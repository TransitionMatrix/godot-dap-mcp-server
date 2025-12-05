package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
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
	SpecRequired  []string      // Fields REQUIRED by DAP spec
	SpecOptional  []string      // Fields OPTIONAL by DAP spec
	GodotExpects  []string      // Fields Godot expects (causing errors if missing)
	ExpectedError string        // What Dictionary error this should trigger
	Timeout       time.Duration // Timeout for waiting for response (default: 2s)
}

var registry map[string]TestCase

func init() {
	// Calculate default path for breakpoints
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// Assuming we run from project root
	defaultScriptPath := filepath.Join(wd, "tests/fixtures/test-project/test_script.gd")

	registry = map[string]TestCase{
		"initialize": {
			Name: "Initialize with LLM-optimized client capabilities",
			Message: map[string]interface{}{
				"type":    "request",
				"command": "initialize",
				"arguments": map[string]interface{}{
					"adapterID":                    "godot",
					"clientID":                     "godot-dap-mcp-server",
					"clientName":                   "Godot DAP MCP Server",
					"linesStartAt1":                true,
					"columnsStartAt1":              true,
					"supportsVariableType":         true,
					"pathFormat":                   "path",
					"supportsInvalidatedEvent":     false,
					"supportsProgressReporting":    false,
					"supportsRunInTerminalRequest": false,
					"supportsVariablePaging":       false,
				},
			},
			SpecRequired:  []string{"seq", "type", "command", "arguments", "arguments.adapterID"},
			ExpectedError: "(none - should work cleanly)",
			Timeout:       2 * time.Second,
		},
		"launch": {
			Name: "Launch - minimal (empty arguments, all defaults)",
			Message: map[string]interface{}{
				"type":      "request",
				"command":   "launch",
				"arguments": map[string]interface{}{},
			},
			SpecRequired:  []string{"seq", "type", "command", "arguments"},
			ExpectedError: "(none - all reads Dictionary-safe)",
			Timeout:       2 * time.Second,
		},
		"attach": {
			Name: "Attach - connect to running game",
			Message: map[string]interface{}{
				"type":      "request",
				"command":   "attach",
				"arguments": map[string]interface{}{},
			},
			SpecRequired:  []string{"seq", "type", "command", "arguments"},
			ExpectedError: "(none - should attach if game is running)",
			Timeout:       2 * time.Second,
		},
		"setBreakpoints": {
			Name: "setBreakpoints with breakpoint at line 4",
			Message: map[string]interface{}{
				"type":    "request",
				"command": "setBreakpoints",
				"arguments": map[string]interface{}{
					"source": map[string]interface{}{
						"path": defaultScriptPath,
					},
					"breakpoints": []map[string]interface{}{
						{"line": 4},
					},
				},
			},
			SpecRequired:  []string{"seq", "type", "command", "arguments", "arguments.source"},
			ExpectedError: "(none - valid request with correct path)",
			Timeout:       2 * time.Second,
		},
		"configurationDone": {
			Name: "configurationDone",
			Message: map[string]interface{}{
				"type":    "request",
				"command": "configurationDone",
			},
			SpecRequired:  []string{"seq", "type", "command"},
			ExpectedError: "(none - required step per DAP protocol)",
			Timeout:       5 * time.Second,
		},
		"next": {
			Name: "next (step over)",
			Message: map[string]interface{}{
				"type":    "request",
				"command": "next",
				"arguments": map[string]interface{}{
					"threadId": 1,
				},
			},
			SpecRequired:  []string{"seq", "type", "command", "arguments", "arguments.threadId"},
			ExpectedError: "(none - threadId never accessed)",
			Timeout:       2 * time.Second,
		},
		"stackTrace": {
			Name: "stackTrace",
			Message: map[string]interface{}{
				"type":    "request",
				"command": "stackTrace",
				"arguments": map[string]interface{}{
					"threadId": 1,
				},
			},
			SpecRequired:  []string{"seq", "type", "command", "arguments", "arguments.threadId"},
			ExpectedError: "(none - response shows current location)",
			Timeout:       2 * time.Second,
		},
		"stepIn": {
			Name: "stepIn (step into function)",
			Message: map[string]interface{}{
				"type":    "request",
				"command": "stepIn",
				"arguments": map[string]interface{}{
					"threadId": 1,
				},
			},
			SpecRequired:  []string{"seq", "type", "command", "arguments", "arguments.threadId"},
			ExpectedError: "(none - arguments completely ignored)",
			Timeout:       2 * time.Second,
		},
		"stepOut": {
			Name: "stepOut (step out of function)",
			Message: map[string]interface{}{
				"type":    "request",
				"command": "stepOut",
				"arguments": map[string]interface{}{
					"threadId": 1,
				},
			},
			SpecRequired:  []string{"seq", "type", "command", "arguments", "arguments.threadId"},
			ExpectedError: "(none - arguments completely ignored)",
			Timeout:       2 * time.Second,
		},
		"terminate": {
			Name: "terminate - stop the game",
			Message: map[string]interface{}{
				"type":    "request",
				"command": "terminate",
			},
			SpecRequired:  []string{"seq", "type", "command"},
			ExpectedError: "(none - expect 'terminate' response)",
			Timeout:       5 * time.Second,
		},
		"disconnect": {
			Name: "disconnect - cleanup",
			Message: map[string]interface{}{
				"type":    "request",
				"command": "disconnect",
			},
			SpecRequired:  []string{"seq", "type", "command"},
			ExpectedError: "(none - final cleanup command)",
			Timeout:       5 * time.Second,
		},
	}
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: go run main.go <scenario_file>")
		os.Exit(1)
	}

	scenarioFile := args[0]
	commands, err := readScenario(scenarioFile)
	if err != nil {
		fmt.Printf("Error reading scenario: %v\n", err)
		os.Exit(1)
	}

	printHeader(scenarioFile)

	conn, err := connectToGodot()
	if err != nil {
		fmt.Printf("%s✗ Failed to connect: %v%s\n", colorRed, err, colorReset)
		fmt.Println("Make sure Godot editor is running with DAP enabled")
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("%s✓ Connected to localhost:6006%s\n\n", colorGreen, colorReset)

	reader := bufio.NewReader(conn)
	stdin := bufio.NewReader(os.Stdin)

	printInstructions()
	waitForEnter(stdin)

	var receivedTerminated bool
	var receivedExited bool

	seqCounter := 1

	for _, cmdName := range commands {
		testCase, ok := registry[cmdName]
		if !ok {
			fmt.Printf("%s[WARN] Unknown command in scenario: %s%s\n", colorYellow, cmdName, colorReset)
			continue
		}

		// Update Sequence Number
		testCase.Message["seq"] = seqCounter
		seqCounter++

		// Run Test
		messages := runTest(seqCounter-1, testCase, conn, reader, stdin)

		// Check for events
		term, exited := checkTerminationEvents(messages)
		if term {
			receivedTerminated = true
		}
		if exited {
			receivedExited = true
		}

		// Short circuit if we are disconnecting and already exited (logic from original)
		if cmdName == "disconnect" && receivedExited {
			fmt.Printf("%s[INFO] Already received 'exited', proceeding with disconnect anyway for cleanliness%s\n", colorYellow, colorReset)
		}
	}

	printSummary(receivedTerminated, receivedExited)
}

func readScenario(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var commands []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			commands = append(commands, line)
		}
	}
	return commands, scanner.Err()
}

// ... (Copy helper functions from test-dap-protocol: connectToGodot, sendDAPMessage, readAllDAPMessages, readDAPMessage, checkTerminationEvents)
// ... (Also printHeader, printInstructions, printSummary, runTest)

func connectToGodot() (net.Conn, error) {
	return net.DialTimeout("tcp", "localhost:6006", 5*time.Second)
}

func sendDAPMessage(conn net.Conn, message map[string]interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
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
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		conn.SetReadDeadline(time.Now().Add(remaining))

		msg, err := readDAPMessage(reader)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			break
		}
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(msg), &parsed); err == nil {
			messages = append(messages, parsed)
		}
	}
	conn.SetReadDeadline(time.Time{})
	return messages
}

func readDAPMessage(reader *bufio.Reader) (string, error) {
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
			break
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

func printHeader(scenario string) {
	fmt.Printf("%s%s", colorBold, strings.Repeat("═", 60))
	fmt.Printf("\n  Godot DAP Runner: %s\n", scenario)
	fmt.Printf("%s%s\n\n", strings.Repeat("═", 60), colorReset)
}

func printInstructions() {
	fmt.Printf("%sInstructions:%s\n", colorYellow, colorReset)
	fmt.Println("• Press ENTER to send each test message")
	fmt.Println("• Press Ctrl-C to exit at any time")
	fmt.Printf("%s→ Press ENTER to start testing%s ", colorYellow, colorReset)
}

func printSummary(receivedTerminated, receivedExited bool) {
	fmt.Printf("\n%s%s", colorBold, strings.Repeat("═", 60))
	fmt.Printf("\n  Testing Complete\n")
	fmt.Printf("%s%s\n\n", strings.Repeat("═", 60), colorReset)
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
}

func runTest(testNum int, test TestCase, conn net.Conn, reader *bufio.Reader, stdin *bufio.Reader) []map[string]interface{} {
	fmt.Printf("%s%s\n", colorBold, strings.Repeat("━", 60))
	fmt.Printf("%sSTEP %d: %s%s\n", colorBlue, testNum, test.Name, colorReset)
	fmt.Printf("%s%s%s\n\n", colorBold, strings.Repeat("━", 60), colorReset)

	fmt.Printf("%sGodot Implementation:%s\n", colorRed, colorReset)
	fmt.Printf("  ! Will trigger error: %s\n", test.ExpectedError)
	fmt.Println()

	fmt.Printf("%s>>> WILL SEND TO GODOT >>>%s\n", colorGreen, colorReset)
	prettyJSON, _ := json.MarshalIndent(test.Message, "", "  ")
	fmt.Println(string(prettyJSON))
	fmt.Println()

	fmt.Printf("%s→ Press ENTER to SEND message (Ctrl-C to exit)%s ", colorYellow, colorReset)
	waitForEnter(stdin)

	fmt.Printf("\n%s⟳ Sending message to Godot...%s\n", colorGreen+colorBold, colorReset)
	err := sendDAPMessage(conn, test.Message)
	if err != nil {
		fmt.Printf("%s✗ Failed to send: %v%s\n\n", colorRed, err, colorReset)
		return nil
	}

	fmt.Printf("%s✓ Message sent, waiting for response (timeout: %v)...%s\n\n", colorGreen+colorBold, test.Timeout, colorReset)
	messages := readAllDAPMessages(conn, reader, test.Timeout)

	fmt.Printf("%s<<< RECEIVED FROM GODOT <<<<%s\n\n", colorRed, colorReset)
	if len(messages) == 0 {
		fmt.Printf("%s(No response received)%s\n", colorYellow, colorReset)
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
	return messages
}
