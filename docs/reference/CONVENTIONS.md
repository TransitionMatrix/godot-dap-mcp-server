# Godot DAP MCP Server - Conventions

**Last Updated**: 2025-11-07

This document describes naming conventions, coding patterns, and error message guidelines for the godot-dap-mcp-server.

---

## Table of Contents

1. [Tool Naming Convention](#tool-naming-convention)
2. [Error Message Guidelines](#error-message-guidelines)
3. [Code Style Guidelines](#code-style-guidelines)
4. [Documentation Patterns](#documentation-patterns)

---

## Tool Naming Convention

### Pattern

**Format**: `godot_<action>_<object>`

**Rules**:
- Prefix all tools with `godot_` for clear namespace
- Use verb for action (connect, launch, set, get, etc.)
- Use noun for object (breakpoint, scene, variables, etc.)
- Use snake_case throughout
- Keep names concise but descriptive

### Examples

✅ **Good**:
```
godot_connect
godot_disconnect
godot_set_breakpoint
godot_clear_breakpoint
godot_launch_main_scene
godot_launch_scene
godot_get_variables
godot_get_stack_trace
godot_step_over
godot_step_into
godot_continue
godot_pause
godot_evaluate
```

❌ **Bad**:
```
connect_to_godot           # Action first, not prefixed
set_godot_breakpoint       # Object first
GodotConnect               # Wrong case
godot-connect              # Wrong separator
launch                     # Too generic, not prefixed
get_vars                   # Unclear abbreviation
```

### Rationale

1. **Clear Namespace**: `godot_` prefix prevents conflicts with other MCP servers
2. **AI Recognition**: AI agents easily identify Godot-specific tools
3. **Consistent Pattern**: Verb-object pattern is predictable and scalable
4. **Discoverability**: Tools group together alphabetically
5. **No Ambiguity**: `godot_set_breakpoint` is clearer than `set_breakpoint`

### Special Cases

**Tools with Multiple Objects**:
- Use most specific object: `godot_launch_main_scene` (not `godot_launch_scene_main`)
- Separate with additional underscores: `godot_get_stack_trace`

**Tools with Modifiers**:
- Append modifier: `godot_launch_current_scene`
- Keep action first: `godot_step_over` (not `godot_over_step`)

---

## Error Message Guidelines

### Pattern: Problem + Context + Solution

Every error message should contain three parts:

1. **Problem**: What went wrong
2. **Context**: Why it might have happened
3. **Solution**: How to fix it

### Good vs Bad Examples

#### Example 1: Connection Failure

❌ **Bad**:
```
Error: connection failed
```

✅ **Good**:
```
Failed to connect to Godot DAP server at localhost:6006

Possible causes:
1. Godot editor is not running
2. DAP server is not enabled in editor settings
3. DAP server is using a different port

Solutions:
1. Launch Godot editor
2. Enable DAP in Editor → Editor Settings → Network → Debug Adapter
3. Check port setting (default: 6006)

For more help, see: https://docs.godotengine.org/en/stable/tutorials/editor/debugger_panel.html
```

#### Example 2: Project Path Error

❌ **Bad**:
```
Error: invalid path
```

✅ **Good**:
```
Invalid project path: project.godot not found at /path/to/project

Possible causes:
1. Path does not point to a Godot project directory
2. The path is relative instead of absolute
3. The project.godot file has been moved or deleted

Solutions:
1. Ensure the path points to the directory containing project.godot
2. Use an absolute path: /full/path/to/project
3. Verify the project exists: ls /path/to/project/project.godot
```

#### Example 3: Breakpoint Not Hit

❌ **Bad**:
```
Error: breakpoint failed
```

✅ **Good**:
```
Breakpoint not hit at player.gd:42

Possible causes:
1. Line 42 is not executable code (comment, blank line, etc.)
2. The file path doesn't match the running game's path
3. Code has changed since breakpoint was set

Solutions:
1. Set breakpoint on an executable line (function call, assignment, etc.)
2. Verify file path is absolute and matches game's script location
3. Restart game after setting breakpoints
4. Check breakpoint was verified: look for "verified: true" in response
```

### Error Message Template

```go
return fmt.Errorf(`[PROBLEM STATEMENT]

Possible causes:
1. [Cause 1]
2. [Cause 2]
3. [Cause 3]

Solutions:
1. [Solution 1]
2. [Solution 2]
3. [Solution 3]

[Optional: Link to documentation]`)
```

### Implementation Example

```go
func (c *Client) Connect(ctx context.Context) error {
    conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", c.address())
    if err != nil {
        return fmt.Errorf(`Failed to connect to Godot DAP server at %s

Possible causes:
1. Godot editor is not running
2. DAP server is not enabled in editor settings
3. DAP server is using a different port
4. Firewall is blocking the connection

Solutions:
1. Launch Godot editor and open your project
2. Enable DAP: Editor → Editor Settings → Network → Debug Adapter
3. Check port setting (default: 6006)
4. Check firewall settings for port %d

For more information, see: %s`,
            c.address(),
            c.port,
            "https://docs.godotengine.org/en/stable/tutorials/editor/debugger_panel.html")
    }
    return nil
}
```

### Guidelines

**Do**:
- ✅ Be specific about what failed
- ✅ Provide context (file paths, port numbers, etc.)
- ✅ Suggest concrete actions
- ✅ Link to relevant documentation
- ✅ Use numbered lists for multiple items
- ✅ Be empathetic and helpful

**Don't**:
- ❌ Use generic error messages
- ❌ Use jargon without explanation
- ❌ Blame the user
- ❌ Provide only technical details without solutions
- ❌ Use error codes without descriptions

---

## Code Style Guidelines

### General Go Style

Follow standard Go conventions:
- Use `gofmt` for formatting
- Run `go vet` before committing
- Follow [Effective Go](https://golang.org/doc/effective_go.html)

### Variable Naming

```go
// Good: Clear, descriptive names
projectPath := "/path/to/project"
breakpointLine := 42
dapClient := dap.NewClient("localhost", 6006)

// Bad: Abbreviations, unclear meaning
pp := "/path/to/project"
bp := 42
c := dap.NewClient("localhost", 6006)
```

### Function Naming

```go
// Good: Verbs for actions, clear purpose
func (c *Client) Connect(ctx context.Context) error
func (c *Client) SetBreakpoints(ctx context.Context, file string, lines []int) error
func (t *Tool) Validate() error

// Bad: Unclear, overly abbreviated
func (c *Client) Conn(ctx context.Context) error
func (c *Client) BP(ctx context.Context, f string, l []int) error
func (t *Tool) V() error
```

### Comments

```go
// Good: Explain why, not what
// ReadWithTimeout prevents permanent hangs by wrapping read operations
// with a context deadline. This is critical because some DAP servers
// may not respond to invalid requests.
func (c *Client) ReadWithTimeout(ctx context.Context) (dap.Message, error) {
    // ...
}

// Bad: Restates the obvious
// ReadWithTimeout reads with timeout
func (c *Client) ReadWithTimeout(ctx context.Context) (dap.Message, error) {
    // ...
}
```

### Error Handling

```go
// Good: Wrap errors with context
if err := client.Connect(ctx); err != nil {
    return fmt.Errorf("failed to establish DAP connection: %w", err)
}

// Bad: Lost context
if err := client.Connect(ctx); err != nil {
    return err
}
```

---

## Documentation Patterns

### Tool Description Pattern

```go
Tool{
    Name: "godot_set_breakpoint",
    Description: `[1-sentence summary]

    Prerequisites:
    - [Prerequisite 1]
    - [Prerequisite 2]

    [1-2 sentences describing behavior]

    Use this when you want to:
    - [Use case 1]
    - [Use case 2]
    - [Use case 3]

    Example: [Concrete example]
    godot_set_breakpoint(file="/path/to/script.gd", line=42)`,
}
```

### Function Documentation

```go
// [One-line summary]
//
// [Detailed explanation if needed]
//
// Parameters:
//   - param1: [Description]
//   - param2: [Description]
//
// Returns:
//   - [Description of return value]
//
// Errors:
//   - [Error condition 1]
//   - [Error condition 2]
//
// Example:
//   result, err := Function(param1, param2)
func Function(param1 string, param2 int) (Result, error) {
    // ...
}
```

### Package Documentation

```go
// Package mcp implements the Model Context Protocol server layer.
//
// This package handles stdio-based JSONRPC 2.0 communication with MCP clients,
// tool registration and routing, and response formatting.
//
// Example usage:
//   server := mcp.NewServer()
//   server.RegisterTool(myTool)
//   server.ListenStdio()
package mcp
```

---

## File Organization

### Directory Structure

```
internal/
├── mcp/           # MCP protocol layer
│   ├── server.go      # Core server
│   ├── types.go       # Type definitions
│   ├── transport.go   # stdio communication
│   └── server_test.go # Tests
├── dap/           # DAP client layer
│   ├── client.go      # Core client
│   ├── session.go     # Session management
│   ├── events.go      # Event filtering
│   ├── timeout.go     # Timeout wrappers
│   ├── godot.go       # Godot-specific
│   └── dap_test.go    # Tests
└── tools/         # MCP tools
    ├── connect.go     # Connection tools
    ├── launch.go      # Launch tools
    ├── breakpoints.go # Breakpoint tools
    ├── execution.go   # Execution control
    ├── inspection.go  # Inspection tools
    └── registry.go    # Tool registration
```

### File Naming

- Use lowercase with underscores: `stack_trace.go`
- Group related functionality: `timeout.go`, `timeout_test.go`
- Test files match source: `client.go` → `client_test.go`

---

## Commit Message Convention

### Format

```
<type>: <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions/changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Build/tooling changes

### Examples

```
feat: add godot_step_over tool

Implement step-over command using DAP next request.
Includes timeout protection and event filtering.

Closes #42
```

```
fix: correct timeout handling in ReadWithTimeout

The goroutine was not properly canceled on timeout,
leading to goroutine leaks. Now uses proper context
cancellation.

Fixes #73
```

---

## References

- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [IMPLEMENTATION_GUIDE.md](../IMPLEMENTATION_GUIDE.md) - Component implementation details
