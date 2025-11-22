package dap

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-dap"
)

// SessionState represents the current state of the DAP session
type SessionState int

const (
	StateDisconnected SessionState = iota
	StateConnected
	StateInitialized
	StateConfigured
	StateLaunched
)

func (s SessionState) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnected:
		return "connected"
	case StateInitialized:
		return "initialized"
	case StateConfigured:
		return "configured"
	case StateLaunched:
		return "launched"
	default:
		return "unknown"
	}
}

// Session manages the lifecycle of a DAP debugging session
type Session struct {
	client *Client
	state  SessionState
}

// NewSession creates a new DAP session
func NewSession(host string, port int) *Session {
	return &Session{
		client: NewClient(host, port),
		state:  StateDisconnected,
	}
}

// GetClient returns the underlying DAP client
func (s *Session) GetClient() *Client {
	return s.client
}

// GetState returns the current session state
func (s *Session) GetState() SessionState {
	return s.state
}

// InitializeSession performs the full initialization sequence:
// Connect → Initialize → ConfigurationDone
func (s *Session) InitializeSession(ctx context.Context) error {
	// Connect to DAP server
	if err := s.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Send initialize request
	if err := s.Initialize(ctx); err != nil {
		s.client.Disconnect() // Clean up on error
		s.state = StateDisconnected
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// Note: ConfigurationDone is NOT sent here.
	// It must be sent after the Launch request.
	// See internal/tools/launch.go or godot_launch_* tools.

	return nil
}

// Connect establishes a connection to the DAP server
func (s *Session) Connect(ctx context.Context) error {
	if s.state != StateDisconnected {
		return fmt.Errorf("cannot connect: session is in state %s", s.state)
	}

	ctx, cancel := WithConnectTimeout(ctx)
	defer cancel()

	if err := s.client.Connect(ctx); err != nil {
		return err
	}

	s.state = StateConnected
	return nil
}

// Initialize sends the initialize request
func (s *Session) Initialize(ctx context.Context) error {
	if s.state != StateConnected {
		return fmt.Errorf("cannot initialize: session is in state %s (must be connected)", s.state)
	}

	ctx, cancel := WithCommandTimeout(ctx)
	defer cancel()

	_, err := s.client.Initialize(ctx)
	if err != nil {
		return err
	}

	s.state = StateInitialized
	return nil
}

// ConfigurationDone sends the configurationDone request
func (s *Session) ConfigurationDone(ctx context.Context) error {
	if s.state != StateInitialized {
		return fmt.Errorf("cannot send configurationDone: session is in state %s (must be initialized)", s.state)
	}

	ctx, cancel := WithCommandTimeout(ctx)
	defer cancel()

	if err := s.client.ConfigurationDone(ctx); err != nil {
		return err
	}

	s.state = StateConfigured
	return nil
}

// Close closes the session and disconnects from the DAP server
func (s *Session) Close() error {
	if s.state == StateDisconnected {
		return nil
	}

	err := s.client.Disconnect()
	s.state = StateDisconnected
	return err
}

// IsReady returns whether the session is ready for debugging operations
// (i.e., in Configured or Launched state)
func (s *Session) IsReady() bool {
	return s.state == StateConfigured || s.state == StateLaunched
}

// RequireReady returns an error if the session is not ready
func (s *Session) RequireReady() error {
	if !s.IsReady() {
		return fmt.Errorf("session not ready: current state is %s", s.state)
	}
	return nil
}

// SetLaunched marks the session as launched (called after successful launch)
func (s *Session) SetLaunched() {
	s.state = StateLaunched
}

// Launch is a convenience method that sends a launch request
// This will be implemented in godot.go with Godot-specific parameters
//
// IMPORTANT: Must be called BEFORE ConfigurationDone!
// The launch request is stored by Godot and executed when ConfigurationDone is sent.
func (s *Session) Launch(ctx context.Context, args map[string]interface{}) (*dap.LaunchResponse, error) {
	// Launch must be called after Initialize but BEFORE ConfigurationDone
	if s.state != StateInitialized {
		return nil, fmt.Errorf("cannot launch: session is in state %s (must be initialized)", s.state)
	}

	ctx, cancel := WithCommandTimeout(ctx)
	defer cancel()

	// Marshal args to JSON as required by LaunchRequest.Arguments
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal launch arguments: %w", err)
	}

	request := &dap.LaunchRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  s.client.nextRequestSeq(),
				Type: "request",
			},
			Command: "launch",
		},
		Arguments: argsJSON,
	}

	if err := s.client.write(request); err != nil {
		return nil, fmt.Errorf("failed to send launch request: %w", err)
	}

	response, err := s.client.waitForResponse(ctx, "launch")
	if err != nil {
		return nil, err
	}

	launchResp, ok := response.(*dap.LaunchResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", response)
	}

	// Note: Don't call SetLaunched() here! The launch request is just stored.
	// The actual game launch happens when ConfigurationDone is sent.
	// State remains StateInitialized until ConfigurationDone.
	return launchResp, nil
}
