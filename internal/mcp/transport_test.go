package mcp

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// TestReadRequest_Valid verifies that valid JSON-RPC 2.0 requests are parsed correctly
func TestReadRequest_Valid(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"test","params":{"foo":"bar"}}`
	stdin := strings.NewReader(input + "\n")
	stdout := &bytes.Buffer{}

	transport := NewTransportWithStreams(stdin, stdout)
	req, err := transport.ReadRequest()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.JSONRPC != "2.0" {
		t.Errorf("Expected jsonrpc '2.0', got '%s'", req.JSONRPC)
	}
	if req.Method != "test" {
		t.Errorf("Expected method 'test', got '%s'", req.Method)
	}
	if req.Params["foo"] != "bar" {
		t.Errorf("Expected params['foo'] = 'bar', got '%v'", req.Params["foo"])
	}
}

// TestReadRequest_EOF verifies that EOF is handled correctly
func TestReadRequest_EOF(t *testing.T) {
	stdin := strings.NewReader("") // Empty input = immediate EOF
	stdout := &bytes.Buffer{}

	transport := NewTransportWithStreams(stdin, stdout)
	req, err := transport.ReadRequest()

	if err != io.EOF {
		t.Fatalf("Expected io.EOF, got %v", err)
	}
	if req != nil {
		t.Errorf("Expected nil request on EOF, got %+v", req)
	}
}

// TestReadRequest_InvalidJSON verifies that malformed JSON is rejected
func TestReadRequest_InvalidJSON(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,INVALID}` // Missing quotes around INVALID
	stdin := strings.NewReader(input + "\n")
	stdout := &bytes.Buffer{}

	transport := NewTransportWithStreams(stdin, stdout)
	req, err := transport.ReadRequest()

	if err == nil {
		t.Fatal("Expected error for invalid JSON, got none")
	}
	if req != nil {
		t.Errorf("Expected nil request on error, got %+v", req)
	}
}

// TestReadRequest_WrongVersion verifies that non-2.0 JSON-RPC versions are rejected
func TestReadRequest_WrongVersion(t *testing.T) {
	input := `{"jsonrpc":"1.0","id":1,"method":"test","params":{}}`
	stdin := strings.NewReader(input + "\n")
	stdout := &bytes.Buffer{}

	transport := NewTransportWithStreams(stdin, stdout)
	_, err := transport.ReadRequest()

	if err == nil {
		t.Fatal("Expected error for wrong JSON-RPC version, got none")
	}
	if !strings.Contains(err.Error(), "invalid JSON-RPC version") {
		t.Errorf("Expected version error, got: %v", err)
	}
}

// TestWriteResponse_Success verifies that success responses are formatted correctly
func TestWriteResponse_Success(t *testing.T) {
	stdin := strings.NewReader("")
	stdout := &bytes.Buffer{}

	transport := NewTransportWithStreams(stdin, stdout)
	err := transport.WriteSuccess(123, map[string]string{"status": "ok"})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, `"jsonrpc":"2.0"`) {
		t.Error("Response missing jsonrpc field")
	}
	if !strings.Contains(output, `"id":123`) {
		t.Error("Response missing or incorrect id")
	}
	if !strings.Contains(output, `"result"`) {
		t.Error("Response missing result field")
	}
}

// TestWriteResponse_Error verifies that error responses are formatted correctly
func TestWriteResponse_Error(t *testing.T) {
	stdin := strings.NewReader("")
	stdout := &bytes.Buffer{}

	transport := NewTransportWithStreams(stdin, stdout)
	err := transport.WriteError(456, -32600, "Invalid request")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, `"jsonrpc":"2.0"`) {
		t.Error("Response missing jsonrpc field")
	}
	if !strings.Contains(output, `"id":456`) {
		t.Error("Response missing or incorrect id")
	}
	if !strings.Contains(output, `"error"`) {
		t.Error("Response missing error field")
	}
	if !strings.Contains(output, `"code":-32600`) {
		t.Error("Response missing or incorrect error code")
	}
	if !strings.Contains(output, "Invalid request") {
		t.Error("Response missing error message")
	}
}
