# Integration Test Guide

This guide walks through testing the godot-dap-mcp-server with Godot.

## Prerequisites

1. **Build the MCP server**:
   ```bash
   go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go
   ```

2. **Install Godot 4.2.2+**:
   - Download from: https://godotengine.org/download
   - Ensure `godot` is in PATH, or set `GODOT_BIN` environment variable

## Automated Integration Test (RECOMMENDED)

**Fully automated** - Launches Godot as subprocess, no manual setup required!

```bash
./scripts/automated-integration-test.sh
```

**What it does:**
1. ✅ Launches Godot in headless mode as subprocess
2. ✅ Waits for DAP server to be ready (auto-starts on port 6006)
3. ✅ Tests MCP server initialization
4. ✅ Tests tool registration (all 7 Phase 3 tools)
5. ✅ Connects to Godot DAP server
6. ✅ Sets and clears breakpoints
7. ✅ Tests disconnection
8. ✅ Terminates Godot subprocess cleanly

**Expected output:**
```
========================================
Godot DAP MCP Server
Automated Integration Test
========================================

✓ Found godot-dap-mcp-server binary
✓ Found Godot: 4.3.stable.official
✓ Found test project: /path/to/tests/fixtures/test-project

Launching Godot editor in headless mode...
✓ Godot started (PID: 12345)
✓ DAP server is ready on port 6006

========================================
Running Integration Tests
========================================

[... all tests pass ...]

========================================
All Integration Tests Passed! ✓
========================================
```

**Configuration:**
- Custom Godot path: `export GODOT_BIN=/path/to/godot`
- Custom DAP port: Edit `DAP_PORT` in script (default: 6006)

## Manual Integration Test (ALTERNATIVE)

If you prefer manual control or automated test fails:

```bash
./scripts/integration-test.sh
```

**Prerequisites:**
1. Godot editor must be running manually
2. Test project must be opened: `tests/fixtures/test-project/`
3. DAP must be enabled in **Editor → Editor Settings → Network → Debug Adapter**
4. Port set to 6006

This script tests the same functionality but requires manual Godot setup.

## Manual End-to-End Test

Test the complete debugging workflow:

### 1. Start MCP Server

```bash
./godot-dap-mcp-server
```

The server will wait for JSON-RPC requests on stdin.

### 2. Initialize MCP Session

Send (via stdin):
```json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1.0","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}
```

### 3. List Available Tools

```json
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
```

You should see all 8 tools:
- godot_ping
- godot_connect
- godot_disconnect
- godot_continue
- godot_step_over
- godot_step_into
- godot_set_breakpoint
- godot_clear_breakpoint

### 4. Connect to Godot

```json
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"godot_connect","arguments":{"port":6006}}}
```

**Expected response:**
```json
{"jsonrpc":"2.0","id":3,"result":{"status":"connected","message":"Connected to Godot DAP server at localhost:6006","state":"configured"}}
```

### 5. Set Breakpoint

```json
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"godot_set_breakpoint","arguments":{"file":"res://test_script.gd","line":13}}}
```

**Expected response:**
```json
{"jsonrpc":"2.0","id":4,"result":{"status":"verified","message":"Breakpoint set at res://test_script.gd:13","file":"res://test_script.gd","requested_line":13,"actual_line":13,"id":1}}
```

### 6. Run Scene in Godot

- Press **F5** in Godot editor
- Scene should run and pause at line 13 in test_script.gd
- Godot debugger should show "Paused"

### 7. Test Stepping Commands

**Step Over:**
```json
{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"godot_step_over","arguments":{}}}
```

**Step Into:**
```json
{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"godot_step_into","arguments":{}}}
```

**Continue:**
```json
{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"godot_continue","arguments":{}}}
```

### 8. Clear Breakpoint

```json
{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"godot_clear_breakpoint","arguments":{"file":"res://test_script.gd"}}}
```

### 9. Disconnect

```json
{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"godot_disconnect","arguments":{}}}
```

## Troubleshooting

### Connection Failed

**Error:** `Failed to connect to Godot DAP server at localhost:6006`

**Solutions:**
1. Verify Godot editor is running
2. Check DAP server is enabled in Editor Settings
3. Confirm port is 6006
4. Try clicking "Start Server" in DAP settings

### Breakpoint Not Verified

**Status:** `"status":"unverified"`

**Causes:**
1. File not loaded in Godot yet (run scene once first)
2. Incorrect file path (use `res://` prefix)
3. Invalid line number (blank line or comment)

**Solutions:**
1. Run the scene once (F5) to load scripts
2. Use absolute `res://` paths
3. Set breakpoint on executable code line

### Step Commands Not Working

**Possible causes:**
1. Not paused at breakpoint
2. No active debugging session
3. Game already exited

**Solutions:**
1. Ensure scene is running and paused
2. Check Godot debugger shows "Paused" state
3. Reconnect if session was lost

## Test Project Details

**Location:** `tests/fixtures/test-project/`

**Test Script:** `test_script.gd`
- Line 13: Inside `calculate_sum()` function
- Line 18: Inside `test_loop()` function

**Expected Behavior:**
- Scene prints output when running
- Pauses at breakpoints when set
- Stepping advances one line at a time
- Variables visible in Godot debugger

## Success Criteria

✅ All automated tests pass
✅ Connection established to Godot
✅ Breakpoint verified and hit when running scene
✅ Step commands advance execution
✅ Continue resumes to next breakpoint or completion
✅ Clean disconnect without errors

## Next Steps

After successful integration testing:
1. Test with real Godot game projects
2. Test with multiple breakpoints
3. Test error conditions (invalid paths, etc.)
4. Test with Claude Code MCP client
