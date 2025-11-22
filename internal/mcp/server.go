package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
)

// Server is the core MCP server that handles tool registration and request routing
type Server struct {
	transport *Transport
	tools     map[string]Tool
}

// NewServer creates a new MCP server with default stdio transport
func NewServer() *Server {
	return &Server{
		transport: NewTransport(),
		tools:     make(map[string]Tool),
	}
}

// NewServerWithTransport creates a new MCP server with custom transport (for testing)
func NewServerWithTransport(transport *Transport) *Server {
	return &Server{
		transport: transport,
		tools:     make(map[string]Tool),
	}
}

// RegisterTool registers a new tool with the server
func (s *Server) RegisterTool(tool Tool) {
	s.tools[tool.Name] = tool
	log.Printf("Registered tool: %s", tool.Name)
}

// ListenAndServe starts the server and processes requests until EOF or error
func (s *Server) ListenAndServe() error {
	log.Println("MCP server started, listening on stdin...")

	for {
		// Read next request
		req, err := s.transport.ReadRequest()
		if err != nil {
			if err == io.EOF {
				// Clean shutdown
				log.Println("EOF reached, shutting down...")
				return nil
			}
			// Log error but continue (send error response if possible)
			log.Printf("Error reading request: %v", err)
			continue
		}

		// Handle request
		resp := s.handleRequest(req)

		// If request ID is nil, it's a notification - do not send response
		if req.ID == nil {
			continue
		}

		// Send response
		if err := s.transport.WriteResponse(resp); err != nil {
			log.Printf("Error writing response: %v", err)
			// If we can't write responses, we should shut down
			return fmt.Errorf("failed to write response: %w", err)
		}
	}
}

// handleRequest routes a request to the appropriate handler
func (s *Server) handleRequest(req *MCPRequest) MCPResponse {
	// Handle notifications that don't require a response
	if req.ID == nil {
		switch req.Method {
		case "notifications/initialized":
			// Just log and return empty response (which won't be sent)
			log.Println("Client initialized notification received")
			return MCPResponse{}
		}
	}

	var id interface{}
	if req.ID != nil {
		id = *req.ID
	}

	switch req.Method {
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	case "initialize":
		return s.handleInitialize(req)
	default:
		return s.errorResponse(id, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}

// handleInitialize handles the initialize method (optional but good practice)
func (s *Server) handleInitialize(req *MCPRequest) MCPResponse {
	var id interface{}
	if req.ID != nil {
		id = *req.ID
	}

	return s.successResponse(id, map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "godot-dap-mcp-server",
			"version": "0.1.0",
		},
	})
}

// handleToolsList handles the tools/list method
func (s *Server) handleToolsList(req *MCPRequest) MCPResponse {
	tools := make([]ToolMetadata, 0, len(s.tools))

	for _, tool := range s.tools {
		// Build input schema from parameters
		properties := make(map[string]PropertyDefinition)
		required := []string{}

		for _, param := range tool.Parameters {
			properties[param.Name] = PropertyDefinition{
				Type:        param.Type,
				Description: param.Description,
				Default:     param.Default,
			}

			if param.Required {
				required = append(required, param.Name)
			}
		}

		tools = append(tools, ToolMetadata{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: ToolInputSchema{
				Type:       "object",
				Properties: properties,
				Required:   required,
			},
		})
	}

	var id interface{}
	if req.ID != nil {
		id = *req.ID
	}

	return s.successResponse(id, ToolListResult{Tools: tools})
}

// handleToolsCall handles the tools/call method
func (s *Server) handleToolsCall(req *MCPRequest) MCPResponse {
	var id interface{}
	if req.ID != nil {
		id = *req.ID
	}

	// Extract tool name
	name, ok := req.Params["name"].(string)
	if !ok {
		return s.errorResponse(id, -32602, "missing or invalid 'name' parameter")
	}

	// Find tool
	tool, exists := s.tools[name]
	if !exists {
		return s.errorResponse(id, -32601, fmt.Sprintf("tool not found: %s", name))
	}

	// Extract arguments
	arguments, ok := req.Params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	// Apply defaults and validate required parameters
	params := s.applyDefaults(tool, arguments)
	if err := s.validateRequired(tool, params); err != nil {
		return s.errorResponse(id, -32602, err.Error())
	}

	// Call tool handler
	result, err := tool.Handler(params)
	if err != nil {
		return s.errorResponse(id, -32000, fmt.Sprintf("tool execution failed: %v", err))
	}

	// Format result as tool call result
	toolResult := ToolCallResult{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: formatResult(result),
			},
		},
	}

	return s.successResponse(id, toolResult)
}

// applyDefaults applies default values to parameters
func (s *Server) applyDefaults(tool Tool, arguments map[string]interface{}) map[string]interface{} {
	params := make(map[string]interface{})

	// Copy provided arguments
	for k, v := range arguments {
		params[k] = v
	}

	// Apply defaults for missing parameters
	for _, param := range tool.Parameters {
		if _, exists := params[param.Name]; !exists && param.Default != nil {
			params[param.Name] = param.Default
		}
	}

	return params
}

// validateRequired checks that all required parameters are present
func (s *Server) validateRequired(tool Tool, params map[string]interface{}) error {
	for _, param := range tool.Parameters {
		if param.Required {
			if _, exists := params[param.Name]; !exists {
				return fmt.Errorf("missing required parameter: %s", param.Name)
			}
		}
	}
	return nil
}

// formatResult converts a tool result to a string
func formatResult(result interface{}) string {
	switch v := result.(type) {
	case string:
		return v
	case error:
		return v.Error()
	default:
		// For complex types, return JSON representation
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			// Fallback to fmt if JSON marshaling fails
			return fmt.Sprintf("%+v", v)
		}
		return string(jsonBytes)
	}
}

// successResponse creates a success response
func (s *Server) successResponse(id interface{}, result interface{}) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// errorResponse creates an error response
func (s *Server) errorResponse(id interface{}, code int, message string) MCPResponse {
	return MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}
}
