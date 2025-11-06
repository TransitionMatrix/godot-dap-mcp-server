# Godot DAP MCP Server

**Status**: 🚧 Under Development

An MCP (Model Context Protocol) server that enables AI agents to perform interactive runtime debugging of Godot games via the Debug Adapter Protocol (DAP).

## Features (Planned)

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

🚧 Not yet available - see [docs/PLAN.md](docs/PLAN.md) for development roadmap.

## Development

See [CLAUDE.md](CLAUDE.md) for development setup and architecture details.

```bash
# Build
go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go

# Run tests
go test ./...
```

## License

MIT License - see [LICENSE](LICENSE)
