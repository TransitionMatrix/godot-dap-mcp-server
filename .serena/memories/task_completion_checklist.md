# Task Completion Checklist

When a coding task is completed, perform the following checks:

## 1. Code Quality
```bash
# Format code
go fmt ./...

# Static analysis
go vet ./...
```

## 2. Testing
```bash
# Run all tests
go test ./...

# For new features, ensure:
# - Unit tests written for new code
# - Tests cover main success paths
# - Tests cover error conditions
# - Integration tests added if applicable
```

## 3. Documentation
- [ ] Update CLAUDE.md if architecture changed
- [ ] Update relevant docs/reference/ files if needed
- [ ] Add/update code comments for complex logic
- [ ] Update PLAN.md if phase status changed

## 4. Conventions Check
- [ ] Tool names follow `godot_<action>_<object>` pattern
- [ ] Error messages follow Problem + Context + Solution pattern
- [ ] Variable/function names are clear and descriptive
- [ ] Comments explain "why", not "what"

## 5. Build Verification
```bash
# Verify build succeeds
go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go

# Optional: Test cross-compilation
./scripts/build.sh
```

## 6. Git Operations (if committing)
```bash
# Check status
git status

# Stage relevant changes
git add <files>

# Commit with conventional format
git commit -m "<type>: <subject>"
# Types: feat, fix, docs, test, refactor, perf, chore

# Push if appropriate
git push origin main
```

## Quick Command Sequence
```bash
# Complete quality check
go fmt ./... && go vet ./... && go test ./... && go build -o godot-dap-mcp-server cmd/godot-dap-mcp-server/main.go
```

## Special Considerations

### For MCP Tools
- Tool description is clear and AI-optimized
- Prerequisites listed
- Examples provided
- Parameters documented with types and validation

### For DAP Client Code
- Timeout protection added (use context.Context)
- Event filtering implemented (async events mixed with responses)
- Error messages follow Problem + Context + Solution pattern
- Session state properly managed

### For Critical Changes
- Update docs/ARCHITECTURE.md if design patterns change
- Update docs/IMPLEMENTATION_GUIDE.md if component specs change
- Consider updating docs/reference/GODOT_DAP_FAQ.md for common issues
