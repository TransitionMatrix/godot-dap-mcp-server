# Godot DAP MCP Server - Implementation Guide

**Last Updated**: 2025-11-08

This document provides detailed component specifications and implementation patterns for the godot-dap-mcp-server. Use this as a reference when implementing or extending the server.

---

## Table of Contents

1. [MCP Server Layer](#mcp-server-layer)
2. [DAP Client Layer](#dap-client-layer)
3. [Tool Layer](#tool-layer)
   - [Formatting Godot Types](#formatting-godot-types)
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

func (c *Client) Attach(ctx context.Context) (*dap.AttachResponse, error) {
    request := &dap.AttachRequest{
        Request: dap.Request{
            ProtocolMessage: dap.ProtocolMessage{
                Seq:  c.nextRequestSeq(),
                Type: "request",
            },
            Command: "attach",
        },
        Arguments: map[string]interface{}{}, // No arguments required for Godot attach
    }

    if err := c.write(request); err != nil {
        return nil, fmt.Errorf("failed to write attach request: %w", err)
    }

    msg, err := c.waitForResponse(ctx, "attach")
    if err != nil {
        return nil, err
    }

    resp, ok := msg.(*dap.AttachResponse)
    if !ok {
        return nil, fmt.Errorf("unexpected response type: %T", msg)
    }

    return resp, nil
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

### Example: attach.go

Attach tool implementation:

```go
package tools

import (
    "fmt"
    "context"

    "your-repo/internal/dap"
    "your-repo/internal/mcp"
)

func RegisterAttachTool(server *mcp.Server, dapClient *dap.Client) {
    server.RegisterTool(mcp.Tool{
        Name: "godot_attach",
        Description: `Attach the debugger to an already running Godot game instance.

        This tool connects to a game that is already running and waiting for a debugger.
        The game must have been started with debugging enabled and configured to connect
        to the editor's port (usually 6007).

        Prerequisites:
        - Must be connected to Godot DAP server (call godot_connect first)
        - Must complete DAP configuration handshake
        - Game must be running and attempting to connect to the editor

        Use this tool:
        - When you want to debug a game that was launched externally
        - When you want to attach to a game running on a device
        - As an alternative to launching the game through the DAP server

        Attach Flow:
        1. Sends attach request
        2. Sends configurationDone
        3. Debugger attaches to the running game session

        Example: Attach to running game
        godot_attach()`,

        Parameters: []mcp.Parameter{}, // No parameters required

        Handler: func(params map[string]interface{}) (interface{}, error) {
            // Check connection state
            session, err := GetSession()
            if err != nil {
                return nil, err
            }

            // Perform attach
            if err := session.Attach(context.Background()); err != nil {
                 return nil, fmt.Errorf("failed to attach: %w", err)
            }

            return map[string]interface{}{
                "status": "attached",
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

### Formatting Godot Types

Location: `internal/tools/formatting.go`

**Purpose**: Enhance DAP variable responses with Godot-specific semantic formatting for better AI readability.

When implementing tools that return Godot variables (like `godot_get_variables` or `godot_evaluate`), use the formatting utilities to add human-readable representations of Godot types:

```go
// Format a single variable
func formatVariable(variable dap.Variable) map[string]interface{} {
    result := map[string]interface{}{
        "name":  variable.Name,
        "value": variable.Value,
        "type":  variable.Type,
    }

    // Add formatted version if it's a Godot type
    if formatted := formatGodotType(variable.Type, variable.Value); formatted != "" {
        result["formatted"] = formatted
    }

    return result
}

// Format a list of variables
variables := formatVariableList(resp.Body.Variables)
```

**Automatically formatted types:**
- **Vectors**: `Vector2(x=10, y=20)`, `Vector3(x=1, y=2, z=3)`
- **Colors**: `Color(r=1.0, g=0.5, b=0.0, a=1.0)`
- **Bounding boxes**: `Rect2(pos=(10, 20), size=(100, 50))`
- **Nodes**: `CharacterBody2D (ID:456)`
- **Collections**: `Array(8): [1, 2, 3, ...]`, `Dictionary(5 keys)`

**Example output:**
```json
{
  "name": "position",
  "value": "(100.5, 200.3)",
  "type": "Vector2",
  "formatted": "Vector2(x=100.5, y=200.3)"
}
```

The `formatted` field is optional and only added when a Godot-specific type is detected. Original `value` and `type` fields are always preserved.

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

1. **Verify DAP specification** - Check `docs/reference/debugAdapterProtocol.json` for the request definition
   ```bash
   jq '.definitions.YourRequestName' docs/reference/debugAdapterProtocol.json
   ```
   - Identify required vs optional fields in the `required` array
   - Note any default values or special handling needed

2. **Consult Godot implementation** - Check `godot-source` for Godot's behavior
   ```bash
   mcp__godot-source__find_symbol("req_yourCommand", relative_path="editor/debugger/debug_adapter")
   ```

3. Create file in `internal/tools/` (e.g., `pause.go`)

4. Implement tool with proper description pattern
   - Document required vs optional fields from DAP spec
   - Use safe `.get()` access for all optional fields
   - Include example showing minimal required fields

5. Register in `registry.go`

6. Add tests in `internal/tools/*_test.go`
   - Test with minimal required fields
   - Test with optional fields omitted

7. Document in user-facing docs

8. Add to troubleshooting FAQ if needed

### Adding a New DAP Command

1. **Verify DAP specification** - Check protocol requirements first
   ```bash
   # Check request definition
   jq '.definitions.YourRequest' docs/reference/debugAdapterProtocol.json

   # Check response definition
   jq '.definitions.YourResponse' docs/reference/debugAdapterProtocol.json
   ```

2. Add method to `internal/dap/client.go`
   - Use safe field access (`.get()` for optional fields)
   - Follow required field specifications from DAP spec

3. Add timeout wrapper if blocking

4. Add response type to event filter in `events.go`

5. Add tests in `internal/dap/dap_test.go`
   - Test with minimal required fields per spec
   - Test with optional fields omitted

6. Document any Godot-specific behavior

### Extending Godot Launch Options

1. Update `GodotLaunchConfig` in `internal/dap/godot.go`
2. Update `ToLaunchArgs()` method
3. Update `Validate()` if new validations needed
4. Document in [DAP_PROTOCOL.md](reference/DAP_PROTOCOL.md)
5. Update tool descriptions that use launch

---

---

## Integration Testing Patterns

These patterns were discovered while implementing Phase 3 integration tests. For the full debugging story, see [LESSONS_LEARNED_PHASE_3.md](LESSONS_LEARNED_PHASE_3.md).

### Pattern: Persistent Subprocess Communication

When testing MCP servers, you need to maintain a persistent server process across multiple tool calls (for session state). Use named pipes with file descriptors:

```bash
#!/bin/bash

# 1. Create named pipes
rm -f /tmp/mcp-stdin /tmp/mcp-stdout
mkfifo /tmp/mcp-stdin /tmp/mcp-stdout

# 2. Open file descriptors (keeps pipes alive)
exec 3<>/tmp/mcp-stdin
exec 4<>/tmp/mcp-stdout

# 3. Start subprocess with pipes
./godot-dap-mcp-server </tmp/mcp-stdin >/tmp/mcp-stdout 2>/dev/null &
SERVER_PID=$!

# 4. Setup cleanup trap
trap "exec 3>&- 4>&-; kill $SERVER_PID 2>/dev/null; rm -f /tmp/mcp-stdin /tmp/mcp-stdout" EXIT

# 5. Communication function
send_mcp_request() {
    local request="$1"
    echo "$request" >&3  # Write to FD 3
    read -r response <&4  # Read from FD 4
    echo "$response"
}

# 6. Use it
RESPONSE=$(send_mcp_request '{"jsonrpc":"2.0","id":1,"method":"tools/list"}')
```

**Why This Works:**
- File descriptors (FD 3, 4) keep pipes open even after individual writes
- Server's stdin never sees EOF
- Session persists across all tool calls
- Clean shutdown via trap

**When to Use:**
- Testing MCP servers with session state (like `globalSession`)
- Any subprocess that needs to maintain state across requests
- Request/response protocols over stdin/stdout

**Common Mistake:**
```bash
# ❌ WRONG - Pipe closes after each write
echo "$data" > /tmp/pipe

# ✅ RIGHT - Use file descriptors
exec 3<>/tmp/pipe
echo "$data" >&3
```

### Pattern: Port Conflict Handling

Development tools often run multiple instances. Use graceful port fallback:

```bash
# Check if port is in use
is_port_in_use() {
    nc -z 127.0.0.1 $1 2>/dev/null
}

# Find available port in range
find_available_port() {
    local start=$1
    local end=$2
    for port in $(seq $start $end); do
        if ! is_port_in_use $port; then
            echo $port
            return 0
        fi
    done
    return 1
}

# Use with fallback
DEFAULT_PORT=6006
if is_port_in_use $DEFAULT_PORT; then
    PORT=$(find_available_port 6006 6020)
    if [ -z "$PORT" ]; then
        echo "Error: No available ports in range 6006-6020"
        exit 1
    fi
    echo "Port $DEFAULT_PORT in use, using $PORT instead"
else
    PORT=$DEFAULT_PORT
fi

echo "Using port: $PORT"
```

**When to Use:**
- Development tools with default ports
- CI/CD environments where port conflicts are common
- Any networked service that may have multiple instances

### Pattern: Robust JSON Parsing in Bash

MCP wraps tool results in nested JSON, causing escaped quotes. Use regex that handles both:

```bash
# Match both escaped and unescaped JSON
PATTERN='(\\"|")status(\\"|"):(\\"|")connected'

if echo "$RESPONSE" | grep -qE "$PATTERN"; then
    echo "Connection successful!"
fi

# Extract values (handles escaped quotes)
extract_status() {
    local response="$1"
    echo "$response" | grep -oE '(\\"|")status(\\"|"):(\\"|")[^"\\]*' |
        sed 's/.*://' | tr -d '"\\'
}

STATUS=$(extract_status "$RESPONSE")
```

**Why This Is Needed:**

MCP protocol wraps tool results in nested JSON:
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [{
      "type": "text",
      "text": "{\\\"status\\\":\\\"connected\\\"}"  // Escaped JSON string
    }]
  }
}
```

**Pattern Breakdown:**
- `(\\"|")` - Matches either `\"` (escaped) or `"` (unescaped)
- Applied to both key and value quotes
- Works with MCP's nested JSON structure

**When to Use:**
- Parsing MCP tool results in shell scripts
- Any situation where JSON might be escaped in a string
- Integration tests that parse structured responses

---

## References

- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture and design decisions
- [GODOT_DAP_FAQ.md](reference/GODOT_DAP_FAQ.md) - Common troubleshooting questions
- [DAP_PROTOCOL.md](reference/DAP_PROTOCOL.md) - Detailed protocol specifications
- [CONVENTIONS.md](reference/CONVENTIONS.md) - Naming and coding conventions
- [LESSONS_LEARNED_PHASE_3.md](LESSONS_LEARNED_PHASE_3.md) - Phase 3 debugging case study
