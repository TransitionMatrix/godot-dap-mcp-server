package tools

import (
	"testing"

	"github.com/TransitionMatrix/godot-dap-mcp-server/internal/mcp"
)

func TestAdvancedTools_Registration(t *testing.T) {
	server := mcp.NewServer()
	RegisterAdvancedTools(server)

	// Verify registration doesn't panic
	// The tools should be registered successfully
}

func TestAdvancedTools_RequireSession(t *testing.T) {
	// Reset global session
	globalSession = nil

	// All advanced tools should fail when no session exists
	// This is tested through the GetSession() function
	_, err := GetSession()
	if err == nil {
		t.Error("Advanced tools should require an active session")
	}
}

func TestValidVariableName(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		// Valid names
		{"simple", "health", true},
		{"with_underscore", "player_health", true},
		{"starts_with_underscore", "_internal", true},
		{"with_numbers", "var123", true},
		{"camelCase", "playerHealth", true},
		{"all_caps", "MAX_HEALTH", true},

		// Invalid names
		{"with_space", "player health", false},
		{"with_operator", "health + 10", false},
		{"with_parens", "get_node(\"Player\")", false},
		{"with_dot", "player.health", false},
		{"with_slash", "node/health", false},
		{"starts_with_number", "123var", false},
		{"with_special_char", "health!", false},
		{"with_equals", "health = 100", false},
		{"with_semicolon", "var; evil()", false},
		{"with_quote", "player\"name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidVariableName(tt.varName)
			if result != tt.expected {
				t.Errorf("isValidVariableName(%q) = %v, expected %v", tt.varName, result, tt.expected)
			}
		})
	}
}

func TestFormatValueForGDScript(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		// Strings
		{"simple_string", "hello", `"hello"`},
		{"string_with_quotes", `say "hi"`, `"say \"hi\""`},
		{"string_with_backslash", `path\to\file`, `"path\\to\\file"`},

		// Numbers
		{"integer", 100, "100"},
		{"float", 3.14, "3.14"},
		{"negative", -42, "-42"},

		// Booleans
		{"bool_true", true, "true"},
		{"bool_false", false, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValueForGDScript(tt.value)
			if result != tt.expected {
				t.Errorf("formatValueForGDScript(%v) = %q, expected %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestEscapeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no_special_chars", "hello", "hello"},
		{"with_quote", `say "hi"`, `say \"hi\"`},
		{"with_backslash", `path\to\file`, `path\\to\\file`},
		{"with_both", `path\to\"file\"`, `path\\to\\\"file\\\"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeString(tt.input)
			if result != tt.expected {
				t.Errorf("escapeString(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
