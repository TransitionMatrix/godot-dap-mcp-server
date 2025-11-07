# Codebase Structure

## Repository Layout
```
godot-dap-mcp-server/
├── cmd/
│   └── godot-dap-mcp-server/
│       └── main.go              # Entry point
├── internal/
│   ├── mcp/                     # MCP Layer - stdio protocol
│   │   ├── server.go           # Core MCP server, request routing
│   │   ├── types.go            # JSONRPC 2.0 request/response types
│   │   ├── transport.go        # stdin/stdout message handling
│   │   ├── server_test.go      # MCP server tests
│   │   └── transport_test.go   # Transport tests
│   ├── dap/                     # DAP Client Layer - TCP to Godot
│   │   ├── client.go           # DAP protocol implementation
│   │   ├── session.go          # Session lifecycle management
│   │   ├── events.go           # Async event filtering
│   │   ├── timeout.go          # Timeout wrappers (prevents hangs)
│   │   ├── godot.go            # Godot-specific DAP extensions
│   │   └── dap_test.go         # DAP client tests
│   └── tools/                   # Tool Layer - Godot MCP tools
│       ├── registry.go         # Tool registration
│       ├── ping.go             # Test tool
│       └── ping_test.go        # Tool tests
├── tests/
│   ├── unit/                    # Unit tests (if separate from internal/)
│   └── integration/             # Integration tests (require Godot)
├── docs/
│   ├── PLAN.md                  # Implementation phases and status
│   ├── ARCHITECTURE.md          # System design and patterns
│   ├── IMPLEMENTATION_GUIDE.md  # Component specifications
│   ├── TESTING.md              # Testing strategies
│   ├── DEPLOYMENT.md           # Build and distribution
│   └── reference/
│       ├── CONVENTIONS.md      # Coding standards
│       ├── DAP_PROTOCOL.md     # Godot DAP details
│       ├── GODOT_DAP_FAQ.md    # Troubleshooting
│       └── GODOT_SOURCE_ANALYSIS.md  # Godot source findings
├── scripts/
│   └── test-stdio.sh            # Test script for stdio communication
├── examples/                    # Example usage and demos
├── .claude/                     # Claude Code configuration
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── CLAUDE.md                    # Claude Code guidance
└── README.md                    # Project readme

## Layer Architecture

### Layer 1: MCP Layer (`internal/mcp/`)
**Purpose**: Handle stdio-based MCP protocol communication
- Reads JSONRPC 2.0 requests from stdin
- Routes to appropriate tool handlers
- Writes JSONRPC 2.0 responses to stdout
- Manages tool registry

**Key Files**:
- `server.go`: Request handling, tool routing, response formatting
- `types.go`: Request, Response, Tool, ToolParameter types
- `transport.go`: stdin/stdout reading/writing

### Layer 2: DAP Client Layer (`internal/dap/`)
**Purpose**: Communicate with Godot editor's DAP server via TCP
- Connects to Godot DAP server (localhost:6006 default)
- Implements DAP protocol using `github.com/google/go-dap`
- Filters async events from responses (critical pattern)
- Provides timeout protection for all operations
- Manages single debugging session lifecycle

**Key Files**:
- `client.go`: DAP protocol implementation, request/response
- `session.go`: Session state, connection management
- `events.go`: Event filtering (separates events from responses)
- `timeout.go`: Timeout wrappers for all DAP calls
- `godot.go`: Godot-specific launch parameters, scene modes

### Layer 3: Tool Layer (`internal/tools/`)
**Purpose**: Implement Godot-specific MCP tools
- Follows naming convention: `godot_<action>_<object>`
- Translates MCP tool calls to DAP operations
- Provides AI-optimized tool descriptions
- Handles parameter validation

**Key Files**:
- `registry.go`: Tool registration system
- Individual tool files: `connect.go`, `launch.go`, `breakpoints.go`, etc.

## Module Structure
- **Package**: `github.com/TransitionMatrix/godot-dap-mcp-server`
- **Internal packages**: Cannot be imported by external projects
- **Go version**: 1.25.3
- **Single dependency**: `github.com/google/go-dap v0.12.0`

## File Naming Conventions
- Lowercase with underscores: `stack_trace.go`
- Test files: `<name>_test.go`
- Group related functionality in same file
- Keep files focused on single responsibility

## Test Organization
- Unit tests alongside source: `internal/*/`
- Integration tests: `tests/integration/`
- Test helpers: In same package as tests
- Coverage target: >80% for critical paths
