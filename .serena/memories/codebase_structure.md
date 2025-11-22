# Codebase Structure

## Repository Layout
```
godot-dap-mcp-server/
├── cmd/
│   ├── godot-dap-mcp-server/ # Main Entry point
│   ├── debug-launch/         # Utility: Isolated launch testing
│   ├── dump-setbreakpoints/  # Utility: Inspect DAP serialization
│   ├── launch-test/          # Utility: Test launch sequence
│   ├── test-dap-protocol/    # Utility: Protocol compliance test
│   ├── test-full-debug-workflow/ # Utility: End-to-end workflow
│   └── test-minimal-dap/     # Utility: Minimal protocol check
├── internal/
│   ├── mcp/                     # MCP Layer - stdio protocol
│   │   ├── server.go           # Core MCP server
│   │   ├── types.go            # JSONRPC types
│   │   └── transport.go        # stdin/stdout handling
│   ├── dap/                     # DAP Client Layer - TCP to Godot
│   │   ├── client.go           # Protocol implementation
│   │   ├── session.go          # Session lifecycle (State Machine)
│   │   ├── events.go           # Event filtering
│   │   ├── timeout.go          # Timeout wrappers
│   │   └── godot.go            # Godot-specific launch config
│   └── tools/                   # Tool Layer - MCP Tools
│       ├── registry.go         # Tool registration
│       ├── connect.go          # Connection tools
│       ├── launch.go           # Launch tools (Main, Scene, Current)
│       ├── breakpoints.go      # Breakpoint tools
│       ├── execution.go        # Execution control
│       ├── inspection.go       # Runtime inspection
│       ├── advanced.go         # Advanced tools (pause, set_variable)
│       └── formatting.go       # Type formatting
├── tests/
│   └── fixtures/
│       └── test-project/       # Godot test project
├── docs/                        # Documentation
│   ├── PLAN.md                  # Status and Roadmap
│   ├── reference/               # Stable reference
│   ├── implementation-notes/    # Phase notes
│   ├── research/                # Test results & Analysis
│   └── godot-upstream/          # Upstream contribution info
├── scripts/
│   ├── automated-integration-test.sh
│   └── integration-test.sh     # Manual integration helper
├── .claude/                     # Claude Code config
├── .serena/                     # Serena memory config
├── go.mod                       # Go module def
└── README.md                    # Project readme

## Layer Architecture

### Layer 1: MCP Layer (`internal/mcp/`)
**Purpose**: Handle stdio-based MCP protocol communication.
- **Key Files**: `server.go`, `types.go`, `transport.go`
- **Recent Changes**: Updated `types.go` for schema validation.

### Layer 2: DAP Client Layer (`internal/dap/`)
**Purpose**: Communicate with Godot editor's DAP server via TCP.
- **Key Files**: `client.go`, `session.go`, `godot.go`
- **Recent Changes**: Implemented deferred `ConfigurationDone` and `LaunchWithConfigurationDone`.

### Layer 3: Tool Layer (`internal/tools/`)
**Purpose**: Implement Godot-specific MCP tools.
- **Key Files**: `registry.go`, `launch.go`, `connect.go`, `execution.go`
- **Recent Changes**: Added `launch.go` with 3 new tools.

**Tool Count**: 17 tools across 4 phases.

## Test Organization
- Unit tests alongside source: `internal/*/`
- Integration tests: `scripts/`
- Debug Utilities: `cmd/*/`
- **Total tests**: ~65 (covering MCP, DAP, Tools)
