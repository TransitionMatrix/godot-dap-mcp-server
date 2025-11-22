package main

import (
	"log"
	"os"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/tools"
)

func main() {
	// Configure logging to file
	logFile, err := os.OpenFile("/tmp/godot-dap-mcp-server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
	} else {
		defer logFile.Close()
		log.SetOutput(logFile)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("==========================================")
	log.Println("Starting Godot DAP MCP Server...")

	// Create MCP server
	server := mcp.NewServer()

	// Register all tools
	tools.RegisterAll(server)

	log.Println("Tools registered, ready to accept requests")

	// Start server (blocks until EOF or error)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server shutdown complete")
}
