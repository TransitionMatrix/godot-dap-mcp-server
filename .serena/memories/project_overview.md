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
- Stdio-based JSONRPC 2.0 communication with MCP clients
- Tool registration and routing
- Request/response handling
- JSON Schema generation with proper "any" type handling

### 2. DAP Client Layer (`internal/dap/`)
- TCP connection to Godot editor's DAP server (default port: 6006)
- DAP protocol implementation using `github.com/google/go-dap`
- **Deferred ConfigurationDone**: Handles `launch` -> `configurationDone` sequence correctly
- Event filtering (critical: DAP sends async events mixed with responses)
- ErrorResponse handling in waitForResponse
- Explicit event type cases for all DAP events
- Timeout protection (prevents hangs from unresponsive DAP server)
- Session lifecycle management with state machine (Disconnected → Connected → Initialized → Configured → Launched)

### 3. Tool Layer (`internal/tools/`)
- Godot-specific MCP tools with consistent naming: `godot_<action>_<object>`
- Global session management via `GetSession()` shared across tool calls
- **17 tools across 4 phases**:
  - **Phase 3 (7 tools)**: connect, disconnect, set/clear breakpoints, continue, step-over, step-in
  - **Phase 4 (5 tools)**: get_threads, get_stack_trace, get_scopes, get_variables, evaluate
  - **Phase 5 (3 tools)**: launch_main_scene, launch_scene, launch_current_scene
  - **Phase 6 (2 tools)**: pause, set_variable
- **Formatting utilities** (`formatting.go`): Pretty-print 15+ Godot types

## Protocol Flow
```
MCP Client → stdio → MCP Server → TCP/DAP → Godot Editor → Game Instance
```

## Implementation Status
- ✅ Phase 1 Complete: Core MCP Server (16 tests)
- ✅ Phase 2 Complete: DAP Client Layer (28 total tests)
- ✅ Phase 3 Complete: Core Debugging Tools (43 total tests)
- ✅ Phase 4 Complete: Runtime Inspection (61 total tests, Godot formatting)
- ✅ Phase 5 Complete: Launch Tools (Integrated & Verified)
- ⏳ **Phase 6 WIP: Advanced Tools** (pause, set_variable implemented, runtime testing in progress)
- 🔲 Phase 7-8: Polish, documentation

## Key Technical Insights
- **Launch Handshake**: `godot_connect` stops at `initialized`. `godot_launch_*` sends `launch` then `configurationDone` to complete handshake.
- **Single DAP session design** (one debugging session at a time)
- **Event filtering required**: DAP sends async events mixed with responses
- All DAP operations need timeout protection (10-30s contexts)
- **ConfigurationDone required**: After `Initialize()`, must call `ConfigurationDone()` or breakpoints timeout
- Godot validates project paths in launch requests (must have project.godot)
- **Absolute paths required**: Godot DAP doesn't support `res://` paths (Godot limitation, not DAP)
- **Godot type formatting**: Vector2/3, Color, Nodes, Arrays automatically get semantic labels for AI readability
- **Scene tree navigation**: No dedicated command; use object expansion with Node/* properties
- **Security**: Variable names strictly validated to prevent code injection

## Testing Infrastructure

**Compliance Testing** (`cmd/test-dap-protocol/`, `scripts/test-dap-compliance.sh`):
- Automated verification of Godot DAP server protocol compliance
- Validates Optional field handling, response ordering, Dictionary safety

**Debug/Test Utilities**:
- `cmd/debug-launch/` - Isolated launch testing utility
- `cmd/test-minimal-dap/` - Minimal protocol verification
- `cmd/dump-setbreakpoints/` - Inspect DAP message serialization
- `cmd/launch-test/` - Validate launch → breakpoints → configurationDone sequence
- `cmd/test-full-debug-workflow/` - End-to-end 17-step debugging workflow test

## Documentation Organization

**Top-level docs/** - Active development documentation
- Core: PLAN, ARCHITECTURE, IMPLEMENTATION_GUIDE, TESTING, DEPLOYMENT
- DOCUMENTATION_WORKFLOW.md - Hybrid approach with phase-specific lessons

**docs/reference/** - Stable reference materials
- **DAP_SESSION_GUIDE.md** - Complete DAP command reference with session flow examples
- Protocol details, conventions, FAQ, debugAdapterProtocol.json

**docs/godot-upstream/** - Upstream submission materials
- STRATEGY.md, TESTING_GUIDE.md, PROGRESS.md

**docs/implementation-notes/** - Phase-specific insights
- PHASE_N_IMPLEMENTATION_NOTES.md, LESSONS_LEARNED_PHASE_N.md

**docs/research/** - Research archive
- TEST_RESULTS_PROTOCOL_FIX.md - Verification of DAP protocol fix
