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

## Documentation Access
```bash
# View CLAUDE.md for project guidance
cat CLAUDE.md

# View architecture documentation
cat docs/ARCHITECTURE.md

# View implementation guide
cat docs/IMPLEMENTATION_GUIDE.md

# View conventions
cat docs/reference/CONVENTIONS.md
```
