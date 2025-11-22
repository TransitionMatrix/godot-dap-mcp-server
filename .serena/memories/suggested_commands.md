# Suggested Commands for Development

## Building
```bash
# Development build
go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go

# Cross-platform builds
./scripts/build.sh
```

## Testing
```bash
# Run all tests
go test ./...

# Run specific package tests
go test -v ./internal/dap/...
go test -v ./internal/tools/...

# Run integration tests
./scripts/automated-integration-test.sh  # Fully automated
./scripts/integration-test.sh            # Manual (requires running editor)
```

## Debug/Test Utilities

**Launch & Protocol Verification:**
```bash
# Test launch sequence (Launch -> ConfigDone)
go run cmd/debug-launch/main.go

# Test minimal DAP handshake
go run cmd/test-minimal-dap/main.go

# Test full debugging workflow (17 steps)
go run cmd/test-full-debug-workflow/main.go
```

**DAP Inspection:**
```bash
# Inspect setBreakpoints serialization
go run cmd/dump-setbreakpoints/main.go
```

## Code Quality
```bash
# Format and Vet
go fmt ./... && go vet ./... && go test ./...
```

## Code Navigation (MCP Tools)
```bash
# Get symbols
mcp__source__get_symbols_overview("internal/dap/client.go")

# Find symbol
mcp__source__find_symbol(
    name_path="Client/Launch",
    relative_path="internal/dap/client.go"
)
```

## Git Operations
```bash
# Check status
git status

# Commit changes
git commit -m "feat: description"
```

## Documentation
```bash
# View plan
cat docs/PLAN.md

# View test results
cat docs/research/TEST_RESULTS_PROTOCOL_FIX.md
```
