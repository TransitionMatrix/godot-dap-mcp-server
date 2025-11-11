# Testing Unsafe Dictionary Code in Godot DAP

This guide explains how to send minimal DAP requests to expose the unsafe Dictionary access patterns in godot-upstream.

---

## Quick Start

### Interactive Go Test Program (Recommended)

```bash
# Run directly (builds on first run)
cd /Users/adp/Projects/godot-dap-mcp-server
go run cmd/test-dap-protocol/main.go

# Or build once and run multiple times
go build -o test-dap-protocol cmd/test-dap-protocol/main.go
./test-dap-protocol
```

**Features:**
- Interactive prompts between tests
- Colored output showing what's sent vs received
- **Clearly shows DAP spec requirements vs Godot's expectations**
- Proper multi-message DAP parsing
- Pretty-printed JSON
- **Spec-compliant messages** (based on official DAP specification)

**DAP Specification Reference:**
The official Debug Adapter Protocol specification is saved in:
- `docs/reference/debugAdapterProtocol.json` (machine-readable, 178KB)
- Source: https://microsoft.github.io/debug-adapter-protocol/debugAdapterProtocol.json

**Prerequisites:**
1. Godot editor running with DAP enabled
2. Go 1.16+ installed

---

## What To Look For

### In Terminal Output

The test will send 4 minimal requests:

1. **Initialize without 'seq'** → Should trigger `p_params["seq"]` error
2. **Initialize without 'arguments'** → Should trigger `p_params["arguments"]` error
3. **Launch without 'arguments'** → Should trigger `p_params["arguments"]` error
4. **setBreakpoints without 'breakpoints'** → Should trigger `args["breakpoints"]` error

### In Godot Console

**If unsafe code is present**, you should see errors like:

```
ERROR: Bug: Dictionary::operator[] used when there was no value for the given key, please report.
   at: operator[] (core/variant/dictionary.cpp:136)
ERROR: Condition "!value" is true. Returning: empty
   at: prepare_success_response (editor/debugger/debug_adapter/debug_adapter_parser.cpp:74)
```

**If code is safe** (with our fixes), you should see:
- Clean execution
- Proper error responses from DAP server
- No Dictionary errors in console

---

## Manual Testing with netcat

For interactive testing:

```bash
# Connect to DAP server
nc localhost 6006

# Then type (calculate Content-Length manually):
Content-Length: 54

{"seq":1,"type":"request","command":"initialize"}
```

**Content-Length Calculation:**
```bash
# Use this to calculate:
echo -n '{"seq":1,"type":"request","command":"initialize"}' | wc -c
# Output: 54
```

---

## Test Cases Explained

### Test 1: Initialize with Minimal Required Fields

**Request:**
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

**Validity:** ✅ VALID - Includes all required fields per DAP spec
- Required: `seq`, `type`, `command`, `arguments`, `arguments.adapterID`
- Omitted optional: `linesStartAt1`, `columnsStartAt1`, `supportsVariableType`, etc.

**What this tests:**
```cpp
// In req_initialize() - upstream vs fixed
Dictionary args = p_params.get("arguments", Dictionary());  // ✅ Safe
peer->linesStartAt1 = args.get("linesStartAt1", false);     // ✅ Safe
```

**Result:** Should work cleanly with both upstream and fixed versions (no Dictionary errors expected)

---

### Test 2: Missing Nested 'breakpoints' Field (VALID per DAP spec)

**Request:**
```json
{
  "seq":3,
  "type":"request",
  "command":"setBreakpoints",
  "arguments": {
    "source": {"path": "/test.gd"}
  }
}
```

**Validity:** ✅ VALID - Omitting "breakpoints" means "clear all breakpoints" per DAP spec

**What breaks:**
```cpp
// In req_setBreakpoints()
Array breakpoints = args["breakpoints"];  // ← breakpoints missing, Dictionary error!
```

**Expected error:**
```
ERROR: Bug: Dictionary::operator[] used when there was no value for the given key
   at: operator[] (core/variant/dictionary.cpp:136)
```

---

## Why Our Regular DAP Client Doesn't Trigger These

Our go-dap based client always sends all fields:

```go
// From internal/dap/client.go
request := &dap.InitializeRequest{
    Request: dap.Request{
        ProtocolMessage: dap.ProtocolMessage{
            Seq:  c.nextRequestSeq(),     // ← Always present
            Type: "request",
        },
        Command: "initialize",             // ← Always present
    },
    Arguments: dap.InitializeRequestArguments{ // ← Always present
        ClientID: "godot-dap-mcp-server",
        // ...
    },
}
```

Because all expected fields are present, the Dictionary lookups succeed and no errors occur.

---

## Comparison: Upstream vs With Fixes

### Upstream (Unsafe Code)

```cpp
// prepare_success_response() - UNSAFE
response["request_seq"] = p_params["seq"];
response["command"] = p_params["command"];

// req_initialize() - UNSAFE
Dictionary args = p_params["arguments"];
peer->linesStartAt1 = args["linesStartAt1"];
```

**Result with minimal requests:** Dictionary errors in console

### With Our Fixes (Safe Code)

```cpp
// prepare_success_response() - SAFE
response["request_seq"] = p_params.get("seq", 0);
response["command"] = p_params.get("command", "");

// req_initialize() - SAFE
Dictionary args = p_params.get("arguments", Dictionary());
peer->linesStartAt1 = args.get("linesStartAt1", false);
```

**Result with minimal requests:** Clean execution, proper error handling

---

## Expected Results

### Testing godot-upstream (2d86b69bf1)

- ❌ Dictionary errors in console for missing fields
- ⚠️ May still return responses (ERR_FAIL_COND_V_MSG returns a value)
- ❌ Console filled with error messages

### Testing with Our PR (fix/dap-dictionary-safety)

- ✅ No Dictionary errors
- ✅ Clean console output
- ✅ Proper handling of missing optional fields
- ✅ Default values used appropriately

---

## Troubleshooting

### Connection Refused

```
Failed to connect to localhost:6006
```

**Solution:**
1. Start Godot editor
2. Enable DAP: Editor → Editor Settings → Network → Debug Adapter → Enable
3. Check port: Default is 6006

### No Response

If test hangs waiting for response:
- Godot might have crashed/frozen due to error
- Check Godot console for errors
- Restart Godot and try again

### "Works Fine" With Upstream

If you don't see errors with upstream:
- You might be testing with a well-behaved client that sends all fields
- Try the test programs above that send minimal requests
- Check Godot console carefully for error messages

---

## Additional Manual Tests

### Using Python

```python
#!/usr/bin/env python3
import socket
import json

def send_dap(msg):
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.connect(('localhost', 6006))

    data = json.dumps(msg)
    header = f"Content-Length: {len(data)}\r\n\r\n"
    s.send(header.encode() + data.encode())

    s.close()

# Test: Initialize without arguments (VALID - arguments is optional)
send_dap({"seq": 1, "type": "request", "command": "initialize"})

# Test: Launch without arguments (VALID - arguments is optional)
send_dap({"seq": 2, "type": "request", "command": "launch"})

# Test: setBreakpoints without breakpoints array (VALID - means clear all)
send_dap({
    "seq": 3,
    "type": "request",
    "command": "setBreakpoints",
    "arguments": {"source": {"path": "/test.gd"}}
})
```

### Using curl (won't work - DAP isn't HTTP)

DAP uses raw TCP sockets, not HTTP, so curl won't work directly.

---

## Conclusion

These tests demonstrate that:

1. ✅ Upstream code has unsafe Dictionary access patterns
2. ✅ Well-behaved clients (like ours) don't trigger them
3. ✅ Minimal/strict clients expose the bugs
4. ✅ Our fixes make the code robust for all clients

Use these tests to validate that our PR is necessary even though normal testing doesn't reveal the issues.
