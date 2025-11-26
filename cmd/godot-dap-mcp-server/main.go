package main

import (
	"log"
	"os"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/tools"
)

func main() {
	// Configure logging
	// By default, log to stderr (MCP clients usually capture this)
	// Can be overridden by GODOT_MCP_LOG_FILE environment variable
	logOutput := os.Stderr
	
	if logPath := os.Getenv("GODOT_MCP_LOG_FILE"); logPath != "" {
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Failed to open log file %s: %v", logPath, err)
		} else {
			defer f.Close()
			logOutput = f
		}
	}
	
	log.SetOutput(logOutput)
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
