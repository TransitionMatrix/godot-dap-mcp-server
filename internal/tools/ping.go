package tools

import (
	"fmt"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
)

// RegisterPingTool registers a simple ping test tool
// This tool is used to verify the MCP server is working correctly
func RegisterPingTool(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name: "godot_ping",
		Description: `Test tool that echoes back a message to verify MCP server is working.

This is a diagnostic tool used to test the MCP connection and server functionality.
It simply echoes back the message you provide, or returns "pong" if no message is given.

Use this to:
- Verify the MCP server is running and responsive
- Test the communication channel between client and server
- Confirm tool calling mechanism is working

Example: Test connection
godot_ping(message="Hello from Claude!")`,

		Parameters: []mcp.Parameter{
			{
				Name:        "message",
				Type:        "string",
				Required:    false,
				Default:     "pong",
				Description: "Message to echo back (default: 'pong')",
			},
		},

		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Get message parameter
			message, ok := params["message"].(string)
			if !ok {
				message = "pong"
			}

			// Echo back the message
			return fmt.Sprintf("Echo: %s", message), nil
		},
	})
}
