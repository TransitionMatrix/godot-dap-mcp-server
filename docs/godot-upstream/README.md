# Godot Upstream Submission Materials

This directory contains all materials for submitting Dictionary safety fixes to the upstream Godot engine repository.

## Active Strategy

**[STRATEGY.md](STRATEGY.md)** - Multi-PR submission strategy (function-by-function approach)
- Why we're doing small PRs instead of one large PR
- Three-phase implementation (Test → Issues → PRs)
- Batch organization and timeline
- Risk mitigation and success metrics

## Guides

**[SUBMISSION_GUIDE.md](SUBMISSION_GUIDE.md)** - Step-by-step submission process
- How to fork and set up Godot repository
- Building and testing Godot
- Creating pull requests
- Following Godot's PR workflow

**[TESTING_GUIDE.md](TESTING_GUIDE.md)** - How to test Dictionary safety issues
- Using `cmd/test-dap-protocol/` to demonstrate bugs
- Test cases that expose unsafe Dictionary access
- Comparison of upstream vs fixed behavior

## Templates

**[ISSUE_TEMPLATE.md](ISSUE_TEMPLATE.md)** - Template for GitHub issues
**[PR_TEMPLATE.md](PR_TEMPLATE.md)** - Template for pull requests

## Progress Tracking

**[PROGRESS.md](PROGRESS.md)** - Track which issues/PRs have been created and their status

---

**Next Steps:**
1. Run test-dap-protocol against godot-upstream
2. Document all Dictionary errors
3. Create GitHub issues using ISSUE_TEMPLATE.md
4. Submit PRs in batches following STRATEGY.md
