package dap

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-dap"
)

// EventHandler is called when an event is received
type EventHandler func(event dap.Event)

// Event storage for async events
type eventStore struct {
	events []dap.Event
}

// Client event handling fields
func (c *Client) initEventHandling() {
	// Initialize event handling if needed
	// This will be called during client creation
}

// waitForResponse reads messages until it finds the expected response
// This is the critical event filtering pattern that prevents hangs
func (c *Client) waitForResponse(ctx context.Context, command string) (dap.Message, error) {
	for {
		// Read with timeout to prevent permanent hangs
		msg, err := c.ReadWithTimeout(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to read response for %s: %w", command, err)
		}

		// Handle different message types
		switch m := msg.(type) {
		case *dap.Response:
			// Check if this is the response we're waiting for
			if m.Command == command {
				// Check if the response indicates success
				if !m.Success {
					return nil, fmt.Errorf("command %s failed: %s", command, m.Message)
				}
				return m, nil
			}
			// If it's a different response, log and continue waiting
			log.Printf("Received unexpected response for command: %s (waiting for %s)", m.Command, command)
			continue

		case *dap.Event:
			// Log the event and continue waiting for the response
			c.logEvent(m)
			continue

		case *dap.InitializeResponse:
			// Special case: InitializeResponse is the actual type we want for initialize
			if command == "initialize" {
				return m, nil
			}
			log.Printf("Received unexpected InitializeResponse (waiting for %s)", command)
			continue

		case *dap.ConfigurationDoneResponse:
			// Special case: ConfigurationDoneResponse
			if command == "configurationDone" {
				return m, nil
			}
			log.Printf("Received unexpected ConfigurationDoneResponse (waiting for %s)", command)
			continue

		case *dap.LaunchResponse:
			// Special case: LaunchResponse
			if command == "launch" {
				return m, nil
			}
			log.Printf("Received unexpected LaunchResponse (waiting for %s)", command)
			continue

		case *dap.SetBreakpointsResponse:
			// Special case: SetBreakpointsResponse
			if command == "setBreakpoints" {
				return m, nil
			}
			log.Printf("Received unexpected SetBreakpointsResponse (waiting for %s)", command)
			continue

		case *dap.ContinueResponse:
			// Special case: ContinueResponse
			if command == "continue" {
				return m, nil
			}
			log.Printf("Received unexpected ContinueResponse (waiting for %s)", command)
			continue

		case *dap.NextResponse:
			// Special case: NextResponse
			if command == "next" {
				return m, nil
			}
			log.Printf("Received unexpected NextResponse (waiting for %s)", command)
			continue

		case *dap.StepInResponse:
			// Special case: StepInResponse
			if command == "stepIn" {
				return m, nil
			}
			log.Printf("Received unexpected StepInResponse (waiting for %s)", command)
			continue

		case *dap.StepOutResponse:
			// Special case: StepOutResponse
			if command == "stepOut" {
				return m, nil
			}
			log.Printf("Received unexpected StepOutResponse (waiting for %s)", command)
			continue

		case *dap.ThreadsResponse:
			// Special case: ThreadsResponse
			if command == "threads" {
				return m, nil
			}
			log.Printf("Received unexpected ThreadsResponse (waiting for %s)", command)
			continue

		case *dap.StackTraceResponse:
			// Special case: StackTraceResponse
			if command == "stackTrace" {
				return m, nil
			}
			log.Printf("Received unexpected StackTraceResponse (waiting for %s)", command)
			continue

		case *dap.ScopesResponse:
			// Special case: ScopesResponse
			if command == "scopes" {
				return m, nil
			}
			log.Printf("Received unexpected ScopesResponse (waiting for %s)", command)
			continue

		case *dap.VariablesResponse:
			// Special case: VariablesResponse
			if command == "variables" {
				return m, nil
			}
			log.Printf("Received unexpected VariablesResponse (waiting for %s)", command)
			continue

		case *dap.EvaluateResponse:
			// Special case: EvaluateResponse
			if command == "evaluate" {
				return m, nil
			}
			log.Printf("Received unexpected EvaluateResponse (waiting for %s)", command)
			continue

		case *dap.DisconnectResponse:
			// Special case: DisconnectResponse
			if command == "disconnect" {
				return m, nil
			}
			log.Printf("Received unexpected DisconnectResponse (waiting for %s)", command)
			continue

		default:
			// Unknown message type, log and continue
			log.Printf("Received unknown message type: %T (waiting for %s)", msg, command)
			continue
		}
	}
}

// logEvent logs a DAP event
func (c *Client) logEvent(event interface{}) {
	switch e := event.(type) {
	case *dap.InitializedEvent:
		log.Printf("[DAP Event] Initialized")
	case *dap.StoppedEvent:
		log.Printf("[DAP Event] Stopped: reason=%s, threadId=%d", e.Body.Reason, e.Body.ThreadId)
	case *dap.ContinuedEvent:
		log.Printf("[DAP Event] Continued: threadId=%d", e.Body.ThreadId)
	case *dap.ExitedEvent:
		log.Printf("[DAP Event] Exited: exitCode=%d", e.Body.ExitCode)
	case *dap.TerminatedEvent:
		log.Printf("[DAP Event] Terminated")
	case *dap.ThreadEvent:
		log.Printf("[DAP Event] Thread: reason=%s, threadId=%d", e.Body.Reason, e.Body.ThreadId)
	case *dap.OutputEvent:
		log.Printf("[DAP Event] Output: category=%s, output=%s", e.Body.Category, e.Body.Output)
	case *dap.BreakpointEvent:
		log.Printf("[DAP Event] Breakpoint: reason=%s", e.Body.Reason)
	case *dap.ModuleEvent:
		log.Printf("[DAP Event] Module: reason=%s", e.Body.Reason)
	case *dap.LoadedSourceEvent:
		log.Printf("[DAP Event] LoadedSource: reason=%s", e.Body.Reason)
	case *dap.ProcessEvent:
		log.Printf("[DAP Event] Process: name=%s", e.Body.Name)
	case *dap.CapabilitiesEvent:
		log.Printf("[DAP Event] Capabilities")
	default:
		log.Printf("[DAP Event] Unknown event type: %T", event)
	}
}
