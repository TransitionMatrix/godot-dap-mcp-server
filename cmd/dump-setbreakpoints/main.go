package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/go-dap"
)

func main() {
	// Create a setBreakpoints request exactly as our client does
	request := &dap.SetBreakpointsRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  3,
				Type: "request",
			},
			Command: "setBreakpoints",
		},
		Arguments: dap.SetBreakpointsArguments{
			Source: dap.Source{
				Path: "res://src/scenes/main_scene/main_scene.gd",
			},
			Breakpoints: []dap.SourceBreakpoint{
				{Line: 56},
			},
		},
	}

	// Serialize to JSON to see what's actually sent
	jsonBytes, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== SetBreakpoints Request JSON ===")
	fmt.Println(string(jsonBytes))
	fmt.Println("===================================")

	// Also create an Initialize request for comparison
	initRequest := &dap.InitializeRequest{
		Request: dap.Request{
			ProtocolMessage: dap.ProtocolMessage{
				Seq:  1,
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

	jsonBytes2, err := json.MarshalIndent(initRequest, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n=== Initialize Request JSON (for comparison) ===")
	fmt.Println(string(jsonBytes2))
	fmt.Println("===============================================")

	// Test actual DAP protocol encoding
	fmt.Println("\n=== ACTUAL DAP PROTOCOL OUTPUT (SetBreakpoints) ===")
	var buf bytes.Buffer
	if err := dap.WriteProtocolMessage(&buf, request); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total bytes: %d\n", buf.Len())
	fmt.Println("Raw output:")
	fmt.Println(buf.String())
	fmt.Println("===================================================")

	// Wait a moment so output is visible
	time.Sleep(100 * time.Millisecond)
}
