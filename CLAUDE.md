# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**godot-dap-mcp-server** is an MCP (Model Context Protocol) server that enables AI agents to perform interactive runtime debugging of Godot games via the Debug Adapter Protocol (DAP).

**Key Architecture:**
- Language: Go (single binary, no runtime dependencies)
- Protocol Flow: MCP Client (Claude Code) → stdio → MCP Server → TCP/DAP → Godot Editor → Game Instance
- Based on previous DAP experimentation (11/13 commands proven working)
- Single DAP session design (one debugging session at a time)

## Available Skills

The following Claude Code skills are available for this project:

### mcp-builder
**Location:** `.claude/skills/mcp-builder/` (installed via `.claude/scripts/setup-skills.sh`)

**Purpose:** Provides guidance for creating high-quality MCP servers, including:
- MCP protocol best practices
- Tool design patterns
- Error handling strategies
- Testing approaches for MCP servers

**When to use:** Invoke this skill when designing new MCP tools, implementing protocol features, or needing MCP-specific architectural guidance.

### memory-sync
**Location:** `.claude/skills/memory-sync/`

**Purpose:** Guided workflow for maintaining strategic redundancy between project memories and documentation:
- Assess what changed since last sync
- Update memories with concise summaries
- Flag documentation that needs comprehensive updates
- Verify sync completeness

**When to use:** After phase completions, architecture changes, new pattern discoveries, or major refactoring (5+ files changed).

## Documentation

Comprehensive documentation is organized into focused documents:

### Core Documentation
- **[docs/PLAN.md](docs/PLAN.md)** - Project status, implementation phases, and timeline
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - System design, patterns, and critical implementation details
- **[docs/IMPLEMENTATION_GUIDE.md](docs/IMPLEMENTATION_GUIDE.md)** - Component specifications and code examples
- **[docs/TESTING.md](docs/TESTING.md)** - Testing strategies, procedures, and examples
- **[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)** - Build process, distribution, and installation

### Reference Documentation
- **[docs/reference/GODOT_DAP_FAQ.md](docs/reference/GODOT_DAP_FAQ.md)** - Common questions and troubleshooting
- **[docs/reference/DAP_PROTOCOL.md](docs/reference/DAP_PROTOCOL.md)** - Godot DAP protocol details
- **[docs/reference/CONVENTIONS.md](docs/reference/CONVENTIONS.md)** - Naming conventions and coding standards
- **[docs/reference/GODOT_SOURCE_ANALYSIS.md](docs/reference/GODOT_SOURCE_ANALYSIS.md)** - Findings from analyzing Godot source code

**When implementing features**: Refer to ARCHITECTURE.md for patterns, IMPLEMENTATION_GUIDE.md for component specs, and reference docs for protocol details.

## Code Navigation and Memory Management

### Using the Source MCP Servers

This project uses two symbolic code navigation MCP servers for token-efficient code exploration:

**MCP Servers:**
- `source` (tools: `mcp__source__*`) - Navigate this project's Go/Bash codebase
- `godot-source` (tools: `mcp__godot-source__*`) - Navigate Godot engine C++ source code

**Configured Languages** (`.serena/project.yml`):
- **Go** (`.go`) - Full LSP support with symbolic navigation
- **Bash** (`.sh`) - Functions and variables indexed
- **Markdown** (`.md`) - Pattern search only (no symbols)

**Prefer symbolic tools over reading entire files:**
- Use `get_symbols_overview` to understand file structure before reading
- Use `find_symbol` with `include_body: false` to explore without loading full code
- Use `find_referencing_symbols` to understand dependencies
- Use `search_for_pattern` for flexible regex searches
- Only read full files when absolutely necessary

**Example workflow:**
```bash
# 1. Get overview of a file
get_symbols_overview(relative_path="internal/dap/client.go")

# 2. Find specific symbol with depth
find_symbol(name_path="Client", depth=1, include_body=false)

# 3. Read only needed method
find_symbol(name_path="Client/Connect", include_body=true)

# 4. Navigate bash scripts symbolically
get_symbols_overview(relative_path="scripts/automated-integration-test.sh")
find_symbol(name_path="send_mcp_request", include_body=true)
```

### Project Memories (Strategic Redundancy)

Project knowledge is stored in memory files (`.serena/memories/`) managed by the `source` MCP server:
- **project_overview.md** - Architecture, purpose, implementation status
- **tech_stack.md** - Dependencies and build system
- **code_style_and_conventions.md** - Naming patterns and error handling
- **suggested_commands.md** - Development commands
- **task_completion_checklist.md** - Post-task verification steps
- **codebase_structure.md** - Directory layout and layers
- **critical_implementation_patterns.md** - Essential patterns (event filtering, timeouts, Godot launch flow)

**Memory vs Documentation Strategy:**

**Project Memories** (`.serena/memories/`) - Quick context for code navigation (token-efficient)
- Updated frequently as code evolves
- Concise summaries (1-2 pages max)
- What you need to know before exploring code
- Reflects current code reality
- Accessed via `mcp__source__read_memory()`, `mcp__source__write_memory()`

**Documentation** (`docs/`) - Comprehensive reference for humans
- Stable, version-controlled knowledge
- Full details with examples and rationale
- Read these when implementing new features
- Reflects intended design and best practices

**When memories diverge from docs**: Memories reflect current implementation; docs reflect design intent. This divergence signals needed doc updates.

### Memory Management Workflow

**When to sync memories** (invoke `/memory-sync` skill):

**Must sync immediately:**
- ✅ Phase completion (update `project_overview.md` status)
- ✅ New architectural pattern discovered (update `critical_implementation_patterns.md`)
- ✅ Tool naming or error message pattern changes (update `code_style_and_conventions.md`)
- ✅ Major refactoring (5+ files changed)

**Should sync soon:**
- New development commands added to workflow
- Directory structure changes
- Convention updates

**Don't sync for:**
- Bug fixes in existing code
- Test additions without new patterns
- Minor documentation updates
- Comment improvements

**To sync memories:**
```bash
# Invoke the memory-sync skill
/memory-sync

# Or manually:
# 1. Review changes: git log --oneline -10
# 2. Update affected memories: write_memory(...)
# 3. Flag docs for comprehensive updates
# 4. Commit changes: git commit -m "docs: sync memories with..."
```

### Godot Source Reference (godot-source)

The `godot-source` MCP server provides access to the **Godot engine C++ source code** at `/Users/adp/Projects/godot`. Use this when implementing DAP features to understand Godot's reference implementation.

**Available memory files (14 total):**

**DAP Protocol Implementation:**
- `dap_architecture_overview` - Architecture of Godot's DAP server
- `dap_implementation_reference` - Code examples and implementation patterns
- `dap_supported_commands` - Which DAP commands Godot implements
- `dap_events_and_notifications` - Event handling and async notifications
- `dap_connection_and_config` - Connection setup and configuration
- `dap_known_issues_and_quirks` - Known bugs (e.g., stepOut hang)
- `dap_faq_for_clients` - Common questions for DAP client developers

**Godot Engine Core:**
- `codebase_structure` - Directory layout and module organization
- `build_system` - SCons build system details
- `object_system_architecture` - Godot's Object class system
- `code_style_and_conventions` - C++ coding standards
- `project_overview` - Overview of the Godot engine codebase
- `suggested_commands` - Common development commands
- `task_completion_checklist` - Verification steps

**When to consult godot-source:**
- ✅ Implementing new DAP tools (see how Godot handles the command)
- ✅ Debugging protocol issues (understand Godot's expectations)
- ✅ Investigating quirks or unexpected behavior
- ✅ Validating command parameters and response formats
- ✅ Understanding event filtering patterns

**Example workflow:**
```bash
# Check if Godot implements a command before adding tool
mcp__godot-source__read_memory("dap_supported_commands")

# Understand event handling before implementing
mcp__godot-source__read_memory("dap_events_and_notifications")

# Find how Godot handles a specific DAP command
mcp__godot-source__find_symbol("DebugAdapterProtocol/handle_request")
```

**Tool naming clarity:**
- `mcp__source__*` → Navigate **our** Go DAP client code
- `mcp__godot-source__*` → Navigate **Godot's** C++ DAP server code
- `./godot-dap-mcp-server` → The compiled MCP server binary

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

**Current Phase**: Phase 3 (Core Debugging Tools)

The project is being built in phases:

- ✅ **Phase 1 (COMPLETE):** Core MCP Server - stdio server with test tool
  - All success criteria met (2025-11-06)
  - 16 unit tests passing
  - Verified working with Claude Code
  - See commit `8b1aa15`
- ✅ **Phase 2 (COMPLETE):** DAP Client Implementation - connection and protocol handling
  - All success criteria met (2025-11-06)
  - 12 unit tests passing (28 total)
  - Event filtering and timeout protection implemented
  - Godot-specific launch configs and session management
  - See commit TBD
- ⏳ **Phase 3 (IN PROGRESS):** Core Debugging Tools - 8 essential tools (connect, breakpoints, stepping)
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

Comprehensive testing documentation is available in **[docs/TESTING.md](docs/TESTING.md)**.

**Quick Reference**:
- Unit tests: `go test ./...`
- Integration tests: Require running Godot editor with DAP enabled
- Test fixture project: `tests/fixtures/test-project/`

See [TESTING.md](docs/TESTING.md) for detailed strategies, test examples, and CI setup.

## Godot DAP Protocol Details

Complete protocol documentation is available in **[docs/reference/DAP_PROTOCOL.md](docs/reference/DAP_PROTOCOL.md)**.

**Quick Reference**:
- Default port: 6006
- Connection sequence: `initialize` → `launch` → `configurationDone`
- Scene modes: `"main"` | `"current"` | `"res://path/to/scene.tscn"`
- Platform: `"host"` (default) | `"android"` | `"web"`

**Critical Finding**: stepOut is NOT implemented in Godot's DAP server. See [GODOT_SOURCE_ANALYSIS.md](docs/reference/GODOT_SOURCE_ANALYSIS.md).

For troubleshooting, see [GODOT_DAP_FAQ.md](docs/reference/GODOT_DAP_FAQ.md).

## Common Patterns

Complete implementation patterns are documented in **[docs/IMPLEMENTATION_GUIDE.md](docs/IMPLEMENTATION_GUIDE.md)**.

**Quick Reference**:
- Adding a new tool: See [IMPLEMENTATION_GUIDE.md - Tool Layer](docs/IMPLEMENTATION_GUIDE.md#tool-layer)
- Adding DAP functionality: See [IMPLEMENTATION_GUIDE.md - DAP Layer](docs/IMPLEMENTATION_GUIDE.md#dap-client-layer)
- Tool naming convention: `godot_<action>_<object>` (see [CONVENTIONS.md](docs/reference/CONVENTIONS.md))
- Error message pattern: Problem + Context + Solution (see [CONVENTIONS.md](docs/reference/CONVENTIONS.md#error-message-guidelines))

## Dependencies

Core dependencies:
- `github.com/google/go-dap`: DAP protocol implementation
- Standard library: `encoding/json`, `net`, `context`, `bufio`

Optional:
- `github.com/spf13/cobra`: CLI commands (if needed)
- `github.com/sirupsen/logrus`: Structured logging

## Distribution

Complete deployment documentation is available in **[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)**.

**Quick Reference**:
- Binary location: `build/`
- Supported platforms: macOS (Intel/ARM64), Linux, Windows
- Build script: `./scripts/build.sh`
- Installation: Download binary → Register with Claude Code

See [DEPLOYMENT.md](docs/DEPLOYMENT.md) for detailed build instructions, installation methods, and troubleshooting.

## Code Style

Complete coding conventions are documented in **[docs/reference/CONVENTIONS.md](docs/reference/CONVENTIONS.md)**.

**Quick Reference**:
- Follow standard Go conventions (gofmt, go vet)
- Tool naming: `godot_<action>_<object>`
- Error messages: Problem + Context + Solution pattern
- Use clear variable names (avoid abbreviations)
- Use context.Context for all potentially long-running operations

See [CONVENTIONS.md](docs/reference/CONVENTIONS.md) for detailed guidelines and examples.
