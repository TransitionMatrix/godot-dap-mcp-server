package tools

import (
	"testing"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
)

func TestInspectionTools_Registration(t *testing.T) {
	server := mcp.NewServer()
	RegisterInspectionTools(server)

	// Verify registration doesn't panic
	// The 5 inspection tools should be registered successfully:
	// - godot_get_threads
	// - godot_get_stack_trace
	// - godot_get_scopes
	// - godot_get_variables
	// - godot_evaluate
}

func TestInspectionTools_RequireSession(t *testing.T) {
	// Reset global session
	globalSession = nil

	// All inspection tools should fail when no session exists
	// This is tested through the GetSession() function
	_, err := GetSession()
	if err == nil {
		t.Error("Inspection tools should require an active session")
	}
}
