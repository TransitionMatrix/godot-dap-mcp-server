package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== Minimal DAP Request Tester ===")
	fmt.Println("Tests Godot DAP server with minimal requests to expose Dictionary bugs")
	fmt.Println()

	// Connect to Godot DAP server
	conn, err := net.DialTimeout("tcp", "localhost:6006", 5*time.Second)
	if err != nil {
		fmt.Printf("Failed to connect to DAP server: %v\n", err)
		fmt.Println("Make sure Godot editor is running with DAP enabled")
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("✓ Connected to localhost:6006")
	fmt.Println()

	// Test cases - VALID DAP messages that should work but expose unsafe Dictionary access
	// Note: seq, type, and command are REQUIRED per DAP spec
	// arguments is OPTIONAL per DAP spec
	tests := []struct {
		name    string
		message map[string]interface{}
		expect  string
	}{
		{
			name: "Initialize without 'arguments' (VALID - arguments is optional)",
			message: map[string]interface{}{
				"seq":     1,
				"type":    "request",
				"command": "initialize",
			},
			expect: "Should trigger: p_params[\"arguments\"]",
		},
		{
			name: "Launch without 'arguments' (VALID - arguments is optional)",
			message: map[string]interface{}{
				"seq":     2,
				"type":    "request",
				"command": "launch",
			},
			expect: "Should trigger: p_params[\"arguments\"]",
		},
		{
			name: "setBreakpoints without 'breakpoints' array (VALID - optional per spec)",
			message: map[string]interface{}{
				"seq":     3,
				"type":    "request",
				"command": "setBreakpoints",
				"arguments": map[string]interface{}{
					"source": map[string]interface{}{"path": "/test.gd"},
				},
			},
			expect: "Should trigger: args[\"breakpoints\"]",
		},
		{
			name: "Launch with arguments but missing optional 'project' field",
			message: map[string]interface{}{
				"seq":     4,
				"type":    "request",
				"command": "launch",
				"arguments": map[string]interface{}{
					"scene": "main",
					// "project" is optional but code does: args["project"]
				},
			},
			expect: "Should trigger: args[\"project\"]",
		},
	}

	reader := bufio.NewReader(conn)

	for i, test := range tests {
		fmt.Printf("TEST %d: %s\n", i+1, test.name)
		fmt.Printf("  Expected: %s\n", test.expect)

		// Send request
		err := sendDAPMessage(conn, test.message)
		if err != nil {
			fmt.Printf("  ✗ Failed to send: %v\n", err)
			continue
		}

		// Try to read response (may fail if Godot crashes/errors)
		response, err := readDAPMessage(reader)
		if err != nil {
			fmt.Printf("  ✗ Failed to read response: %v\n", err)
			fmt.Println("  → This might indicate the request caused an error!")
		} else {
			fmt.Printf("  ✓ Got response: %s\n", truncate(response, 100))
		}

		fmt.Println()
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("=== Check Godot Console Output ===")
	fmt.Println("Look for errors like:")
	fmt.Println("  ERROR: Bug: Dictionary::operator[] used when there was no value for the given key")
	fmt.Println("  at: operator[] (core/variant/dictionary.cpp:136)")
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

func readDAPMessage(reader *bufio.Reader) (string, error) {
	// Read Content-Length header
	var contentLength int
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
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
	_, err := reader.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	return string(buf), nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
