# Godot DAP MCP Server

**Status**: 🚧 Active Development - Phase 3 Complete ✅ | Phase 4 Next ⏳

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

**Phase 3 Complete** (2025-11-07):
- ✅ 7 core debugging tools: connect, disconnect, breakpoints, stepping
- ✅ Breakpoint management: set, verify, clear breakpoints
- ✅ Execution control: continue, step over, step into
- ✅ Global session management across tool calls
- ✅ 15 additional unit tests (43 total)
- ✅ Fully automated integration tests with Godot subprocess
- ✅ Manual integration test support

**Coming Soon** (Phase 4-8):
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

**Lessons Learned:**
- **[docs/LESSONS_LEARNED_PHASE_3.md](docs/LESSONS_LEARNED_PHASE_3.md)** - Integration testing debugging journey

**For Contributors:**
- See [CLAUDE.md](CLAUDE.md) for AI-assisted development guidance, including Serena memory navigation and workflow tools
- See [docs/DOCUMENTATION_WORKFLOW.md](docs/DOCUMENTATION_WORKFLOW.md) for how to document patterns and debugging insights

## Installation

🚧 **Early Access** - Core debugging tools (Phase 3) are complete and working!

**Current Status**: You can now connect to Godot, set breakpoints, and control execution (continue, step over, step into). Runtime inspection coming in Phase 4.

**For Testing/Development**:

```bash
# Build from source
go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go

# Run all tests
go test ./...  # 43 tests covering MCP, DAP, and tools

# Run integration tests (requires Godot 4.2.2+)
./scripts/automated-integration-test.sh  # Fully automated
./scripts/integration-test.sh            # Manual with running editor

# Cross-platform builds
./scripts/build.sh
```

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for detailed build instructions and [docs/PLAN.md](docs/PLAN.md) for the development roadmap.

## Development

See [CLAUDE.md](CLAUDE.md) for complete development setup, architecture details, and AI-assisted development workflows.

**Development Progress**: Phase 3 complete ✅ | Phase 4 next ⏳ - See [docs/PLAN.md](docs/PLAN.md) for detailed status

### Quick Start

```bash
# Build
go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go

# Run all tests (43 tests: 16 MCP + 12 DAP + 15 tools)
go test ./...

# Run with verbose output
go test -v ./...

# Run integration tests
./scripts/automated-integration-test.sh

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
internal/tools/      # Godot-specific MCP tools (Phase 3 ✅)
cmd/                 # Entry point
docs/                # Comprehensive documentation
tests/               # Integration tests and fixtures
scripts/             # Build and test automation
.claude/             # Claude Code skills and hooks
```

## License

MIT License - see [LICENSE](LICENSE)
