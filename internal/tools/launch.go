package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/dap"
	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
)

// RegisterLaunchTools registers scene launching tools (main, custom, current)
func RegisterLaunchTools(server *mcp.Server) {
	// godot_launch_main_scene - Launch project's main scene
	server.RegisterTool(mcp.Tool{
		Name: "godot_launch_main_scene",
		Description: `Launch the project's main scene defined in project.godot.

This tool starts the game using the main scene configured in your Godot project.
This is equivalent to pressing F5 in the Godot editor.

Prerequisites:
- Must be connected to Godot DAP server (call godot_connect first)
- Must complete DAP configuration handshake
- Project must have a main scene defined in project.godot

Use this tool:
- To start debugging the game from the main entry point
- To test the complete game flow from the beginning
- When you want the default launch behavior

Launch Flow:
1. Sends launch request with scene="main"
2. Sends configurationDone to trigger actual launch
3. Game starts and runs until breakpoint/pause/exit

Example: Launch main scene with default settings
godot_launch_main_scene(project="/path/to/godot/project")

Example: Launch with debugging disabled
godot_launch_main_scene(project="/path/to/project", no_debug=true)

Example: Launch with profiling enabled
godot_launch_main_scene(project="/path/to/project", profiling=true)`,

		Parameters: []mcp.Parameter{
			{
				Name:        "project",
				Type:        "string",
				Required:    true,
				Description: "Absolute path to Godot project directory (must contain project.godot)",
			},
			{
				Name:        "no_debug",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "If true, run without debugger (breakpoints will be ignored)",
			},
			{
				Name:        "profiling",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Enable performance profiling",
			},
			{
				Name:        "debug_collisions",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Show collision shapes visually",
			},
			{
				Name:        "debug_navigation",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Show navigation mesh",
			},
		},

		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Get active session
			session, err := GetSession()
			if err != nil {
				return nil, fmt.Errorf("%w\n\nPlease call godot_connect first to establish a DAP session", err)
			}

			// Get and validate project path
			projectPath, ok := params["project"].(string)
			if !ok || projectPath == "" {
				return nil, fmt.Errorf("project parameter is required and must be a string")
			}

			// Validate project path
			if err := validateProjectPath(projectPath); err != nil {
				return nil, err
			}

			// Build launch configuration
			config := &dap.GodotLaunchConfig{
				Project:         projectPath,
				Scene:           dap.SceneLaunchMain,
				Platform:        dap.PlatformHost,
				NoDebug:         getBoolParam(params, "no_debug"),
				Profiling:       getBoolParam(params, "profiling"),
				DebugCollisions: getBoolParam(params, "debug_collisions"),
				DebugNavigation: getBoolParam(params, "debug_navigation"),
			}

			// Launch scene
			ctx, cancel := dap.WithCommandTimeout(context.Background())
			defer cancel()

			if _, err := session.LaunchGodotScene(ctx, config); err != nil {
				return nil, fmt.Errorf("failed to launch main scene: %w", err)
			}

			return map[string]interface{}{
				"status":  "launched",
				"message": "Main scene launched successfully",
				"project": projectPath,
				"scene":   "main",
			}, nil
		},
	})

	// godot_launch_scene - Launch specific scene by path
	server.RegisterTool(mcp.Tool{
		Name: "godot_launch_scene",
		Description: `Launch a specific scene by resource path.

This tool starts the game from a specific scene file, allowing you to test
individual scenes without changing your project configuration.

Prerequisites:
- Must be connected to Godot DAP server (call godot_connect first)
- Must complete DAP configuration handshake
- Scene file must exist in the project

Use this tool:
- To test a specific scene in isolation
- To debug a particular game level or screen
- When you want to skip to a specific point in the game

Scene Path Format:
- Use Godot resource path format: "res://path/to/scene.tscn"
- Path is relative to project root
- Must be a valid .tscn or .scn file

Launch Flow:
1. Sends launch request with scene="res://path/to/scene.tscn"
2. Sends configurationDone to trigger actual launch
3. Game starts from specified scene

Example: Launch specific test scene
godot_launch_scene(project="/path/to/project", scene="res://scenes/test_level.tscn")

Example: Launch with debugging disabled
godot_launch_scene(project="/path/to/project", scene="res://scenes/level_2.tscn", no_debug=true)

Example: Launch with collision visualization
godot_launch_scene(project="/path/to/project", scene="res://test.tscn", debug_collisions=true)`,

		Parameters: []mcp.Parameter{
			{
				Name:        "project",
				Type:        "string",
				Required:    true,
				Description: "Absolute path to Godot project directory (must contain project.godot)",
			},
			{
				Name:        "scene",
				Type:        "string",
				Required:    true,
				Description: "Godot resource path to scene file (e.g., \"res://scenes/test.tscn\")",
			},
			{
				Name:        "no_debug",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "If true, run without debugger (breakpoints will be ignored)",
			},
			{
				Name:        "profiling",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Enable performance profiling",
			},
			{
				Name:        "debug_collisions",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Show collision shapes visually",
			},
			{
				Name:        "debug_navigation",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Show navigation mesh",
			},
		},

		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Get active session
			session, err := GetSession()
			if err != nil {
				return nil, fmt.Errorf("%w\n\nPlease call godot_connect first to establish a DAP session", err)
			}

			// Get and validate project path
			projectPath, ok := params["project"].(string)
			if !ok || projectPath == "" {
				return nil, fmt.Errorf("project parameter is required and must be a string")
			}

			// Validate project path
			if err := validateProjectPath(projectPath); err != nil {
				return nil, err
			}

			// Get scene path
			scenePath, ok := params["scene"].(string)
			if !ok || scenePath == "" {
				return nil, fmt.Errorf("scene parameter is required and must be a string")
			}

			// Build launch configuration
			config := &dap.GodotLaunchConfig{
				Project:         projectPath,
				Scene:           dap.SceneLaunchCustom,
				ScenePath:       scenePath,
				Platform:        dap.PlatformHost,
				NoDebug:         getBoolParam(params, "no_debug"),
				Profiling:       getBoolParam(params, "profiling"),
				DebugCollisions: getBoolParam(params, "debug_collisions"),
				DebugNavigation: getBoolParam(params, "debug_navigation"),
			}

			// Launch scene
			ctx, cancel := dap.WithCommandTimeout(context.Background())
			defer cancel()

			if _, err := session.LaunchGodotScene(ctx, config); err != nil {
				return nil, fmt.Errorf("failed to launch scene %s: %w", scenePath, err)
			}

			return map[string]interface{}{
				"status":  "launched",
				"message": fmt.Sprintf("Scene %s launched successfully", scenePath),
				"project": projectPath,
				"scene":   scenePath,
			}, nil
		},
	})

	// godot_launch_current_scene - Launch currently open scene in editor
	server.RegisterTool(mcp.Tool{
		Name: "godot_launch_current_scene",
		Description: `Launch the currently open scene in the Godot editor.

This tool starts the game from whatever scene is currently open/active in the
Godot editor. This is equivalent to pressing F6 in the Godot editor.

Prerequisites:
- Must be connected to Godot DAP server (call godot_connect first)
- Must complete DAP configuration handshake
- Godot editor must have a scene open

Use this tool:
- To quickly test the scene you're currently editing
- When iterating on a specific scene during development
- To debug the scene in your current editor tab

Launch Flow:
1. Sends launch request with scene="current"
2. Godot determines which scene is currently open in the editor
3. Sends configurationDone to trigger actual launch
4. Game starts from editor's active scene

Example: Launch current scene with default settings
godot_launch_current_scene(project="/path/to/godot/project")

Example: Launch with debugging disabled
godot_launch_current_scene(project="/path/to/project", no_debug=true)

Example: Launch with profiling and collision debug
godot_launch_current_scene(project="/path/to/project", profiling=true, debug_collisions=true)`,

		Parameters: []mcp.Parameter{
			{
				Name:        "project",
				Type:        "string",
				Required:    true,
				Description: "Absolute path to Godot project directory (must contain project.godot)",
			},
			{
				Name:        "no_debug",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "If true, run without debugger (breakpoints will be ignored)",
			},
			{
				Name:        "profiling",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Enable performance profiling",
			},
			{
				Name:        "debug_collisions",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Show collision shapes visually",
			},
			{
				Name:        "debug_navigation",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Show navigation mesh",
			},
		},

		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Get active session
			session, err := GetSession()
			if err != nil {
				return nil, fmt.Errorf("%w\n\nPlease call godot_connect first to establish a DAP session", err)
			}

			// Get and validate project path
			projectPath, ok := params["project"].(string)
			if !ok || projectPath == "" {
				return nil, fmt.Errorf("project parameter is required and must be a string")
			}

			// Validate project path
			if err := validateProjectPath(projectPath); err != nil {
				return nil, err
			}

			// Build launch configuration
			config := &dap.GodotLaunchConfig{
				Project:         projectPath,
				Scene:           dap.SceneLaunchCurrent,
				Platform:        dap.PlatformHost,
				NoDebug:         getBoolParam(params, "no_debug"),
				Profiling:       getBoolParam(params, "profiling"),
				DebugCollisions: getBoolParam(params, "debug_collisions"),
				DebugNavigation: getBoolParam(params, "debug_navigation"),
			}

			// Launch scene
			ctx, cancel := dap.WithCommandTimeout(context.Background())
			defer cancel()

			if _, err := session.LaunchGodotScene(ctx, config); err != nil {
				return nil, fmt.Errorf("failed to launch current scene: %w", err)
			}

			return map[string]interface{}{
				"status":  "launched",
				"message": "Current scene launched successfully",
				"project": projectPath,
				"scene":   "current",
			}, nil
		},
	})
}

// validateProjectPath checks if the project path is valid and contains project.godot
func validateProjectPath(path string) error {
	projectFile := filepath.Join(path, "project.godot")
	if _, err := os.Stat(projectFile); os.IsNotExist(err) {
		return fmt.Errorf(`Invalid project path: project.godot not found at %s

Possible causes:
1. Path does not point to a Godot project directory
2. The path is relative instead of absolute
3. The project.godot file has been moved or deleted

Solutions:
1. Ensure the path points to the directory containing project.godot
2. Use an absolute path: /full/path/to/project
3. Verify the project exists: ls %s`, projectFile, projectFile)
	}
	return nil
}

// buildLaunchArgs constructs the launch arguments map from parameters
func buildLaunchArgs(projectPath string, scene string, params map[string]interface{}) map[string]interface{} {
	args := map[string]interface{}{
		"project":  projectPath,
		"scene":    scene,
		"platform": "host", // Always use host platform for now
	}

	// Add optional boolean parameters
	if noDebug, ok := params["no_debug"].(bool); ok {
		args["noDebug"] = noDebug
	}
	if profiling, ok := params["profiling"].(bool); ok {
		args["profiling"] = profiling
	}
	if debugCollisions, ok := params["debug_collisions"].(bool); ok {
		args["debug_collisions"] = debugCollisions
	}
	if debugNavigation, ok := params["debug_navigation"].(bool); ok {
		args["debug_navigation"] = debugNavigation
	}

	return args
}

// getBoolParam extracts a boolean parameter from the map, returning false if not present or invalid
func getBoolParam(params map[string]interface{}, name string) bool {
	if val, ok := params[name]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return false
}
