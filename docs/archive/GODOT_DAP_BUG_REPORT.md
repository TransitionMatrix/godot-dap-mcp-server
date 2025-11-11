# Godot DAP Server Dictionary Access Crash - Bug Report

**Date:** 2025-11-09
**Godot Version:** 4.6.dev.custom_build.6fd949a6d (2025-11-07 16:32:39 UTC)
**Affects:** Debug Adapter Protocol (DAP) server
**Severity:** Critical - Prevents DAP clients from setting breakpoints
**Status:** Confirmed regression (works in 4.5.1 stable)

---

## Summary

The Godot DAP server crashes with a Dictionary access error when processing valid `setBreakpoints` requests from DAP clients. The error occurs in `prepare_success_response()` when trying to access dictionary keys that don't exist, causing the DAP server to become non-functional.

---

## Error Output

```
ERROR: Bug: Dictionary::operator[] used when there was no value for the given key, please report.
   at: operator[] (core/variant/dictionary.cpp:136)
ERROR: Bug: Dictionary::operator[] used when there was no value for the given key, please report.
   at: operator[] (core/variant/dictionary.cpp:136)
```

The error appears twice, suggesting two separate failed dictionary accesses (likely `params["seq"]` and `params["command"]`).

---

## Steps to Reproduce

1. **Start Godot 4.6.dev nightly build:**
   ```bash
   godot-dev -e ~/Projects/xiang-qi-game-2d-godot
   ```

2. **Ensure DAP server is enabled:**
   - Editor → Editor Settings → Network → Debug Adapter
   - Verify DAP server is listening on port 6006

3. **Connect a DAP client and send a setBreakpoints request:**
   ```json
   Content-Length: 157

   {
     "seq": 3,
     "type": "request",
     "command": "setBreakpoints",
     "arguments": {
       "source": {
         "path": "res://src/scenes/main_scene/main_scene.gd"
       },
       "breakpoints": [
         {"line": 56}
       ]
     }
   }
   ```

4. **Observe the error:**
   - DAP client times out waiting for response
   - Godot editor prints Dictionary access error twice
   - No breakpoint is set

---

## Expected Behavior

1. Godot DAP server receives and parses the JSON request
2. Request is routed to `req_setBreakpoints` handler
3. Breakpoint is set at the specified line
4. Response is sent back to the client:
   ```json
   {
     "seq": <response_seq>,
     "type": "response",
     "request_seq": 3,
     "command": "setBreakpoints",
     "success": true,
     "body": {
       "breakpoints": [...]
     }
   }
   ```

---

## Actual Behavior

1. ✅ Godot receives the request
2. ✅ JSON is parsed successfully into a Dictionary
3. ❌ **Timeout check triggers** (possibly due to race condition or changed timeout behavior)
4. ❌ **Crash attempting to build error response** - `prepare_success_response()` tries to access `params["seq"]` and `params["command"]` without checking if keys exist
5. ❌ Dictionary access fails with `operator[]` error
6. ❌ No response sent to client
7. ❌ Client times out after 30 seconds

---

## Root Cause Analysis

### Code Location
`editor/debugger/debug_adapter/debug_adapter_parser.cpp`

### Problem Code

**Line 70-77: `prepare_success_response()`**
```cpp
Dictionary DebugAdapterParser::prepare_success_response(const Dictionary &p_params) const {
    Dictionary response;
    response["type"] = "response";
    response["request_seq"] = p_params["seq"];        // ❌ CRASHES if "seq" missing
    response["command"] = p_params["command"];        // ❌ CRASHES if "command" missing
    response["success"] = true;

    return response;
}
```

**Line 860-864: Timeout check in `process_message()`**
```cpp
if (OS::get_singleton()->get_ticks_msec() - _current_peer->timestamp > _request_timeout) {
    Dictionary response = parser->prepare_error_response(params, DAP::ErrorType::TIMEOUT);
    _current_peer->res_queue.push_front(response);
    return true;
}
```

### Why This Happens

The code uses C++'s `Dictionary::operator[]` which **crashes** when accessing non-existent keys instead of returning null or throwing a catchable exception. This is a known Godot Dictionary anti-pattern.

The bug is triggered when:
1. A timeout occurs (possibly due to a race condition in 4.6.dev)
2. Godot tries to build a timeout error response
3. The `params` Dictionary is empty or corrupt
4. Accessing `params["seq"]` crashes immediately

### Why It Doesn't Happen with `initialize`

The `initialize` command succeeds because:
- It's the first command sent
- No previous state corruption
- Likely completes before any timeout can trigger

Subsequent commands like `setBreakpoints` encounter the issue, possibly due to:
- State corruption from previous processing
- Changed timing behavior in 4.6.dev
- Race conditions in the polling loop

---

## Proposed Fix

### Option 1: Defensive Dictionary Access (Recommended)

Add key existence checks before accessing:

```cpp
Dictionary DebugAdapterParser::prepare_success_response(const Dictionary &p_params) const {
    Dictionary response;
    response["type"] = "response";

    // Defensive access
    if (p_params.has("seq")) {
        response["request_seq"] = p_params["seq"];
    } else {
        ERR_PRINT("Missing 'seq' field in DAP request");
        response["request_seq"] = 0;  // Default value
    }

    if (p_params.has("command")) {
        response["command"] = p_params["command"];
    } else {
        ERR_PRINT("Missing 'command' field in DAP request");
        response["command"] = "";
    }

    response["success"] = true;
    return response;
}
```

Apply same fix to `prepare_error_response()` at line 80-124.

### Option 2: Use Dictionary::get() with Default Values

```cpp
Dictionary DebugAdapterParser::prepare_success_response(const Dictionary &p_params) const {
    Dictionary response;
    response["type"] = "response";
    response["request_seq"] = p_params.get("seq", 0);      // Returns 0 if missing
    response["command"] = p_params.get("command", "");     // Returns "" if missing
    response["success"] = true;
    return response;
}
```

### Option 3: Fix Root Cause - Timeout Logic

Investigate why the timeout is triggering inappropriately in 4.6.dev:

1. Check if `_current_peer->timestamp` is being set correctly (line 99-101 in `debug_adapter_protocol.cpp`)
2. Verify timeout value is loaded from settings (default: 5000ms, line 121 in `debug_adapter_protocol.h`)
3. Check for race conditions in the polling loop (line 1201-1228 in `debug_adapter_protocol.cpp`)

---

## Testing the Fix

### Test Case 1: Basic setBreakpoints
```bash
# 1. Apply fix to Godot source
# 2. Rebuild Godot: scons platform=macos

# 3. Start Godot with test project
godot -e ~/Projects/test-project

# 4. Connect DAP client and send:
{
  "seq": 3,
  "type": "request",
  "command": "setBreakpoints",
  "arguments": {
    "source": {"path": "res://test.gd"},
    "breakpoints": [{"line": 10}]
  }
}

# Expected: Response received, no Dictionary errors
```

### Test Case 2: Malformed Request (Edge Case)
```bash
# Send request with missing fields
{
  "type": "request",
  "command": "setBreakpoints"
  # Missing "seq" field
}

# Expected: Graceful error response, no crash
```

### Test Case 3: Timeout Scenario
```bash
# Set very low timeout in editor settings
# Send slow request
# Expected: Timeout error response, no crash
```

---

## Verification Commands

### Check Godot Version
```bash
godot --version
# Should show: 4.6.dev.custom_build.6fd949a6d or later with fix
```

### Search for Unsafe Dictionary Access
```bash
# Find all unsafe Dictionary accesses in DAP code
cd /Users/adp/Projects/godot
grep -n 'p_params\["' editor/debugger/debug_adapter/*.cpp
```

### Run Godot with DAP Logging
```bash
# Enable verbose logging
godot --verbose -e ~/Projects/test-project
# Watch for DAP-related output
```

---

## Additional Context

### Affected Files
- `editor/debugger/debug_adapter/debug_adapter_parser.cpp` (lines 70-77, 80-124)
- `editor/debugger/debug_adapter/debug_adapter_protocol.cpp` (line 860-864)

### DAP Client Used for Testing
- **Client:** godot-dap-mcp-server (MCP server for AI agents)
- **Library:** github.com/google/go-dap (official DAP Go implementation)
- **Verified Request Structure:** Matches DAP specification exactly

### Request Structure Verification
```json
{
  "seq": 3,
  "type": "request",
  "command": "setBreakpoints",
  "arguments": {
    "source": {
      "path": "res://src/scenes/main_scene/main_scene.gd"
    },
    "breakpoints": [
      {"line": 56}
    ]
  }
}
```

This structure is **100% valid** according to the Debug Adapter Protocol specification.

---

## Related Issues

- Similar pattern exists in multiple request handlers
- All use `params["field"]` without checking existence
- Could affect other DAP commands under timeout conditions

### Other Potentially Affected Functions
```cpp
// Line 129: req_initialize
Dictionary args = p_params["arguments"];  // Could crash

// Line 170: req_launch
Dictionary args = p_params["arguments"];  // Could crash

// Line 366: req_setBreakpoints
Dictionary args = p_params["arguments"];  // Could crash

// Many more...
```

**Recommendation:** Audit all DAP parser functions for unsafe Dictionary access.

---

## Regression Status

### Works in Godot 4.5.1 Stable ✅
- setBreakpoints command succeeds
- No Dictionary errors
- Proper responses sent

### Broken in Godot 4.6.dev Nightly ❌
- setBreakpoints command fails
- Dictionary access crashes
- No responses sent

### Likely Cause of Regression
- Changed timeout behavior
- Modified polling loop timing
- Race condition introduced in state management
- Dictionary implementation changes

---

## Impact Assessment

**Severity:** Critical
**User Impact:** DAP debugging completely non-functional
**Workaround:** Use Godot 4.5.1 stable
**Fix Complexity:** Low (add defensive checks)
**Testing Required:** Moderate (test all DAP commands)

---

## Fix Priority

**P0 - Critical** - This blocks all DAP clients from using breakpoints, which is a core debugging feature. The fix is straightforward and low-risk.

---

## Notes for Implementation

1. **Apply fix to all request/response builders** - Don't just fix `prepare_success_response`
2. **Add unit tests** - Test Dictionary access with missing keys
3. **Add logging** - Log when missing keys are encountered
4. **Consider protocol validation** - Add early validation that rejects malformed requests
5. **Review Dictionary usage patterns** - Audit codebase for similar anti-patterns

---

## References

- DAP Specification: https://microsoft.github.io/debug-adapter-protocol/specification
- Godot Dictionary Documentation: https://docs.godotengine.org/en/stable/classes/class_dictionary.html
- Related Godot Issue: https://github.com/godotengine/godot/issues/108288 (seq field type conversion)
