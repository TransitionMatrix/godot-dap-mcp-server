# DAP Protocol: Required vs Optional Fields

**Sources:**
- [Official DAP Specification](https://microsoft.github.io/debug-adapter-protocol/specification)
- **Machine-readable spec**: `docs/reference/debugAdapterProtocol.json` (saved locally, 178KB)
- Direct link: https://microsoft.github.io/debug-adapter-protocol/debugAdapterProtocol.json

---

## Required vs Optional Fields

### ProtocolMessage (Base Class)

```typescript
interface ProtocolMessage {
  seq: number;      // ✅ REQUIRED
  type: string;     // ✅ REQUIRED ('request', 'response', 'event')
}
```

### Request (Base for All Requests)

```json
{
  "required": ["type", "command"]
}
```

**Fields:**
- **`type`** - ✅ REQUIRED - Always "request"
- **`command`** - ✅ REQUIRED - Command name
- **`seq`** - ✅ REQUIRED (inherited from ProtocolMessage)
- **`arguments`** - ❌ OPTIONAL in base Request

**Note:** Specific request types may override this and make `arguments` required.

### InitializeRequest (Specific Request Type)

```json
{
  "required": ["command", "arguments"]
}
```

**Fields:**
- **`command`** - ✅ REQUIRED - Must be "initialize"
- **`arguments`** - ✅ REQUIRED - Must be InitializeRequestArguments
- **`arguments.adapterID`** - ✅ REQUIRED - Adapter identifier
- **`arguments.linesStartAt1`** - ❌ OPTIONAL - Line numbering
- **`arguments.columnsStartAt1`** - ❌ OPTIONAL - Column numbering
- **`arguments.supportsVariableType`** - ❌ OPTIONAL - Variable type support
- *(and 10+ other optional capability fields)*

**Important:** Per the official spec, `InitializeRequest` **requires** the `arguments` field, unlike the base `Request` type.

---

## Godot DAP Code Analysis

### What Godot Can Safely Assume

**These fields are REQUIRED by DAP spec:**
- ✅ `seq` - Sequence number
- ✅ `type` - Message type ('request', 'response', 'event')
- ✅ `command` - Command name

**Godot code can use `operator[]` for these without safety checks.**

### What Godot Cannot Assume

**These fields are OPTIONAL per DAP spec:**
- ❌ `arguments` - Command arguments object
- ❌ All nested fields within `arguments`

**Godot MUST use `.get()` for these fields.**

---

## Current Godot Bugs

### Bug 1: Unsafe Access to Optional `arguments`

**Location:** Throughout `debug_adapter_parser.cpp`

```cpp
// ❌ WRONG - "arguments" is optional per DAP spec
Dictionary args = p_params["arguments"];
```

**Fix:**
```cpp
// ✅ CORRECT - Safe access with default
Dictionary args = p_params.get("arguments", Dictionary());
```

**Impact:** If client omits `arguments` field (which is valid per spec), Dictionary error occurs.

### Bug 2: Unsafe Access to Nested Optional Fields

**Location:** Multiple handlers in `debug_adapter_parser.cpp`

```cpp
// In req_launch()
Dictionary args = p_params["arguments"];  // May be OK if we know arguments exists
String project = args["project"];          // ❌ WRONG - "project" is optional!
String scene = args["scene"];              // ❌ WRONG - "scene" is optional!
```

**Fix:**
```cpp
Dictionary args = p_params.get("arguments", Dictionary());
String project = args.get("project", "");  // ✅ CORRECT
String scene = args.get("scene", "main");  // ✅ CORRECT
```

**Impact:** If client omits optional nested fields, Dictionary errors occur.

### Bug 3: Response Preparation (Debatable)

**Location:** `prepare_success_response()` and `prepare_error_response()`

```cpp
// Current code assumes seq and command are present
response["request_seq"] = p_params["seq"];     // Technically OK (seq is required)
response["command"] = p_params["command"];      // Technically OK (command is required)
```

**Analysis:**
- Per spec, these ARE required fields
- However, defensive programming suggests using `.get()` anyway for robustness
- Our PR includes these in the 32 fixes for consistency

**Recommendation:** Use `.get()` for defensive programming even though spec says they're required.

---

## Valid Test Cases

### Test 1: Initialize with Minimal Required Fields (VALID)

```json
{
  "seq": 1,
  "type": "request",
  "command": "initialize",
  "arguments": {
    "adapterID": "godot"
  }
}
```

**Expected:** Should work correctly (includes all required fields, omits optional capability fields)

**What this tests:** Safe access to optional fields within `arguments` (linesStartAt1, columnsStartAt1, etc.)

### Test 2: setBreakpoints Without breakpoints Array (VALID)

```json
{
  "seq": 2,
  "type": "request",
  "command": "setBreakpoints",
  "arguments": {
    "source": {"path": "/file.gd"}
  }
}
```

**Expected:** Should clear all breakpoints (per DAP spec)

**Godot behavior with unsafe code:** Dictionary error on `args["breakpoints"]`

### Test 3: Launch Without project Field (VALID)

```json
{
  "seq": 3,
  "type": "request",
  "command": "launch",
  "arguments": {
    "scene": "main"
  }
}
```

**Expected:** Should launch with default project path

**Godot behavior with unsafe code:** Dictionary error on `args["project"]` (if accessed)

---

## Invalid Test Cases (Don't Test These)

### Missing Required seq

```json
{
  "type": "request",
  "command": "initialize"
}
```

**Status:** INVALID per DAP spec
**Verdict:** Godot is not obligated to handle this

### Missing Required command

```json
{
  "seq": 1,
  "type": "request"
}
```

**Status:** INVALID per DAP spec
**Verdict:** Godot is not obligated to handle this

---

## Summary of Our PR Fixes

### Category 1: Fixing Required-But-Treated-As-Optional (6 fixes)

Locations like `prepare_success_response()`:
```cpp
// These are technically required by spec, but we use .get() for robustness
response["request_seq"] = p_params.get("seq", 0);
response["command"] = p_params.get("command", "");
```

**Rationale:** Defensive programming - handle malformed/buggy clients gracefully.

### Category 2: Fixing Actually Optional Fields (25 fixes)

Locations throughout request handlers:
```cpp
// These are ACTUALLY optional per spec - MUST use .get()
Dictionary args = p_params.get("arguments", Dictionary());
String project = args.get("project", "");
String scene = args.get("scene", "main");
Array breakpoints = args.get("breakpoints", Array());
```

**Rationale:** DAP spec compliance - these fields are optional.

### Category 3: Launch Response Fix (1 fix)

```cpp
// Before: req_launch() returned Dictionary() (no response)
return Dictionary();

// After: req_launch() returns proper response
return prepare_success_response(p_params);
```

**Rationale:** DAP spec requires every request gets a response.

---

## Revised Testing Strategy

Our test program now sends **VALID** DAP messages that should work according to spec:

1. ✅ Initialize with minimal required fields (includes `arguments` and `adapterID`, omits optional capability fields)
2. ✅ Launch with minimal arguments (omits optional `project`, `scene` fields)
3. ✅ setBreakpoints without `breakpoints` array (optional per spec - means "clear all")
4. ✅ Launch with missing nested optional fields (optional per spec)

These are all legitimate messages a DAP client might send, and Godot should handle them correctly.

---

## Conclusion

### What We Learned

1. **`seq`, `type`, `command` are REQUIRED** by DAP spec (for all requests)
2. **`arguments` requirements vary by request type:**
   - Base `Request`: `arguments` is optional
   - `InitializeRequest`: `arguments` is **required** (with required `adapterID`)
   - Other requests: varies (check spec for each)
3. **All nested fields in `arguments` vary by command** (most are optional)

### What Godot Should Do

1. ✅ Can assume `seq`, `type`, `command` are present (but `.get()` is safer for robustness)
2. For `arguments`:
   - For `initialize`: Can assume present (required by spec), but `.get()` is still safer
   - For other commands: Check spec - some require it, some don't
   - **Best practice**: Always use `.get()` for defensive programming
3. ❌ Cannot assume nested fields in `arguments` are present - MUST use `.get()` for all

### What Our PR Fixes

- All 31 unsafe Dictionary accesses replaced with safe `.get()` calls
- Handles both required-but-better-safe and actually-optional fields
- Makes Godot DAP implementation robust and spec-compliant
- Launch response fix (separate issue)

**Result:** Godot DAP implementation that works with all DAP clients, not just well-behaved ones.
