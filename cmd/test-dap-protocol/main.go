package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorBold   = "\033[1m"
)

// TestCase represents a DAP protocol test
type TestCase struct {
	Name          string
	Message       map[string]interface{}
	SpecRequired  []string // Fields REQUIRED by DAP spec
	SpecOptional  []string // Fields OPTIONAL by DAP spec
	GodotExpects  []string // Fields Godot expects (causing errors if missing)
	ExpectedError string   // What Dictionary error this should trigger
}

func main() {
	printHeader()

	// Connect to Godot DAP server
	conn, err := connectToGodot()
	if err != nil {
		fmt.Printf("%s✗ Failed to connect: %v%s\n", colorRed, err, colorReset)
		fmt.Println("Make sure Godot editor is running with DAP enabled")
		fmt.Println("  Editor → Editor Settings → Network → Debug Adapter → Enable")
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("%s✓ Connected to localhost:6006%s\n\n", colorGreen, colorReset)

	// Define test cases
	tests := []TestCase{
		{
			Name: "Initialize with LLM-optimized client capabilities",
			Message: map[string]interface{}{
				"seq":     1,
				"type":    "request",
				"command": "initialize",
				"arguments": map[string]interface{}{
					"adapterID": "godot",

					// Client identification
					"clientID":   "godot-dap-mcp-server",
					"clientName": "Godot DAP MCP Server",

					// Human-friendly numbering (1-based)
					"linesStartAt1":   true,
					"columnsStartAt1": true,

					// Information richness
					"supportsVariableType": true,

					// Path format (standard file paths)
					"pathFormat": "path",

					// Advanced features we don't implement
					"supportsInvalidatedEvent":     false,
					"supportsProgressReporting":    false,
					"supportsRunInTerminalRequest": false,
					"supportsVariablePaging":       false,
				},
			},
			SpecRequired: []string{"seq", "type", "command", "arguments", "arguments.adapterID"},
			SpecOptional: []string{
				"arguments.clientID",
				"arguments.clientName",
				"arguments.linesStartAt1",
				"arguments.columnsStartAt1",
				"arguments.supportsVariableType",
				"arguments.pathFormat",
				"arguments.supportsInvalidatedEvent",
				"arguments.supportsProgressReporting",
				"arguments.supportsRunInTerminalRequest",
				"arguments.supportsVariablePaging",
			},
			GodotExpects:  []string{},
			ExpectedError: "(none - should work cleanly)",
		},
		{
			Name: "Launch - minimal (empty arguments, all defaults)",
			Message: map[string]interface{}{
				"seq":     2,
				"type":    "request",
				"command": "launch",
				"arguments": map[string]interface{}{
					// Empty - tests that Godot safely handles missing fields
				},
			},
			SpecRequired:  []string{"seq", "type", "command", "arguments"},
			SpecOptional:  []string{"arguments.request", "arguments.noDebug", "arguments.project", "arguments.scene", "arguments.platform", "arguments.device", "arguments.playArgs", "arguments.godot/custom_data"},
			GodotExpects:  []string{"All fields safe: project (.has check:171), godot/custom_data (.has check:178), noDebug (.get default:204), platform (.get default:208), scene (.get default:211), device (.get default:220), playArgs (.has check:189)"},
			ExpectedError: "(none - all reads Dictionary-safe, will use defaults: scene='main', platform='host', noDebug=false, device=-1, playArgs=[])",
		},
		{
			Name: "setBreakpoints with breakpoint at line 4 (configuration phase)",
			Message: map[string]interface{}{
				"seq":     3,
				"type":    "request",
				"command": "setBreakpoints",
				"arguments": map[string]interface{}{
					"source": map[string]interface{}{
						"path": "/Users/adp/Projects/godot-dap-mcp-server/tests/fixtures/test-project/test_script.gd",
					},
					"breakpoints": []map[string]interface{}{
						{"line": 4},
					},
				},
			},
			SpecRequired:  []string{"seq", "type", "command", "arguments", "arguments.source"},
			SpecOptional:  []string{"arguments.breakpoints"},
			GodotExpects:  []string{},
			ExpectedError: "(none - valid request with correct path)",
		},
		{
			Name: "configurationDone (triggers actual launch)",
			Message: map[string]interface{}{
				"seq":     4,
				"type":    "request",
				"command": "configurationDone",
			},
			SpecRequired:  []string{"seq", "type", "command"},
			SpecOptional:  []string{},
			GodotExpects:  []string{},
			ExpectedError: "(none - required step per DAP protocol)",
		},
		{
			Name: "next (step over) - assumes breakpoint was hit",
			Message: map[string]interface{}{
				"seq":     5,
				"type":    "request",
				"command": "next",
				"arguments": map[string]interface{}{
					"threadId": 1,
				},
			},
			SpecRequired:  []string{"seq", "type", "command", "arguments", "arguments.threadId"},
			SpecOptional:  []string{"arguments.granularity", "arguments.singleThread"},
			GodotExpects:  []string{"seq (safe .get), command (safe .get), does NOT read 'arguments' at all"},
			ExpectedError: "(none - threadId never accessed, no Dictionary errors possible)",
		},
		{
			Name: "stackTrace - discover WHERE execution stopped after step",
			Message: map[string]interface{}{
				"seq":     6,
				"type":    "request",
				"command": "stackTrace",
				"arguments": map[string]interface{}{
					"threadId": 1,
				},
			},
			SpecRequired:  []string{"seq", "type", "command", "arguments", "arguments.threadId"},
			SpecOptional:  []string{"arguments.startFrame", "arguments.levels", "arguments.format"},
			GodotExpects:  []string{"threadId (safe .get with default)"},
			ExpectedError: "(none - response shows current location: file, function, line, column)",
		},
		{
			Name: "stepIn (step into function) - Godot ignores ALL arguments",
			Message: map[string]interface{}{
				"seq":     7,
				"type":    "request",
				"command": "stepIn",
				"arguments": map[string]interface{}{
					"threadId": 1,
				},
			},
			SpecRequired:  []string{"seq", "type", "command", "arguments", "arguments.threadId"},
			SpecOptional:  []string{"arguments.granularity", "arguments.singleThread", "arguments.targetId"},
			GodotExpects:  []string{"seq (safe .get), command (safe .get), does NOT read 'arguments' at all"},
			ExpectedError: "(none - arguments completely ignored, no Dictionary access)",
		},
		{
			Name: "stackTrace - verify we stepped INTO calculate_sum function",
			Message: map[string]interface{}{
				"seq":     8,
				"type":    "request",
				"command": "stackTrace",
				"arguments": map[string]interface{}{
					"threadId": 1,
				},
			},
			SpecRequired:  []string{"seq", "type", "command", "arguments", "arguments.threadId"},
			SpecOptional:  []string{"arguments.startFrame", "arguments.levels", "arguments.format"},
			GodotExpects:  []string{"threadId (safe .get with default)"},
			ExpectedError: "(none - should show calculate_sum at line 15 in test_script.gd)",
		},
		{
			Name: "stepOut (step out of function) - Godot ignores ALL arguments",
			Message: map[string]interface{}{
				"seq":     9,
				"type":    "request",
				"command": "stepOut",
				"arguments": map[string]interface{}{
					"threadId": 1,
				},
			},
			SpecRequired:  []string{"seq", "type", "command", "arguments", "arguments.threadId"},
			SpecOptional:  []string{"arguments.granularity", "arguments.singleThread"},
			GodotExpects:  []string{"seq (safe .get), command (safe .get), does NOT read 'arguments' at all"},
			ExpectedError: "(none - arguments completely ignored, no Dictionary access)",
		},
		{
			Name: "stackTrace - verify we stepped OUT of calculate_sum back to caller",
			Message: map[string]interface{}{
				"seq":     10,
				"type":    "request",
				"command": "stackTrace",
				"arguments": map[string]interface{}{
					"threadId": 1,
				},
			},
			SpecRequired:  []string{"seq", "type", "command", "arguments", "arguments.threadId"},
			SpecOptional:  []string{"arguments.startFrame", "arguments.levels", "arguments.format"},
			GodotExpects:  []string{"threadId (safe .get with default)"},
			ExpectedError: "(none - should show caller context after stepping out of calculate_sum)",
		},
		{
			Name: "terminate - stop the game",
			Message: map[string]interface{}{
				"seq":     11,
				"type":    "request",
				"command": "terminate",
			},
			SpecRequired:  []string{"seq", "type", "command"},
			SpecOptional:  []string{"arguments.restart"},
			GodotExpects:  []string{"seq (safe .get), command (safe .get)"},
			ExpectedError: "(none - expect 'terminate' response + 'terminated' event + 'exited' event)",
		},
		{
			Name: "disconnect - cleanup (only if 'exited' event not received)",
			Message: map[string]interface{}{
				"seq":     12,
				"type":    "request",
				"command": "disconnect",
			},
			SpecRequired:  []string{"seq", "type", "command"},
			SpecOptional:  []string{"arguments.restart", "arguments.terminateDebuggee"},
			GodotExpects:  []string{"seq (safe .get), command (safe .get)"},
			ExpectedError: "(none - final cleanup command)",
		},
	}

	reader := bufio.NewReader(conn)
	stdin := bufio.NewReader(os.Stdin)

	printInstructions()
	waitForEnter(stdin)

	// Track events for summary
	var receivedTerminated bool
	var receivedExited bool
	var lastMessages []map[string]interface{}

	// Run tests
	for i, test := range tests {
		testNum := i + 1

		// Special handling for terminate test (test 11)
		if testNum == 11 {
			lastMessages = runTest(testNum, test, conn, reader, stdin, 5*time.Second)
			// Check for events
			receivedTerminated, receivedExited = checkTerminationEvents(lastMessages)
			continue
		}

		// Special handling for disconnect test (test 12)
		if testNum == 12 {
			if receivedExited {
				fmt.Printf("%s[SKIP] Test 12: Already received 'exited' event, disconnect not needed%s\n\n", colorYellow, colorReset)
				continue
			}
			fmt.Printf("%s[INFO] 'exited' event not received after 5s, sending disconnect...%s\n\n", colorYellow, colorReset)
			lastMessages = runTest(testNum, test, conn, reader, stdin, 5*time.Second)
			// Check again for events
			_, exitedFromDisconnect := checkTerminationEvents(lastMessages)
			receivedExited = receivedExited || exitedFromDisconnect
			continue
		}

		// Regular test
		runTest(testNum, test, conn, reader, stdin, 2*time.Second)
	}

	printSummary(receivedTerminated, receivedExited)
}

func printHeader() {
	fmt.Printf("%s%s", colorBold, strings.Repeat("═", 60))
	fmt.Printf("\n  Godot DAP Protocol Compliance Tester\n")
	fmt.Printf("%s%s\n\n", strings.Repeat("═", 60), colorReset)
	fmt.Println("This tool tests Godot's DAP server with minimal valid messages")
	fmt.Println("to demonstrate the difference between:")
	fmt.Printf("  • %sDAP Specification Requirements%s (what MUST be present)\n", colorGreen, colorReset)
	fmt.Printf("  • %sGodot's Expectations%s (what Godot assumes is present)\n", colorRed, colorReset)
	fmt.Println()
}

func printInstructions() {
	fmt.Printf("%sInstructions:%s\n", colorYellow, colorReset)
	fmt.Println("• This tool follows the DAP protocol sequence:")
	fmt.Println("  1. initialize → 2. launch → 3. setBreakpoints → 4. configurationDone")
	fmt.Println("• Configuration (breakpoints) happens between launch and configurationDone")
	fmt.Println("• Watch the Godot editor console for Dictionary errors")
	fmt.Println("• Press ENTER to send each test message")
	fmt.Println("• Press Ctrl-C to exit at any time")
	fmt.Println()
	fmt.Printf("%s→ Press ENTER to start testing%s ", colorYellow, colorReset)
}

func printSummary(receivedTerminated, receivedExited bool) {
	fmt.Printf("\n%s%s", colorBold, strings.Repeat("═", 60))
	fmt.Printf("\n  Testing Complete\n")
	fmt.Printf("%s%s\n\n", strings.Repeat("═", 60), colorReset)

	// Report on termination events
	fmt.Printf("%sTermination Event Status:%s\n", colorYellow, colorReset)
	if receivedTerminated {
		fmt.Printf("  %s✓ Received 'terminated' event%s\n", colorGreen, colorReset)
	} else {
		fmt.Printf("  %s✗ Did NOT receive 'terminated' event%s\n", colorRed, colorReset)
	}
	if receivedExited {
		fmt.Printf("  %s✓ Received 'exited' event%s\n", colorGreen, colorReset)
	} else {
		fmt.Printf("  %s✗ Did NOT receive 'exited' event%s\n", colorRed, colorReset)
	}
	fmt.Println()

	fmt.Printf("%sCheck Godot Console Output:%s\n", colorYellow, colorReset)
	fmt.Println("  • With UNSAFE code: Dictionary error messages")
	fmt.Println("  • With SAFE code: No errors, clean responses")
	fmt.Println()
	fmt.Printf("%sExpected error format:%s\n", colorYellow, colorReset)
	fmt.Println("  ERROR: Bug: Dictionary::operator[] used when there was no value")
	fmt.Println("  at: operator[] (core/variant/dictionary.cpp:136)")
	fmt.Println()
}

func runTest(testNum int, test TestCase, conn net.Conn, reader *bufio.Reader, stdin *bufio.Reader, timeout time.Duration) []map[string]interface{} {
	// Print test header
	fmt.Printf("%s%s\n", colorBold, strings.Repeat("━", 60))
	fmt.Printf("%sTEST %d: %s%s\n", colorBlue, testNum, test.Name, colorReset)
	fmt.Printf("%s%s%s\n\n", colorBold, strings.Repeat("━", 60), colorReset)

	// Show what the DAP spec says
	fmt.Printf("%sDAP Specification:%s\n", colorGreen, colorReset)
	fmt.Printf("  ✓ Required fields: %s\n", strings.Join(test.SpecRequired, ", "))
	fmt.Printf("  ○ Optional fields: %s\n", strings.Join(test.SpecOptional, ", "))
	fmt.Println()

	// Show what Godot expects
	fmt.Printf("%sGodot Implementation:%s\n", colorRed, colorReset)
	fmt.Printf("  ! Expects (but shouldn't require): %s\n", strings.Join(test.GodotExpects, ", "))
	fmt.Printf("  ! Will trigger error: %s\n", test.ExpectedError)
	fmt.Println()

	// Show the message that will be sent
	fmt.Printf("%s>>> WILL SEND TO GODOT >>>%s\n", colorGreen, colorReset)
	prettyJSON, _ := json.MarshalIndent(test.Message, "", "  ")
	fmt.Println(string(prettyJSON))
	fmt.Println()

	fmt.Printf("%s→ Press ENTER to SEND message (Ctrl-C to exit)%s ", colorYellow, colorReset)
	waitForEnter(stdin)

	// Send the message
	fmt.Printf("\n%s⟳ Sending message to Godot...%s\n", colorGreen+colorBold, colorReset)
	err := sendDAPMessage(conn, test.Message)
	if err != nil {
		fmt.Printf("%s✗ Failed to send: %v%s\n\n", colorRed, err, colorReset)
		return nil
	}

	// Read all responses (with timeout)
	fmt.Printf("%s✓ Message sent, waiting for response (timeout: %v)...%s\n\n", colorGreen+colorBold, timeout, colorReset)
	messages := readAllDAPMessages(conn, reader, timeout)

	// Display responses
	fmt.Printf("%s<<< RECEIVED FROM GODOT <<<<%s\n\n", colorRed, colorReset)
	if len(messages) == 0 {
		fmt.Printf("%s(No response received - Godot may have encountered an error)%s\n", colorYellow, colorReset)
	} else {
		for i, msg := range messages {
			if i > 0 {
				fmt.Printf("\n%s━━━ Additional Message ━━━%s\n\n", colorRed, colorReset)
			}
			prettyJSON, _ := json.MarshalIndent(msg, "", "  ")
			fmt.Println(string(prettyJSON))
		}
	}

	fmt.Println()
	fmt.Printf("%s→ Press ENTER to continue to next test (Ctrl-C to exit)%s ", colorYellow, colorReset)
	waitForEnter(stdin)
	fmt.Println()

	return messages
}

func connectToGodot() (net.Conn, error) {
	return net.DialTimeout("tcp", "localhost:6006", 5*time.Second)
}

func sendDAPMessage(conn net.Conn, message map[string]interface{}) error {
	// Encode to JSON
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Send with Content-Length header
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	_, err = conn.Write([]byte(header))
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

func readAllDAPMessages(conn net.Conn, reader *bufio.Reader, timeout time.Duration) []map[string]interface{} {
	messages := []map[string]interface{}{}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		// Set read deadline on connection
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}

		conn.SetReadDeadline(time.Now().Add(remaining))

		msg, err := readDAPMessage(reader)
		if err != nil {
			// Check if it's a timeout (expected when no more messages)
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			// Other errors also mean we're done
			break
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(msg), &parsed); err == nil {
			messages = append(messages, parsed)
		}
	}

	// Clear deadline
	conn.SetReadDeadline(time.Time{})

	return messages
}

func readDAPMessage(reader *bufio.Reader) (string, error) {
	// Read Content-Length header
	var contentLength int
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return "", fmt.Errorf("EOF while reading headers")
			}
			return "", fmt.Errorf("failed to read header: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break // End of headers
		}

		if strings.HasPrefix(line, "Content-Length:") {
			lengthStr := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			contentLength, err = strconv.Atoi(lengthStr)
			if err != nil {
				return "", fmt.Errorf("invalid Content-Length: %w", err)
			}
		}
	}

	if contentLength == 0 {
		return "", fmt.Errorf("no Content-Length header")
	}

	// Read message body
	buf := make([]byte, contentLength)
	n, err := io.ReadFull(reader, buf)
	if err != nil {
		return "", fmt.Errorf("failed to read body (read %d/%d bytes): %w", n, contentLength, err)
	}

	return string(buf), nil
}

func waitForEnter(stdin *bufio.Reader) {
	stdin.ReadString('\n')
}

func checkTerminationEvents(messages []map[string]interface{}) (hasTerminated bool, hasExited bool) {
	for _, msg := range messages {
		msgType, ok := msg["type"].(string)
		if !ok || msgType != "event" {
			continue
		}

		event, ok := msg["event"].(string)
		if !ok {
			continue
		}

		if event == "terminated" {
			hasTerminated = true
		}
		if event == "exited" {
			hasExited = true
		}
	}
	return hasTerminated, hasExited
}
