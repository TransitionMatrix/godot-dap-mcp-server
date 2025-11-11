# GitHub Issue Template

Use this template when creating issues in the Godot repository.

---

## Title Format
```
Dictionary error in req_<command>() when optional field omitted
```

Example: `Dictionary error in req_initialize() when optional field omitted`

---

## Issue Body

```markdown
## Dictionary Error in req_<command>() with Spec-Compliant Request

**Description**
The `req_<command>()` function in `debug_adapter_parser.cpp` uses unsafe
Dictionary access (`operator[]`) for optional fields. When a DAP client sends
a valid request that omits optional fields, Godot logs Dictionary errors.

**Steps to Reproduce**
1. Enable DAP server in Godot: Editor → Editor Settings → Network → Debug Adapter
2. Send this spec-compliant <command> request:
   ```json
   {
     "seq": 1,
     "type": "request",
     "command": "<command>",
     "arguments": {
       "requiredField": "value"
     }
   }
   ```
3. Observe console output

**Expected Behavior**
No errors. Per the DAP specification, fields like `<optionalField1>`, `<optionalField2>`,
etc. are optional. The server should use default values when they're omitted.

**Actual Behavior**
```
ERROR: Bug: Dictionary::operator[] used when there was no value for the given key
   at: operator[] (core/variant/dictionary.cpp:136)
```

**DAP Specification**
Per the [Debug Adapter Protocol specification](https://microsoft.github.io/debug-adapter-protocol/specification):
- `arguments.<optionalField1>` is optional (defaults to <default1>)
- `arguments.<optionalField2>` is optional (defaults to <default2>)
- ...

**Root Cause**
File: `editor/debugger/debug_adapter/debug_adapter_parser.cpp`, lines X-Y

```cpp
// Current unsafe code
Dictionary args = p_params["arguments"];  // Unsafe!
field_value = args["optionalField"];       // Unsafe!
```

**Proposed Fix**
Use Dictionary's safe `.get()` method with appropriate defaults:

```cpp
Dictionary args = p_params.get("arguments", Dictionary());
field_value = args.get("optionalField", default_value);
```

**Impact**
- Severity: Low (functionality works, but logs errors)
- Frequency: Every time a strict DAP client connects
- User Experience: Console spam, confusion about errors

**Test Tool**
A test program demonstrating this issue is available at:
https://github.com/<your-username>/godot-dap-mcp-server/tree/main/cmd/test-dap-protocol

**Godot Version**
- Tested on: 4.4-dev (commit <hash>)
- Affects: All versions since DAP implementation
```

---

## Labels to Add

- `bug`
- `debugger`
- `priority:low` or `priority:medium` (depending on command importance)

---

## Checklist Before Submitting

- [ ] Tested against godot-upstream with test-dap-protocol
- [ ] Verified error appears in Godot console
- [ ] Checked DAP spec for field requirements
- [ ] Identified specific code location and line numbers
- [ ] Prepared fix with `.get()` calls
- [ ] Ready to submit PR after issue is created
