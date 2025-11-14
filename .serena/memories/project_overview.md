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
- Launch game scenes for debugging

## Architecture
The system follows a three-layer architecture:

### 1. MCP Layer (`internal/mcp/`)
- Stdio-based JSONRPC 2.0 communication with MCP clients (Claude Code)
- Tool registration and routing
- Request/response handling
- JSON Schema generation with proper "any" type handling

### 2. DAP Client Layer (`internal/dap/`)
- TCP connection to Godot editor's DAP server (default port: 6006)
- DAP protocol implementation using `github.com/google/go-dap`
- Event filtering (critical: DAP sends async events mixed with responses)
- **ErrorResponse handling** in waitForResponse
- **Explicit event type cases** for all DAP events
- Timeout protection (prevents hangs from unresponsive DAP server)
- Session lifecycle management with state machine

### 3. Tool Layer (`internal/tools/`)
- Godot-specific MCP tools with consistent naming: `godot_<action>_<object>`
- Global session management via `GetSession()` shared across tool calls
- **14 tools across 3 phases**:
  - Phase 3 (7 tools): connect, disconnect, set/clear breakpoints, continue, step-over, step-in
  - Phase 4 (5 tools): get_threads, get_stack_trace, get_scopes, get_variables, evaluate
  - Phase 6 (2 tools): pause, set_variable
- **Formatting utilities** (`formatting.go`): Pretty-print 15+ Godot types

## Protocol Flow
```
MCP Client (Claude Code) → stdio → MCP Server → TCP/DAP → Godot Editor → Game Instance
```

## Implementation Status
- ✅ Phase 1 Complete: Core MCP Server (16 tests)
- ✅ Phase 2 Complete: DAP Client Layer (28 total tests)
- ✅ Phase 3 Complete: Core Debugging Tools (43 total tests, integration tests working)
- ✅ Phase 4 Complete: Runtime Inspection (61 total tests, Godot formatting)
- ⏳ **Phase 6 WIP: Advanced Tools** (pause, set_variable implemented, runtime testing in progress)
- 🔲 Phase 5, 7-8: Launch tools, polish, documentation

## Key Technical Insights
- Single DAP session design (one debugging session at a time)
- Event filtering required: DAP sends async events mixed with responses
- All DAP operations need timeout protection (10-30s contexts)
- **ConfigurationDone required**: After `Initialize()`, must call `ConfigurationDone()` or breakpoints timeout
- Godot validates project paths in launch requests (must have project.godot)
- **Known Godot DAP issues**:
  - `stepOut` command not implemented (will hang)
  - `setVariable` advertised but not implemented (use evaluate() workaround)
  - **pause command may affect DAP response handling** (under investigation)
- **Absolute paths required**: Godot DAP doesn't support `res://` paths (Godot limitation, not DAP)
- **Godot type formatting**: Vector2/3, Color, Nodes, Arrays automatically get semantic labels for AI readability
- **Scene tree navigation**: No dedicated command; use object expansion with Node/* properties
- **Security**: Variable names strictly validated to prevent code injection (regex: `^[a-zA-Z_][a-zA-Z0-9_]*$`)

## Testing Infrastructure

**Compliance Testing** (`cmd/test-dap-protocol/`, `scripts/test-dap-compliance.sh`):
- **Purpose**: Automated verification of Godot DAP server protocol compliance
- **Client capabilities**: LLM-optimized (1-based indexing, type support, client identification)
- **Test sequence**: initialize → launch → setBreakpoints → configurationDone
- **Validates**: Optional field handling, response ordering, Dictionary safety
- **Usage**: `GODOT_BIN=/path/to/godot ./scripts/test-dap-compliance.sh`
- **Detects**: Unsafe Dictionary access patterns in godot-upstream

**Debug/Test Utilities**:
- `cmd/dump-setbreakpoints/` - Inspect DAP message serialization (JSON + wire format)
- `cmd/launch-test/` - Validate launch → breakpoints → configurationDone sequence
- `cmd/test-full-debug-workflow/` - End-to-end 17-step debugging workflow test (13 DAP commands)

**Key Finding**: Godot uses SAFE `.get()` for client capabilities but UNSAFE `[]` for optional DAP request fields (Source.name, Source.checksums)

## Documentation Organization (2025-11-14 Update)

The project uses a **purpose-based documentation structure**:

**Top-level docs/** - Active development documentation
- Core: PLAN, ARCHITECTURE, IMPLEMENTATION_GUIDE, TESTING, DEPLOYMENT
- DOCUMENTATION_WORKFLOW.md - Hybrid approach with phase-specific lessons

**docs/reference/** - Stable reference materials
- **DAP_SESSION_GUIDE.md** - Complete DAP command reference with session flow examples (1373 lines)
- Protocol details, conventions, FAQ
- debugAdapterProtocol.json - Official DAP specification (178KB)

**docs/godot-upstream/** - Upstream submission materials
- STRATEGY.md - Multi-PR submission approach for Godot Dictionary safety fixes
- TESTING_GUIDE.md - How to test Dictionary safety issues
- Templates ready: ISSUE_TEMPLATE.md, PR_TEMPLATE.md
- PROGRESS.md tracks submission status

**docs/implementation-notes/** - Phase-specific insights
- PHASE_N_IMPLEMENTATION_NOTES.md - Pre-implementation research (from /phase-prep skill)
- LESSONS_LEARNED_PHASE_N.md - Post-implementation debugging narratives

**docs/research/** - Research archive
- test-3-analysis.md - DAP compliance test findings and Dictionary safety analysis

**docs/archive/** - Superseded documents (old strategies, early drafts)

### Documentation Workflow

1. **Before phase**: Run `/phase-prep` → creates PHASE_N_IMPLEMENTATION_NOTES.md
2. **During phase**: Write code, test, debug
3. **After phase**: Document lessons in LESSONS_LEARNED_PHASE_N.md
4. **Extract patterns**: Update ARCHITECTURE.md and IMPLEMENTATION_GUIDE.md with reusable patterns
5. **Sync memories**: Run `/memory-sync` to update Serena memories

See `docs/DOCUMENTATION_WORKFLOW.md` for complete details.

## Upstream Contribution Status

**Dictionary Safety Fixes for Godot**:
- **Phase**: Research and testing (WIP)
- **Test tool**: `cmd/test-dap-protocol/` demonstrates unsafe Dictionary access
- **Automation**: `scripts/test-dap-compliance.sh` for automated verification
- **Analysis**: `docs/research/test-3-analysis.md` documents findings
- **Strategy**: Multi-PR approach (8-10 small PRs) in `docs/godot-upstream/STRATEGY.md`
- **Verified**: Godot HEAD (0b5ad6c73c) still has Dictionary safety issues

**Current findings**:
- `Source::from_json` uses unsafe `[]` for optional fields (name, checksums)
- These fields are output-only (Godot ignores client values, regenerates from path)
- Test confirms 2 Dictionary errors with minimal DAP-compliant messages
