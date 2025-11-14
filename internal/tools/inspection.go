package tools

import (
	"context"
	"fmt"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/dap"
	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
)

// RegisterInspectionTools registers all runtime inspection MCP tools.
func RegisterInspectionTools(server *mcp.Server) {
	// godot_get_threads - Get list of active threads
	server.RegisterTool(mcp.Tool{
		Name: "godot_get_threads",
		Description: `Get the list of active threads in the debugged game.

This tool returns information about all threads in the running game. Godot games
typically run on a single thread (ID: 1, Name: "Main").

Prerequisites:
- Must be connected to Godot DAP server (call godot_connect first)
- Game must be running (launched or attached)

Use this tool:
- To get the thread ID for stack trace requests
- To verify the game is running and responsive
- Before inspecting variables or evaluating expressions

The response includes thread ID and name for each active thread.

Example: Get all threads
godot_get_threads()`,

		Parameters: []mcp.Parameter{},

		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Get active session
			session, err := GetSession()
			if err != nil {
				return nil, fmt.Errorf("%w\n\nPlease call godot_connect first to establish a DAP session", err)
			}

			// Request threads
			ctx, cancel := dap.WithCommandTimeout(context.Background())
			defer cancel()

			client := session.GetClient()
			resp, err := client.Threads(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get threads: %w", err)
			}

			// Format response
			threads := make([]map[string]interface{}, len(resp.Body.Threads))
			for i, thread := range resp.Body.Threads {
				threads[i] = map[string]interface{}{
					"id":   thread.Id,
					"name": thread.Name,
				}
			}

			return map[string]interface{}{
				"status":  "success",
				"threads": threads,
				"count":   len(threads),
			}, nil
		},
	})

	// godot_get_stack_trace - Get call stack
	server.RegisterTool(mcp.Tool{
		Name: "godot_get_stack_trace",
		Description: `Get the call stack for the paused game.

This tool returns the current call stack showing the sequence of function calls
that led to the current execution point. Each frame includes the function name,
source file, and line number.

Prerequisites:
- Must be connected to Godot DAP server (call godot_connect first)
- Game must be paused (at breakpoint or manually paused)

Use this tool:
- To understand the execution path that led to a breakpoint
- To see which function called the current function
- To get frame IDs for inspecting variables in different stack frames

The response includes frames from most recent (index 0) to oldest.

Example: Get full stack trace
godot_get_stack_trace(thread_id=1)

Example: Get top 5 frames only
godot_get_stack_trace(thread_id=1, max_frames=5)`,

		Parameters: []mcp.Parameter{
			{
				Name:        "thread_id",
				Type:        "number",
				Required:    false,
				Default:     1,
				Description: "Thread ID to get stack trace for (default: 1)",
			},
			{
				Name:        "max_frames",
				Type:        "number",
				Required:    false,
				Default:     20,
				Description: "Maximum number of stack frames to return (default: 20)",
			},
		},

		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Get active session
			session, err := GetSession()
			if err != nil {
				return nil, fmt.Errorf("%w\n\nPlease call godot_connect first to establish a DAP session", err)
			}

			// Get parameters
			threadId := 1
			if tid, ok := params["thread_id"].(float64); ok {
				threadId = int(tid)
			}

			maxFrames := 20
			if max, ok := params["max_frames"].(float64); ok {
				maxFrames = int(max)
			}

			// Request stack trace
			ctx, cancel := dap.WithCommandTimeout(context.Background())
			defer cancel()

			client := session.GetClient()
			resp, err := client.StackTrace(ctx, threadId, 0, maxFrames)
			if err != nil {
				return nil, fmt.Errorf("failed to get stack trace: %w", err)
			}

			// Format stack frames
			frames := make([]map[string]interface{}, len(resp.Body.StackFrames))
			for i, frame := range resp.Body.StackFrames {
				frameData := map[string]interface{}{
					"id":     frame.Id,
					"name":   frame.Name,
					"line":   frame.Line,
					"column": frame.Column,
				}

				// Add source file if available
				if frame.Source != nil {
					frameData["source"] = map[string]interface{}{
						"name": frame.Source.Name,
						"path": frame.Source.Path,
					}
				}

				frames[i] = frameData
			}

			return map[string]interface{}{
				"status":       "success",
				"frames":       frames,
				"total_frames": resp.Body.TotalFrames,
			}, nil
		},
	})

	// godot_get_scopes - Get variable scopes for stack frame
	server.RegisterTool(mcp.Tool{
		Name: "godot_get_scopes",
		Description: `Get variable scopes for a stack frame.

This tool returns the available variable scopes (Locals, Members, Globals) for
a specific stack frame. Each scope has a variablesReference that can be used
with godot_get_variables to retrieve the actual variables.

Prerequisites:
- Must be connected to Godot DAP server (call godot_connect first)
- Game must be paused (at breakpoint or manually paused)
- Must have a valid frame ID (from godot_get_stack_trace)

Use this tool:
- To discover what variable scopes are available in a stack frame
- To get variablesReference IDs for retrieving variables
- Before calling godot_get_variables

Godot always returns three scopes:
- Locals: Function-local variables
- Members: Instance/class member variables (if in a method)
- Globals: Global variables and autoloads

Example: Get scopes for top frame
godot_get_scopes(frame_id=1)`,

		Parameters: []mcp.Parameter{
			{
				Name:        "frame_id",
				Type:        "number",
				Required:    true,
				Description: "Stack frame ID (from godot_get_stack_trace)",
			},
		},

		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Get active session
			session, err := GetSession()
			if err != nil {
				return nil, fmt.Errorf("%w\n\nPlease call godot_connect first to establish a DAP session", err)
			}

			// Get frame ID parameter
			frameIdFloat, ok := params["frame_id"].(float64)
			if !ok {
				return nil, fmt.Errorf("frame_id is required and must be a number")
			}
			frameId := int(frameIdFloat)

			// Request scopes
			ctx, cancel := dap.WithCommandTimeout(context.Background())
			defer cancel()

			client := session.GetClient()
			resp, err := client.Scopes(ctx, frameId)
			if err != nil {
				return nil, fmt.Errorf("failed to get scopes: %w", err)
			}

			// Format scopes
			scopes := make([]map[string]interface{}, len(resp.Body.Scopes))
			for i, scope := range resp.Body.Scopes {
				scopeData := map[string]interface{}{
					"name":                scope.Name,
					"variables_reference": scope.VariablesReference,
					"expensive":           scope.Expensive,
				}

				if scope.PresentationHint != "" {
					scopeData["hint"] = scope.PresentationHint
				}

				scopes[i] = scopeData
			}

			return map[string]interface{}{
				"status": "success",
				"scopes": scopes,
				"count":  len(scopes),
			}, nil
		},
	})

	// godot_get_variables - Get variables in scope
	server.RegisterTool(mcp.Tool{
		Name: "godot_get_variables",
		Description: `Get variables in a scope or expand a complex variable.

This tool retrieves variables using a variablesReference obtained from
godot_get_scopes or from a variable with variablesReference > 0.

Prerequisites:
- Must be connected to Godot DAP server (call godot_connect first)
- Game must be paused (at breakpoint or manually paused)
- Must have a valid variablesReference (from godot_get_scopes or another variable)

Use this tool:
- To view variable values in a scope (Locals, Members, Globals)
- To expand complex objects (Vector2, Node, Array, Dictionary)
- To inspect object properties and array elements
- To navigate the scene tree through Node objects

Variables with variablesReference > 0 can be expanded by calling this tool
again with their variablesReference.

Scene Tree Navigation:
To navigate the scene tree and inspect nodes:
1. Get Members scope (contains 'self' - the current Node)
2. Expand 'self' to see Node properties
3. Look for properties with 'Node/' prefix (name, parent, children)
4. Expand 'Node/children' array to see child nodes
5. Expand each child to inspect its properties

When expanding a Node object, properties are categorized:
- Members/* - Script member variables (if script attached)
- Constants/* - Script constants (if script attached)
- Node/* - Node-specific properties (name, parent, children, scene path)
- Transform2D/* - Position, rotation, scale (for 2D nodes)
- Other categories based on node type (CanvasItem, Control, etc.)

Example: Get all local variables
godot_get_variables(variables_reference=1000)

Example: Expand a Vector2 variable
godot_get_variables(variables_reference=2000)

Example: Scene tree navigation workflow
1. godot_get_scopes(frame_id=0)
   → Returns scopes, Members scope has variables_reference=1001
2. godot_get_variables(variables_reference=1001)
   → Returns 'self' with variables_reference=2000
3. godot_get_variables(variables_reference=2000)
   → Returns Node properties including 'Node/children' with variables_reference=2050
4. godot_get_variables(variables_reference=2050)
   → Returns array of child nodes, each expandable`,

		Parameters: []mcp.Parameter{
			{
				Name:        "variables_reference",
				Type:        "number",
				Required:    true,
				Description: "Variables reference ID (from godot_get_scopes or a complex variable)",
			},
		},

		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Get active session
			session, err := GetSession()
			if err != nil {
				return nil, fmt.Errorf("%w\n\nPlease call godot_connect first to establish a DAP session", err)
			}

			// Get variables reference parameter
			varRefFloat, ok := params["variables_reference"].(float64)
			if !ok {
				return nil, fmt.Errorf("variables_reference is required and must be a number")
			}
			varRef := int(varRefFloat)

			// Request variables
			ctx, cancel := dap.WithCommandTimeout(context.Background())
			defer cancel()

			client := session.GetClient()
			resp, err := client.Variables(ctx, varRef)
			if err != nil {
				return nil, fmt.Errorf("failed to get variables: %w", err)
			}

				// Format variables with Godot-specific formatting
			variables := formatVariableList(resp.Body.Variables)

			return map[string]interface{}{
				"status":    "success",
				"variables": variables,
				"count":     len(variables),
			}, nil
		},
	})

	// godot_evaluate - Evaluate GDScript expression
	server.RegisterTool(mcp.Tool{
		Name: "godot_evaluate",
		Description: `Evaluate a GDScript expression in the current debugging context.

This tool evaluates arbitrary GDScript expressions and returns the result.
The expression is evaluated in the context of the specified stack frame,
so it has access to local variables, member variables, and global variables.

Prerequisites:
- Must be connected to Godot DAP server (call godot_connect first)
- Game must be paused (at breakpoint or manually paused)
- Must have a valid frame ID (from godot_get_stack_trace)

Use this tool:
- To compute values based on current variables (e.g., "player.health * 2")
- To test conditions (e.g., "position.x > 100")
- To access object properties not visible in variables
- To call getter functions

WARNING: The expression CAN modify game state. For example, evaluating
"player.health = 0" will actually change the player's health. Use
godot_set_variable for intentional modifications.

Example: Evaluate simple expression
godot_evaluate(expression="player.health * 2", frame_id=1)

Example: Check condition
godot_evaluate(expression="position.x > 100 and velocity.y < 0", frame_id=1)

Example: Access nested property
godot_evaluate(expression="$Player/Sprite.texture.get_size()", frame_id=1)`,

		Parameters: []mcp.Parameter{
			{
				Name:        "expression",
				Type:        "string",
				Required:    true,
				Description: "GDScript expression to evaluate",
			},
			{
				Name:        "frame_id",
				Type:        "number",
				Required:    false,
				Default:     0,
				Description: "Stack frame ID for evaluation context (default: 0 = top frame)",
			},
			{
				Name:        "context",
				Type:        "string",
				Required:    false,
				Default:     "repl",
				Description: "Evaluation context: 'watch', 'repl', or 'hover' (default: 'repl')",
			},
		},

		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Get active session
			session, err := GetSession()
			if err != nil {
				return nil, fmt.Errorf("%w\n\nPlease call godot_connect first to establish a DAP session", err)
			}

			// Get expression parameter
			expression, ok := params["expression"].(string)
			if !ok || expression == "" {
				return nil, fmt.Errorf("expression is required and must be a non-empty string")
			}

			// Get optional parameters
			frameId := 0
			if fid, ok := params["frame_id"].(float64); ok {
				frameId = int(fid)
			}

			evalContext := "repl"
			if ctx, ok := params["context"].(string); ok {
				evalContext = ctx
			}

			// Evaluate expression
			ctx, cancel := dap.WithCommandTimeout(context.Background())
			defer cancel()

			client := session.GetClient()
			resp, err := client.Evaluate(ctx, expression, frameId, evalContext)
			if err != nil {
				return nil, fmt.Errorf("failed to evaluate expression: %w", err)
			}

			// Format response with Godot-specific formatting
			result := map[string]interface{}{
				"status": "success",
				"result": resp.Body.Result,
				"type":   resp.Body.Type,
			}

			// Add formatted version if it's a Godot type
			if formatted := formatGodotType(resp.Body.Type, resp.Body.Result); formatted != "" {
				result["formatted"] = formatted
			}

			// Mark if result is expandable
			if resp.Body.VariablesReference > 0 {
				result["expandable"] = true
				result["variables_reference"] = resp.Body.VariablesReference
			}

			return result, nil
		},
	})
}
