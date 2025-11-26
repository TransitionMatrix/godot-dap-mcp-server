# Godot DAP MCP Server

**Status**: 🚧 Active Development - Phase 8 Complete ✅ | Documentation ⏳

An MCP (Model Context Protocol) server that enables AI agents to perform interactive runtime debugging of Godot games via the Debug Adapter Protocol (DAP).

## Current Capabilities

**Phase 1-3: Core Debugging** (Complete):
- ✅ Connect to Godot editor via DAP
- ✅ Set, verify, and clear breakpoints
- ✅ Control execution (continue, step over, step into)
- ✅ Session management

**Phase 4: Runtime Inspection** (Complete):
- ✅ Get stack traces with source locations
- ✅ Inspect variable scopes (Locals, Members, Globals)
- ✅ Browse complex objects (Arrays, Dictionaries, Nodes)
- ✅ Evaluate GDScript expressions at runtime
- ✅ Godot-specific type formatting (Vector2, Color, etc.)

**Phase 5: Scene Launching** (Complete):
- ✅ Launch main scene, specific scene, or current editor scene
- ✅ Configure debug options (profiling, collisions, navigation)
- ✅ Validates project paths

**Phase 6-8: Advanced & Polish** (Complete):
- ✅ `godot_pause`: Pause running game
- ✅ `godot_set_variable`: *Available but currently disabled due to upstream Godot limitation*
- ✅ **Robust Error Handling**: Problem/Context/Solution error messages
- ✅ **Path Resolution**: Auto-converts `res://` paths to absolute paths
- ✅ **Timeout Protection**: Prevents hangs from unresponsive DAP server
- ✅ **Event-Driven Architecture**: Handles complex interleaved DAP events

## Planned Features

- **Upstream Improvements**:
  - `stepOut` support (PR Submitted: [godotengine/godot#112875](https://github.com/godotengine/godot/pull/112875))
  - `setVariable` support (Planned PR)
- **Documentation**: Complete tool reference and examples

## Architecture

```
MCP Client (Claude Code) → stdio → MCP Server → TCP/DAP → Godot Editor → Game Instance
```

This server bridges Claude Code (or any MCP client) to Godot's built-in DAP server, enabling AI-assisted debugging workflows.

**Architecture:**
- **MCP Layer** (`internal/mcp/`): stdio-based JSONRPC 2.0 communication
- **DAP Client** (`internal/dap/`): Event-driven TCP client for Godot's DAP server
- **Tool Layer** (`internal/tools/`): Godot-specific MCP tools with error handling & path resolution

## Documentation

Comprehensive documentation is available in the `docs/` directory:

- **[docs/PLAN.md](docs/PLAN.md)** - Project status and roadmap
- **[docs/TOOLS.md](docs/TOOLS.md)** - Complete tool reference (Coming Soon)
- **[docs/EXAMPLES.md](docs/EXAMPLES.md)** - Usage examples (Coming Soon)
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - System design
- **[docs/TESTING.md](docs/TESTING.md)** - Testing guide

## Installation

**For Testing/Development**:

```bash
# Build from source
go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go

# Run all tests
go test ./...

# Run integration tests (requires Godot 4.2.2+)
./scripts/automated-integration-test.sh
```

## Quick Start

1. **Start Godot Editor** with your project.
2. **Run MCP Server** (via Claude Code or manually).
3. **Connect**: `godot_connect(project="/path/to/your/project")`
4. **Set Breakpoint**: `godot_set_breakpoint(file="res://player.gd", line=10)`
5. **Launch**: `godot_launch_main_scene(project="/path/to/your/project")`
6. **Debug**: Use `godot_step_over`, `godot_get_stack_trace`, `godot_evaluate`, etc.

## License

MIT License - see [LICENSE](LICENSE)
