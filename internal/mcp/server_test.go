package mcp

import (
	"fmt"
	"testing"
)

// TestRegisterTool verifies that tools can be registered
func TestRegisterTool(t *testing.T) {
	server := NewServer()

	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters:  []Parameter{},
		Handler: func(params map[string]interface{}) (interface{}, error) {
			return "test result", nil
		},
	}

	server.RegisterTool(tool)

	if _, exists := server.tools["test_tool"]; !exists {
		t.Error("Tool was not registered")
	}
}

// TestHandleInitialize verifies the initialize method response
func TestHandleInitialize(t *testing.T) {
	server := NewServer()
	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  map[string]interface{}{},
	}

	resp := server.handleInitialize(req)

	if resp.ID != 1 {
		t.Errorf("Expected response ID 1, got %v", resp.ID)
	}
	if resp.Error != nil {
		t.Errorf("Expected no error, got %+v", resp.Error)
	}
	if resp.Result == nil {
		t.Fatal("Expected result, got nil")
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocol version 2024-11-05, got %v", result["protocolVersion"])
	}
}

// TestHandleToolsList_Empty verifies tools/list with no tools registered
func TestHandleToolsList_Empty(t *testing.T) {
	server := NewServer()
	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	}

	resp := server.handleToolsList(req)

	if resp.Error != nil {
		t.Errorf("Expected no error, got %+v", resp.Error)
	}

	result, ok := resp.Result.(ToolListResult)
	if !ok {
		t.Fatal("Expected ToolListResult")
	}
	if len(result.Tools) != 0 {
		t.Errorf("Expected 0 tools, got %d", len(result.Tools))
	}
}

// TestHandleToolsList_WithTools verifies tools/list returns registered tools
func TestHandleToolsList_WithTools(t *testing.T) {
	server := NewServer()
	server.RegisterTool(Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: []Parameter{
			{
				Name:        "param1",
				Type:        "string",
				Required:    true,
				Description: "First parameter",
			},
		},
		Handler: func(params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	})

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	}

	resp := server.handleToolsList(req)

	if resp.Error != nil {
		t.Fatalf("Expected no error, got %+v", resp.Error)
	}

	result, ok := resp.Result.(ToolListResult)
	if !ok {
		t.Fatal("Expected ToolListResult")
	}
	if len(result.Tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(result.Tools))
	}
	if result.Tools[0].Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", result.Tools[0].Name)
	}
}

// TestHandleToolsCall_Success verifies successful tool execution
func TestHandleToolsCall_Success(t *testing.T) {
	server := NewServer()
	server.RegisterTool(Tool{
		Name:        "echo",
		Description: "Echoes input",
		Parameters: []Parameter{
			{
				Name:        "message",
				Type:        "string",
				Required:    false,
				Default:     "default",
				Description: "Message to echo",
			},
		},
		Handler: func(params map[string]interface{}) (interface{}, error) {
			msg := params["message"].(string)
			return fmt.Sprintf("Echo: %s", msg), nil
		},
	})

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "echo",
			"arguments": map[string]interface{}{
				"message": "test",
			},
		},
	}

	resp := server.handleToolsCall(req)

	if resp.Error != nil {
		t.Fatalf("Expected no error, got %+v", resp.Error)
	}

	result, ok := resp.Result.(ToolCallResult)
	if !ok {
		t.Fatal("Expected ToolCallResult")
	}
	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(result.Content))
	}
	if result.Content[0].Text != "Echo: test" {
		t.Errorf("Expected 'Echo: test', got '%s'", result.Content[0].Text)
	}
}

// TestHandleToolsCall_ToolNotFound verifies error when tool doesn't exist
func TestHandleToolsCall_ToolNotFound(t *testing.T) {
	server := NewServer()
	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "nonexistent",
			"arguments": map[string]interface{}{},
		},
	}

	resp := server.handleToolsCall(req)

	if resp.Error == nil {
		t.Fatal("Expected error for nonexistent tool, got none")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("Expected error code -32601, got %d", resp.Error.Code)
	}
}

// TestHandleToolsCall_MissingRequired verifies error when required param is missing
func TestHandleToolsCall_MissingRequired(t *testing.T) {
	server := NewServer()
	server.RegisterTool(Tool{
		Name:        "require_param",
		Description: "Requires a parameter",
		Parameters: []Parameter{
			{
				Name:        "required_param",
				Type:        "string",
				Required:    true,
				Description: "A required parameter",
			},
		},
		Handler: func(params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	})

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      6,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "require_param",
			"arguments": map[string]interface{}{}, // Missing required_param
		},
	}

	resp := server.handleToolsCall(req)

	if resp.Error == nil {
		t.Fatal("Expected error for missing required parameter, got none")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("Expected error code -32602, got %d", resp.Error.Code)
	}
}

// TestHandleRequest_UnknownMethod verifies error for unknown methods
func TestHandleRequest_UnknownMethod(t *testing.T) {
	server := NewServer()
	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      7,
		Method:  "unknown/method",
		Params:  map[string]interface{}{},
	}

	resp := server.handleRequest(req)

	if resp.Error == nil {
		t.Fatal("Expected error for unknown method, got none")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("Expected error code -32601, got %d", resp.Error.Code)
	}
}
