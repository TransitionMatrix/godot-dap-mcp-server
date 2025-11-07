# Godot DAP MCP Server - Architecture

**Last Updated**: 2025-11-07

This document describes the system architecture, design decisions, and critical implementation patterns for the godot-dap-mcp-server.

---

## Table of Contents

1. [System Overview](#system-overview)
2. [Architecture Layers](#architecture-layers)
3. [Technical Foundation](#technical-foundation)
4. [Project Structure](#project-structure)
5. [Critical Implementation Patterns](#critical-implementation-patterns)
6. [Design Decisions](#design-decisions)

---

## System Overview

**godot-dap-mcp-server** bridges AI agents (via MCP) to Godot's Debug Adapter Protocol server, enabling interactive runtime debugging of Godot games.

### High-Level Flow

```
┌─────────────────┐
│  Claude Code    │  (or any MCP client)
│  (MCP Client)   │
└────────┬────────┘
         │ MCP Protocol
         │ (stdin/stdout)
         ↓
┌─────────────────────────────────────┐
│  godot-dap-mcp-server               │
│  ┌──────────────────────────────┐   │
│  │ MCP Layer (stdio)            │   │
│  │ - Parse requests             │   │
│  │ - Route to tools             │   │
│  │ - Format responses           │   │
│  └──────────┬───────────────────┘   │
│             ↓                       │
│  ┌──────────────────────────────┐   │
│  │ Tool Layer (Godot-specific)  │   │
│  │ - godot_connect              │   │
│  │ - godot_launch_scene         │   │
│  │ - godot_set_breakpoint       │   │
│  │ - godot_step_over/in         │   │
│  │ - godot_evaluate             │   │
│  │ - godot_get_variables        │   │
│  └──────────┬───────────────────┘   │
│             ↓                       │
│  ┌──────────────────────────────┐   │
│  │ DAP Client Layer             │   │
│  │ - TCP connection to Godot    │   │
│  │ - Protocol handling          │   │
│  │ - Response parsing           │   │
│  │ - Event filtering            │   │
│  │ - Timeout management         │   │
│  └──────────┬───────────────────┘   │
└─────────────┼───────────────────────┘
              │ DAP Protocol
              │ (TCP, port 6006)
              ↓
┌─────────────────────────────────────┐
│  Godot Editor                       │
│  ┌──────────────────────────────┐   │
│  │ DAP Server (port 6006)       │   │
│  └──────────┬───────────────────┘   │
│             ↓                       │
│  ┌──────────────────────────────┐   │
│  │ EditorRunBar                 │   │
│  │ (launches game)              │   │
│  └──────────┬───────────────────┘   │
└─────────────┼───────────────────────┘
              ↓
┌─────────────────────────────────────┐
│  Game Instance (Running)            │
│  - Breakpoints active               │
│  - Debugging enabled                │
└─────────────────────────────────────┘
```

---

## Architecture Layers

### 1. MCP Layer (`internal/mcp/`)

**Purpose**: Handle stdio-based MCP protocol communication

**Responsibilities**:
- Parse JSONRPC 2.0 requests from stdin
- Route tool calls to registered handlers
- Format and send responses to stdout
- Maintain tool registry

**Key Files**:
- `server.go`: Core MCP server, request routing, stdio communication
- `types.go`: MCP request/response types (JSONRPC 2.0)
- `transport.go`: stdin/stdout message handling

**Protocol**: JSONRPC 2.0 over stdio (line-delimited JSON)

### 2. DAP Client Layer (`internal/dap/`)

**Purpose**: Communicate with Godot's DAP server via TCP

**Responsibilities**:
- Establish and manage TCP connection to Godot (port 6006)
- Send DAP requests and parse responses
- Filter async events from command responses
- Implement timeout protection for all operations
- Manage DAP session lifecycle

**Key Files**:
- `client.go`: DAP protocol implementation using `github.com/google/go-dap`
- `session.go`: DAP session lifecycle management
- `events.go`: Async event filtering (critical for response parsing)
- `timeout.go`: Timeout wrappers for all DAP operations (prevents hangs)
- `godot.go`: Godot-specific DAP extensions and launch parameters

**Protocol**: DAP over TCP (Content-Length header format)

**Session States**:
```
Disconnected → Connect() → Connected → Initialize() → Initialized
→ ConfigurationDone() → Configured → Launch() → Launched
```

### 3. Tool Layer (`internal/tools/`)

**Purpose**: Implement Godot-specific MCP tools

**Responsibilities**:
- Expose DAP functionality as MCP tools
- Validate parameters and provide helpful errors
- Format Godot objects for AI consumption
- Manage global session state

**Key Files**:
- `connect.go`: Connection management tools
- `launch.go`: Scene launching tools (main/current/custom)
- `breakpoints.go`: Breakpoint management
- `execution.go`: Execution control (continue, step-over, step-in)
- `inspection.go`: Runtime inspection (stack, variables, evaluation)
- `registry.go`: Tool registration and metadata

**Tool Naming Pattern**: `godot_<action>_<object>`
- Examples: `godot_connect`, `godot_set_breakpoint`, `godot_launch_main_scene`

---

## Technical Foundation

### Language Choice: Go

**Why Go:**
- ✅ Prior validation: 11/13 DAP commands proven working through experimentation
- ✅ Single binary: No runtime dependencies (Python, Node.js)
- ✅ Fast startup: <10ms (critical for stdio-spawned processes)
- ✅ Excellent concurrency: Goroutines handle async DAP events naturally
- ✅ Strong DAP library: `github.com/google/go-dap` (production-grade)
- ✅ Cross-platform: Compile for Mac/Linux/Windows from any platform

**Development Time:**
- Starting fresh: 2-3 weeks
- Leveraging existing code: 2-3 days

### Dependencies

```go
// Core dependencies
github.com/google/go-dap          // DAP protocol handling
encoding/json                      // MCP protocol JSON
os, bufio, io                      // stdio communication
context, time                      // Timeout management
net                                // TCP to Godot

// Optional (if needed)
github.com/spf13/cobra            // CLI if adding commands
github.com/sirupsen/logrus        // Structured logging
```

### go-dap API Pattern

The `github.com/google/go-dap` library uses a Codec pattern for message handling:

```go
// Reading messages
reader := bufio.NewReader(conn)
codec := dap.NewCodec()
data, err := dap.ReadBaseMessage(reader)
msg, err := codec.DecodeMessage(data)

// Writing messages
err := dap.WriteProtocolMessage(conn, message)
```

---

## Project Structure

```
godot-dap-mcp-server/
├── cmd/
│   └── godot-dap-mcp-server/
│       └── main.go                 # Entry point, stdio MCP server
│
├── internal/
│   ├── mcp/
│   │   ├── server.go              # MCP protocol handler (stdio)
│   │   ├── types.go               # MCP request/response types
│   │   └── transport.go           # stdin/stdout communication
│   │
│   ├── dap/
│   │   ├── client.go              # DAP client (TCP to Godot)
│   │   ├── session.go             # DAP session management
│   │   ├── events.go              # Event filtering/handling
│   │   ├── timeout.go             # Timeout wrapper utilities
│   │   └── godot.go               # Godot-specific DAP extensions
│   │
│   └── tools/
│       ├── connect.go             # godot_connect tool
│       ├── launch.go              # godot_launch_* tools
│       ├── breakpoints.go         # godot_set_breakpoint tool
│       ├── execution.go           # godot_continue, step_* tools
│       ├── inspection.go          # godot_get_*, godot_evaluate tools
│       └── registry.go            # Tool registration
│
├── pkg/
│   └── godot/
│       ├── types.go               # Godot-specific types (Node, Scene, etc.)
│       └── formatting.go          # Pretty-print Godot objects
│
├── docs/
│   ├── PLAN.md                    # Project planning and status
│   ├── ARCHITECTURE.md            # This document
│   ├── IMPLEMENTATION_GUIDE.md    # Component implementation specs
│   ├── TESTING.md                 # Testing strategies
│   ├── DEPLOYMENT.md              # Build and distribution
│   └── reference/
│       ├── GODOT_DAP_FAQ.md       # DAP troubleshooting Q&A
│       ├── DAP_PROTOCOL.md        # Protocol details
│       ├── CONVENTIONS.md         # Naming and error patterns
│       └── GODOT_SOURCE_ANALYSIS.md # Source code findings
│
├── tests/
│   ├── integration/               # Integration tests with Godot
│   ├── fixtures/                  # Test Godot project
│   └── unit/                      # Unit tests
│
├── scripts/
│   ├── build.sh                   # Cross-platform build script
│   └── test-godot.sh              # Launch test Godot project
│
├── go.mod
├── go.sum
└── README.md
```

---

## Critical Implementation Patterns

### 1. Event Filtering Pattern

**Problem**: DAP servers send async events mixed with command responses. Without filtering, commands hang waiting for responses that may never arrive in order.

**Solution**: Implemented in `internal/dap/events.go`:

```go
func (c *Client) waitForResponse(ctx context.Context, command string) (dap.Message, error) {
    for {
        msg, err := c.ReadWithTimeout(ctx)
        if err != nil {
            return nil, fmt.Errorf("failed to read response for %s: %w", command, err)
        }

        switch m := msg.(type) {
        case *dap.Response:
            if m.Command == command {
                if !m.Success {
                    return nil, fmt.Errorf("command %s failed: %s", command, m.Message)
                }
                return m, nil
            }
            // Not our response, continue waiting
            log.Printf("Received unexpected response for command: %s (waiting for %s)",
                m.Command, command)
            continue

        case *dap.Event:
            // Log event but continue waiting for response
            c.logEvent(m)
            continue

        case *dap.InitializeResponse:
            if command == "initialize" {
                return m, nil
            }
            continue

        // ... handle 13+ more specific response types
        }
    }
}
```

**Key Points**:
- Reads messages in loop with timeout
- Type-switches on message type
- Logs and discards events
- Continues waiting until matching response found
- Handles 15+ specific response types for type safety

**This is the most critical reliability pattern in the DAP client.**

### 2. Timeout Protection Pattern

**Problem**: Some DAP commands can hang indefinitely (especially when Godot encounters errors).

**Solution**: Implemented in `internal/dap/timeout.go`:

```go
const (
    DefaultConnectTimeout  = 10 * time.Second
    DefaultCommandTimeout  = 30 * time.Second
    DefaultReadTimeout     = 5 * time.Second
)

func (c *Client) ReadWithTimeout(ctx context.Context) (dap.Message, error) {
    type result struct {
        msg dap.Message
        err error
    }

    resultChan := make(chan result, 1)

    go func() {
        msg, err := c.read()
        resultChan <- result{msg, err}
    }()

    select {
    case <-ctx.Done():
        return nil, fmt.Errorf("read timeout: %w", ctx.Err())
    case res := <-resultChan:
        return res.msg, res.err
    }
}
```

**Usage Example**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

return c.waitForResponse(ctx, "launch")
```

**Key Points**:
- All DAP operations use context-based timeouts (10-30s)
- Prevents permanent hangs
- Clear timeout error messages
- Goroutine + select pattern for cancellation

### 3. Godot Launch Flow Pattern

**Problem**: Godot requires a specific sequence to launch games via DAP.

**Solution**: Two-step launch process:

```go
// Step 1: Send launch request (stores parameters, doesn't launch yet)
launchReq := &dap.LaunchRequest{
    Arguments: argsJSON,  // Must be json.RawMessage
}
c.write(launchReq)
c.waitForResponse(ctx, "launch")

// Step 2: Send configurationDone (triggers actual launch)
configReq := &dap.ConfigurationDoneRequest{}
c.write(configReq)
c.waitForResponse(ctx, "configurationDone")

// Now game is launching
```

**Critical**: The `launch` request only *stores* parameters. Must send `configurationDone` to actually start the game.

### 4. Session State Machine Pattern

**Problem**: DAP operations have dependencies (must connect before launching, etc.).

**Solution**: Explicit state tracking in `internal/dap/session.go`:

```go
type SessionState int

const (
    StateDisconnected SessionState = iota
    StateConnected
    StateInitialized
    StateConfigured
    StateLaunched
)

type Session struct {
    client *Client
    state  SessionState
}

func (s *Session) Launch(ctx context.Context, args map[string]interface{}) error {
    if s.state != StateConfigured {
        return fmt.Errorf("cannot launch: session not configured (current state: %v)", s.state)
    }

    // ... perform launch

    s.state = StateLaunched
    return nil
}
```

**Key Points**:
- Enforces correct operation sequence
- Clear error messages when operations done out of order
- State transitions are explicit and validated

---

## Design Decisions

### 1. stdio-based MCP (not HTTP)

**Decision**: Use stdin/stdout for MCP communication

**Rationale**:
- Standard MCP pattern - works with all MCP clients
- No separate server process to manage
- Claude Code spawns server automatically
- Simpler deployment (single binary)

**Trade-offs**:
- ✅ Simple, reliable
- ✅ No port conflicts
- ❌ Can't inspect traffic easily (but can log to file)

### 2. Single DAP Session

**Decision**: Support only one debugging session at a time

**Rationale**:
- AI workflows typically debug one game at a time
- Simplifies state management
- Godot's DAP server model is single-session anyway

**Trade-offs**:
- ✅ Simpler implementation
- ✅ Clearer state management
- ❌ Can't debug multiple games simultaneously

### 3. Godot-Specific Tools (not generic DAP)

**Decision**: Implement Godot-specific MCP tools, not a generic DAP bridge

**Rationale**:
- AI agents benefit from high-level abstractions
- Godot-specific formatting (Vector2, Node paths, etc.)
- Hide DAP complexity (multi-step variable inspection)
- Optimize for common Godot workflows

**Trade-offs**:
- ✅ Better AI agent experience
- ✅ Godot-optimized formatting
- ❌ Less flexible than generic DAP

### 4. Timeout at All Layers

**Decision**: Implement timeouts on every potentially blocking operation

**Rationale**:
- Discovered during DAP experimentation that some commands hang
- Prevents MCP server from becoming unresponsive
- Better error messages than "hung forever"

**Trade-offs**:
- ✅ Prevents hangs
- ✅ Better UX
- ❌ Slight code complexity

### 5. Error Messages: Problem + Context + Solution

**Decision**: Structured error messages with actionable guidance

**Example**:
```
Failed to connect to Godot DAP server at localhost:6006

Possible causes:
1. Godot editor is not running
2. DAP server is not enabled in editor settings
3. DAP server is using a different port

Solutions:
1. Launch Godot editor
2. Enable DAP in Editor → Editor Settings → Network → Debug Adapter
3. Check port setting (default: 6006)
```

**Rationale**:
- AI agents can understand and act on errors
- Human users get clear troubleshooting steps
- Reduces support burden

---

## References

For detailed protocol information, see:
- [GODOT_DAP_FAQ.md](reference/GODOT_DAP_FAQ.md) - Common questions and troubleshooting
- [DAP_PROTOCOL.md](reference/DAP_PROTOCOL.md) - Detailed protocol specifications
- [GODOT_SOURCE_ANALYSIS.md](reference/GODOT_SOURCE_ANALYSIS.md) - Insights from Godot source code

For implementation details, see:
- [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md) - Component implementation specifications
