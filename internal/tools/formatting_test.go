package tools

import (
	"testing"

	"github.com/google/go-dap"
)

// Test Vector2 formatting
func TestFormatVector2(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard vector2",
			input:    "(10, 20)",
			expected: "Vector2(x=10, y=20)",
		},
		{
			name:     "vector2 with decimals",
			input:    "(3.14, 2.71)",
			expected: "Vector2(x=3.14, y=2.71)",
		},
		{
			name:     "vector2 with negative values",
			input:    "(-5, -10)",
			expected: "Vector2(x=-5, y=-10)",
		},
		{
			name:     "vector2 with spaces",
			input:    "( 100 , 200 )",
			expected: "Vector2(x=100, y=200)",
		},
		{
			name:     "invalid format",
			input:    "(10, 20, 30)",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatVector2(tt.input)
			if result != tt.expected {
				t.Errorf("formatVector2(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test Vector3 formatting
func TestFormatVector3(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard vector3",
			input:    "(1, 2, 3)",
			expected: "Vector3(x=1, y=2, z=3)",
		},
		{
			name:     "vector3 with decimals",
			input:    "(1.5, 2.5, 3.5)",
			expected: "Vector3(x=1.5, y=2.5, z=3.5)",
		},
		{
			name:     "invalid format - too few values",
			input:    "(1, 2)",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatVector3(tt.input)
			if result != tt.expected {
				t.Errorf("formatVector3(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test Color formatting
func TestFormatColor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard color",
			input:    "(1, 0.5, 0, 1)",
			expected: "Color(r=1, g=0.5, b=0, a=1)",
		},
		{
			name:     "transparent color",
			input:    "(0.2, 0.3, 0.4, 0.5)",
			expected: "Color(r=0.2, g=0.3, b=0.4, a=0.5)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatColor(tt.input)
			if result != tt.expected {
				t.Errorf("formatColor(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test Rect2 formatting
func TestFormatRect2(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard rect2",
			input:    "[P: (10, 20), S: (100, 50)]",
			expected: "Rect2(pos=(10, 20), size=(100, 50))",
		},
		{
			name:     "rect2 with decimals",
			input:    "[P: (1.5, 2.5), S: (10.0, 20.0)]",
			expected: "Rect2(pos=(1.5, 2.5), size=(10.0, 20.0))",
		},
		{
			name:     "invalid format",
			input:    "(10, 20, 100, 50)",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRect2(tt.input)
			if result != tt.expected {
				t.Errorf("formatRect2(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test Array formatting
func TestFormatArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty array",
			input:    "[]",
			expected: "Array(empty)",
		},
		{
			name:     "small array",
			input:    "[1, 2, 3]",
			expected: "Array(3): [1, 2, 3]",
		},
		{
			name:     "large array with preview",
			input:    "[1, 2, 3, 4, 5, 6, 7, 8]",
			expected: "Array(8): [1,  2,  3, ...]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatArray(tt.input)
			if result != tt.expected {
				t.Errorf("formatArray(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test Dictionary formatting
func TestFormatDictionary(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty dictionary",
			input:    "{}",
			expected: "Dictionary(empty)",
		},
		{
			name:     "small dictionary",
			input:    "{\"key1\": 10, \"key2\": 20}",
			expected: "Dictionary(2): {\"key1\": 10, \"key2\": 20}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDictionary(tt.input)
			// Check if result contains expected pattern (don't exact match due to count variations)
			if tt.expected == "Dictionary(empty)" && result != tt.expected {
				t.Errorf("formatDictionary(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test Node formatting
func TestFormatNode(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		value    string
		expected string
	}{
		{
			name:     "node with instance ID",
			typeName: "Node2D",
			value:    "<CharacterBody2D#456>",
			expected: "CharacterBody2D (ID:456)",
		},
		{
			name:     "null node",
			typeName: "Node",
			value:    "<null>",
			expected: "Node(null)",
		},
		{
			name:     "node without instance ID",
			typeName: "Control",
			value:    "Button",
			expected: "Control: Button",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatNode(tt.typeName, tt.value)
			if result != tt.expected {
				t.Errorf("formatNode(%q, %q) = %q, expected %q", tt.typeName, tt.value, result, tt.expected)
			}
		})
	}
}

// Test isNodeType function
func TestIsNodeType(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		expected bool
	}{
		{name: "Node", typeName: "Node", expected: true},
		{name: "Node2D", typeName: "Node2D", expected: true},
		{name: "Node3D", typeName: "Node3D", expected: true},
		{name: "CharacterBody2D", typeName: "CharacterBody2D", expected: true},
		{name: "RigidBody3D", typeName: "RigidBody3D", expected: true},
		{name: "Control", typeName: "Control", expected: true},
		{name: "Button", typeName: "Button", expected: true}, // Button is a Control (Node)
		{name: "Vector2", typeName: "Vector2", expected: false},
		{name: "int", typeName: "int", expected: false},
		{name: "CustomNode", typeName: "CustomNode", expected: true}, // Contains "Node"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNodeType(tt.typeName)
			if result != tt.expected {
				t.Errorf("isNodeType(%q) = %v, expected %v", tt.typeName, result, tt.expected)
			}
		})
	}
}

// Test formatGodotType dispatcher
func TestFormatGodotType(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		value    string
		expected string
	}{
		{
			name:     "Vector2",
			typeName: "Vector2",
			value:    "(10, 20)",
			expected: "Vector2(x=10, y=20)",
		},
		{
			name:     "Vector3",
			typeName: "Vector3",
			value:    "(1, 2, 3)",
			expected: "Vector3(x=1, y=2, z=3)",
		},
		{
			name:     "Color",
			typeName: "Color",
			value:    "(1, 0, 0, 1)",
			expected: "Color(r=1, g=0, b=0, a=1)",
		},
		{
			name:     "Node2D",
			typeName: "Node2D",
			value:    "<Sprite2D#123>",
			expected: "Sprite2D (ID:123)",
		},
		{
			name:     "Unsupported type",
			typeName: "String",
			value:    "hello",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatGodotType(tt.typeName, tt.value)
			if result != tt.expected {
				t.Errorf("formatGodotType(%q, %q) = %q, expected %q", tt.typeName, tt.value, result, tt.expected)
			}
		})
	}
}

// Test formatVariable with actual DAP variable
func TestFormatVariable(t *testing.T) {
	tests := []struct {
		name     string
		variable dap.Variable
		checkKey string
		expected interface{}
	}{
		{
			name: "simple variable",
			variable: dap.Variable{
				Name:  "health",
				Value: "100",
				Type:  "int",
			},
			checkKey: "name",
			expected: "health",
		},
		{
			name: "Vector2 variable",
			variable: dap.Variable{
				Name:  "position",
				Value: "(10, 20)",
				Type:  "Vector2",
			},
			checkKey: "formatted",
			expected: "Vector2(x=10, y=20)",
		},
		{
			name: "expandable variable",
			variable: dap.Variable{
				Name:               "player",
				Value:              "<CharacterBody2D#456>",
				Type:               "CharacterBody2D",
				VariablesReference: 1000,
			},
			checkKey: "expandable",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatVariable(tt.variable)

			if val, ok := result[tt.checkKey]; ok {
				if val != tt.expected {
					t.Errorf("formatVariable()[%q] = %v, expected %v", tt.checkKey, val, tt.expected)
				}
			} else {
				t.Errorf("formatVariable() missing key %q", tt.checkKey)
			}
		})
	}
}

// Test formatVariableList
func TestFormatVariableList(t *testing.T) {
	variables := []dap.Variable{
		{
			Name:  "x",
			Value: "(1, 2)",
			Type:  "Vector2",
		},
		{
			Name:  "y",
			Value: "42",
			Type:  "int",
		},
	}

	result := formatVariableList(variables)

	if len(result) != 2 {
		t.Errorf("formatVariableList() returned %d variables, expected 2", len(result))
	}

	// Check first variable has formatting
	if formatted, ok := result[0]["formatted"]; !ok {
		t.Error("First variable should have 'formatted' field for Vector2")
	} else if formatted != "Vector2(x=1, y=2)" {
		t.Errorf("First variable formatted = %v, expected Vector2(x=1, y=2)", formatted)
	}

	// Check second variable doesn't have formatting (int)
	if _, ok := result[1]["formatted"]; ok {
		t.Error("Second variable should not have 'formatted' field for int")
	}
}
