# Suggested Commands for Development

## Building
```bash
# Development build (creates binary in project root)
go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go

# Cross-platform builds (creates binaries in build/ directory)
./scripts/build.sh

# Manual cross-compilation examples
GOOS=darwin GOARCH=arm64 go build -o build/godot-dap-mcp-server-darwin-arm64 cmd/godot-dap-mcp-server/main.go
GOOS=linux GOARCH=amd64 go build -o build/godot-dap-mcp-server-linux-amd64 cmd/godot-dap-mcp-server/main.go
GOOS=windows GOARCH=amd64 go build -o build/godot-dap-mcp-server-windows-amd64.exe cmd/godot-dap-mcp-server/main.go
```

## Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./internal/mcp/...
go test ./internal/dap/...
go test ./internal/tools/...

# Run integration tests (requires Godot editor with DAP enabled)
./scripts/automated-integration-test.sh  # Fully automated
./scripts/integration-test.sh            # Manual with running editor

# Test coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Code Quality
```bash
# Format code
go fmt ./...
gofmt -w .

# Static analysis
go vet ./...

# Run all quality checks
go fmt ./... && go vet ./... && go test ./...
```

## Code Navigation (MCP Tools)

Token-efficient navigation using `source` MCP server:

```bash
# Get file overview (Go files)
mcp__source__get_symbols_overview("internal/dap/client.go")

# Find specific symbol
mcp__source__find_symbol(
    name_path="Client/Connect",
    relative_path="internal/dap/client.go",
    include_body=true
)

# Navigate bash scripts
mcp__source__get_symbols_overview("scripts/automated-integration-test.sh")
mcp__source__find_symbol(
    name_path="send_mcp_request",
    relative_path="scripts/automated-integration-test.sh",
    include_body=true
)

# Search for patterns
mcp__source__search_for_pattern(
    substring_pattern="ConfigurationDone",
    paths_include_glob="**/*.go"
)

# Memory management
mcp__source__read_memory("critical_implementation_patterns")
mcp__source__write_memory("memory_name", "content")
mcp__source__list_memories()
```

**Configured languages** (`.serena/project.yml`):
- Go: Full symbolic navigation
- Bash: Function and variable indexing
- Markdown: Pattern search only

## Dependencies
```bash
# Install/update dependencies
go mod tidy

# Add new dependency
go get github.com/package/name

# Update dependency
go get -u github.com/package/name

# Verify dependencies
go mod verify

# View dependency graph
go mod graph
```

## Running the Server
```bash
# Run directly (development)
go run cmd/godot-dap-mcp-server/main.go

# Run compiled binary
./godot-dap-mcp-server

# Test stdio communication
./scripts/test-stdio.sh
```

## Debugging
```bash
# View Go documentation
go doc github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp
go doc github.com/TransitionMatrix/godot-dap-mcp-server/internal/dap

# Build with debug symbols
go build -gcflags="all=-N -l" -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go

# View build details
go version -m godot-dap-mcp-server
```

## Darwin (macOS) Specific Commands
```bash
# List files (BSD ls)
ls -la

# Find files (BSD find)
find . -name "*.go"

# Search in files (BSD grep)
grep -r "pattern" .

# View file
cat file.go

# Directory navigation
cd /path/to/directory
pwd
```

## Git Operations
```bash
# Status
git status

# Stage changes
git add .

# Commit
git commit -m "feat: descriptive message"

# Push
git push origin main

# View log
git log --oneline -10
```

## Claude Code Skills

Guided workflows via Claude Code skills:

```bash
# Memory sync - Update memories and docs after significant changes
/memory-sync

# Phase prep - Research Godot source before implementing a phase
/phase-prep

# MCP builder - Guidance for MCP server development
# (external skill, installed via .claude/scripts/setup-skills.sh)
```

**When to use skills:**
- `/memory-sync` - After phase completion, architecture changes, major refactoring, or doc reorganization
- `/phase-prep` - Before implementing Phase 5+ to discover Godot quirks upfront
- Memory sync and phase-prep are **project-specific skills** (checked into repo)

See `CLAUDE.md` for detailed skill documentation.

## Documentation Access

**Core documentation** (top-level docs/):
```bash
cat docs/PLAN.md                # Implementation phases and status
cat docs/ARCHITECTURE.md         # System design and patterns
cat docs/IMPLEMENTATION_GUIDE.md # Component specifications
cat docs/TESTING.md             # Testing strategies
cat docs/DEPLOYMENT.md          # Build and distribution
cat docs/DOCUMENTATION_WORKFLOW.md # Hybrid docs approach
```

**Reference documentation** (docs/reference/):
```bash
# DAP Protocol
cat docs/reference/DAP_SESSION_GUIDE.md # Complete DAP command reference + session flow
cat docs/reference/DAP_PROTOCOL.md      # Godot DAP details (launch-specific)
cat docs/reference/GODOT_DAP_FAQ.md     # Common questions and troubleshooting

# Other references
cat docs/reference/CONVENTIONS.md            # Coding standards
cat docs/reference/GODOT_SOURCE_ANALYSIS.md  # Godot source findings

# Query DAP specification
jq '.definitions.InitializeRequest' docs/reference/debugAdapterProtocol.json
```

**Godot upstream submission** (docs/godot-upstream/):
```bash
cat docs/godot-upstream/STRATEGY.md        # Multi-PR submission approach
cat docs/godot-upstream/TESTING_GUIDE.md   # Testing Dictionary safety issues
cat docs/godot-upstream/PROGRESS.md        # Track submission status
cat docs/godot-upstream/ISSUE_TEMPLATE.md  # GitHub issue template
cat docs/godot-upstream/PR_TEMPLATE.md     # Pull request template
```

**Implementation notes** (docs/implementation-notes/):
```bash
cat docs/implementation-notes/LESSONS_LEARNED_PHASE_3.md
cat docs/implementation-notes/PHASE_6_IMPLEMENTATION_NOTES.md
```

**CLAUDE.md** (project guidance for AI assistants):
```bash
cat CLAUDE.md  # Start here for project overview and guidance
```

## Godot Upstream Testing

Test Dictionary safety issues against Godot builds:

**Automated testing** (recommended):
```bash
# Set Godot binary path
export GODOT_BIN=/Users/adp/Projects/godot/bin/godot.macos.editor.arm64

# Run automated compliance test
./scripts/test-dap-compliance.sh

# With custom options
DAP_PORT=6007 PROJECT_PATH=path/to/project.godot ./scripts/test-dap-compliance.sh

# Test against upstream Godot
export GODOT_BIN=/Users/adp/Projects/godot-upstream/bin/godot.macos.editor.arm64
./scripts/test-dap-compliance.sh
```

The automated script:
- Starts Godot in the background with DAP enabled
- Runs test-dap-protocol in automated mode (no manual input)
- Captures output to timestamped file
- Analyzes Godot console for Dictionary errors
- Cleans up automatically
- Returns exit code 0 if no errors, 1 if errors found

**Manual testing**:
```bash
# Start Godot editor manually
/Users/adp/Projects/godot-upstream/bin/godot.macos.editor.arm64

# In another terminal: Run test tool (interactive mode)
cd /Users/adp/Projects/godot-dap-mcp-server
go run cmd/test-dap-protocol/main.go

# Observe Godot console for Dictionary errors
```

**Output locations**:
- Test output: `/tmp/godot-dap-test-output-YYYYMMDD-HHMMSS.txt`
- Godot console: `/tmp/godot-console-<PID>.log`

See `docs/godot-upstream/TESTING_GUIDE.md` for comprehensive testing strategy.
