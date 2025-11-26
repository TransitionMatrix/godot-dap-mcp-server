package tools

import (
	"context"
	"fmt"
	"regexp"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/dap"
	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
)

// RegisterAdvancedTools registers Phase 6 advanced debugging tools
func RegisterAdvancedTools(server *mcp.Server) {
	// godot_pause - Pause execution of running game
	server.RegisterTool(mcp.Tool{
		Name: "godot_pause",
		Description: `Pause execution of the running Godot game.

This tool pauses the game at its current execution point. Use this when you want to:
- Inspect game state mid-execution
- Pause before setting breakpoints to examine current state
- Stop animation/physics to examine variables
- Interrupt running code to investigate behavior

Prerequisites:
- Must be connected to Godot DAP server (call godot_connect first)
- Game must be running (not already paused)

After pausing:
- The game will send a 'stopped' event with reason='pause'
- Use godot_get_stack_trace to see where execution stopped
- Use godot_get_scopes and godot_get_variables to inspect state
- Use godot_continue to resume execution

The pause happens immediately and execution stops at the current line.

Example: Pause running game
godot_pause()

Example: Pause specific thread (Godot uses thread ID 1)
godot_pause(thread_id=1)`,

		Parameters: []mcp.Parameter{
			{
				Name:        "thread_id",
				Type:        "number",
				Required:    false,
				Default:     1,
				Description: "Thread ID to pause (default: 1, Godot typically uses single thread)",
			},
		},

		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Get active session
			session, err := GetSession()
			if err != nil {
				return nil, fmt.Errorf("%w\n\nPlease call godot_connect first to establish a DAP session", err)
			}

			// Get thread ID parameter
			threadId := 1 // default
			if tid, ok := params["thread_id"].(float64); ok {
				threadId = int(tid)
			}

			// Send pause request
			ctx, cancel := dap.WithCommandTimeout(context.Background())
			defer cancel()

			client := session.GetClient()
			_, err = client.Pause(ctx, threadId)
			if err != nil {
				return nil, fmt.Errorf("failed to pause execution: %w\n\nPossible causes:\n1. Game is not running\n2. Game is already paused\n3. Connection was lost\n\nSolutions:\n1. Ensure game is running (not at a breakpoint)\n2. If already paused, use godot_get_stack_trace to inspect state\n3. Check connection with godot_get_threads", err)
			}

			return map[string]interface{}{
				"status":  "paused",
				"message": "Execution paused. Use godot_get_stack_trace to inspect current state, then godot_continue to resume.",
			}, nil
		},
	})

	// godot_set_variable - Modify variable value at runtime
	server.RegisterTool(mcp.Tool{
		Name: "godot_set_variable",
		Description: `Set a variable's value at runtime during debugging.

This tool modifies a variable's value while the game is paused. Use this when you want to:
- Test different values without restarting the game
- Fix game state during debugging
- Inject test data to reproduce specific scenarios
- Change variables to test edge cases

Prerequisites:
- Must be connected to Godot DAP server (call godot_connect first)
- Game must be paused (at breakpoint or after godot_pause)
- Variable must exist in current scope (Locals, Members, or Globals)

Parameters:
- variable_name: Must be a valid GDScript identifier (letters, numbers, underscores only)
  - ✅ Valid: player_health, _internal_var, score
  - ❌ Invalid: player health, health+10, get_node("Player")
- value: New value (will be formatted based on type)
  - Numbers: 100, 3.14
  - Strings: "hello"
  - Booleans: true, false
- frame_id: Stack frame (0 = current frame, get from godot_get_stack_trace)

Security:
- Variable names are strictly validated to prevent code injection
- Only simple variable assignment is supported
- Complex expressions should use godot_evaluate instead

Implementation Note:
Godot's DAP server advertises setVariable support but doesn't actually implement it.
This tool works around the limitation by using evaluate() with an assignment expression.

Example: Set player health
godot_set_variable(variable_name="player_health", value=100, frame_id=0)

Example: Change a string variable
godot_set_variable(variable_name="player_name", value="TestPlayer", frame_id=0)

Example: Toggle a boolean
godot_set_variable(variable_name="debug_mode", value=true, frame_id=0)

Returns: Variable name, new value, and type`,

		Parameters: []mcp.Parameter{
			{
				Name:        "variable_name",
				Type:        "string",
				Required:    true,
				Description: "Name of the variable to modify (must be valid GDScript identifier)",
			},
			{
				Name:        "value",
				Type:        "", // Empty type accepts any value (omitted from schema)
				Required:    true,
				Description: "New value for the variable",
			},
			{
				Name:        "frame_id",
				Type:        "number",
				Required:    false,
				Default:     0,
				Description: "Stack frame ID (default: 0 = top frame)",
			},
		},

		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Get active session
			_, err := GetSession()
			if err != nil {
				return nil, fmt.Errorf("%w\n\nPlease call godot_connect first to establish a DAP session", err)
			}

			// Return explanatory error
			return nil, fmt.Errorf("godot_set_variable is currently unavailable.\n\nAnalysis of Godot 4.x source code confirms that while Godot advertises 'supportsSetVariable', the implementation is missing from the engine's DAP server.\n\nWorkarounds using expression evaluation also fail because GDScript assignments are statements, not expressions.\n\nFuture Work: We plan to submit a Pull Request to the Godot Engine to implement this feature. Until then, variables can only be inspected, not modified.")
		},
	})
}

// isValidVariableName validates that a variable name is a valid GDScript identifier
// Pattern: ^[a-zA-Z_][a-zA-Z0-9_]*$
// This prevents code injection by rejecting expressions with operators, spaces, etc.
func isValidVariableName(name string) bool {
	// Must start with letter or underscore, followed by letters, numbers, or underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
	return matched
}

// formatValueForGDScript formats a value for use in a GDScript expression
func formatValueForGDScript(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Quote strings and escape internal quotes
		return fmt.Sprintf(`"%s"`, escapeString(v))
	case int, int64, float64:
		// Numbers as-is
		return fmt.Sprintf("%v", v)
	case bool:
		// Booleans as true/false
		return fmt.Sprintf("%v", v)
	default:
		// Default: convert to string (may not work for complex types)
		return fmt.Sprintf("%v", v)
	}
}

// escapeString escapes quotes and backslashes in a string for GDScript
func escapeString(s string) string {
	// Replace backslashes first, then quotes
	s = regexp.MustCompile(`\\`).ReplaceAllString(s, `\\`)
	s = regexp.MustCompile(`"`).ReplaceAllString(s, `\"`)
	return s
}
