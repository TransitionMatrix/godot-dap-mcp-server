# Godot DAP Server - FAQ for Client Implementers

Common questions when implementing a DAP client/bridge for Godot Engine.

## Connection & Setup

### Q: What port does Godot's DAP server use?
**A**: Default port **6006** (TCP, localhost only)
- Configurable: `Editor → Editor Settings → Network → Debug Adapter → remote_port`
- Command-line override: Set `DebugAdapterServer::port_override` before editor starts

### Q: Does the DAP server start automatically?
**A**: Yes, when editor starts (on `NOTIFICATION_ENTER_TREE`)
- Binds to `127.0.0.1:6006`
- No UI toggle to disable

### Q: How many concurrent connections are supported?
**A**: Maximum **8 clients** (`DAP_MAX_CLIENTS`)
- All clients see same debugging state
- Single session model (one game instance)

### Q: What's the message format?
**A**: Content-Length header format:
```
Content-Length: <bytes>\r\n
\r\n
<JSON_content>
```
- Maximum message size: 4MB

---

## Initialization

### Q: What's the correct initialization sequence?
**A**: Critical 3-step sequence (discovered through Phase 3 debugging):
```
1. initialize          → Get capabilities (State: Initialized)
2. configurationDone   → Signal ready for debugging (State: Configured)
3. setBreakpoints      → Optional, can set breakpoints now
4. launch/attach       → Store params (doesn't launch yet!)
5. configurationDone   → Actually launches game!
```

**IMPORTANT**: Without step 2 (`configurationDone` after `initialize`), the session remains in "initialized" state and breakpoint commands will timeout. This is a **required step**, not optional.

For full details, see [ARCHITECTURE.md - DAP Protocol Handshake Pattern](../ARCHITECTURE.md#3-dap-protocol-handshake-pattern).

### Q: Why isn't my game launching after the launch request?
**A**: **Common mistake!** The `launch` request only *stores* the launch parameters.
- Must send `configurationDone` to actually launch
- Game starts after `configurationDone` response
- Server sends `process` and `initialized` events

### Q: What capabilities does Godot support?
**A**: From `initialize` response:
```json
{
  "supportsConfigurationDoneRequest": true,
  "supportsEvaluateForHovers": true,
  "supportsSetVariable": true,
  "supportedChecksumAlgorithms": ["MD5", "SHA1", "SHA256"],
  "supportsRestartRequest": true,
  "supportsValueFormattingOptions": true,
  "supportTerminateDebuggee": true,
  "supportSuspendDebuggee": true,
  "supportsTerminateRequest": true,
  "supportsBreakpointLocationsRequest": true
}
```

---

## Launch & Attach

### Q: What launch parameters does Godot support?
**A**: Key parameters:
```json
{
  "project": "/absolute/path/to/project",  // Optional, validated
  "scene": "main" | "current" | "res://path/to/scene.tscn",
  "platform": "host" | "android" | "web",
  "device": -1,  // For Android
  "noDebug": false,  // Skip breakpoints
  "playArgs": ["--arg1", "value1"]  // CLI args
}
```
- `"main"`: Launch main scene from project.godot
- `"current"`: Launch currently open scene in editor
- Resource path: Launch specific scene

### Q: Why am I getting "WRONG_PATH" error?
**A**: Godot validates that paths match editor's project path
- **Solution**: Use absolute paths
- **Solution**: Match editor's working directory exactly
- **Beware**: Symbolic links, case sensitivity
- Windows: Convert `\` → `/`, uppercase drive letter

### Q: How do I launch on Android or Web?
**A**: Set `platform` parameter:
```json
{
  "platform": "android",  // or "web"
  "device": 0  // Android device index
}
```
- Requires export templates configured
- May get `UNKNOWN_PLATFORM` or `MISSING_DEVICE` errors

---

## Breakpoints

### Q: How do I set breakpoints?
**A**: Use `setBreakpoints` request:
```json
{
  "source": { "path": "/absolute/path/to/file.gd" },
  "breakpoints": [
    { "line": 10 },
    { "line": 25 }
  ]
}
```
- **Important**: Replaces ALL breakpoints in file (not additive)
- Empty array clears all breakpoints

### Q: How do I clear breakpoints?
**A**: Send `setBreakpoints` with empty array:
```json
{
  "source": { "path": "/path/to/file.gd" },
  "breakpoints": []
}
```

### Q: Does Godot support conditional breakpoints?
**A**: **No**, conditional breakpoints are not supported
- `condition` field is ignored
- Workaround: Set breakpoint, use `evaluate` to check condition

### Q: Are line numbers 0-based or 1-based?
**A**: Client specifies in `initialize`:
```json
{
  "linesStartAt1": true  // or false
}
```
- Godot adjusts internally: `line + !linesStartAt1`

---

## Execution Control

### Q: Which stepping commands are supported?
**A**:
- ✅ `next` (step over) - Supported
- ✅ `stepIn` (step into) - Supported
- ❌ `stepOut` (step out) - **NOT IMPLEMENTED**

### Q: Why is stepOut not working?
**A**: **Godot does not implement stepOut**
- No `req_stepOut()` method exists in Godot's source code
- Confirmed in: `editor/debugger/debug_adapter/debug_adapter_parser.cpp`
- Workarounds:
  - Set breakpoint after function call, use `continue`
  - Step through remaining lines manually
  - Notify user that stepOut is unavailable

### Q: Why don't I receive a "continued" event after continue?
**A**: It's **asynchronous** - separate from response
```
1. Send continue request
2. Receive continue response (immediate)
3. [Later] Receive continued event
```
- Don't block waiting for event after response
- Handle events in event loop

### Q: When do I get "stopped" events?
**A**: Four reasons:
- `"pause"` - After pause request
- `"breakpoint"` - Hit breakpoint
- `"step"` - After next/stepIn completes
- `"exception"` - Runtime error occurred

---

## Inspection

### Q: How do I inspect variables?
**A**: Four-step process:
```
1. stackTrace → Get stack frames
2. scopes(frameId) → Get scopes (Locals, Members, Globals)
3. variables(scopeReference) → Get variables in scope
4. variables(variableReference) → Expand complex variables
```
- Each level requires separate request

### Q: What scopes does Godot provide?
**A**: Always exactly **3 scopes**:
1. **Locals** - Function-local variables
2. **Members** - Instance/class members
3. **Globals** - Global variables and autoloads

- Returned even if empty

### Q: How do I expand Vector2, Vector3, etc.?
**A**: Use `variables` request with `variablesReference`:
```
Vector2 variable: variablesReference = 2000
→ variables(2000) returns: { x: 10, y: 20 }
```
- Complex types have `variablesReference > 0`
- Primitives have `variablesReference = 0`

### Q: Can I evaluate GDScript expressions?
**A**: Yes, use `evaluate` request:
```json
{
  "expression": "player.health * 2",
  "frameId": 1  // Optional stack frame
}
```
- Evaluates in context of stack frame
- Can access locals, members, globals
- Results are volatile (cached temporarily)

---

## Threading

### Q: How many threads does Godot report?
**A**: Always exactly **1 thread**
```json
{
  "threads": [
    { "id": 1, "name": "Main" }
  ]
}
```
- All debugging on single thread
- Thread ID always 1 in all requests

---

## Events & Async

### Q: How do I handle async events?
**A**: Event filtering pattern:
```
Send request
Wait for response:
    Read message
    If Response with matching seq:
        Return response
    If Event:
        Handle event (update UI, queue)
        Continue waiting
```
- Don't block on events
- Events sent to all connected clients

### Q: What events should I expect?
**A**: Common events:
- `initialized` - Ready for debugging
- `process` - Game launched
- `stopped` - Execution paused
- `continued` - Execution resumed
- `output` - Console output
- `breakpoint` - Breakpoint changed
- `terminated` - Game stopped

---

## Path Handling

### Q: How should I format file paths?
**A**: Best practices:
- **Always** use absolute paths
- **Always** use forward slashes (`/`)
- Windows: Uppercase drive letter (e.g., `C:/project/script.gd`)
- Match editor's project directory exactly

**Note**: `res://` paths are NOT supported in DAP protocol. Godot requires absolute filesystem paths for all file references (breakpoints, source files, etc.). This is a **Godot limitation**, not a DAP protocol limitation.

For MCP servers bridging to Godot DAP, you'll need to convert `res://` paths to absolute paths. See planned enhancement in [PLAN.md - Phase 7](../PLAN.md).

### Q: Why are Windows paths failing?
**A**: Godot auto-converts but validation runs first
- Convert before sending: `c:\path` → `C:/path`
- Backslashes → Forward slashes
- Uppercase drive letter

---

## Error Handling

### Q: What errors should I handle?
**A**: Common DAP errors:
- `WRONG_PATH` - Path doesn't match project
- `NOT_RUNNING` - No active session (attach)
- `TIMEOUT` - Request timed out
- `UNKNOWN_PLATFORM` - Invalid platform
- `MISSING_DEVICE` - Android device not found

### Q: What's the maximum message size?
**A**: 4MB (`DAP_MAX_BUFFER_SIZE`)
- Error: "Response content too big"
- Paginate large requests (stackTrace, variables)

### Q: How do I handle timeouts?
**A**: Implement client-side timeouts:
- Server has `request_timeout` setting but enforcement varies
- Recommended: 10-30 second timeouts on all requests

---

## Implementation Tips

### Q: Should I implement stepOut anyway?
**A**: No, it will fail
- Godot will return "command not supported" error
- Document limitation in your client
- Provide workarounds (continue to breakpoint)

### Q: How do I test my implementation?
**A**:
1. Start Godot editor with project open
2. Connect to port 6006
3. Send initialize → launch → configurationDone
4. Set breakpoints, let game hit them
5. Test stepping, variable inspection

### Q: What's the most common mistake?
**A**: Forgetting `configurationDone`
- Launch won't execute without it
- Appears like "hanging"
- Always send after launch request

### Q: Where can I find implementation examples?
**A**: Check existing DAP clients:
- VS Code Godot plugins
- JetBrains Rider Godot support
- These implement Godot DAP protocol

---

## Quick Reference

### Minimum Viable Workflow
```
1. connect to localhost:6006
2. initialize → Get capabilities
3. launch → Store params
4. setBreakpoints → Set BPs (optional)
5. configurationDone → Actually launch!
6. [Wait for stopped event]
7. threads → Get thread list
8. stackTrace → Get frames
9. scopes → Get scopes
10. variables → Get variable values
```

### Commands That Return Immediately (Async)
- `continue` - Response immediate, `continued` event later
- `next` - Response immediate, `stopped` event later
- `stepIn` - Response immediate, `stopped` event later

### Commands That Block/Wait
- `stackTrace` - Waits for debugger
- `variables` - May wait if stack dump in progress
- `evaluate` - Waits for evaluation

---

---

## MCP Bridge Considerations

### Q: How do I maintain session state across MCP tool calls?
**A**: MCP servers must use a persistent process, not per-request spawning:

**Problem**: Each MCP tool call (e.g., `godot_connect`, `godot_set_breakpoint`) needs to share the same DAP session.

**Solution**: Use named pipes with file descriptors to maintain persistent stdin/stdout:
```bash
mkfifo /tmp/mcp-stdin /tmp/mcp-stdout
exec 3<>/tmp/mcp-stdin
exec 4<>/tmp/mcp-stdout
./godot-dap-mcp-server </tmp/mcp-stdin >/tmp/mcp-stdout &
```

Without persistent pipes, each tool call spawns a new server process and loses the DAP session. See [IMPLEMENTATION_GUIDE.md - Integration Testing Patterns](../IMPLEMENTATION_GUIDE.md#integration-testing-patterns) for details.

### Q: How do I parse MCP responses with nested JSON?
**A**: MCP wraps tool results, causing escaped quotes:

```json
{
  "result": {
    "content": [{
      "text": "{\\\"status\\\":\\\"connected\\\"}"
    }]
  }
}
```

Use regex that handles both escaped and unescaped:
```bash
grep -qE '(\\"|")status(\\"|"):(\\"|")connected'
```

See [IMPLEMENTATION_GUIDE.md - Robust JSON Parsing](../IMPLEMENTATION_GUIDE.md#pattern-robust-json-parsing-in-bash) for complete pattern.

---

## Additional Resources

For more detailed information, consult the Godot Engine source code:
- `editor/debugger/debug_adapter/debug_adapter_server.{h,cpp}` - Server setup
- `editor/debugger/debug_adapter/debug_adapter_protocol.{h,cpp}` - Protocol handling
- `editor/debugger/debug_adapter/debug_adapter_parser.{h,cpp}` - Command implementations
- `editor/debugger/debug_adapter/debug_adapter_types.h` - Type definitions

Or reference the DAP specification: https://microsoft.github.io/debug-adapter-protocol/

**For debugging insights**: See [LESSONS_LEARNED_PHASE_3.md](../LESSONS_LEARNED_PHASE_3.md) for the complete Phase 3 debugging journey.