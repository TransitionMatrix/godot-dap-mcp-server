# Pull Request Template

Use this template when creating PRs in the Godot repository.

---

## Title Format
```
Fix Dictionary safety in req_<command>()
```

Example: `Fix Dictionary safety in req_initialize()`

---

## PR Body

```markdown
## Fix Dictionary safety in req_<command>()

Fixes godotengine/godot#<issue-number>

### Problem

The `req_<command>()` function uses unsafe Dictionary access (`operator[]`)
for optional fields defined in the Debug Adapter Protocol specification. When
a DAP client sends a spec-compliant request omitting optional fields, Godot
logs Dictionary errors to the console.

### Evidence

**Test tool**: https://github.com/<your-username>/godot-dap-mcp-server/tree/main/cmd/test-dap-protocol

Running the test tool against current Godot master shows:
```
ERROR: Bug: Dictionary::operator[] used when there was no value for the given key
   at: operator[] (core/variant/dictionary.cpp:136)
```

**DAP Specification**: Per [Microsoft's DAP spec](https://microsoft.github.io/debug-adapter-protocol/specification),
the following fields are optional in <Command>RequestArguments:
- `<field1>` (optional, defaults to <default1>)
- `<field2>` (optional, defaults to <default2>)
- ...

### Solution

Replace unsafe `operator[]` with safe `.get()` method using appropriate
default values per the DAP specification.

**Before**:
```cpp
Dictionary args = p_params["arguments"];
field1 = args["field1"];
field2 = args["field2"];
```

**After**:
```cpp
Dictionary args = p_params.get("arguments", Dictionary());
field1 = args.get("field1", default1);
field2 = args.get("field2", default2);
```

### Testing

**Before (upstream master)**:
```bash
# Start Godot with DAP enabled
./bin/godot.macos.editor.arm64

# Run test tool (sends spec-compliant messages)
go run cmd/test-dap-protocol/main.go

# Result: Dictionary errors in Godot console
```

**After (with this PR)**:
```bash
# Same test
go run cmd/test-dap-protocol/main.go

# Result: No errors, clean execution
```

### Impact

- **Severity**: Low (functionality works but logs errors)
- **Scope**: Affects all DAP clients that send minimal requests
- **Risk**: Very low - adds safe defaults, no behavior change for clients
  that send all fields

### Files Changed

- `editor/debugger/debug_adapter/debug_adapter_parser.cpp` - N lines
```

---

## Commit Message Format

```
Fix Dictionary safety in req_<command>()

The req_<command>() function used unsafe Dictionary access (operator[])
for optional DAP protocol fields. This caused Dictionary errors when
clients omitted optional fields.

Replace with safe .get() calls using appropriate defaults per the
Debug Adapter Protocol specification.

Fixes #<issue-number>
```

---

## Checklist Before Submitting

- [ ] Branch created from latest Godot master
- [ ] Code follows Godot coding style (checked with clang-format)
- [ ] Tested with test-dap-protocol showing no errors
- [ ] Commit message follows Godot conventions
- [ ] References GitHub issue number
- [ ] PR targets correct Godot branch (master)
- [ ] Ready for review
