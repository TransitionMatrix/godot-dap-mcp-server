package tools

import (
	"strings"
	"testing"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
)

func TestConnectionTools_Registration(t *testing.T) {
	server := mcp.NewServer()
	RegisterConnectionTools(server)

	// Verify registration doesn't panic
	// The tools should be registered successfully
}

func TestGetSession_NoSession(t *testing.T) {
	// Reset global session
	globalSession = nil

	_, err := GetSession()
	if err == nil {
		t.Error("GetSession should error when no session exists")
	}

	expectedMsg := "not connected to Godot DAP server"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedMsg, err)
	}
}

func TestGodotConnect_ToolMetadata(t *testing.T) {
	server := mcp.NewServer()
	RegisterConnectionTools(server)

	// We can't easily test the full handler without exposing internals,
	// but we can verify the tools are registered by checking they don't panic
	// and that the registration structure is correct

	// Test that godot_connect tool is registered with correct schema
	// This is validated through integration tests or by examining the server
	// For unit tests, we verify registration succeeds
}

func TestGodotDisconnect_NoConnection(t *testing.T) {
	// Reset global session
	globalSession = nil

	server := mcp.NewServer()
	RegisterConnectionTools(server)

	// In a real scenario, we'd call the disconnect tool handler
	// For unit tests, we verify the session management logic
	session, err := GetSession()
	if err == nil {
		t.Error("Should error when no session exists")
	}
	if session != nil {
		t.Error("Should return nil session when disconnected")
	}
}
