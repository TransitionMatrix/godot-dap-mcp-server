# Lessons Learned - Phase 3 Integration Testing

**Date**: 2025-11-07
**Context**: Implementing automated integration tests for Phase 3 core debugging tools

This document captures the debugging journey and key insights gained while implementing the integration test infrastructure for godot-dap-mcp-server.

## 📚 Documentation Organization (Hybrid Approach)

This document preserves the **debugging narrative** - the story of how we discovered critical patterns through iterative debugging. The **reusable patterns** have been extracted to appropriate reference documentation for discoverability:

**Reusable Patterns → [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md#integration-testing-patterns)**:
- Pattern 1: Persistent Subprocess Communication (file descriptors + named pipes)
- Pattern 2: Port Conflict Handling (auto-detect + fallback)
- Pattern 4: Robust JSON Parsing in Bash (handling escaped quotes)

**Critical Implementation Details → [ARCHITECTURE.md](ARCHITECTURE.md#critical-implementation-patterns)**:
- Pattern 3: DAP Protocol Handshake (ConfigurationDone requirement and state transitions)

**Known Quirks & FAQ → [GODOT_DAP_FAQ.md](reference/GODOT_DAP_FAQ.md)**:
- Absolute path requirements (res:// not supported)
- Session persistence for MCP bridges
- JSON response escaping in nested structures

**This document provides the CONTEXT and STORY behind those patterns.** Read this to understand WHY we made these decisions and HOW we discovered them through debugging.

---

## Table of Contents

1. [The Challenge](#the-challenge)
2. [Critical Issues Encountered](#critical-issues-encountered)
3. [Solutions and Patterns](#solutions-and-patterns)
4. [Key Takeaways](#key-takeaways)
5. [Future Considerations](#future-considerations)

---

## The Challenge

**Goal**: Create a fully automated integration test that:
- Launches Godot editor as a subprocess
- Starts MCP server with persistent session
- Tests all 7 Phase 3 debugging tools
- Runs without manual setup

**Why This Was Hard**:
- MCP server needs persistent stdin/stdout for session management
- Bash subprocess communication is tricky with bidirectional pipes
- Godot's DAP protocol has specific handshake requirements
- JSON response parsing in bash shell scripts

---

## Critical Issues Encountered

### Issue 1: Coprocess Not Available

**Problem**:
```bash
coproc ./godot-dap-mcp-server
# Error: coproc: command not found
```

**Root Cause**: The `coproc` builtin may not be available in all shell environments (sh vs bash differences).

**Symptom**: Test failed immediately at "Starting MCP server..." step.

**Why This Mattered**:
- Need persistent MCP server process for the entire test
- Each tool call must use the same server instance
- The `globalSession` variable in `internal/tools/connect.go` only persists within a single MCP server process

**What We Tried**:
1. ✅ Started Godot as subprocess successfully
2. ✅ DAP server auto-starts and port detection works
3. ✅ MCP server works when invoked per-request (but loses session!)
4. ❌ Coprocess for persistent session fails

---

### Issue 2: Named Pipes Close After Each Write

**Problem**:
```bash
mkfifo /tmp/mcp-in /tmp/mcp-out
./godot-dap-mcp-server < /tmp/mcp-in > /tmp/mcp-out &

send_request() {
    echo "$1" > /tmp/mcp-in  # This closes the pipe!
}
```

**Root Cause**: Writing to a named pipe without keeping it open causes EOF when the writer closes, which kills the MCP server.

**Symptom**: MCP server PID disappeared from process list after first request.

**Debugging Steps**:
1. Checked if MCP server was running: `ps aux | grep godot-dap-mcp-server` - NOT FOUND
2. Realized each `echo > /tmp/mcp-in` closes the pipe
3. Pipe close → stdin gets EOF → MCP server shuts down gracefully

**The "Aha!" Moment**: Named pipes need to stay open for the entire session, not just per-request.

---

### Issue 3: File Descriptor Solution

**Solution**:
```bash
# Create named pipes
mkfifo /tmp/mcp-stdin /tmp/mcp-stdout

# Open file descriptors to KEEP PIPES OPEN
exec 3<>/tmp/mcp-stdin   # FD 3 for writing requests
exec 4<>/tmp/mcp-stdout  # FD 4 for reading responses

# Start MCP server (pipes stay open because FDs are open)
./godot-dap-mcp-server </tmp/mcp-stdin >/tmp/mcp-stdout 2>/dev/null &
MCP_PID=$!

# Send requests via file descriptor
send_mcp_request() {
    echo "$request" >&3  # Write to FD 3
    read -r response <&4  # Read from FD 4
    echo "$response"
}
```

**Why This Works**:
- File descriptors keep pipes open even after individual writes
- MCP server's stdin never sees EOF
- Session persists across all tool calls
- Clean shutdown via `exec 3>&- 4>&-` in trap

**Key Learning**: Bash file descriptors are the proper way to maintain persistent bidirectional communication with a subprocess.

---

### Issue 4: DAP Protocol Handshake Incomplete

**Problem**:
```bash
# Connection succeeds, state is "initialized"
✓ Connection successful
  State: initialized

# But then breakpoint setting times out
✗ Failed to set breakpoint
Response: {...timeout: context deadline exceeded...}
```

**Root Cause**: We called `Initialize()` but forgot to call `ConfigurationDone()`. According to Godot's DAP implementation:
1. Client sends `initialize` → Server returns capabilities
2. Client sends `configurationDone` → **Server transitions to ready state**
3. Only THEN can you set breakpoints, launch, etc.

**The Fix** (in `internal/tools/connect.go`):
```go
// Initialize the session
if err := session.Initialize(ctx); err != nil {
    session.Close()
    return nil, fmt.Errorf("failed to initialize DAP session: %w", err)
}

// Complete the handshake with configurationDone
if err := session.ConfigurationDone(ctx); err != nil {
    session.Close()
    return nil, fmt.Errorf("failed to complete DAP configuration: %w", err)
}
```

**State Transition**:
- Before fix: `initialized` (not ready for debugging)
- After fix: `configured` (ready for breakpoints and execution control)

**Key Learning**: Read the DAP protocol docs carefully! The handshake isn't just `initialize` - you need `configurationDone` to actually be ready.

---

### Issue 5: Path Format Mismatch

**Problem**:
```bash
# Test uses res:// path
BREAKPOINT_REQUEST='..."file":"res://test_script.gd"...'

# Godot responds with timeout (never responds)
```

**Root Cause**: From Godot DAP documentation:
> **setBreakpoints Path Handling:**
> - Windows paths: `\` → `/`, drive letter uppercased
> - **Uses absolute paths only**

**The Fix**:
```bash
# Before (WRONG)
file="res://test_script.gd"

# After (CORRECT)
TEST_SCRIPT_PATH="$TEST_PROJECT/test_script.gd"  # Absolute path
file="$TEST_SCRIPT_PATH"
```

**Why Godot Requires Absolute Paths**:
- Path validation ensures you're debugging files from the correct project
- Symbolic links and case sensitivity handled consistently
- Windows path normalization is easier with absolute paths

**Key Learning**: This is a **Godot DAP limitation**, not a DAP protocol limitation. We'll add `res://` → absolute path conversion in Phase 7.

---

### Issue 6: JSON Response Escaping

**Problem**:
```bash
# Response has escaped quotes
RESPONSE='{"result":{"content":[{"type":"text","text":"{\"status\":\"connected\"}"}]}}'

# Simple grep fails
if echo "$RESPONSE" | grep -q '"status":"connected"'; then
    # Never matches! ❌
fi
```

**Root Cause**: MCP protocol wraps tool results in nested JSON:
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"status\":\"connected\"}"  // Escaped JSON string
    }]
  }
}
```

**The Fix**:
```bash
# Regex pattern that handles both escaped and unescaped quotes
if echo "$RESPONSE" | grep -qE '(\\"|")status(\\"|"):(\\"|")connected'; then
    # Now it matches! ✅
fi
```

**Pattern Breakdown**:
- `(\\"|")` - Matches either `\"` (escaped) or `"` (unescaped)
- Applied to both key and value quotes
- Works with MCP's nested JSON structure

**Key Learning**: When parsing JSON in bash scripts, always account for escaped quotes in nested JSON strings.

---

### Issue 7: Manual Test Script Had Same Issues

**Problem**: After fixing automated test, the manual test (`integration-test.sh`) failed with the exact same issues.

**Root Cause**: Manual test was using the old approach:
```bash
send_mcp_request() {
    echo "$request" | ./godot-dap-mcp-server | head -1
}
```

This spawns a NEW MCP server for each request = loses session!

**The Fix**: Applied the same file descriptor pattern to manual test.

**Key Learning**: When you find a solution to a fundamental problem (persistent sessions), audit ALL scripts that might have the same issue.

---

## Solutions and Patterns

### Pattern 1: Persistent Subprocess Communication

**Template for maintaining persistent bidirectional communication:**

```bash
#!/bin/bash

# 1. Create named pipes
rm -f /tmp/mcp-stdin /tmp/mcp-stdout
mkfifo /tmp/mcp-stdin /tmp/mcp-stdout

# 2. Open file descriptors (keeps pipes alive)
exec 3<>/tmp/mcp-stdin
exec 4<>/tmp/mcp-stdout

# 3. Start subprocess with pipes
./your-server < /tmp/mcp-stdin > /tmp/mcp-stdout 2>/dev/null &
SERVER_PID=$!

# 4. Setup cleanup trap
trap "exec 3>&- 4>&-; kill $SERVER_PID; rm -f /tmp/mcp-stdin /tmp/mcp-stdout" EXIT

# 5. Communication function
send_request() {
    echo "$1" >&3  # Write to FD 3
    read -r response <&4  # Read from FD 4
    echo "$response"
}

# 6. Use it
send_request '{"method":"test"}'
```

**When to Use This Pattern**:
- Any subprocess that needs to maintain state across requests
- Request/response protocols over stdin/stdout
- Session-based services

---

### Pattern 2: Port Conflict Handling

**Template for graceful port fallback:**

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
        echo "No available ports!"
        exit 1
    fi
else
    PORT=$DEFAULT_PORT
fi
```

**When to Use This Pattern**:
- Development tools that might have multiple instances running
- CI/CD environments where port conflicts are common
- Any networked service with a default port

---

### Pattern 3: DAP Protocol Handshake

**Complete handshake sequence:**

```go
// 1. Connect to DAP server
session := dap.NewSession("localhost", port)
if err := session.Connect(ctx); err != nil {
    return err
}

// 2. Initialize (get capabilities)
if err := session.Initialize(ctx); err != nil {
    return err
}

// 3. ConfigurationDone (transition to ready state)
if err := session.ConfigurationDone(ctx); err != nil {
    return err
}

// 4. Now you can set breakpoints, launch, etc.
```

**Why Each Step Matters**:
- `Connect()` - Establishes TCP connection
- `Initialize()` - Negotiates protocol version and capabilities
- `ConfigurationDone()` - Signals "I'm ready to debug" (CRITICAL!)
- Without `ConfigurationDone()`, server remains in "initialized" state and won't accept debugging commands

---

### Pattern 4: Robust JSON Parsing in Bash

**Template for parsing nested/escaped JSON:**

```bash
# Match both escaped and unescaped JSON
PATTERN='(\\"|")key(\\"|"):(\\"|")value'

if echo "$RESPONSE" | grep -qE "$PATTERN"; then
    echo "Match found!"
fi

# Extract values (handles escaped quotes)
VALUE=$(echo "$RESPONSE" | grep -oE '(\\"|")status(\\"|"):(\\"|")[^"\\]*' |
        sed 's/.*://' | tr -d '"\')
```

**When to Use This Pattern**:
- Parsing MCP tool results (which are JSON-in-JSON)
- Any situation where JSON might be escaped in a string
- Shell scripts that parse structured data

---

## Key Takeaways

### 1. Read the Protocol Docs Thoroughly

**Lesson**: We initially thought `Initialize()` was sufficient, but Godot's DAP requires `configurationDone` to transition to ready state.

**Best Practice**:
- Read protocol specifications completely, not just the basics
- Check for multi-step handshakes
- Look for state machine diagrams in docs

**Resources**:
- Godot DAP memories: `mcp__godot-source__read_memory("dap_supported_commands")`
- DAP specification: https://microsoft.github.io/debug-adapter-protocol/

---

### 2. Bash Named Pipes Require File Descriptors

**Lesson**: Named pipes close when the writer closes, causing EOF. File descriptors keep them open.

**Best Practice**:
```bash
# ❌ WRONG - Pipe closes after each write
echo "$data" > /tmp/pipe

# ✅ RIGHT - Use file descriptors
exec 3<>/tmp/pipe
echo "$data" >&3
```

**Why This Matters**: Many subprocess communication patterns need persistent connections.

---

### 3. Always Test Both Automated and Manual Flows

**Lesson**: We fixed the automated test but the manual test had identical issues because it used the old pattern.

**Best Practice**:
- When fixing a fundamental issue, grep for similar patterns in ALL scripts
- Maintain consistency across automated and manual test infrastructure
- Consider creating shared functions for common patterns

---

### 4. Protocol Limitations Need Workarounds

**Lesson**: Godot DAP requires absolute paths, not `res://` paths. This is a Godot limitation, not a protocol limitation.

**Best Practice**:
- Document limitations clearly
- Plan workarounds (Phase 7: add `res://` → absolute path conversion)
- Add to FAQ/troubleshooting docs
- Inform users about temporary constraints

---

### 5. Integration Tests Reveal Real-World Issues

**Issues Found Only Through Integration Testing**:
- `ConfigurationDone()` missing from handshake
- Path format requirements
- JSON escaping in responses
- Port conflicts in development

**Best Practice**:
- Write integration tests early
- Test with real servers/editors, not just mocks
- Automate as much as possible
- Document manual testing procedures for edge cases

---

## Future Considerations

### 1. Cross-Platform Testing

**Current Status**: Tested on macOS only

**TODO**:
- Test on Linux (different shell behaviors)
- Test on Windows (GitBash, WSL, PowerShell)
- Handle Windows path separators in test scripts

---

### 2. CI/CD Integration

**Considerations for GitHub Actions / GitLab CI**:
- Godot headless mode works in CI (verified!)
- Need to install Godot in CI environment
- Port conflicts less likely but still handle gracefully
- Timeout values may need adjustment for slower CI runners

**Sample CI Workflow**:
```yaml
- name: Install Godot
  run: |
    wget https://downloads.tuxfamily.org/godotengine/4.2.2/Godot_v4.2.2-stable_linux.x86_64.zip
    unzip Godot_v4.2.2-stable_linux.x86_64.zip
    export GODOT_BIN=./Godot_v4.2.2-stable_linux.x86_64

- name: Run Integration Tests
  run: ./scripts/automated-integration-test.sh
```

---

### 3. res:// Path Resolution (Phase 7)

**Design Approach**:
```go
// Add to Session struct
type Session struct {
    client      *Client
    state       SessionState
    projectRoot string  // NEW: for path resolution
}

// Helper function
func (s *Session) ResolveGodotPath(path string) string {
    if strings.HasPrefix(path, "res://") && s.projectRoot != "" {
        relPath := strings.TrimPrefix(path, "res://")
        return filepath.Join(s.projectRoot, relPath)
    }
    return path
}

// Usage in godot_set_breakpoint
file = session.ResolveGodotPath(params["file"])
```

**User Experience**:
```
# Specify project root once at connect
godot_connect(port=6006, project="/path/to/project")

# Then use res:// paths naturally
godot_set_breakpoint(file="res://player.gd", line=25)
```

---

### 4. Error Recovery and Reconnection

**Future Enhancement**: Handle Godot editor restart gracefully

**Approach**:
- Detect connection loss
- Auto-retry with exponential backoff
- Preserve breakpoint state across reconnections

---

### 5. Test Coverage Expansion

**Additional Test Scenarios**:
- Multiple breakpoints in different files
- Breakpoint on non-executable lines (should adjust line)
- Debugging multiple scenes
- Long-running game sessions
- Stress testing (rapid start/stop cycles)

---

## Conclusion

The journey to working integration tests taught us:

1. **Protocol knowledge is critical** - Read specs thoroughly
2. **Bash subprocess patterns** - File descriptors for persistent pipes
3. **Real testing reveals real issues** - Integration tests are invaluable
4. **Consistency across scripts** - Apply fixes everywhere
5. **Document as you go** - Future you will thank present you

**Final Test Results**:
```
✓ Found godot-dap-mcp-server binary
✓ Found Godot: 4.5.1.stable.official
✓ Godot started (PID: 16114)
✓ DAP server is ready on port 6006
✓ MCP server started (PID: 16123)

========================================
All Integration Tests Passed! ✓
========================================

Summary:
  ✓ Godot launched as subprocess (headless mode)
  ✓ DAP server started automatically
  ✓ MCP server initialized
  ✓ All 7 Phase 3 tools registered
  ✓ Connected to Godot DAP server
  ✓ Set and cleared breakpoint
  ✓ Clean disconnection
```

Phase 3 is complete! 🎉

---

## Addendum: stackTrace Verification Pattern

**Date**: 2025-11-15
**Context**: Enhancing DAP compliance testing with execution location verification

### Discovery

While implementing comprehensive DAP protocol compliance tests (`cmd/test-dap-protocol/main.go`), we discovered a critical verification pattern:

**Problem**: After sending stepping commands (stepIn, next), how do you verify the command actually worked?

**Solution**: Use `stackTrace` immediately after stepping to confirm execution location.

### Implementation Pattern

```json
// Test 7: Step into function
{
  "command": "stepIn",
  "arguments": { "threadId": 1 }
}

// Test 8: Verify we stepped into the expected function
{
  "command": "stackTrace",
  "arguments": { "threadId": 1 }
}
```

**Expected stackTrace response**:
```json
{
  "body": {
    "stackFrames": [
      {
        "id": 0,
        "name": "calculate_sum",  // ✅ Confirms stepIn worked
        "line": 15,               // ✅ At expected line
        "column": 0,              // Always 0 for GDScript
        "source": {
          "path": "/path/to/test_script.gd",
          "name": "test_script.gd",
          "checksums": [          // Godot includes both MD5 and SHA256
            {
              "algorithm": "MD5",
              "checksum": "187c6f2eb1e8016309f0c8875f1a2061"
            },
            {
              "algorithm": "SHA256",
              "checksum": "5baba6b7784def4d93dfd08e53aa306159cf0d22edaa323da1e7ca8eadbd888c"
            }
          ]
        }
      },
      {
        "id": 1,
        "name": "_ready",         // Caller function
        "line": 6,
        "column": 0
      }
    ]
  }
}
```

### Key Insights

1. **Checksums Included**: Godot's stackTrace responses include both MD5 and SHA256 checksums for source files. This is an optional DAP feature that Godot implements.

2. **Frame Ordering**: Stack frames are ordered innermost-first. The current execution point is always `stackFrames[0]`.

3. **Column Always Zero**: GDScript doesn't track column positions, so `column` is always 0.

4. **Frame IDs Sequential**: Frame IDs are 0, 1, 2... and are used for subsequent `scopes` and `variables` requests.

5. **Verification Workflow**:
   - Send stepping command (stepIn, next, stepOver)
   - Wait for response (command completes asynchronously)
   - Send stackTrace request
   - Verify `stackFrames[0].name` matches expected function
   - Verify `stackFrames[0].line` is expected line number

### Value for Testing

This pattern is now used in `cmd/test-dap-protocol/main.go` to:
- Verify stepIn successfully enters functions
- Confirm execution location after stepping
- Demonstrate proper DAP protocol workflows
- Test that Godot's stackTrace implementation includes all expected fields

### Related Documentation

- Pattern added to `critical_implementation_patterns.md` memory (Pattern #9)
- Enhanced `DAP_SESSION_GUIDE.md` with detailed stackTrace example showing checksums
- Test implementation: `cmd/test-dap-protocol/main.go` (Tests 7-8)

---

**Document Version**: 1.1
**Last Updated**: 2025-11-15
**Related Files**:
- `scripts/automated-integration-test.sh`
- `scripts/integration-test.sh`
- `tests/INTEGRATION_TEST.md`
- `internal/tools/connect.go`
- `cmd/test-dap-protocol/main.go`
