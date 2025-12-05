package tools

import (
	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
)

// RegisterAll registers all available tools with the MCP server
// This is the central place where all tools are registered
func RegisterAll(server *mcp.Server) {
	// Register test tool
	RegisterPingTool(server)

	// Phase 3: Core debugging tools
	RegisterConnectionTools(server)
	RegisterExecutionTools(server)
	RegisterBreakpointTools(server)

	// Phase 4: Runtime inspection tools
	RegisterInspectionTools(server)

	// Phase 5: Launch tools
	RegisterLaunchTools(server)
	RegisterAttachTools(server)

	// Phase 6: Advanced debugging tools
	RegisterAdvancedTools(server)
}
