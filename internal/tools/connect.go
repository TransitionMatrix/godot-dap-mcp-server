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
		return nil, ErrNotConnected()
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

Example: Connect with project path (enables res:// path resolution)
godot_connect(project="/path/to/my/project")`,

		Parameters: []mcp.Parameter{
			{
				Name:        "port",
				Type:        "number",
				Required:    false,
				Default:     6006,
				Description: "DAP server port number (default: 6006)",
			},
			{
				Name:        "project",
				Type:        "string",
				Required:    false,
				Description: "Absolute path to project root (optional, enables res:// path resolution)",
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

			// Set project root if provided
			if proj, ok := params["project"].(string); ok && proj != "" {
				session.SetProjectRoot(proj)
			}

			// Connect with timeout
			ctx, cancel := dap.WithConnectTimeout(context.Background())
			defer cancel()

			if err := session.Connect(ctx); err != nil {
				return nil, FormatError(
					"Failed to connect to Godot DAP server",
					fmt.Sprintf("localhost:%d", port),
					[]string{
						"Launch Godot editor",
						"Enable DAP in Editor → Editor Settings → Network → Debug Adapter",
						fmt.Sprintf("Check port setting (default: 6006, tried: %d)", port),
					},
					err,
				)
			}

			// Initialize the session
			if err := session.Initialize(ctx); err != nil {
				session.Close()
				return nil, fmt.Errorf("failed to initialize DAP session: %w", err)
			}

			// Note: We do NOT send configurationDone here.
			// It must be sent AFTER the launch request.
			// The session remains in 'initialized' state until a launch tool is called.

			// Session is now ready for debugging
			globalSession = session

			return map[string]interface{}{
				"status":  "connected",
				"message": fmt.Sprintf("Connected to Godot DAP server at localhost:%d. Ready to launch.", port),
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
