package tools

import (
	"context"
	"fmt"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/dap"
	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
)

// Global DAP session (single session design)
// All debugging tools share this session
var globalSession *dap.Session

// GetSession returns the global DAP session
// Returns error if no session is active
func GetSession() (*dap.Session, error) {
	if globalSession == nil {
		return nil, fmt.Errorf("not connected to Godot DAP server")
	}
	return globalSession, nil
}

// RegisterConnectionTools registers godot_connect and godot_disconnect tools
func RegisterConnectionTools(server *mcp.Server) {
	// godot_connect - Establish DAP connection to Godot
	server.RegisterTool(mcp.Tool{
		Name: "godot_connect",
		Description: `Connect to Godot's Debug Adapter Protocol (DAP) server.

This tool establishes a connection to the Godot editor's DAP server, which must be
running and have the DAP server enabled in editor settings.

Prerequisites:
1. Godot editor must be running
2. DAP server must be enabled in: Editor → Editor Settings → Network → Debug Adapter
3. DAP server must be listening on the specified port (default: 6006)

After connecting, the DAP session is initialized and configured, making it ready
for debugging operations (breakpoints, stepping, inspection).

Use this tool:
- Before setting breakpoints or launching scenes
- After starting the Godot editor
- When you want to begin a debugging session

Example: Connect to default port
godot_connect()

Example: Connect to custom port
godot_connect(port=6007)`,

		Parameters: []mcp.Parameter{
			{
				Name:        "port",
				Type:        "number",
				Required:    false,
				Default:     6006,
				Description: "DAP server port number (default: 6006)",
			},
		},

		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Check if already connected
			if globalSession != nil && globalSession.IsReady() {
				return map[string]interface{}{
					"status":  "already_connected",
					"message": "Already connected to Godot DAP server",
				}, nil
			}

			// Get port parameter
			port := 6006 // default
			if p, ok := params["port"].(float64); ok {
				port = int(p)
			}

			// Create new session
			session := dap.NewSession("localhost", port)

			// Connect with timeout
			ctx, cancel := dap.WithConnectTimeout(context.Background())
			defer cancel()

			if err := session.Connect(ctx); err != nil {
				return nil, fmt.Errorf(`Failed to connect to Godot DAP server at localhost:%d

Possible causes:
1. Godot editor is not running
2. DAP server is not enabled in editor settings
3. DAP server is using a different port

Solutions:
1. Launch Godot editor
2. Enable DAP in Editor → Editor Settings → Network → Debug Adapter
3. Check port setting (default: 6006)

Error: %w`, port, err)
			}

			// Initialize the session
			if err := session.Initialize(ctx); err != nil {
				session.Close()
				return nil, fmt.Errorf("failed to initialize DAP session: %w", err)
			}

			// Complete the handshake with configurationDone
			if err := session.ConfigurationDone(ctx); err != nil {
				session.Close()
				return nil, fmt.Errorf("failed to complete DAP configuration: %w", err)
			}

			// Session is now ready for debugging
			globalSession = session

			return map[string]interface{}{
				"status":  "connected",
				"message": fmt.Sprintf("Connected to Godot DAP server at localhost:%d", port),
				"state":   session.GetState().String(),
			}, nil
		},
	})

	// godot_disconnect - Close DAP connection
	server.RegisterTool(mcp.Tool{
		Name: "godot_disconnect",
		Description: `Disconnect from the Godot DAP server.

This tool closes the active DAP session and cleans up the connection.

Use this tool:
- When finished debugging
- Before shutting down the MCP server
- To reset the connection state

After disconnecting, you'll need to call godot_connect again before
performing any debugging operations.

Example: Disconnect from DAP server
godot_disconnect()`,

		Parameters: []mcp.Parameter{},

		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Check if connected
			if globalSession == nil {
				return map[string]interface{}{
					"status":  "not_connected",
					"message": "Not currently connected to Godot DAP server",
				}, nil
			}

			// Close the session
			if err := globalSession.Close(); err != nil {
				return nil, fmt.Errorf("failed to disconnect: %w", err)
			}

			globalSession = nil

			return map[string]interface{}{
				"status":  "disconnected",
				"message": "Disconnected from Godot DAP server",
			}, nil
		},
	})
}
