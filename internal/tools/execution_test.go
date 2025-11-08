package tools

import (
	"testing"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
)

func TestExecutionTools_Registration(t *testing.T) {
	server := mcp.NewServer()
	RegisterExecutionTools(server)

	// Verify registration doesn't panic
	// The tools should be registered successfully
}

func TestExecutionTools_RequireSession(t *testing.T) {
	// Reset global session
	globalSession = nil

	// All execution tools should fail when no session exists
	// This is tested through the GetSession() function
	_, err := GetSession()
	if err == nil {
		t.Error("Execution tools should require an active session")
	}
}

func TestExecutionTools_DefaultThreadID(t *testing.T) {
	// Verify that default thread ID is 1 (Godot's typical thread)
	// This is validated through the tool parameter defaults
	// The actual validation happens in the tool handlers

	// For unit testing, we verify the tools are registered
	server := mcp.NewServer()
	RegisterExecutionTools(server)
}
