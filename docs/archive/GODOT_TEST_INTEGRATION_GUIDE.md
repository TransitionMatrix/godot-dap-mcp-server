# Unit Test Integration Guide for Godot

This guide explains how to add the Dictionary safety unit tests to the Godot repository.

## Test File Location

**Create:** `tests/core/variant/test_debug_adapter_dictionary.h`

**Rationale:**
- Tests Dictionary safety patterns (core variant functionality)
- Fits existing test structure: `tests/core/variant/test_dictionary.h`
- Demonstrates the fix without requiring full debug adapter setup
- Can be run with standard test suite

## Integration Steps

### Step 1: Copy Test File

```bash
# In Godot repository root
cp /path/to/test_debug_adapter_dictionary.h tests/core/variant/
```

### Step 2: Register Test in test_main.cpp

Edit `tests/test_main.cpp` and add the include near the other variant tests (around line 120):

```cpp
#include "tests/core/variant/test_array.h"
#include "tests/core/variant/test_callable.h"
#include "tests/core/variant/test_debug_adapter_dictionary.h"  // ADD THIS LINE
#include "tests/core/variant/test_dictionary.h"
#include "tests/core/variant/test_variant.h"
#include "tests/core/variant/test_variant_utility.h"
```

### Step 3: Build and Run Tests

```bash
# Build with tests enabled
scons tests=yes

# Run all tests
./bin/godot.linuxbsd.editor.x86_64.tests --test

# Run only the new tests
./bin/godot.linuxbsd.editor.x86_64.tests --test --test-case="[DebugAdapter]*"

# Run specific test
./bin/godot.linuxbsd.editor.x86_64.tests --test --test-case="[DebugAdapter] Dictionary.get() safely handles missing keys"
```

## Expected Test Output

```
[doctest] doctest version is "2.4.11"
[doctest] run with "--help" for options
===============================================================================
[doctest] test cases:  5 |  5 passed | 0 failed | 0 skipped
[doctest] assertions: 12 | 12 passed | 0 failed |
[doctest] Status: SUCCESS!
```

## Alternative: Integration Tests

If you want to test the actual debug adapter code directly, you would need:

1. **Create:** `tests/editor/debugger/test_debug_adapter_parser.h`
2. **Challenges:**
   - Requires TOOLS_ENABLED
   - Needs mock DAP client
   - More complex setup
   - Requires linking editor code

**Recommendation:** The pattern-demonstration tests in `test_debug_adapter_dictionary.h` are sufficient for this PR, as they:
- Demonstrate the fix principle
- Run quickly in standard test suite
- Don't require complex setup
- Verify Dictionary.get() behavior

Integration tests of the actual debug_adapter classes can be added in a future PR if needed.

## Verification Checklist

- [ ] Test file copied to `tests/core/variant/test_debug_adapter_dictionary.h`
- [ ] Include added to `tests/test_main.cpp`
- [ ] Project builds successfully: `scons tests=yes`
- [ ] All 5 test cases pass
- [ ] No compilation errors or warnings
- [ ] Tests run in CI pipeline (if enabled)

## Notes for PR

In the PR description, mention:

```markdown
## Unit Tests

Added 5 unit tests demonstrating Dictionary safety:
- `tests/core/variant/test_debug_adapter_dictionary.h`
- Tests verify Dictionary.get() safely handles missing keys
- All tests pass (12 assertions)

Run with:
\`\`\`bash
./bin/godot.*.tests --test --test-case="[DebugAdapter]*"
\`\`\`
```

## CI Integration

The tests will automatically run in Godot's CI pipeline if:
- [ ] Test file is in `tests/` directory
- [ ] Include is added to `tests/test_main.cpp`
- [ ] No compilation errors

No additional CI configuration is needed - Godot's existing test infrastructure handles it.
