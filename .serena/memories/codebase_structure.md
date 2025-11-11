# Codebase Structure

## Repository Layout
```
godot-dap-mcp-server/
├── cmd/
│   └── godot-dap-mcp-server/
│       └── main.go              # Entry point
├── internal/
│   ├── mcp/                     # MCP Layer - stdio protocol
│   │   ├── server.go           # Core MCP server, request routing
│   │   ├── types.go            # JSONRPC 2.0 request/response types
│   │   ├── transport.go        # stdin/stdout message handling
│   │   ├── server_test.go      # MCP server tests (incl. schema validation)
│   │   └── transport_test.go   # Transport tests
│   ├── dap/                     # DAP Client Layer - TCP to Godot
│   │   ├── client.go           # DAP protocol implementation
│   │   ├── session.go          # Session lifecycle management
│   │   ├── events.go           # Async event filtering
│   │   ├── timeout.go          # Timeout wrappers (prevents hangs)
│   │   ├── godot.go            # Godot-specific DAP extensions
│   │   └── dap_test.go         # DAP client tests
│   └── tools/                   # Tool Layer - Godot MCP tools (15 files)
│       ├── registry.go         # Tool registration
│       ├── session.go          # Global session management
│       ├── formatting.go       # Godot type formatting
│       ├── ping.go             # Test tool
│       ├── connect.go          # Phase 3: Connection tools
│       ├── breakpoints.go      # Phase 3: Breakpoint tools
│       ├── execution.go        # Phase 3: Execution control
│       ├── inspection.go       # Phase 4: Runtime inspection
│       ├── advanced.go         # Phase 6: Advanced tools (pause, set_variable)
│       └── *_test.go           # Tool tests (7 test files)
├── tests/
│   └── fixtures/
│       └── test-project/       # Godot test project for integration tests
├── docs/                        # Documentation (organized by purpose)
│   ├── PLAN.md                  # Implementation phases and status
│   ├── ARCHITECTURE.md          # System design and patterns
│   ├── IMPLEMENTATION_GUIDE.md  # Component specifications
│   ├── TESTING.md              # Testing strategies
│   ├── DEPLOYMENT.md           # Build and distribution
│   ├── DOCUMENTATION_WORKFLOW.md  # Hybrid docs approach
│   ├── godot-upstream/          # Upstream submission materials
│   │   ├── STRATEGY.md         # Multi-PR submission strategy
│   │   ├── SUBMISSION_GUIDE.md # Step-by-step submission process
│   │   ├── TESTING_GUIDE.md    # Test Dictionary safety issues
│   │   ├── ISSUE_TEMPLATE.md   # GitHub issue template
│   │   ├── PR_TEMPLATE.md      # Pull request template
│   │   ├── PROGRESS.md         # Track submission status
│   │   └── README.md
│   ├── reference/               # Reference documentation
│   │   ├── CONVENTIONS.md      # Coding standards
│   │   ├── DAP_PROTOCOL.md     # Godot DAP details
│   │   ├── GODOT_DAP_FAQ.md    # Troubleshooting
│   │   ├── GODOT_SOURCE_ANALYSIS.md  # Godot source findings
│   │   ├── POPULAR_GODOT_DAP_CLIENTS.md # DAP client survey
│   │   └── debugAdapterProtocol.json    # Official DAP spec (178KB)
│   ├── implementation-notes/    # Phase notes and lessons
│   │   ├── PHASE_6_IMPLEMENTATION_NOTES.md
│   │   ├── LAUNCH_RESPONSE_FIX_EXPLANATION.md
│   │   ├── LESSONS_LEARNED_DAP_DEBUGGING.md
│   │   ├── LESSONS_LEARNED_PHASE_3.md
│   │   ├── LESSONS_LEARNED_PHASE_6.md
│   │   └── README.md
│   ├── research/                # Research and analysis archive
│   │   └── (Dictionary safety audits, test results, analysis)
│   └── archive/                 # Superseded documents
│       └── (Old PR strategies, drafts)
├── scripts/
│   ├── automated-integration-test.sh  # Automated integration test
│   └── build.sh                # Cross-platform build script
├── .claude/                     # Claude Code configuration
│   ├── commands/               # Slash commands
│   └── skills/                 # Claude Code skills (phase-prep, memory-sync)
├── .serena/                     # Serena MCP configuration
│   ├── memories/               # Project memories (7 files)
│   └── project.yml             # Serena config
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── CLAUDE.md                    # Claude Code guidance
└── README.md                    # Project readme

## Documentation Organization (2025-11-10 Update)

Documentation is now organized by purpose:

**Top-level** (`docs/`) - Active development documentation
- Core docs: PLAN, ARCHITECTURE, IMPLEMENTATION_GUIDE, TESTING, DEPLOYMENT
- DOCUMENTATION_WORKFLOW.md - How to use the hybrid approach

**godot-upstream/** - Upstream submission materials (active work)
- STRATEGY.md - Multi-PR submission approach
- Templates (ISSUE_TEMPLATE.md, PR_TEMPLATE.md)
- PROGRESS.md - Track submission status

**reference/** - Stable reference materials
- Protocol details, conventions, FAQ
- debugAdapterProtocol.json - Official DAP specification

**implementation-notes/** - Phase notes and lessons learned
- PHASE_N_IMPLEMENTATION_NOTES.md - Pre-implementation research
- LESSONS_LEARNED_PHASE_N.md - Post-implementation insights

**research/** - Research and analysis (historical reference)
- Dictionary safety audits and test results
- DAP protocol analysis documents

**archive/** - Superseded documents (preserved for history)
- Old PR strategies, early drafts

## Layer Architecture

### Layer 1: MCP Layer (`internal/mcp/`)
**Purpose**: Handle stdio-based MCP protocol communication
- Reads JSONRPC 2.0 requests from stdin
- Routes to appropriate tool handlers
- Writes JSONRPC 2.0 responses to stdout
- Manages tool registry

**Key Files**:
- `server.go`: Request handling, tool routing, response formatting
- `types.go`: Request, Response, Tool, Parameter, PropertyDefinition types
- `transport.go`: stdin/stdout reading/writing

**Phase 6 Addition**: `types.go` now has `omitempty` on PropertyDefinition.Type for valid JSON Schema generation

### Layer 2: DAP Client Layer (`internal/dap/`)
**Purpose**: Communicate with Godot editor's DAP server via TCP
- Connects to Godot DAP server (localhost:6006 default)
- Implements DAP protocol using `github.com/google/go-dap`
- Filters async events from responses (critical pattern)
- Provides timeout protection for all operations
- Manages single debugging session lifecycle

**Key Files**:
- `client.go`: DAP protocol implementation, request/response
- `session.go`: Session state, connection management
- `events.go`: Event filtering (separates events from responses)
- `timeout.go`: Timeout wrappers for all DAP calls
- `godot.go`: Godot-specific launch parameters, scene modes

**Phase 6 Addition**: `client.go` now has `Pause()` method for pausing running game

### Layer 3: Tool Layer (`internal/tools/`)
**Purpose**: Implement Godot-specific MCP tools
- Follows naming convention: `godot_<action>_<object>`
- Translates MCP tool calls to DAP operations
- Provides AI-optimized tool descriptions
- Handles parameter validation and security

**Key Files**:
- `registry.go`: Tool registration system
- `session.go`: Global session management via `GetSession()`
- `formatting.go`: Godot type formatting (Vector2/3, Color, Nodes)
- `connect.go`: Phase 3 connection tools (connect, disconnect)
- `breakpoints.go`: Phase 3 breakpoint tools (set, clear)
- `execution.go`: Phase 3 execution control (continue, step-over, step-in)
- `inspection.go`: Phase 4 runtime inspection (threads, stack, scopes, variables, evaluate)
- `advanced.go`: Phase 6 advanced tools (pause, set_variable with security validation)

**Tool Count**: 14 tools across 3 phases (7 Phase 3, 5 Phase 4, 2 Phase 6)

## Module Structure
- **Package**: `github.com/TransitionMatrix/godot-dap-mcp-server`
- **Internal packages**: Cannot be imported by external projects
- **Go version**: 1.25.3
- **Single dependency**: `github.com/google/go-dap v0.12.0`

## File Naming Conventions
- Lowercase with underscores: `stack_trace.go`
- Test files: `<name>_test.go`
- Group related functionality in same file
- Keep files focused on single responsibility

## Test Organization
- Unit tests alongside source: `internal/*/`
- Integration tests: Automated via `scripts/automated-integration-test.sh`
- Test fixture: `tests/fixtures/test-project/` (Godot project)
- Coverage target: >70% for critical paths
- **Total tests**: 61 (28 MCP, 0 DAP, 33 tools)
