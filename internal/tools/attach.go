package tools

import (
	"context"
	"fmt"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/dap"
	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
)

// RegisterAttachTools registers the attach tool
func RegisterAttachTools(server *mcp.Server) {
	// godot_attach - Attach to running game
	server.RegisterTool(mcp.Tool{
		Name: "godot_attach",
		Description: `Attach the debugger to an already running Godot game instance.

This tool connects to a game that is already running and waiting for a debugger.
The game must have been started with debugging enabled and configured to connect
to the editor's port (usually 6007).

Prerequisites:
- Must be connected to Godot DAP server (call godot_connect first)
- Must complete DAP configuration handshake
- Game must be running and attempting to connect to the editor

Use this tool:
- When you want to debug a game that was launched externally
- When you want to attach to a game running on a device
- As an alternative to launching the game through the DAP server

Attach Flow:
1. Sends attach request
2. Sends configurationDone
3. Debugger attaches to the running game session

Example: Attach to running game
godot_attach()`,

		Parameters: []mcp.Parameter{}, // No parameters required for attach

		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Get active session
			session, err := GetSession()
			if err != nil {
				return nil, fmt.Errorf("%w\n\nPlease call godot_connect first to establish a DAP session", err)
			}

			// Attach to game
			ctx, cancel := dap.WithCommandTimeout(context.Background())
			defer cancel()

			if _, err := session.AttachGodot(ctx); err != nil {
				return nil, FormatError(
					"Failed to attach to running game",
					"command=attach",
					[]string{
						"Ensure the game is running",
						"Ensure the game was started with --remote-debug",
						"Check that the game is connecting to the correct port (default 6007)",
					},
					err,
				)
			}

			return map[string]interface{}{
				"status":  "attached",
				"message": "Successfully attached to running game",
			}, nil
		},
	})
}
