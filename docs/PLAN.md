# Godot DAP MCP Server - Implementation Plan

**Date**: 2025-11-05
**Language**: Go
**Purpose**: MCP server providing interactive runtime debugging for Godot games via DAP protocol

---

## Executive Summary

Build a specialized MCP server that bridges Claude Code (or any MCP client) to Godot's Debug Adapter Protocol (DAP) server, enabling AI agents to perform interactive debugging of Godot games.

**Key Differentiators**:
- First MCP server for Godot interactive debugging
- Complements existing Godot MCP servers (GDAI focuses on editor/scene manipulation)
- Leverages existing tested Go codebase (11/13 commands proven working)
- stdio-based standard MCP pattern (no separate server process)
- Single binary distribution (no runtime dependencies)

---

## Table of Contents

1. [Project Goals](#project-goals)
2. [Architecture Overview](#architecture-overview)
3. [Technical Foundation](#technical-foundation)
4. [Project Structure](#project-structure)
5. [Implementation Phases](#implementation-phases)
6. [Detailed Component Design](#detailed-component-design)
7. [Testing Strategy](#testing-strategy)
8. [Deployment & Distribution](#deployment--distribution)
9. [Documentation Plan](#documentation-plan)
10. [Success Metrics](#success-metrics)

---

## Project Goals

### Primary Goals

1. **Enable AI-driven debugging**: Allow Claude Code and other MCP clients to debug Godot games interactively
2. **Leverage prior experiments**: Build on previous DAP testing (11/13 commands proven working)
3. **Godot-specific enhancements**: Support GDScript syntax, scene tree navigation, Godot object inspection
4. **Production-ready**: Implement timeout mechanisms, error handling, graceful failure recovery
5. **Easy deployment**: Single binary, stdio-based, zero configuration

### Secondary Goals

1. **Fill market gap**: Provide capability that GDAI MCP and other Godot servers don't offer
2. **Industry standard**: Use Microsoft's DAP protocol (future-proof)
3. **AI agent optimized**: Tool descriptions and parameters designed for LLM understanding
4. **Open source contribution**: Share with Godot and MCP communities

---

## Architecture Overview

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
│             ↓                        │
│  ┌──────────────────────────────┐   │
│  │ Tool Layer (Godot-specific)  │   │
│  │ - godot_connect              │   │
│  │ - godot_launch_scene         │   │
│  │ - godot_set_breakpoint       │   │
│  │ - godot_step_over/in/out     │   │
│  │ - godot_evaluate             │   │
│  │ - godot_get_variables        │   │
│  └──────────┬───────────────────┘   │
│             ↓                        │
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
│             ↓                        │
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

### Key Architectural Decisions

1. **stdio-based MCP**: Standard pattern, no separate server process to manage
2. **Single DAP session**: One debugging session at a time (sufficient for AI workflows)
3. **Godot-specific tools**: Not generic DAP - optimized for Godot use cases
4. **Timeout at all layers**: 10-30s timeouts prevent permanent hangs
5. **Graceful degradation**: Clear error messages, no crashes

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

### Patterns to Implement

From previous DAP experiments:
1. ✅ DAP connection handling
2. ✅ Response parsing with event filtering (infinite loop + type switch pattern)
3. ✅ Breakpoint management
4. ✅ Stepping commands (next, step-in)
5. ✅ Stack trace inspection
6. ✅ Variable inspection with deep drilling
7. ✅ Expression evaluation
8. ✅ Scope management

**What needs to be added:**
1. ❌ MCP stdio layer
2. ❌ Godot-specific tool wrappers
3. ❌ Timeout mechanisms (critical!)
4. ❌ Launch request with Godot parameters
5. ❌ step-out hang fix

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
│   ├── README.md                  # Main documentation
│   ├── TOOLS.md                   # Tool reference
│   ├── EXAMPLES.md                # Usage examples
│   └── ARCHITECTURE.md            # This document
│
├── examples/
│   ├── claude-code-config.json    # Example MCP client config
│   └── debugging-workflow.md      # Example AI debugging session
│
├── tests/
│   ├── integration/               # Integration tests with Godot
│   └── unit/                      # Unit tests
│
├── scripts/
│   ├── build.sh                   # Cross-platform build script
│   └── test-godot.sh              # Launch test Godot project
│
├── .github/
│   └── workflows/
│       └── ci.yml                 # GitHub Actions CI
│
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

---

## Implementation Phases

### Phase 1: Core MCP Server (Day 1) - PRIORITY

**Goal**: Get stdio MCP server running with one test tool

**Tasks**:
1. Create project structure
2. Implement stdio MCP server (`internal/mcp/server.go`)
3. Implement basic tool registration (`internal/tools/registry.go`)
4. Create one simple tool (`godot_ping` - just echoes back)
5. Test with manual stdio input/output
6. Test with Claude Code MCP client

**Deliverables**:
- Working stdio MCP server
- Tool registration system
- One test tool proven working

**Success Criteria**:
- Claude Code can spawn server
- Server receives MCP requests via stdin
- Server responds via stdout
- Clean shutdown on EOF

---

### Phase 2: DAP Client Implementation (Day 1-2) - PRIORITY

**Goal**: Implement DAP client patterns from experimentation

**Tasks**:
1. Implement DAP client code in `internal/dap/client.go`
2. Structure using patterns discovered during testing
3. Add timeout wrappers (`internal/dap/timeout.go`)
4. Implement connection management
5. Migrate response parsing patterns (infinite loop + type switch)
6. Add event filtering logic

**Patterns to Implement**:
```go
// From previous experiments:
- connectToDebugger()
- configurationDone()
- setBreakpoints()
- threads()
- stackTrace()
- scopes()
- variables()
- evaluate()
- continue()
- next()
- stepIn()
- stepOut() // with hang fix
```

**Deliverables**:
- DAP client with all proven working commands
- Timeout protection (10-30s on all requests)
- Event filtering for async DAP events

**Success Criteria**:
- Can connect to Godot DAP server (port 6006)
- Can send initialize/configurationDone
- Can set breakpoints
- All inspection commands work

---

### Phase 3: Core Debugging Tools (Day 2) - HIGH PRIORITY

**Goal**: Implement essential debugging tools using DAP client

**Tools to Implement** (8 core tools):

1. **`godot_connect`**
   ```
   Parameters: port (default: 6006)
   Returns: Connection status, server capabilities
   ```

2. **`godot_disconnect`**
   ```
   Parameters: none
   Returns: Disconnection status
   ```

3. **`godot_set_breakpoint`**
   ```
   Parameters: file (absolute path), line (integer)
   Returns: Breakpoint ID, verification status
   ```

4. **`godot_clear_breakpoint`**
   ```
   Parameters: file, line
   Returns: Success status
   ```

5. **`godot_continue`**
   ```
   Parameters: none
   Returns: Success status
   ```

6. **`godot_step_over`**
   ```
   Parameters: none
   Returns: New location (file, line)
   ```

7. **`godot_step_in`**
   ```
   Parameters: none
   Returns: New location (file, line)
   ```

8. **`godot_step_out`**
   ```
   Parameters: none
   Returns: New location (file, line)
   Note: Fix the hang issue from testing!
   ```

**Deliverables**:
- 8 core tools implemented
- Tool registration in MCP server
- Clear error messages
- Timeout protection

**Success Criteria**:
- Can set breakpoints via MCP tool
- Can control execution (continue, stepping)
- No hangs (step-out issue fixed)
- Clear error messages on failure

---

### Phase 4: Inspection Tools (Day 2-3) - HIGH PRIORITY

**Goal**: Implement runtime inspection tools

**Tools to Implement** (5 inspection tools):

1. **`godot_get_stack_trace`**
   ```
   Parameters: none
   Returns: Array of stack frames with file/line/function
   ```

2. **`godot_get_scopes`**
   ```
   Parameters: frame_id (default: 0)
   Returns: Locals, Members, Globals scopes
   ```

3. **`godot_get_variables`**
   ```
   Parameters: scope (Locals|Members|Globals), frame_id (default: 0)
   Returns: Array of variables with name/type/value
   ```

4. **`godot_evaluate`**
   ```
   Parameters: expression (GDScript), frame_id (default: 0)
   Returns: Result value, type
   ```

5. **`godot_get_threads`**
   ```
   Parameters: none
   Returns: Array of threads (usually just "Main" for Godot)
   ```

**Godot-Specific Enhancements**:
- Pretty-print Node objects (show name, type, parent path)
- Format Vector2/Vector3 nicely (e.g., "Vector2(100, 200)")
- Show scene tree path for Node references
- Handle GDScript null vs. empty string

**Deliverables**:
- 5 inspection tools
- Godot object formatting
- Deep drilling support (reference chains)

**Success Criteria**:
- Can inspect call stack at breakpoint
- Can view variable values in all scopes
- Can evaluate GDScript expressions
- Godot objects formatted nicely

---

### Phase 5: Launch Tools (Day 3) - MEDIUM PRIORITY

**Goal**: Implement scene launching via DAP

**Research Finding**: Godot's DAP server supports `launch` request with these parameters:
- `project`: Project path (validated)
- `scene`: "main" | "current" | "res://path/to/scene.tscn"
- `platform`: "host" (default) | "android" | "web"
- `noDebug`: boolean
- `godot/custom_data`: boolean
- Additional play arguments

**Tools to Implement** (3 launch tools):

1. **`godot_launch_main_scene`**
   ```
   Parameters:
     - project_path (required)
     - profiling (default: false)
     - debug_collisions (default: false)
     - debug_paths (default: false)
     - debug_navigation (default: false)
     - additional_options (optional string)

   Returns: Launch status, game PID

   Implementation:
     - Send DAP launch request with scene="main"
     - Wait for configurationDone
     - Verify game started
   ```

2. **`godot_launch_scene`**
   ```
   Parameters:
     - scene_path (required, e.g., "res://scenes/player.tscn")
     - project_path (required)
     - profiling (default: false)
     - additional_options (optional)

   Returns: Launch status

   Implementation:
     - Send DAP launch request with scene=<path>
     - Handle file validation
   ```

3. **`godot_launch_current_scene`**
   ```
   Parameters:
     - project_path (required)

   Returns: Launch status

   Implementation:
     - Send DAP launch request with scene="current"
     - Useful for AI testing currently open scene
   ```

**Critical Implementation Details**:

Based on Godot source (`debug_adapter_parser.cpp`):
```cpp
Dictionary DebugAdapterParser::req_launch(const Dictionary &p_params) const {
    Dictionary args = p_params["arguments"];

    // Godot validates project path
    if (args.has("project") && !is_valid_path(args["project"])) {
        return prepare_error_response(p_params, DAP::ErrorType::WRONG_PATH, ...);
    }

    // Launch deferred until configurationDone
    DebugAdapterProtocol::get_singleton()->get_current_peer()->pending_launch = p_params;
    return Dictionary();
}

Dictionary DebugAdapterParser::_launch_process(const Dictionary &p_params) const {
    Dictionary args = p_params["arguments"];
    const String scene = args.get("scene", "main");

    if (scene == "main") {
        EditorRunBar::get_singleton()->play_main_scene(false, play_args);
    } else if (scene == "current") {
        EditorRunBar::get_singleton()->play_current_scene(false, play_args);
    } else {
        EditorRunBar::get_singleton()->play_custom_scene(scene, play_args);
    }
}
```

**Go Implementation Pattern**:
```go
func (s *DAPSession) LaunchScene(project, scene string, options LaunchOptions) error {
    // Build launch arguments
    args := map[string]interface{}{
        "project": project,
        "scene":   scene,
        "noDebug": false,
    }

    if options.Profiling {
        args["profiling"] = true
    }
    // ... add other options

    // Send launch request
    if err := s.client.LaunchRequest(args); err != nil {
        return fmt.Errorf("launch failed: %w", err)
    }

    // Send configurationDone to trigger actual launch
    if err := s.client.ConfigurationDone(); err != nil {
        return fmt.Errorf("configuration failed: %w", err)
    }

    // Wait for debug_enter or error
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    return s.waitForLaunch(ctx)
}
```

**Deliverables**:
- 3 launch tools
- Full Godot launch parameter support
- Error handling for invalid paths/scenes
- Timeout protection

**Success Criteria**:
- Can launch main scene via MCP tool (no manual F5!)
- Can launch specific scenes by path
- Clear errors for invalid scenes
- Game launches with breakpoints active

---

### Phase 6: Advanced Tools (Day 3-4) - OPTIONAL

**Goal**: Add nice-to-have debugging tools

**Tools to Implement** (4 advanced tools):

1. **`godot_pause`**
   ```
   Parameters: none
   Returns: Pause status
   ```

2. **`godot_set_variable`**
   ```
   Parameters: variable_name, new_value, frame_id
   Returns: Success status, new value
   ```

3. **`godot_get_scene_tree`**
   ```
   Parameters: none (when paused at breakpoint)
   Returns: Current scene tree structure
   Note: Godot-specific enhancement
   ```

4. **`godot_inspect_node`**
   ```
   Parameters: node_path (e.g., "/root/Main/Player")
   Returns: Node properties, signals, children
   Note: Godot-specific enhancement
   ```

**Deliverables**:
- 4 advanced tools (if time permits)
- Enhanced Godot object inspection

**Success Criteria**:
- Can pause running game
- Can modify variables at runtime
- Can inspect scene tree structure

---

### Phase 7: Error Handling & Polish (Day 4) - CRITICAL

**Goal**: Production-ready error handling and user experience

**Tasks**:

1. **Timeout Implementation** (CRITICAL - from testing findings):
   ```go
   // Wrapper for all DAP requests
   func (c *DAPClient) withTimeout(ctx context.Context, fn func() error) error {
       ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
       defer cancel()

       errChan := make(chan error, 1)
       go func() {
           errChan <- fn()
       }()

       select {
       case err := <-errChan:
           return err
       case <-ctx.Done():
           return fmt.Errorf("request timeout after 30s: %w", ctx.Err())
       }
   }
   ```

2. **Error Message Formatting**:
   - Clear, actionable error messages
   - Suggest fixes (e.g., "Is Godot editor running with DAP enabled?")
   - Include context (file, line, state)

3. **Graceful Degradation**:
   - Handle Godot editor restart
   - Recover from connection drops
   - Clear state on disconnect

4. **step-out Hang Fix** (from testing):
   - Investigate why step-out hangs
   - Implement proper response parsing
   - Add extra timeout protection

5. **Logging**:
   - Structured logging (optional: use logrus)
   - Debug mode for troubleshooting
   - Log DAP protocol traffic (when enabled)

**Deliverables**:
- Timeout on all DAP requests
- Clear error messages
- Graceful failure recovery
- step-out working reliably

**Success Criteria**:
- No permanent hangs (all requests timeout)
- Helpful error messages
- step-out command works
- Survives Godot editor restart

---

### Phase 8: Documentation (Day 4-5) - HIGH PRIORITY

**Goal**: Comprehensive documentation for users and AI agents

**Documents to Create**:

1. **README.md**:
   - Project overview
   - Installation instructions
   - Quick start guide
   - Example usage

2. **TOOLS.md**:
   - Complete tool reference
   - Each tool with: parameters, returns, examples, notes
   - Organized by category (connection, execution, inspection, launch)

3. **EXAMPLES.md**:
   - Real debugging workflows
   - AI agent example sessions
   - Common patterns

4. **ARCHITECTURE.md**:
   - System design
   - DAP flow diagrams
   - Extension points

5. **CONTRIBUTING.md**:
   - How to add new tools
   - Code style guide
   - Testing guidelines

**Tool Description Pattern** (for AI agents):
```go
Tool{
    Name: "godot_set_breakpoint",
    Description: `Set a breakpoint in GDScript code. The game will pause execution when this line is reached, allowing you to inspect variables and step through code.

    Use this when you want to:
    - Pause execution at a specific line
    - Inspect variable values at a certain point
    - Debug why certain code isn't executing

    Example: Set breakpoint at player controller's collision check
    godot_set_breakpoint(file="/path/to/player.gd", line=42)`,

    Parameters: []Parameter{
        {Name: "file", Type: "string", Required: true, Description: "Absolute path to .gd file"},
        {Name: "line", Type: "integer", Required: true, Description: "Line number (1-indexed)"},
    },

    Returns: "Breakpoint ID and verification status",
}
```

**Deliverables**:
- 5 comprehensive documentation files
- AI-optimized tool descriptions
- Code examples for common workflows

**Success Criteria**:
- AI agents can use server without human help
- Clear tool descriptions
- Working code examples

---

## Detailed Component Design

### 1. MCP Server Layer (`internal/mcp/`)

**Purpose**: Handle stdio-based MCP protocol communication

**Key Files**:

#### `server.go`
```go
package mcp

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io"
    "os"
)

type Server struct {
    tools map[string]Tool
    stdin io.Reader
    stdout io.Writer
}

type Tool struct {
    Name        string
    Description string
    Parameters  []Parameter
    Handler     func(params map[string]interface{}) (interface{}, error)
}

func NewServer() *Server {
    return &Server{
        tools:  make(map[string]Tool),
        stdin:  os.Stdin,
        stdout: os.Stdout,
    }
}

func (s *Server) RegisterTool(tool Tool) {
    s.tools[tool.Name] = tool
}

func (s *Server) ListenStdio() error {
    scanner := bufio.NewScanner(s.stdin)

    for scanner.Scan() {
        line := scanner.Text()

        // Parse MCP request
        var req MCPRequest
        if err := json.Unmarshal([]byte(line), &req); err != nil {
            s.sendError(fmt.Sprintf("invalid request: %v", err))
            continue
        }

        // Handle request
        resp := s.handleRequest(req)

        // Send response
        if err := s.sendResponse(resp); err != nil {
            return err
        }
    }

    return scanner.Err()
}

func (s *Server) handleRequest(req MCPRequest) MCPResponse {
    switch req.Method {
    case "tools/list":
        return s.listTools()
    case "tools/call":
        return s.callTool(req.Params)
    default:
        return s.errorResponse(fmt.Sprintf("unknown method: %s", req.Method))
    }
}

func (s *Server) callTool(params map[string]interface{}) MCPResponse {
    toolName, ok := params["name"].(string)
    if !ok {
        return s.errorResponse("missing tool name")
    }

    tool, exists := s.tools[toolName]
    if !exists {
        return s.errorResponse(fmt.Sprintf("tool not found: %s", toolName))
    }

    // Call tool handler
    result, err := tool.Handler(params["arguments"].(map[string]interface{}))
    if err != nil {
        return s.errorResponse(fmt.Sprintf("tool error: %v", err))
    }

    return s.successResponse(result)
}
```

#### `types.go`
```go
package mcp

type MCPRequest struct {
    JSONRPC string                 `json:"jsonrpc"`
    ID      interface{}            `json:"id"`
    Method  string                 `json:"method"`
    Params  map[string]interface{} `json:"params"`
}

type MCPResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Result  interface{} `json:"result,omitempty"`
    Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

type Parameter struct {
    Name        string
    Type        string
    Required    bool
    Description string
    Default     interface{}
}
```

---

### 2. DAP Client Layer (`internal/dap/`)

**Purpose**: Communicate with Godot's DAP server via TCP

**Key Files**:

#### `client.go`
```go
package dap

import (
    "context"
    "fmt"
    "net"
    "time"

    "github.com/google/go-dap"
)

type Client struct {
    conn      net.Conn
    protocol  *dap.Client
    seq       int
    timeout   time.Duration
}

func NewClient(timeout time.Duration) *Client {
    return &Client{
        timeout: timeout,
    }
}

func (c *Client) Connect(host string, port int) error {
    addr := fmt.Sprintf("%s:%d", host, port)

    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()

    conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", addr)
    if err != nil {
        return fmt.Errorf("failed to connect to %s: %w", addr, err)
    }

    c.conn = conn
    c.protocol = dap.NewClient(conn)

    return c.initialize()
}

func (c *Client) initialize() error {
    // Send initialize request
    req := &dap.InitializeRequest{
        Request: *c.newRequest("initialize"),
        Arguments: dap.InitializeRequestArguments{
            ClientID: "godot-dap-mcp-server",
            AdapterID: "godot",
        },
    }

    if err := c.protocol.WriteProtocolMessage(req); err != nil {
        return err
    }

    // Wait for initialize response
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()

    return c.waitForResponse(ctx, "initialize")
}

func (c *Client) SetBreakpoints(file string, lines []int) error {
    req := &dap.SetBreakpointsRequest{
        Request: *c.newRequest("setBreakpoints"),
        Arguments: dap.SetBreakpointsArguments{
            Source: dap.Source{Path: file},
            Lines:  lines,
        },
    }

    if err := c.protocol.WriteProtocolMessage(req); err != nil {
        return err
    }

    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()

    return c.waitForResponse(ctx, "setBreakpoints")
}

// ... more DAP commands
```

#### `timeout.go`
```go
package dap

import (
    "context"
    "fmt"
    "time"
)

// WithTimeout wraps any operation with a timeout
func WithTimeout(timeout time.Duration, fn func(ctx context.Context) error) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    errChan := make(chan error, 1)
    go func() {
        errChan <- fn(ctx)
    }()

    select {
    case err := <-errChan:
        return err
    case <-ctx.Done():
        return fmt.Errorf("operation timeout after %v: %w", timeout, ctx.Err())
    }
}

// ReadWithTimeout reads a DAP message with timeout
func (c *Client) ReadWithTimeout(ctx context.Context) (dap.Message, error) {
    msgChan := make(chan dap.Message, 1)
    errChan := make(chan error, 1)

    go func() {
        msg, err := c.protocol.ReadProtocolMessage()
        if err != nil {
            errChan <- err
            return
        }
        msgChan <- msg
    }()

    select {
    case msg := <-msgChan:
        return msg, nil
    case err := <-errChan:
        return nil, err
    case <-ctx.Done():
        return nil, fmt.Errorf("read timeout: %w", ctx.Err())
    }
}
```

#### `events.go`
```go
package dap

import (
    "context"
    "fmt"

    "github.com/google/go-dap"
)

// waitForResponse reads messages until expected response arrives
// Filters out async events (StoppedEvent, ContinuedEvent, etc.)
func (c *Client) waitForResponse(ctx context.Context, command string) error {
    for {
        msg, err := c.ReadWithTimeout(ctx)
        if err != nil {
            return err
        }

        switch m := msg.(type) {
        case *dap.Response:
            // Check if it's the expected response
            if m.Command == command {
                if !m.Success {
                    return fmt.Errorf("command failed: %s", m.Message)
                }
                return nil
            }

        case *dap.Event:
            // Log event but continue waiting
            c.logEvent(m)

        default:
            // Unknown message type, continue
            continue
        }
    }
}

func (c *Client) logEvent(event *dap.Event) {
    // Optional: log async events for debugging
    // Could emit these to MCP client as notifications
}
```

#### `godot.go`
```go
package dap

import (
    "context"
    "fmt"
)

// GodotLaunchOptions contains Godot-specific launch parameters
type GodotLaunchOptions struct {
    Project             string
    Scene               string // "main", "current", or path
    Platform            string // "host", "android", "web"
    NoDebug             bool
    Profiling           bool
    DebugCollisions     bool
    DebugPaths          bool
    DebugNavigation     bool
    DebugAvoidance      bool
    AdditionalOptions   string
}

// LaunchGodotScene sends a launch request with Godot-specific parameters
func (c *Client) LaunchGodotScene(opts GodotLaunchOptions) error {
    // Build launch arguments
    args := map[string]interface{}{
        "project": opts.Project,
        "scene":   opts.Scene,
        "platform": "host",
        "noDebug": opts.NoDebug,
    }

    if opts.Profiling {
        args["profiling"] = true
    }
    if opts.DebugCollisions {
        args["debug_collisions"] = true
    }
    // ... add other options

    // Send launch request
    req := &dap.LaunchRequest{
        Request: *c.newRequest("launch"),
        Arguments: args,
    }

    if err := c.protocol.WriteProtocolMessage(req); err != nil {
        return fmt.Errorf("failed to send launch: %w", err)
    }

    // Wait for launch response
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()

    if err := c.waitForResponse(ctx, "launch"); err != nil {
        return err
    }

    // Send configurationDone to trigger actual launch
    configReq := &dap.ConfigurationDoneRequest{
        Request: *c.newRequest("configurationDone"),
    }

    if err := c.protocol.WriteProtocolMessage(configReq); err != nil {
        return fmt.Errorf("failed to send configurationDone: %w", err)
    }

    ctx2, cancel2 := context.WithTimeout(context.Background(), c.timeout)
    defer cancel2()

    return c.waitForResponse(ctx2, "configurationDone")
}
```

---

### 3. Tool Layer (`internal/tools/`)

**Purpose**: Implement Godot-specific MCP tools

**Example Tool** (`connect.go`):
```go
package tools

import (
    "fmt"

    "your-repo/internal/dap"
    "your-repo/internal/mcp"
)

func RegisterConnectTool(server *mcp.Server, dapClient *dap.Client) {
    server.RegisterTool(mcp.Tool{
        Name: "godot_connect",
        Description: `Connect to Godot's DAP server to begin debugging session.

        Prerequisites:
        - Godot editor must be running
        - DAP server must be enabled in editor settings
        - Default port is 6006

        Use this as the first step before any debugging operations.

        Example: Connect to local Godot editor
        godot_connect(port=6006)`,

        Parameters: []mcp.Parameter{
            {
                Name:        "port",
                Type:        "integer",
                Required:    false,
                Default:     6006,
                Description: "DAP server port (default: 6006)",
            },
        },

        Handler: func(params map[string]interface{}) (interface{}, error) {
            port := 6006
            if p, ok := params["port"].(float64); ok {
                port = int(p)
            }

            // Connect to Godot
            if err := dapClient.Connect("localhost", port); err != nil {
                return nil, fmt.Errorf("failed to connect to Godot DAP server: %w. Is Godot editor running with DAP enabled?", err)
            }

            // Get capabilities
            caps := dapClient.GetCapabilities()

            return map[string]interface{}{
                "status": "connected",
                "port":   port,
                "capabilities": caps,
            }, nil
        },
    })
}
```

**Example Tool** (`launch.go`):
```go
package tools

import (
    "fmt"
    "os"

    "your-repo/internal/dap"
    "your-repo/internal/mcp"
)

func RegisterLaunchTools(server *mcp.Server, dapClient *dap.Client) {
    // godot_launch_main_scene
    server.RegisterTool(mcp.Tool{
        Name: "godot_launch_main_scene",
        Description: `Launch the project's main scene (defined in project.godot) with debugging enabled.

        This is equivalent to pressing F5 in the Godot editor, but controlled programmatically.

        The scene will run until:
        - A breakpoint is hit
        - The game is closed
        - You call godot_pause()

        Use this when you want to:
        - Test the game from the beginning
        - Debug initialization code
        - Run the full game flow

        Example: Launch main scene with profiling
        godot_launch_main_scene(project_path="/path/to/project", profiling=true)`,

        Parameters: []mcp.Parameter{
            {
                Name:        "project_path",
                Type:        "string",
                Required:    true,
                Description: "Absolute path to project directory (contains project.godot)",
            },
            {
                Name:        "profiling",
                Type:        "boolean",
                Required:    false,
                Default:     false,
                Description: "Enable performance profiling",
            },
            {
                Name:        "debug_collisions",
                Type:        "boolean",
                Required:    false,
                Default:     false,
                Description: "Show collision shapes visually",
            },
        },

        Handler: func(params map[string]interface{}) (interface{}, error) {
            projectPath, ok := params["project_path"].(string)
            if !ok || projectPath == "" {
                return nil, fmt.Errorf("project_path is required")
            }

            // Validate project path
            if _, err := os.Stat(projectPath + "/project.godot"); err != nil {
                return nil, fmt.Errorf("invalid project path: project.godot not found at %s", projectPath)
            }

            // Build launch options
            opts := dap.GodotLaunchOptions{
                Project:         projectPath,
                Scene:           "main",
                Profiling:       params["profiling"].(bool),
                DebugCollisions: params["debug_collisions"].(bool),
            }

            // Launch scene
            if err := dapClient.LaunchGodotScene(opts); err != nil {
                return nil, fmt.Errorf("failed to launch scene: %w", err)
            }

            return map[string]interface{}{
                "status": "launched",
                "scene":  "main",
                "project": projectPath,
            }, nil
        },
    })

    // godot_launch_scene (custom scene)
    server.RegisterTool(mcp.Tool{
        Name: "godot_launch_scene",
        Description: `Launch a specific scene by path with debugging enabled.

        Use this when you want to:
        - Test a specific scene in isolation
        - Debug a scene that isn't the main scene
        - Skip to a particular game state

        Example: Launch player controller test scene
        godot_launch_scene(scene_path="res://tests/test_player.tscn", project_path="/path/to/project")`,

        Parameters: []mcp.Parameter{
            {
                Name:        "scene_path",
                Type:        "string",
                Required:    true,
                Description: "Scene path (use res:// format, e.g., res://scenes/player.tscn)",
            },
            {
                Name:        "project_path",
                Type:        "string",
                Required:    true,
                Description: "Absolute path to project directory",
            },
        },

        Handler: func(params map[string]interface{}) (interface{}, error) {
            scenePath := params["scene_path"].(string)
            projectPath := params["project_path"].(string)

            opts := dap.GodotLaunchOptions{
                Project: projectPath,
                Scene:   scenePath,
            }

            if err := dapClient.LaunchGodotScene(opts); err != nil {
                return nil, err
            }

            return map[string]interface{}{
                "status": "launched",
                "scene":  scenePath,
            }, nil
        },
    })
}
```

---

## Testing Strategy

### Unit Tests

**What to Test**:
- MCP protocol parsing (request/response)
- Tool parameter validation
- Error message formatting
- Timeout mechanisms

**Example**:
```go
func TestMCPServer_CallTool(t *testing.T) {
    server := mcp.NewServer()

    // Register test tool
    called := false
    server.RegisterTool(mcp.Tool{
        Name: "test_tool",
        Handler: func(params map[string]interface{}) (interface{}, error) {
            called = true
            return "success", nil
        },
    })

    // Call tool
    req := mcp.MCPRequest{
        Method: "tools/call",
        Params: map[string]interface{}{
            "name": "test_tool",
            "arguments": map[string]interface{}{},
        },
    }

    resp := server.handleRequest(req)

    if !called {
        t.Error("tool not called")
    }
    if resp.Error != nil {
        t.Errorf("unexpected error: %v", resp.Error)
    }
}
```

---

### Integration Tests

**What to Test**:
- Full MCP → DAP → Godot flow
- Breakpoint setting and hitting
- Variable inspection
- Stepping commands
- Launch functionality

**Setup**:
1. Launch Godot editor in CI (headless mode with DAP)
2. Load test project
3. Run integration tests

**Example**:
```go
func TestGodotDebugging_SetBreakpointAndHit(t *testing.T) {
    // Connect to Godot (running in CI)
    client := dap.NewClient(30 * time.Second)
    if err := client.Connect("localhost", 6006); err != nil {
        t.Fatalf("failed to connect: %v", err)
    }
    defer client.Disconnect()

    // Set breakpoint
    testFile := "/path/to/test/player.gd"
    if err := client.SetBreakpoints(testFile, []int{42}); err != nil {
        t.Fatalf("failed to set breakpoint: %v", err)
    }

    // Launch scene
    opts := dap.GodotLaunchOptions{
        Project: "/path/to/test/project",
        Scene:   "main",
    }
    if err := client.LaunchGodotScene(opts); err != nil {
        t.Fatalf("failed to launch: %v", err)
    }

    // Wait for breakpoint hit (with timeout)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    stopped, err := client.WaitForStop(ctx)
    if err != nil {
        t.Fatalf("breakpoint not hit: %v", err)
    }

    if stopped.Line != 42 {
        t.Errorf("wrong line: got %d, want 42", stopped.Line)
    }
}
```

---

### Manual Testing

**Test Scenarios**:
1. Connect to Godot and list capabilities
2. Set breakpoint, launch scene, verify pause
3. Step through code (next, step-in, step-out)
4. Inspect variables at different stack frames
5. Evaluate GDScript expressions
6. Launch different scenes
7. Handle Godot editor restart
8. Test timeout on hung commands

**Test Project**:
Create minimal Godot project in `tests/fixtures/test-project/`:
```
test-project/
├── project.godot
├── main.tscn
└── player.gd          # Simple script with testable code
```

---

## Deployment & Distribution

### Build Process

**Cross-compilation** (from Mac):
```bash
# scripts/build.sh
#!/bin/bash

set -e

VERSION=$(git describe --tags --always)
BUILD_DIR="build"

mkdir -p $BUILD_DIR

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o $BUILD_DIR/godot-dap-mcp-server-darwin-amd64 cmd/godot-dap-mcp-server/main.go

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o $BUILD_DIR/godot-dap-mcp-server-darwin-arm64 cmd/godot-dap-server/main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o $BUILD_DIR/godot-dap-mcp-server-linux-amd64 cmd/godot-dap-mcp-server/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o $BUILD_DIR/godot-dap-mcp-server-windows-amd64.exe cmd/godot-dap-mcp-server/main.go

echo "Build complete: version $VERSION"
ls -lh $BUILD_DIR/
```

---

### Claude Code Integration

#### Installation Methods

Unlike npm or Python MCP servers that can be installed on-the-fly, Go binary MCP servers require a two-step process: obtaining the binary, then registering it with Claude Code.

**Prerequisites**:
- Godot 4.0+ with DAP enabled
- Claude Code CLI installed

---

**Option A: Install from GitHub Releases (Recommended)**

```bash
# 1. Create directory for MCP servers
mkdir -p ~/.claude/mcp-servers/godot-dap/

# 2. Download the appropriate binary for your platform
# macOS (Apple Silicon)
curl -L -o ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server \
  https://github.com/username/godot-dap-mcp-server/releases/download/v1.0.0/godot-dap-mcp-server-darwin-arm64

# macOS (Intel)
curl -L -o ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server \
  https://github.com/username/godot-dap-mcp-server/releases/download/v1.0.0/godot-dap-mcp-server-darwin-amd64

# Linux
curl -L -o ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server \
  https://github.com/username/godot-dap-mcp-server/releases/download/v1.0.0/godot-dap-mcp-server-linux-amd64

# 3. Make binary executable
chmod +x ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server

# 4. Register with Claude Code
claude mcp add godot-dap ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server
```

---

**Option B: Build from Source**

```bash
# 1. Clone repository
git clone https://github.com/username/godot-dap-mcp-server.git
cd godot-dap-mcp-server

# 2. Build binary
go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go

# 3. Create installation directory and move binary
mkdir -p ~/.claude/mcp-servers/godot-dap/
mv godot-dap-mcp-server ~/.claude/mcp-servers/godot-dap/

# 4. Register with Claude Code
claude mcp add godot-dap ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server
```

---

**Option C: System-Wide Installation**

```bash
# 1. Download or build binary (see above)

# 2. Move to system location
sudo mv godot-dap-mcp-server /usr/local/bin/

# 3. Register with Claude Code (no path needed - it's in PATH)
claude mcp add godot-dap /usr/local/bin/godot-dap-mcp-server
```

---

**Verify Installation**

```bash
# Check MCP server is registered
claude mcp list

# Test the server can be spawned
~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server
# Should start and wait for stdin input (Ctrl+C to exit)
```

---

**Manual Configuration** (if not using `claude mcp add`)

Edit `~/.claude/mcp.json`:
```json
{
  "mcpServers": {
    "godot-dap": {
      "command": "/Users/username/.claude/mcp-servers/godot-dap/godot-dap-mcp-server",
      "args": [],
      "env": {
        "GODOT_DAP_DEBUG": "false"
      }
    }
  }
}
```

Replace `/Users/username/` with your actual home directory path, or use the full path where you placed the binary.

---

### GitHub Releases

**Automated Releases** (GitHub Actions):
```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build binaries
        run: ./scripts/build.sh

      - name: Create release
        uses: softprops/action-gh-release@v1
        with:
          files: build/*
          body_path: CHANGELOG.md
```

---

## Documentation Plan

### README.md Structure

```markdown
# Godot DAP MCP Server

Interactive runtime debugging for Godot games via MCP protocol.

## Features
- Set breakpoints and pause execution
- Step through code (step-over, step-in, step-out)
- Inspect variables and call stacks
- Evaluate GDScript expressions
- Launch scenes programmatically
- Works with Claude Code and other MCP clients

## Installation
[Download binary, setup instructions]

## Quick Start
[Simple example: connect, set breakpoint, launch, inspect]

## Documentation
- [Tool Reference](docs/TOOLS.md)
- [Examples](docs/EXAMPLES.md)
- [Architecture](docs/ARCHITECTURE.md)

## Requirements
- Godot 4.0+ with DAP enabled
- MCP-compatible client (Claude Code, etc.)

## License
MIT
```

---

## Success Metrics

### Functional Metrics
- ✅ All core debugging tools working (connect, breakpoints, stepping, inspection)
- ✅ Launch functionality working (main, custom, current scenes)
- ✅ No permanent hangs (timeout mechanisms working)
- ✅ step-out issue resolved
- ✅ Clear error messages on failure

### Performance Metrics
- Startup time: <100ms
- Connection time: <1s
- Breakpoint hit latency: <500ms
- Step command latency: <500ms
- Binary size: <10MB

### User Experience Metrics
- AI agents can debug without human intervention
- Clear tool descriptions for LLM understanding
- Helpful error messages suggest fixes
- Works on first try (zero configuration)

### Community Metrics
- GitHub stars: 100+ (first month)
- Issues resolved: <48 hours
- Documentation completeness: 90%+
- Example workflows: 5+ scenarios

---

## Timeline Summary

| Phase | Days | Priority | Deliverable |
|-------|------|----------|-------------|
| 1. MCP Server Core | 1 | CRITICAL | stdio server with test tool |
| 2. DAP Client Migration | 1-2 | CRITICAL | Port working code + timeouts |
| 3. Core Debugging Tools | 1 | HIGH | 8 essential tools |
| 4. Inspection Tools | 1 | HIGH | 5 inspection tools |
| 5. Launch Tools | 1 | MEDIUM | 3 launch variants |
| 6. Advanced Tools | 1 | OPTIONAL | 4 nice-to-have tools |
| 7. Error Handling | 1 | CRITICAL | Timeouts, recovery, step-out fix |
| 8. Documentation | 1 | HIGH | Complete docs |
| **Total** | **4-5 days** | | **Production-ready server** |

---

## Next Steps

### Immediate Actions

1. **Create GitHub Repository**:
   ```bash
   mkdir godot-dap-mcp-server
   cd godot-dap-mcp-server
   git init
   go mod init github.com/yourusername/godot-dap-mcp-server
   ```

2. **Set Up Project Structure**:
   - Create directory layout
   - Add `.gitignore`, `LICENSE`, `README.md`
   - Initialize Go modules

3. **Phase 1 Implementation**:
   - Implement `internal/mcp/server.go`
   - Create test tool
   - Test with manual stdio input

4. **Phase 2 Implementation**:
   - Implement DAP client with proven patterns
   - Add timeout wrappers
   - Test connection to Godot

### Questions to Answer

1. **Repository location**: GitHub username?
2. **License**: MIT or Apache 2.0?
3. **Versioning**: Start at v0.1.0 or v1.0.0?
4. **Logging**: Include debug logging or minimal?

---

## Appendix A: Validated Patterns Reference

### From Previous DAP Experiments

**Proven Working** (11/13 commands):
```
✅ connect-to-debugger     (src: connectToDebugger())
✅ configuration-done      (src: configurationDone())
✅ set-breakpoints         (src: setBreakpoints())
✅ threads                 (src: listThreads())
✅ stack-trace             (src: getStackTrace())
✅ scopes                  (src: getScopes())
✅ variables               (src: getVariables())
✅ evaluate                (src: evaluateExpression())
✅ continue                (src: continue())
✅ next                    (src: next())
✅ step-in                 (src: stepIn())
```

**Blocking Issues**:
```
❌ step-out                (hangs permanently)
⚠️ pause                   (untested)
```

**Refactoring Patterns** (to reuse):
```go
// Pattern: Event filtering with infinite loop
for {
    msg, err := conn.ReadMessage()
    if err != nil {
        return err
    }

    switch m := msg.(type) {
    case *dap.ExpectedResponse:
        // Process response
        return nil
    default:
        // Skip events, continue loop
        continue
    }
}
```

---

## Appendix B: Godot DAP Protocol Details

### Launch Request Format

Based on Godot source (`debug_adapter_parser.cpp`):

```cpp
Dictionary args = p_params["arguments"];

// Required
String project = args["project"];  // Validated path

// Optional
String scene = args.get("scene", "main");  // "main" | "current" | path
String platform = args.get("platform", "host");  // "host" | "android" | "web"
bool noDebug = args.get("noDebug", false);

// Godot extension
bool customData = args.get("godot/custom_data", false);

// Additional play arguments (extracted separately)
Array additionalArgs = args.get("additional_arguments", []);
```

### Scene Launch Behavior

```cpp
if (scene == "main") {
    EditorRunBar::get_singleton()->play_main_scene(false, play_args);
} else if (scene == "current") {
    EditorRunBar::get_singleton()->play_current_scene(false, play_args);
} else {
    EditorRunBar::get_singleton()->play_custom_scene(scene, play_args);
}
```

---

## Appendix C: Tool Naming Convention

**Pattern**: `godot_<action>_<object>`

**Examples**:
- `godot_connect` (not `connect_to_godot`)
- `godot_set_breakpoint` (not `set_godot_breakpoint`)
- `godot_launch_main_scene` (clear hierarchy)
- `godot_get_variables` (consistent verb prefix)

**Rationale**:
- AI agents easily recognize Godot-specific tools
- Consistent verb-object pattern
- Namespace prevents conflicts with other MCP servers

---

## Appendix D: Error Message Guidelines

**Pattern**: Problem + Context + Solution

**Bad**:
```
Error: connection failed
```

**Good**:
```
Error: Failed to connect to Godot DAP server at localhost:6006

Possible causes:
1. Godot editor is not running
2. DAP server is not enabled in editor settings
3. DAP server is using a different port

Solutions:
1. Launch Godot editor
2. Enable DAP in Editor → Editor Settings → Network → Debug Adapter
3. Check port setting (default: 6006)

For more help, see: https://docs.godotengine.org/en/stable/tutorials/editor/debugger_panel.html
```

---

**End of Implementation Plan**

---

**Ready to Start?**

This plan provides everything needed to build the `godot-dap-mcp-server`:
- ✅ Clear architecture and design
- ✅ Detailed implementation phases
- ✅ Code examples and patterns
- ✅ Testing strategy
- ✅ Deployment approach
- ✅ Success metrics

**Estimated Timeline**: 4-5 days for production-ready server

**Next Step**: Create GitHub repository and begin Phase 1 (MCP Server Core)
