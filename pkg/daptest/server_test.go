package daptest

import (
	"bufio"
	"encoding/json"
	"net"
	"strconv"
	"testing"

	"github.com/google/go-dap"
)

func TestMockServer(t *testing.T) {
	// Start server
	server := NewServer(t)
	defer server.Close()

	// Client connects
	conn, err := net.Dial("tcp", server.Address())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Client sends request
	req := &dap.InitializeRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  1,
				Type: "request",
			},
			Command: "initialize",
		},
		Arguments: dap.InitializeRequestArguments{
			AdapterID: "test-client",
		},
	}

	body, _ := json.Marshal(req)
	// header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))

	conn.Write([]byte("Content-Length: " + strconv.Itoa(len(body)) + "\r\n\r\n"))
	conn.Write(body)
}

func TestMockServer_RoundTrip(t *testing.T) {
	server := NewServer(t)
	defer server.Close()

	// Use real DAP client to talk to it?
	// Or just raw net.Conn for simplicity in this unit test.

	conn, err := net.Dial("tcp", server.Address())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Send Request
	req := dap.InitializeRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  1,
				Type: "request",
			},
			Command: "initialize",
		},
		Arguments: dap.InitializeRequestArguments{
			AdapterID: "test",
		},
	}

	reqBytes, _ := json.Marshal(req)
	conn.Write([]byte("Content-Length: " + strconv.Itoa(len(reqBytes)) + "\r\n\r\n"))
	conn.Write(reqBytes)

	// Server expects request
	msg, err := server.ExpectRequest("initialize")
	if err != nil {
		t.Fatalf("ExpectRequest failed: %v", err)
	}

	if reqMsg, ok := msg.(*dap.InitializeRequest); !ok {
		t.Errorf("Expected InitializeRequest, got %T", msg)
	} else {
		if reqMsg.Arguments.AdapterID != "test" {
			t.Errorf("Wrong AdapterID: %s", reqMsg.Arguments.AdapterID)
		}
	}

	// Server sends response
	resp := &dap.InitializeResponse{
		Response: dap.Response{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  2,
				Type: "response",
			},
			RequestSeq: 1,
			Success:    true,
			Command:    "initialize",
		},
	}

	if err := server.Send(resp); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	// Client reads response
	reader := bufio.NewReader(conn)
	respMsg, err := dap.ReadProtocolMessage(reader)
	if err != nil {
		t.Fatalf("ReadProtocolMessage failed: %v", err)
	}

	if _, ok := respMsg.(*dap.InitializeResponse); !ok {
		t.Errorf("Expected InitializeResponse, got %T", respMsg)
	}
}
