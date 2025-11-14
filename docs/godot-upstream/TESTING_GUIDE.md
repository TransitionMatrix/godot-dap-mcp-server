# Testing Unsafe Dictionary Code in Godot DAP

This guide explains how to send minimal DAP requests to expose the unsafe Dictionary access patterns in godot-upstream.

---

## Quick Start

### Automated Testing (Recommended for PR Strategy)

The fastest way to test Dictionary safety issues against any Godot build:

```bash
# Set the Godot binary to test
export GODOT_BIN=/Users/adp/Projects/godot-upstream/bin/godot.macos.editor.arm64

# Run automated test
./scripts/test-dap-compliance.sh
```

**What it does:**
1. Starts Godot in the background with DAP enabled
2. Waits for DAP server to be ready
3. Runs all compliance tests automatically (no manual input)
4. Captures output to timestamped files
5. Analyzes Godot console for Dictionary errors
6. Reports pass/fail with error counts
7. Cleans up automatically

**Output files:**
- Test output: `/tmp/godot-dap-test-output-YYYYMMDD-HHMMSS.txt`
- Godot console: `/tmp/godot-console-<PID>.log`

**Exit codes:**
- `0` - All tests passed, no Dictionary errors
- `1` - Dictionary errors detected

**Custom options:**
```bash
# Use different port
DAP_PORT=6007 ./scripts/test-dap-compliance.sh

# Use different project
PROJECT_PATH=/path/to/project.godot ./scripts/test-dap-compliance.sh

# Custom output location
OUTPUT_FILE=/custom/path/output.txt ./scripts/test-dap-compliance.sh
```

**Prerequisites:**
- `GODOT_BIN` environment variable set to Godot binary path
- `nc` (netcat) command available (standard on macOS/Linux)
- Go 1.16+ installed

See [Automated Testing Workflow](#automated-testing-workflow) below for detailed usage in PR strategy.

---

### Interactive Go Test Program

For manual exploration and understanding:

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

**When to use:**
- Learning about DAP protocol requirements
- Understanding specific test cases
- Observing Godot's responses in detail
- Debugging specific commands

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

## Automated Testing Workflow

### Phase 1: Document Dictionary Errors (PR Strategy)

Use automated testing to systematically document all Dictionary errors in upstream Godot:

**Step 1: Test upstream Godot**
```bash
# Set upstream binary
export GODOT_BIN=/Users/adp/Projects/godot-upstream/bin/godot.macos.editor.arm64

# Run automated test
./scripts/test-dap-compliance.sh

# Output saved to /tmp/godot-dap-test-output-YYYYMMDD-HHMMSS.txt
```

**Step 2: Analyze results**
```bash
# Check Dictionary error count
# Script reports: "Dictionary Errors: N"

# Review Godot console for specific errors
cat /tmp/godot-console-<PID>.log | grep -A 2 "Dictionary::operator\[\]"
```

**Step 3: Document findings**

For each DAP command that triggers errors, create an entry in your tracking document:

```markdown
## req_initialize()
- **File**: `editor/debugger/debug_adapter/debug_adapter_parser.cpp`
- **Line**: ~42
- **Unsafe access**: `p_params["arguments"]`
- **Test**: Initialize with minimal fields
- **Error**: Dictionary::operator[] at core/variant/dictionary.cpp:136
- **Fix**: `p_params.get("arguments", Dictionary())`
```

**Step 4: Repeat for development branch**
```bash
# Test your fixes
export GODOT_BIN=/Users/adp/Projects/godot/bin/godot.macos.editor.arm64
./scripts/test-dap-compliance.sh

# Should report: "Dictionary Errors: 0"
```

### Continuous Testing

Use automated testing during development:

```bash
# After each fix, verify it works
export GODOT_BIN=/path/to/your/godot/build
./scripts/test-dap-compliance.sh

# Script returns exit code 0 on success, 1 on errors
# Perfect for CI/CD integration
```

### Comparing Builds

Test multiple Godot builds to compare behavior:

```bash
# Test upstream (expect errors)
GODOT_BIN=/path/to/upstream/godot OUTPUT_FILE=/tmp/upstream-results.txt \
  ./scripts/test-dap-compliance.sh

# Test your fixes (expect no errors)
GODOT_BIN=/path/to/fixed/godot OUTPUT_FILE=/tmp/fixed-results.txt \
  ./scripts/test-dap-compliance.sh

# Compare results
diff /tmp/upstream-results.txt /tmp/fixed-results.txt
```

---

## Conclusion

These tests demonstrate that:

1. ✅ Upstream code has unsafe Dictionary access patterns
2. ✅ Well-behaved clients (like ours) don't trigger them
3. ✅ Minimal/strict clients expose the bugs
4. ✅ Our fixes make the code robust for all clients

**Testing strategies:**
- **Automated**: Fast, repeatable, perfect for PR strategy Phase 1
- **Interactive**: Educational, detailed observation, good for understanding
- **Manual**: Flexible, ad-hoc testing, custom scenarios

Use these tests to validate that our PR is necessary even though normal testing doesn't reveal the issues.
