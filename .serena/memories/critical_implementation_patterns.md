# Critical Implementation Patterns

These are essential patterns discovered through testing and implementation. Following these patterns is critical for correctness.

## 1. DAP Protocol Handshake Pattern (Revised Phase 5)

**Problem**: Sending `ConfigurationDone` immediately after `Initialize` prevents `Launch` requests from working (protocol violation).

**Solution**: Defer `ConfigurationDone` until after the Launch request.

```go
// 1. Connect & Initialize (godot_connect)
session.Connect(ctx)
session.Initialize(ctx)
// State: Initialized (Waiting for configuration/launch)

// 2. Launch (godot_launch_*)
// Client sends Launch request
client.Send(LaunchRequest)

// 3. Signal Ready
// Client sends ConfigurationDone immediately after Launch
client.Send(ConfigurationDoneRequest)

// 4. Wait for Responses
// Godot sends LaunchResponse ONLY after receiving ConfigurationDone
// We must wait for both responses in any order
client.WaitForLaunchAndConfigDone()
// State: Launched (Game running)
```

## 2. Concurrent Tool Execution Pattern (Phase 8 Fix)

**Problem**: Single-threaded MCP servers hang completely if one tool (e.g., `godot_pause`) blocks waiting for an unresponsive upstream service (Godot). Subsequent requests queue up indefinitely.

**Solution**: Process MCP requests in goroutines and use thread-safe transport.

```go
// internal/mcp/server.go
go func(r *MCPRequest) {
    resp := s.handleRequest(r)
    s.transport.WriteResponse(resp) // Thread-safe write
}(req)

// internal/mcp/transport.go
func (t *Transport) WriteResponse(resp MCPResponse) {
    t.mu.Lock() // Critical: prevent interleaved JSON
    defer t.mu.Unlock()
    // ... write to stdout ...
}
```

**Key Points**:
- Never block the main read loop.
- Always protect stdout with a mutex.
- Allows "Cancel" signals or new connections to be processed while a tool is blocked.

## 3. Event and Response Handling Pattern

**Problem**: Godot's DAP server sends async events mixed with command responses. `go-dap` returns specific response types.

**Solution**: Loop and filter until the specific expected response type is received.

```go
func (c *Client) waitForResponse(ctx context.Context, expectedCommand string) (dap.Message, error) {
    for {
        msg, err := c.ReadWithTimeout(ctx)
        switch m := msg.(type) {
        case *dap.Response: // Generic
            if m.Command == expectedCommand { return m, nil }
        case *dap.Event:
            c.logEvent(m) // Log and continue
        case *dap.PauseResponse: // Explicit types required
            if expectedCommand == "pause" { return m, nil }
        // ... other types ...
        }
    }
}
```

## 4. Timeout Protection Pattern

**Problem**: DAP server may hang.

**Solution**: Wrap all DAP operations with context timeouts.

```go
ctx, cancel := dap.WithCommandTimeout(context.Background()) // Default 30s
defer cancel()
```

## 5. Global Session Management

**Pattern**: Single `globalSession` variable shared across tool calls.

```go
var globalSession *dap.Session
func GetSession() (*dap.Session, error) { ... }
```

## 6. Error Message Pattern

**Pattern**: Problem + Context + Solution.

```go
return fmt.Errorf("Invalid project path: %s\nPossible causes: ...\nSolutions: ...", path)
```

## 7. Path Resolution Pattern

**Problem**: Godot DAP requires absolute paths; users/scripts use `res://`.

**Solution**: Automatically convert `res://` paths using the project root stored in session.

```go
// internal/tools/path.go
func ResolveGodotPath(path string, projectRoot string) string {
    if strings.HasPrefix(path, "res://") {
        return filepath.Join(projectRoot, strings.TrimPrefix(path, "res://"))
    }
    return path
}
```

## 8. Security Validation Pattern

**Solution**: Strict whitelist validation for user inputs (variable names).

```go
regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
```

## 9. Node Inspection Pattern

**Solution**: Use existing variables system. Expand `self` -> `Node/children`.

## 10. JSON Schema Validation Pattern

**Solution**: Use empty string for "any type" parameters.

```go
{ Name: "value", Type: "" } // Omitted = any
```
