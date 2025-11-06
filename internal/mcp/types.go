package mcp

// MCPRequest represents an incoming JSON-RPC 2.0 request from the MCP client
type MCPRequest struct {
	JSONRPC string                 `json:"jsonrpc"` // Always "2.0"
	ID      interface{}            `json:"id"`      // Request ID (number or string)
	Method  string                 `json:"method"`  // Method name (e.g., "tools/list", "tools/call")
	Params  map[string]interface{} `json:"params"`  // Method parameters
}

// MCPResponse represents an outgoing JSON-RPC 2.0 response to the MCP client
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`           // Always "2.0"
	ID      interface{} `json:"id"`                // Must match request ID
	Result  interface{} `json:"result,omitempty"`  // Success result
	Error   *MCPError   `json:"error,omitempty"`   // Error details (mutually exclusive with Result)
}

// MCPError represents a JSON-RPC 2.0 error
type MCPError struct {
	Code    int         `json:"code"`              // Error code
	Message string      `json:"message"`           // Human-readable error message
	Data    interface{} `json:"data,omitempty"`    // Optional additional error data
}

// Tool represents a callable MCP tool with metadata and handler
type Tool struct {
	Name        string                                                  // Tool name (e.g., "godot_connect")
	Description string                                                  // AI-friendly description
	Parameters  []Parameter                                             // Tool parameters
	Handler     func(params map[string]interface{}) (interface{}, error) // Handler function
}

// Parameter represents a tool parameter definition
type Parameter struct {
	Name        string      // Parameter name
	Type        string      // Parameter type ("string", "integer", "boolean", "object", "array")
	Required    bool        // Whether parameter is required
	Description string      // Human-readable description
	Default     interface{} // Default value if not provided
}

// ToolListResult represents the response to tools/list
type ToolListResult struct {
	Tools []ToolMetadata `json:"tools"`
}

// ToolMetadata represents tool metadata for tools/list response
type ToolMetadata struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	InputSchema ToolInputSchema      `json:"inputSchema"`
}

// ToolInputSchema defines the JSON schema for tool parameters
type ToolInputSchema struct {
	Type       string                        `json:"type"`       // Always "object"
	Properties map[string]PropertyDefinition `json:"properties"` // Parameter definitions
	Required   []string                      `json:"required"`   // Required parameter names
}

// PropertyDefinition defines a single parameter's schema
type PropertyDefinition struct {
	Type        string      `json:"type"`                  // Parameter type
	Description string      `json:"description"`           // Parameter description
	Default     interface{} `json:"default,omitempty"`     // Default value
}

// ToolCallResult represents the response to tools/call
type ToolCallResult struct {
	Content []ContentBlock `json:"content"`
}

// ContentBlock represents a content block in tool response
type ContentBlock struct {
	Type string `json:"type"` // "text" or other types
	Text string `json:"text"` // Text content
}
