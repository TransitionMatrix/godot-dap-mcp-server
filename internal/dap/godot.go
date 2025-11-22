package dap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-dap"
)

// SceneLaunchMode defines how Godot should launch scenes
type SceneLaunchMode string

const (
	// SceneLaunchMain launches the project's main scene (from project.godot)
	SceneLaunchMain SceneLaunchMode = "main"

	// SceneLaunchCurrent launches the currently open scene in the editor
	SceneLaunchCurrent SceneLaunchMode = "current"

	// SceneLaunchCustom launches a specific scene by path (use with ScenePath)
	SceneLaunchCustom SceneLaunchMode = "custom"
)

// Platform defines the target platform for launching
type Platform string

const (
	PlatformHost    Platform = "host"
	PlatformAndroid Platform = "android"
	PlatformWeb     Platform = "web"
)

// GodotLaunchConfig contains configuration for launching a Godot scene
type GodotLaunchConfig struct {
	// Project is the absolute path to the Godot project directory
	// Must contain a project.godot file
	Project string

	// Scene determines which scene to launch
	Scene SceneLaunchMode

	// ScenePath is the path to the scene file (e.g., "res://scenes/level1.tscn")
	// Only used when Scene is SceneLaunchCustom
	ScenePath string

	// Platform is the target platform (default: host)
	Platform Platform

	// NoDebug disables debugging features
	NoDebug bool

	// Profiling enables profiling
	Profiling bool

	// DebugCollisions shows collision shapes
	DebugCollisions bool

	// DebugPaths shows navigation paths
	DebugPaths bool

	// DebugNavigation shows navigation debug
	DebugNavigation bool

	// AdditionalOptions contains additional command-line options
	AdditionalOptions string
}

// Validate checks if the launch configuration is valid
func (c *GodotLaunchConfig) Validate() error {
	// Check that project path is provided
	if c.Project == "" {
		return fmt.Errorf("project path is required")
	}

	// Check that project.godot exists
	projectFile := filepath.Join(c.Project, "project.godot")
	if _, err := os.Stat(projectFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("project.godot not found in %s", c.Project)
		}
		return fmt.Errorf("failed to check project.godot: %w", err)
	}

	// Check scene configuration
	if c.Scene == SceneLaunchCustom && c.ScenePath == "" {
		return fmt.Errorf("scene path is required when using custom scene launch mode")
	}

	// Set default platform if not specified
	if c.Platform == "" {
		c.Platform = PlatformHost
	}

	return nil
}

// ToLaunchArgs converts the config to DAP launch request arguments
func (c *GodotLaunchConfig) ToLaunchArgs() map[string]interface{} {
	args := map[string]interface{}{
		"project":          c.Project,
		"platform":         string(c.Platform),
		"noDebug":          c.NoDebug,
		"profiling":        c.Profiling,
		"debug_collisions": c.DebugCollisions,
		"debug_paths":      c.DebugPaths,
		"debug_navigation": c.DebugNavigation,
	}

	// Set scene based on launch mode
	switch c.Scene {
	case SceneLaunchMain:
		args["scene"] = "main"
	case SceneLaunchCurrent:
		args["scene"] = "current"
	case SceneLaunchCustom:
		args["scene"] = c.ScenePath
	}

	// Add additional options if provided
	if c.AdditionalOptions != "" {
		args["additional_options"] = c.AdditionalOptions
	}

	return args
}

// LaunchGodotScene launches a Godot scene with the given configuration
func (s *Session) LaunchGodotScene(ctx context.Context, config *GodotLaunchConfig) (*dap.LaunchResponse, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid launch configuration: %w", err)
	}

	// Launch with the converted arguments using the Godot-specific sequence
	// (Launch -> ConfigurationDone -> Wait for ConfigDone -> Wait for Launch)
	return s.client.LaunchWithConfigurationDone(ctx, config.ToLaunchArgs())
}

// LaunchMainScene is a convenience method to launch the project's main scene
func (s *Session) LaunchMainScene(ctx context.Context, projectPath string) (*dap.LaunchResponse, error) {
	config := &GodotLaunchConfig{
		Project:  projectPath,
		Scene:    SceneLaunchMain,
		Platform: PlatformHost,
	}
	return s.LaunchGodotScene(ctx, config)
}

// LaunchCurrentScene is a convenience method to launch the currently open scene
func (s *Session) LaunchCurrentScene(ctx context.Context, projectPath string) (*dap.LaunchResponse, error) {
	config := &GodotLaunchConfig{
		Project:  projectPath,
		Scene:    SceneLaunchCurrent,
		Platform: PlatformHost,
	}
	return s.LaunchGodotScene(ctx, config)
}

// LaunchCustomScene is a convenience method to launch a specific scene
func (s *Session) LaunchCustomScene(ctx context.Context, projectPath string, scenePath string) (*dap.LaunchResponse, error) {
	config := &GodotLaunchConfig{
		Project:   projectPath,
		Scene:     SceneLaunchCustom,
		ScenePath: scenePath,
		Platform:  PlatformHost,
	}
	return s.LaunchGodotScene(ctx, config)
}

// Additional Godot-specific DAP commands can be added here as needed
// For example: breakpoint management, stepping, evaluation, etc.
