# Upstream Submission Progress

Track the status of issues and PRs submitted to godotengine/godot.

**Last Updated:** 2025-11-10

---

## Summary

| Status | Count |
|--------|-------|
| Issues Created | 0 |
| PRs Submitted | 0 |
| PRs Merged | 0 |
| Remaining | TBD |

---

## Batch 1: Critical Path (High Priority)

### Issue #1: req_initialize()
- **Status:** Not started
- **GitHub Issue:** N/A
- **PR:** N/A
- **Priority:** High
- **Files:** debug_adapter_parser.cpp
- **Lines Changed:** ~4

### Issue #2: req_launch()
- **Status:** Not started
- **GitHub Issue:** N/A
- **PR:** N/A
- **Priority:** High
- **Files:** debug_adapter_parser.cpp
- **Lines Changed:** ~8

### Issue #3: req_setBreakpoints()
- **Status:** Not started
- **GitHub Issue:** N/A
- **PR:** N/A
- **Priority:** High
- **Files:** debug_adapter_parser.cpp
- **Lines Changed:** ~3

---

## Batch 2: Debugging Workflow (Medium Priority)

### Issue #4: Inspection commands (stackTrace, scopes, variables)
- **Status:** Not started
- **GitHub Issue:** N/A
- **PR:** N/A
- **Priority:** Medium
- **Files:** debug_adapter_parser.cpp
- **Lines Changed:** ~6

### Issue #5: Execution control (continue, next, stepIn)
- **Status:** Not started
- **GitHub Issue:** N/A
- **PR:** N/A
- **Priority:** Medium
- **Files:** debug_adapter_parser.cpp
- **Lines Changed:** ~4

### Issue #6: evaluate
- **Status:** Not started
- **GitHub Issue:** N/A
- **PR:** N/A
- **Priority:** Medium
- **Files:** debug_adapter_parser.cpp
- **Lines Changed:** ~2

---

## Batch 3: Remaining Commands (Low Priority)

### Issue #7: Session management (restart, terminate, disconnect)
- **Status:** Not started
- **GitHub Issue:** N/A
- **PR:** N/A
- **Priority:** Low
- **Files:** debug_adapter_parser.cpp
- **Lines Changed:** ~5

### Issue #8: Response preparation (prepare_success_response, prepare_error_response)
- **Status:** Not started
- **GitHub Issue:** N/A
- **PR:** N/A
- **Priority:** Low
- **Files:** debug_adapter_parser.cpp
- **Lines Changed:** ~3

---

## Timeline

- **Week 1-2:** Test & Document (CURRENT)
- **Week 3:** Create Batch 1 Issues
- **Week 4-5:** Submit Batch 1 PRs
- **Week 6-8:** Submit Batch 2 PRs (after Batch 1 feedback)
- **Week 9-11:** Submit Batch 3 PRs (after Batch 2 feedback)

---

## Notes

- Wait for feedback on Batch 1 before submitting Batch 2
- Adjust strategy based on maintainer response
- Each PR must reference its GitHub issue
- Keep PRs small and focused (1-3 functions per PR)

---

## Links

- Godot Repository: https://github.com/godotengine/godot
- Our Fork: https://github.com/TransitionMatrix/godot
- Test Tool: https://github.com/<username>/godot-dap-mcp-server/cmd/test-dap-protocol
