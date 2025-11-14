# DAP Compliance Test - Dictionary Error Analysis

## Summary

The automated DAP compliance test sends minimally-valid DAP requests to identify unsafe Dictionary accesses in Godot's DAP server. Despite sending only spec-required fields, the test triggers 2 Dictionary errors from Test 3 (`setBreakpoints`). This document explains where those errors come from and why they demonstrate Godot's non-compliance with optional field handling.

## Test Sequence (Final)

**Optimized sequence** (follows DAP spec configuration phase):
1. **Test 1** (seq=1): `initialize` → ✅ Success
2. **Test 2** (seq=2): `launch` with project/scene/noDebug → Stores in pending_launch, returns empty Dict (expected)
3. **Test 3** (seq=3): `setBreakpoints` without breakpoints array → ❌ **Triggers Dictionary errors** (configuration phase)
4. **Test 4** (seq=4): `configurationDone` → ✅ Triggers _launch_process with pending_launch, returns both responses

**Why this order:**
- DAP spec allows configuration requests (like `setBreakpoints`) **after** `launch` but **before** `configurationDone`
- This is the "configuration phase" where breakpoints, exception filters, etc. are set
- `configurationDone` signals the end of configuration and triggers the actual game launch

## The 2 Dictionary Errors

### Error #1: `req_setBreakpoints` line 385

**Location**: `editor/debugger/debug_adapter/debug_adapter_parser.cpp:385`

```cpp
Array breakpoints = args["breakpoints"], lines;  // ❌ UNSAFE
```

**Problem**: Test 4 sends `setBreakpoints` without the `breakpoints` array, which is **optional** per DAP specification. Omitting `breakpoints` means "clear all breakpoints in this file."

**DAP Spec**: According to the Debug Adapter Protocol specification, the `breakpoints` field in `SetBreakpointsArguments` is optional:
- If present: Set these breakpoints
- If absent: Clear all breakpoints in the file

**Test 4 Request**:
```json
{
  "arguments": {
    "source": {
      "path": "/test.gd"
    }
  },
  "command": "setBreakpoints",
  "seq": 4,
  "type": "request"
}
```

**Fix**: Use safe Dictionary access:
```cpp
Array breakpoints = args.get("breakpoints", Array()), lines;
```

### Error #2: `Source::from_json` lines 87-89

**Location**: `editor/debugger/debug_adapter/debug_adapter_types.h:87-89`

```cpp
_FORCE_INLINE_ void from_json(const Dictionary &p_params) {
    name = p_params["name"];           // ❌ UNSAFE line 87
    path = p_params["path"];           // ❌ UNSAFE line 88 (but Test 4 has this)
    _checksums = p_params["checksums"]; // ❌ UNSAFE line 89
}
```

**Problem**: Test 4's `source` object only contains `"path"`, missing optional fields `"name"` and `"checksums"`.

**DAP Spec**: According to the Debug Adapter Protocol specification, `Source` has these fields:
- `path` (optional): The path of the source file
- `name` (optional): The short name of the source
- `sourceReference` (optional): Reference to source content
- `checksums` (optional): Array of checksums for the file
- etc.

**Test 4's Source Object**:
```json
{
  "path": "/test.gd"
}
```

**Fix**: Use safe Dictionary access:
```cpp
_FORCE_INLINE_ void from_json(const Dictionary &p_params) {
    name = p_params.get("name", "");
    path = p_params.get("path", "");
    _checksums = p_params.get("checksums", Array());
}
```

## Why Test 2 (Launch) Doesn't Cause Errors

Test 2's `launch` request includes all commonly-used fields:

```json
{
  "arguments": {
    "noDebug": false,
    "project": "/Users/adp/Projects/godot-dap-mcp-server/tests/fixtures/test-project",
    "scene": "main"
  },
  "command": "launch",
  "seq": 3,
  "type": "request"
}
```

The `req_launch` function stores these parameters in `pending_launch` and returns an empty Dictionary (which means no response is sent immediately - this is correct DAP behavior). No unsafe Dictionary accesses are triggered because:
- Line 170: `p_params["arguments"]` exists ✅
- Line 171-173: Only executed if `args.has("project")` is true, and uses safe access
- Line 179: Only executed if `args.has("godot/custom_data")` is true

## Client Capabilities Configuration (LLM-Optimized)

**Context:** Since our DAP client will be used by LLMs via MCP, we configure capabilities for human-friendly output and information richness.

### Test 1: Initialize Request with Client Capabilities

```json
{
  "arguments": {
    "adapterID": "godot",

    // Client identification
    "clientID": "godot-dap-mcp-server",
    "clientName": "Godot DAP MCP Server",

    // Human-friendly numbering (1-based)
    "linesStartAt1": true,
    "columnsStartAt1": true,

    // Information richness
    "supportsVariableType": true,

    // Path format (standard file paths)
    "pathFormat": "path",

    // Advanced features we don't implement
    "supportsInvalidatedEvent": false,
    "supportsProgressReporting": false,
    "supportsRunInTerminalRequest": false,
    "supportsVariablePaging": false
  },
  "command": "initialize",
  "seq": 1,
  "type": "request"
}
```

### Design Rationale

**For LLM/MCP Use Case:**
1. **Human-aligned numbering**: LLMs interact with humans, so use 1-based indexing
   - When Claude says "set breakpoint at line 10", users expect line 10 (not 11)
   - Matches behavior of all major DAP clients (VSCode, Zed, Neovim, Rider)

2. **Information richness**: Request type information for better debugging
   - `supportsVariableType: true` → Godot includes type field in variables
   - Helps Claude understand `player: CharacterBody2D` vs `player: Node`

3. **Client identification**: Clearly identify ourselves for debugging
   - `clientID` and `clientName` appear in Godot logs/telemetry
   - Helps debug issues specific to our MCP client

4. **Honest feature reporting**: Only claim features we implement
   - Set unsupported features to `false` rather than omitting
   - Makes our capabilities explicit and clear

### Godot's Safe Handling of Client Capabilities

**Important finding:** Godot uses SAFE Dictionary access for reading client capabilities:

```cpp
// From req_initialize (lines 133-136):
peer->linesStartAt1 = args.get("linesStartAt1", false);
peer->columnsStartAt1 = args.get("columnsStartAt1", false);
peer->supportsVariableType = args.get("supportsVariableType", false);
peer->supportsInvalidatedEvent = args.get("supportsInvalidatedEvent", false);
```

**This demonstrates:**
- Godot KNOWS how to use safe Dictionary access (`.get()` with defaults)
- The unsafe accesses (`dict["key"]`) elsewhere are implementation oversights
- Client capabilities work correctly whether present, absent, or false

### Test Results with Client Capabilities

**Test execution confirmed:**
- ✅ Initialize succeeds with all capabilities
- ✅ Godot correctly reads and stores all capability flags
- ✅ Still triggers 2 Dictionary errors from Test 3 (setBreakpoints)
- ✅ Demonstrates Godot inconsistently uses safe vs unsafe Dictionary access

**Capability usage confirmed:**
- `linesStartAt1`/`columnsStartAt1`: Used in `req_stackTrace` and `req_setBreakpoints` to adjust line numbers
- `supportsVariableType`: Used in `req_variables` to strip type field if false
- `supportsInvalidatedEvent`: Stored but no usage found

## Why "_launch_process" Isn't Called

The test sequence sends `configurationDone` (Test 2) BEFORE any `launch` request (Tests 3 and 5).

**DAP Specification:**
According to the DAP spec, the protocol allows flexible ordering of `launch` and `configurationDone` requests, as long as the **responses** are sent in the correct order:
- `configurationDone` response must be sent before `launch` response

**Godot's Implementation:**
Godot implements a **compliant subset** of this flexibility:
- **Requires**: `launch` request → `configurationDone` request (in that order)
- **Guarantees**: `configurationDone` response → `launch` response (correct order)
- **Mechanism**:
  1. `req_launch` stores parameters in `pending_launch`, returns empty Dict (no response sent)
  2. `req_configurationDone` calls `_launch_process(pending_launch)` and queues responses:
     - First: configurationDone response (via `push_front`)
     - Second: launch response (via `push_back`, added before push_front)

**Our test sequence** (incorrect for Godot):
1. `initialize`
2. `configurationDone` (pending_launch is empty, _launch_process not called) ❌
3. `launch` (stores parameters, but configurationDone already happened) ❌

**Correct sequence** (for Godot):
1. `initialize`
2. `launch` (stores parameters in pending_launch)
3. `configurationDone` (triggers _launch_process with stored parameters) ✅

**Why this matters:**
While our test sequence is incorrect for actually launching the game, it doesn't affect the Dictionary error analysis because:
- The errors come from `req_setBreakpoints` and `Source::from_json`, not from `_launch_process`
- `_launch_process` is never called in this test sequence
- The Dictionary errors occur regardless of whether the launch completes

## Implications for Upstream Godot PR

These findings demonstrate that:

1. **Godot violates the DAP specification** by using unsafe Dictionary access for optional fields
2. **Multiple locations need fixing**:
   - `req_setBreakpoints` (breakpoints array)
   - `Source::from_json` (name, checksums)
   - And potentially many other locations (see `docs/research/godot-dictionary-safety-issues.md`)

3. **The fix is simple**: Replace `dict["key"]` with `dict.get("key", default)` for all optional DAP fields

4. **Testing is essential**: The compliance tester successfully identified these issues by sending minimally-valid DAP messages (only required fields)

## Next Steps

1. ✅ Identified root cause of 2 Dictionary errors
2. ✅ Fixed test sequence to send `launch` before `configurationDone`
3. ✅ Re-run test to verify our analysis (confirmed 2 errors from setBreakpoints)
4. ✅ Configured LLM-optimized client capabilities (1-based indexing, type support, client identification)
5. Create comprehensive list of all unsafe Dictionary accesses in Godot DAP server
6. Prepare GitHub issue with specific line numbers and recommended fixes
