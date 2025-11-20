# Godot DAP Session Guide

**Purpose**: Complete reference for Godot's Debug Adapter Protocol implementation, showing all supported commands in the context of typical debugging sessions.

**Audience**: DAP client implementers, MCP tool developers, and contributors working on Godot's DAP server.

---

## Table of Contents

1. [Command Categories](#command-categories)
2. [Typical Session Flow](#typical-session-flow)
3. [Complete Session Example](#complete-session-example)
4. [Command Reference by Category](#command-reference-by-category)
5. [Not Supported: stepOut](#not-supported-stepout)

---

## Command Categories

Godot's DAP server supports **20+ commands** organized into 5 categories:

### 1. Session Management (6 commands)
Commands for establishing and managing the DAP connection:
- `initialize` - Start session, negotiate capabilities
- `configurationDone` - Signal ready (triggers launch)
- `disconnect` - End session
- `terminate` - Stop running game
- `restart` - Restart debugging
- `attach` - Attach to running game

### 2. Launch & Scene Control (1 command)
Commands for starting the game:
- `launch` - Launch game with parameters (scene, platform, options)

### 3. Breakpoints (2 commands)
Commands for managing breakpoints:
- `setBreakpoints` - Set/update/clear breakpoints for a file
- `breakpointLocations` - Query valid breakpoint locations

### 4. Execution Control (4 commands)
Commands for controlling execution flow:
- `continue` - Resume execution
- `pause` - Pause execution
- `next` - Step over (don't enter functions)
- `stepIn` - Step into function calls
- ⚠️ `stepOut` - **NOT SUPPORTED** (see below)

### 5. Runtime Inspection (5 commands)
Commands for inspecting game state:
- `threads` - Get thread list (always returns 1 thread)
- `stackTrace` - Get call stack when paused
- `scopes` - Get variable scopes (Locals, Members, Globals)
- `variables` - Get variables in scope or expand complex types
- `evaluate` - Evaluate GDScript expressions

### 6. Custom Godot Extensions (1 command)
Non-standard commands:
- `godot/put_msg` - Send custom messages to debugger

---

## Typical Session Flow

A typical Godot debugging session follows this sequence:

```
┌─────────────────────────────────────────────────────────────┐
│ Phase 1: Connection & Initialization                        │
└─────────────────────────────────────────────────────────────┘

1. Client connects to localhost:6006
2. Client → initialize (with capabilities)
   Server → initialize response (server capabilities)
   Server → initialized event
3. Client → configurationDone ⚠️ REQUIRED STEP
   Server → configurationDone response

┌─────────────────────────────────────────────────────────────┐
│ Phase 2: Breakpoint Setup (Optional)                        │
└─────────────────────────────────────────────────────────────┘

4. Client → setBreakpoints (for each file)
   Server → setBreakpoints response (verified breakpoints)
   Server → breakpoint events (one per breakpoint)

┌─────────────────────────────────────────────────────────────┐
│ Phase 3: Launch Game                                        │
└─────────────────────────────────────────────────────────────┘

5. Client → launch (scene, project, options)
   Server → launch response (empty - stored for later)
6. Client → configurationDone ⚠️ TRIGGERS ACTUAL LAUNCH
   Server → configurationDone response
   Server → process event (game started)
   Game runs...

┌─────────────────────────────────────────────────────────────┐
│ Phase 4: Debugging Loop                                     │
└─────────────────────────────────────────────────────────────┘

[Game hits breakpoint or user pauses]

7. Server → stopped event (reason: "breakpoint" | "pause" | "step")

8. Client → threads
   Server → threads response (threadId: 1)

9. Client → stackTrace (threadId: 1)
   Server → stackTrace response (stack frames)

10. Client → scopes (frameId: N)
    Server → scopes response (Locals, Members, Globals)

11. Client → variables (variablesReference: N)
    Server → variables response (variable list)

12. [Optional] Client → evaluate (expression, frameId)
    Server → evaluate response (result)

13. Client → continue | next | stepIn
    Server → continue/next/stepIn response
    [Later] Server → continued event
    Game resumes...

[Repeat debugging loop as needed]

┌─────────────────────────────────────────────────────────────┐
│ Phase 5: Termination                                        │
└─────────────────────────────────────────────────────────────┘

14. Client → terminate (or user stops game)
    Server → terminate response
    Server → terminated event

15. Client → disconnect
    Server → disconnect response
    Connection closes
```

---

## Complete Session Example

This example shows a minimal debugging session with actual JSON messages:

### 1. Initialize Session

**Request**:
```json
{
  "seq": 1,
  "type": "request",
  "command": "initialize",
  "arguments": {
    "clientID": "godot-dap-mcp-server",
    "clientName": "Godot DAP MCP Server",
    "adapterID": "godot",
    "linesStartAt1": true,
    "columnsStartAt1": true,
    "supportsVariableType": true,
    "supportsInvalidatedEvent": true
  }
}
```

**Response**:
```json
{
  "seq": 1,
  "type": "response",
  "request_seq": 1,
  "success": true,
  "command": "initialize",
  "body": {
    "supportsConfigurationDoneRequest": true,
    "supportsEvaluateForHovers": true,
    "supportsSetVariable": true,
    "supportedChecksumAlgorithms": ["MD5", "SHA1", "SHA256"],
    "supportsRestartRequest": true,
    "supportsValueFormattingOptions": true,
    "supportTerminateDebuggee": true,
    "supportSuspendDebuggee": true,
    "supportsTerminateRequest": true,
    "supportsBreakpointLocationsRequest": true
  }
}
```

**Event**:
```json
{
  "seq": 2,
  "type": "event",
  "event": "initialized"
}
```

### 2. Configuration Done (Required)

**Request**:
```json
{
  "seq": 2,
  "type": "request",
  "command": "configurationDone"
}
```

**Response**:
```json
{
  "seq": 3,
  "type": "response",
  "request_seq": 2,
  "success": true,
  "command": "configurationDone"
}
```

### 3. Set Breakpoints

**Request**:
```json
{
  "seq": 3,
  "type": "request",
  "command": "setBreakpoints",
  "arguments": {
    "source": {
      "path": "/Users/user/project/player.gd"
    },
    "breakpoints": [
      { "line": 10 },
      { "line": 25 }
    ]
  }
}
```

**Response**:
```json
{
  "seq": 4,
  "type": "response",
  "request_seq": 3,
  "success": true,
  "command": "setBreakpoints",
  "body": {
    "breakpoints": [
      {
        "id": 1,
        "verified": true,
        "line": 10,
        "source": {
          "path": "/Users/user/project/player.gd",
          "checksums": [
            {
              "algorithm": "MD5",
              "checksum": "a1b2c3d4e5f6..."
            }
          ]
        }
      },
      {
        "id": 2,
        "verified": true,
        "line": 25,
        "source": {
          "path": "/Users/user/project/player.gd",
          "checksums": [
            {
              "algorithm": "MD5",
              "checksum": "a1b2c3d4e5f6..."
            }
          ]
        }
      }
    ]
  }
}
```

**Events**:
```json
{
  "seq": 5,
  "type": "event",
  "event": "breakpoint",
  "body": {
    "reason": "new",
    "breakpoint": {
      "id": 1,
      "verified": true,
      "line": 10,
      "source": { "path": "/Users/user/project/player.gd" }
    }
  }
}
```
```json
{
  "seq": 6,
  "type": "event",
  "event": "breakpoint",
  "body": {
    "reason": "new",
    "breakpoint": {
      "id": 2,
      "verified": true,
      "line": 25,
      "source": { "path": "/Users/user/project/player.gd" }
    }
  }
}
```

### 4. Launch Game

**Request**:
```json
{
  "seq": 4,
  "type": "request",
  "command": "launch",
  "arguments": {
    "project": "/Users/user/project",
    "scene": "main",
    "platform": "host",
    "noDebug": false
  }
}
```

**Response**:
```json
{
  "seq": 7,
  "type": "response",
  "request_seq": 4,
  "success": true,
  "command": "launch"
}
```

**Note**: Launch request is stored but not executed yet.

### 5. Configuration Done (Triggers Launch)

**Request**:
```json
{
  "seq": 5,
  "type": "request",
  "command": "configurationDone"
}
```

**Response**:
```json
{
  "seq": 8,
  "type": "response",
  "request_seq": 5,
  "success": true,
  "command": "configurationDone"
}
```

**Event** (Game Started):
```json
{
  "seq": 9,
  "type": "event",
  "event": "process",
  "body": {
    "name": "Godot Game",
    "systemProcessId": 12345,
    "startMethod": "launch"
  }
}
```

### 6. Game Hits Breakpoint

**Event**:
```json
{
  "seq": 10,
  "type": "event",
  "event": "stopped",
  "body": {
    "reason": "breakpoint",
    "threadId": 1,
    "allThreadsStopped": true
  }
}
```

### 7. Inspect State - Get Threads

**Request**:
```json
{
  "seq": 6,
  "type": "request",
  "command": "threads"
}
```

**Response**:
```json
{
  "seq": 11,
  "type": "response",
  "request_seq": 6,
  "success": true,
  "command": "threads",
  "body": {
    "threads": [
      { "id": 1, "name": "Main" }
    ]
  }
}
```

### 8. Get Stack Trace

**Request**:
```json
{
  "seq": 7,
  "type": "request",
  "command": "stackTrace",
  "arguments": {
    "threadId": 1,
    "startFrame": 0,
    "levels": 20
  }
}
```

**Response**:
```json
{
  "seq": 12,
  "type": "response",
  "request_seq": 7,
  "success": true,
  "command": "stackTrace",
  "body": {
    "stackFrames": [
      {
        "id": 0,
        "name": "calculate_sum",
        "source": {
          "path": "/Users/user/project/test_script.gd",
          "name": "test_script.gd",
          "checksums": [
            {
              "algorithm": "MD5",
              "checksum": "187c6f2eb1e8016309f0c8875f1a2061"
            },
            {
              "algorithm": "SHA256",
              "checksum": "5baba6b7784def4d93dfd08e53aa306159cf0d22edaa323da1e7ca8eadbd888c"
            }
          ]
        },
        "line": 15,
        "column": 0
      },
      {
        "id": 1,
        "name": "_ready",
        "source": {
          "path": "/Users/user/project/test_script.gd",
          "name": "test_script.gd",
          "checksums": [
            {
              "algorithm": "MD5",
              "checksum": "187c6f2eb1e8016309f0c8875f1a2061"
            },
            {
              "algorithm": "SHA256",
              "checksum": "5baba6b7784def4d93dfd08e53aa306159cf0d22edaa323da1e7ca8eadbd888c"
            }
          ]
        },
        "line": 6,
        "column": 0
      }
    ],
    "totalFrames": 2
  }
}
```

**Key observations**:
- **Checksums**: Godot includes both MD5 and SHA256 checksums in source objects (DAP optional feature)
- **Frame ordering**: Innermost frame first (current function at index 0)
- **Frame IDs**: Sequential numbering (0, 1, 2...) used for subsequent scopes/variables requests
- **Column position**: Always 0 for GDScript (language doesn't track column positions)
- **Use case**: Call stackTrace after stepping commands (stepIn, next) to verify execution location

### 9. Get Variable Scopes

**Request**:
```json
{
  "seq": 8,
  "type": "request",
  "command": "scopes",
  "arguments": {
    "frameId": 1
  }
}
```

**Response**:
```json
{
  "seq": 13,
  "type": "response",
  "request_seq": 8,
  "success": true,
  "command": "scopes",
  "body": {
    "scopes": [
      {
        "name": "Locals",
        "presentationHint": "locals",
        "variablesReference": 1000,
        "expensive": false
      },
      {
        "name": "Members",
        "presentationHint": "members",
        "variablesReference": 1001,
        "expensive": false
      },
      {
        "name": "Globals",
        "presentationHint": "globals",
        "variablesReference": 1002,
        "expensive": false
      }
    ]
  }
}
```

### 10. Get Variables in Scope

**Request**:
```json
{
  "seq": 9,
  "type": "request",
  "command": "variables",
  "arguments": {
    "variablesReference": 1000
  }
}
```

**Response**:
```json
{
  "seq": 14,
  "type": "response",
  "request_seq": 9,
  "success": true,
  "command": "variables",
  "body": {
    "variables": [
      {
        "name": "health",
        "value": "100",
        "type": "int",
        "variablesReference": 0
      },
      {
        "name": "position",
        "value": "(10, 20)",
        "type": "Vector2",
        "variablesReference": 2000
      }
    ]
  }
}
```

### 11. Evaluate Expression

**Request**:
```json
{
  "seq": 10,
  "type": "request",
  "command": "evaluate",
  "arguments": {
    "expression": "health * 2",
    "frameId": 1,
    "context": "watch"
  }
}
```

**Response**:
```json
{
  "seq": 15,
  "type": "response",
  "request_seq": 10,
  "success": true,
  "command": "evaluate",
  "body": {
    "result": "200",
    "type": "int",
    "variablesReference": 0
  }
}
```

### 12. Continue Execution

**Request**:
```json
{
  "seq": 11,
  "type": "request",
  "command": "continue",
  "arguments": {
    "threadId": 1
  }
}
```

**Response**:
```json
{
  "seq": 16,
  "type": "response",
  "request_seq": 11,
  "success": true,
  "command": "continue",
  "body": {
    "allThreadsContinued": true
  }
}
```

**Event** (Later):
```json
{
  "seq": 17,
  "type": "event",
  "event": "continued",
  "body": {
    "threadId": 1,
    "allThreadsContinued": true
  }
}
```

### 13. Terminate Game

**Request**:
```json
{
  "seq": 12,
  "type": "request",
  "command": "terminate"
}
```

**Response**:
```json
{
  "seq": 18,
  "type": "response",
  "request_seq": 12,
  "success": true,
  "command": "terminate"
}
```

**Event**:
```json
{
  "seq": 19,
  "type": "event",
  "event": "terminated"
}
```

### 14. Disconnect

**Request**:
```json
{
  "seq": 13,
  "type": "request",
  "command": "disconnect",
  "arguments": {
    "restart": false,
    "terminateDebuggee": false
  }
}
```

**Response**:
```json
{
  "seq": 20,
  "type": "response",
  "request_seq": 13,
  "success": true,
  "command": "disconnect"
}
```

---

## Command Reference by Category

### Session Management

#### initialize

**Purpose**: Start DAP session and negotiate capabilities

**When to use**: First command after connecting to port 6006

**Request**:
```json
{
  "command": "initialize",
  "arguments": {
    "clientID": "your-client-id",
    "clientName": "Your Client Name",
    "adapterID": "godot",
    "linesStartAt1": true,
    "columnsStartAt1": true,
    "supportsVariableType": true,
    "supportsInvalidatedEvent": true
  }
}
```

**Response**: Server capabilities object

**Key capabilities**:
- `supportsConfigurationDoneRequest: true` - REQUIRED for launch
- `supportsSetVariable: true` - Can modify variables at runtime
- `supportsRestartRequest: true` - Can restart debugging
- `supportedChecksumAlgorithms: ["MD5", "SHA1", "SHA256"]` - File verification

**Notes**:
- Must be first command
- Server sends `initialized` event after response
- `linesStartAt1` affects all subsequent line numbers

---

#### configurationDone

**Purpose**: Signal configuration complete; triggers launch if pending

**When to use**:
1. Immediately after `initialize` (required for proper initialization)
2. After `launch` to actually start the game

**Request**:
```json
{
  "command": "configurationDone"
}
```

**Response**: Success response

**Critical behavior**:
- First call: Marks session as "configured" (enables breakpoint commands)
- Second call (after `launch`): Executes stored launch request, starts game

**Common mistake**: Forgetting this command causes:
- Breakpoint commands timeout
- Launch request never executes

---

#### disconnect

**Purpose**: End DAP session

**When to use**: When closing debugging session

**Request**:
```json
{
  "command": "disconnect",
  "arguments": {
    "restart": false,
    "terminateDebuggee": false
  }
}
```

**Response**: Success response

**Parameters**:
- `restart`: Whether disconnecting to restart (optional)
- `terminateDebuggee`: Whether to stop running game (optional)

---

#### terminate

**Purpose**: Stop running game

**When to use**: To stop game while keeping DAP session open

**Request**:
```json
{
  "command": "terminate"
}
```

**Response**: Success response

**Event**: Sends `terminated` event

**Notes**:
- Stops game via `EditorRunBar::stop_playing()`
- Session remains active (can launch again)

---

#### restart

**Purpose**: Restart debugging session with same launch parameters

**When to use**: To restart game without reconfiguring

**Request**:
```json
{
  "command": "restart"
}
```

**Response**: Success response

**Notes**:
- Uses stored launch parameters from previous `launch` request
- Equivalent to: terminate → launch (same params) → configurationDone

---

#### attach

**Purpose**: Attach to already-running game

**When to use**: Game launched manually in editor, want to debug it

**Request**:
```json
{
  "command": "attach"
}
```

**Response**: Success response

**Error**: `NOT_RUNNING` if no active game session

**Notes**:
- Requires game already running
- Checks `ScriptEditorDebugger::is_session_active()`

---

### Launch & Scene Control

#### launch

**Purpose**: Launch game with debugging enabled

**When to use**: After setting breakpoints, before `configurationDone`

**Request**:
```json
{
  "command": "launch",
  "arguments": {
    "project": "/absolute/path/to/project",
    "scene": "main",
    "platform": "host",
    "device": -1,
    "noDebug": false,
    "playArgs": ["--arg1", "value1"]
  }
}
```

**Response**: Empty success response

**Parameters**:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `project` | string | Current project | Absolute path (validated) |
| `scene` | string | `"main"` | Scene mode (see below) |
| `platform` | string | `"host"` | Platform (host/android/web) |
| `device` | integer | `-1` | Android device index |
| `noDebug` | boolean | `false` | Skip breakpoints |
| `playArgs` | array | `[]` | CLI arguments |

**Scene modes**:
- `"main"` - Launch main scene from project.godot
- `"current"` - Launch currently open scene in editor
- `"res://path/to/scene.tscn"` - Launch specific scene

**Platform modes**:
- `"host"` - Run on local machine (default)
- `"android"` - Deploy to Android device
- `"web"` - Run in browser

**Errors**:
- `WRONG_PATH` - Project path doesn't match editor
- `UNKNOWN_PLATFORM` - Invalid platform
- `MISSING_DEVICE` - Android device not found

**Critical behavior**:
- Launch request is **stored**, not executed
- Actual launch happens when `configurationDone` is sent
- This allows setting breakpoints before game starts

---

### Breakpoints

#### setBreakpoints

**Purpose**: Set or update breakpoints for a source file

**When to use**: Before launch, or while paused

**Request**:
```json
{
  "command": "setBreakpoints",
  "arguments": {
    "source": {
      "path": "/absolute/path/to/file.gd"
    },
    "breakpoints": [
      { "line": 10 },
      { "line": 25 }
    ]
  }
}
```

**Response**: Verified breakpoints with IDs and checksums

**Clearing breakpoints**:
```json
{
  "breakpoints": []  // Empty array clears all
}
```

**Notes**:
- **Replaces** all breakpoints in file (not incremental)
- Must use absolute paths
- Line numbers adjusted based on `linesStartAt1` from initialize
- Sends `breakpoint` event for each breakpoint

**Path handling**:
- Windows: `C:\path` → `C:/path` (forward slashes, uppercase drive)
- Must match editor's project path exactly

---

#### breakpointLocations

**Purpose**: Query valid breakpoint locations (for UI hints)

**When to use**: To show user where breakpoints can be set

**Request**:
```json
{
  "command": "breakpointLocations",
  "arguments": {
    "source": { "path": "/path/to/file.gd" },
    "line": 10,
    "endLine": 20
  }
}
```

**Response**: Array of valid locations

**Notes**:
- Godot's implementation returns requested line as-is
- Doesn't validate against actual code

---

### Execution Control

#### continue

**Purpose**: Resume execution from breakpoint/pause

**When to use**: After inspecting state at breakpoint

**Request**:
```json
{
  "command": "continue",
  "arguments": {
    "threadId": 1
  }
}
```

**Response**:
```json
{
  "body": {
    "allThreadsContinued": true
  }
}
```

**Event**: `continued` event sent asynchronously (later)

**Notes**:
- Returns immediately (async operation)
- Don't block waiting for `continued` event

---

#### pause

**Purpose**: Pause execution

**When to use**: To inspect running game

**Request**:
```json
{
  "command": "pause",
  "arguments": {
    "threadId": 1
  }
}
```

**Response**: Success response

**Event**: `stopped` event with reason "pause"

**Notes**:
- Calls `ScriptEditorDebugger::debug_break()`
- Game pauses at current execution point

---

#### next (Step Over)

**Purpose**: Execute current line, don't enter function calls

**When to use**: Stepping through code at same level

**Request**:
```json
{
  "command": "next",
  "arguments": {
    "threadId": 1,
    "granularity": "statement"
  }
}
```

**Response**: Success response

**Event**: `stopped` event with reason "step" when complete

**Notes**:
- Returns immediately (async)
- Calls `ScriptEditorDebugger::debug_next()`
- Sets `_stepping = true` internally

---

#### stepIn (Step Into)

**Purpose**: Step into function call

**When to use**: Want to debug inside called function

**Request**:
```json
{
  "command": "stepIn",
  "arguments": {
    "threadId": 1,
    "granularity": "statement"
  }
}
```

**Response**: Success response

**Event**: `stopped` event with reason "step" when complete

**Notes**:
- Returns immediately (async)
- Calls `ScriptEditorDebugger::debug_step()`
- If current line has no function call, behaves like `next`

---

### Runtime Inspection

#### threads

**Purpose**: Get list of threads

**When to use**: After `stopped` event, before `stackTrace`

**Request**:
```json
{
  "command": "threads"
}
```

**Response**:
```json
{
  "body": {
    "threads": [
      { "id": 1, "name": "Main" }
    ]
  }
}
```

**Notes**:
- Godot always returns single thread (ID: 1, Name: "Main")
- All debugging happens on this thread

---

#### stackTrace

**Purpose**: Get call stack when paused

**When to use**:
- After `stopped` event, to see execution context
- After stepping commands (stepIn, next) to verify location
- To discover current function and call hierarchy

**Request**:
```json
{
  "command": "stackTrace",
  "arguments": {
    "threadId": 1,
    "startFrame": 0,
    "levels": 20
  }
}
```

**Response**: Array of stack frames with source locations

**Response structure**:
- `stackFrames[0]` - Current execution point (innermost frame)
- `stackFrames[1..n]` - Caller chain (outer frames)
- Each frame includes:
  - `id` - Frame identifier for scopes/variables requests
  - `name` - Function/method name
  - `source` - File path with checksums (MD5 + SHA256)
  - `line` - Current line number (1-based if `linesStartAt1: true`)
  - `column` - Always 0 for GDScript

**Notes**:
- Must be paused (at breakpoint or after step)
- Frame IDs used in subsequent `scopes` requests
- `totalFrames` shows complete stack depth
- **Godot includes checksums**: Both MD5 and SHA256 for source file verification
- **Verification pattern**: Use stackTrace immediately after stepIn to confirm you entered the expected function

---

#### scopes

**Purpose**: Get variable scopes for stack frame

**When to use**: After `stackTrace`, before inspecting variables

**Request**:
```json
{
  "command": "scopes",
  "arguments": {
    "frameId": 1
  }
}
```

**Response**: Always exactly 3 scopes:
1. **Locals** - Function-local variables
2. **Members** - Instance/class members
3. **Globals** - Global variables and autoloads

**Notes**:
- `variablesReference` used in `variables` request
- All scopes returned even if empty

---

#### variables

**Purpose**: Get variables in scope or expand complex types

**When to use**:
- After `scopes` to see variables in scope
- To expand complex types (Vector2, Object, Array, etc.)

**Request**:
```json
{
  "command": "variables",
  "arguments": {
    "variablesReference": 1000
  }
}
```

**Response**: Array of variables with values and types

**Variable expansion**:
- Primitive types: `variablesReference = 0` (leaf node)
- Complex types: `variablesReference > 0` (can expand)

**Expandable types**:
- Vector2/3: x, y, z components
- Object: properties and methods
- Array: indexed elements
- Dictionary: key-value pairs

**Notes**:
- Hierarchical navigation (expand tree)
- May wait for debugger if more data needed

---

#### evaluate

**Purpose**: Evaluate GDScript expression in current context

**When to use**:
- Watch expressions
- REPL/console
- Hover evaluations

**Request**:
```json
{
  "command": "evaluate",
  "arguments": {
    "expression": "player.health * 2",
    "frameId": 1,
    "context": "watch"
  }
}
```

**Response**:
```json
{
  "body": {
    "result": "200",
    "type": "int",
    "variablesReference": 0
  }
}
```

**Context values**:
- `"watch"` - Watch expression
- `"repl"` - REPL/console
- `"hover"` - Hover tooltip

**Notes**:
- Evaluates in context of specified stack frame
- Can access locals, members, globals
- Can modify game state (side effects!)
- Results cached temporarily (volatile)

---

## Not Supported: stepOut

⚠️ **IMPORTANT**: Godot's DAP server does **NOT** implement the `stepOut` command (as of 4.5.1).

**Status Update**: A PR has been submitted to add stepOut support to the 4.x branch: https://github.com/godotengine/godot/pull/112875 (no milestone set yet).

**What is stepOut?**: Step out of current function (run until function returns)

**Why not supported?**: No `req_stepOut()` method in Godot's `DebugAdapterParser`

**Confirmed in**: `editor/debugger/debug_adapter/debug_adapter_parser.cpp`

**Workarounds** (until PR merges):
1. Set temporary breakpoint after function call, use `continue`
2. Step through remaining lines manually with `next`
3. Notify user that stepOut is unavailable

**Impact for clients**: Must handle gracefully (disable UI, show message)

---

## Events

Godot sends asynchronous events during debugging:

### initialized
**When**: After `initialize` response
**Meaning**: DAP session ready for configuration

### process
**When**: After game launches
**Meaning**: Game process started
**Body**: `{ name, systemProcessId, startMethod }`

### stopped
**When**: Execution pauses
**Reasons**:
- `"breakpoint"` - Hit breakpoint
- `"pause"` - User paused
- `"step"` - Step operation completed
- `"exception"` - Runtime error

**Body**: `{ reason, threadId, allThreadsStopped }`

### continued
**When**: Execution resumes (async after `continue` response)
**Body**: `{ threadId, allThreadsContinued }`

### output
**When**: Game produces console output
**Body**: `{ category, output }`
**Categories**: `"stdout"`, `"stderr"`, `"console"`

### breakpoint
**When**: Breakpoint added/changed/removed
**Body**: `{ reason, breakpoint }`
**Reasons**: `"new"`, `"changed"`, `"removed"`

### terminated
**When**: Game stops
**Meaning**: Debugging session ended

---

## Additional Notes

### Threading
- Godot always reports 1 thread (ID: 1, Name: "Main")
- All requests use `threadId: 1`
- No multi-threaded debugging support

### Path Handling
- Always use absolute paths (not `res://`)
- Always use forward slashes (`/`)
- Windows: Uppercase drive letter (e.g., `C:/project`)
- Paths must match editor's project directory exactly

### Line Numbers
- Client specifies indexing in `initialize`
- `linesStartAt1: true` → lines start at 1
- `linesStartAt1: false` → lines start at 0
- Godot adjusts internally

### Async Operations
- `continue`, `next`, `stepIn` return immediately
- Completion signaled by `continued` or `stopped` event
- Don't block waiting for events

### Error Handling
Common errors:
- `WRONG_PATH` - Path validation failed
- `NOT_RUNNING` - No active game (for attach)
- `UNKNOWN_PLATFORM` - Invalid platform
- `MISSING_DEVICE` - Android device not found

---

## References

- Official DAP Specification: https://microsoft.github.io/debug-adapter-protocol/specification
- Godot Source: `editor/debugger/debug_adapter/debug_adapter_parser.cpp`
- This project's DAP implementation: `internal/dap/`

---

**Last Updated**: 2025-11-11
