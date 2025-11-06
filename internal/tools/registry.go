package tools

import (
	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
)

// RegisterAll registers all available tools with the MCP server
// This is the central place where all tools are registered
func RegisterAll(server *mcp.Server) {
	// Register test tool
	RegisterPingTool(server)

	// Future tools will be registered here as we implement them:
	// RegisterConnectTool(server, dapClient)
	// RegisterLaunchTools(server, dapClient)
	// RegisterBreakpointTools(server, dapClient)
	// RegisterExecutionTools(server, dapClient)
	// RegisterInspectionTools(server, dapClient)
}
