package tools

import (
	"context"
	"fmt"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/dap"
	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
)

// RegisterBreakpointTools registers breakpoint management tools
func RegisterBreakpointTools(server *mcp.Server) {
	// godot_set_breakpoint - Set a breakpoint
	server.RegisterTool(mcp.Tool{
		Name: "godot_set_breakpoint",
		Description: `Set a breakpoint in a GDScript file at the specified line.

This tool sets a breakpoint that will pause game execution when that line is reached.
The breakpoint will be active for all subsequent runs until cleared.

Prerequisites:
- Must be connected to Godot DAP server (call godot_connect first)
- File path must be absolute OR start with "res://" (if project path was set in godot_connect)
- Line number must be valid (positive integer)

Use this tool:
- Before launching a scene to pause at specific points
- To investigate code behavior at runtime
- To inspect variables at specific locations

Godot will verify the breakpoint and may adjust the line number if the specified
line is not executable (e.g., blank line, comment). The verified breakpoint
location will be returned.

File path requirements:
- Can be absolute path: /path/to/project/scripts/player.gd
- Can be res:// path: res://scripts/player.gd (Requires 'project' arg in godot_connect)
- Must point to a .gd (GDScript) file

Example: Set breakpoint in player script
godot_set_breakpoint(file="res://scripts/player.gd", line=45)

Example: Set breakpoint with absolute path
godot_set_breakpoint(file="/Users/dev/myproject/player.gd", line=12)`,

		Parameters: []mcp.Parameter{
			{
				Name:        "file",
				Type:        "string",
				Required:    true,
				Description: "Path to GDScript file (absolute or res:// path)",
			},
			{
				Name:        "line",
				Type:        "number",
				Required:    true,
				Description: "Line number where breakpoint should be set (1-indexed)",
			},
		},

		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Get active session
			session, err := GetSession()
			if err != nil {
				return nil, fmt.Errorf("%w\n\nPlease call godot_connect first to establish a DAP session", err)
			}

			// Get file parameter
			file, ok := params["file"].(string)
			if !ok || file == "" {
				return nil, fmt.Errorf("file parameter is required and must be a non-empty string")
			}

			// Get line parameter
			lineFloat, ok := params["line"].(float64)
			if !ok || lineFloat < 1 {
				return nil, fmt.Errorf("line parameter is required and must be a positive integer")
			}
			line := int(lineFloat)

			// Resolve file path
			normalizedFile, err := resolveGodotPath(file, session.GetProjectRoot())
			if err != nil {
				return nil, err
			}

			// Send setBreakpoints request
			ctx, cancel := dap.WithCommandTimeout(context.Background())
			defer cancel()

			client := session.GetClient()
			resp, err := client.SetBreakpoints(ctx, normalizedFile, []int{line})
			if err != nil {
				return nil, fmt.Errorf("failed to set breakpoint: %w", err)
			}

			// Check if breakpoint was verified
			if len(resp.Body.Breakpoints) == 0 {
				return nil, fmt.Errorf("no breakpoints were set (file may not exist or line may be invalid)")
			}

			bp := resp.Body.Breakpoints[0]
			if !bp.Verified {
				return map[string]interface{}{
					"status":   "unverified",
					"message":  "Breakpoint set but not verified by Godot",
					"file":     file,
					"requested_line": line,
					"actual_line":    bp.Line,
					"reason":   "File may not be loaded or line may not be executable",
				}, nil
			}

			result := map[string]interface{}{
				"status":          "verified",
				"message":         fmt.Sprintf("Breakpoint set at %s:%d", file, bp.Line),
				"file":            file,
				"requested_line":  line,
				"actual_line":     bp.Line,
				"id":              bp.Id,
			}

			// Add message if line was adjusted
			if bp.Line != line {
				result["adjusted"] = true
				result["message"] = fmt.Sprintf("Breakpoint set at %s:%d (adjusted from line %d)", file, bp.Line, line)
			}

			return result, nil
		},
	})

	// godot_clear_breakpoint - Clear a breakpoint
	server.RegisterTool(mcp.Tool{
		Name: "godot_clear_breakpoint",
		Description: `Clear a breakpoint from a GDScript file.

This tool removes the breakpoint at the specified line in the given file.
Technically, this sets an empty breakpoint list for the file, which clears
all breakpoints in that file.

Prerequisites:
- Must be connected to Godot DAP server (call godot_connect first)
- Breakpoint must have been previously set at the specified location

Use this tool:
- When you no longer need a breakpoint
- To disable debugging at a specific location
- To clean up breakpoints after debugging

Note: Due to DAP protocol design, this clears ALL breakpoints in the specified file.
If you want to keep some breakpoints and remove others, you'll need to set
breakpoints again for the lines you want to keep.

Example: Clear breakpoint in player script
godot_clear_breakpoint(file="res://scripts/player.gd")

Example: Clear with absolute path
godot_clear_breakpoint(file="/Users/dev/myproject/player.gd")`,

		Parameters: []mcp.Parameter{
			{
				Name:        "file",
				Type:        "string",
				Required:    true,
				Description: "Path to GDScript file (absolute or res:// path)",
			},
		},

		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Get active session
			session, err := GetSession()
			if err != nil {
				return nil, fmt.Errorf("%w\n\nPlease call godot_connect first to establish a DAP session", err)
			}

			// Get file parameter
			file, ok := params["file"].(string)
			if !ok || file == "" {
				return nil, fmt.Errorf("file parameter is required and must be a non-empty string")
			}

			// Resolve file path
			normalizedFile, err := resolveGodotPath(file, session.GetProjectRoot())
			if err != nil {
				return nil, err
			}

			// Send setBreakpoints with empty list to clear all breakpoints
			ctx, cancel := dap.WithCommandTimeout(context.Background())
			defer cancel()

			client := session.GetClient()
			_, err = client.SetBreakpoints(ctx, normalizedFile, []int{})
			if err != nil {
				return nil, fmt.Errorf("failed to clear breakpoints: %w", err)
			}

			return map[string]interface{}{
				"status":  "cleared",
				"message": fmt.Sprintf("All breakpoints cleared in %s", file),
				"file":    file,
			}, nil
		},
	})
}
