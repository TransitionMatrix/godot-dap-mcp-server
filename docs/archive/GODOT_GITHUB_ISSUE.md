# GitHub Issue Draft for Godot Repository

## Title
`[Debugger] DAP implementation has unsafe Dictionary access patterns causing crashes and launch timeouts`

## Labels
- `bug`
- `debugger`
- `priority:high`

## Description

### Summary

The Debug Adapter Protocol (DAP) implementation in Godot contains two critical bugs that completely break debugging for external DAP clients:

1. **Dictionary Safety**: 31 unsafe Dictionary accesses causing crashes
2. **Launch Response**: Missing launch response causing 30-second timeouts

Both bugs are latent issues that have existed since the DAP implementation was added in 2021, but were recently exposed by Dictionary const-correctness improvements in commit 0be2a771.

### Issue 1: Dictionary Access Crashes

**Error Message:**
```
ERROR: Bug: Dictionary::operator[] used when there was no value for the given key, please report.
   at: operator[] (core/variant/dictionary.cpp:136)
```

**Root Cause:**
The DAP server code uses unsafe `dict["key"]` access patterns throughout. After commit 0be2a771 fixed Dictionary const-correctness, these unsafe accesses began triggering errors instead of silently inserting default values.

**Affected Files:**
- `editor/debugger/debug_adapter/debug_adapter_parser.cpp` (25 locations)
- `editor/debugger/debug_adapter/debug_adapter_protocol.cpp` (6 locations)
- `editor/debugger/debug_adapter/debug_adapter_types.h` (2 locations)

**Impact:**
- All DAP commands crash when clients omit optional fields (per DAP spec, most fields are optional)
- External DAP clients (VS Code extensions, AI debugging tools) cannot connect or debug games
- Complete debugging failure for any non-Godot-editor DAP client

**Example Crash:**
```cpp
// Current unsafe code
Dictionary args = p_params["arguments"];  // Crashes if "arguments" not present
String value = args["field"];              // Crashes if "field" not present
```

### Issue 2: Launch Request Timeout

**Symptom:**
Launch requests timeout after 30 seconds with no response from Godot.

**Root Cause:**
The `req_launch()` handler in `debug_adapter_parser.cpp:184` returns `Dictionary()` instead of a success response. While launch execution is correctly deferred until `configurationDone` (per DAP spec), the launch REQUEST still requires an immediate acknowledgment.

**Impact:**
- DAP clients wait 30 seconds before timing out
- Poor user experience even when launch eventually succeeds
- Violates DAP protocol requirement that every request receives a response

### Steps to Reproduce

1. **Setup:**
   - Build Godot from master
   - Enable DAP server: Editor → Editor Settings → Network → Debug Adapter
   - Use any external DAP client (VS Code extension, custom client)

2. **Trigger Dictionary Crash:**
   ```jsonrpc
   // Send initialize request without optional fields
   {"type": "request", "seq": 1, "command": "initialize", "arguments": {}}

   // Result: Godot console shows Dictionary error
   ```

3. **Trigger Launch Timeout:**
   ```jsonrpc
   // Send launch request
   {"type": "request", "seq": 2, "command": "launch", "arguments": {"project": "/path"}}

   // Result: No response received, client times out after 30s
   ```

4. **Observe:**
   - Godot console shows Dictionary errors
   - No responses sent to client
   - DAP debugging completely non-functional

### Testing Evidence

**Before Fixes:**
- ❌ All DAP commands crashed or timed out
- ❌ Dictionary errors in console for every command
- ❌ 0% success rate (0/13 commands working)

**After Fixes:**
- ✅ Connect, Initialize, Launch working
- ✅ Breakpoints set and verified
- ✅ Runtime inspection (threads, stack, variables)
- ✅ Execution control (step-over, step-into, continue)
- ✅ Expression evaluation
- ✅ No Dictionary errors
- ✅ 85% success rate (11/13 commands working)

Test program: https://github.com/TransitionMatrix/godot-dap-mcp-server

### Related Issues

This may be related to:
- #110749 - DAP crash when removing breakpoint (also Dictionary-related)
- #106969 - Debug Android export via DAP doesn't work
- #108518 - DAP disconnects immediately after launch (closed)

### Proposed Solution

**1. Dictionary Safety (31 locations):**
Replace unsafe accesses with safe `.get()` calls:

```cpp
// Before (unsafe)
Dictionary args = p_params["arguments"];
String value = args["field"];

// After (safe)
Dictionary args = p_params.get("arguments", Dictionary());
String value = args.get("field", "");
```

**2. Launch Response (1 location):**
Return proper success response:

```cpp
// Before
return Dictionary();  // No response sent!

// After
return prepare_success_response(p_params);  // Response sent immediately
```

**Note:** Launch execution remains correctly deferred until `configurationDone`. This fix only addresses the protocol response requirement.

### Additional Context

- **Pattern Precedent:** Commit 0be2a771 fixed similar Dictionary issues in AudioStreamWav
- **DAP Spec Compliance:** Most DAP protocol fields are optional; server must handle missing fields gracefully
- **Impact Scope:** All external DAP clients (not just Godot editor's built-in debugger)
- **Risk:** Low - defensive programming improvements only
- **Lines Changed:** 32 lines across 3 files

### Fix Availability

I have a PR ready with both fixes, including:
- All 32 safety fixes applied
- Testing with real DAP client showing 85% success rate (11/13 commands working)
- Follows established Godot patterns (commit 0be2a771)

**Note on Testing**: The DAP implementation currently has no unit test coverage. Adding comprehensive unit tests would require editor infrastructure (EditorDebuggerServer, ScriptDebugger) that doesn't exist in the standard test framework. The fix has been validated with real DAP client testing showing 85% success rate (11/13 commands working).

Would you like me to submit the PR? I can link it to this issue once created.

---

**Environment:**
- Godot Version: 4.6.dev (also affects 4.5.stable)
- Platform: All platforms
- Godot Commit: master (post-0be2a771)
