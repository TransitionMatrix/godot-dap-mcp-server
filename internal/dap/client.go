package dap

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	// DEBUG: Log the message being sent
	if jsonBytes, err := json.MarshalIndent(msg, "", "  "); err == nil {
		log.Printf("=== SENDING DAP MESSAGE ===\n%s\n=========================", string(jsonBytes))
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

// SetBreakpoints sets breakpoints for a specific file
// Returns the verified breakpoint information from the server
func (c *Client) SetBreakpoints(ctx context.Context, file string, lines []int) (*dap.SetBreakpointsResponse, error) {
	// Convert line numbers to breakpoints
	breakpoints := make([]dap.SourceBreakpoint, len(lines))
	for i, line := range lines {
		breakpoints[i] = dap.SourceBreakpoint{
			Line: line,
		}
	}

	request := &dap.SetBreakpointsRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "setBreakpoints",
		},
		Arguments: dap.SetBreakpointsArguments{
			Source: dap.Source{
				Path: file,
			},
			Breakpoints: breakpoints,
		},
	}

	if err := c.write(request); err != nil {
		return nil, fmt.Errorf("failed to send setBreakpoints request: %w", err)
	}

	// Wait for setBreakpoints response
	response, err := c.waitForResponse(ctx, "setBreakpoints")
	if err != nil {
		return nil, err
	}

	bpResp, ok := response.(*dap.SetBreakpointsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}

	return bpResp, nil
}

// Continue resumes execution of the specified thread
// Use threadId 0 to continue all threads (Godot typically uses single thread)
func (c *Client) Continue(ctx context.Context, threadId int) (*dap.ContinueResponse, error) {
	request := &dap.ContinueRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "continue",
		},
		Arguments: dap.ContinueArguments{
			ThreadId: threadId,
		},
	}

	if err := c.write(request); err != nil {
		return nil, fmt.Errorf("failed to send continue request: %w", err)
	}

	// Wait for continue response
	response, err := c.waitForResponse(ctx, "continue")
	if err != nil {
		return nil, err
	}

	contResp, ok := response.(*dap.ContinueResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}

	return contResp, nil
}

// Next steps over the current line (step over)
// Use threadId from the stopped event
func (c *Client) Next(ctx context.Context, threadId int) (*dap.NextResponse, error) {
	request := &dap.NextRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "next",
		},
		Arguments: dap.NextArguments{
			ThreadId: threadId,
		},
	}

	if err := c.write(request); err != nil {
		return nil, fmt.Errorf("failed to send next request: %w", err)
	}

	// Wait for next response
	response, err := c.waitForResponse(ctx, "next")
	if err != nil {
		return nil, err
	}

	nextResp, ok := response.(*dap.NextResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}

	return nextResp, nil
}

// StepIn steps into the function at the current line
// Use threadId from the stopped event
func (c *Client) StepIn(ctx context.Context, threadId int) (*dap.StepInResponse, error) {
	request := &dap.StepInRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "stepIn",
		},
		Arguments: dap.StepInArguments{
			ThreadId: threadId,
		},
	}

	if err := c.write(request); err != nil {
		return nil, fmt.Errorf("failed to send stepIn request: %w", err)
	}

	// Wait for stepIn response
	response, err := c.waitForResponse(ctx, "stepIn")
	if err != nil {
		return nil, err
	}

	stepInResp, ok := response.(*dap.StepInResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}

	return stepInResp, nil
}

// Pause pauses execution of the specified thread
// Use threadId 1 for Godot (single thread)
// This will trigger a 'stopped' event with reason='pause'
func (c *Client) Pause(ctx context.Context, threadId int) (*dap.PauseResponse, error) {
	request := &dap.PauseRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "pause",
		},
		Arguments: dap.PauseArguments{
			ThreadId: threadId,
		},
	}

	if err := c.write(request); err != nil {
		return nil, fmt.Errorf("failed to send pause request: %w", err)
	}

	// Wait for pause response
	response, err := c.waitForResponse(ctx, "pause")
	if err != nil {
		return nil, err
	}

	pauseResp, ok := response.(*dap.PauseResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}

	return pauseResp, nil
}

// Threads requests the list of active threads.
// Godot always returns a single thread with ID 1 named "Main".
func (c *Client) Threads(ctx context.Context) (*dap.ThreadsResponse, error) {
	request := &dap.ThreadsRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "threads",
		},
	}

	if err := c.write(request); err != nil {
		return nil, fmt.Errorf("failed to send threads request: %w", err)
	}

	// Wait for threads response
	response, err := c.waitForResponse(ctx, "threads")
	if err != nil {
		return nil, err
	}

	threadsResp, ok := response.(*dap.ThreadsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}

	return threadsResp, nil
}

// StackTrace requests the call stack for the specified thread.
// Returns stack frames with source file paths, line numbers, and frame IDs.
func (c *Client) StackTrace(ctx context.Context, threadId int, startFrame int, levels int) (*dap.StackTraceResponse, error) {
	request := &dap.StackTraceRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "stackTrace",
		},
		Arguments: dap.StackTraceArguments{
			ThreadId:   threadId,
			StartFrame: startFrame,
			Levels:     levels,
		},
	}

	if err := c.write(request); err != nil {
		return nil, fmt.Errorf("failed to send stackTrace request: %w", err)
	}

	// Wait for stackTrace response
	response, err := c.waitForResponse(ctx, "stackTrace")
	if err != nil {
		return nil, err
	}

	stackResp, ok := response.(*dap.StackTraceResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}

	return stackResp, nil
}

// Scopes requests the variable scopes for the specified stack frame.
// Returns Locals, Members, and Globals scopes with variablesReference IDs.
func (c *Client) Scopes(ctx context.Context, frameId int) (*dap.ScopesResponse, error) {
	request := &dap.ScopesRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "scopes",
		},
		Arguments: dap.ScopesArguments{
			FrameId: frameId,
		},
	}

	if err := c.write(request); err != nil {
		return nil, fmt.Errorf("failed to send scopes request: %w", err)
	}

	// Wait for scopes response
	response, err := c.waitForResponse(ctx, "scopes")
	if err != nil {
		return nil, err
	}

	scopesResp, ok := response.(*dap.ScopesResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}

	return scopesResp, nil
}

// Variables requests the variables in the specified scope or expands a complex variable.
// Use variablesReference from scopes response or from a variable with variablesReference > 0.
func (c *Client) Variables(ctx context.Context, variablesReference int) (*dap.VariablesResponse, error) {
	request := &dap.VariablesRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "variables",
		},
		Arguments: dap.VariablesArguments{
			VariablesReference: variablesReference,
		},
	}

	if err := c.write(request); err != nil {
		return nil, fmt.Errorf("failed to send variables request: %w", err)
	}

	// Wait for variables response
	response, err := c.waitForResponse(ctx, "variables")
	if err != nil {
		return nil, err
	}

	varsResp, ok := response.(*dap.VariablesResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}

	return varsResp, nil
}

// Evaluate evaluates the specified expression in the context of the specified stack frame.
// Returns the result value, type, and variablesReference (if result is complex).
// Context can be "watch", "repl", or "hover" to indicate the evaluation context.
func (c *Client) Evaluate(ctx context.Context, expression string, frameId int, context string) (*dap.EvaluateResponse, error) {
	request := &dap.EvaluateRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "evaluate",
		},
		Arguments: dap.EvaluateArguments{
			Expression: expression,
			FrameId:    frameId,
			Context:    context,
		},
	}

	if err := c.write(request); err != nil {
		return nil, fmt.Errorf("failed to send evaluate request: %w", err)
	}

	// Wait for evaluate response
	response, err := c.waitForResponse(ctx, "evaluate")
	if err != nil {
		return nil, err
	}

	evalResp, ok := response.(*dap.EvaluateResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}

	return evalResp, nil
}
