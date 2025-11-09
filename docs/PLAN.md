# Godot DAP MCP Server - Implementation Plan

**Date**: 2025-11-05
**Last Updated**: 2025-11-08
**Language**: Go
**Purpose**: MCP server providing interactive runtime debugging for Godot games via DAP protocol

---

## Current Status

**Repository**: https://github.com/TransitionMatrix/godot-dap-mcp-server

### Phase Completion

- ✅ **Phase 1: Core MCP Server** - COMPLETE (2025-11-06)
  - All success criteria met
  - 16 unit tests passing
  - Binary tested with Claude Code MCP client
  - Commit: `8b1aa15`

- ✅ **Phase 2: DAP Client Implementation** - COMPLETE (2025-11-06)
  - All success criteria met
  - 12 unit tests passing (28 total with Phase 1)
  - Event filtering and timeout protection implemented
  - Session management and Godot-specific launch configs
  - Commit: `3079ee8`

- ✅ **Phase 3: Core Debugging Tools** - COMPLETE (2025-11-07)
  - All 7 core debugging tools implemented
  - 43 unit tests passing
  - Integration tests working with automated Godot launch
  - Session management and event filtering patterns established

- ✅ **Phase 4: Runtime Inspection Tools** - COMPLETE (2025-11-08)
  - All 5 inspection tools implemented
  - 61 unit tests passing (including 11 formatting tests)
  - Stack traces, scopes, variables, and expression evaluation working
  - Godot-specific type formatting (Vector2/3, Color, Nodes, Arrays, etc.)
  - Integration test verifies all tools registered

- 🔲 **Phase 5: Launch Tools** - PENDING
- 🔲 **Phase 6: Advanced Tools** - PENDING
- 🔲 **Phase 7: Error Handling & Polish** - PENDING
- 🔲 **Phase 8: Documentation** - PENDING

---

## Executive Summary

Build a specialized MCP server that bridges Claude Code (or any MCP client) to Godot's Debug Adapter Protocol (DAP) server, enabling AI agents to perform interactive debugging of Godot games.

**Key Differentiators**:
- First MCP server for Godot interactive debugging
- Complements existing Godot MCP servers (GDAI focuses on editor/scene manipulation)
- Leverages existing tested Go codebase (11/13 commands proven working)
- stdio-based standard MCP pattern (no separate server process)
- Single binary distribution (no runtime dependencies)

---

## Project Goals

### Primary Goals

1. **Enable AI-driven debugging**: Allow Claude Code and other MCP clients to debug Godot games interactively
2. **Leverage prior experiments**: Build on previous DAP testing (11/13 commands proven working)
3. **Godot-specific enhancements**: Support GDScript syntax, scene tree navigation, Godot object inspection
4. **Production-ready**: Implement timeout mechanisms, error handling, graceful failure recovery
5. **Easy deployment**: Single binary, stdio-based, zero configuration

### Secondary Goals

1. **Fill market gap**: Provide capability that GDAI MCP and other Godot servers don't offer
2. **Industry standard**: Use Microsoft's DAP protocol (future-proof)
3. **AI agent optimized**: Tool descriptions and parameters designed for LLM understanding
4. **Open source contribution**: Share with Godot and MCP communities

---

## Documentation Structure

This plan references detailed documentation in separate files:

- **[ARCHITECTURE.md](ARCHITECTURE.md)** - System design, patterns, and design decisions
- **[IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md)** - Component implementation specifications
- **[TESTING.md](TESTING.md)** - Testing strategies and procedures
- **[DEPLOYMENT.md](DEPLOYMENT.md)** - Build process and distribution
- **[reference/GODOT_DAP_FAQ.md](reference/GODOT_DAP_FAQ.md)** - Troubleshooting Q&A
- **[reference/DAP_PROTOCOL.md](reference/DAP_PROTOCOL.md)** - Protocol details
- **[reference/CONVENTIONS.md](reference/CONVENTIONS.md)** - Naming and coding conventions
- **[reference/GODOT_SOURCE_ANALYSIS.md](reference/GODOT_SOURCE_ANALYSIS.md)** - Godot source code findings

---

## Implementation Phases

### Phase 1: Core MCP Server - ✅ COMPLETE

**Goal**: Get stdio MCP server running with one test tool

**Completion Date**: 2025-11-06
**Commit**: `8b1aa15`

**Deliverables**:
- ✅ Working stdio MCP server
- ✅ Tool registration system
- ✅ One test tool (`godot_ping`)
- ✅ Comprehensive test suite (16 tests)
- ✅ Verified with Claude Code

**Success Criteria**:
- ✅ Claude Code can spawn server
- ✅ Server receives MCP requests via stdin
- ✅ Server responds via stdout
- ✅ Clean shutdown on EOF

---

### Phase 2: DAP Client Implementation - ✅ COMPLETE

**Goal**: Implement DAP client with patterns from experimentation

**Completion Date**: 2025-11-06
**Commit**: `3079ee8`

**Key Components Implemented**:
- ✅ DAP client (`internal/dap/client.go`)
- ✅ Session management (`internal/dap/session.go`)
- ✅ Event filtering (`internal/dap/events.go`)
- ✅ Timeout wrappers (`internal/dap/timeout.go`)
- ✅ Godot-specific configs (`internal/dap/godot.go`)

**Deliverables**:
- ✅ DAP client with proven commands
- ✅ Timeout protection (10-30s timeouts)
- ✅ Event filtering for async events
- ✅ 12 unit tests (28 total with Phase 1)

**Success Criteria**:
- ✅ Can connect to Godot DAP server (port 6006)
- ✅ Can send initialize/configurationDone
- ✅ Session state machine working
- ✅ All tests passing

**See**: [ARCHITECTURE.md](ARCHITECTURE.md) for critical implementation patterns

---

### Phase 3: Core Debugging Tools - ✅ COMPLETE

**Goal**: Implement essential debugging tools using DAP client

**Tools Implemented** (7 core tools):

1. ✅ **`godot_connect`** - Establish DAP connection with configurationDone handshake
   - Parameters: `port` (default: 6006)
   - Returns: Connection status, state (configured)
   - Location: `internal/tools/connect.go`

2. ✅ **`godot_disconnect`** - Close DAP connection
   - Parameters: none
   - Returns: Disconnection status
   - Location: `internal/tools/connect.go`

3. ✅ **`godot_set_breakpoint`** - Set breakpoint in GDScript file
   - Parameters: `file` (absolute path), `line` (integer)
   - Returns: Breakpoint ID, verification status
   - Location: `internal/tools/breakpoints.go`

4. ✅ **`godot_clear_breakpoint`** - Clear all breakpoints in file
   - Parameters: `file` (absolute path)
   - Returns: Success status
   - Location: `internal/tools/breakpoints.go`

5. ✅ **`godot_continue`** - Resume execution
   - Parameters: none (uses threadId=1)
   - Returns: Success status
   - Location: `internal/tools/execution.go`

6. ✅ **`godot_step_over`** - Step over current line
   - Parameters: none (uses threadId=1)
   - Returns: Success status
   - Location: `internal/tools/execution.go`

7. ✅ **`godot_step_into`** - Step into function
   - Parameters: none (uses threadId=1)
   - Returns: Success status
   - Location: `internal/tools/execution.go`

**⚠️ IMPORTANT**: `step_out` is **NOT implemented** in Godot's DAP server. See [GODOT_SOURCE_ANALYSIS.md](reference/GODOT_SOURCE_ANALYSIS.md).

**Deliverables**:
- ✅ 7 MCP tools for debugging
- ✅ 4 new DAP client methods (SetBreakpoints, Continue, Next, StepIn)
- ✅ Global session management via GetSession()
- ✅ 15 unit tests (43 total with Phases 1-2)
- ✅ Fully automated integration test with Godot subprocess
- ✅ Manual integration test for running editor

**Success Criteria**:
- ✅ Can set/clear breakpoints via MCP tools
- ✅ Can control execution (continue, step over, step into)
- ✅ No hangs (30s timeout protection on all operations)
- ✅ Event filtering works for async DAP responses
- ✅ Persistent MCP session via named pipes
- ✅ All integration tests passing (6/6)
- Clear error messages on failure

**See**: [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md) for tool implementation patterns

---

### Phase 4: Inspection Tools - ✅ COMPLETE

**Goal**: Implement runtime inspection tools

**Tools Implemented** (5 tools):
1. ✅ `godot_get_threads` - Get active threads (always returns thread ID 1)
2. ✅ `godot_get_stack_trace` - Get call stack with frame IDs
3. ✅ `godot_get_scopes` - Get variable scopes (Locals, Members, Globals)
4. ✅ `godot_get_variables` - Get variables in scope or expand complex objects
5. ✅ `godot_evaluate` - Evaluate GDScript expressions in context

**Implementation Details**:
- Added 5 DAP client methods: `Threads()`, `StackTrace()`, `Scopes()`, `Variables()`, `Evaluate()`
- Created `internal/tools/inspection.go` with 5 MCP tools
- All tools follow event filtering and timeout protection patterns
- Variables marked as expandable when `variablesReference > 0`

**Testing**:
- 61 passing unit tests (up from 43 in Phase 3)
- Integration test verifies all 5 tools registered
- Tool descriptions AI-optimized with prerequisites, use cases, examples
- Comprehensive formatting tests for 15+ Godot types

**Godot-Specific Formatting** (Polish):
- ✅ Vector2/3/4 - `Vector2(x=10, y=20)`
- ✅ Color - `Color(r=1.0, g=0.5, b=0.0, a=1.0)`
- ✅ Rect2, AABB - Bounding boxes with pos/size labels
- ✅ Transform2D/3D, Basis - Simplified representations
- ✅ Node types - Class name with instance ID
- ✅ Arrays - Element count with preview
- ✅ Dictionaries - Key count hints

**Success Criteria**:
- ✅ Can inspect call stack at breakpoint
- ✅ Can view variable values in all scopes (Locals, Members, Globals)
- ✅ Can evaluate GDScript expressions
- ✅ Godot objects formatted nicely with semantic labels

**Completion Date**: 2025-11-08

---

### Phase 5: Launch Tools - 🔲 PENDING

**Goal**: Implement scene launching via DAP

**Tools to Implement** (3 tools):
1. `godot_launch_main_scene` - Launch project's main scene
2. `godot_launch_scene` - Launch specific scene by path
3. `godot_launch_current_scene` - Launch currently open scene

**Launch Parameters**:
- Project path (validated)
- Scene mode (main/current/custom)
- Platform (host/android/web)
- Debug options (profiling, collision visualization, etc.)

**Critical Implementation**:
- Two-step launch process (launch + configurationDone)
- Project path validation (must contain project.godot)
- See [DAP_PROTOCOL.md](reference/DAP_PROTOCOL.md) for details

**Success Criteria**:
- Can launch main scene via MCP tool (no manual F5!)
- Can launch specific scenes by path
- Clear errors for invalid scenes
- Game launches with breakpoints active

---

### Phase 6: Advanced Tools - 🔲 PENDING

**Goal**: Add nice-to-have debugging tools

**Tools to Implement** (4 tools):
1. `godot_pause` - Pause execution
2. `godot_set_variable` - Modify variable at runtime
3. `godot_get_scene_tree` - Inspect scene structure
4. `godot_inspect_node` - Inspect Node properties

**Success Criteria**:
- Can pause running game
- Can modify variables at runtime
- Can inspect scene tree structure

---

### Phase 7: Error Handling & Polish - 🔲 CRITICAL

**Goal**: Production-ready error handling and user experience

**Tasks**:
1. **Timeout Implementation**: Ensure all DAP requests have timeouts
2. **Error Message Formatting**: Problem + Context + Solution pattern
3. **Graceful Degradation**: Handle Godot editor restart
4. **Logging**: Structured logging with debug mode
5. **Path Resolution Enhancement**:
   - Add optional `project` parameter to `godot_connect`
   - Store project root in Session for `res://` path conversion
   - Create `ResolveGodotPath()` utility to convert `res://` → absolute paths
   - Auto-convert paths in `godot_set_breakpoint` and `godot_clear_breakpoint`
   - Update tool descriptions to document both path formats
   - Add integration tests for `res://` path handling

**Success Criteria**:
- No permanent hangs (all requests timeout)
- Helpful error messages
- Survives Godot editor restart
- Supports both absolute and `res://` paths transparently

**Technical Note**: Godot's DAP server requires absolute paths (not a DAP protocol limitation). We convert `res://` paths client-side for better UX.

**See**: [CONVENTIONS.md](reference/CONVENTIONS.md) for error message guidelines

---

### Phase 8: Documentation - 🔲 HIGH PRIORITY

**Goal**: Comprehensive documentation for users and AI agents

**Documents to Create/Update**:
1. **README.md** - Project overview, installation, quick start
2. **TOOLS.md** - Complete tool reference
3. **EXAMPLES.md** - Real debugging workflows
4. **CHANGELOG.md** - Version history

**Tool Description Pattern**: See [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md#tool-description-guidelines)

**Success Criteria**:
- AI agents can use server without human help
- Clear tool descriptions
- Working code examples

---

## Timeline Summary

| Phase | Days | Priority | Status | Deliverable |
|-------|------|----------|--------|-------------|
| 1. MCP Server Core | 1 | CRITICAL | ✅ COMPLETE | stdio server with test tool |
| 2. DAP Client | 1-2 | CRITICAL | ✅ COMPLETE | DAP client + timeouts |
| 3. Core Debugging Tools | 1 | HIGH | ✅ COMPLETE | 7 essential tools + tests |
| 4. Inspection Tools | 1 | HIGH | 🔲 PENDING | 5 inspection tools |
| 5. Launch Tools | 1 | MEDIUM | 🔲 PENDING | 3 launch variants |
| 6. Advanced Tools | 1 | OPTIONAL | 🔲 PENDING | 4 nice-to-have tools |
| 7. Error Handling | 1 | CRITICAL | 🔲 PENDING | Timeouts, recovery, paths |
| 8. Documentation | 1 | HIGH | 🔲 PENDING | Complete docs |
| **Total** | **4-5 days** | | | **Production-ready server** |

---

## Success Metrics

### Functional Metrics
- ✅ All core debugging tools working (connect, breakpoints, stepping, inspection)
- ⏳ Launch functionality working (main, custom, current scenes)
- ⏳ No permanent hangs (timeout mechanisms working)
- ⏳ Clear error messages on failure

### Performance Metrics
- Startup time: <100ms
- Connection time: <1s
- Breakpoint hit latency: <500ms
- Step command latency: <500ms
- Binary size: <10MB

### User Experience Metrics
- AI agents can debug without human intervention
- Clear tool descriptions for LLM understanding
- Helpful error messages suggest fixes
- Works on first try (zero configuration)

---

## Next Steps

### Immediate Actions (Phase 3)

1. **Implement `godot_connect` tool**
   - Create `internal/tools/connect.go`
   - Use Phase 2 Session to establish connection
   - Return connection status and capabilities

2. **Implement execution control tools**
   - `godot_continue`, `godot_step_over`, `godot_step_into`
   - Use DAP client methods from Phase 2
   - Handle async events properly

3. **Implement breakpoint tools**
   - `godot_set_breakpoint`, `godot_clear_breakpoint`
   - Use absolute paths (validate)
   - Handle Godot's path requirements

4. **Testing**
   - Add unit tests for each tool
   - Create integration test with running Godot
   - Verify end-to-end workflow

**See**: [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md) for implementation patterns

---

## Key References

### Implementation
- [ARCHITECTURE.md](ARCHITECTURE.md) - Critical implementation patterns
- [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md) - Component specifications
- [TESTING.md](TESTING.md) - Testing procedures
- [DEPLOYMENT.md](DEPLOYMENT.md) - Build and distribution

### Protocol Details
- [reference/GODOT_DAP_FAQ.md](reference/GODOT_DAP_FAQ.md) - Common questions
- [reference/DAP_PROTOCOL.md](reference/DAP_PROTOCOL.md) - Protocol specifications
- [reference/GODOT_SOURCE_ANALYSIS.md](reference/GODOT_SOURCE_ANALYSIS.md) - Source findings

### Conventions
- [reference/CONVENTIONS.md](reference/CONVENTIONS.md) - Naming and error patterns

---

## Critical Findings

### stepOut Not Implemented
⚠️ **Godot's DAP server does NOT implement stepOut command**. Do not attempt to implement `godot_step_out` tool. See [GODOT_SOURCE_ANALYSIS.md](reference/GODOT_SOURCE_ANALYSIS.md) for details.

### Launch Requires Two Steps
The DAP `launch` request only stores parameters. Must send `configurationDone` to actually launch game. See [DAP_PROTOCOL.md](reference/DAP_PROTOCOL.md#launch-flow).

### Event Filtering is Critical
DAP sends async events mixed with responses. Event filtering pattern (in Phase 2) is essential to prevent hangs. See [ARCHITECTURE.md](ARCHITECTURE.md#critical-implementation-patterns).

---

## Questions & Answers

**Q: Can I test anything end-to-end yet?**
**A**: After Phase 2, only Phase 1 MCP server (godot_ping) is testable. Phase 3 will add first real debugging tools.

**Q: Where do I find detailed implementation specs?**
**A**: See [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md) for complete component specifications with code examples.

**Q: How do I handle errors properly?**
**A**: Follow Problem + Context + Solution pattern in [CONVENTIONS.md](reference/CONVENTIONS.md#error-message-guidelines).

**Q: What if I encounter DAP issues?**
**A**: Check [GODOT_DAP_FAQ.md](reference/GODOT_DAP_FAQ.md) for common troubleshooting.

---

**Last Updated**: 2025-11-07
**Project Status**: Phase 2 Complete, Phase 3 In Progress
