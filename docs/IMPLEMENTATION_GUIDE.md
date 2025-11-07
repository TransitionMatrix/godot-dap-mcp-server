# Godot DAP MCP Server - Implementation Guide

**Last Updated**: 2025-11-07

This document provides detailed component specifications and implementation patterns for the godot-dap-mcp-server. Use this as a reference when implementing or extending the server.

---

## Table of Contents

1. [MCP Server Layer](#mcp-server-layer)
2. [DAP Client Layer](#dap-client-layer)
3. [Tool Layer](#tool-layer)
4. [Tool Description Guidelines](#tool-description-guidelines)
5. [Adding New Components](#adding-new-components)

---

## MCP Server Layer

Location: `internal/mcp/`

**Purpose**: Handle stdio-based MCP protocol communication (JSONRPC 2.0 over stdin/stdout).

### server.go

Core MCP server implementation with request routing:

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

### types.go

MCP protocol type definitions:

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

## DAP Client Layer

Location: `internal/dap/`

**Purpose**: Communicate with Godot's DAP server via TCP. Implements timeout protection and event filtering.

### client.go

Core DAP client with TCP connection management:

```go
package dap

import (
    "bufio"
    "context"
    "fmt"
    "net"
    "sync"
    "time"

    "github.com/google/go-dap"
)

type Client struct {
    host   string
    port   int
    conn   net.Conn
    reader *bufio.Reader
    codec  *dap.Codec

    mu      sync.Mutex
    nextSeq int

    connected bool
}

func NewClient(host string, port int) *Client {
    return &Client{
        host:    host,
        port:    port,
        codec:   dap.NewCodec(),
        nextSeq: 1,
    }
}

func (c *Client) Connect(ctx context.Context) error {
    addr := fmt.Sprintf("%s:%d", c.host, c.port)

    conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", addr)
    if err != nil {
        return fmt.Errorf("failed to connect to %s: %w", addr, err)
    }

    c.conn = conn
    c.reader = bufio.NewReader(conn)
    c.connected = true

    return nil
}

func (c *Client) Disconnect() error {
    if c.conn == nil {
        return nil
    }
    c.connected = false
    return c.conn.Close()
}

func (c *Client) nextRequestSeq() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    seq := c.nextSeq
    c.nextSeq++
    return seq
}

// read reads a DAP message from the connection
func (c *Client) read() (dap.Message, error) {
    if !c.connected {
        return nil, fmt.Errorf("not connected")
    }

    // Read base message
    data, err := dap.ReadBaseMessage(c.reader)
    if err != nil {
        return nil, fmt.Errorf("failed to read base message: %w", err)
    }

    // Decode message
    msg, err := c.codec.DecodeMessage(data)
    if err != nil {
        return nil, fmt.Errorf("failed to decode message: %w", err)
    }

    return msg, nil
}

// write sends a DAP message to the connection
func (c *Client) write(msg dap.Message) error {
    if !c.connected {
        return fmt.Errorf("not connected")
    }
    return dap.WriteProtocolMessage(c.conn, msg)
}

func (c *Client) Initialize(ctx context.Context) (*dap.InitializeResponse, error) {
    request := &dap.InitializeRequest{
        Request: dap.Request{
            ProtocolMessage: dap.ProtocolMessage{
                Seq:  c.nextRequestSeq(),
                Type: "request",
            },
            Command: "initialize",
        },
        Arguments: dap.InitializeRequestArguments{
            ClientID:       "godot-dap-mcp-server",
            AdapterID:      "godot",
            LinesStartAt1:  true,
            ColumnsStartAt1: true,
        },
    }

    if err := c.write(request); err != nil {
        return nil, fmt.Errorf("failed to write initialize request: %w", err)
    }

    msg, err := c.waitForResponse(ctx, "initialize")
    if err != nil {
        return nil, err
    }

    initResp, ok := msg.(*dap.InitializeResponse)
    if !ok {
        return nil, fmt.Errorf("unexpected response type: %T", msg)
    }

    return initResp, nil
}

func (c *Client) ConfigurationDone(ctx context.Context) error {
    request := &dap.ConfigurationDoneRequest{
        Request: dap.Request{
            ProtocolMessage: dap.ProtocolMessage{
                Seq:  c.nextRequestSeq(),
                Type: "request",
            },
            Command: "configurationDone",
        },
    }

    if err := c.write(request); err != nil {
        return fmt.Errorf("failed to write configurationDone request: %w", err)
    }

    _, err := c.waitForResponse(ctx, "configurationDone")
    return err
}

// SetBreakpoints sets breakpoints for a file
func (c *Client) SetBreakpoints(ctx context.Context, file string, lines []int) error {
    breakpoints := make([]dap.SourceBreakpoint, len(lines))
    for i, line := range lines {
        breakpoints[i] = dap.SourceBreakpoint{Line: line}
    }

    request := &dap.SetBreakpointsRequest{
        Request: dap.Request{
            ProtocolMessage: dap.ProtocolMessage{
                Seq:  c.nextRequestSeq(),
                Type: "request",
            },
            Command: "setBreakpoints",
        },
        Arguments: dap.SetBreakpointsArguments{
            Source:      dap.Source{Path: file},
            Breakpoints: breakpoints,
        },
    }

    if err := c.write(request); err != nil {
        return fmt.Errorf("failed to write setBreakpoints request: %w", err)
    }

    _, err := c.waitForResponse(ctx, "setBreakpoints")
    return err
}

// Additional DAP commands: Continue, Next, StepIn, Threads, StackTrace,
// Scopes, Variables, Evaluate, etc.
```

### timeout.go

Timeout wrappers for all DAP operations:

```go
package dap

import (
    "context"
    "fmt"
    "time"

    "github.com/google/go-dap"
)

const (
    DefaultConnectTimeout  = 10 * time.Second
    DefaultCommandTimeout  = 30 * time.Second
    DefaultReadTimeout     = 5 * time.Second
)

// ReadWithTimeout reads a DAP message with timeout protection
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

// Helper functions for creating contexts with standard timeouts
func WithConnectTimeout(parent context.Context) (context.Context, context.CancelFunc) {
    if parent == nil {
        parent = context.Background()
    }
    return context.WithTimeout(parent, DefaultConnectTimeout)
}

func WithCommandTimeout(parent context.Context) (context.Context, context.CancelFunc) {
    if parent == nil {
        parent = context.Background()
    }
    return context.WithTimeout(parent, DefaultCommandTimeout)
}

func WithReadTimeout(parent context.Context) (context.Context, context.CancelFunc) {
    if parent == nil {
        parent = context.Background()
    }
    return context.WithTimeout(parent, DefaultReadTimeout)
}

func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
    if parent == nil {
        parent = context.Background()
    }
    return context.WithTimeout(parent, timeout)
}
```

### events.go

Event filtering pattern - filters async events from command responses:

```go
package dap

import (
    "context"
    "fmt"
    "log"

    "github.com/google/go-dap"
)

// waitForResponse reads messages until expected response arrives
// This is CRITICAL - filters out async events to prevent hangs
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
            log.Printf("Received unexpected response for command: %s (waiting for %s)",
                m.Command, command)
            continue

        case *dap.Event:
            c.logEvent(m)
            continue

        case *dap.InitializeResponse:
            if command == "initialize" {
                return m, nil
            }
            continue

        case *dap.LaunchResponse:
            if command == "launch" {
                return m, nil
            }
            continue

        case *dap.ConfigurationDoneResponse:
            if command == "configurationDone" {
                return m, nil
            }
            continue

        case *dap.SetBreakpointsResponse:
            if command == "setBreakpoints" {
                return m, nil
            }
            continue

        case *dap.ContinueResponse:
            if command == "continue" {
                return m, nil
            }
            continue

        case *dap.NextResponse:
            if command == "next" {
                return m, nil
            }
            continue

        case *dap.StepInResponse:
            if command == "stepIn" {
                return m, nil
            }
            continue

        case *dap.ThreadsResponse:
            if command == "threads" {
                return m, nil
            }
            continue

        case *dap.StackTraceResponse:
            if command == "stackTrace" {
                return m, nil
            }
            continue

        case *dap.ScopesResponse:
            if command == "scopes" {
                return m, nil
            }
            continue

        case *dap.VariablesResponse:
            if command == "variables" {
                return m, nil
            }
            continue

        case *dap.EvaluateResponse:
            if command == "evaluate" {
                return m, nil
            }
            continue

        default:
            log.Printf("Received unknown message type: %T", m)
            continue
        }
    }
}

func (c *Client) logEvent(event *dap.Event) {
    log.Printf("DAP Event: %s", event.Event)
    // Could emit these to MCP client as notifications
}
```

### session.go

DAP session lifecycle management with state machine:

```go
package dap

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/google/go-dap"
)

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

func NewSession(client *Client) *Session {
    return &Session{
        client: client,
        state:  StateDisconnected,
    }
}

func (s *Session) Connect(ctx context.Context) error {
    if s.state != StateDisconnected {
        return fmt.Errorf("already connected")
    }

    if err := s.client.Connect(ctx); err != nil {
        return err
    }

    s.state = StateConnected
    return nil
}

func (s *Session) Initialize(ctx context.Context) (*dap.InitializeResponse, error) {
    if s.state != StateConnected {
        return nil, fmt.Errorf("not connected")
    }

    resp, err := s.client.Initialize(ctx)
    if err != nil {
        return nil, err
    }

    s.state = StateInitialized
    return resp, nil
}

func (s *Session) ConfigurationDone(ctx context.Context) error {
    if s.state != StateInitialized {
        return fmt.Errorf("not initialized")
    }

    if err := s.client.ConfigurationDone(ctx); err != nil {
        return err
    }

    s.state = StateConfigured
    return nil
}

func (s *Session) InitializeSession(ctx context.Context) error {
    if err := s.Connect(ctx); err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }

    if _, err := s.Initialize(ctx); err != nil {
        s.client.Disconnect()
        s.state = StateDisconnected
        return fmt.Errorf("failed to initialize: %w", err)
    }

    if err := s.ConfigurationDone(ctx); err != nil {
        s.client.Disconnect()
        s.state = StateDisconnected
        return fmt.Errorf("failed to send configurationDone: %w", err)
    }

    return nil
}

func (s *Session) Launch(ctx context.Context, args map[string]interface{}) (*dap.LaunchResponse, error) {
    if s.state != StateConfigured {
        return nil, fmt.Errorf("session not configured")
    }

    // Marshal args to JSON as required by LaunchRequest.Arguments
    argsJSON, err := json.Marshal(args)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal launch arguments: %w", err)
    }

    request := &dap.LaunchRequest{
        Request: dap.Request{
            ProtocolMessage: dap.ProtocolMessage{
                Seq:  s.client.nextRequestSeq(),
                Type: "request",
            },
            Command: "launch",
        },
        Arguments: argsJSON,
    }

    if err := s.client.write(request); err != nil {
        return nil, fmt.Errorf("failed to write launch request: %w", err)
    }

    msg, err := s.client.waitForResponse(ctx, "launch")
    if err != nil {
        return nil, err
    }

    launchResp, ok := msg.(*dap.LaunchResponse)
    if !ok {
        return nil, fmt.Errorf("unexpected response type: %T", msg)
    }

    s.state = StateLaunched
    return launchResp, nil
}
```

### godot.go

Godot-specific DAP extensions and launch configurations:

```go
package dap

import (
    "fmt"
    "os"
    "path/filepath"
)

// SceneLaunchMode defines how to launch a scene
type SceneLaunchMode string

const (
    SceneLaunchMain    SceneLaunchMode = "main"
    SceneLaunchCurrent SceneLaunchMode = "current"
    SceneLaunchCustom  SceneLaunchMode = "custom"
)

// Platform defines the target platform
type Platform string

const (
    PlatformHost    Platform = "host"
    PlatformAndroid Platform = "android"
    PlatformWeb     Platform = "web"
)

// GodotLaunchConfig contains Godot-specific launch parameters
type GodotLaunchConfig struct {
    Project           string
    Scene             SceneLaunchMode
    ScenePath         string
    Platform          Platform
    NoDebug           bool
    Profiling         bool
    DebugCollisions   bool
    DebugPaths        bool
    DebugNavigation   bool
    AdditionalOptions string
}

// Validate checks if the launch configuration is valid
func (c *GodotLaunchConfig) Validate() error {
    if c.Project == "" {
        return fmt.Errorf("project path is required")
    }

    // Validate project.godot exists
    projectFile := filepath.Join(c.Project, "project.godot")
    if _, err := os.Stat(projectFile); err != nil {
        if os.IsNotExist(err) {
            return fmt.Errorf("project.godot not found in %s", c.Project)
        }
        return fmt.Errorf("failed to check project.godot: %w", err)
    }

    // Validate scene path if custom scene
    if c.Scene == SceneLaunchCustom && c.ScenePath == "" {
        return fmt.Errorf("scene path is required when using custom scene launch mode")
    }

    return nil
}

// ToLaunchArgs converts the config to DAP launch arguments
func (c *GodotLaunchConfig) ToLaunchArgs() map[string]interface{} {
    args := map[string]interface{}{
        "project":  c.Project,
        "platform": string(c.Platform),
        "noDebug":  c.NoDebug,
    }

    // Set scene parameter
    if c.Scene == SceneLaunchCustom {
        args["scene"] = c.ScenePath
    } else {
        args["scene"] = string(c.Scene)
    }

    // Add optional parameters
    if c.Profiling {
        args["profiling"] = true
    }
    if c.DebugCollisions {
        args["debug_collisions"] = true
    }
    if c.DebugPaths {
        args["debug_paths"] = true
    }
    if c.DebugNavigation {
        args["debug_navigation"] = true
    }
    if c.AdditionalOptions != "" {
        args["additional_options"] = c.AdditionalOptions
    }

    return args
}

// Helper functions for common launch scenarios
func LaunchMainScene(project string) *GodotLaunchConfig {
    return &GodotLaunchConfig{
        Project:  project,
        Scene:    SceneLaunchMain,
        Platform: PlatformHost,
    }
}

func LaunchCurrentScene(project string) *GodotLaunchConfig {
    return &GodotLaunchConfig{
        Project:  project,
        Scene:    SceneLaunchCurrent,
        Platform: PlatformHost,
    }
}

func LaunchCustomScene(project, scenePath string) *GodotLaunchConfig {
    return &GodotLaunchConfig{
        Project:   project,
        Scene:     SceneLaunchCustom,
        ScenePath: scenePath,
        Platform:  PlatformHost,
    }
}
```

---

## Tool Layer

Location: `internal/tools/`

**Purpose**: Implement Godot-specific MCP tools that wrap DAP functionality with AI-optimized descriptions.

### Example: connect.go

Connection management tool:

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

### Example: launch.go

Scene launching tools with comprehensive descriptions:

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

    // Additional launch tools: godot_launch_scene, godot_launch_current_scene
}
```

---

## Tool Description Guidelines

When implementing MCP tools, follow this pattern for descriptions:

### Structure

1. **Summary** (1 sentence): What the tool does
2. **Prerequisites** (if any): What must be true before calling
3. **Behavior** (1-2 sentences): What happens when called
4. **Use Cases** (3-5 bullets): When to use this tool
5. **Example** (1-2 lines): Concrete usage example

### Example Pattern

```go
Description: `[Summary - one sentence describing what it does]

Prerequisites:
- [Prerequisite 1]
- [Prerequisite 2]

[Behavior - what happens when called]

Use this when you want to:
- [Use case 1]
- [Use case 2]
- [Use case 3]

Example: [Concrete example with actual parameters]
tool_name(param1="value", param2=123)`,
```

### Error Message Pattern

Follow: **Problem + Context + Solution**

```go
return nil, fmt.Errorf(`Failed to connect to Godot DAP server at localhost:%d

Possible causes:
1. Godot editor is not running
2. DAP server is not enabled in editor settings
3. DAP server is using a different port

Solutions:
1. Launch Godot editor
2. Enable DAP in Editor → Editor Settings → Network → Debug Adapter
3. Check port setting (default: 6006)`, port)
```

---

## Adding New Components

### Adding a New MCP Tool

1. Create file in `internal/tools/` (e.g., `pause.go`)
2. Implement tool with proper description pattern
3. Register in `registry.go`
4. Add tests in `internal/tools/*_test.go`
5. Document in user-facing docs
6. Add to troubleshooting FAQ if needed

### Adding a New DAP Command

1. Add method to `internal/dap/client.go`
2. Add timeout wrapper if blocking
3. Add response type to event filter in `events.go`
4. Add tests in `internal/dap/dap_test.go`
5. Document any Godot-specific behavior

### Extending Godot Launch Options

1. Update `GodotLaunchConfig` in `internal/dap/godot.go`
2. Update `ToLaunchArgs()` method
3. Update `Validate()` if new validations needed
4. Document in [DAP_PROTOCOL.md](reference/DAP_PROTOCOL.md)
5. Update tool descriptions that use launch

---

## References

- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture and design decisions
- [GODOT_DAP_FAQ.md](reference/GODOT_DAP_FAQ.md) - Common troubleshooting questions
- [DAP_PROTOCOL.md](reference/DAP_PROTOCOL.md) - Detailed protocol specifications
- [CONVENTIONS.md](reference/CONVENTIONS.md) - Naming and coding conventions
