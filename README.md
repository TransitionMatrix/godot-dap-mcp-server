# Godot DAP MCP Server

**Status**: 🚧 Active Development - Phase 1 Complete ✅

An MCP (Model Context Protocol) server that enables AI agents to perform interactive runtime debugging of Godot games via the Debug Adapter Protocol (DAP).

## Current Capabilities

**Phase 1 Complete** (2025-11-06):
- ✅ Working MCP server with stdio communication
- ✅ Tool registration system
- ✅ Test tool (`godot_ping`) verified with Claude Code
- ✅ 16-test suite covering critical paths
- ✅ Binary tested and working with MCP clients

**Coming Soon** (Phase 2-8):
- DAP client for Godot connection (Phase 2)
- Breakpoint management (Phase 3)
- Execution control (continue, stepping) (Phase 3)
- Runtime inspection (stack, variables, evaluation) (Phase 4)
- Scene launching (Phase 5)

## Planned Features

- **Interactive Debugging**: Set breakpoints, step through code, inspect variables
- **Scene Launching**: Launch Godot scenes programmatically with debugging enabled
- **Runtime Inspection**: Examine call stacks, evaluate GDScript expressions, inspect game state
- **AI-Optimized**: Tool descriptions and parameters designed for LLM understanding
- **Zero Dependencies**: Single Go binary, no runtime dependencies

## Architecture

```
MCP Client (Claude Code) → stdio → MCP Server → TCP/DAP → Godot Editor → Game Instance
```

This server bridges Claude Code (or any MCP client) to Godot's built-in DAP server, enabling AI-assisted debugging workflows.

## Installation

🚧 Not yet available for end users - Godot debugging features coming in Phase 2+.

**For Testing/Development**: You can test the Phase 1 MCP server:

```bash
# Build from source
go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go

# Register with Claude Code (for testing)
mkdir -p ~/.claude/mcp-servers/godot-dap/
cp godot-dap-mcp-server ~/.claude/mcp-servers/godot-dap/
claude mcp add godot-dap ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server

# Test the godot_ping tool in Claude Code
```

See [docs/PLAN.md](docs/PLAN.md) for development roadmap.

## Development

See [CLAUDE.md](CLAUDE.md) for complete development setup and architecture details.

**Development Progress**: Phase 1 complete ✅ - See [docs/PLAN.md](docs/PLAN.md) for status

### Build and Test

```bash
# Build
go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go

# Run unit tests (16 tests)
go test ./...

# Run with verbose output
go test -v ./...

# Test with stdio directly
./scripts/test-stdio.sh
```

### Project Structure

```
internal/mcp/        # MCP protocol implementation (Phase 1 ✅)
internal/dap/        # DAP client (Phase 2 - coming soon)
internal/tools/      # Godot-specific tools
cmd/                 # Entry point
```

## License

MIT License - see [LICENSE](LICENSE)
