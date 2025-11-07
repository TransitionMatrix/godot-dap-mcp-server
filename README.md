# Godot DAP MCP Server

**Status**: 🚧 Active Development - Phase 2 Complete ✅ | Phase 3 In Progress ⏳

An MCP (Model Context Protocol) server that enables AI agents to perform interactive runtime debugging of Godot games via the Debug Adapter Protocol (DAP).

## Current Capabilities

**Phase 1 Complete** (2025-11-06):
- ✅ Working MCP server with stdio communication
- ✅ Tool registration system
- ✅ Test tool (`godot_ping`) verified with Claude Code
- ✅ 16 unit tests covering critical paths
- ✅ Binary tested and working with MCP clients

**Phase 2 Complete** (2025-11-06):
- ✅ DAP client with TCP connection to Godot editor
- ✅ Event filtering (handles async DAP events mixed with responses)
- ✅ Timeout protection (prevents hangs from unresponsive DAP server)
- ✅ Session lifecycle management
- ✅ Godot-specific launch configurations
- ✅ 12 additional unit tests (28 total)

**Coming Soon** (Phase 3-8):
- Core debugging tools: connect, breakpoints, stepping (Phase 3 - In Progress)
- Runtime inspection (stack, variables, evaluation) (Phase 4)
- Scene launching tools (Phase 5)
- Advanced features and polish (Phase 6-8)

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

**Three-Layer Architecture:**
- **MCP Layer** (`internal/mcp/`): stdio-based JSONRPC 2.0 communication
- **DAP Client Layer** (`internal/dap/`): TCP connection to Godot's DAP server ✅
- **Tool Layer** (`internal/tools/`): Godot-specific MCP tools

## Documentation

Comprehensive documentation is available in the `docs/` directory:

**Core Documentation:**
- **[docs/PLAN.md](docs/PLAN.md)** - Implementation phases and project timeline
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - System design and critical patterns
- **[docs/IMPLEMENTATION_GUIDE.md](docs/IMPLEMENTATION_GUIDE.md)** - Component specifications
- **[docs/TESTING.md](docs/TESTING.md)** - Testing strategies and procedures
- **[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)** - Build and distribution guide

**Reference Documentation:**
- **[docs/reference/CONVENTIONS.md](docs/reference/CONVENTIONS.md)** - Coding standards and naming conventions
- **[docs/reference/DAP_PROTOCOL.md](docs/reference/DAP_PROTOCOL.md)** - Godot DAP protocol details
- **[docs/reference/GODOT_DAP_FAQ.md](docs/reference/GODOT_DAP_FAQ.md)** - Common questions and troubleshooting
- **[docs/reference/GODOT_SOURCE_ANALYSIS.md](docs/reference/GODOT_SOURCE_ANALYSIS.md)** - Findings from Godot source analysis

**For Contributors:** See [CLAUDE.md](CLAUDE.md) for AI-assisted development guidance, including Serena memory navigation and workflow tools.

## Installation

🚧 **Not yet available for end users** - Core debugging tools (Phase 3) are in development.

**Current Status**: The MCP server and DAP client infrastructure are complete. Debugging tools (connect, breakpoints, stepping) are coming in Phase 3.

**For Testing/Development**:

```bash
# Build from source
go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go

# Run tests
go test ./...  # 28 tests covering MCP and DAP layers

# Cross-platform builds
./scripts/build.sh
```

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for detailed build instructions and [docs/PLAN.md](docs/PLAN.md) for the development roadmap.

## Development

See [CLAUDE.md](CLAUDE.md) for complete development setup, architecture details, and AI-assisted development workflows.

**Development Progress**: Phase 2 complete ✅ | Phase 3 in progress ⏳ - See [docs/PLAN.md](docs/PLAN.md) for detailed status

### Quick Start

```bash
# Build
go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go

# Run all tests (28 tests: 16 MCP + 12 DAP)
go test ./...

# Run with verbose output
go test -v ./...

# Format and lint
go fmt ./... && go vet ./...
```

### Security & Workflows

**Secret Detection** (Pre-commit hook):
```bash
# Install gitleaks for secret detection (recommended for public repo)
brew install gitleaks

# Enable pre-commit hook (includes gitleaks + memory sync reminders)
ln -sf ../../.claude/hooks/pre-commit.sh .git/hooks/pre-commit
```

**Memory Sync Workflow** (For contributors):
- Use `/memory-sync` skill in Claude Code to maintain project memories
- See [CLAUDE.md](CLAUDE.md) for when to sync memories vs documentation

### Project Structure

```
internal/mcp/        # MCP protocol implementation (Phase 1 ✅)
internal/dap/        # DAP client (Phase 2 ✅)
internal/tools/      # Godot-specific MCP tools
cmd/                 # Entry point
docs/                # Comprehensive documentation
.claude/             # Claude Code skills and hooks
```

## License

MIT License - see [LICENSE](LICENSE)
