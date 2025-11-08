# DAP MCP Test Project

This is a minimal Godot project for testing the godot-dap-mcp-server.

## Project Structure

- `project.godot` - Godot project configuration
- `test_scene.tscn` - Simple scene with test script attached
- `test_script.gd` - GDScript with code for debugging

## Test Script Details

The `test_script.gd` file contains:
- `calculate_sum(a, b)` - Function with breakpoint on line 13
- `test_loop()` - Loop with breakpoint on line 18

These functions are called when the scene runs, providing predictable locations for testing breakpoints and stepping.

## How to Use

### Setup
1. Open this project in Godot editor
2. Enable DAP server in Godot:
   - Go to: **Editor → Editor Settings → Network → Debug Adapter**
   - Check **Enable Debug Adapter**
   - Set **Remote Port** to `6006`
   - Click **Start Server**

### Manual Testing
1. Run the integration test: `./scripts/integration-test.sh`
2. The script will test connecting and setting breakpoints
3. Press F5 in Godot to run the scene
4. Scene will pause at the breakpoint

### Debugging Workflow
1. Connect: `godot_connect()`
2. Set breakpoint: `godot_set_breakpoint(file="res://test_script.gd", line=13)`
3. Run scene in Godot (F5)
4. When breakpoint hits, use stepping commands:
   - `godot_continue()` - Resume execution
   - `godot_step_over()` - Step over current line
   - `godot_step_into()` - Step into function call

## Breakpoint Locations

Good locations to test breakpoints:
- **Line 13**: `var sum: int = a + b` - Inside calculate_sum function
- **Line 18**: `print("Loop iteration: ", i)` - Inside test_loop

## Expected Output

When running the scene normally (without breakpoints):
```
Test script starting...
Result: 15
Loop iteration: 0
  Squared: 0
Loop iteration: 1
  Squared: 1
Loop iteration: 2
  Squared: 4
Test script finished!
```

When stopped at breakpoint on line 13:
- Execution pauses before calculating sum
- You can inspect variables `a` and `b`
- You can step through the function

## Godot Version

This project is configured for Godot 4.3+
