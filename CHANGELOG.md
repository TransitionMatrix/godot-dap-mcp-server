# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- **Mock DAP Server**: `pkg/daptest` for simulating Godot DAP behavior.
- **Event-Driven Architecture**: Refactored DAP client to handle asynchronous events and responses robustly.
- **Path Resolution**: `godot_connect` now accepts `project` path to enable automatic `res://` to absolute path conversion for breakpoints.
- **Error Handling**: Standardized "Problem/Context/Solution" error format across all tools.
- **Logging**: Configurable logging via `GODOT_MCP_LOG_FILE` (defaults to stderr).
- **Godot 4.x Support**: Verified with Godot 4.6.dev.

### Changed
- **`godot_set_variable`**: Disabled with an explanatory error message due to missing upstream implementation in Godot Engine.
- **Timeouts**: All DAP requests now have strict timeouts to prevent hangs.
- **Documentation**: Updated README, added TOOLS.md and EXAMPLES.md.

### Fixed
- **Event Interleaving**: Fixed race conditions where `process` or `output` events arriving during `launch` would cause timeouts or missed responses.
- **Launch Flow**: Correctly implemented the `Launch` -> `ConfigurationDone` two-step handshake required by Godot.

## [Phase 3] - 2025-11-07

### Added
- Core debugging tools: `godot_connect`, `godot_set_breakpoint`, `godot_continue`, etc.
- Automated integration test suite.

## [Phase 1-2] - 2025-11-06

### Added
- Initial MCP server implementation.
- Basic DAP client.
