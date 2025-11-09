# Project Overview: godot-dap-mcp-server

## Purpose
**godot-dap-mcp-server** is an MCP (Model Context Protocol) server that enables AI agents (like Claude Code) to perform interactive runtime debugging of Godot game engine projects via the Debug Adapter Protocol (DAP).

## Key Capabilities
- Connect to running Godot editor's DAP server
- Set/clear breakpoints in GDScript code
- Control execution (continue, step-over, step-in)
- **Inspect runtime state** (stack traces, variables, scopes, evaluate expressions)
- **Format Godot types** (Vector2/3, Color, Nodes with semantic labels)
- Launch game scenes for debugging

## Architecture
The system follows a three-layer architecture:

### 1. MCP Layer (`internal/mcp/`)
- Stdio-based JSONRPC 2.0 communication with MCP clients (Claude Code)
- Tool registration and routing
- Request/response handling

### 2. DAP Client Layer (`internal/dap/`)
- TCP connection to Godot editor's DAP server (default port: 6006)
- DAP protocol implementation using `github.com/google/go-dap`
- Event filtering (critical: DAP sends async events mixed with responses)
- Timeout protection (prevents hangs from unresponsive DAP server)
- Session lifecycle management with state machine

### 3. Tool Layer (`internal/tools/`)
- Godot-specific MCP tools with consistent naming: `godot_<action>_<object>`
- Global session management via `GetSession()` shared across tool calls
- **12 tools across 2 phases**:
  - Phase 3 (7 tools): connect, disconnect, set/clear breakpoints, continue, step-over, step-in
  - Phase 4 (5 tools): get_threads, get_stack_trace, get_scopes, get_variables, evaluate
- **Formatting utilities** (`formatting.go`): Pretty-print 15+ Godot types

## Protocol Flow
```
MCP Client (Claude Code) → stdio → MCP Server → TCP/DAP → Godot Editor → Game Instance
```

## Implementation Status
- ✅ Phase 1 Complete: Core MCP Server (16 tests)
- ✅ Phase 2 Complete: DAP Client Layer (28 total tests)
- ✅ Phase 3 Complete: Core Debugging Tools (43 total tests, integration tests working)
- ✅ **Phase 4 Complete: Runtime Inspection** (61 total tests, Godot formatting)
- 🔲 Phase 5-8: Launch tools, advanced features, polish

## Key Technical Insights
- Single DAP session design (one debugging session at a time)
- Event filtering required: DAP sends async events mixed with responses
- All DAP operations need timeout protection (10-30s contexts)
- **ConfigurationDone required**: After `Initialize()`, must call `ConfigurationDone()` or breakpoints timeout
- Godot validates project paths in launch requests (must have project.godot)
- Known issue: `stepOut` command not implemented in Godot's DAP server
- **Absolute paths required**: Godot DAP doesn't support `res://` paths (Godot limitation, not DAP)
- **Godot type formatting**: Vector2/3, Color, Nodes, Arrays automatically get semantic labels for AI readability

## Documentation Organization (Phase 3+)
The project uses a **hybrid documentation approach**:
- **Phase-specific lessons learned** (`docs/LESSONS_LEARNED_PHASE_N.md`) - Debugging narratives
- **IMPLEMENTATION_GUIDE.md** - Reusable patterns extracted from lessons
- **ARCHITECTURE.md** - Critical patterns with rationale
- **FAQ** - Quick troubleshooting answers

See `docs/DOCUMENTATION_WORKFLOW.md` for the complete workflow.
