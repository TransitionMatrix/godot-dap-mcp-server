package dap

import (
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
