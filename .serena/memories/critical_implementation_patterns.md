# Critical Implementation Patterns

These are essential patterns discovered through testing and implementation. Following these patterns is critical for correctness.

## 1. Event Filtering Pattern (DAP Client)

**Problem**: Godot's DAP server sends async events (stopped, continued, output, etc.) mixed with command responses. Reading without filtering will return events instead of expected responses.

**Solution**: Always filter for expected response type
```go
func (c *Client) waitForResponse(ctx context.Context, expectedCommand string) (*dap.Response, error) {
    for {
        msg, err := c.ReadWithTimeout(ctx)
        if err != nil {
            return nil, err
        }

        switch m := msg.(type) {
        case *dap.Response:
            if m.Command == expectedCommand {
                return m, nil
            }
            // Wrong response, keep waiting
        case *dap.Event:
            // Log event but don't return - continue waiting
            c.logEvent(m)
        default:
            // Unknown message type, continue
            continue
        }
    }
}
```

**Key Points**:
- Never return on first message - it might be an event
- Loop until expected response arrives
- Log events but don't treat them as responses
- Use timeouts to prevent infinite loops

## 2. Timeout Protection Pattern

**Problem**: DAP server may hang or not respond to certain requests, causing permanent blocking.

**Solution**: Wrap all DAP operations with context timeouts
```go
func (t *Tool) Execute(args map[string]interface{}) (interface{}, error) {
    // Create timeout context (10-30 seconds depending on operation)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Use context for all DAP operations
    if err := client.SomeOperation(ctx); err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return nil, fmt.Errorf("operation timed out after 30 seconds")
        }
        return nil, err
    }

    return result, nil
}
```

**Timeout Guidelines**:
- Quick operations: 10 seconds (connect, breakpoint)
- Launch operations: 30 seconds (scene loading)
- Step operations: 15 seconds (may hit breakpoint)
- Read operations: 5 seconds (should be fast)

## 3. Godot Launch Flow Pattern

**Problem**: Godot requires specific sequence to launch game for debugging.

**Solution**: Follow exact initialization sequence
```go
// 1. Initialize DAP connection
err := client.Initialize(ctx)

// 2. Send launch request with arguments
err = client.Launch(ctx, LaunchConfig{
    Mode:    "launch",
    Request: "launch",
    Project: "/absolute/path/to/project",
    Scene:   "main",  // or "current" or "res://path/to/scene.tscn"
    // Optional:
    Platform:     "host",
    BreakOnStart: false,
})

// 3. Send configurationDone to trigger actual launch
err = client.ConfigurationDone(ctx)

// Game is now running until breakpoint/pause/exit
```

**Key Points**:
- Must call initialize before launch
- Must call configurationDone after launch
- Project path must be absolute and contain project.godot
- Scene modes: "main", "current", or explicit path

## 4. Error Message Pattern

**Pattern**: Problem + Context + Solution

```go
func validateProjectPath(path string) error {
    projectFile := filepath.Join(path, "project.godot")
    if _, err := os.Stat(projectFile); os.IsNotExist(err) {
        return fmt.Errorf(`Invalid project path: project.godot not found at %s

Possible causes:
1. Path does not point to a Godot project directory
2. The path is relative instead of absolute
3. The project.godot file has been moved or deleted

Solutions:
1. Ensure the path points to the directory containing project.godot
2. Use an absolute path: /full/path/to/project
3. Verify the project exists: ls %s`, path, projectFile)
    }
    return nil
}
```

**Key Points**:
- State the problem clearly
- List 2-4 possible causes
- Provide concrete solutions
- Include relevant paths/values

## 5. Session State Management

**Problem**: Only one DAP session can be active at a time. Must track connection and session state.

**Solution**: Centralized session state
```go
type Session struct {
    client      *Client
    connected   bool
    initialized bool
    launched    bool
    mu          sync.RWMutex
}

func (s *Session) RequireConnected() error {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    if !s.connected {
        return fmt.Errorf("not connected to Godot DAP server")
    }
    return nil
}

func (s *Session) RequireLaunched() error {
    if err := s.RequireConnected(); err != nil {
        return err
    }
    
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    if !s.launched {
        return fmt.Errorf("game not launched - use godot_launch_* first")
    }
    return nil
}
```

**Key Points**:
- Protect state with mutex
- Validate state before operations
- Clear error messages about required state
- Single active session at a time

## 6. Known Issues and Workarounds

### stepOut Not Implemented
```go
// ❌ This will hang - stepOut not implemented in Godot
err := client.StepOut(ctx)

// ✅ Workaround: Use continue or step-over instead
err := client.Continue(ctx)
```

### Breakpoint Verification
```go
// Always check if breakpoint was verified by Godot
response, err := client.SetBreakpoints(ctx, file, lines)
for _, bp := range response.Breakpoints {
    if !bp.Verified {
        log.Printf("Warning: breakpoint at line %d not verified", bp.Line)
    }
}
```

### Path Handling
```go
// Always use absolute paths for Godot
projectPath, err := filepath.Abs(relativePath)
scriptPath := filepath.Join(projectPath, "scripts/player.gd")
```

## 7. Tool Registration Pattern

```go
func RegisterTools(server *mcp.Server, session *dap.Session) {
    server.RegisterTool(&mcp.Tool{
        Name:        "godot_set_breakpoint",
        Description: `[AI-optimized description with prerequisites, use cases, examples]`,
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "file": map[string]interface{}{
                    "type":        "string",
                    "description": "Absolute path to script file",
                },
                "line": map[string]interface{}{
                    "type":        "integer",
                    "description": "Line number (1-indexed)",
                },
            },
            "required": []string{"file", "line"},
        },
        Handler: func(args map[string]interface{}) (interface{}, error) {
            // Validate session state
            if err := session.RequireLaunched(); err != nil {
                return nil, err
            }
            
            // Extract and validate parameters
            file, _ := args["file"].(string)
            line, _ := args["line"].(float64)  // JSON numbers are float64
            
            // Execute with timeout
            ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            defer cancel()
            
            return session.Client.SetBreakpoint(ctx, file, int(line))
        },
    })
}
```
