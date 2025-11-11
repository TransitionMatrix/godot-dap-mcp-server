# Godot PR Submission Guide - Option A: Full Compliance

This guide walks through submitting the DAP fixes PR with full compliance to Godot contributing guidelines (unit tests + GitHub issue).

## Overview

**Status**: ✅ All preparation complete
**Branch**: `fix/dap-dictionary-safety` (clean, 2 commits)
**Ready for**: GitHub issue creation → Unit test integration → PR submission

---

## Step 1: Create GitHub Issue

### 1.1 Navigate to Godot Issues
- Go to: https://github.com/godotengine/godot/issues
- Click "New Issue"
- Use "Bug Report" template if available

### 1.2 Fill in Issue Details

**Source**: Use content from `docs/GODOT_GITHUB_ISSUE.md`

**Title**:
```
[Debugger] DAP implementation has unsafe Dictionary access patterns causing crashes and launch timeouts
```

**Labels** (if you can set them):
- `bug`
- `debugger`
- `priority:high`

**Copy entire description** from `GODOT_GITHUB_ISSUE.md`, which includes:
- Summary of both bugs
- Error messages and symptoms
- Root causes
- Steps to reproduce
- Testing evidence
- Proposed solution with code examples
- Related issues (#110749, #106969)

### 1.3 Submit and Note Issue Number
- Click "Submit new issue"
- **Record the issue number** (e.g., #12345)
- You'll need this for Step 2 and Step 3

---

## Step 2: Add Unit Tests to Godot

### 2.1 Copy Test File

In your Godot repository:

```bash
cd /Users/adp/Projects/godot

# Copy the test file
cp /Users/adp/Projects/godot-dap-mcp-server/docs/GODOT_UNIT_TEST.h \
   tests/core/variant/test_debug_adapter_dictionary.h
```

### 2.2 Register Test in test_main.cpp

Edit `tests/test_main.cpp` around line 120 (near other variant tests):

```cpp
#include "tests/core/variant/test_array.h"
#include "tests/core/variant/test_callable.h"
#include "tests/core/variant/test_debug_adapter_dictionary.h"  // ADD THIS
#include "tests/core/variant/test_dictionary.h"
#include "tests/core/variant/test_variant.h"
```

### 2.3 Build and Test

```bash
# Build with tests enabled
scons tests=yes

# Run all tests to ensure nothing broke
./bin/godot.macos.editor.dev.arm64.tests --test

# Run just the new tests
./bin/godot.macos.editor.dev.arm64.tests --test --test-case="[DebugAdapter]*"
```

**Expected Output**:
```
[doctest] test cases:  5 |  5 passed | 0 failed | 0 skipped
[doctest] assertions: 12 | 12 passed | 0 failed |
[doctest] Status: SUCCESS!
```

### 2.4 Commit the Test

```bash
git add tests/core/variant/test_debug_adapter_dictionary.h
git add tests/test_main.cpp

git commit -m "Add unit tests for DAP Dictionary safety

Tests demonstrate safe Dictionary.get() usage vs unsafe operator[].
Related to issue #XXXXX."
```

**Important**: Replace `#XXXXX` with the actual issue number from Step 1.

---

## Step 3: Update PR Description

### 3.1 Replace Placeholder Issue Numbers

Edit `docs/GODOT_PR_DESCRIPTION.md` and replace ALL instances of `#XXXXX` with your actual issue number:

```bash
# Example if issue number is #12345
sed -i '' 's/#XXXXX/#12345/g' docs/GODOT_PR_DESCRIPTION.md
```

Or manually edit these locations:
- Line 3: `Fixes #XXXXX`
- Line 134: `Fixes #XXXXX - DAP implementation...`
- Line 149: `- [x] References GitHub issue #XXXXX`

### 3.2 Verify PR Description

Review `docs/GODOT_PR_DESCRIPTION.md` to ensure:
- [x] Issue number is correct everywhere
- [x] Unit tests section is present
- [x] Testing evidence is included
- [x] Related issues are mentioned
- [x] Checklist is complete

---

## Step 4: Push Branch and Create PR

### 4.1 Verify Branch Status

```bash
cd /Users/adp/Projects/godot

# Verify you're on the correct branch
git branch

# Check commit history (should show 3 commits now)
git log --oneline -3

# Expected:
# <hash3> Add unit tests for DAP Dictionary safety
# 6fcba562 [Debugger] Fix DAP launch request not returning response
# bde72e9c [Debugger] Fix unsafe Dictionary access in DAP implementation
```

### 4.2 Push to Your Fork

```bash
# Push to your fork
git push fork fix/dap-dictionary-safety

# If branch already exists on fork, use force push
git push --force fork fix/dap-dictionary-safety
```

### 4.3 Create Pull Request on GitHub

1. Go to your fork: https://github.com/TransitionMatrix/godot
2. You should see a yellow banner: "Compare & pull request"
3. Click "Compare & pull request"
4. **Base repository**: `godotengine/godot` (base: `master`)
5. **Head repository**: `TransitionMatrix/godot` (compare: `fix/dap-dictionary-safety`)

### 4.4 Fill in PR Details

**Title**:
```
[Debugger] Fix DAP Dictionary crashes and launch timeout
```

**Description**: Copy entire content from `docs/GODOT_PR_DESCRIPTION.md`

**Important Checks**:
- [x] Issue number is correct (not #XXXXX)
- [x] Base branch is `master`
- [x] Your branch shows 3 commits (2 fixes + 1 test)
- [x] Files changed shows ~35 lines (32 fixes + 3 test registration)

### 4.5 Submit PR

- Click "Create pull request"
- **Do not** check "Allow edits by maintainers" (disabled by default for security)
- PR is now submitted!

---

## Step 5: Post-Submission

### 5.1 Monitor PR

- Watch for CI test results (usually takes 10-30 minutes)
- Respond to any reviewer comments promptly
- Be prepared to make changes if requested

### 5.2 CI Checklist

The PR should pass:
- [ ] All unit tests (including your 5 new tests)
- [ ] Compilation on all platforms
- [ ] Code formatting checks
- [ ] No new warnings

### 5.3 Common Review Requests

Be prepared to address:
- **"Can you add more tests?"** - You already have 5, but can add integration tests if requested
- **"Why .get() instead of .has()?"** - Explain: simpler, cleaner, follows AudioStreamWav pattern
- **"Does this break anything?"** - No, purely defensive improvements with default values
- **"Have you tested this?"** - Yes, point to test evidence in PR description

---

## Verification Checklist

Before submitting, verify:

**Issue Created**:
- [x] Issue created on Godot repository
- [x] Issue number recorded: #_____
- [x] Issue includes full problem description
- [x] Issue includes testing evidence

**Unit Tests**:
- [x] Test file in `tests/core/variant/test_debug_adapter_dictionary.h`
- [x] Test registered in `tests/test_main.cpp`
- [x] Tests compile successfully
- [x] All 5 tests pass (12 assertions)
- [x] Tests committed to branch

**PR Description**:
- [x] Issue number replaced everywhere (no #XXXXX)
- [x] Unit tests section present
- [x] Testing evidence included
- [x] Related issues mentioned
- [x] Checklist complete

**Branch Status**:
- [x] On `fix/dap-dictionary-safety` branch
- [x] 3 commits total (2 fixes + 1 test)
- [x] Pushed to fork
- [x] Ready to create PR

**PR Submission**:
- [x] Base: `godotengine/godot:master`
- [x] Head: `YourUsername/godot:fix/dap-dictionary-safety`
- [x] Title matches convention
- [x] Description from `GODOT_PR_DESCRIPTION.md`
- [x] CI tests passing

---

## Quick Reference

**Files Created** (in this project):
- `docs/GODOT_GITHUB_ISSUE.md` - Issue content
- `docs/GODOT_UNIT_TEST.h` - Test file to copy
- `docs/GODOT_TEST_INTEGRATION_GUIDE.md` - Detailed test integration guide
- `docs/GODOT_PR_DESCRIPTION.md` - PR description
- `docs/GODOT_SUBMISSION_GUIDE.md` - This file

**Files Modified** (in Godot repository):
- `tests/core/variant/test_debug_adapter_dictionary.h` - NEW (copied from GODOT_UNIT_TEST.h)
- `tests/test_main.cpp` - Added include for new test
- `editor/debugger/debug_adapter/debug_adapter_parser.cpp` - 25 Dictionary safety fixes + 1 launch fix
- `editor/debugger/debug_adapter/debug_adapter_protocol.cpp` - 6 Dictionary safety fixes
- `editor/debugger/debug_adapter/debug_adapter_types.h` - 2 Dictionary safety fixes

**Commands Summary**:
```bash
# Copy test file
cp docs/GODOT_UNIT_TEST.h /path/to/godot/tests/core/variant/test_debug_adapter_dictionary.h

# Edit test_main.cpp to add include
# (manual edit around line 120)

# Build and test
cd /path/to/godot
scons tests=yes
./bin/godot.*.tests --test --test-case="[DebugAdapter]*"

# Commit test
git add tests/core/variant/test_debug_adapter_dictionary.h tests/test_main.cpp
git commit -m "Add unit tests for DAP Dictionary safety"

# Push
git push fork fix/dap-dictionary-safety

# Create PR on GitHub
# (use web interface)
```

---

## Timeline Estimate

- **Step 1** (Create issue): 5 minutes
- **Step 2** (Add tests): 15 minutes
- **Step 3** (Update PR description): 5 minutes
- **Step 4** (Push & create PR): 10 minutes
- **Total**: ~35 minutes

---

## Need Help?

**If tests fail to compile**:
- Check that `#include "tests/test_macros.h"` is at top of test file
- Verify you edited correct line in `test_main.cpp`
- Ensure path is exact: `tests/core/variant/test_debug_adapter_dictionary.h`

**If tests fail to run**:
- Check that all 5 `TEST_CASE` blocks are present
- Verify no syntax errors in test file
- Try running all tests to see if issue is with your tests or existing tests

**If PR creation fails**:
- Verify you're logged into GitHub
- Check that fork is up to date with upstream
- Ensure branch name is correct
- Verify you have pushed latest commits

**If CI fails**:
- Check CI logs for specific error
- Most likely cause: test file compilation error
- Fix and force-push updated branch

---

## Success Criteria

You'll know you're done when:
- ✅ GitHub issue created and has issue number
- ✅ Unit tests pass locally (5/5 tests, 12/12 assertions)
- ✅ PR created with correct issue reference
- ✅ CI tests are passing (green checkmark)
- ✅ PR is awaiting review

**Next**: Wait for Godot maintainers to review your PR. Response time is typically 1-7 days.
