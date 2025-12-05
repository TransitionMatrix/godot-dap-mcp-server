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

func TestResolveGodotPath(t *testing.T) {
	projectRoot := "/Users/dev/project"

	tests := []struct {
		name        string
		path        string
		projectRoot string
		wantPath    string
		wantErr     bool
	}{
		{
			name:        "valid res path with root",
			path:        "res://scripts/player.gd",
			projectRoot: projectRoot,
			wantPath:    "/Users/dev/project/scripts/player.gd",
			wantErr:     false,
		},
		{
			name:        "valid res path without root",
			path:        "res://scripts/player.gd",
			projectRoot: "",
			wantErr:     true,
		},
		{
			name:        "absolute path",
			path:        "/Users/dev/project/player.gd",
			projectRoot: projectRoot,
			wantPath:    "/Users/dev/project/player.gd",
			wantErr:     false,
		},
		{
			name:        "relative path",
			path:        "scripts/player.gd",
			projectRoot: projectRoot,
			wantErr:     true,
		},
		{
			name:        "empty path",
			path:        "",
			projectRoot: projectRoot,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveGodotPath(tt.path, tt.projectRoot)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveGodotPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantPath {
				t.Errorf("resolveGodotPath() = %v, want %v", got, tt.wantPath)
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
