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

// TestJSONSchemaValidation_AnyType verifies that parameters accepting any type
// generate valid JSON Schema (omit type field instead of "type": "any")
func TestJSONSchemaValidation_AnyType(t *testing.T) {
	server := NewServer()
	server.RegisterTool(Tool{
		Name:        "test_any_type",
		Description: "Test tool with any-type parameter",
		Parameters: []Parameter{
			{
				Name:        "string_param",
				Type:        "string",
				Required:    true,
				Description: "A string parameter",
			},
			{
				Name:        "any_param",
				Type:        "", // Empty type = accepts any value
				Required:    true,
				Description: "An any-type parameter",
			},
			{
				Name:        "number_param",
				Type:        "number",
				Required:    false,
				Description: "A number parameter",
			},
		},
		Handler: func(params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	})

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      8,
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

	tool := result.Tools[0]
	schema := tool.InputSchema

	// Verify string_param has type "string"
	stringProp, ok := schema.Properties["string_param"]
	if !ok {
		t.Fatal("Expected string_param property")
	}
	if stringProp.Type != "string" {
		t.Errorf("Expected string_param type 'string', got '%s'", stringProp.Type)
	}

	// Verify any_param has NO type field (omitted)
	// This is valid JSON Schema for "accepts any type"
	anyProp, ok := schema.Properties["any_param"]
	if !ok {
		t.Fatal("Expected any_param property")
	}
	if anyProp.Type != "" {
		t.Errorf("Expected any_param type to be empty (omitted), got '%s'", anyProp.Type)
	}

	// Verify number_param has type "number"
	numberProp, ok := schema.Properties["number_param"]
	if !ok {
		t.Fatal("Expected number_param property")
	}
	if numberProp.Type != "number" {
		t.Errorf("Expected number_param type 'number', got '%s'", numberProp.Type)
	}

	// Verify required fields
	if len(schema.Required) != 2 {
		t.Fatalf("Expected 2 required parameters, got %d", len(schema.Required))
	}
}

// TestJSONSchemaValidation_InvalidAnyType demonstrates the bug that occurs
// when Type: "any" is used (it outputs invalid "type": "any" in JSON Schema)
// This test documents the anti-pattern - DO NOT use Type: "any" in real code!
func TestJSONSchemaValidation_InvalidAnyType(t *testing.T) {
	server := NewServer()

	// ANTI-PATTERN: This is WRONG - do NOT do this in real code!
	// Using Type: "any" will cause Claude API validation errors
	invalidTool := Tool{
		Name:        "invalid_tool",
		Description: "Tool with invalid type (anti-pattern demo)",
		Parameters: []Parameter{
			{
				Name:        "bad_param",
				Type:        "any", // ❌ WRONG: This is invalid JSON Schema!
				Required:    true,
				Description: "Invalid parameter type",
			},
		},
		Handler: func(params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}

	server.RegisterTool(invalidTool)

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      9,
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

	tool := result.Tools[0]
	badParam := tool.InputSchema.Properties["bad_param"]

	// Document the bug: Type: "any" gets output as-is in JSON Schema
	// This causes Claude API error: "JSON schema is invalid. It must match JSON Schema draft 2020-12"
	if badParam.Type != "any" {
		t.Errorf("Expected Type: \"any\" to be preserved (demonstrating the bug), got '%s'", badParam.Type)
	}

	// The correct approach is documented in TestJSONSchemaValidation_AnyType:
	// Use Type: "" (empty string) which gets omitted from JSON via omitempty tag
	t.Log("✓ This test documents the anti-pattern")
	t.Log("✓ To accept any type, use Type: \"\" instead of Type: \"any\"")
	t.Log("✓ See TestJSONSchemaValidation_AnyType for the correct pattern")
}
