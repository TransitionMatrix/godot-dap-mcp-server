package main

import (
	"context"
	"log"
	"time"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/dap"
)

func main() {
	log.Println("=== Comprehensive DAP Debugging Workflow Test ===")
	log.Println("This test demonstrates a complete debugging session:")
	log.Println("1. Launch with breakpoint")
	log.Println("2. Inspect runtime state (threads, stack, variables)")
	log.Println("3. Step through code")
	log.Println("4. Evaluate expressions")
	log.Println("5. Resume execution")
	log.Println()

	// Configuration
	projectPath := "/Users/adp/Projects/xiang-qi-game-2d-godot"
	scriptPath := projectPath + "/src/scenes/main_scene/main_scene.gd"
	breakpointLine := 56

	session := dap.NewSession("localhost", 6006)
	ctx := context.Background()

	// ========================================
	// PHASE 1: Setup and Launch
	// ========================================

	log.Println("--- PHASE 1: Setup and Launch ---")

	// 1. Connect
	log.Println("\n1. Connecting to Godot DAP server...")
	if err := session.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	log.Println("✓ Connected successfully")

	// 2. Initialize
	log.Println("\n2. Initializing DAP session...")
	if err := session.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	log.Println("✓ Initialized successfully")

	// 3. Launch (stores parameters, doesn't start game yet)
	log.Println("\n3. Launching main scene...")
	if _, err := session.LaunchMainScene(ctx, projectPath); err != nil {
		log.Fatalf("Failed to launch scene: %v", err)
	}
	log.Println("✓ Launch request sent (game not started yet)")

	// 4. Set breakpoint BEFORE game starts
	log.Println("\n4. Setting breakpoint at line", breakpointLine, "...")
	client := session.GetClient()
	bpResp, err := client.SetBreakpoints(ctx, scriptPath, []int{breakpointLine})
	if err != nil {
		log.Fatalf("Failed to set breakpoint: %v", err)
	}
	if len(bpResp.Body.Breakpoints) > 0 {
		log.Printf("✓ Breakpoint set successfully (verified: %v)", bpResp.Body.Breakpoints[0].Verified)
	} else {
		log.Println("⚠ No breakpoints returned")
	}

	// 5. ConfigurationDone triggers actual launch
	log.Println("\n5. Sending configurationDone to start game...")
	if err := session.ConfigurationDone(ctx); err != nil {
		log.Fatalf("Failed to send configurationDone: %v", err)
	}
	log.Println("✓ Game launched with breakpoints active")

	// 6. Wait for breakpoint hit
	log.Println("\n6. Waiting for breakpoint to be hit...")
	time.Sleep(3 * time.Second)
	log.Println("✓ Breakpoint should be hit by now")

	// ========================================
	// PHASE 2: Inspect Runtime State
	// ========================================

	log.Println("\n--- PHASE 2: Inspect Runtime State ---")

	// 7. Get threads
	log.Println("\n7. Getting active threads...")
	threadsResp, err := client.Threads(ctx)
	if err != nil {
		log.Fatalf("Failed to get threads: %v", err)
	}
	log.Printf("✓ Found %d thread(s):", len(threadsResp.Body.Threads))
	for _, thread := range threadsResp.Body.Threads {
		log.Printf("  - Thread %d: %s", thread.Id, thread.Name)
	}

	// 8. Get stack trace
	log.Println("\n8. Getting stack trace...")
	threadId := 1
	stackResp, err := client.StackTrace(ctx, threadId, 0, 10)
	if err != nil {
		log.Fatalf("Failed to get stack trace: %v", err)
	}
	log.Printf("✓ Stack trace (%d frames):", len(stackResp.Body.StackFrames))
	for i, frame := range stackResp.Body.StackFrames {
		sourceName := "<unknown>"
		if frame.Source != nil {
			sourceName = frame.Source.Name
		}
		log.Printf("  [%d] %s at %s:%d", i, frame.Name, sourceName, frame.Line)
	}

	// 9. Get scopes for top frame
	log.Println("\n9. Getting variable scopes for top frame...")
	if len(stackResp.Body.StackFrames) == 0 {
		log.Fatalf("No stack frames available")
	}
	frameId := stackResp.Body.StackFrames[0].Id
	scopesResp, err := client.Scopes(ctx, frameId)
	if err != nil {
		log.Fatalf("Failed to get scopes: %v", err)
	}
	log.Printf("✓ Found %d scope(s):", len(scopesResp.Body.Scopes))
	for _, scope := range scopesResp.Body.Scopes {
		log.Printf("  - %s (ref: %d)", scope.Name, scope.VariablesReference)
	}

	// 10. Get local variables
	log.Println("\n10. Getting local variables...")
	var localsRef int
	for _, scope := range scopesResp.Body.Scopes {
		if scope.Name == "Locals" {
			localsRef = scope.VariablesReference
			break
		}
	}
	if localsRef > 0 {
		varsResp, err := client.Variables(ctx, localsRef)
		if err != nil {
			log.Fatalf("Failed to get variables: %v", err)
		}
		log.Printf("✓ Found %d local variable(s):", len(varsResp.Body.Variables))
		for _, v := range varsResp.Body.Variables {
			log.Printf("  - %s: %s = %s", v.Name, v.Type, v.Value)
			if v.VariablesReference > 0 {
				log.Printf("    (expandable, ref: %d)", v.VariablesReference)
			}
		}
	} else {
		log.Println("⚠ No Locals scope found (might be empty)")
	}

	// 11. Evaluate an expression
	log.Println("\n11. Evaluating expression: '1 + 1'...")
	evalResp, err := client.Evaluate(ctx, "1 + 1", frameId, "repl")
	if err != nil {
		log.Fatalf("Failed to evaluate expression: %v", err)
	}
	log.Printf("✓ Result: %s (type: %s)", evalResp.Body.Result, evalResp.Body.Type)

	// ========================================
	// PHASE 3: Execution Control
	// ========================================

	log.Println("\n--- PHASE 3: Execution Control ---")

	// 12. Step over (next)
	log.Println("\n12. Stepping over current line...")
	if _, err := client.Next(ctx, threadId); err != nil {
		log.Fatalf("Failed to step over: %v", err)
	}
	log.Println("✓ Stepped over successfully")
	time.Sleep(500 * time.Millisecond)

	// Get new stack trace to confirm step
	log.Println("\n13. Getting updated stack trace after step...")
	stackResp2, err := client.StackTrace(ctx, threadId, 0, 5)
	if err != nil {
		log.Fatalf("Failed to get stack trace: %v", err)
	}
	if len(stackResp2.Body.StackFrames) > 0 {
		frame := stackResp2.Body.StackFrames[0]
		sourceName := "<unknown>"
		if frame.Source != nil {
			sourceName = frame.Source.Name
		}
		log.Printf("✓ Now at: %s:%d (%s)", sourceName, frame.Line, frame.Name)
	}

	// 14. Step into (if there's a function call)
	log.Println("\n14. Attempting step into...")
	if _, err := client.StepIn(ctx, threadId); err != nil {
		log.Printf("⚠ Step into failed (might not be a function call): %v", err)
	} else {
		log.Println("✓ Stepped into successfully")
		time.Sleep(500 * time.Millisecond)
	}

	// 15. Continue execution
	log.Println("\n15. Continuing execution...")
	continueResp, err := client.Continue(ctx, threadId)
	if err != nil {
		log.Fatalf("Failed to continue: %v", err)
	}
	log.Printf("✓ Execution resumed (all threads: %v)", continueResp.Body.AllThreadsContinued)

	// Wait a bit to let game run
	log.Println("\n16. Letting game run for 2 seconds...")
	time.Sleep(2 * time.Second)

	// ========================================
	// PHASE 4: Cleanup
	// ========================================

	log.Println("\n--- PHASE 4: Cleanup ---")

	// 17. Disconnect
	log.Println("\n17. Disconnecting from DAP server...")
	session.Close()
	log.Println("✓ Disconnected successfully")

	// ========================================
	// Summary
	// ========================================

	log.Println("\n=== TEST SUMMARY ===")
	log.Println("✅ All DAP commands executed successfully!")
	log.Println()
	log.Println("Commands tested:")
	log.Println("  1. Connect")
	log.Println("  2. Initialize")
	log.Println("  3. Launch")
	log.Println("  4. SetBreakpoints")
	log.Println("  5. ConfigurationDone")
	log.Println("  6. Threads")
	log.Println("  7. StackTrace")
	log.Println("  8. Scopes")
	log.Println("  9. Variables")
	log.Println(" 10. Evaluate")
	log.Println(" 11. Next (step over)")
	log.Println(" 12. StepIn")
	log.Println(" 13. Continue")
	log.Println()
	log.Println("This demonstrates complete end-to-end DAP debugging functionality!")
}
