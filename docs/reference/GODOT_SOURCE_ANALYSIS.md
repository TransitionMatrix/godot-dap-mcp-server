# Godot Source Analysis

This document captures findings from analyzing the upstream Godot engine source code, specifically focusing on the DAP implementation.

## 1. Missing `stepOut` Implementation

**Status**: Acknowledged bug / Missing feature.
**PR**: https://github.com/godotengine/godot/pull/112875 (Submitted to 4.x branch)

Godot's DAP server does not implement the `stepOut` command. While not explicitly advertised as supported in capabilities (it's optional), its absence limits debugging workflows. The PR implements `req_stepOut` by leveraging the engine's existing "step out" debugger command.

## 2. Missing `setVariable` Implementation (Confirmed)

**Status**: **False Advertising / Bug**.
**Advertised Capability**: `supportsSetVariable = true` (in `debug_adapter_types.h`)
**Implementation**: **None**.

### Analysis
- **Declaration**: `Capabilities` struct sets `supportsSetVariable = true`.
- **Handler**: No `req_setVariable` method exists in `DebugAdapterParser` class (`debug_adapter_parser.h/cpp`).
- **Impact**: DAP clients (like VSCode or our MCP server) see the capability and enable UI for variable editing, but calls to `setVariable` fail (likely with "unknown command" or similar).

### Workaround Failure (Evaluate Assignment)
- **Attempt**: Using `evaluate` request with an assignment expression (e.g., `var = val`).
- **Result**: Fails.
- **Cause**: The `evaluate` request uses Godot's `core/math/Expression` class.
- **Limitation**: The `Expression` class is designed for **evaluating mathematical/logic expressions**, not executing statements. Assignments in GDScript are **statements**, not expressions, so `Expression` fails to parse them (Error: `Expected '='`).

### Conclusion
There is **currently no supported way** to modify variable values via DAP in Godot. The feature is falsely advertised and unimplemented.

## 3. DAP Event Handling Behavior

**Status**: Complex / Chatty.

Godot's DAP server is highly asynchronous and chatty with events:
- **Interleaved Events**: Events (like `output`, `stopped`) often arrive interleaved with command responses.
- **Multiple Events**: Single commands (like `launch`) trigger multiple events (`process`, `initialized`, etc.).
- **Source of Truth**: The `stopped` event is the definitive source for the debuggee's state, not the response to a `pause` or `step` command.

See [DAP_EVENT_ANALYSIS.md](../../docs/implementation-notes/DAP_EVENT_ANALYSIS.md) for detailed analysis and strategy.