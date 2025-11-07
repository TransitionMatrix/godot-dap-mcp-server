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

# Run integration tests (requires running Godot editor with DAP enabled)
go test ./tests/integration/...

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
