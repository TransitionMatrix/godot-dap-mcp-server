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

**Key Points**:
- `godot_connect` ONLY initializes.
- `godot_launch_*` handles the `Launch` -> `ConfigurationDone` sequence.
- `LaunchWithConfigurationDone` method handles the specific ordering Godot requires.

## 2. Event and Response Handling Pattern (DAP Client)

**Problem**: Godot's DAP server sends async events (stopped, continued, output, etc.) mixed with command responses. Additionally, the `go-dap` library decodes messages into specific response types (e.g., `*dap.PauseResponse`, `*dap.LaunchResponse`), which must be explicitly handled in `waitForResponse` to avoid unexpected type assertions and timeouts.

**Solution**: Implement robust `waitForResponse` that filters events and explicitly handles all expected DAP response types.

```go
func (c *Client) waitForResponse(ctx context.Context, expectedCommand string) (dap.Message, error) {
    for {
        msg, err := c.ReadWithTimeout(ctx)
        // ... error handling ...

        switch m := msg.(type) {
        case *dap.Response:
            // Generic DAP response. Check its command field.
            if m.Command == expectedCommand {
                return m, nil
            }
        case *dap.ErrorResponse:
            // Handle error responses directly.
            if m.Command == expectedCommand {
                return nil, fmt.Errorf("command %s error: %s", command, m.Message)
            }
        case *dap.Event:
            // Log known events and continue waiting.
            c.logEvent(m)
            
        // Explicitly handle specific DAP response types to match expectedCommand
        case *dap.InitializeResponse:
            if expectedCommand == "initialize" { return m, nil }
        case *dap.ConfigurationDoneResponse:
            if expectedCommand == "configurationDone" { return m, nil }
        case *dap.LaunchResponse:
            if expectedCommand == "launch" { return m, nil }
        case *dap.SetBreakpointsResponse:
            if expectedCommand == "setBreakpoints" { return m, nil }
        case *dap.ContinueResponse:
            if expectedCommand == "continue" { return m, nil }
        case *dap.NextResponse:
            if expectedCommand == "next" { return m, nil }
        case *dap.StepInResponse:
            if expectedCommand == "stepIn" { return m, nil }
        case *dap.PauseResponse: // ✅ New: Correctly handles PauseResponse
            if expectedCommand == "pause" { return m, nil }
        // ... other specific response types ...

        default:
            // Log unknown messages or other event types that might not implement *dap.Event
            c.logEvent(m) // Fallback for various event types
        }
    }
}
```

**Key Points**:
- Never return on first message - it might be an event or a response for another command.
- Loop until the exact `expectedCommand` response is received.
- `ErrorResponse` returns an error immediately.
- Log events and unexpected responses, but continue waiting for the target response.
- **Critical**: Explicitly `case` match all specific `*dap.XxxResponse` types returned by the `go-dap` library to ensure correct type assertion and command matching. This was the fix for `godot_pause` timeout.

## 3. Timeout Protection Pattern

**Problem**: DAP server may hang or not respond to certain requests, causing permanent blocking.

**Solution**: Wrap all DAP operations with context timeouts.

```go
func (t *Tool) Execute(args map[string]interface{}) (interface{}, error) {
    // Create timeout context (10-30 seconds depending on operation)
    ctx, cancel := dap.WithCommandTimeout(context.Background()) // Default 30s
    defer cancel()

    if err := client.SomeOperation(ctx); err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return nil, fmt.Errorf("operation timed out")
        }
        return nil, err
    }
    return result, nil
}
```

**Timeout Guidelines**:
- Connect: 10s
- Command: 30s (Launch, Step, etc.)
- Read: 5s

## 4. Global Session Management

**Pattern**: Single `globalSession` variable shared across tool calls.

```go
var globalSession *dap.Session

func GetSession() (*dap.Session, error) {
    if globalSession == nil {
        return nil, fmt.Errorf("not connected - call godot_connect first")
    }
    return globalSession, nil
}
```

**Key Points**:
- One debugging session at a time
- Tools use `GetSession()` to access shared session
- Clear error if not connected

## 5. Error Message Pattern

**Pattern**: Problem + Context + Solution

```go
return fmt.Errorf(`Invalid project path: project.godot not found at %s

Possible causes:
1. Path does not point to a Godot project directory
2. The path is relative instead of absolute

Solutions:
1. Ensure the path points to the directory containing project.godot
2. Use an absolute path: /full/path/to/project`, path)
```

## 6. Security Validation Pattern (Phase 6)

**Problem**: Tools that execute code (evaluate, set variable) risk code injection attacks.

**Solution**: Strict whitelist validation for user inputs.

```go
// Validate variable names to prevent code injection
func isValidVariableName(name string) bool {
    // Only allow: letters, numbers, underscores
    // Must start with letter or underscore
    matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
    return matched
}
```

**Key Points**:
- Never trust user input for code execution
- Whitelist approach (only allow safe characters)

## 7. stackTrace Verification Pattern

**Problem**: Verifying location after stepping.

**Solution**: Use stackTrace immediately after stepping.

```go
// 1. stepIn
client.StepIn(ctx, threadId)

// 2. stackTrace
stack, err := client.StackTrace(ctx, threadId, 0, 1)
// Verify stack.StackFrames[0].Name / Line
```

## 8. Node Inspection Pattern

**Problem**: Inspect scene tree and node properties.

**Solution**: Use existing variables system with object expansion.

1. Get `Members` scope (contains `self`)
2. Expand `self` to get current Node
3. Look for `Node/children` property
4. Expand children array to navigate tree

**Key Points**:
- NO separate "scene tree" DAP command exists
- Uses standard variable inspection tools

## 9. JSON Schema Validation Pattern

**Problem**: "any" type is invalid in JSON Schema 2020-12.

**Solution**: Use empty string for "any type" parameters with omitempty tag.

```go
// ✅ CORRECT: Empty type with omitempty omits field from JSON
{
    Name: "value",
    Type: "",  // Omitted from schema = accepts any type
}
```
