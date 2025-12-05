package tools

import (
	"fmt"
	"path/filepath"
	"strings"
)

// resolveGodotPath converts a file path to an absolute path that Godot DAP can understand.
// It handles:
// 1. Absolute paths: returned as-is
// 2. res:// paths: converted to absolute path using projectRoot
// 3. Relative paths: rejected (must be absolute or res://)
func resolveGodotPath(path string, projectRoot string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Handle res:// paths
	if strings.HasPrefix(path, "res://") {
		if projectRoot == "" {
			return "", fmt.Errorf("cannot resolve res:// path '%s': project root not set. Please provide 'project' argument in godot_connect or use absolute path", path)
		}

		relativePath := strings.TrimPrefix(path, "res://")
		return filepath.Join(projectRoot, relativePath), nil
	}

	// Handle absolute paths
	if filepath.IsAbs(path) {
		return path, nil
	}

	return "", fmt.Errorf("path must be absolute or start with res:// (got: %s)", path)
}
