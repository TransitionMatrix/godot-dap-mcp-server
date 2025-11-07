# Godot DAP Protocol Details

**Last Updated**: 2025-11-07

This document describes Godot-specific DAP protocol details, based on Godot Engine source code analysis.

---

## Table of Contents

1. [Launch Request Format](#launch-request-format)
2. [Scene Launch Modes](#scene-launch-modes)
3. [Platform Support](#platform-support)
4. [Supported Commands](#supported-commands)
5. [Path Validation](#path-validation)

---

## Launch Request Format

Based on Godot source (`editor/debugger/debug_adapter/debug_adapter_parser.cpp`):

### Launch Arguments

```json
{
  "project": "/absolute/path/to/project",
  "scene": "main" | "current" | "res://path/to/scene.tscn",
  "platform": "host" | "android" | "web",
  "device": -1,
  "noDebug": false,
  "profiling": false,
  "debug_collisions": false,
  "debug_paths": false,
  "debug_navigation": false,
  "debug_avoidance": false,
  "additional_options": "string",
  "playArgs": ["--arg1", "value1"]
}
```

### Field Details

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `project` | string | No | Editor's current project | Absolute path to project directory |
| `scene` | string | No | `"main"` | Scene to launch (see modes below) |
| `platform` | string | No | `"host"` | Target platform |
| `device` | integer | No | `-1` | Android device index |
| `noDebug` | boolean | No | `false` | Skip breakpoints |
| `profiling` | boolean | No | `false` | Enable performance profiling |
| `debug_collisions` | boolean | No | `false` | Show collision shapes visually |
| `debug_paths` | boolean | No | `false` | Show path navigation |
| `debug_navigation` | boolean | No | `false` | Show navigation mesh |
| `debug_avoidance` | boolean | No | `false` | Show avoidance obstacles |
| `additional_options` | string | No | `""` | Additional CLI arguments |
| `playArgs` | array | No | `[]` | Array of CLI arguments |

### Godot Source Implementation

```cpp
Dictionary DebugAdapterParser::req_launch(const Dictionary &p_params) const {
    Dictionary args = p_params["arguments"];

    // Validate project path if provided
    if (args.has("project") && !is_valid_path(args["project"])) {
        return prepare_error_response(p_params, DAP::ErrorType::WRONG_PATH, ...);
    }

    // Launch deferred until configurationDone
    DebugAdapterProtocol::get_singleton()->get_current_peer()->pending_launch = p_params;
    return Dictionary();
}

Dictionary DebugAdapterParser::_launch_process(const Dictionary &p_params) const {
    Dictionary args = p_params["arguments"];
    const String scene = args.get("scene", "main");

    if (scene == "main") {
        EditorRunBar::get_singleton()->play_main_scene(false, play_args);
    } else if (scene == "current") {
        EditorRunBar::get_singleton()->play_current_scene(false, play_args);
    } else {
        EditorRunBar::get_singleton()->play_custom_scene(scene, play_args);
    }
}
```

**Critical**: The `launch` request only *stores* parameters. Actual launch happens when `configurationDone` is sent.

---

## Scene Launch Modes

### Mode: "main"

**Behavior**: Launches the main scene defined in `project.godot`

**Example**:
```json
{
  "scene": "main"
}
```

**Godot Implementation**:
```cpp
EditorRunBar::get_singleton()->play_main_scene(false, play_args);
```

### Mode: "current"

**Behavior**: Launches the currently open scene in the editor

**Example**:
```json
{
  "scene": "current"
}
```

**Use Case**: Testing the scene you're actively editing

**Godot Implementation**:
```cpp
EditorRunBar::get_singleton()->play_current_scene(false, play_args);
```

### Mode: Custom Scene Path

**Behavior**: Launches a specific scene by resource path

**Example**:
```json
{
  "scene": "res://scenes/test_player.tscn"
}
```

**Path Format**:
- Must use Godot resource path format: `res://path/to/scene.tscn`
- Relative to project root
- Must be a valid `.tscn` or `.scn` file

**Godot Implementation**:
```cpp
EditorRunBar::get_singleton()->play_custom_scene(scene, play_args);
```

---

## Platform Support

### Supported Platforms

| Platform Value | Description | Requirements |
|----------------|-------------|--------------|
| `"host"` | Local desktop platform (Windows/Mac/Linux) | Default, always available |
| `"android"` | Android device/emulator | Requires Android export templates, device connected |
| `"web"` | Web browser | Requires Web export templates, HTTP server |

### Platform-Specific Parameters

**Android**:
```json
{
  "platform": "android",
  "device": 0
}
```

The `device` parameter specifies which Android device to use (index into connected devices list).

**Web**:
```json
{
  "platform": "web"
}
```

Launches game in default web browser.

---

## Supported Commands

### Initialization Sequence

**Required sequence for all debugging**:

```
1. initialize          → Get capabilities
2. launch/attach       → Store launch parameters
3. setBreakpoints      → Optional, set breakpoints
4. configurationDone   → Actually launches game
```

### Stepping Commands

| Command | Supported | DAP Method | Godot Method |
|---------|-----------|------------|--------------|
| `next` | ✅ Yes | `req_next()` | `debug_next()` |
| `stepIn` | ✅ Yes | `req_stepIn()` | `debug_step()` |
| `stepOut` | ❌ **NOT IMPLEMENTED** | - | - |

**Critical**: `stepOut` is **not implemented** in Godot's DAP server. Do not attempt to use it.

### Execution Control

| Command | Supported | Notes |
|---------|-----------|-------|
| `continue` | ✅ Yes | Resumes execution |
| `pause` | ✅ Yes | Pauses execution |
| `terminate` | ✅ Yes | Stops the game |
| `restart` | ✅ Yes | Restarts the game |

### Inspection Commands

| Command | Supported | Notes |
|---------|-----------|-------|
| `threads` | ✅ Yes | Always returns 1 thread: "Main" |
| `stackTrace` | ✅ Yes | Returns call stack |
| `scopes` | ✅ Yes | Always returns 3 scopes: Locals, Members, Globals |
| `variables` | ✅ Yes | Returns variables in scope |
| `evaluate` | ✅ Yes | Evaluates GDScript expressions |
| `setVariable` | ✅ Yes | Modifies variable values |

### Breakpoint Commands

| Command | Supported | Notes |
|---------|-----------|-------|
| `setBreakpoints` | ✅ Yes | Replaces all breakpoints in file |
| `setFunctionBreakpoints` | ❌ No | - |
| `setExceptionBreakpoints` | ❌ No | - |
| `breakpointLocations` | ✅ Yes | Available |

---

## Path Validation

### Project Path Validation

Godot validates the project path in launch requests:

```cpp
if (args.has("project") && !is_valid_path(args["project"])) {
    return prepare_error_response(p_params, DAP::ErrorType::WRONG_PATH,
        "Project path is not valid");
}
```

**Requirements**:
- Must be absolute path
- Must contain `project.godot` file
- Must match editor's current project (in most cases)

**Common Errors**:
- `WRONG_PATH`: Path doesn't exist or doesn't contain project.godot
- Path mismatch: Symbolic links or case sensitivity issues

### File Path Formatting

**Best Practices**:
- **Always use absolute paths**
- **Always use forward slashes (`/`)**
- Windows: Uppercase drive letter (e.g., `C:/project/script.gd`)
- Match editor's project directory exactly

**Example**:
```json
{
  "source": {
    "path": "/Users/username/Projects/MyGame/player.gd"
  },
  "breakpoints": [
    { "line": 42 }
  ]
}
```

---

## Request Timeout

### Server-Side Timeout

Godot has built-in 5-second timeout for requests:

```cpp
// debug_adapter_protocol.cpp:860
if (OS::get_singleton()->get_ticks_msec() - _current_peer->timestamp > _request_timeout) {
    return prepare_error_response(params, DAP::ErrorType::TIMEOUT);
}
// _request_timeout = 5000 (5 seconds)
```

### Client-Side Timeout Recommendations

Despite server-side timeout, client-side timeouts (10-30s) are still needed for:
- Network issues
- Editor freezes
- Commands that don't respond at all

**Recommended Timeouts**:
- Connect: 10 seconds
- Commands: 30 seconds
- Read operations: 5 seconds

---

## Command Routing

Godot prepends `"req_"` to all commands internally:

```cpp
bool DebugAdapterProtocol::process_message(const String &p_text) {
    String command = "req_" + (String)params["command"];
    if (parser->has_method(command)) {
        // Calls parser->req_initialize(), parser->req_continue(), etc.
    }
}
```

**Impact**: All DAP commands are routed to `req_<command>` methods in `DebugAdapterParser`.

---

## Thread Model

Godot debugging uses a **single-thread model**:

```json
{
  "threads": [
    { "id": 1, "name": "Main" }
  ]
}
```

**Always**:
- Exactly 1 thread
- Thread ID is always `1`
- Thread name is always `"Main"`

---

## Scopes Model

Godot always returns **exactly 3 scopes**:

```json
{
  "scopes": [
    { "name": "Locals", "variablesReference": 1000, "expensive": false },
    { "name": "Members", "variablesReference": 1001, "expensive": false },
    { "name": "Globals", "variablesReference": 1002, "expensive": false }
  ]
}
```

**Scope Descriptions**:
- **Locals**: Function-local variables
- **Members**: Instance/class members (properties)
- **Globals**: Global variables and autoloads

---

## References

For complete protocol specification:
- [Microsoft DAP Specification](https://microsoft.github.io/debug-adapter-protocol/)
- [Godot DAP FAQ](GODOT_DAP_FAQ.md) - Common questions and troubleshooting
- [Godot Source Analysis](GODOT_SOURCE_ANALYSIS.md) - Additional findings
