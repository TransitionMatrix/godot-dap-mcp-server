# Popular Godot DAP Clients for Testing

**Document Purpose:** Identify DAP clients to demonstrate the Dictionary regression and validate fixes
**Investigation Date:** November 9, 2025
**Context:** Testing DAP server fixes for unsafe Dictionary access patterns

---

## Overview

This document lists popular Debug Adapter Protocol (DAP) clients used with Godot, ranked by adoption and ease of testing. These clients can be used to demonstrate the Dictionary regression bugs and validate that fixes work correctly.

---

## 1. Visual Studio Code (VSCode)

**Status:** ⭐ **MOST POPULAR** - Official support
**Ease of Setup:** ⭐⭐⭐⭐⭐ Very Easy
**Best For:** Quick testing, widest adoption

### Official Extensions

**For GDScript:**
- **Name:** Godot Tools
- **Repository:** https://github.com/godotengine/godot-vscode-plugin
- **Marketplace:** Search "Godot Tools" in VSCode
- **Features:** Full DAP debugging, LSP support, launch configurations

**For C#:**
- **Name:** C# Tools for Godot
- **Repository:** https://github.com/godotengine/godot-csharp-vscode
- **Marketplace:** https://marketplace.visualstudio.com/items?itemName=neikeq.godot-csharp-vscode
- **Features:** C# debugging, debugger adapter based on vscode-mono-debug

### Setup Instructions

1. **Install Godot Tools extension** from VSCode marketplace
2. **Configure LSP port:** Set to 6005 in VSCode settings
3. **Create launch configuration:** Use Debug panel or "Generate Assets for Build and Debug"
4. **Connect to Godot:** Start Godot editor with project open

### DAP Configuration

VSCode connects to Godot's DAP server on:
- **Host:** 127.0.0.1
- **Port:** 6006 (default DAP port)
- **Port:** 6007 (debug adapter alternative)

### Known Issues with Regression

- Issue #97819 - Completions request not working
- Will encounter Dictionary errors on malformed/incomplete requests

### Testing Value

**Why Use VSCode:**
- ✅ Most widely used Godot DAP client
- ✅ Easy to install and configure
- ✅ Official support from Godot team
- ✅ Good error reporting
- ✅ Representative of typical user experience

**Best for demonstrating:**
- Basic DAP functionality
- Common debugging workflows
- Standard protocol compliance

---

## 2. Zed Editor

**Status:** ⭐⭐⭐⭐ **ACTIVELY REPORTED ISSUES** - Community maintained
**Ease of Setup:** ⭐⭐⭐⭐ Easy
**Best For:** Demonstrating strict protocol compliance issues

### Extension Details

- **Name:** GDScript for Zed
- **Extension Page:** https://zed.dev/extensions/gdscript
- **Repository:** https://github.com/GDQuest/zed-gdscript
- **Maintainer:** GDQuest (community)
- **Language:** Rust

### Setup Instructions

1. **Install Zed editor** from https://zed.dev
2. **Install GDScript extension** from Zed extensions marketplace
3. **Create debug configuration:** Add `.zed/debug.json` to project root:

```json
{
  "configurations": [
    {
      "name": "Godot (Launch)",
      "type": "godot",
      "request": "launch"
    },
    {
      "name": "Godot (Attach)",
      "type": "godot",
      "request": "attach"
    }
  ]
}
```

4. **Run debug task:** Press F4, select task from menu

### Known Issues with Regression

- **Issue #108288** - `request_seq` float regression (CRITICAL)
  - Reporter: @fstxz
  - **Zed requires strict integer types**
  - Fails to parse responses with `1.0` instead of `1`
  - **This is THE issue that exposed the float problem**

### Compatibility Notes

⚠️ **IMPORTANT:** At time of writing:
- **Works:** Godot 4.3 and 4.5.1-stable
- **BROKEN:** Godot 4.4, 4.6-dev/master (due to Dictionary regression)

### Testing Value

**Why Use Zed:**
- ✅ **STRICT protocol compliance** - catches spec violations
- ✅ **Already broken by regression** - perfect test case
- ✅ **Reporter is active** - @fstxz can validate fixes
- ✅ Fast, modern editor with growing adoption
- ✅ Rust-based, good error messages

**Best for demonstrating:**
- Protocol compliance failures
- Type strictness issues (float vs int)
- The exact regression that was reported
- Before/after fix comparison

**Testing Steps:**
1. Test with Godot 4.5.1-stable (should work)
2. Test with current master (should fail with deserialization error)
3. Test with your fix applied (should work again)

---

## 3. Neovim (nvim-dap)

**Status:** ⭐⭐⭐⭐ **POPULAR** - Community plugins available
**Ease of Setup:** ⭐⭐⭐ Moderate (requires Lua config)
**Best For:** Power users, demonstrating CLI workflows

### Plugin Options

**Option 1: Manual nvim-dap Configuration**
- **Repository:** https://github.com/mfussenegger/nvim-dap
- **Requires:** Manual Lua configuration

**Option 2: godot.nvim (Recommended)**
- **Repository:** https://github.com/Lommix/godot.nvim
- **Features:** Integrated setup, run and debug from Neovim

**Option 3: godotdev.nvim (Full-featured)**
- **Repository:** https://github.com/Mathijs-Bakker/godotdev.nvim
- **Features:** LSP, DAP, Tree-sitter, external editor integration
- **Dependencies:** nvim-lspconfig, nvim-dap, nvim-dap-ui, nvim-treesitter

**Option 4: godot-lsp.nvim**
- **Repository:** https://github.com/Mathijs-Bakker/godot-lsp.nvim
- **Features:** Experimental DAP with breakpoints, step-through, variable inspection

### Manual Configuration (nvim-dap)

```lua
local dap = require("dap")

-- DAP adapter configuration
dap.adapters.godot = {
  type = "server",
  host = "127.0.0.1",
  port = 6006,
}

-- Debug configurations
dap.configurations.gdscript = {
  {
    type = "godot",
    request = "launch",
    name = "Launch scene",
    project = "${workspaceFolder}",
    launch_scene = true,
  },
}
```

### Known Issues with Regression

- **Issue #110749** - Crash when removing breakpoints (CRITICAL)
  - Reporter: @jalelegenda
  - Godot crashes (signal 11) when removing breakpoints during DAP session
  - Reproducible in 4.4.1-stable
  - Uses Neovim as external editor with DAP
  - **Likely Dictionary-related crash**

- **Issue #111077** - Path handling issues with Neovim DAP
  - Linux path problems during debugging

### Testing Value

**Why Use Neovim:**
- ✅ Popular among developers
- ✅ **Known crash with breakpoint removal** - critical test case
- ✅ CLI-based, good for automated testing
- ✅ Multiple plugin options for different test scenarios
- ✅ Reported issues provide clear reproduction steps

**Best for demonstrating:**
- Breakpoint removal crashes
- External editor integration
- CLI debugging workflows
- Crash regression testing

**Testing Steps:**
1. Set up Neovim with nvim-dap and Godot config
2. Set breakpoints during debugging session
3. **Remove breakpoint while debugging** (triggers crash on master)
4. Verify crash is fixed with Dictionary patches

---

## 4. JetBrains Rider

**Status:** ⭐⭐⭐⭐⭐ **PROFESSIONAL/COMMERCIAL** - Official JetBrains support
**Ease of Setup:** ⭐⭐⭐⭐⭐ Very Easy (bundled since 2024.2)
**Best For:** Professional workflows, IDE integration testing

### Official Support

- **Product Page:** https://www.jetbrains.com/lp/rider-godot/
- **Documentation:** https://www.jetbrains.com/help/rider/Godot.html
- **Plugin:** Bundled since Rider 2024.2 (no separate installation needed)
- **Repository:** https://github.com/JetBrains/godot-support

### Features

- **Out-of-the-box support** - No configuration needed
- **Automatic run configurations** - Created when opening Godot project
- **C# debugging** - Full support for Godot C# projects
- **GDScript support** - Native + LSP integration
- **DAP-based debugging** - Uses Debug Adapter Protocol

### Key Contributor

**Ivan Shakhov (@van800)** - JetBrains employee, active Godot DAP contributor
- Authored PR #109637 - Scene selection and CLI args
- Authored PR #109987 - Device debugging refactor
- Reports issues from Rider perspective
- Professional use case validation

### Known Issues with Regression

- Issue #106969 - Debug Android export via DAP (reported by @van800)
- Likely affected by Dictionary regression on master branch

### Testing Value

**Why Use Rider:**
- ✅ **Commercial product** - represents professional use case
- ✅ **Active JetBrains involvement** - direct line to @van800
- ✅ Advanced debugging features
- ✅ Good for IDE integration validation
- ✅ Strict requirements for professional use

**Best for demonstrating:**
- Professional workflow reliability
- IDE integration correctness
- Advanced debugging features
- Commercial product expectations

**Limitations:**
- ⚠️ Commercial license required (trial available)
- ⚠️ Heavy IDE, slower to start
- ⚠️ Windows/macOS/Linux support varies

---

## 5. Emacs (dape / dap-mode)

**Status:** ⭐⭐⭐ **NICHE** - Community supported
**Ease of Setup:** ⭐⭐ Advanced (Elisp configuration)
**Best For:** Emacs users, demonstrating protocol flexibility

### DAP Client Options

**Option 1: dape (Modern, Recommended for Godot 4)**
- **Repository:** https://github.com/svaante/dape
- **Package:** GNU ELPA - https://elpa.gnu.org/packages/dape.html
- **Features:** Modern DAP client for Emacs
- **Godot Support:** Godot 4 (via DAP)

**Option 2: dap-mode (Traditional)**
- **Package:** Available via MELPA
- **Features:** Established DAP client
- **Godot Support:** Godot 4 (via DAP)
- **Doom Emacs:** Supported with `(:lang gdscript +lsp)`

### GDScript Mode

- **Repository:** https://github.com/godotengine/emacs-gdscript-mode
- **Official:** Maintained by Godot organization
- **Features:** GDScript syntax, LSP integration
- **Debugging:** Godot 3 (custom), Godot 4 (use dape/dap-mode)

### Configuration Example (dape)

```elisp
;; Add to your Emacs config
(use-package dape
  :config
  (add-to-list 'dape-configs
               '(godot-launch
                 modes (gdscript-mode)
                 host "127.0.0.1"
                 port 6006
                 command-cwd dape-cwd-fn
                 :type "godot"
                 :request "launch"
                 :name "Launch scene"
                 :project "${workspaceFolder}")))
```

### Testing Value

**Why Use Emacs:**
- ✅ Different client implementation
- ✅ Tests protocol flexibility
- ✅ Represents text-based workflow
- ✅ Good for edge case discovery

**Best for demonstrating:**
- Alternative DAP client implementations
- Text-based debugging workflows
- Protocol parser differences

**Limitations:**
- ⚠️ Smaller user base
- ⚠️ Configuration complexity
- ⚠️ Less likely to catch issues than mainstream clients

---

## Testing Strategy & Recommendations

### Minimum Test Coverage

For demonstrating the regression and validating fixes:

**MUST TEST (Priority 1):**
1. **Zed Editor** - Already broken, strict compliance, reported issue
2. **VSCode** - Most popular, easy setup, widest impact
3. **Neovim** - Known crash issue, CLI-based

**SHOULD TEST (Priority 2):**
4. **Rider** - Professional use case, JetBrains validation

**OPTIONAL (Priority 3):**
5. **Emacs** - Alternative implementation, edge cases

### Test Scenarios

#### Scenario 1: Basic Request/Response
**Tests:** `prepare_success_response()`, `prepare_error_response()`
**Clients:** All
**Steps:**
1. Connect client to Godot DAP server
2. Send `initialize` request
3. Verify response has correct `request_seq` (int, not float)
4. Verify `command` field is present

**Expected on Master (Broken):**
- Zed: Fails to parse response (deserialization error)
- Others: May work but with errors

**Expected with Fix:**
- All clients: Parse response successfully

#### Scenario 2: Error Responses
**Tests:** `prepare_error_response()` with missing keys
**Clients:** All
**Steps:**
1. Send malformed request (missing `seq` or `command`)
2. Verify error response is generated
3. Check for Dictionary access crashes/errors

**Expected on Master (Broken):**
- Error: "Bug: Dictionary::operator[] used when there was no value for the given key"
- Possible crash or empty fields

**Expected with Fix:**
- Clean error response with default values

#### Scenario 3: Breakpoint Removal
**Tests:** Breakpoint management during debugging
**Clients:** Neovim (known issue), VSCode
**Steps:**
1. Start debugging session
2. Add breakpoints
3. **Remove breakpoint while debugging**
4. Verify no crash

**Expected on Master (Broken):**
- Neovim: Crash (signal 11) - Issue #110749
- Related to Dictionary access in breakpoint handling

**Expected with Fix:**
- No crash, clean removal

#### Scenario 4: Type Deserialization
**Tests:** `Source::from_json()`, `SourceBreakpoint::from_json()`
**Clients:** All
**Steps:**
1. Send `setBreakpoints` request with incomplete source info
2. Verify fields default properly

**Expected on Master (Broken):**
- Dictionary access errors for missing fields

**Expected with Fix:**
- Clean defaults applied

### Demonstration Script

**For showing the regression:**

```bash
# Terminal 1: Start Godot with current master
./bin/godot.linuxbsd.editor.x86_64

# Terminal 2: Connect Zed editor
# Open project in Zed, start debugging
# OBSERVE: Deserialization error on initialize

# Terminal 3: Monitor Godot console
# OBSERVE: Dictionary::operator[] errors
```

**For showing the fix:**

```bash
# Terminal 1: Start Godot with fix applied
./bin/godot.linuxbsd.editor.x86_64

# Terminal 2: Connect Zed editor
# Open project in Zed, start debugging
# OBSERVE: Successful connection and debugging

# Terminal 3: Monitor Godot console
# OBSERVE: No Dictionary errors
```

---

## DAP Connection Details

### Standard Godot DAP Configuration

**Default Ports:**
- **DAP Server:** 6006
- **DAP Alternative:** 6007
- **LSP Server:** 6005 (GDScript language server)

**Connection Type:** TCP socket
**Host:** 127.0.0.1 (localhost)
**Protocol:** Debug Adapter Protocol (JSON-RPC over TCP)

### Command Line Options

```bash
# Start Godot with custom DAP port
./godot --dap-port 6006

# Start Godot with custom LSP port
./godot --lsp-port 6005
```

### Editor Settings

In Godot editor:
- **Editor Settings** → **Network** → **Debug Adapter**
  - Remote Port: 6007 (default)
- **Editor Settings** → **Network** → **Debug**
  - Remote Port: 6006
  - Sync Breakpoints: On (recommended)

---

## Client Comparison Matrix

| Client | Ease of Setup | Popularity | Known Issues | Best For | License |
|--------|---------------|------------|--------------|----------|---------|
| **VSCode** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Completions | General testing | Free |
| **Zed** | ⭐⭐⭐⭐ | ⭐⭐⭐ | Float regression (#108288) | **Regression demo** | Free |
| **Neovim** | ⭐⭐⭐ | ⭐⭐⭐⭐ | Crash (#110749) | **Crash demo** | Free |
| **Rider** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Android debug | Professional validation | Commercial |
| **Emacs** | ⭐⭐ | ⭐⭐ | None reported | Edge cases | Free |

### Recommendation Priority

**For demonstrating the regression:**
1. **Zed** - Most direct evidence (Issue #108288)
2. **Neovim** - Crash evidence (Issue #110749)
3. **VSCode** - Widespread impact

**For validating the fix:**
1. **VSCode** - Ensures fix doesn't break common case
2. **Zed** - Ensures strict compliance is restored
3. **Neovim** - Ensures crash is fixed
4. **Rider** - Ensures professional use case works

---

## Additional Resources

### Official Documentation
- Godot DAP Spec: https://microsoft.github.io/debug-adapter-protocol/
- Godot Editor Debugger: `editor/debugger/` in source

### Community Resources
- Godot Discord: #programming channel
- Godot Forums: https://forum.godotengine.org/
- Reddit: r/godot

### Issue References
- #108288 - Zed float regression
- #110749 - Neovim crash on breakpoint removal
- #111077 - Neovim path issues
- #97819 - VSCode completions request
- #106969 - Rider Android debugging

---

## Conclusion

### Recommended Testing Approach

**Minimal viable demonstration:**
1. **Zed Editor** - Shows the exact reported regression
2. **VSCode** - Shows it affects the most popular client
3. **Monitor console** - Shows Dictionary::operator[] errors

**Comprehensive validation:**
1. All 3 minimal clients above
2. **Neovim** - Validates crash fix
3. **Rider** (if available) - Professional validation

### Expected Results

**Before Fix (Master):**
- ❌ Zed: Connection fails (deserialization error)
- ⚠️ VSCode: May work but logs errors
- ❌ Neovim: Crashes on breakpoint removal
- ⚠️ Rider: Likely encounters issues

**After Fix (Your PR):**
- ✅ Zed: Connection succeeds, debugging works
- ✅ VSCode: Works without errors
- ✅ Neovim: No crash on breakpoint removal
- ✅ Rider: Professional features work correctly
- ✅ Console: No Dictionary::operator[] errors

This demonstrates that your fix:
1. Resolves reported regressions
2. Maintains compatibility with popular clients
3. Fixes crashes and errors
4. Follows proper Godot patterns (`.get()` with defaults)
