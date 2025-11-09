package dap

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("localhost", 6006)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.host != "localhost" {
		t.Errorf("expected host localhost, got %s", client.host)
	}

	if client.port != 6006 {
		t.Errorf("expected port 6006, got %d", client.port)
	}

	if client.connected {
		t.Error("client should not be connected initially")
	}

	if client.nextSeq != 1 {
		t.Errorf("expected initial sequence to be 1, got %d", client.nextSeq)
	}
}

func TestRequestSeqGeneration(t *testing.T) {
	client := NewClient("localhost", 6006)

	// Test sequential sequence generation
	seq1 := client.nextRequestSeq()
	seq2 := client.nextRequestSeq()
	seq3 := client.nextRequestSeq()

	if seq1 != 1 {
		t.Errorf("expected first sequence to be 1, got %d", seq1)
	}

	if seq2 != 2 {
		t.Errorf("expected second sequence to be 2, got %d", seq2)
	}

	if seq3 != 3 {
		t.Errorf("expected third sequence to be 3, got %d", seq3)
	}
}

func TestClientInitialState(t *testing.T) {
	client := NewClient("localhost", 6006)

	if client.IsConnected() {
		t.Error("client should not be connected initially")
	}
}

func TestDisconnectWhenNotConnected(t *testing.T) {
	client := NewClient("localhost", 6006)

	// Should not error when disconnecting a non-connected client
	err := client.Disconnect()
	if err != nil {
		t.Errorf("Disconnect should not error on non-connected client: %v", err)
	}
}

func TestNewSession(t *testing.T) {
	session := NewSession("localhost", 6006)

	if session == nil {
		t.Fatal("NewSession returned nil")
	}

	if session.client == nil {
		t.Fatal("session.client is nil")
	}

	if session.state != StateDisconnected {
		t.Errorf("expected initial state to be StateDisconnected, got %s", session.state)
	}
}

func TestSessionStateTransitions(t *testing.T) {
	session := NewSession("localhost", 6006)

	// Test state string representation
	states := []struct {
		state    SessionState
		expected string
	}{
		{StateDisconnected, "disconnected"},
		{StateConnected, "connected"},
		{StateInitialized, "initialized"},
		{StateConfigured, "configured"},
		{StateLaunched, "launched"},
	}

	for _, tc := range states {
		if tc.state.String() != tc.expected {
			t.Errorf("expected state %d to be %s, got %s", tc.state, tc.expected, tc.state.String())
		}
	}

	// Test IsReady
	session.state = StateDisconnected
	if session.IsReady() {
		t.Error("session should not be ready when disconnected")
	}

	session.state = StateConnected
	if session.IsReady() {
		t.Error("session should not be ready when connected")
	}

	session.state = StateInitialized
	if session.IsReady() {
		t.Error("session should not be ready when initialized")
	}

	session.state = StateConfigured
	if !session.IsReady() {
		t.Error("session should be ready when configured")
	}

	session.state = StateLaunched
	if !session.IsReady() {
		t.Error("session should be ready when launched")
	}
}

func TestSessionRequireReady(t *testing.T) {
	session := NewSession("localhost", 6006)

	// Should error when not ready
	session.state = StateDisconnected
	if err := session.RequireReady(); err == nil {
		t.Error("RequireReady should error when state is disconnected")
	}

	// Should not error when ready
	session.state = StateConfigured
	if err := session.RequireReady(); err != nil {
		t.Errorf("RequireReady should not error when state is configured: %v", err)
	}

	session.state = StateLaunched
	if err := session.RequireReady(); err != nil {
		t.Errorf("RequireReady should not error when state is launched: %v", err)
	}
}

func TestGodotLaunchConfigValidation(t *testing.T) {
	// Create a temporary directory with project.godot for testing
	tempDir := t.TempDir()
	projectFile := filepath.Join(tempDir, "project.godot")
	if err := os.WriteFile(projectFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create test project.godot: %v", err)
	}

	tests := []struct {
		name    string
		config  *GodotLaunchConfig
		wantErr bool
	}{
		{
			name: "valid main scene config",
			config: &GodotLaunchConfig{
				Project: tempDir,
				Scene:   SceneLaunchMain,
			},
			wantErr: false,
		},
		{
			name: "valid current scene config",
			config: &GodotLaunchConfig{
				Project: tempDir,
				Scene:   SceneLaunchCurrent,
			},
			wantErr: false,
		},
		{
			name: "valid custom scene config",
			config: &GodotLaunchConfig{
				Project:   tempDir,
				Scene:     SceneLaunchCustom,
				ScenePath: "res://scenes/level1.tscn",
			},
			wantErr: false,
		},
		{
			name: "missing project path",
			config: &GodotLaunchConfig{
				Scene: SceneLaunchMain,
			},
			wantErr: true,
		},
		{
			name: "project.godot not found",
			config: &GodotLaunchConfig{
				Project: "/nonexistent/path",
				Scene:   SceneLaunchMain,
			},
			wantErr: true,
		},
		{
			name: "custom scene without path",
			config: &GodotLaunchConfig{
				Project: tempDir,
				Scene:   SceneLaunchCustom,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGodotLaunchConfigToLaunchArgs(t *testing.T) {
	config := &GodotLaunchConfig{
		Project:           "/path/to/project",
		Scene:             SceneLaunchMain,
		Platform:          PlatformHost,
		NoDebug:           false,
		Profiling:         true,
		DebugCollisions:   true,
		DebugPaths:        false,
		DebugNavigation:   false,
		AdditionalOptions: "--verbose",
	}

	args := config.ToLaunchArgs()

	// Check all expected fields
	if args["project"] != "/path/to/project" {
		t.Errorf("expected project to be /path/to/project, got %v", args["project"])
	}

	if args["scene"] != "main" {
		t.Errorf("expected scene to be main, got %v", args["scene"])
	}

	if args["platform"] != "host" {
		t.Errorf("expected platform to be host, got %v", args["platform"])
	}

	if args["noDebug"] != false {
		t.Errorf("expected noDebug to be false, got %v", args["noDebug"])
	}

	if args["profiling"] != true {
		t.Errorf("expected profiling to be true, got %v", args["profiling"])
	}

	if args["debug_collisions"] != true {
		t.Errorf("expected debug_collisions to be true, got %v", args["debug_collisions"])
	}

	if args["additional_options"] != "--verbose" {
		t.Errorf("expected additional_options to be --verbose, got %v", args["additional_options"])
	}
}

func TestGodotLaunchConfigSceneModes(t *testing.T) {
	tests := []struct {
		name      string
		scene     SceneLaunchMode
		scenePath string
		expected  string
	}{
		{
			name:     "main scene",
			scene:    SceneLaunchMain,
			expected: "main",
		},
		{
			name:     "current scene",
			scene:    SceneLaunchCurrent,
			expected: "current",
		},
		{
			name:      "custom scene",
			scene:     SceneLaunchCustom,
			scenePath: "res://scenes/level1.tscn",
			expected:  "res://scenes/level1.tscn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &GodotLaunchConfig{
				Project:   "/path/to/project",
				Scene:     tt.scene,
				ScenePath: tt.scenePath,
			}

			args := config.ToLaunchArgs()
			if args["scene"] != tt.expected {
				t.Errorf("expected scene to be %s, got %v", tt.expected, args["scene"])
			}
		})
	}
}

func TestTimeoutContextHelpers(t *testing.T) {
	// Test WithConnectTimeout
	ctx, cancel := WithConnectTimeout(context.Background())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Error("WithConnectTimeout should set a deadline")
	}

	expectedDeadline := time.Now().Add(DefaultConnectTimeout)
	if deadline.After(expectedDeadline.Add(time.Second)) || deadline.Before(expectedDeadline.Add(-time.Second)) {
		t.Errorf("WithConnectTimeout deadline is not close to expected: got %v, expected around %v", deadline, expectedDeadline)
	}

	// Test WithCommandTimeout
	ctx2, cancel2 := WithCommandTimeout(context.Background())
	defer cancel2()

	deadline2, ok := ctx2.Deadline()
	if !ok {
		t.Error("WithCommandTimeout should set a deadline")
	}

	expectedDeadline2 := time.Now().Add(DefaultCommandTimeout)
	if deadline2.After(expectedDeadline2.Add(time.Second)) || deadline2.Before(expectedDeadline2.Add(-time.Second)) {
		t.Errorf("WithCommandTimeout deadline is not close to expected: got %v, expected around %v", deadline2, expectedDeadline2)
	}

	// Test WithTimeout with custom duration
	customTimeout := 5 * time.Second
	ctx3, cancel3 := WithTimeout(context.Background(), customTimeout)
	defer cancel3()

	deadline3, ok := ctx3.Deadline()
	if !ok {
		t.Error("WithTimeout should set a deadline")
	}

	expectedDeadline3 := time.Now().Add(customTimeout)
	if deadline3.After(expectedDeadline3.Add(time.Second)) || deadline3.Before(expectedDeadline3.Add(-time.Second)) {
		t.Errorf("WithTimeout deadline is not close to expected: got %v, expected around %v", deadline3, expectedDeadline3)
	}
}

func TestTimeoutContextHelpersWithNilParent(t *testing.T) {
	// Test that helpers handle nil parent context gracefully
	ctx, cancel := WithConnectTimeout(nil)
	defer cancel()

	if ctx == nil {
		t.Error("WithConnectTimeout should not return nil context even with nil parent")
	}

	_, ok := ctx.Deadline()
	if !ok {
		t.Error("WithConnectTimeout should set a deadline even with nil parent")
	}
}

func TestClientSetBreakpoints_NotConnected(t *testing.T) {
	client := NewClient("localhost", 6006)
	ctx := context.Background()

	// Should error when not connected
	_, err := client.SetBreakpoints(ctx, "res://test.gd", []int{10})
	if err == nil {
		t.Error("SetBreakpoints should error when not connected")
	}
}

func TestClientContinue_NotConnected(t *testing.T) {
	client := NewClient("localhost", 6006)
	ctx := context.Background()

	// Should error when not connected
	_, err := client.Continue(ctx, 1)
	if err == nil {
		t.Error("Continue should error when not connected")
	}
}

func TestClientNext_NotConnected(t *testing.T) {
	client := NewClient("localhost", 6006)
	ctx := context.Background()

	// Should error when not connected
	_, err := client.Next(ctx, 1)
	if err == nil {
		t.Error("Next should error when not connected")
	}
}

func TestClientStepIn_NotConnected(t *testing.T) {
	client := NewClient("localhost", 6006)
	ctx := context.Background()

	// Should error when not connected
	_, err := client.StepIn(ctx, 1)
	if err == nil {
		t.Error("StepIn should error when not connected")
	}
}

func TestClientThreads_NotConnected(t *testing.T) {
	client := NewClient("localhost", 6006)
	ctx := context.Background()

	// Should error when not connected
	_, err := client.Threads(ctx)
	if err == nil {
		t.Error("Threads should error when not connected")
	}
}

func TestClientStackTrace_NotConnected(t *testing.T) {
	client := NewClient("localhost", 6006)
	ctx := context.Background()

	// Should error when not connected
	_, err := client.StackTrace(ctx, 1, 0, 20)
	if err == nil {
		t.Error("StackTrace should error when not connected")
	}
}

func TestClientScopes_NotConnected(t *testing.T) {
	client := NewClient("localhost", 6006)
	ctx := context.Background()

	// Should error when not connected
	_, err := client.Scopes(ctx, 1)
	if err == nil {
		t.Error("Scopes should error when not connected")
	}
}

func TestClientVariables_NotConnected(t *testing.T) {
	client := NewClient("localhost", 6006)
	ctx := context.Background()

	// Should error when not connected
	_, err := client.Variables(ctx, 1000)
	if err == nil {
		t.Error("Variables should error when not connected")
	}
}

func TestClientEvaluate_NotConnected(t *testing.T) {
	client := NewClient("localhost", 6006)
	ctx := context.Background()

	// Should error when not connected
	_, err := client.Evaluate(ctx, "1 + 1", 0, "repl")
	if err == nil {
		t.Error("Evaluate should error when not connected")
	}
}
