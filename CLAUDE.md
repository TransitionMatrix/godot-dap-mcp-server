# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**godot-dap-mcp-server** is an MCP (Model Context Protocol) server that enables AI agents to perform interactive runtime debugging of Godot games via the Debug Adapter Protocol (DAP).

**Key Architecture:**
- Language: Go (single binary, no runtime dependencies)
- Protocol Flow: MCP Client (Claude Code) → stdio → MCP Server → TCP/DAP → Godot Editor → Game Instance
- Based on previous DAP experimentation (11/13 commands proven working)
- Single DAP session design (one debugging session at a time)

## Development Commands

### Building
```bash
# Development build
go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go

# Cross-platform builds (run from project root)
./scripts/build.sh

# Manual cross-compilation examples
GOOS=darwin GOARCH=arm64 go build -o build/godot-dap-mcp-server-darwin-arm64 cmd/godot-dap-mcp-server/main.go
GOOS=linux GOARCH=amd64 go build -o build/godot-dap-mcp-server-linux-amd64 cmd/godot-dap-mcp-server/main.go
```

### Testing
```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./internal/mcp/...

# Run integration tests (requires running Godot editor)
go test ./tests/integration/...

# Launch test Godot project
./scripts/test-godot.sh
```

### Dependencies
```bash
# Install/update dependencies
go mod tidy

# Add new dependency
go get github.com/package/name
```

## Architecture

### Layer Structure

The codebase is organized in three primary layers:

1. **MCP Layer** (`internal/mcp/`): stdio-based MCP protocol handling
   - `server.go`: Core MCP server, request routing, stdio communication
   - `types.go`: MCP request/response types (JSONRPC 2.0)
   - `transport.go`: stdin/stdout message handling

2. **DAP Client Layer** (`internal/dap/`): TCP connection to Godot's DAP server
   - `client.go`: DAP protocol implementation using `github.com/google/go-dap`
   - `session.go`: DAP session lifecycle management
   - `events.go`: Async event filtering (critical for response parsing)
   - `timeout.go`: Timeout wrappers for all DAP operations (prevents hangs)
   - `godot.go`: Godot-specific DAP extensions and launch parameters

3. **Tool Layer** (`internal/tools/`): Godot-specific MCP tools
   - `connect.go`: Connection management tools
   - `launch.go`: Scene launching tools (main/current/custom)
   - `breakpoints.go`: Breakpoint management
   - `execution.go`: Execution control (continue, step-over, step-in, step-out)
   - `inspection.go`: Runtime inspection (stack, variables, evaluation)
   - `registry.go`: Tool registration and metadata

### Critical Implementation Patterns

**Event Filtering (discovered through DAP testing):**
```go
// DAP sends async events mixed with responses
// Must filter events to get expected response
for {
    msg, err := c.ReadWithTimeout(ctx)
    if err != nil {
        return err
    }

    switch m := msg.(type) {
    case *dap.Response:
        if m.Command == expectedCommand {
            return processResponse(m)
        }
    case *dap.Event:
        // Log but don't return - continue waiting
        c.logEvent(m)
    default:
        continue
    }
}
```

**Timeout Protection:**
All DAP operations must use context-based timeouts (10-30s) to prevent permanent hangs:
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

return c.waitForResponse(ctx, "commandName")
```

**Godot Launch Flow:**
1. Send `launch` request with arguments (project, scene, options)
2. Send `configurationDone` to trigger actual launch
3. Wait for launch response
4. Game runs until breakpoint/pause/exit

### Known Issues

**step-out Command:** The step-out DAP command has a known hang issue from testing. When implementing, add extra timeout protection and investigate response parsing.

**Project Path Validation:** Godot validates the project path in launch requests. Always verify `project.godot` exists before sending launch request.

## Tool Design Principles

### Naming Convention
- Pattern: `godot_<action>_<object>`
- Examples: `godot_connect`, `godot_set_breakpoint`, `godot_launch_main_scene`
- Rationale: AI agents easily recognize Godot-specific tools; consistent namespace

### Tool Description Format
Tool descriptions must be AI-optimized with:
- Clear purpose statement
- Use cases ("Use this when you want to...")
- Concrete examples
- Parameter descriptions with types and validation rules
- Expected return values

### Error Messages
Follow pattern: **Problem + Context + Solution**

Good error message example:
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
```

## Implementation Status

**Current Phase**: Phase 2 (DAP Client Implementation)

The project is being built in phases:

- ✅ **Phase 1 (COMPLETE):** Core MCP Server - stdio server with test tool
  - All success criteria met (2025-11-06)
  - 16 unit tests passing
  - Verified working with Claude Code
  - See commit `8b1aa15`
- ⏳ **Phase 2 (IN PROGRESS):** DAP Client Migration - port working code with timeouts
- 🔲 **Phase 3:** Core Debugging Tools - 8 essential tools (connect, breakpoints, stepping)
- 🔲 **Phase 4:** Inspection Tools - 5 inspection tools (stack, variables, evaluate)
- 🔲 **Phase 5:** Launch Tools - 3 launch variants (main, custom, current scene)
- 🔲 **Phase 6:** Advanced Tools - pause, set variable, scene tree inspection
- 🔲 **Phase 7:** Error Handling & Polish - timeout implementation, step-out fix
- 🔲 **Phase 8:** Documentation - README, TOOLS.md, EXAMPLES.md

**Phase 1 Deliverables**:
- `internal/mcp/`: Complete MCP protocol implementation (types, transport, server)
- `internal/tools/`: Tool registration system with `godot_ping` test tool
- `cmd/godot-dap-mcp-server/main.go`: Entry point
- Test suite: 16 tests covering critical paths
- Binary: 2.9MB, tested with Claude Code MCP client

## Testing Strategy

### Unit Tests
Focus on:
- MCP protocol parsing
- Tool parameter validation
- Error message formatting
- Timeout mechanisms

### Integration Tests
Requires running Godot editor with DAP enabled. Tests should verify:
- Full MCP → DAP → Godot flow
- Breakpoint setting and hitting
- Variable inspection at breakpoints
- Stepping commands execution
- Scene launching functionality

Test fixture project located at: `tests/fixtures/test-project/`

### Manual Testing
Use the test Godot project to verify:
1. Connection and capability detection
2. Breakpoint workflow (set → launch → hit → inspect)
3. Stepping through code (next, step-in, step-out)
4. Variable inspection at different stack frames
5. GDScript expression evaluation
6. Different launch scenarios
7. Error recovery (Godot restart, connection drops)
8. Timeout handling on hung commands

## Godot DAP Protocol Details

### Connection
- Default port: 6006
- Protocol: TCP to localhost
- Must send `initialize` then `configurationDone`

### Launch Request Arguments
```go
{
    "project": "/absolute/path/to/project",  // Required, validated
    "scene": "main" | "current" | "res://path/to/scene.tscn",
    "platform": "host" | "android" | "web",  // Default: "host"
    "noDebug": false,
    "profiling": false,
    "debug_collisions": false,
    "debug_paths": false,
    "debug_navigation": false,
    "additional_options": "string"
}
```

### Scene Launch Modes
- `"main"`: Launch project's main scene (from project.godot)
- `"current"`: Launch currently open scene in editor
- `"res://path"`: Launch specific scene by path

## Common Patterns

### Adding a New Tool

1. Create handler in appropriate file (`internal/tools/`)
2. Define tool metadata (name, description, parameters)
3. Implement handler function calling DAP client
4. Register tool in `registry.go`
5. Add unit tests
6. Add integration test (if applicable)
7. Document in `docs/TOOLS.md`

### Adding DAP Functionality

1. Implement DAP request in `internal/dap/client.go`
2. Add timeout wrapper
3. Add event filtering in response handler
4. Create corresponding tool wrapper in `internal/tools/`
5. Test with running Godot instance

## Dependencies

Core dependencies:
- `github.com/google/go-dap`: DAP protocol implementation
- Standard library: `encoding/json`, `net`, `context`, `bufio`

Optional:
- `github.com/spf13/cobra`: CLI commands (if needed)
- `github.com/sirupsen/logrus`: Structured logging

## Distribution

### Binary Locations
Compiled binaries go to: `build/`

### Platforms
- macOS (Intel): `darwin-amd64`
- macOS (Apple Silicon): `darwin-arm64`
- Linux: `linux-amd64`
- Windows: `windows-amd64.exe`

### User Installation

Users must obtain the binary then register it with Claude Code (unlike npm/Python servers):

```bash
# Quick install from releases
mkdir -p ~/.claude/mcp-servers/godot-dap/
curl -L -o ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server \
  https://github.com/username/godot-dap-mcp-server/releases/download/v1.0.0/godot-dap-mcp-server-[platform]
chmod +x ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server
claude mcp add godot-dap ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server
```

Alternative: Build from source then register:
```bash
go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go
mkdir -p ~/.claude/mcp-servers/godot-dap/
mv godot-dap-mcp-server ~/.claude/mcp-servers/godot-dap/
claude mcp add godot-dap ~/.claude/mcp-servers/godot-dap/godot-dap-mcp-server
```

Manual config (`~/.claude/mcp.json`):
```json
{
  "mcpServers": {
    "godot-dap": {
      "command": "/Users/username/.claude/mcp-servers/godot-dap/godot-dap-mcp-server",
      "args": [],
      "env": {}
    }
  }
}
```

## Code Style

- Follow standard Go conventions (gofmt, go vet)
- Use clear variable names (avoid abbreviations)
- Add comments for non-obvious logic
- Include context in error messages
- Prefer explicit over clever code
- Use context.Context for all potentially long-running operations
