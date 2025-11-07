# Code Style and Conventions

## General Go Style
- Follow standard Go conventions
- Use `gofmt` for automatic formatting
- Run `go vet` before committing
- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines

## Tool Naming Convention
**Pattern**: `godot_<action>_<object>`

Examples:
- `godot_connect` - Connect to DAP server
- `godot_set_breakpoint` - Set a breakpoint
- `godot_launch_main_scene` - Launch main scene
- `godot_step_over` - Step over current line
- `godot_get_stack_trace` - Get stack trace

Rationale:
- Clear namespace with `godot_` prefix
- AI agents easily recognize Godot-specific tools
- Verb-object pattern is predictable and scalable
- Tools group together alphabetically

## Variable Naming
- Use clear, descriptive names (avoid abbreviations)
- Examples:
  - ✅ `projectPath`, `breakpointLine`, `dapClient`
  - ❌ `pp`, `bp`, `c`

## Function Naming
- Use verbs for actions: `Connect()`, `SetBreakpoints()`, `Validate()`
- Be specific and clear about purpose
- Avoid abbreviations

## Error Handling Pattern
**Pattern**: Problem + Context + Solution

Every error message should contain:
1. **Problem**: What went wrong
2. **Context**: Possible causes
3. **Solution**: How to fix it

Example:
```go
return fmt.Errorf(`Failed to connect to Godot DAP server at localhost:6006

Possible causes:
1. Godot editor is not running
2. DAP server is not enabled in editor settings
3. DAP server is using a different port

Solutions:
1. Launch Godot editor
2. Enable DAP in Editor → Editor Settings → Network → Debug Adapter
3. Check port setting (default: 6006)`)
```

## Comments
- Explain **why**, not **what**
- Avoid restating the obvious
- Document complex logic and gotchas

Good:
```go
// ReadWithTimeout prevents permanent hangs by wrapping read operations
// with a context deadline. This is critical because some DAP servers
// may not respond to invalid requests.
```

Bad:
```go
// ReadWithTimeout reads with timeout
```

## Error Wrapping
Always wrap errors with context:
```go
if err := client.Connect(ctx); err != nil {
    return fmt.Errorf("failed to establish DAP connection: %w", err)
}
```

## File Organization
- Use lowercase with underscores: `stack_trace.go`
- Group related functionality: `timeout.go`, `timeout_test.go`
- Test files match source: `client.go` → `client_test.go`

## Commit Message Convention
Format: `<type>: <subject>`

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions/changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Build/tooling changes

Example:
```
feat: add godot_step_over tool

Implement step-over command using DAP next request.
Includes timeout protection and event filtering.

Closes #42
```
