# DAP Launch Response Fix - Technical Explanation

**Date:** November 9, 2025
**Issue:** Launch request timing out after 30 seconds
**Root Cause:** Missing response to launch request
**Fix:** One-line change to return success response

---

## The Problem

The DAP protocol requires that **every request receives a response**. This is fundamental to the request/response pattern - clients track requests by sequence number and wait for matching responses.

### What Was Happening

```cpp
// In req_launch() - BEFORE fix (line 184)
DebugAdapterProtocol::get_singleton()->get_current_peer()->pending_launch = p_params;
return Dictionary();  // ❌ Returns EMPTY - no response sent!
```

**The flow:**
1. Client sends `launch` request (seq=2)
2. Godot stores launch params in `pending_launch`
3. Godot returns empty Dictionary
4. Protocol handler sees empty Dictionary and **doesn't queue a response**
5. **Client waits forever** for seq=2 response (times out after 30s)

---

## Why Empty Dictionary?

Looking at the protocol handler in `debug_adapter_protocol.cpp`:

```cpp
Dictionary response = parser->callv(command, args);
if (!response.is_empty()) {
    _current_peer->res_queue.push_front(response);  // Queue response
} else {
    // Launch request needs to be deferred until we receive a configurationDone request.
    if (command != "req_launch") {
        completed = false;
    }
}
```

The empty Dictionary was a signal that **"this command doesn't need immediate handling"**. But this conflates two concepts:
- **Execution deferral** (launch defers until configurationDone) ✓ Correct
- **Response deferral** (no response sent) ✗ **Protocol violation!**

---

## The Confusion: Deferred Launch vs Deferred Response

Godot correctly implements the DAP pattern where:
1. `launch` request stores parameters but doesn't start the game
2. `configurationDone` triggers the actual launch
3. This allows setting breakpoints before the game starts

**But** - the `launch` REQUEST itself still needs a response! The response just means "I received your launch request and stored it", not "I launched the game".

This is a subtle but critical distinction:
- ✓ **Deferred execution** = Correct (launch waits for configurationDone)
- ✗ **Deferred response** = Protocol violation (every request needs response)

---

## The Fix

```cpp
// In req_launch() - AFTER fix (line 184)
DebugAdapterProtocol::get_singleton()->get_current_peer()->pending_launch = p_params;
return prepare_success_response(p_params);  // ✅ Returns success response!
```

**Result:**
- ✅ Client gets immediate response to `launch` request
- ✅ Launch execution still deferred until `configurationDone`
- ✅ Protocol satisfied (every request gets response)
- ✅ No timeout

---

## Visual Flow Comparison

### BEFORE (broken)

```
Client                          Godot
  |                               |
  |-- launch (seq=2) ----------->|
  |                               | (stores in pending_launch)
  |                               | (returns empty Dictionary)
  |                               | (no response queued)
  |   ⏳ waiting...                |
  |   ⏳ waiting...                |
  |   ⏳ TIMEOUT! (30s)            |
```

### AFTER (fixed)

```
Client                          Godot
  |                               |
  |-- launch (seq=2) ----------->|
  |                               | (stores in pending_launch)
  |<- launch response (seq=2) ---| ✓ Success!
  |                               |
  |-- configurationDone (seq=3) ->|
  |                               | (triggers actual launch from pending_launch)
  |<- configurationDone resp ----| ✓ Success!
  |                               | [Game starts]
```

---

## Test Results

**Before fix:**
```
2025/11/09 16:03:21 3. Launching main scene...
2025/11/09 16:03:51 Failed to launch scene: read timeout: context deadline exceeded
exit status 1
```

**After fix:**
```
2025/11/09 16:18:44 3. Launching main scene...
2025/11/09 16:18:44 ✓ Launch request sent (state: initialized)
2025/11/09 16:18:44   Launch response: &{Response:{...RequestSeq:2 Success:true Command:launch...}}
```

Launch now responds immediately (< 1 second) instead of timing out after 30 seconds.

---

## DAP Protocol Reference

From the [Debug Adapter Protocol specification](https://microsoft.github.io/debug-adapter-protocol/specification):

> "The Debug Adapter Protocol defines the abstract protocol used between a development tool (for example, IDE or editor) and a debugger. Each **request** receives a **response** message with a unique sequence number."

The launch request is defined as:
- **Request:** `launch` with arguments
- **Response:** Success or error response
- **Effect:** Debugger should prepare to launch (but actual launch timing is implementation-specific)

Godot's implementation correctly defers the launch **effect**, but was incorrectly deferring the **response**.

---

## Related Code Locations

**Modified file:** `editor/debugger/debug_adapter/debug_adapter_parser.cpp`

**Line 184 change:**
```cpp
- return Dictionary();
+ return prepare_success_response(p_params);
```

**Related functions:**
- `req_launch()` - Stores launch parameters and NOW returns response
- `req_configurationDone()` - Triggers deferred launch from `pending_launch`
- `_launch_process()` - Actual launch execution

**Response handling:** `editor/debugger/debug_adapter/debug_adapter_protocol.cpp`
- Lines 867-878: Empty dictionary detection and special launch handling

---

## Impact

**Before:**
- ❌ All DAP clients timeout on launch (30s wait)
- ❌ Confusing error messages ("read timeout")
- ❌ Protocol violation

**After:**
- ✅ Launch responds immediately
- ✅ Protocol compliant
- ✅ Clear success/failure indication
- ✅ Correct deferred execution preserved

---

## Lessons Learned

1. **Protocol compliance is critical** - Every request MUST receive a response, even if the execution is deferred
2. **Empty Dictionary as signal** - Using empty return values as control flow signals can violate protocol expectations
3. **Separation of concerns** - Response acknowledgment vs. execution timing are separate concerns
4. **Testing revealed the issue** - Integration testing with an actual DAP client exposed the timeout
5. **Simple fix, big impact** - One line change fixes 30-second timeout

This fix ensures Godot's DAP implementation is fully protocol-compliant while maintaining the correct deferred launch behavior.
