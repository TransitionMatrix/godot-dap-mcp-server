package tools

import (
	"context"
	"fmt"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/dap"
	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
)

// RegisterExecutionTools registers execution control tools (continue, step-over, step-into)
func RegisterExecutionTools(server *mcp.Server) {
	// godot_continue - Resume execution
	server.RegisterTool(mcp.Tool{
		Name: "godot_continue",
		Description: `Resume execution of the paused game.

This tool continues execution after hitting a breakpoint or pausing. The game will
run until it hits another breakpoint, pauses, or exits.

Prerequisites:
- Must be connected to Godot DAP server (call godot_connect first)
- Game must be paused (at breakpoint or manually paused)

Use this tool:
- After inspecting variables at a breakpoint
- To resume execution after stepping
- When you're done debugging the current pause

The tool will wait for the continue operation to complete. You'll receive a
"stopped" event when the game hits the next breakpoint.

Example: Continue execution
godot_continue()

Example: Continue specific thread (Godot uses thread ID 1)
godot_continue(thread_id=1)`,

		Parameters: []mcp.Parameter{
			{
				Name:        "thread_id",
				Type:        "number",
				Required:    false,
				Default:     1,
				Description: "Thread ID to continue (default: 1, Godot typically uses single thread)",
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

			// Send continue request
			ctx, cancel := dap.WithCommandTimeout(context.Background())
			defer cancel()

			client := session.GetClient()
			resp, err := client.Continue(ctx, threadId)
			if err != nil {
				return nil, fmt.Errorf("failed to continue execution: %w", err)
			}

			return map[string]interface{}{
				"status":           "continued",
				"message":          "Execution resumed",
				"all_threads_continued": resp.Body.AllThreadsContinued,
			}, nil
		},
	})

	// godot_step_over - Step over current line
	server.RegisterTool(mcp.Tool{
		Name: "godot_step_over",
		Description: `Step over the current line of code.

This tool executes the current line and pauses at the next line in the same function.
If the current line calls a function, it will execute the entire function and pause
at the next line after the function call.

Prerequisites:
- Must be connected to Godot DAP server (call godot_connect first)
- Game must be paused (at breakpoint or manually paused)

Use this tool:
- To step through code line by line
- When you want to skip over function calls
- To quickly navigate through a function's logic

The game will pause at the next line of code in the current function.

Example: Step over current line
godot_step_over()

Example: Step over with specific thread ID
godot_step_over(thread_id=1)`,

		Parameters: []mcp.Parameter{
			{
				Name:        "thread_id",
				Type:        "number",
				Required:    false,
				Default:     1,
				Description: "Thread ID to step (default: 1, Godot typically uses single thread)",
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

			// Send next (step over) request
			ctx, cancel := dap.WithCommandTimeout(context.Background())
			defer cancel()

			client := session.GetClient()
			_, err = client.Next(ctx, threadId)
			if err != nil {
				return nil, fmt.Errorf("failed to step over: %w", err)
			}

			return map[string]interface{}{
				"status":  "stepped_over",
				"message": "Stepped over current line",
			}, nil
		},
	})

	// godot_step_into - Step into function call
	server.RegisterTool(mcp.Tool{
		Name: "godot_step_into",
		Description: `Step into the function call at the current line.

This tool steps into the function being called on the current line. If the current
line doesn't call a function, it behaves the same as step_over.

Prerequisites:
- Must be connected to Godot DAP server (call godot_connect first)
- Game must be paused (at breakpoint or manually paused)

Use this tool:
- To investigate the implementation of a function
- When you want to debug inside a function call
- To trace execution into called functions

The game will pause at the first line of the called function.

Note: If the current line calls a built-in function or C++ function (not GDScript),
this will behave like step_over since you can't step into native code.

Example: Step into function
godot_step_into()

Example: Step into with specific thread ID
godot_step_into(thread_id=1)`,

		Parameters: []mcp.Parameter{
			{
				Name:        "thread_id",
				Type:        "number",
				Required:    false,
				Default:     1,
				Description: "Thread ID to step (default: 1, Godot typically uses single thread)",
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

			// Send stepIn request
			ctx, cancel := dap.WithCommandTimeout(context.Background())
			defer cancel()

			client := session.GetClient()
			_, err = client.StepIn(ctx, threadId)
			if err != nil {
				return nil, fmt.Errorf("failed to step into: %w", err)
			}

			return map[string]interface{}{
				"status":  "stepped_in",
				"message": "Stepped into function",
			}, nil
		},
	})
}
