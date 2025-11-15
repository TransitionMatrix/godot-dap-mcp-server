# Godot Source Code Analysis - DAP Server

**Last Updated**: 2025-11-07
**Godot Version Analyzed**: 4.x (commit: cb3af5afff)
**Source Location**: `/Users/adp/Projects/godot`

This document contains critical implementation details discovered from analyzing Godot Engine's DAP server source code.

---

## Table of Contents

1. [Critical Finding: stepOut Not Implemented](#critical-finding-stepout-not-implemented)
2. [Request Processing Architecture](#request-processing-architecture)
3. [Timeout Mechanisms](#timeout-mechanisms)
4. [Launch Flow](#launch-flow)
5. [Path Validation](#path-validation)
6. [Validated Commands](#validated-commands)

---

## Critical Finding: stepOut Not Implemented

### 🚨 CRITICAL

**Finding**: Godot's DAP server **does not implement stepOut** command at all.

### Evidence

1. **No `req_stepOut` method** in `editor/debugger/debug_adapter/debug_adapter_parser.h:84`
2. **Missing debugger method** in `editor/debugger/script_editor_debugger.h:325-328`:
   ```cpp
   // Only these methods exist:
   void debug_next();      // step over
   void debug_step();      // step into
   void debug_break();     // pause
   void debug_continue();  // continue
   // NO debug_step_out() method
   ```
3. **Zero mentions** of "stepOut" in entire `debug_adapter/` directory

### Source Code

`editor/debugger/debug_adapter/debug_adapter_parser.cpp`:

```cpp
// Stepping commands that ARE implemented:
Dictionary DebugAdapterParser::req_next(const Dictionary &p_params) const {
    DebugAdapterProtocol::get_singleton()->debug_next();
    return _prepare_success_response(p_params);
}

Dictionary DebugAdapterParser::req_stepIn(const Dictionary &p_params) const {
    DebugAdapterProtocol::get_singleton()->debug_step();
    return _prepare_success_response(p_params);
}

// stepOut does NOT exist - no req_stepOut() method anywhere
```

### Impact

**For godot-dap-mcp-server**:
- ❌ **Do NOT implement** `godot_step_out` tool
- ✅ Only implement: `godot_step_over` and `godot_step_into`
- 📝 Document this limitation clearly in tool descriptions
- ⚠️ Warn users if they try to step out

**Workarounds for users**:
1. Set breakpoint after function call, use `continue`
2. Step through remaining lines manually with `step_over`
3. Use `pause` then inspect stack to understand position

---

## Request Processing Architecture

### Command Routing

Location: `editor/debugger/debug_adapter/debug_adapter_protocol.cpp:849`

```cpp
bool DebugAdapterProtocol::process_message(const String &p_text) {
    // Parse JSON request
    Dictionary params = JSON::parse_string(p_text);

    // Godot prepends "req_" to all commands
    String command = "req_" + (String)params["command"];

    // Route to parser method
    if (parser->has_method(command)) {
        Dictionary result = parser->call(command, params);
        send_response(result);
        return true;
    }

    return false;
}
```

**Pattern**: All DAP commands are routed to `req_<command>` methods in `DebugAdapterParser`.

**Examples**:
- `initialize` → `req_initialize()`
- `setBreakpoints` → `req_setBreakpoints()`
- `continue` → `req_continue()`
- `next` → `req_next()`

### Message Format

Godot uses standard DAP Content-Length header format:

```
Content-Length: <bytes>\r\n
\r\n
<JSON_content>
```

**Maximum message size**: 4MB (`DAP_MAX_BUFFER_SIZE`)

---

## Timeout Mechanisms

### Server-Side Timeout

Location: `editor/debugger/debug_adapter/debug_adapter_protocol.cpp:860`

```cpp
// Godot has built-in 5-second timeout
if (OS::get_singleton()->get_ticks_msec() - _current_peer->timestamp > _request_timeout) {
    Dictionary error = prepare_error_response(params, DAP::ErrorType::TIMEOUT,
        "Request timeout");
    send_response(error);
    return true;
}

// _request_timeout = 5000 (5 seconds)
```

**Key Points**:
- Timeout is **5 seconds** by default
- Applies to request processing, not full command execution
- Returns `TIMEOUT` error type

**Implication**: Client-side timeouts (10-30s) are still needed for:
- Network failures
- Editor freezes (timeout mechanism won't run)
- Commands that never send response

### Recommended Client Timeouts

```go
const (
    ConnectTimeout  = 10 * time.Second  // TCP connection
    CommandTimeout  = 30 * time.Second  // DAP commands
    ReadTimeout     = 5 * time.Second   // Single read operation
)
```

---

## Launch Flow

### Two-Step Process

**Critical Pattern**: Launch happens in two phases:

#### Phase 1: Store Parameters

Location: `editor/debugger/debug_adapter/debug_adapter_parser.cpp:200`

```cpp
Dictionary DebugAdapterParser::req_launch(const Dictionary &p_params) const {
    Dictionary args = p_params["arguments"];

    // Validate project path
    if (args.has("project")) {
        String project_path = args["project"];
        if (!is_valid_path(project_path)) {
            return prepare_error_response(p_params, DAP::ErrorType::WRONG_PATH,
                "Project path is not valid");
        }
    }

    // Store launch params (does NOT launch yet!)
    DebugAdapterProtocol::get_singleton()->get_current_peer()->pending_launch = p_params;

    // Return success (but game not launched)
    return _prepare_success_response(p_params);
}
```

#### Phase 2: Actual Launch

Location: `editor/debugger/debug_adapter/debug_adapter_parser.cpp:250`

```cpp
Dictionary DebugAdapterParser::req_configurationDone(const Dictionary &p_params) const {
    // Get stored launch params
    Dictionary pending = get_current_peer()->pending_launch;

    if (!pending.is_empty()) {
        // NOW actually launch the game
        _launch_process(pending);
    }

    return _prepare_success_response(p_params);
}

Dictionary DebugAdapterParser::_launch_process(const Dictionary &p_params) const {
    Dictionary args = p_params["arguments"];
    const String scene = args.get("scene", "main");

    // Route to appropriate launch method
    if (scene == "main") {
        EditorRunBar::get_singleton()->play_main_scene(false, play_args);
    } else if (scene == "current") {
        EditorRunBar::get_singleton()->play_current_scene(false, play_args);
    } else {
        EditorRunBar::get_singleton()->play_custom_scene(scene, play_args);
    }

    return Dictionary();
}
```

**Why Two Steps?**
- Allows setting breakpoints before game starts
- Standard DAP pattern for all debug adapters
- Prevents race conditions

**Client Implementation**:
```go
// Step 1: Send launch request
err := client.Launch(ctx, launchArgs)

// Step 2: Send configurationDone (triggers actual launch)
err := client.ConfigurationDone(ctx)
```

### Launch Argument Processing

**Source Analysis**: Based on `debug_adapter_parser.cpp` analysis, here's exactly when and how Godot reads launch arguments:

| Field | Type | Default | Purpose | Read When | Dictionary Safety |
|-------|------|---------|---------|-----------|-------------------|
| `request` | string | - | **Ignored by Godot** | Never | N/A (never accessed) |
| `project` | string | - | Validate project path | `req_launch:171` | ✅ Safe (`.has()` check before `.get()`) |
| `godot/custom_data` | bool | `false` | Enable custom data events | `req_launch:178` | ✅ Safe (`.has()` check before `.get()`) |
| `noDebug` | bool | `false` | Skip breakpoints | `_launch_process:204` | ✅ Safe (`.get()` with default) |
| `platform` | string | `"host"` | Platform: "host", "android", "web" | `_launch_process:208` | ✅ Safe (`.get()` with default) |
| `scene` | string | `"main"` | Scene: "main", "current", or path | `_launch_process:211` | ✅ Safe (`.get()` with default) |
| `playArgs` | array | `[]` | Command-line arguments | `_extract_play_arguments:189` | ✅ Safe (`.has()` check, type validation) |
| `device` | int | `-1` | Device ID (android/web) | `_launch_process:220` | ✅ Safe (`.get()` with default) |

**Key Findings**:

1. **`request` field ignored**: Many DAP clients (e.g., Zed) send `"request": "launch"` for spec compliance, but Godot never reads this field. The command type (`launch` vs `attach`) determines behavior, not the field value.

2. **All reads are Dictionary-safe**: Every field access uses either:
   - `.has()` check before `.get()` (for project, godot/custom_data, playArgs)
   - `.get(key, default)` with default value (for noDebug, platform, scene, device)

   This prevents Dictionary access errors even when clients send minimal launch requests.

3. **Two-stage reading**:
   - `req_launch()` (lines 169-185): Reads only `project` and `godot/custom_data`
   - `_launch_process()` (lines 201-257): Reads remaining fields when actually launching

4. **`godot/custom_data` extension**: Godot-specific extension that enables forwarding of internal debugger messages as `godot/custom_data` events. See [DAP_PROTOCOL.md - Godot-Specific Extensions](DAP_PROTOCOL.md#godot-specific-extensions).

**Source Code References**:

```cpp
// req_launch - Initial validation (lines 169-185)
Dictionary DebugAdapterParser::req_launch(const Dictionary &p_params) const {
    Dictionary args = p_params.get("arguments", Dictionary());

    // Safe: .has() check before access
    if (args.has("project") && !is_valid_path(args.get("project", ""))) {
        return prepare_error_response(p_params, DAP::ErrorType::WRONG_PATH, ...);
    }

    // Safe: .has() check before access
    if (args.has("godot/custom_data")) {
        get_current_peer()->supportsCustomData = args.get("godot/custom_data", false);
    }

    // Store for later (phase 2)
    get_current_peer()->pending_launch = p_params;
    return prepare_success_response(p_params);
}

// _launch_process - Actual launch (lines 201-257)
Dictionary DebugAdapterParser::_launch_process(const Dictionary &p_params) const {
    Dictionary args = p_params.get("arguments", Dictionary());

    // All safe: .get() with defaults
    bool noDebug = args.get("noDebug", false);
    String platform = args.get("platform", "host");
    String scene = args.get("scene", "main");
    int device = args.get("device", -1);
    Vector<String> playArgs = _extract_play_arguments(args);  // Safe internally

    // Launch based on scene mode
    if (scene == "main") {
        EditorRunBar::get_singleton()->play_main_scene(false, playArgs);
    } else if (scene == "current") {
        EditorRunBar::get_singleton()->play_current_scene(false, playArgs);
    } else {
        EditorRunBar::get_singleton()->play_custom_scene(scene, playArgs);
    }

    return prepare_success_response(p_params);
}

// _extract_play_arguments - Safe array extraction (lines 187-199)
Vector<String> DebugAdapterParser::_extract_play_arguments(const Dictionary &p_args) const {
    Vector<String> play_args;

    // Safe: .has() check before access
    if (p_args.has("playArgs")) {
        Variant v = p_args.get("playArgs", Array());

        // Type validation
        if (v.get_type() == Variant::ARRAY) {
            Array arr = v;
            for (const Variant &arg : arr) {
                play_args.push_back(String(arg));
            }
        }
    }

    return play_args;
}
```

**Implications for DAP Clients**:

1. **Minimal launch requests work**: Can send just `{ "command": "launch", "arguments": {} }` - all fields have safe defaults
2. **Extra fields ignored**: Sending spec-compliant fields like `"request": "launch"` is safe (Godot ignores them)
3. **No Dictionary errors in launch**: All reads are protected, unlike some other commands (e.g., `Source::from_json`)

---

## Path Validation

### Project Path Validation

Location: `editor/debugger/debug_adapter/debug_adapter_parser.cpp:205`

```cpp
bool DebugAdapterParser::is_valid_path(const String &p_path) {
    // Check if path exists
    if (!DirAccess::exists(p_path)) {
        return false;
    }

    // Check if project.godot exists
    String project_file = p_path.path_join("project.godot");
    if (!FileAccess::exists(project_file)) {
        return false;
    }

    // Additional validation: must match editor's project
    // (in some configurations)
    if (EditorSettings::get_singleton()->get_project_path() != p_path) {
        // May return error depending on configuration
    }

    return true;
}
```

**Validation Rules**:
1. Path must exist as directory
2. Directory must contain `project.godot` file
3. Path may need to match editor's current project (context-dependent)

**Common Errors**:
- `WRONG_PATH`: Path invalid or project.godot not found
- Path mismatch: Symbolic links cause issues
- Case sensitivity: Windows vs Unix path differences

### File Path Handling

Godot auto-converts paths but validates before conversion:

```cpp
// Windows path conversion
String convert_path(const String &p_path) {
    // Convert backslashes to forward slashes
    String result = p_path.replace("\\", "/");

    // Ensure uppercase drive letter on Windows
    if (result.length() >= 2 && result[1] == ':') {
        result[0] = _unicode_toupper(result[0]);
    }

    return result;
}
```

**Best Practice**: Send paths pre-converted:
- Use forward slashes: `/`
- Windows: Uppercase drive letter: `C:/path/to/file.gd`
- Always absolute paths

---

## Validated Commands

### Commands Proven Working

From previous DAP experimentation and source code analysis:

| Command | Status | Source Method | Notes |
|---------|--------|---------------|-------|
| `initialize` | ✅ Tested | `req_initialize()` | Returns capabilities |
| `configurationDone` | ✅ Tested | `req_configurationDone()` | Triggers launch |
| `setBreakpoints` | ✅ Tested | `req_setBreakpoints()` | Replaces all BPs in file |
| `threads` | ✅ Tested | `req_threads()` | Always returns 1 thread |
| `stackTrace` | ✅ Tested | `req_stackTrace()` | Returns call stack |
| `scopes` | ✅ Tested | `req_scopes()` | 3 scopes: Locals, Members, Globals |
| `variables` | ✅ Tested | `req_variables()` | Returns variables in scope |
| `evaluate` | ✅ Tested | `req_evaluate()` | Evaluates GDScript |
| `continue` | ✅ Tested | `req_continue()` | Resumes execution |
| `next` | ✅ Tested | `req_next()` | Step over |
| `stepIn` | ✅ Tested | `req_stepIn()` | Step into |
| `pause` | ⚠️ Untested | `req_pause()` | Should work (implemented) |
| `stepOut` | ❌ Not Implemented | - | **DO NOT USE** |

### Commands Not Implemented

| Command | Status | Reason |
|---------|--------|--------|
| `stepOut` | ❌ Missing | No debugger method exists |
| `setFunctionBreakpoints` | ❌ Missing | Not in parser |
| `setExceptionBreakpoints` | ❌ Missing | Not in parser |
| `dataBreakpointInfo` | ❌ Missing | Not in parser |
| `setDataBreakpoints` | ❌ Missing | Not in parser |
| `setInstructionBreakpoints` | ❌ Missing | Not in parser |

---

## Event Types

### Async Events Sent by Godot

Location: `editor/debugger/debug_adapter/debug_adapter_protocol.cpp`

```cpp
// Events sent during debugging:
- "initialized"     // Ready for debugging
- "process"         // Game launched
- "stopped"         // Execution paused (breakpoint/step/pause)
- "continued"       // Execution resumed
- "output"          // Console output
- "breakpoint"      // Breakpoint state changed
- "terminated"      // Game stopped
- "exited"          // Game process exited
```

**Client Implementation**: Must filter these events from command responses using event filtering pattern.

---

## Connection Limits

Location: `editor/debugger/debug_adapter/debug_adapter_server.h:50`

```cpp
#define DAP_MAX_CLIENTS 8
```

**Maximum concurrent connections**: 8 clients

**Behavior**: All clients see the same debugging state (single session model).

---

## References

### Godot Source Files

Key files for understanding DAP implementation:

```
editor/debugger/debug_adapter/
├── debug_adapter_server.h         # Server setup, connection handling
├── debug_adapter_server.cpp
├── debug_adapter_protocol.h       # Protocol handling, message routing
├── debug_adapter_protocol.cpp
├── debug_adapter_parser.h         # Command implementations
├── debug_adapter_parser.cpp
└── debug_adapter_types.h          # Type definitions, error codes
```

### Official Documentation

- [Godot Debugger Documentation](https://docs.godotengine.org/en/stable/tutorials/editor/debugger_panel.html)
- [Microsoft DAP Specification](https://microsoft.github.io/debug-adapter-protocol/)

### Related Documentation

- [GODOT_DAP_FAQ.md](GODOT_DAP_FAQ.md) - Common questions and troubleshooting
- [DAP_PROTOCOL.md](DAP_PROTOCOL.md) - Protocol details and usage
