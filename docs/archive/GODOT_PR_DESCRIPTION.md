# PR: Fix DAP Dictionary crashes and launch timeout

Fixes #XXXXX

## Summary

This PR fixes two critical bugs in the Debug Adapter Protocol (DAP) implementation that prevent DAP clients from debugging Godot games:

1. **Dictionary Safety**: Fixes 31 unsafe Dictionary accesses causing crashes
2. **Launch Response**: Fixes missing launch response causing 30-second timeout

Both bugs have existed since the DAP implementation was added in 2021 but were masked by old Dictionary behavior.

## Problem

### Issue 1: Dictionary Access Crashes

**Symptom:**
```
ERROR: Bug: Dictionary::operator[] used when there was no value for the given key, please report.
   at: operator[] (core/variant/dictionary.cpp:136)
```

**Root Cause:**
The DAP code used unsafe `dict["key"]` access patterns throughout. After commit 0be2a771 fixed Dictionary const-correctness, these unsafe accesses began triggering errors instead of silently inserting default values.

**Impact:**
- All DAP commands crash when clients omit optional fields
- DAP clients cannot connect or debug games
- Complete debugging failure

### Issue 2: Launch Request Timeout

**Symptom:**
Launch requests timeout after 30 seconds with no response.

**Root Cause:**
The `req_launch()` handler returns empty `Dictionary()` instead of a success response. While launch execution is correctly deferred until `configurationDone` (per DAP spec), the launch REQUEST still needs immediate acknowledgment.

**Impact:**
- DAP clients wait 30 seconds before timing out
- Poor user experience even when everything else works
- Violates DAP protocol requirement that every request gets a response

## Solution

### Fix 1: Dictionary Safety (31 locations)

Replace unsafe accesses with safe `.get()` calls:

```cpp
// Before (unsafe)
Dictionary args = p_params["arguments"];
String value = args["field"];

// After (safe)
Dictionary args = p_params.get("arguments", Dictionary());
String value = args.get("field", "");
```

**Files changed:**
- `debug_adapter_parser.cpp`: 25 fixes
- `debug_adapter_protocol.cpp`: 6 fixes
- `debug_adapter_types.h`: 2 fixes

### Fix 2: Launch Response (1 location)

Return proper success response:

```cpp
// Before
return Dictionary();  // No response sent!

// After
return prepare_success_response(p_params);  // Response sent immediately
```

**Note:** Launch execution is still correctly deferred until `configurationDone`. This fix only addresses the protocol response.

## Testing

Tested with comprehensive integration test exercising all 13 DAP commands:

**Before fixes:**
- ❌ All commands crashed or timed out
- ❌ Dictionary errors in console
- ❌ 0% success rate

**After fixes:**
- ✅ Connect, Initialize, Launch working
- ✅ Breakpoints set and verified
- ✅ Runtime inspection (threads, stack, variables)
- ✅ Execution control (step, continue)
- ✅ Expression evaluation
- ✅ No Dictionary errors
- ✅ 85% success rate (11/13 commands working)

**Remaining limitations** (pre-existing, not addressed by this PR):
- `pause` - Not implemented in DAP server
- `setVariable` - Advertised but not implemented

Test program available at: https://github.com/TransitionMatrix/godot-dap-mcp-server/blob/main/cmd/debug-workflow-test/main.go

## Testing Note

The DAP implementation currently has no unit test coverage. Adding comprehensive unit tests would require editor infrastructure (EditorDebuggerServer, ScriptDebugger) that doesn't exist in the standard test framework. The fix has been validated through real DAP client testing showing 85% success rate (11/13 commands working).

## Rationale

These are **bug fixes** for latent issues exposed by correctness improvements:

1. **Dictionary Safety**: Follows established Godot pattern from commit 0be2a771 (AudioStreamWav fixes)
2. **Launch Response**: Restores DAP protocol compliance (every request must get response)
3. **Size**: Only 32 lines changed (mechanical safety fixes)
4. **Risk**: Low - defensive programming improvements
5. **Benefit**: Enables DAP debugging for all clients

## Related Issues

Fixes #XXXXX - DAP implementation has unsafe Dictionary access patterns causing crashes

This PR may also help with:
- #110749 - DAP crash when removing breakpoint (also Dictionary-related)
- #106969 - Debug Android export via DAP doesn't work

## Checklist

- [x] Code follows Godot coding style
- [x] Fixes follow established patterns (commit 0be2a771)
- [x] All changes are defensive programming improvements
- [x] Tested with real DAP client (godot-dap-mcp-server)
- [x] No breaking changes
- [x] Commit messages follow Godot conventions
- [x] References GitHub issue #XXXXX
