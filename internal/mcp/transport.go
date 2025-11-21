package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Transport handles stdin/stdout communication for MCP protocol
type Transport struct {
	stdin   io.Reader
	stdout  io.Writer
	decoder *json.Decoder
}

// NewTransport creates a new transport using os.Stdin and os.Stdout
func NewTransport() *Transport {
	return NewTransportWithStreams(os.Stdin, os.Stdout)
}

// NewTransportWithStreams creates a new transport with custom streams (useful for testing)
func NewTransportWithStreams(stdin io.Reader, stdout io.Writer) *Transport {
	return &Transport{
		stdin:   stdin,
		stdout:  stdout,
		decoder: json.NewDecoder(stdin),
	}
}

// ReadRequest reads and parses a single MCP request from stdin
// Returns the parsed request or an error if reading/parsing fails
func (t *Transport) ReadRequest() (*MCPRequest, error) {
	var req MCPRequest
	if err := t.decoder.Decode(&req); err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("failed to parse JSON request: %w", err)
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		return nil, fmt.Errorf("invalid JSON-RPC version: %s (expected 2.0)", req.JSONRPC)
	}

	return &req, nil
}

// WriteResponse writes an MCP response to stdout
func (t *Transport) WriteResponse(resp MCPResponse) error {
	// Ensure JSON-RPC version is set
	resp.JSONRPC = "2.0"

	// Serialize to JSON
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Write to stdout with newline
	if _, err := t.stdout.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}

	return nil
}

// WriteError is a convenience method to write an error response
func (t *Transport) WriteError(requestID interface{}, code int, message string) error {
	return t.WriteResponse(MCPResponse{
		ID: requestID,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	})
}

// WriteSuccess is a convenience method to write a success response
func (t *Transport) WriteSuccess(requestID interface{}, result interface{}) error {
	return t.WriteResponse(MCPResponse{
		ID:     requestID,
		Result: result,
	})
}
