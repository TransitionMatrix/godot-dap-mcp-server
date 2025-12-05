package dap

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

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

	// Pending requests (seq -> channel)
	pendingReqs map[int]chan dap.Message
	reqMu       sync.Mutex

	// Event listeners
	eventListeners []chan dap.Message
	eventMu        sync.Mutex

	// Connection state
	connected bool
}

// NewClient creates a new DAP client for connecting to Godot
func NewClient(host string, port int) *Client {
	return &Client{
		host:           host,
		port:           port,
		nextSeq:        1,
		codec:          dap.NewCodec(),
		pendingReqs:    make(map[int]chan dap.Message),
		eventListeners: make([]chan dap.Message, 0),
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

	// Start background read loop
	go c.readLoop()

	return nil
}

// readLoop continuously reads messages from the connection
func (c *Client) readLoop() {
	for {
		msg, err := c.read()
		if err != nil {
			if c.connected {
				log.Printf("Connection error: %v", err)
				c.connected = false
			}
			return
		}
		c.handleMessage(msg)
	}
}

// handleMessage dispatches incoming messages
func (c *Client) handleMessage(msg dap.Message) {
	switch m := msg.(type) {
	case *dap.Response:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.ErrorResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.InitializeResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.ConfigurationDoneResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.LaunchResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.AttachResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.SetBreakpointsResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.ContinueResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.NextResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.StepInResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.StepOutResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.PauseResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.ThreadsResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.StackTraceResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.ScopesResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.VariablesResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.EvaluateResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.SetVariableResponse:
		c.dispatchResponse(m.RequestSeq, m)
	case *dap.DisconnectResponse:
		c.dispatchResponse(m.RequestSeq, m)
	default:
		// It's an event or unknown message
		// For now, just log it or handle via event listeners
		if _, ok := msg.(dap.EventMessage); ok {
			c.logEvent(msg)
			c.broadcastEvent(msg)
		} else {
			log.Printf("Received unknown message type: %T", msg)
		}
	}
}

// SubscribeToEvents subscribes to all DAP events.
// Returns a channel to receive events and a cleanup function.
func (c *Client) SubscribeToEvents() (<-chan dap.Message, func()) {
	ch := make(chan dap.Message, 100) // Buffer to prevent blocking
	c.eventMu.Lock()
	c.eventListeners = append(c.eventListeners, ch)
	c.eventMu.Unlock()

	cleanup := func() {
		c.eventMu.Lock()
		defer c.eventMu.Unlock()
		for i, listener := range c.eventListeners {
			if listener == ch {
				// Remove (swap with last and shrink)
				c.eventListeners[i] = c.eventListeners[len(c.eventListeners)-1]
				c.eventListeners = c.eventListeners[:len(c.eventListeners)-1]
				// close(ch) // Don't close, just stop sending. Caller might still be reading.
				break
			}
		}
	}
	return ch, cleanup
}

// broadcastEvent sends an event to all listeners
func (c *Client) broadcastEvent(event dap.Message) {
	c.eventMu.Lock()
	defer c.eventMu.Unlock()
	for _, ch := range c.eventListeners {
		select {
		case ch <- event:
		default:
			log.Printf("Warning: Event listener buffer full, dropping event")
		}
	}
}

// dispatchResponse sends a response to the waiting request
func (c *Client) dispatchResponse(seq int, msg dap.Message) {
	c.reqMu.Lock()
	ch, ok := c.pendingReqs[seq]
	if ok {
		delete(c.pendingReqs, seq)
	}
	c.reqMu.Unlock()

	if ok {
		ch <- msg
	} else {
		log.Printf("Received response for unknown/timed-out request seq %d: %T", seq, msg)
	}
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

// read reads a message from the connection
func (c *Client) read() (dap.Message, error) {
	// Use the bufio reader that was initialized in Connect
	reader := c.reader

	// Read Content-Length header
	// DAP headers are HTTP-like: "Content-Length: 123\r\n\r\n"
	contentLength := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read header: %w", err)
		}

		// Remove trailing whitespace
		line = line[:len(line)-2] // remove \r\n

		if line == "" {
			// Empty line marks end of headers
			break
		}

		// Parse Content-Length
		var length int
		if n, _ := fmt.Sscanf(line, "Content-Length: %d", &length); n == 1 {
			contentLength = length
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("missing or invalid Content-Length header")
	}

	// Read body
	body := make([]byte, contentLength)
	_, err := io.ReadFull(reader, body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	// Log incoming message (pretty printed)
	var rawMsg map[string]interface{}
	if err := json.Unmarshal(body, &rawMsg); err == nil {
		if prettyBytes, err := json.MarshalIndent(rawMsg, "", "  "); err == nil {
			log.Printf("[DAP RCVD] %s", string(prettyBytes))
		} else {
			log.Printf("[DAP RCVD] %s", string(body))
		}
	} else {
		log.Printf("[DAP RCVD] %s", string(body))
	}

	// Decode into specific type based on Type and Command/Event
	return dap.DecodeProtocolMessage(body)
}

// write sends a message to the connection
func (c *Client) write(msg dap.Message) error {
	if !c.connected {
		return fmt.Errorf("not connected")
	}

	if jsonBytes, err := json.MarshalIndent(msg, "", "  "); err == nil {
		log.Printf("[DAP SENT] %s", string(jsonBytes))
	} else {
		log.Printf("[DAP SENT] (failed to marshal for logging): %v", msg)
	}

	return dap.WriteProtocolMessage(c.conn, msg)
}

// sendRequestAndWait sends a request and waits for the response
func (c *Client) sendRequestAndWait(ctx context.Context, req dap.Message) (dap.Message, error) {
	seq := req.GetSeq()
	ch := make(chan dap.Message, 1)

	c.reqMu.Lock()
	c.pendingReqs[seq] = ch
	c.reqMu.Unlock()

	// Ensure cleanup
	defer func() {
		c.reqMu.Lock()
		delete(c.pendingReqs, seq)
		c.reqMu.Unlock()
	}()

	if err := c.write(req); err != nil {
		return nil, err
	}

	select {
	case resp := <-ch:
		// Check for ErrorResponse
		if errResp, ok := resp.(*dap.ErrorResponse); ok {
			return nil, fmt.Errorf("DAP error: %s", errResp.Message)
		}
		return resp, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("request timed out: %w", ctx.Err())
	}
}

// Initialize sends the initialize request to the DAP server
// This must be the first request sent after connecting
func (c *Client) Initialize(ctx context.Context) (*dap.InitializeResponse, error) {
	// Subscribe to events BEFORE sending request to avoid missing "initialized" event
	events, cleanup := c.SubscribeToEvents()
	defer cleanup()

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

	resp, err := c.sendRequestAndWait(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to send initialize request: %w", err)
	}

	initResp, ok := resp.(*dap.InitializeResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}

	// Wait for initialized event
	log.Println("Waiting for initialized event...")
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for initialized event: %w", ctx.Err())
		case msg := <-events:
			if _, ok := msg.(*dap.InitializedEvent); ok {
				log.Println("Received initialized event")
				return initResp, nil
			}
		}
	}
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

	_, err := c.sendRequestAndWait(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to send configurationDone request: %w", err)
	}

	return nil
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

	resp, err := c.sendRequestAndWait(ctx, request)
	if err != nil {
		return nil, err
	}

	bpResp, ok := resp.(*dap.SetBreakpointsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
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

	resp, err := c.sendRequestAndWait(ctx, request)
	if err != nil {
		return nil, err
	}

	contResp, ok := resp.(*dap.ContinueResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
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

	resp, err := c.sendRequestAndWait(ctx, request)
	if err != nil {
		return nil, err
	}

	nextResp, ok := resp.(*dap.NextResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
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

	resp, err := c.sendRequestAndWait(ctx, request)
	if err != nil {
		return nil, err
	}

	stepInResp, ok := resp.(*dap.StepInResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
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

	resp, err := c.sendRequestAndWait(ctx, request)
	if err != nil {
		return nil, err
	}

	pauseResp, ok := resp.(*dap.PauseResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
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

	resp, err := c.sendRequestAndWait(ctx, request)
	if err != nil {
		return nil, err
	}

	threadsResp, ok := resp.(*dap.ThreadsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
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

	resp, err := c.sendRequestAndWait(ctx, request)
	if err != nil {
		return nil, err
	}

	stackResp, ok := resp.(*dap.StackTraceResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
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

	resp, err := c.sendRequestAndWait(ctx, request)
	if err != nil {
		return nil, err
	}

	scopesResp, ok := resp.(*dap.ScopesResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
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

	resp, err := c.sendRequestAndWait(ctx, request)
	if err != nil {
		return nil, err
	}

	varsResp, ok := resp.(*dap.VariablesResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
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

	resp, err := c.sendRequestAndWait(ctx, request)
	if err != nil {
		return nil, err
	}

	evalResp, ok := resp.(*dap.EvaluateResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}

	return evalResp, nil
}

// Launch sends a launch request to start the Godot game with specified parameters.
// Note: The launch request only stores parameters. The game won't actually launch
// until configurationDone() is called after this.
//
// Common arguments:
//   - project: Absolute path to Godot project directory (must contain project.godot)
//   - scene: "main" (project's main scene), "current" (editor's active scene), or "res://path/to/scene.tscn"
//   - platform: "host" (default), "android", or "web"
//   - noDebug: If true, breakpoints will be ignored
//   - additional launch arguments can be passed in args map
func (c *Client) Launch(ctx context.Context, args map[string]interface{}) (*dap.LaunchResponse, error) {
	// Marshal arguments to JSON
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal launch arguments: %w", err)
	}

	request := &dap.LaunchRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "launch",
		},
		Arguments: argsJSON,
	}

	resp, err := c.sendRequestAndWait(ctx, request)
	if err != nil {
		return nil, err
	}

	launchResp, ok := resp.(*dap.LaunchResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}

	return launchResp, nil
}

// LaunchWithConfigurationDone sends a launch request followed immediately by configurationDone.
// This is required for Godot, which only sends the launch response AFTER receiving configurationDone.
func (c *Client) LaunchWithConfigurationDone(ctx context.Context, args map[string]interface{}) (*dap.LaunchResponse, error) {
	// 1. Send Launch Request and Wait
	// Godot 4.x sends LaunchResponse immediately before/during ConfigurationDone?
	// Actually, standard DAP says LaunchResponse comes first.
	// Godot might be interleaving.
	// If Godot blocks sending LaunchResponse until it gets ConfigurationDone, we have a deadlock if we wait synchronously.
	// But my simulation showed LaunchResponse came back.
	// However, the comment "This is required for Godot, which only sends the launch response AFTER receiving configurationDone"
	// contradicts my finding or implies a deadlock risk.

	// Let's assume Godot MIGHT block launch response.
	// If so, we should send both requests asynchronously?
	// But sendRequestAndWait blocks.

	// If Godot waits for configDone to send launch response, we CANNOT wait for launch response before sending configDone.
	// We must send Launch, then Send ConfigDone, then wait for both.

	// To do this with sendRequestAndWait, we'd need goroutines.

	// Marshal arguments to JSON
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal launch arguments: %w", err)
	}

	launchRequest := &dap.LaunchRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "launch",
		},
		Arguments: argsJSON,
	}

	configDoneRequest := &dap.ConfigurationDoneRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "configurationDone",
		},
	}

	// Use channels to wait
	launchSeq := launchRequest.Seq
	configSeq := configDoneRequest.Seq

	launchCh := make(chan dap.Message, 1)
	configCh := make(chan dap.Message, 1)

	c.reqMu.Lock()
	c.pendingReqs[launchSeq] = launchCh
	c.pendingReqs[configSeq] = configCh
	c.reqMu.Unlock()

	defer func() {
		c.reqMu.Lock()
		delete(c.pendingReqs, launchSeq)
		delete(c.pendingReqs, configSeq)
		c.reqMu.Unlock()
	}()

	log.Println("DEBUG: Sending Launch Request...")
	if err := c.write(launchRequest); err != nil {
		return nil, fmt.Errorf("failed to send launch request: %w", err)
	}

	log.Println("DEBUG: Sending ConfigurationDone Request...")
	if err := c.write(configDoneRequest); err != nil {
		return nil, fmt.Errorf("failed to send configurationDone request: %w", err)
	}

	// Wait for both
	var launchResp *dap.LaunchResponse

	// We need to collect both. Order doesn't matter.
	timeout := time.After(30 * time.Second) // Default timeout
	if d, ok := ctx.Deadline(); ok {
		timeout = time.After(time.Until(d))
	}

	gotLaunch := false
	gotConfig := false

	for !gotLaunch || !gotConfig {
		select {
		case msg := <-launchCh:
			gotLaunch = true
			if m, ok := msg.(*dap.LaunchResponse); ok {
				if !m.Success {
					return nil, fmt.Errorf("launch failed: %s", m.Message)
				}
				launchResp = m
			} else if m, ok := msg.(*dap.ErrorResponse); ok {
				return nil, fmt.Errorf("launch error: %s", m.Message)
			}
		case msg := <-configCh:
			gotConfig = true
			if m, ok := msg.(*dap.ConfigurationDoneResponse); ok {
				if !m.Success {
					return nil, fmt.Errorf("configurationDone failed: %s", m.Message)
				}
			} else if m, ok := msg.(*dap.ErrorResponse); ok {
				return nil, fmt.Errorf("configurationDone error: %s", m.Message)
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for launch/configDone responses")
		}
	}

	return launchResp, nil
}

// Attach sends an attach request to connect to an already running Godot game.
// Note: The attach request only stores parameters. The connection won't actually happen
// until configurationDone() is called after this.
func (c *Client) Attach(ctx context.Context, args map[string]interface{}) (*dap.AttachResponse, error) {
	// Marshal arguments to JSON
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal attach arguments: %w", err)
	}

	request := &dap.AttachRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "attach",
		},
		Arguments: argsJSON,
	}

	resp, err := c.sendRequestAndWait(ctx, request)
	if err != nil {
		return nil, err
	}

	attachResp, ok := resp.(*dap.AttachResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}

	return attachResp, nil
}

// AttachWithConfigurationDone sends an attach request followed immediately by configurationDone.
// This is required for Godot, which only sends the attach response AFTER receiving configurationDone.
func (c *Client) AttachWithConfigurationDone(ctx context.Context, args map[string]interface{}) (*dap.AttachResponse, error) {
	// Marshal arguments to JSON
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal attach arguments: %w", err)
	}

	attachRequest := &dap.AttachRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "attach",
		},
		Arguments: argsJSON,
	}

	configDoneRequest := &dap.ConfigurationDoneRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  c.nextRequestSeq(),
				Type: "request",
			},
			Command: "configurationDone",
		},
	}

	// Use channels to wait
	attachSeq := attachRequest.Seq
	configSeq := configDoneRequest.Seq

	attachCh := make(chan dap.Message, 1)
	configCh := make(chan dap.Message, 1)

	c.reqMu.Lock()
	c.pendingReqs[attachSeq] = attachCh
	c.pendingReqs[configSeq] = configCh
	c.reqMu.Unlock()

	defer func() {
		c.reqMu.Lock()
		delete(c.pendingReqs, attachSeq)
		delete(c.pendingReqs, configSeq)
		c.reqMu.Unlock()
	}()

	log.Println("DEBUG: Sending Attach Request...")
	if err := c.write(attachRequest); err != nil {
		return nil, fmt.Errorf("failed to send attach request: %w", err)
	}

	log.Println("DEBUG: Sending ConfigurationDone Request...")
	if err := c.write(configDoneRequest); err != nil {
		return nil, fmt.Errorf("failed to send configurationDone request: %w", err)
	}

	// Wait for both
	var attachResp *dap.AttachResponse

	// We need to collect both. Order doesn't matter.
	timeout := time.After(30 * time.Second) // Default timeout
	if d, ok := ctx.Deadline(); ok {
		timeout = time.After(time.Until(d))
	}

	gotAttach := false
	gotConfig := false

	for !gotAttach || !gotConfig {
		select {
		case msg := <-attachCh:
			gotAttach = true
			if m, ok := msg.(*dap.AttachResponse); ok {
				if !m.Success {
					return nil, fmt.Errorf("attach failed: %s", m.Message)
				}
				attachResp = m
			} else if m, ok := msg.(*dap.ErrorResponse); ok {
				return nil, fmt.Errorf("attach error: %s", m.Message)
			}
		case msg := <-configCh:
			gotConfig = true
			if m, ok := msg.(*dap.ConfigurationDoneResponse); ok {
				if !m.Success {
					return nil, fmt.Errorf("configurationDone failed: %s", m.Message)
				}
			} else if m, ok := msg.(*dap.ErrorResponse); ok {
				return nil, fmt.Errorf("configurationDone error: %s", m.Message)
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for attach/configDone responses")
		}
	}

	return attachResp, nil
}
