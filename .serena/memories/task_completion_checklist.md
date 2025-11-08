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

# Integration tests (requires Godot 4.2.2+)
./scripts/automated-integration-test.sh  # Fully automated
./scripts/integration-test.sh            # Manual with running editor

# For new features, ensure:
# - Unit tests written for new code
# - Tests cover main success paths
# - Tests cover error conditions
# - Integration tests added if applicable
```

## 3. Documentation

**For all changes:**
- [ ] Update code comments for complex logic
- [ ] Update PLAN.md if phase status changed

**For phase completions with significant debugging:**
- [ ] Create `docs/LESSONS_LEARNED_PHASE_N.md` with debugging narrative
- [ ] Extract reusable patterns to `docs/IMPLEMENTATION_GUIDE.md`
- [ ] Add critical patterns to `docs/ARCHITECTURE.md`
- [ ] Add Q&A entries to `docs/reference/GODOT_DAP_FAQ.md`
- [ ] Add cross-references between documents

See `docs/DOCUMENTATION_WORKFLOW.md` for complete guidance on when and how to document lessons learned.

**For architecture/convention changes:**
- [ ] Update CLAUDE.md if workflow changed
- [ ] Update relevant docs/reference/ files

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

# Pre-commit hook runs:
# - gitleaks (secret detection)
# - Memory sync reminder (if internal/ or cmd/ changed)
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
- ConfigurationDone called after Initialize (critical!)
- Error messages follow Problem + Context + Solution pattern
- Session state properly managed

### For Phase Completions
If phase involved significant debugging (3+ issue/solution cycles):
1. Create phase-specific lessons learned document
2. Extract patterns to implementation guide
3. Update architecture doc with critical discoveries
4. Run `/memory-sync` to update Serena memories
5. Commit both code and documentation changes

### For Critical Changes
- Update docs/ARCHITECTURE.md if design patterns change
- Update docs/IMPLEMENTATION_GUIDE.md if component specs change
- Consider updating docs/reference/GODOT_DAP_FAQ.md for common issues
