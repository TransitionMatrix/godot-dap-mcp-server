package dap

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/google/go-dap"
)

// Client manages a connection to a Godot DAP server
type Client struct {
	host   string
	port   int
	conn   net.Conn
	reader *bufio.Reader
	codec  *dap.Codec

	// Request ID management
	mu      sync.Mutex
	nextSeq int

	// Connection state
	connected bool
}

// NewClient creates a new DAP client for connecting to Godot
func NewClient(host string, port int) *Client {
	return &Client{
		host:    host,
		port:    port,
		nextSeq: 1,
		codec:   dap.NewCodec(),
	}
}

// Connect establishes a TCP connection to the Godot DAP server
func (c *Client) Connect(ctx context.Context) error {
	if c.connected {
		return fmt.Errorf("already connected")
	}

	address := fmt.Sprintf("%s:%d", c.host, c.port)

	// Establish TCP connection
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.connected = true

	return nil
}

// Disconnect closes the connection to the DAP server
func (c *Client) Disconnect() error {
	if !c.connected {
		return nil
	}

	c.connected = false
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected returns whether the client is currently connected
func (c *Client) IsConnected() bool {
	return c.connected
}

// nextRequestSeq returns the next sequence number for a request
func (c *Client) nextRequestSeq() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	seq := c.nextSeq
	c.nextSeq++
	return seq
}

// read reads a single DAP message from the connection
// This is a low-level method - use ReadWithTimeout for actual operations
func (c *Client) read() (dap.Message, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected")
	}

	// Read the base message (raw bytes)
	data, err := dap.ReadBaseMessage(c.reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read base message: %w", err)
	}

	// Decode the message using the codec
	msg, err := c.codec.DecodeMessage(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode message: %w", err)
	}

	return msg, nil
}

// write sends a DAP message to the connection
func (c *Client) write(msg dap.Message) error {
	if !c.connected {
		return fmt.Errorf("not connected")
	}
	return dap.WriteProtocolMessage(c.conn, msg)
}

// Initialize sends the initialize request to the DAP server
// This must be the first request sent after connecting
func (c *Client) Initialize(ctx context.Context) (*dap.InitializeResponse, error) {
	request := &dap.InitializeRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "initialize",
		},
		Arguments: dap.InitializeRequestArguments{
			ClientID:                     "godot-dap-mcp-server",
			ClientName:                   "Godot DAP MCP Server",
			AdapterID:                    "godot",
			Locale:                       "en-US",
			LinesStartAt1:                true,
			ColumnsStartAt1:              true,
			PathFormat:                   "path",
			SupportsVariableType:         true,
			SupportsVariablePaging:       false,
			SupportsRunInTerminalRequest: false,
			SupportsMemoryReferences:     false,
			SupportsProgressReporting:    false,
			SupportsInvalidatedEvent:     false,
		},
	}

	if err := c.write(request); err != nil {
		return nil, fmt.Errorf("failed to send initialize request: %w", err)
	}

	// Wait for initialize response using timeout wrapper
	response, err := c.waitForResponse(ctx, "initialize")
	if err != nil {
		return nil, err
	}

	initResp, ok := response.(*dap.InitializeResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}

	return initResp, nil
}

// ConfigurationDone tells the DAP server that configuration is complete
// This must be sent after Initialize and before launching/attaching
func (c *Client) ConfigurationDone(ctx context.Context) error {
	request := &dap.ConfigurationDoneRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "configurationDone",
		},
	}

	if err := c.write(request); err != nil {
		return fmt.Errorf("failed to send configurationDone request: %w", err)
	}

	// Wait for configurationDone response
	_, err := c.waitForResponse(ctx, "configurationDone")
	return err
}
