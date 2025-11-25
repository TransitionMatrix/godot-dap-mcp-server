# Godot DAP Event Handling Analysis

**Date:** 2025-11-24
**Status:** DRAFT
**Related Phase:** Phase 6 (Advanced Tools) / Phase 7 (Error Handling)

## Context

During the implementation of Phase 6 and compliance testing (`cmd/test-dap-protocol/`), we analyzed the behavior of Godot's DAP server under load and during standard debugging sessions. The log file (`test-dap-protocol` output) revealed critical insights into how Godot handles asynchronous events, which differs significantly from a simple request-response model.

This document captures those observations and outlines a strategy for a robust, event-driven DAP client architecture.

## Key Observations

### 1. Asynchronous & Interleaved Protocol
Godot's DAP server is highly asynchronous. Events are not just notifications; they are often the primary source of truth for the debuggee's state.

*   **Multiple Events per Command:** A single request often triggers multiple events.
    *   `configurationDone` triggers: `process` -> `launch` (response) -> `output` -> `stopped`.
    *   `stepIn` triggers: `continued` -> `stopped`.
*   **Interleaved Responses:** Events can arrive *before* the response to the command that triggered them.
*   **Chatty Output:** `OutputEvent` (stdout/stderr) can arrive in high volumes, potentially flooding the client if not managed.

### 2. "Stopped" is the Source of Truth
The client cannot rely on the successful return of a `pause`, `step`, or `launch` command to assume the game is paused.
*   **Observation:** The `stopped` event contains the critical `reason` (breakpoint, step, pause, exception) and `threadId`.
*   **Implication:** State transitions (Running -> Paused) must be driven by the `stopped` *event*, not the command response.

### 3. Termination Complexity
Terminating the debuggee triggers multiple events:
*   `ExitedEvent` (with exit code)
*   `TerminatedEvent`
*   If the client processes `OutputEvent`s sequentially without prioritization, it might delay detecting the termination, leading to "zombie" sessions.

## Strategy for Robust Event Handling

To address these behaviors, the DAP client architecture must evolve from a synchronous request-response model to a robust event-driven system.

### 1. Event Prioritization
**Problem:** High-volume `OutputEvent`s can block processing of critical control events like `Stopped` or `Terminated`.
**Solution:** Implement a **Priority Event Queue**.
*   **High Priority Queue:** `TerminatedEvent`, `ExitedEvent`, `StoppedEvent`. These effectively control the state machine.
*   **Standard Priority Queue:** `OutputEvent`, `ThreadEvent`, `ModuleEvent`, `process` event.
*   **Dispatcher:** Always drain the High Priority queue before processing Standard Priority events.

### 2. Event-Driven State Management
**Problem:** Current state machine transitions on *sending* commands, which is optimistic and sometimes incorrect.
**Solution:** Refine `SessionState` to transition based on *receiving* events.
*   `StateLaunched` -> `StatePaused`: Triggered **only** by `StoppedEvent`.
*   `StatePaused` -> `StateRunning`: Triggered by `ContinuedEvent` (or success of step/continue *if* validated).
*   `Any` -> `StateTerminated`: Triggered by `TerminatedEvent` / `ExitedEvent`.
*   **Data Invalidation:** `StoppedEvent` should automatically invalidate cached stack traces, variables, and scopes.

### 3. Error Handling & Recovery
**Problem:** Malformed events (e.g., missing body fields) or unexpected event types shouldn't crash the client.
**Solution:**
*   **Parser Safety:** Recover from panics during event unmarshalling. Log the error but keep the connection alive.
*   **Sequence Gap Detection:** Monitor `seq` numbers. Log warnings if gaps are detected (indicating dropped messages).

### 4. Performance Optimization (Output Throttling)
**Problem:** Reading every single stdout line as a separate JSONRPC notification is inefficient for the MCP host.
**Solution:** **Output Throttling / Debouncing**.
*   Buffer `OutputEvent`s.
*   Emit a batched notification to the MCP client (e.g., every 50ms or 100 lines).
*   This reduces the JSONRPC overhead significantly during heavy logging.

### 5. Testability (Mock DAP Server)
**Problem:** It is difficult to reproduce race conditions or specific event interleavings with the real Godot engine.
**Solution:** Create a **Mock DAP Server** (`pkg/daptest`).
*   A simple TCP server that speaks DAP.
*   **Scriptable Scenarios:**
    *   *Flood:* Send 1000 log messages followed by a stop event.
    *   *Race:* Send `stopped` event *before* `launch` response.
    *   *Malformed:* Send invalid JSON to test recovery.
*   This allows deterministic unit testing of the new event handling logic.

## Roadmap

1.  **Refactor `Session` State Machine:** Make it listen to events.
2.  **Implement Priority Queue:** update `client.readLoop`.
3.  **Implement Mock DAP Server:** for TDD of the new logic.
4.  **Verify with Integration Tests:** Ensure `godot_pause` and breakpoints still work robustly.
