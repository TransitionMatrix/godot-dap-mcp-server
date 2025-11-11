# Godot DAP Dictionary Safety - PR Strategy

**Last Updated**: 2025-11-10

This document outlines our strategy for submitting Dictionary safety fixes to the Godot engine repository as a series of small, focused PRs rather than one large PR.

---

## Why Function-by-Function PRs?

### Advantages Over Single Large PR

**1. Easier to Review & Accept**
- Small, focused changes (1-3 functions per PR vs 31 fixes in one PR)
- Each PR has clear before/after demonstration
- Maintainers can review in 5-10 minutes per PR
- Lower risk = faster acceptance
- Easier to rollback if issues found

**2. Better Documentation**
Each GitHub issue becomes a clear, focused bug report:
```
Title: Dictionary error in req_initialize() when optional field omitted

Steps to reproduce:
1. Run test-dap-protocol tool
2. Send initialize with minimal required fields per DAP spec
3. Observe Dictionary error in Godot console

Expected: No errors (optional fields should be... optional)
Actual: ERROR: Dictionary::operator[] used when there was no value

Fix: Use .get("fieldName", default) for safe access
```

**3. Prioritization**
Focus on high-impact commands first:
- **High priority**: Core workflow commands (initialize, launch, setBreakpoints)
- **Medium priority**: Debugging commands (stackTrace, variables, continue)
- **Low priority**: Edge cases and rarely-used commands

**4. Test-Driven Approach**
Our `cmd/test-dap-protocol/` tool is perfect for this:
- ✅ Sends spec-compliant messages
- ✅ Demonstrates bugs clearly with interactive output
- ✅ Shows Dictionary errors in real-time
- ✅ Can be run by Godot maintainers to verify
- ✅ Provides before/after proof
- ✅ Documents DAP spec requirements vs Godot's expectations

**5. Incremental Progress**
- Can submit 2-3 PRs at a time (not spammy)
- Build trust as early PRs get accepted
- Learn from feedback on first batch
- Can adjust strategy based on maintainer response
- Can pause if maintainers push back

**6. Risk Mitigation**
- Each PR can be tested independently
- Bugs in one fix don't block others
- Maintainers comfortable with small, auditable changes
- Easier to bisect if regressions occur

---

## Three-Phase Implementation

### Phase 1: Test & Document (1-2 days)

For each DAP command:

**1. Test Against Upstream**
```bash
# Start godot-upstream editor with DAP enabled
/Users/adp/Projects/godot-upstream/bin/godot.macos.editor.arm64

# Run test tool
cd /Users/adp/Projects/godot-dap-mcp-server
go run cmd/test-dap-protocol/main.go
```

**2. Document Issues**
For each command that triggers Dictionary errors:
- Screenshot of test output
- Copy console error messages
- Note which fields trigger errors
- Check DAP spec for field requirements
- Identify `.get()` calls needed

**3. Create Issue Notes**
Template for each issue:
```markdown
## Command: initialize

**DAP Spec**: debugAdapterProtocol.json
- Required: command, arguments, arguments.adapterID
- Optional: arguments.linesStartAt1, arguments.columnsStartAt1, etc.

**Test Result**:
- Sent minimal required fields
- Dictionary errors on: linesStartAt1, columnsStartAt1, supportsVariableType, supportsInvalidatedEvent

**Files Affected**:
- editor/debugger/debug_adapter/debug_adapter_parser.cpp:133-136

**Fixes Needed**:
```cpp
// Line 133
peer->linesStartAt1 = args.get("linesStartAt1", false);
// Line 134
peer->columnsStartAt1 = args.get("columnsStartAt1", false);
// Line 135
peer->supportsVariableType = args.get("supportsVariableType", false);
// Line 136
peer->supportsInvalidatedEvent = args.get("supportsInvalidatedEvent", false);
```

**Impact**: Medium - Works but logs errors for spec-compliant clients
```

### Phase 2: Create GitHub Issues (1 day)

**Issue Organization**:

Create 10-15 issues, grouped by command or functionality:

**Group 1: Connection & Setup (High Priority)**
- Issue #1: Dictionary errors in `initialize` - unsafe access to optional client capabilities
- Issue #2: Dictionary errors in `launch` - unsafe access to optional launch parameters
- Issue #3: Missing response in `launch` - violates DAP spec requirement

**Group 2: Breakpoints (High Priority)**
- Issue #4: Dictionary errors in `setBreakpoints` - unsafe access to optional breakpoints array

**Group 3: Execution Control (Medium Priority)**
- Issue #5: Dictionary errors in `continue` - unsafe access to optional fields
- Issue #6: Dictionary errors in `next` - unsafe access to optional threadId
- Issue #7: Dictionary errors in `stepIn` - unsafe access to optional fields

**Group 4: Inspection (Medium Priority)**
- Issue #8: Dictionary errors in `stackTrace` - unsafe access to optional startFrame/levels
- Issue #9: Dictionary errors in `scopes` - unsafe access to optional fields
- Issue #10: Dictionary errors in `variables` - unsafe access to optional filter

**Group 5: Evaluation (Medium Priority)**
- Issue #11: Dictionary errors in `evaluate` - unsafe access to optional context

**Group 6: Advanced Features (Low Priority)**
- Issue #12: Dictionary errors in `restart` - unsafe nested arguments access
- Issue #13: Dictionary errors in `terminate` - unsafe access to optional fields
- Issue #14: Dictionary errors in `disconnect` - unsafe access to optional fields

**Issue Template**:
```markdown
## Dictionary Error in req_initialize() with Spec-Compliant Request

**Description**
The `req_initialize()` function in `debug_adapter_parser.cpp` uses unsafe
Dictionary access (`operator[]`) for optional fields. When a DAP client sends
a valid request that omits optional fields, Godot logs Dictionary errors.

**Steps to Reproduce**
1. Enable DAP server in Godot: Editor → Editor Settings → Network → Debug Adapter
2. Send this spec-compliant initialize request:
   ```json
   {
     "seq": 1,
     "type": "request",
     "command": "initialize",
     "arguments": {
       "adapterID": "test-client"
     }
   }
   ```
3. Observe console output

**Expected Behavior**
No errors. Per the DAP specification, fields like `linesStartAt1`, `columnsStartAt1`,
etc. are optional. The server should use default values when they're omitted.

**Actual Behavior**
```
ERROR: Bug: Dictionary::operator[] used when there was no value for the given key
   at: operator[] (core/variant/dictionary.cpp:136)
```

**DAP Specification**
Per the [Debug Adapter Protocol specification](https://microsoft.github.io/debug-adapter-protocol/specification):
- `arguments.linesStartAt1` is optional (defaults to true)
- `arguments.columnsStartAt1` is optional (defaults to true)
- `arguments.supportsVariableType` is optional (defaults to false)
- `arguments.supportsInvalidatedEvent` is optional (defaults to false)

**Root Cause**
File: `editor/debugger/debug_adapter/debug_adapter_parser.cpp`, lines 133-136

```cpp
peer->linesStartAt1 = args["linesStartAt1"];  // Unsafe!
peer->columnsStartAt1 = args["columnsStartAt1"];  // Unsafe!
peer->supportsVariableType = args["supportsVariableType"];  // Unsafe!
peer->supportsInvalidatedEvent = args["supportsInvalidatedEvent"];  // Unsafe!
```

**Proposed Fix**
Use Dictionary's safe `.get()` method with appropriate defaults:

```cpp
peer->linesStartAt1 = args.get("linesStartAt1", false);
peer->columnsStartAt1 = args.get("columnsStartAt1", false);
peer->supportsVariableType = args.get("supportsVariableType", false);
peer->supportsInvalidatedEvent = args.get("supportsInvalidatedEvent", false);
```

**Impact**
- Severity: Low (functionality works, but logs errors)
- Frequency: Every time a strict DAP client connects
- User Experience: Console spam, confusion about errors

**Test Tool**
A test program demonstrating this issue is available at:
https://github.com/user/godot-dap-mcp-server/tree/main/cmd/test-dap-protocol

**Godot Version**
- Tested on: 4.4-dev (commit 2d86b69bf1)
- Affects: All versions since DAP implementation
```

### Phase 3: Submit PRs in Batches

**Batch 1: Critical Path (3 PRs)**
Start with the most important commands for basic debugging workflow:

**PR #1: Fix Dictionary safety in req_initialize()**
- Fixes: Issue #1
- Files: 1 file, ~4 lines changed
- Priority: High
- Justification: First command every client sends

**PR #2: Fix Dictionary safety and missing response in req_launch()**
- Fixes: Issue #2, Issue #3
- Files: 1 file, ~8 lines changed
- Priority: High
- Justification: Required to start debugging session

**PR #3: Fix Dictionary safety in req_setBreakpoints()**
- Fixes: Issue #4
- Files: 1 file, ~3 lines changed
- Priority: High
- Justification: Core debugging feature

**Wait for feedback on Batch 1 before proceeding**

**Batch 2: Debugging Workflow (3-4 PRs)**
After Batch 1 accepted:

**PR #4: Fix Dictionary safety in inspection commands**
- Fixes: Issue #8, Issue #9, Issue #10
- Files: 1 file, ~6 lines changed
- Commands: stackTrace, scopes, variables

**PR #5: Fix Dictionary safety in execution control**
- Fixes: Issue #5, Issue #6, Issue #7
- Files: 1 file, ~4 lines changed
- Commands: continue, next, stepIn

**PR #6: Fix Dictionary safety in evaluate**
- Fixes: Issue #11
- Files: 1 file, ~2 lines changed
- Commands: evaluate

**Batch 3: Remaining Commands (2-3 PRs)**
After Batch 2 accepted:

**PR #7: Fix Dictionary safety in session management**
- Fixes: Issue #12, Issue #13, Issue #14
- Files: 1 file, ~5 lines changed
- Commands: restart, terminate, disconnect

**PR #8: Fix Dictionary safety in response preparation**
- Fixes defensive programming issues
- Files: 1 file, ~3 lines changed
- Functions: prepare_success_response, prepare_error_response

---

## PR Template

Each PR should follow this structure:

```markdown
## Fix Dictionary safety in req_initialize()

Fixes godotengine/godot#[issue-number]

### Problem

The `req_initialize()` function uses unsafe Dictionary access (`operator[]`)
for optional fields defined in the Debug Adapter Protocol specification. When
a DAP client sends a spec-compliant request omitting optional fields, Godot
logs Dictionary errors to the console.

### Evidence

**Test tool**: https://github.com/user/godot-dap-mcp-server/tree/main/cmd/test-dap-protocol

Running the test tool against current Godot master shows:
```
ERROR: Bug: Dictionary::operator[] used when there was no value for the given key
   at: operator[] (core/variant/dictionary.cpp:136)
```

**DAP Specification**: Per [Microsoft's DAP spec](https://microsoft.github.io/debug-adapter-protocol/specification),
the following fields are optional in InitializeRequestArguments:
- `linesStartAt1` (optional, defaults to true)
- `columnsStartAt1` (optional, defaults to true)
- `supportsVariableType` (optional, defaults to false)
- `supportsInvalidatedEvent` (optional, defaults to false)

### Solution

Replace unsafe `operator[]` with safe `.get()` method using appropriate
default values per the DAP specification.

**Before**:
```cpp
peer->linesStartAt1 = args["linesStartAt1"];
peer->columnsStartAt1 = args["columnsStartAt1"];
peer->supportsVariableType = args["supportsVariableType"];
peer->supportsInvalidatedEvent = args["supportsInvalidatedEvent"];
```

**After**:
```cpp
peer->linesStartAt1 = args.get("linesStartAt1", false);
peer->columnsStartAt1 = args.get("columnsStartAt1", false);
peer->supportsVariableType = args.get("supportsVariableType", false);
peer->supportsInvalidatedEvent = args.get("supportsInvalidatedEvent", false);
```

### Testing

**Before (upstream master)**:
```bash
# Start Godot with DAP enabled
./bin/godot.macos.editor.arm64

# Run test tool (sends spec-compliant messages)
go run cmd/test-dap-protocol/main.go

# Result: Dictionary errors in Godot console
```

**After (with this PR)**:
```bash
# Same test
go run cmd/test-dap-protocol/main.go

# Result: No errors, clean execution
```

### Impact

- **Severity**: Low (functionality works but logs errors)
- **Scope**: Affects all DAP clients that send minimal requests
- **Risk**: Very low - adds safe defaults, no behavior change for clients
  that send all fields

### Files Changed

- `editor/debugger/debug_adapter/debug_adapter_parser.cpp` - 4 lines
```

---

## Timeline Estimate

### Optimistic (4-6 weeks)
- Week 1: Test all commands, document issues
- Week 2: Create GitHub issues
- Week 3-4: Submit Batch 1, get feedback, iterate
- Week 5: Submit Batch 2
- Week 6: Submit Batch 3

### Realistic (8-12 weeks)
- Week 1-2: Test, document, create issues
- Week 3-5: Batch 1 (3 PRs) - review cycles
- Week 6-8: Batch 2 (3-4 PRs) - review cycles
- Week 9-11: Batch 3 (2-3 PRs) - review cycles
- Week 12: Final reviews and merges

### Conservative (12-16 weeks)
- Account for longer review cycles
- Maintainer feedback requiring changes
- Possible pushback requiring discussion
- Holiday/vacation delays

---

## Risk Mitigation

### Potential Challenges

**1. Maintainers prefer single PR**
- Response: Explain review burden, testing complexity
- Offer to combine if they insist
- Point to successful multi-PR patterns in other projects

**2. Disagreement on defaults**
- Response: Cite DAP specification
- Show behavior of other DAP implementations (VS Code, etc.)
- Willing to adjust if maintainers have good reasons

**3. "It works fine, why change it?"**
- Response:
  - Spec compliance matters for ecosystem
  - Console spam confuses users
  - Defensive programming prevents future bugs
  - Minimal risk, clear benefit

**4. Request for comprehensive tests**
- Response: Provide test-dap-protocol tool
- Offer to add unit tests if needed
- Point out existing testing via real usage

### Contingency Plans

**If maintainers push back on approach:**
- Option A: Group into 3-4 larger PRs instead
- Option B: Submit all as one PR with clear sections
- Option C: Focus on highest-priority commands only

**If PRs sit idle:**
- Politely ping after 2 weeks
- Offer to address any concerns
- Ask if different approach preferred

**If first PR rejected:**
- Understand reasoning
- Adjust strategy for remaining PRs
- Consider if worth continuing

---

## Success Metrics

### Quantitative
- ✅ 8-10 PRs merged
- ✅ 90%+ of Dictionary errors eliminated
- ✅ Zero regressions in DAP functionality

### Qualitative
- ✅ Positive maintainer feedback
- ✅ Other contributors learn safe Dictionary patterns
- ✅ DAP client developers report better experience
- ✅ Godot console cleaner for users

---

## Communication Strategy

### GitHub Issues
- Clear, concise titles
- Include reproduction steps
- Cite DAP specification
- Professional, factual tone
- Offer to submit PR

### Pull Requests
- Reference issue number
- Include before/after evidence
- Explain rationale briefly
- Show low risk
- Thank maintainers for review

### Community
- Don't spam Discord/forums
- If asked, point to issues/PRs
- Be patient with review process
- Accept feedback gracefully

---

## Resources

### Test Tools
- `cmd/test-dap-protocol/` - Interactive DAP protocol compliance tester
- Shows spec requirements vs Godot's expectations
- Demonstrates Dictionary errors in real-time

### Documentation
- `docs/reference/debugAdapterProtocol.json` - Official DAP specification
- `docs/DAP_REQUIRED_VS_OPTIONAL_FIELDS.md` - Field requirement analysis
- `docs/TESTING_UNSAFE_DICTIONARY_CODE.md` - Testing methodology

### Godot Resources
- [Contributing Guide](https://docs.godotengine.org/en/stable/community/contributing/ways_to_contribute.html)
- [PR Guidelines](https://docs.godotengine.org/en/stable/community/contributing/pr_workflow.html)
- [Code Style Guide](https://docs.godotengine.org/en/stable/community/contributing/code_style_guidelines.html)

---

## Next Steps

1. **Test all commands** - Run test-dap-protocol against upstream, document all Dictionary errors
2. **Create spreadsheet** - Track commands, issues, fixes, priorities
3. **Draft first 3 issues** - Start with initialize, launch, setBreakpoints
4. **Get feedback** - Share approach in Godot contributor channels (if appropriate)
5. **Submit Batch 1** - Create first 3 PRs after issues are filed

---

## Notes

- This is a marathon, not a sprint
- Quality over speed
- Build trust with maintainers
- Be responsive to feedback
- Celebrate small wins
- Each merged PR is progress!
