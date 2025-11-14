package main

import (
	"context"
	"log"
	"time"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/dap"
)

func main() {
	log.Println("=== Testing DAP Launch → SetBreakpoints Sequence ===")

	// Create session
	session := dap.NewSession("localhost", 6006)

	// Step 1: Connect
	log.Println("\n1. Connecting to Godot DAP server...")
	ctx := context.Background()
	if err := session.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	log.Printf("✓ Connected (state: %s)", session.GetState())

	// Step 2: Initialize
	log.Println("\n2. Initializing DAP session...")
	if err := session.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	log.Printf("✓ Initialized (state: %s)", session.GetState())

	// Step 3: Launch the main scene (stores params, doesn't start game yet!)
	log.Println("\n3. Launching main scene...")
	projectPath := "/Users/adp/Projects/xiang-qi-game-2d-godot"
	launchResp, err := session.LaunchMainScene(ctx, projectPath)
	if err != nil {
		log.Fatalf("Failed to launch scene: %v", err)
	}
	log.Printf("✓ Launch request sent (state: %s)", session.GetState())
	log.Printf("  Launch response: %+v", launchResp)

	// Step 4: Set breakpoints BEFORE game starts!
	log.Println("\n4. Setting breakpoints (before game starts)...")
	client := session.GetClient()

	// Use absolute path, not res:// path (Godot requires absolute filesystem paths)
	scriptPath := projectPath + "/src/scenes/main_scene/main_scene.gd"
	breakpoints, err := client.SetBreakpoints(ctx, scriptPath, []int{56})
	if err != nil {
		log.Fatalf("Failed to set breakpoint: %v", err)
	}
	log.Printf("✓ Breakpoint set successfully!")
	log.Printf("  Breakpoints: %+v", breakpoints)

	// Step 5: ConfigurationDone (NOW the game starts with breakpoints active!)
	log.Println("\n5. Sending configurationDone to trigger launch...")
	if err := session.ConfigurationDone(ctx); err != nil {
		log.Fatalf("Failed to send configurationDone: %v", err)
	}
	log.Printf("✓ Configuration done - game launching with breakpoints active (state: %s)", session.GetState())

	// Wait for the game to hit breakpoint or run
	log.Println("\n6. Waiting for breakpoint or game execution...")
	time.Sleep(5 * time.Second)

	log.Println("\n=== SUCCESS! Launch → SetBreakpoints sequence works! ===")

	// Clean up
	log.Println("\n7. Disconnecting...")
	if err := session.Close(); err != nil {
		log.Printf("Warning: disconnect failed: %v", err)
	}
	log.Println("✓ Disconnected")
}
