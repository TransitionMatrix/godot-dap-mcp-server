package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
)

func TestLaunchTools_Registration(t *testing.T) {
	server := mcp.NewServer()
	RegisterLaunchTools(server)

	// Verify registration doesn't panic
	// The tools should be registered successfully
}

func TestValidateProjectPath_Valid(t *testing.T) {
	// Create a temporary directory with project.godot
	tmpDir, err := os.MkdirTemp("", "godot-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create project.godot file
	projectFile := filepath.Join(tmpDir, "project.godot")
	if err := os.WriteFile(projectFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create project.godot: %v", err)
	}

	// Test validation
	err = validateProjectPath(tmpDir)
	if err != nil {
		t.Errorf("validateProjectPath(%q) returned error: %v", tmpDir, err)
	}
}

func TestValidateProjectPath_Missing(t *testing.T) {
	// Create a temporary directory WITHOUT project.godot
	tmpDir, err := os.MkdirTemp("", "godot-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test validation should fail
	err = validateProjectPath(tmpDir)
	if err == nil {
		t.Error("validateProjectPath should fail when project.godot is missing")
	}
}

func TestValidateProjectPath_NonExistent(t *testing.T) {
	// Test with non-existent directory
	err := validateProjectPath("/nonexistent/path/to/project")
	if err == nil {
		t.Error("validateProjectPath should fail for non-existent path")
	}
}

func TestBuildLaunchArgs_Main(t *testing.T) {
	params := map[string]interface{}{
		"no_debug":  true,
		"profiling": true,
	}

	args := buildLaunchArgs("/path/to/project", "main", params)

	// Verify required fields
	if args["project"] != "/path/to/project" {
		t.Errorf("expected project path %q, got %q", "/path/to/project", args["project"])
	}
	if args["scene"] != "main" {
		t.Errorf("expected scene %q, got %q", "main", args["scene"])
	}
	if args["platform"] != "host" {
		t.Errorf("expected platform %q, got %q", "host", args["platform"])
	}

	// Verify optional boolean fields
	if args["noDebug"] != true {
		t.Errorf("expected noDebug true, got %v", args["noDebug"])
	}
	if args["profiling"] != true {
		t.Errorf("expected profiling true, got %v", args["profiling"])
	}
}

func TestBuildLaunchArgs_CustomScene(t *testing.T) {
	params := map[string]interface{}{
		"debug_collisions": true,
		"debug_navigation": true,
	}

	scenePath := "res://scenes/test.tscn"
	args := buildLaunchArgs("/path/to/project", scenePath, params)

	// Verify scene path
	if args["scene"] != scenePath {
		t.Errorf("expected scene %q, got %q", scenePath, args["scene"])
	}

	// Verify debug flags
	if args["debug_collisions"] != true {
		t.Errorf("expected debug_collisions true, got %v", args["debug_collisions"])
	}
	if args["debug_navigation"] != true {
		t.Errorf("expected debug_navigation true, got %v", args["debug_navigation"])
	}
}

func TestBuildLaunchArgs_Current(t *testing.T) {
	params := map[string]interface{}{}

	args := buildLaunchArgs("/path/to/project", "current", params)

	// Verify scene is "current"
	if args["scene"] != "current" {
		t.Errorf("expected scene %q, got %q", "current", args["scene"])
	}

	// Verify platform defaults to "host"
	if args["platform"] != "host" {
		t.Errorf("expected platform %q, got %q", "host", args["platform"])
	}
}

func TestBuildLaunchArgs_EmptyParams(t *testing.T) {
	params := map[string]interface{}{}

	args := buildLaunchArgs("/path/to/project", "main", params)

	// Should have required fields only
	if len(args) != 3 {
		t.Errorf("expected 3 args (project, scene, platform), got %d", len(args))
	}

	// Verify no optional fields are present
	if _, ok := args["noDebug"]; ok {
		t.Error("noDebug should not be present when not specified")
	}
	if _, ok := args["profiling"]; ok {
		t.Error("profiling should not be present when not specified")
	}
}

func TestLaunchTools_RequireSession(t *testing.T) {
	// Reset global session
	globalSession = nil

	// All launch tools should fail when no session exists
	_, err := GetSession()
	if err == nil {
		t.Error("Launch tools should require an active session")
	}
}
