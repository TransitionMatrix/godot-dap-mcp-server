package main

import (
	"context"
	"log"
	"time"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/dap"
)

func main() {
	log.Println("=== Verifying Fix for 'wrong_path' Error ===")

	// Create session
	// Note: We are using the code directly from internal/dap, 
	// so this runs with the changes we just made to godot.go
	session := dap.NewSession("localhost", 6006)

	// 1. Connect
	ctx := context.Background()
	if err := session.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	log.Println("✓ Connected to Godot")

	// 2. Initialize
	if err := session.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	log.Println("✓ Initialized DAP session")

	// 3. Launch Main Scene
	// This calls the modified LaunchGodotScene -> ToLaunchArgs
	// which should now OMIT the "project" parameter.
	projectPath := "/Users/adp/Projects/godot-dap-mcp-server/tests/fixtures/test-project"
	log.Printf("Attempting launch (project path: %s)...", projectPath)
	
	// We use a timeout for the launch
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := session.LaunchMainScene(ctx, projectPath)
	if err != nil {
		log.Fatalf("❌ Launch FAILED: %v", err)
	}
	
	log.Printf("✓ Launch SUCCESS!")
	log.Printf("  Response: %+v", resp)
	
	// Cleanup
	session.Close()
}
