# Godot Debugging Examples

These examples demonstrate common workflows for using the Godot DAP MCP Server.

## Workflow 1: Basic Debugging

1. **Start**: Connect and launch the game.
   ```python
   godot_connect(project="/path/to/game")
   godot_launch_main_scene(project="/path/to/game")
   ```

2. **Pause**: Pause the game to inspect state.
   ```python
   godot_pause()
   ```

3. **Inspect**: Check where we are and what variables are set.
   ```python
   // Where are we?
   godot_get_stack_trace()
   
   // What are the local variables?
   // (Assume godot_get_scopes returned Locals ref=1000)
   godot_get_variables(variables_reference=1000)
   ```

4. **Resume**: Continue playing.
   ```python
   godot_continue()
   ```

## Workflow 2: Debugging with Breakpoints

1. **Setup**: Connect and set a breakpoint before launching.
   ```python
   godot_connect(project="/path/to/game")
   godot_set_breakpoint(file="res://scripts/player.gd", line=25)
   ```

2. **Launch**: Start the game.
   ```python
   godot_launch_main_scene(project="/path/to/game")
   ```

3. **Hit Breakpoint**: The game will pause automatically when execution hits line 25. You'll receive a "stopped" notification (if your client supports it) or you can poll/check status.

4. **Step**: Step through the code.
   ```python
   godot_step_over() // Step to line 26
   godot_step_over() // Step to line 27
   ```

5. **Inspect Complex Object**:
   ```python
   // Evaluate 'self' to get the player node
   result = godot_evaluate(expression="self")
   // Result has variablesReference (e.g., 2000)
   
   // Inspect properties of 'self'
   godot_get_variables(variables_reference=2000)
   ```

## Workflow 3: Fixing a Bug

Scenario: Player health is not decreasing correctly.

1. **Set Breakpoint** in `take_damage` function:
   ```python
   godot_set_breakpoint(file="res://scripts/player.gd", line=40)
   ```

2. **Reproduce**: Play the game until player takes damage. Game pauses.

3. **Inspect Arguments**:
   ```python
   // Get stack trace to find frame ID (e.g., 0)
   godot_get_stack_trace()
   
   // Get Locals scope ID (e.g., 1001)
   godot_get_scopes(frame_id=0)
   
   // Inspect arguments (damage_amount)
   godot_get_variables(variables_reference=1001)
   ```

4. **Evaluate Logic**:
   ```python
   // Check calculation
   godot_evaluate(expression="health - damage_amount")
   ```

5. **Resume**:
   ```python
   godot_continue()
   ```

## Workflow 4: Attaching to a Running Game

Scenario: You have a game already running (e.g., launched via Godot Editor) and want to attach the debugger to it without restarting.

### Standard Method (GUI Editor)

1. **Launch Godot Editor**: Open your project in the Godot Editor.
2. **Run Game**: Press F5 (Play) to start the game.
3. **Attach**: Use the MCP server to attach.
   ```python
   godot_connect(project="/path/to/game")
   godot_attach()
   ```
4. **Debug**: You are now attached. You can pause, step, or inspect the game.

### Advanced Method (CLI / Headless)

Scenario: You want to debug a game running in a headless environment (CI/CD) or manually started via command line.

1. **Start Editor (Debugger Server)**:
   Start the editor in listening mode. This acts as the bridge between the game and the MCP server.
   ```bash
   # Standard (GUI)
   godot --editor --path /path/to/game --debug-server tcp://127.0.0.1:6007

   # Headless (No Window)
   godot --headless --editor --path /path/to/game --debug-server tcp://127.0.0.1:6007
   ```

2. **Start Game (Debuggee)**:
   Launch the game instance and tell it to connect to the debugger.
   ```bash
   godot --path /path/to/game --remote-debug tcp://127.0.0.1:6007
   ```

3. **Attach**:
   ```python
   godot_connect(project="/path/to/game")
   godot_attach()
   ```
