# Godot DAP MCP Server - Tool Reference

This document provides a complete reference for all available tools in the Godot DAP MCP Server.

## Connection Tools

### `godot_connect`
Establishes a connection to the Godot editor's DAP server.

**Parameters**:
- `port` (number, default: 6006): The DAP server port.
- `project` (string, optional): Absolute path to the project root. Enables `res://` path resolution.

**Example**:
```python
// Connect to default port with project context
godot_connect(project="/Users/me/my-game")
```

### `godot_disconnect`
Closes the DAP connection.

**Example**:
```python
godot_disconnect()
```

---

## Launch Tools

### `godot_launch_main_scene`
Launches the project's main scene (defined in `project.godot`).

**Parameters**:
- `project` (string, required): Absolute path to project directory.
- `no_debug` (boolean, default: false): Run without debugger attached.
- `profiling` (boolean, default: false): Enable performance profiling.
- `debug_collisions` (boolean, default: false): Visualize collision shapes.
- `debug_navigation` (boolean, default: false): Visualize navigation meshes.

**Example**:
```python
godot_launch_main_scene(project="/Users/me/my-game")
```

### `godot_launch_scene`
Launches a specific scene file.

**Parameters**:
- `project` (string, required): Absolute path to project directory.
- `scene` (string, required): Resource path (e.g., `res://scenes/level1.tscn`).
- ... (standard launch options)

**Example**:
```python
godot_launch_scene(project="/Users/me/my-game", scene="res://tests/test_unit.tscn")
```

### `godot_launch_current_scene`
Launches the scene currently open in the Godot editor.

**Parameters**:
- `project` (string, required): Absolute path to project directory.
- ... (standard launch options)

**Example**:
```python
godot_launch_current_scene(project="/Users/me/my-game")
```

---

## Breakpoint Tools

### `godot_set_breakpoint`
Sets a breakpoint at a specific line.

**Parameters**:
- `file` (string, required): Path to GDScript file (`res://` or absolute).
- `line` (number, required): Line number (1-based).

**Example**:
```python
godot_set_breakpoint(file="res://player.gd", line=15)
```

### `godot_clear_breakpoint`
Clears all breakpoints in a file.

**Parameters**:
- `file` (string, required): Path to GDScript file (`res://` or absolute).

**Example**:
```python
godot_clear_breakpoint(file="res://player.gd")
```

---

## Execution Control

### `godot_continue`
Resumes execution of the paused game.

**Example**:
```python
godot_continue()
```

### `godot_step_over`
Steps to the next line in the current function.

**Example**:
```python
godot_step_over()
```

### `godot_step_into`
Steps into a function call.

**Example**:
```python
godot_step_into()
```

### `godot_pause`
Pauses the running game.

**Example**:
```python
godot_pause()
```

---

## Inspection Tools

### `godot_get_stack_trace`
Gets the call stack for the paused game.

**Example**:
```python
godot_get_stack_trace()
```

### `godot_get_scopes`
Gets variable scopes (Locals, Members, Globals) for a stack frame.

**Parameters**:
- `frame_id` (number, required): Stack frame ID from `godot_get_stack_trace`.

**Example**:
```python
godot_get_scopes(frame_id=0)
```

### `godot_get_variables`
Gets variables in a scope or expands a complex variable/object.

**Parameters**:
- `variables_reference` (number, required): ID from `godot_get_scopes` or a variable.

**Example**:
```python
// Get local variables
godot_get_variables(variables_reference=1000)

// Expand an object (e.g., 'self')
godot_get_variables(variables_reference=2000)
```

### `godot_evaluate`
Evaluates a GDScript expression in the current context.

**Parameters**:
- `expression` (string, required): GDScript code to evaluate.
- `frame_id` (number, optional): Stack frame ID (default: 0).

**Example**:
```python
godot_evaluate(expression="player.health * 2")
```

### `godot_get_threads`
Lists active threads (Godot typically has one "Main" thread).

**Example**:
```python
godot_get_threads()
```

---

## Known Limitations

- **Set Variable**: `godot_set_variable` is currently disabled because Godot Engine does not implement the underlying DAP functionality (despite advertising support). We plan to submit a PR to Godot Engine to fix this.
- **Step Out**: `stepOut` is not currently supported by Godot Engine. A PR implementing this functionality has been submitted: [godotengine/godot#112875](https://github.com/godotengine/godot/pull/112875).
