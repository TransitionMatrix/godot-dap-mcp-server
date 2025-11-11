# Godot DAP MCP Server - Testing Guide

**Last Updated**: 2025-11-07

This document describes the testing strategy, test types, and testing procedures for the godot-dap-mcp-server.

---

## Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Unit Tests](#unit-tests)
3. [Integration Tests](#integration-tests)
4. [Manual Testing](#manual-testing)
5. [Test Project Setup](#test-project-setup)
6. [Running Tests](#running-tests)
7. [Continuous Integration](#continuous-integration)

---

## Testing Philosophy

### Test Coverage Goals

- **Unit Tests**: Cover critical paths in MCP protocol, DAP client, timeout mechanisms
- **Integration Tests**: Verify full MCP → DAP → Godot flow
- **Manual Tests**: Validate end-to-end workflows and edge cases

### What to Test

✅ **Test**:
- MCP protocol parsing and routing
- Tool parameter validation
- DAP session state transitions
- Timeout mechanisms
- Event filtering logic
- Error message formatting
- Godot launch configurations
- **DAP protocol compliance** - Verify messages match `docs/reference/debugAdapterProtocol.json`

❌ **Don't Test**:
- go-dap library internals (trust the library)
- Godot's DAP server behavior (trust Godot)
- Network layer (trust standard library)

### DAP Protocol Compliance Testing

When testing DAP commands, always verify against the official specification:

**Verification checklist**:
1. ✅ Include all required fields per `docs/reference/debugAdapterProtocol.json`
2. ✅ Test with minimal required fields (omit optional fields)
3. ✅ Use safe `.get()` access for optional fields in implementation
4. ✅ Test messages should be spec-compliant (use `test-dap-protocol` to verify)

**Example workflow**:
```bash
# 1. Check spec for required fields
jq '.definitions.InitializeRequest' docs/reference/debugAdapterProtocol.json
# Shows: "required": ["command", "arguments"]

# 2. Test with minimal required fields
# In test: send initialize with only adapterID (required), omit optional fields

# 3. Verify with test program
go run cmd/test-dap-protocol/main.go
```

**Testing tool**: `cmd/test-dap-protocol/` - Interactive tool that sends spec-compliant minimal messages to verify Godot's handling of optional fields.

---

## Unit Tests

Location: `internal/*/` (alongside source files)

### MCP Layer Tests

**What to Test**:
- Request parsing and validation
- Tool registration and lookup
- Response formatting
- Error handling

**Example**: `internal/mcp/server_test.go`

```go
package mcp

import "testing"

func TestServer_CallTool(t *testing.T) {
    server := NewServer()

    // Register test tool
    called := false
    server.RegisterTool(Tool{
        Name: "test_tool",
        Handler: func(params map[string]interface{}) (interface{}, error) {
            called = true
            return "success", nil
        },
    })

    // Call tool
    req := MCPRequest{
        Method: "tools/call",
        Params: map[string]interface{}{
            "name": "test_tool",
            "arguments": map[string]interface{}{},
        },
    }

    resp := server.handleRequest(req)

    if !called {
        t.Error("tool not called")
    }
    if resp.Error != nil {
        t.Errorf("unexpected error: %v", resp.Error)
    }
}

func TestServer_ToolNotFound(t *testing.T) {
    server := NewServer()

    req := MCPRequest{
        Method: "tools/call",
        Params: map[string]interface{}{
            "name": "nonexistent_tool",
            "arguments": map[string]interface{}{},
        },
    }

    resp := server.handleRequest(req)

    if resp.Error == nil {
        t.Error("expected error for nonexistent tool")
    }
    if resp.Error.Message != "tool not found: nonexistent_tool" {
        t.Errorf("wrong error message: %s", resp.Error.Message)
    }
}
```

### DAP Layer Tests

**What to Test**:
- Client initialization
- Request sequence generation
- Session state transitions
- Timeout helper functions
- Godot launch config validation

**Example**: `internal/dap/dap_test.go`

```go
package dap

import (
    "context"
    "testing"
    "time"
)

func TestNewClient(t *testing.T) {
    client := NewClient("localhost", 6006)

    if client.host != "localhost" {
        t.Errorf("expected host localhost, got %s", client.host)
    }
    if client.port != 6006 {
        t.Errorf("expected port 6006, got %d", client.port)
    }
    if client.nextSeq != 1 {
        t.Errorf("expected nextSeq 1, got %d", client.nextSeq)
    }
}

func TestRequestSeqGeneration(t *testing.T) {
    client := NewClient("localhost", 6006)

    seq1 := client.nextRequestSeq()
    seq2 := client.nextRequestSeq()
    seq3 := client.nextRequestSeq()

    if seq1 != 1 || seq2 != 2 || seq3 != 3 {
        t.Errorf("sequence not incrementing correctly: %d, %d, %d", seq1, seq2, seq3)
    }
}

func TestSessionStateTransitions(t *testing.T) {
    client := NewClient("localhost", 6006)
    session := NewSession(client)

    if session.state != StateDisconnected {
        t.Errorf("initial state should be StateDisconnected, got %v", session.state)
    }

    // Test invalid state transitions
    err := session.Initialize(context.Background())
    if err == nil {
        t.Error("Initialize should fail when not connected")
    }
}

func TestGodotLaunchConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  GodotLaunchConfig
        wantErr bool
    }{
        {
            name: "valid main scene config",
            config: GodotLaunchConfig{
                Project: "/path/to/project",
                Scene:   SceneLaunchMain,
            },
            wantErr: false,
        },
        {
            name: "missing project path",
            config: GodotLaunchConfig{
                Scene: SceneLaunchMain,
            },
            wantErr: true,
        },
        {
            name: "custom scene without path",
            config: GodotLaunchConfig{
                Project: "/path/to/project",
                Scene:   SceneLaunchCustom,
            },
            wantErr: true,
        },
        {
            name: "valid custom scene config",
            config: GodotLaunchConfig{
                Project:   "/path/to/project",
                Scene:     SceneLaunchCustom,
                ScenePath: "res://test.tscn",
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestTimeoutContextHelpers(t *testing.T) {
    ctx, cancel := WithConnectTimeout(nil)
    defer cancel()

    deadline, ok := ctx.Deadline()
    if !ok {
        t.Error("context should have deadline")
    }

    expectedDuration := DefaultConnectTimeout
    actualDuration := time.Until(deadline)

    // Allow 1 second tolerance
    diff := actualDuration - expectedDuration
    if diff < -time.Second || diff > time.Second {
        t.Errorf("timeout duration off: expected ~%v, got %v", expectedDuration, actualDuration)
    }
}
```

### Tool Layer Tests

**What to Test**:
- Parameter extraction and validation
- Error message formatting
- Tool registration

**Example**: `internal/tools/tools_test.go`

```go
package tools

import (
    "testing"

    "your-repo/internal/mcp"
)

func TestToolParameterValidation(t *testing.T) {
    // Test that tools properly validate required parameters
    tool := createTestTool()

    result, err := tool.Handler(map[string]interface{}{})

    if err == nil {
        t.Error("expected error for missing required parameter")
    }
}

func TestErrorMessageFormat(t *testing.T) {
    // Verify error messages follow Problem + Context + Solution pattern
    err := formatConnectionError("localhost", 6006)

    errStr := err.Error()

    // Should contain problem statement
    if !strings.Contains(errStr, "Failed to connect") {
        t.Error("error missing problem statement")
    }

    // Should contain context
    if !strings.Contains(errStr, "localhost:6006") {
        t.Error("error missing context")
    }

    // Should contain solutions
    if !strings.Contains(errStr, "Solutions:") {
        t.Error("error missing solutions")
    }
}
```

---

## Integration Tests

**Status**: ✅ Fully implemented and working (Phase 3)

Location: `scripts/` and `tests/`

**Purpose**: Test full MCP → DAP → Godot flow with real Godot editor.

### Automated Integration Test

**Script**: `scripts/automated-integration-test.sh`

**Features**:
- ✅ Fully automated - no manual setup required
- ✅ Launches Godot as subprocess in headless mode
- ✅ Auto-detects DAP port availability (6006-6020)
- ✅ Persistent MCP session via named pipes
- ✅ Tests all 7 Phase 3 tools
- ✅ Cleans up processes on exit

**Usage**:
```bash
# Requires Godot 4.2.2+ in PATH (or set GODOT_BIN)
./scripts/automated-integration-test.sh
```

**Tests Performed**:
1. MCP server initialization
2. Tool registration (7 tools)
3. Connect to Godot DAP server
4. Set breakpoint (with verification)
5. Clear breakpoint
6. Disconnect from Godot

### Manual Integration Test

**Script**: `scripts/integration-test.sh`

**Use Case**: Testing with a running Godot editor (user-controlled)

**Setup Requirements**:
1. Godot editor running with DAP enabled
2. Test project loaded (from `tests/fixtures/test-project/`)
3. DAP server listening on port 6006

**Usage**:
```bash
# 1. Open Godot editor
# 2. Enable DAP in Editor Settings
# 3. Run script
./scripts/integration-test.sh
```

### Test Project

**Location**: `tests/fixtures/test-project/`

**Contents**:
- `project.godot` - Minimal Godot 4.3 project
- `test_script.gd` - GDScript with breakpoint locations (lines 13, 18)
- `test_scene.tscn` - Scene that runs the test script
- `README.md` - Test project documentation

See `tests/INTEGRATION_TEST.md` for complete integration testing guide.

### Test Scenarios

```go
package integration

import (
    "context"
    "testing"
    "time"

    "your-repo/internal/dap"
)

func TestGodotDebugging_SetBreakpointAndHit(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    // Connect to Godot (running in CI or locally)
    client := dap.NewClient("localhost", 6006)

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := client.Connect(ctx); err != nil {
        t.Fatalf("failed to connect: %v", err)
    }
    defer client.Disconnect()

    // Initialize session
    if _, err := client.Initialize(ctx); err != nil {
        t.Fatalf("failed to initialize: %v", err)
    }

    if err := client.ConfigurationDone(ctx); err != nil {
        t.Fatalf("failed to send configurationDone: %v", err)
    }

    // Set breakpoint
    testFile := "/absolute/path/to/test/player.gd"
    if err := client.SetBreakpoints(ctx, testFile, []int{10}); err != nil {
        t.Fatalf("failed to set breakpoint: %v", err)
    }

    // Launch scene
    config := dap.LaunchMainScene("/path/to/test/project")
    session := dap.NewSession(client)

    if _, err := session.Launch(ctx, config.ToLaunchArgs()); err != nil {
        t.Fatalf("failed to launch: %v", err)
    }

    // Wait for breakpoint hit
    ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel2()

    stopped, err := client.WaitForStop(ctx2)
    if err != nil {
        t.Fatalf("breakpoint not hit: %v", err)
    }

    if stopped.Line != 10 {
        t.Errorf("wrong line: got %d, want 10", stopped.Line)
    }
}

func TestGodotDebugging_VariableInspection(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    // ... (setup same as above through breakpoint hit)

    // Get stack trace
    threads, err := client.Threads(ctx)
    if err != nil {
        t.Fatalf("failed to get threads: %v", err)
    }

    if len(threads.Threads) == 0 {
        t.Fatal("no threads returned")
    }

    // Get stack frames
    frames, err := client.StackTrace(ctx, threads.Threads[0].Id, 0, 10)
    if err != nil {
        t.Fatalf("failed to get stack trace: %v", err)
    }

    if len(frames.StackFrames) == 0 {
        t.Fatal("no stack frames returned")
    }

    // Get scopes
    scopes, err := client.Scopes(ctx, frames.StackFrames[0].Id)
    if err != nil {
        t.Fatalf("failed to get scopes: %v", err)
    }

    // Godot always returns 3 scopes: Locals, Members, Globals
    if len(scopes.Scopes) != 3 {
        t.Errorf("expected 3 scopes, got %d", len(scopes.Scopes))
    }

    // Get variables from Locals scope
    vars, err := client.Variables(ctx, scopes.Scopes[0].VariablesReference)
    if err != nil {
        t.Fatalf("failed to get variables: %v", err)
    }

    if len(vars.Variables) == 0 {
        t.Log("no local variables (might be empty scope)")
    }
}

func TestGodotDebugging_SteppingCommands(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    // ... (setup through breakpoint hit)

    // Test step over
    if err := client.Next(ctx, threadID); err != nil {
        t.Fatalf("failed to step over: %v", err)
    }

    // Wait for stopped event
    stopped, err := client.WaitForStop(ctx)
    if err != nil {
        t.Fatalf("step over did not stop: %v", err)
    }

    // Test step in
    if err := client.StepIn(ctx, threadID); err != nil {
        t.Fatalf("failed to step in: %v", err)
    }

    stopped, err = client.WaitForStop(ctx)
    if err != nil {
        t.Fatalf("step in did not stop: %v", err)
    }

    // Note: stepOut not tested (not implemented in Godot)
}
```

---

## Manual Testing

### Test Project Setup

Create minimal test project in `tests/fixtures/test-project/`:

```
test-project/
├── project.godot
├── main.tscn
└── scripts/
    └── player.gd
```

**player.gd** (simple testable script):

```gdscript
extends Node2D

var health = 100
var position_x = 0

func _ready():
    print("Player ready")  # Line 7 - good breakpoint spot
    initialize_stats()

func initialize_stats():
    health = 100  # Line 11
    position_x = 50

func take_damage(amount):
    health -= amount  # Line 15 - test variable inspection
    if health <= 0:
        die()

func die():
    print("Player died")  # Line 20
    queue_free()

func _process(delta):
    position_x += 10 * delta  # Line 24 - test stepping
    if position_x > 100:
        position_x = 0
```

### Manual Test Scenarios

1. **Connection Test**
   - Start Godot with test project
   - Run: `godot_connect(port=6006)`
   - Verify: Connection successful, capabilities returned

2. **Breakpoint Test**
   - Set breakpoint: `godot_set_breakpoint(file="/path/to/player.gd", line=7)`
   - Launch: `godot_launch_main_scene(project_path="/path/to/test-project")`
   - Verify: Game pauses at line 7

3. **Variable Inspection Test**
   - Hit breakpoint at line 15 (in `take_damage`)
   - Run: `godot_get_stack_trace()`
   - Run: `godot_get_variables(scope="Locals")`
   - Verify: Can see `amount` parameter and `health` variable

4. **Stepping Test**
   - Hit breakpoint
   - Run: `godot_step_over()`
   - Run: `godot_step_in()` (when at function call)
   - Verify: Execution advances correctly

5. **Expression Evaluation Test**
   - Hit breakpoint
   - Run: `godot_evaluate(expression="health * 2")`
   - Verify: Returns correct calculated value

6. **Launch Modes Test**
   - Test `godot_launch_main_scene()`
   - Test `godot_launch_current_scene()`
   - Test `godot_launch_scene(scene_path="res://test.tscn")`
   - Verify: All three modes launch correctly

7. **Error Recovery Test**
   - Launch game with breakpoint
   - Close Godot editor
   - Run any command
   - Verify: Clear error message about disconnection

8. **Timeout Test**
   - Launch game
   - Try command that might hang (if any)
   - Verify: Timeout after 30s with clear message

---

## Running Tests

### Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/mcp/...
go test ./internal/dap/...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Automated integration test (fully automated with Godot subprocess)
./scripts/automated-integration-test.sh

# Manual integration test (requires running Godot editor)
./scripts/integration-test.sh

# Custom Godot path
export GODOT_BIN=/path/to/godot
./scripts/automated-integration-test.sh

# Custom DAP port
export DAP_PORT=6007
./scripts/integration-test.sh
```

See `tests/INTEGRATION_TEST.md` for detailed integration testing guide.

# Run MCP server for manual testing
echo "Starting MCP server..."
echo "Connect with: godot_connect(port=6006)"
./godot-dap-mcp-server

# Cleanup
kill $GODOT_PID
```

---

## Continuous Integration

### GitHub Actions Example

`.github/workflows/test.yml`:

```yaml
name: Tests

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run unit tests
        run: go test -short -v -cover ./...

  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install Godot
        run: |
          wget https://downloads.tuxfamily.org/godotengine/4.2/Godot_v4.2-stable_linux.x86_64.zip
          unzip Godot_v4.2-stable_linux.x86_64.zip
          chmod +x Godot_v4.2-stable_linux.x86_64

      - name: Start Godot with test project
        run: |
          ./Godot_v4.2-stable_linux.x86_64 --headless --path tests/fixtures/test-project &
          sleep 10

      - name: Run integration tests
        run: go test -v ./tests/integration/...
```

---

## Test Coverage Goals

- **Unit Tests**: >80% coverage for critical paths
- **Integration Tests**: Cover all major workflows
- **Manual Tests**: Verify edge cases and UX

### Current Status

As of Phase 2 completion:
- ✅ MCP layer: 16 tests
- ✅ DAP layer: 12 tests
- ⏳ Integration tests: Pending Phase 3+
- ⏳ Manual testing: Ongoing

---

## References

- [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md) - Component implementation details
- [GODOT_DAP_FAQ.md](reference/GODOT_DAP_FAQ.md) - Common issues and troubleshooting
