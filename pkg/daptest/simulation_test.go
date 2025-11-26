package daptest

import (
	"context"
	"testing"
	"time"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/dap"
	godap "github.com/google/go-dap"
)

// TestGodotLaunchSimulation simulates the complex event flow during Godot launch
func TestGodotLaunchSimulation(t *testing.T) {
	// Start mock server
	server := NewServer(t)
	defer server.Close()

	// Connect our real DAP client to the mock server
	client := dap.NewClient("localhost", server.Port())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// 1. Client sends Initialize
	go func() {
		// Server handles Initialize
		msg, err := server.ExpectRequest("initialize")
		if err != nil {
			server.errors <- err // Report error
			return
		}
		req := msg.(*godap.InitializeRequest)
		
		// Send response
		server.Send(&godap.InitializeResponse{
			Response: godap.Response{
				ProtocolMessage: godap.ProtocolMessage{
					Seq:  server.NextSeq(),
					Type: "response",
				},
				RequestSeq: req.Seq,
				Success:    true,
				Command:    "initialize",
			},
		})
		
		// Send "initialized" event immediately after response (standard DAP)
		server.Send(&godap.InitializedEvent{
			Event: godap.Event{
				ProtocolMessage: godap.ProtocolMessage{
					Seq:  server.NextSeq(),
					Type: "event",
				},
				Event: "initialized",
			},
		})
	}()

	if _, err := client.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// 2. Client sends Launch
	go func() {
		msg, err := server.ExpectRequest("launch")
		if err != nil {
			server.errors <- err
			return
		}
		req := msg.(*godap.LaunchRequest)

		// Send response
		server.Send(&godap.LaunchResponse{
			Response: godap.Response{
				ProtocolMessage: godap.ProtocolMessage{
					Seq:  server.NextSeq(),
					Type: "response",
				},
				RequestSeq: req.Seq,
				Success:    true,
				Command:    "launch",
			},
		})
		
		// Godot sends "process" event
		server.Send(&godap.ProcessEvent{
			Event: godap.Event{
				ProtocolMessage: godap.ProtocolMessage{
					Seq:  server.NextSeq(),
					Type: "event",
				},
				Event: "process",
			},
			Body: godap.ProcessEventBody{
				Name: "Godot",
			},
		})
	}()

	// We need to call LaunchMainScene
	// But LaunchMainScene expects to read "project.godot" to validate.
	// We can bypass session logic and use client.Launch directly? 
	// Or we can mock the file system?
	// client.Launch just sends the request.
	
	// Let's use client directly for now to test the event handling
	launchArgs := map[string]interface{}{
		"project": "/tmp/test",
	}
	if _, err := client.Launch(ctx, launchArgs); err != nil {
		t.Fatalf("Launch failed: %v", err)
	}

	// 3. Client sends SetBreakpoints (optional, but usually happens here)
	// 4. Client sends ConfigurationDone
	go func() {
		msg, err := server.ExpectRequest("configurationDone")
		if err != nil {
			server.errors <- err
			return
		}
		req := msg.(*godap.ConfigurationDoneRequest)

		server.Send(&godap.ConfigurationDoneResponse{
			Response: godap.Response{
				ProtocolMessage: godap.ProtocolMessage{
					Seq:  server.NextSeq(),
					Type: "response",
				},
				RequestSeq: req.Seq,
				Success:    true,
				Command:    "configurationDone",
			},
		})
		
		// Godot then sends "stopped" event (entry point)
		// Wait a tiny bit to simulate async
		time.Sleep(10 * time.Millisecond)
		server.Send(&godap.StoppedEvent{
			Event: godap.Event{
				ProtocolMessage: godap.ProtocolMessage{
					Seq:  server.NextSeq(),
					Type: "event",
				},
				Event: "stopped",
			},
			Body: godap.StoppedEventBody{
				Reason:   "entry",
				ThreadId: 1,
			},
		})
	}()

	if err := client.ConfigurationDone(ctx); err != nil {
		t.Fatalf("ConfigurationDone failed: %v", err)
	}

	// Now wait for stopped event
	stopped, err := client.WaitForStop(ctx)
	if err != nil {
		t.Fatalf("WaitForStop failed: %v", err)
	}

	if stopped.Reason != "entry" {
		t.Errorf("Expected stop reason 'entry', got '%s'", stopped.Reason)
	}
}
