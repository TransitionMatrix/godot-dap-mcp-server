package tools

import (
	"testing"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
)

func TestBreakpointTools_Registration(t *testing.T) {
	server := mcp.NewServer()
	RegisterBreakpointTools(server)

	// Verify registration doesn't panic
	// The tools should be registered successfully
}

func TestIsResPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "valid res path",
			path:     "res://scripts/player.gd",
			expected: true,
		},
		{
			name:     "absolute path",
			path:     "/Users/dev/project/player.gd",
			expected: false,
		},
		{
			name:     "relative path",
			path:     "scripts/player.gd",
			expected: false,
		},
		{
			name:     "empty path",
			path:     "",
			expected: false,
		},
		{
			name:     "res prefix only",
			path:     "res:",
			expected: false,
		},
		{
			name:     "res:/ single slash",
			path:     "res:/player.gd",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isResPath(tt.path)
			if result != tt.expected {
				t.Errorf("isResPath(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestBreakpointTools_RequireSession(t *testing.T) {
	// Reset global session
	globalSession = nil

	// All breakpoint tools should fail when no session exists
	_, err := GetSession()
	if err == nil {
		t.Error("Breakpoint tools should require an active session")
	}
}

func TestSetBreakpoint_PathValidation(t *testing.T) {
	// Test path validation logic
	// Valid paths: absolute or res://
	// Invalid paths: relative (without res://)

	tests := []struct {
		name      string
		path      string
		shouldErr bool
	}{
		{
			name:      "absolute path",
			path:      "/Users/dev/project/player.gd",
			shouldErr: false,
		},
		{
			name:      "res:// path",
			path:      "res://scripts/player.gd",
			shouldErr: false,
		},
		{
			name:      "relative path",
			path:      "scripts/player.gd",
			shouldErr: true,
		},
		{
			name:      "empty path",
			path:      "",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test isResPath helper and path validation logic
			isRes := isResPath(tt.path)
			isAbs := len(tt.path) > 0 && tt.path[0] == '/'

			isValid := isRes || isAbs
			if (tt.path == "") {
				isValid = false
			}

			shouldBeValid := !tt.shouldErr
			if isValid != shouldBeValid {
				t.Errorf("Path validation for %q: got valid=%v, want valid=%v", tt.path, isValid, shouldBeValid)
			}
		})
	}
}
