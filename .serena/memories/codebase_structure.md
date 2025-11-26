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
│   │   ├── server.go           # Concurrent MCP server
│   │   ├── types.go            # JSONRPC types
│   │   ├── transport.go        # Thread-safe transport
│   │   └── server_concurrency_test.go # Concurrency verification
│   ├── dap/                     # DAP Client Layer - TCP to Godot
│   │   ├── client.go           # Protocol implementation
│   │   ├── session.go          # Session lifecycle
│   │   ├── events.go           # Event filtering
│   │   ├── timeout.go          # Timeout definitions
│   │   └── godot.go            # Godot-specific config
│   └── tools/                   # Tool Layer - MCP Tools
│       ├── registry.go         # Tool registration
│       ├── connect.go          # Connection tools
│       ├── launch.go           # Launch tools
│       ├── breakpoints.go      # Breakpoint tools
│       ├── execution.go        # Execution control
│       ├── inspection.go       # Runtime inspection
│       ├── advanced.go         # Advanced tools
│       ├── errors.go           # Error formatting
│       ├── path.go             # Path resolution
│       └── formatting.go       # Type formatting
├── pkg/                         # Shared Packages
│   ├── daptest/                # Mock DAP server for testing
│   └── godot/                  # Godot utilities
├── tests/
│   └── fixtures/
│       └── test-project/       # Godot test project
├── docs/                        # Documentation
│   ├── PLAN.md                  # Status and Roadmap
│   ├── TOOLS.md                 # Tool Reference
│   ├── EXAMPLES.md              # Usage Examples
│   ├── reference/               # Stable reference
│   ├── implementation-notes/    # Phase notes
│   ├── research/                # Test results & Analysis
│   └── godot-upstream/          # Upstream contribution info
├── scripts/
│   └── automated-integration-test.sh
├── .claude/                     # Claude Code config
├── .serena/                     # Serena memory config
├── go.mod                       # Go module def
├── CHANGELOG.md                 # Version history
└── README.md                    # Project readme

## Layer Architecture

### Layer 1: MCP Layer (`internal/mcp/`)
**Purpose**: Handle stdio-based MCP protocol communication.
- **Key Files**: `server.go` (Concurrent), `transport.go` (Thread-safe)
- **Recent Changes**: Implemented goroutine-per-request model to prevent deadlocks.

### Layer 2: DAP Client Layer (`internal/dap/`)
**Purpose**: Communicate with Godot editor's DAP server via TCP.
- **Key Files**: `client.go`, `session.go`, `timeout.go`
- **Recent Changes**: Added timeout context wrappers for all operations.

### Layer 3: Tool Layer (`internal/tools/`)
**Purpose**: Implement Godot-specific MCP tools.
- **Key Files**: `launch.go`, `inspection.go`, `path.go`
- **Recent Changes**: Added `path.go` for `res://` resolution.

**Tool Count**: 17 tools across 4 phases.

## Test Organization
- Unit tests: `internal/*/*_test.go`
- Concurrency tests: `internal/mcp/server_concurrency_test.go`
- Integration tests: `scripts/`
- Debug Utilities: `cmd/*/`
