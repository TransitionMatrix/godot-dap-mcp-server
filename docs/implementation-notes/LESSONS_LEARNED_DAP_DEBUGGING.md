# Lessons Learned: DAP Debugging Workflow

**Date:** November 9, 2025
**Context:** Fixing Godot DAP implementation + building DAP client

---

## Discovery Process

### Initial Problem
After fixing Godot's Dictionary safety issues and launch response, we needed to verify that DAP worked end-to-end. The natural next step was testing breakpoints.

### What We Learned

#### 1. Correct DAP Sequence (Critical!)

**Wrong (causes failures):**
```
Initialize → Launch → ConfigurationDone → Wait → SetBreakpoints
                                           ↑
                                    Game already started!
```

**Correct:**
```
Initialize → Launch → SetBreakpoints → ConfigurationDone
                      ↑                 ↑
              BEFORE game starts    Triggers launch
```

**Why this matters:**
- `launch` request **stores** parameters but doesn't start the game
- `configurationDone` **triggers** the actual launch
- Breakpoints must be set BETWEEN these commands
- This is by design - allows debuggers to configure before execution begins

**Reference:** See `godot-source` memory: `dap_faq_for_clients.md` - "Launch Flow"

---

#### 2. Path Validation Requirements

**Error encountered:**
```
Failed to set breakpoint: command setBreakpoints returned error: wrong_path
```

**Problem:**
```go
// ❌ WRONG - res:// paths don't work for breakpoints
client.SetBreakpoints(ctx, "res://src/scenes/main_scene/main_scene.gd", []int{56})
```

**Solution:**
```go
// ✅ CORRECT - absolute filesystem paths required
scriptPath := projectPath + "/src/scenes/main_scene/main_scene.gd"
// Example: /Users/user/Projects/game/src/scenes/main_scene/main_scene.gd
client.SetBreakpoints(ctx, scriptPath, []int{56})
```

**Why:**
- Godot validates paths against editor's project directory
- `res://` is Godot's internal path scheme, not DAP protocol
- DAP expects absolute filesystem paths
- Path must match exactly (case-sensitive, no symlinks)

**Reference:** Godot DAP FAQ - "Why am I getting 'WRONG_PATH' error?"

---

#### 3. Event Filtering Gaps

**Problem discovered:**
When testing setBreakpoints after game exit, we saw:
```
Received unknown message type: *dap.ErrorResponse (waiting for setBreakpoints)
[Hangs forever - timeout after 30 seconds]
```

**Root cause:**
```go
// events.go - BEFORE fix
switch m := msg.(type) {
case *dap.Response:
    // Handle responses
case *dap.Event:
    // Handle events
default:
    // Unknown type - just log and continue
    // ❌ ErrorResponse falls through here!
    log.Printf("Received unknown message type: %T", msg)
    continue  // Infinite loop!
}
```

**Solution:**
```go
// events.go - AFTER fix
case *dap.ErrorResponse:
    // Error response - command failed
    return nil, fmt.Errorf("command %s returned error: %s", command, m.Message)

case *dap.Event:
    // Generic event handler
    c.logEvent(m)
    continue

// Specific event types (don't match generic *dap.Event in Go's type system)
case *dap.InitializedEvent, *dap.StoppedEvent, *dap.ContinuedEvent,
     *dap.ExitedEvent, *dap.TerminatedEvent, *dap.ThreadEvent,
     *dap.OutputEvent, *dap.BreakpointEvent, *dap.ModuleEvent,
     *dap.LoadedSourceEvent, *dap.ProcessEvent, *dap.CapabilitiesEvent:
    // These are all events - log and continue waiting
    c.logEvent(m)
    continue
```

**Why this matters:**
- `*dap.ErrorResponse` is a valid response type that must be handled
- Go's type system requires explicit handling of concrete event types
- Without this, errors cause 30-second timeouts instead of immediate failure
- Affects all DAP commands - not just setBreakpoints

---

#### 4. Integration Testing Value

**Approach:**
1. Fix code based on analysis
2. Write integration test (`cmd/launch-test/main.go`)
3. Run against real Godot instance
4. Discover actual protocol behavior
5. Fix newly discovered issues
6. Repeat until end-to-end works

**Benefits:**
- Discovered sequence requirements (couldn't learn from docs alone)
- Found path validation issues
- Identified event filtering gaps
- Proved fixes work together
- Built confidence for PR submission

**Key insight:** Static analysis and unit tests aren't enough for protocol implementations. Integration tests reveal real-world behavior.

---

## Testing Workflow

### Setup
```bash
# Terminal 1: Start Godot editor (enables DAP server on port 6006)
godot-dev -e ~/Projects/test-game

# Terminal 2: Run integration test
go run cmd/launch-test/main.go
```

### Test Program Structure
```go
// cmd/launch-test/main.go
func main() {
    // 1. Connect to DAP server
    session := dap.NewSession("localhost", 6006)
    session.Connect(ctx)

    // 2. Initialize
    session.Initialize(ctx)

    // 3. Launch (stores params, doesn't start game)
    session.LaunchMainScene(ctx, projectPath)

    // 4. Set breakpoints BEFORE game starts!
    client.SetBreakpoints(ctx, absolutePath, []int{56})

    // 5. ConfigurationDone (NOW game starts)
    session.ConfigurationDone(ctx)

    // 6. Wait for execution
    time.Sleep(5 * time.Second)

    // 7. Disconnect
    session.Close()
}
```

### Expected Output (Success)
```
=== Testing DAP Launch → SetBreakpoints Sequence ===

1. Connecting to Godot DAP server...
✓ Connected (state: connected)

2. Initializing DAP session...
✓ Initialized (state: initialized)

3. Launching main scene...
✓ Launch request sent (state: initialized)

4. Setting breakpoints (before game starts)...
✓ Breakpoint set successfully!
  Breakpoints: [{Verified:true Line:56}]

5. Sending configurationDone to trigger launch...
✓ Configuration done - game launching with breakpoints active

6. Waiting for breakpoint or game execution...

=== SUCCESS! Launch → SetBreakpoints sequence works! ===
```

---

## Development Patterns

### Pattern 1: Protocol Discovery via Integration Testing

**When to use:** Implementing unfamiliar protocols (DAP, LSP, MCP, etc.)

**Steps:**
1. Read specification
2. Implement based on spec understanding
3. Write integration test
4. Discover actual behavior differs from expectation
5. Research (Godot source, community implementations)
6. Update implementation
7. Verify with integration test

**Example from this workflow:**
- Spec says "launch starts the game"
- Actually: launch stores params, configurationDone starts game
- Discovered by testing: breakpoints failed when set after launch
- Fixed by consulting Godot source and FAQ

### Pattern 2: Error Message Analysis

**When error occurs:**
1. Don't just look at error text
2. Look at **what happened right before**
3. Check for patterns (event types, timing)

**Example:**
```
Error: timeout after 30 seconds
Look before error:
  - Received *dap.ErrorResponse
  - Received *dap.ExitedEvent
  - Received *dap.TerminatedEvent
```

**Insight:** Game exited! We were setting breakpoints too late.

### Pattern 3: Type System Clues

Go's type system revealed event filtering bug:

```go
case *dap.Event:
    // ❌ Concrete event types don't match this!
```

The compiler didn't error, but runtime showed "unknown message type" for concrete events. This revealed that Go's type switch requires explicit handling of concrete types even when they implement the same interface.

---

## Godot-Specific Discoveries

### 1. Launch Flow (Deferred Execution)
- `launch` stores parameters in `pending_launch`
- `configurationDone` executes `_launch_process()` using stored params
- This allows breakpoints and configuration between request and execution

**Why designed this way:**
- Standard DAP pattern
- Allows setting breakpoints before code runs
- Provides window for multiple configuration steps

### 2. Path Validation
- Godot validates `source.path` against editor's project directory
- Uses `is_valid_path()` helper
- Must be absolute path, exact match
- Windows: Forward slashes, uppercase drive letter

### 3. Event Types
- Godot sends many events during normal operation
- `InitializedEvent` - After initialize
- `ProcessEvent` - When game launches
- `ExitedEvent` - When game stops
- `TerminatedEvent` - After exit
- Clients must filter events while waiting for responses

---

## Documentation Improvements Needed

Based on this workflow, we should document:

1. **godot-dap-mcp-server README:**
   - Correct DAP sequence with emphasis on setBreakpoints timing
   - Path requirements (absolute, not res://)
   - Integration testing approach

2. **Godot upstream documentation:**
   - Make configurationDone timing more prominent
   - Add example showing breakpoint setting before configurationDone
   - Document path validation requirements clearly

3. **DAP Tool Design Guide:**
   - Event filtering patterns
   - ErrorResponse handling
   - Integration testing strategies

---

## Takeaways

1. **Protocol specs are not enough** - Real implementations have quirks
2. **Integration tests reveal truth** - Unit tests can't catch sequence issues
3. **Error responses must be handled** - Silent failures cause confusion
4. **Path validation matters** - Small details break functionality
5. **Type systems help** - Go's concrete type requirements revealed bug

**Most important:** Don't just fix bugs - test end-to-end to prove the system works.

---

## Next Steps

Now that we have working end-to-end DAP:

1. ✅ Godot Dictionary fixes verified
2. ✅ Launch response fix verified
3. ✅ Client event filtering verified
4. ✅ Breakpoint workflow verified
5. 🔲 Test other DAP commands (continue, step, inspect)
6. 🔲 Create comprehensive PR for Godot
7. 🔲 Document findings in godot-dap-mcp-server

**PR Status:** Ready to demonstrate full functionality!
