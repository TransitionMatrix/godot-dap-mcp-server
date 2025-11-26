# Project Overview: godot-dap-mcp-server

## Purpose
**godot-dap-mcp-server** is an MCP (Model Context Protocol) server that enables AI agents (like Claude Code) to perform interactive runtime debugging of Godot game engine projects via the Debug Adapter Protocol (DAP).

## Key Capabilities
- Connect to running Godot editor's DAP server
- Set/clear breakpoints in GDScript code
- Control execution (continue, step-over, step-in)
- **Inspect runtime state** (stack traces, variables, scopes, evaluate expressions)
- **Format Godot types** (Vector2/3, Color, Nodes with semantic labels)
- **Pause running game** and modify variables at runtime
- **Navigate scene tree** through Node object expansion
- **Launch game scenes** (Main, Current, or Custom scene) for debugging

## Architecture
The system follows a three-layer architecture:

### 1. MCP Layer (`internal/mcp/`)
- Stdio-based JSONRPC 2.0 communication
- **Concurrent Request Handling**: Request handlers run in goroutines to prevent deadlocks from blocked tools.
- **Thread-Safe Transport**: Mutex-protected stdout writing.
- Tool registration and routing.

### 2. DAP Client Layer (`internal/dap/`)
- TCP connection to Godot editor's DAP server (default port: 6006)
- DAP protocol implementation using `github.com/google/go-dap`
- **Event-Driven Architecture**: Robust event filtering and state transitions.
- **Timeout Protection**: All DAP operations wrapped in contexts (10-30s).
- **Deferred ConfigurationDone**: Handles `launch` -> `configurationDone` sequence.
- Session lifecycle management with state machine.

### 3. Tool Layer (`internal/tools/`)
- Godot-specific MCP tools with consistent naming: `godot_<action>_<object>`
- Global session management via `GetSession()` shared across tool calls
- **17 tools across 4 phases**:
  - **Phase 3 (7 tools)**: connect, disconnect, set/clear breakpoints, continue, step-over, step-in
  - **Phase 4 (5 tools)**: get_threads, get_stack_trace, get_scopes, get_variables, evaluate
  - **Phase 5 (3 tools)**: launch_main_scene, launch_scene, launch_current_scene
  - **Phase 6 (2 tools)**: pause, set_variable (set_variable partially supported)
- **Formatting utilities** (`formatting.go`): Pretty-print 15+ Godot types

## Protocol Flow
```
MCP Client → stdio → MCP Server (Concurrent) → TCP/DAP → Godot Editor → Game Instance
```

## Implementation Status
**Version**: v1.0.0 (Released)

- ✅ Phase 1 Complete: Core MCP Server (16 tests)
- ✅ Phase 2 Complete: DAP Client Layer (28 total tests)
- ✅ Phase 3 Complete: Core Debugging Tools (43 total tests)
- ✅ Phase 4 Complete: Runtime Inspection (61 total tests, Godot formatting)
- ✅ Phase 5 Complete: Launch Tools (Integrated & Verified)
- ✅ Phase 6 Complete: Advanced Tools (Pause works; set_variable limited by Godot Engine)
- ✅ Phase 7 Complete: Architecture Refactor (Event-driven, Mock Server)
- ✅ Phase 8 Complete: Error Handling & Polish (Timeouts, Path Resolution, Concurrency Fix)
- ✅ Phase 9 Complete: Documentation (README, TOOLS.md, EXAMPLES.md)

## Key Technical Insights
- **Concurrency prevents deadlocks**: MCP server must handle requests concurrently; otherwise, a blocked `godot_pause` hangs the entire server.
- **Launch Handshake**: `godot_connect` stops at `initialized`. `godot_launch_*` sends `launch` then `configurationDone` to complete handshake.
- **Single DAP session design** (one debugging session at a time)
- **Event filtering required**: DAP sends async events mixed with responses
- **ConfigurationDone required**: After `Initialize()`, must call `ConfigurationDone()` or breakpoints timeout
- Godot validates project paths in launch requests (must have project.godot)
- **Absolute paths required**: Godot DAP doesn't support `res://` paths; server resolves them automatically.
- **Godot type formatting**: Vector2/3, Color, Nodes, Arrays automatically get semantic labels for AI readability

## Testing Infrastructure

**Compliance Testing** (`cmd/test-dap-protocol/`, `scripts/test-dap-compliance.sh`):
- Automated verification of Godot DAP server protocol compliance

**Debug/Test Utilities**:
- `cmd/debug-launch/` - Isolated launch testing utility
- `cmd/test-minimal-dap/` - Minimal protocol verification
- `cmd/dump-setbreakpoints/` - Inspect DAP message serialization
- `cmd/launch-test/` - Validate launch → breakpoints → configurationDone sequence
- `cmd/test-full-debug-workflow/` - End-to-end 17-step debugging workflow test
- `pkg/daptest` - Mock DAP server for offline testing

## Documentation Organization

**Top-level docs/** - Active development documentation
- Core: PLAN, ARCHITECTURE, IMPLEMENTATION_GUIDE, TESTING, DEPLOYMENT
- TOOLS.md - Full tool reference
- EXAMPLES.md - Usage examples

**docs/reference/** - Stable reference materials
- DAP_SESSION_GUIDE.md, GODOT_SOURCE_ANALYSIS.md, CONVENTIONS.md

**docs/godot-upstream/** - Upstream submission materials
