# Comprehensive DAP Dictionary Access Audit

**Date:** November 9, 2025
**Purpose:** Identify and fix ALL unsafe Dictionary access patterns in Godot's DAP implementation

## Executive Summary

Total unsafe Dictionary accesses found: **31 locations**
- `debug_adapter_parser.cpp`: 25 locations
- `debug_adapter_protocol.cpp`: 6 locations
- `debug_adapter_types.h`: 2 locations (ALREADY FIXED)

## Classification

### Safe Patterns (No Fix Needed)
All `dict["key"] = value` in `to_json()` methods are SAFE because they:
- Write TO dictionaries we create
- Never read FROM external dictionaries
- Can't trigger the Dictionary const-correctness error

### Unsafe Patterns (MUST Fix)
Reading from dictionaries with `dict["key"]` where dict comes from:
- DAP protocol messages (`p_params["arguments"]`)
- Parsed arguments (`args["field"]`)
- Debug data structures (`stack_info["file"]`)

**Fix pattern:** Replace `dict["key"]` with `dict.get("key", default_value)`

---

## File: `editor/debugger/debug_adapter/debug_adapter_parser.cpp`

### Function: `req_initialize()` (Line 127)
**Status:** 🔴 UNSAFE

**Line 129:**
```cpp
Dictionary args = p_params["arguments"];  // 🔴 UNSAFE
```

**Fix:**
```cpp
Dictionary args = p_params.get("arguments", Dictionary());
```

**Reason:** DAP spec: `arguments` field is optional in requests

---

### Function: `req_launch()` (Line 169)
**Status:** 🟡 PARTIALLY FIXED

**Line 170:** ✅ ALREADY FIXED
```cpp
Dictionary args = p_params.get("arguments", Dictionary());  // ✅ FIXED
```

**Line 171:** 🔴 UNSAFE (inside has() check)
```cpp
if (args.has("project") && !is_valid_path(args["project"])) {  // 🔴 UNSAFE
```

**Line 173:** 🔴 UNSAFE
```cpp
    variables["clientPath"] = args["project"];  // 🔴 UNSAFE
```

**Fix:**
```cpp
if (args.has("project") && !is_valid_path(args.get("project", ""))) {
    Dictionary variables;
    variables["clientPath"] = args.get("project", "");
```

**Reason:** Even with has() check, using operator[] is unsafe pattern. Always use .get()

---

**Line 179:** 🔴 UNSAFE (inside has() check)
```cpp
if (args.has("godot/custom_data")) {
    DebugAdapterProtocol::get_singleton()->get_current_peer()->supportsCustomData = args["godot/custom_data"];  // 🔴 UNSAFE
}
```

**Fix:**
```cpp
if (args.has("godot/custom_data")) {
    DebugAdapterProtocol::get_singleton()->get_current_peer()->supportsCustomData = args.get("godot/custom_data", false);
}
```

---

### Function: `_build_play_arguments()` (Line 188)
**Status:** 🔴 UNSAFE

**Line 190:** 🔴 UNSAFE (inside has() check)
```cpp
if (p_args.has("playArgs")) {
    Variant v = p_args["playArgs"];  // 🔴 UNSAFE
```

**Fix:**
```cpp
if (p_args.has("playArgs")) {
    Variant v = p_args.get("playArgs", Array());
```

---

### Function: `req_restart()` (Line 271)
**Status:** 🔴 UNSAFE

**Line 273:** 🔴 UNSAFE
```cpp
args = params["arguments"];  // 🔴 UNSAFE
```

**Line 274:** 🔴 UNSAFE
```cpp
args = args["arguments"];  // 🔴 UNSAFE
```

**Line 275:** 🟡 ASSIGNMENT (writing, not reading - but follows unsafe read)
```cpp
params["arguments"] = args;
```

**Fix:**
```cpp
// Extract embedded "arguments" so it can be given to req_launch/req_attach
Dictionary params = p_params, args;
args = params.get("arguments", Dictionary());
args = args.get("arguments", Dictionary());
params["arguments"] = args;  // This assignment is OK after safe reads
```

**Reason:** Double-nested arguments structure in restart request

---

### Function: `req_setBreakpoints()` (Line 364)
**Status:** 🔴 UNSAFE

**Line 366:** 🔴 UNSAFE
```cpp
Dictionary args = p_params["arguments"];  // 🔴 UNSAFE
```

**Line 368:** 🔴 UNSAFE
```cpp
source.from_json(args["source"]);  // 🔴 UNSAFE
```

**Fix:**
```cpp
Dictionary args = p_params.get("arguments", Dictionary());
DAP::Source source;
source.from_json(args.get("source", Dictionary()));
```

**Reason:** `arguments` and `source` are required by DAP spec, but client may send incomplete data

---

**Line 374-375:** 🟢 SAFE (these are assignments to variables Dictionary)
```cpp
variables["clientPath"] = source.path;
variables["editorPath"] = ProjectSettings::get_singleton()->get_resource_path();
```

---

**Line 385:** 🔴 UNSAFE
```cpp
Array breakpoints = args["breakpoints"], lines;  // 🔴 UNSAFE
```

**Fix:**
```cpp
Array breakpoints = args.get("breakpoints", Array()), lines;
```

---

### Function: `req_breakpointLocations()` (Line 402)
**Status:** 🔴 UNSAFE

**Line 404:** 🔴 UNSAFE
```cpp
Dictionary args = p_params["arguments"];  // 🔴 UNSAFE
```

**Line 407:** 🔴 UNSAFE
```cpp
location.line = args["line"];  // 🔴 UNSAFE
```

**Line 409:** 🔴 UNSAFE (inside has() check)
```cpp
if (args.has("endLine")) {
    location.endLine = args["endLine"];  // 🔴 UNSAFE
}
```

**Fix:**
```cpp
Dictionary args = p_params.get("arguments", Dictionary());

DAP::BreakpointLocation location;
location.line = args.get("line", 0);
if (args.has("endLine")) {
    location.endLine = args.get("endLine", -1);
}
```

---

### Function: `req_scopes()` (Line 419)
**Status:** 🔴 UNSAFE

**Line 421:** 🔴 UNSAFE
```cpp
Dictionary args = p_params["arguments"];  // 🔴 UNSAFE
```

**Line 422:** 🔴 UNSAFE
```cpp
int frame_id = args["frameId"];  // 🔴 UNSAFE
```

**Fix:**
```cpp
Dictionary args = p_params.get("arguments", Dictionary());
int frame_id = args.get("frameId", 0);
```

---

### Function: `req_variables()` (Line 461)
**Status:** 🔴 UNSAFE

**Line 463:** 🔴 UNSAFE
```cpp
Dictionary args = p_params["arguments"];  // 🔴 UNSAFE
```

**Line 464:** 🔴 UNSAFE
```cpp
int variable_id = args["variablesReference"];  // 🔴 UNSAFE
```

**Fix:**
```cpp
Dictionary args = p_params.get("arguments", Dictionary());
int variable_id = args.get("variablesReference", 0);
```

---

### Function: `req_evaluate()` (Line 507)
**Status:** 🔴 UNSAFE

**Line 508:** 🔴 UNSAFE
```cpp
Dictionary args = p_params["arguments"];  // 🔴 UNSAFE
```

**Line 509:** 🔴 UNSAFE
```cpp
String expression = args["expression"];  // 🔴 UNSAFE
```

**Line 510:** 🟡 MIXED (has() check but still uses operator[])
```cpp
int frame_id = args.has("frameId") ? static_cast<int>(args["frameId"]) : DebugAdapterProtocol::get_singleton()->_current_frame;
```

**Fix:**
```cpp
Dictionary args = p_params.get("arguments", Dictionary());
String expression = args.get("expression", "");
int frame_id = args.get("frameId", DebugAdapterProtocol::get_singleton()->_current_frame);
```

---

### Function: `req_godot_put_msg()` (Line 531)
**Status:** 🔴 UNSAFE

**Line 532:** 🔴 UNSAFE
```cpp
Dictionary args = p_params["arguments"];  // 🔴 UNSAFE
```

**Line 534:** 🔴 UNSAFE
```cpp
String msg = args["message"];  // 🔴 UNSAFE
```

**Line 535:** 🔴 UNSAFE
```cpp
Array data = args["data"];  // 🔴 UNSAFE
```

**Fix:**
```cpp
Dictionary args = p_params.get("arguments", Dictionary());

String msg = args.get("message", "");
Array data = args.get("data", Array());
```

---

## File: `editor/debugger/debug_adapter/debug_adapter_protocol.cpp`

### Function: `process_message()` (Line 855)
**Status:** 🔴 UNSAFE

**Line 857:** 🔴 UNSAFE (both read and write)
```cpp
if (params.has("seq")) {
    params["seq"] = (int)params["seq"];  // 🔴 UNSAFE - reads then writes
}
```

**Fix:**
```cpp
if (params.has("seq")) {
    params["seq"] = (int)params.get("seq", 0);
}
```

**Reason:** Reading with operator[] to convert type, then writing back. Use .get() for read.

---

### Function: `on_debug_breaked()` (Line 1090)
**Status:** 🔴 UNSAFE

**Line 1091:** 🟢 SAFE (Array access)
```cpp
Dictionary d = p_stack_dump[0];  // Array access, not Dictionary
```

**Line 1092:** 🔴 UNSAFE
```cpp
DAP::Breakpoint breakpoint(fetch_source(d["file"]));  // 🔴 UNSAFE
```

**Line 1093:** 🔴 UNSAFE
```cpp
breakpoint.line = d["line"];  // 🔴 UNSAFE
```

**Fix:**
```cpp
Dictionary d = p_stack_dump[0];
DAP::Breakpoint breakpoint(fetch_source(d.get("file", "")));
breakpoint.line = d.get("line", 0);
```

**Reason:** `p_stack_dump` comes from Godot engine debug data, should be trusted, but still follow safe pattern

---

### Function: `on_debug_stack_dump()` (Line 1109)
**Status:** 🔴 UNSAFE

**Line 1109:** 🟢 SAFE (Array access)
```cpp
Dictionary stack_info = p_stack_dump[i];  // Array access, not Dictionary
```

**Line 1111:** 🔴 UNSAFE
```cpp
DAP::StackFrame stackframe(fetch_source(stack_info["file"]));  // 🔴 UNSAFE
```

**Line 1113:** 🔴 UNSAFE
```cpp
stackframe.name = stack_info["function"];  // 🔴 UNSAFE
```

**Line 1114:** 🔴 UNSAFE
```cpp
stackframe.line = stack_info["line"];  // 🔴 UNSAFE
```

**Fix:**
```cpp
Dictionary stack_info = p_stack_dump[i];

DAP::StackFrame stackframe(fetch_source(stack_info.get("file", "")));
stackframe.id = stackframe_id++;
stackframe.name = stack_info.get("function", "");
stackframe.line = stack_info.get("line", 0);
stackframe.column = 0;
```

**Reason:** Same as above - engine debug data should be trusted, but follow safe pattern

---

## File: `editor/debugger/debug_adapter/debug_adapter_types.h`

### Status: ✅ ALREADY FIXED

**Lines 87-89:** ✅ FIXED in Compilation 3
```cpp
_FORCE_INLINE_ void from_json(const Dictionary &p_params) {
    name = p_params.get("name", "");
    path = p_params.get("path", "");
    _checksums = p_params.get("checksums", Array());
}
```

**Line 215:** ✅ FIXED in Compilation 3
```cpp
_FORCE_INLINE_ void from_json(const Dictionary &p_params) {
    line = p_params.get("line", 0);
}
```

---

## Summary by Priority

### Critical (External DAP Protocol Input)
These MUST be fixed - they parse external DAP client requests:

1. `req_initialize()` - Line 129
2. `req_launch()` - Lines 171, 173, 179
3. `req_restart()` - Lines 273, 274
4. `req_setBreakpoints()` - Lines 366, 368, 385
5. `req_breakpointLocations()` - Lines 404, 407, 409
6. `req_scopes()` - Lines 421, 422
7. `req_variables()` - Lines 463, 464
8. `req_evaluate()` - Lines 508, 509, 510
9. `req_godot_put_msg()` - Lines 532, 534, 535
10. `_build_play_arguments()` - Line 190
11. `process_message()` - Line 857

**Total: 25 locations in debug_adapter_parser.cpp, 1 in debug_adapter_protocol.cpp**

### Important (Internal Engine Data)
These should be fixed for consistency, but less likely to crash (Godot engine controls this data):

1. `on_debug_breaked()` - Lines 1092, 1093
2. `on_debug_stack_dump()` - Lines 1111, 1113, 1114

**Total: 5 locations in debug_adapter_protocol.cpp**

---

## Fix Strategy

### Batch 1: Request Handlers (High Priority)
Fix all `req_*` functions that parse DAP client requests:
- req_initialize
- req_launch
- req_restart
- req_setBreakpoints
- req_breakpointLocations
- req_scopes
- req_variables
- req_evaluate
- req_godot_put_msg

**Estimated compilation time:** 20-30 seconds

### Batch 2: Internal Handlers (Medium Priority)
Fix internal debug event handlers:
- on_debug_breaked
- on_debug_stack_dump
- process_message seq conversion

**Estimated compilation time:** 20-30 seconds

### Batch 3: Test & Verify
- Compile final binary
- Test launch sequence
- Test all DAP commands

**Total estimated time:** ~2-3 minutes compilation + testing

---

## Testing Plan

After all fixes:

1. **Launch Test:**
   ```bash
   cd /Users/adp/Projects/godot-dap-mcp-server
   go run cmd/launch-test/main.go
   ```
   Expected: ✅ Launch succeeds without timeout

2. **SetBreakpoints Test:**
   Expected: ✅ No Dictionary errors (already verified)

3. **Full Integration Test:**
   ```bash
   ./scripts/automated-integration-test.sh
   ```
   Expected: ✅ All Phase 1-4 tools work

---

## Expected Outcome

After fixing all 31 unsafe Dictionary accesses:
- ✅ No more Dictionary const-correctness errors
- ✅ Launch request completes successfully
- ✅ All DAP commands handle missing/malformed fields gracefully
- ✅ Godot DAP implementation follows safe Dictionary access pattern throughout

This comprehensive fix aligns with the regression analysis conclusion: **The DAP code has been unsafe since 2021, and now we're fixing it properly.**
