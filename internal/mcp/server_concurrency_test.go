package mcp

import (
	"encoding/json"
	"io"
	"testing"
	"time"
)

// TestServer_Concurrency verifies that the server processes requests concurrently
func TestServer_Concurrency(t *testing.T) {
	// 1. Setup pipes for communication
	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()

	// 2. Create server with custom transport
	transport := NewTransportWithStreams(inReader, outWriter)
	server := NewServerWithTransport(transport)

	// 3. Register blocking tool
	blockCh := make(chan struct{})
	server.RegisterTool(Tool{
		Name: "blocking_tool",
		Handler: func(params map[string]interface{}) (interface{}, error) {
			// Block until signal received or timeout
			select {
			case <-blockCh:
				return "unblocked", nil
			case <-time.After(2 * time.Second):
				return "timeout", nil
			}
		},
	})

	// 4. Register fast tool
	server.RegisterTool(Tool{
		Name: "fast_tool",
		Handler: func(params map[string]interface{}) (interface{}, error) {
			return "fast", nil
		},
	})

	// 5. Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil {
			// Ignore pipe closed errors
			t.Logf("Server stopped: %v", err)
		}
	}()

	// 6. Send blocking request
	encoder := json.NewEncoder(inWriter)
	req1 := MCPRequest{
		JSONRPC: "2.0",
		ID:      intPtr(1),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "blocking_tool",
		},
	}
	if err := encoder.Encode(req1); err != nil {
		t.Fatalf("Failed to send req1: %v", err)
	}

	// 7. Send fast request IMMEDIATELY
	req2 := MCPRequest{
		JSONRPC: "2.0",
		ID:      intPtr(2),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "fast_tool",
		},
	}
	if err := encoder.Encode(req2); err != nil {
		t.Fatalf("Failed to send req2: %v", err)
	}

	// 8. Read responses
	decoder := json.NewDecoder(outReader)

	// We expect response 2 to arrive quickly, even though request 1 is blocked
	done := make(chan struct{})
	go func() {
		for {
			var resp MCPResponse
			if err := decoder.Decode(&resp); err != nil {
				if err == io.EOF {
					return
				}
				t.Errorf("Failed to decode response: %v", err)
				return
			}

			// Check ID
			idVal := 0
			if resp.ID != nil {
				// Handle varied unmarshal types (float64 for JSON numbers)
				if f, ok := resp.ID.(float64); ok {
					idVal = int(f)
				} else if i, ok := resp.ID.(int); ok {
					idVal = i
				} else if ptr, ok := resp.ID.(*interface{}); ok {
					// Pointer to interface
					if f, ok := (*ptr).(float64); ok {
						idVal = int(f)
					}
				}
			}

			if idVal == 2 {
				// Verify success
				t.Log("Received response for fast_tool")
				close(done)
				return
			} else if idVal == 1 {
				t.Error("Received response for blocking_tool BEFORE fast_tool (sequential processing?)")
			}
		}
	}()

	// 9. Wait for fast response
	select {
	case <-done:
		// Success! Fast tool finished while blocking tool was blocked
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for fast_tool response (server is likely blocking)")
	}

	// 10. Cleanup
	close(blockCh)   // Unblock tool 1
	inWriter.Close() // Stop server
}
