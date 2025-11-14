# Critical Implementation Patterns

These are essential patterns discovered through testing and implementation. Following these patterns is critical for correctness.

## 1. DAP Protocol Handshake Pattern

**Problem**: Missing `ConfigurationDone()` leaves session in "initialized" state where breakpoint commands timeout.

**Solution**: Complete 3-step handshake (Phase 3 discovery)
```go
// Step 1: Initialize - Get server capabilities
if err := session.Initialize(ctx); err != nil {
    return fmt.Errorf("failed to initialize: %w", err)
}
// State: Initialized (NOT ready for debugging yet!)

// Step 2: ConfigurationDone - Signal "ready for debugging"
if err := session.ConfigurationDone(ctx); err != nil {
    return fmt.Errorf("failed to complete configuration: %w", err)
}
// State: Configured (NOW ready for breakpoints, launch, etc.)

// Step 3: Launch (optional) - Start game instance
if err := session.Launch(ctx, launchArgs); err != nil {
    return fmt.Errorf("failed to launch: %w", err)
}
// State: Launched (game running)
```

**Key Points**:
- `ConfigurationDone()` is **required**, not optional
- Without it, session stays in "initialized" state
- Breakpoints set before `ConfigurationDone()` will timeout
- State transitions: Disconnected → Connected → Initialized → Configured → Launched

## 2. Event Filtering Pattern (DAP Client)

**Problem**: Godot's DAP server sends async events (stopped, continued, output, etc.) mixed with command responses. Reading without filtering will return events instead of expected responses.

**Solution**: Always filter for expected response type (Phase 6 refinement adds ErrorResponse handling)
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
        
        case *dap.ErrorResponse:
            // Command failed - return error immediately
            return nil, fmt.Errorf("command %s returned error: %s", command, m.Message)
        
        case *dap.Event:
            // Log event but don't return - continue waiting
            c.logEvent(m)
        
        // Explicit event types (needed since some don't match *dap.Event)
        case *dap.InitializedEvent, *dap.StoppedEvent, *dap.ContinuedEvent,
            *dap.ExitedEvent, *dap.TerminatedEvent, *dap.ThreadEvent,
            *dap.OutputEvent, *dap.BreakpointEvent, *dap.ModuleEvent:
            c.logEvent(m)
            continue
        
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
- ErrorResponse returns error immediately (no retry)
- Log events but don't treat them as responses
- Explicit event type cases needed (some don't match *dap.Event interface)
- Use timeouts to prevent infinite loops

**Phase 6 Addition - Stopped Event Variants**:
The `stopped` event has multiple `reason` values. Our event filtering is **already reason-agnostic** (good design!) but be aware of all variants:
- `"breakpoint"` - Hit breakpoint
- `"step"` - Step operation completed
- `"pause"` - User-requested pause (Phase 6)
- `"exception"` - Runtime error occurred

Our Phase 3 implementation correctly accepts any stopped event regardless of reason.

## 3. Timeout Protection Pattern

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
- Quick operations: 10 seconds (connect, breakpoint, pause)
- Launch operations: 30 seconds (scene loading)
- Step operations: 15 seconds (may hit breakpoint)
- Read operations: 5 seconds (should be fast)

## 4. Persistent MCP Session Pattern (Integration Testing)

**Problem**: MCP tools need to share DAP session across multiple tool calls. Spawning new server per request loses session.

**Solution**: File descriptors with named pipes (Phase 3 discovery)
```bash
# Create named pipes
mkfifo /tmp/mcp-stdin /tmp/mcp-stdout

# Open FDs to keep pipes alive
exec 3<>/tmp/mcp-stdin
exec 4<>/tmp/mcp-stdout

# Start MCP server (pipes stay open because FDs are open)
./godot-dap-mcp-server </tmp/mcp-stdin >/tmp/mcp-stdout &

# Send requests via FD 3, read responses via FD 4
echo "$request" >&3
read -r response <&4
```

**Key Points**:
- File descriptors keep pipes open between writes
- Server stdin never sees EOF
- Session persists across all tool calls
- Clean shutdown via `exec 3>&- 4>&-` in trap

## 5. Error Message Pattern

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

## 6. Global Session Management

**Pattern**: Single `globalSession` variable shared across tool calls

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

## 7. JSON Schema Validation Pattern (Phase 6 Discovery)

**Problem**: Claude API validates tool schemas against JSON Schema draft 2020-12. Using invalid types causes runtime errors.

**Solution**: Use empty string for "any type" parameters with omitempty tag

```go
// ❌ WRONG: "any" is not a valid JSON Schema type
{
    Name: "value",
    Type: "any",  // Causes: "JSON schema is invalid"
}

// ✅ CORRECT: Empty type with omitempty omits field from JSON
{
    Name: "value",
    Type: "",  // Omitted from schema = accepts any type
}

// PropertyDefinition with omitempty tag
type PropertyDefinition struct {
    Type        string      `json:"type,omitempty"`  // Key: omitempty!
    Description string      `json:"description"`
    Default     interface{} `json:"default,omitempty"`
}
```

**Valid JSON Schema Types** (draft 2020-12):
- `string`, `number`, `integer`, `boolean`, `object`, `array`, `null`
- **NOT** `any` (will cause validation error)

**Key Points**:
- Use `Type: ""` to accept any value type
- Add `omitempty` to Type field in PropertyDefinition
- Test generated schemas with `tools/list` before deployment
- Claude API errors point to tool index, not the actual problematic tool

## 8. Security Validation Pattern (Phase 6 Discovery)

**Problem**: Tools that execute code (evaluate, set variable) risk code injection attacks.

**Solution**: Strict whitelist validation for user inputs

```go
// Validate variable names to prevent code injection
func isValidVariableName(name string) bool {
    // Only allow: letters, numbers, underscores
    // Must start with letter or underscore
    matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
    return matched
}

// ✅ Valid: player_health, _internal, score123
// ❌ Reject: "health + 10", "get_node('/root').queue_free()"

// Format values for GDScript
func formatValueForGDScript(value interface{}) string {
    switch v := value.(type) {
    case string:
        return fmt.Sprintf(`"%s"`, escapeString(v))  // Quote strings
    case int, int64, float64:
        return fmt.Sprintf("%v", v)  // Numbers as-is
    case bool:
        return fmt.Sprintf("%v", v)  // true/false
    default:
        return fmt.Sprintf("%v", v)
    }
}
```

**Key Points**:
- Never trust user input for code execution
- Whitelist approach (only allow safe characters)
- Provide clear error messages explaining validation rules
- Test with injection attack scenarios

## 9. Known Issues and Workarounds

### stepOut Not Implemented (Phase 3)
```go
// ❌ This will hang - stepOut not implemented in Godot
err := client.StepOut(ctx)

// ✅ Workaround: Use continue or step-over instead
err := client.Continue(ctx)
```

### setVariable Not Implemented (Phase 6)
```go
// ❌ Godot claims support but doesn't implement it!
// Capabilities advertise: supportsSetVariable: true
// Reality: No req_setVariable() method exists

// ✅ Workaround: Use evaluate() with assignment expression
expression := fmt.Sprintf("%s = %v", varName, newValue)
result, err := client.Evaluate(ctx, expression, frameId)

// ⚠️ SECURITY: Validate variable names to prevent code injection
// Pattern: ^[a-zA-Z_][a-zA-Z0-9_]*$
```

**Critical**: Both `stepOut` and `setVariable` are **false advertising** - Godot's capabilities claim support but don't implement the commands.

### Absolute Paths Required
```go
// ❌ res:// paths not supported by Godot DAP
file := "res://scripts/player.gd"

// ✅ Use absolute paths
file := "/absolute/path/to/project/scripts/player.gd"
```

### pause Command Response Handling (Phase 6 - Under Investigation)
**Observation**: After `pause` command, subsequent DAP requests may timeout or receive malformed responses.

**Symptoms**:
- Commands execute (game steps, pauses) but responses are corrupted
- Protocol parsing errors: "header format is not '^Content-Length: ([0-9]+)$'"
- May indicate event messages not being consumed properly

**Status**: Being investigated. May require event filtering improvements or special handling of pause-related stopped events.

## 10. Node Inspection Pattern (Phase 6 Discovery)

**Problem**: How to inspect scene tree and node properties via DAP?

**Solution**: Use existing variables system with object expansion

```go
// Nodes are inspected through OBJECT case in parse_variant()
// When an object is expanded, parse_object() categorizes properties:
// - "Members/" prefix → Script members
// - "Constants/" prefix → Script constants  
// - "Node/" prefix → Node-specific properties (name, parent, children)
// - Regular properties grouped by category (Transform2D, CanvasItem, etc.)

// Navigation pattern:
// 1. Get Members scope (contains 'self')
// 2. Expand 'self' to get current Node
// 3. Look for 'Node/children' property
// 4. Expand children array to navigate tree
```

**Key Points**:
- NO separate "scene tree" DAP command exists
- Scopes are FIXED at 3 types: Locals, Members, Globals (no "SceneTree" scope)
- Scene tree navigation uses existing `godot_get_variables` tool
- Document the pattern, don't create redundant tools

## 11. Godot Dictionary Safety Pattern (Upstream Research)

**Finding**: Godot's DAP implementation inconsistently uses safe vs unsafe Dictionary access.

**Evidence** (from compliance testing):
```cpp
// ✅ SAFE: Client capabilities (req_initialize:133-136)
peer->linesStartAt1 = args.get("linesStartAt1", false);
peer->columnsStartAt1 = args.get("columnsStartAt1", false);

// ❌ UNSAFE: Source parsing (debug_adapter_types.h:87-89)
name = p_params["name"];           // Crashes if "name" missing
path = p_params["path"];           
_checksums = p_params["checksums"]; // Crashes if "checksums" missing
```

**Impact**:
- DAP spec marks `name` and `checksums` as **optional**
- Godot uses these fields as **output-only** (ignores client values, regenerates from path)
- Test tool triggers 2 Dictionary errors with minimal DAP-compliant messages
- Verified in godot-upstream HEAD (0b5ad6c73c) - issues persist

**Testing**:
- Tool: `cmd/test-dap-protocol/` sends minimal DAP messages
- Automation: `scripts/test-dap-compliance.sh` for verification
- Detects unsafe Dictionary access in req_setBreakpoints path

**Client Capabilities for LLM/MCP**:
- Use 1-based indexing (human-friendly line numbers)
- Request type information (aids debugging context)
- Identify as "godot-dap-mcp-server" in logs

For complete debugging stories, see `docs/LESSONS_LEARNED_PHASE_N.md`.
