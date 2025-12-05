package tools

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-dap"
)

// Godot type formatters for DAP variables
// These functions detect and pretty-print common Godot types for better readability

// formatVariable enhances a DAP variable with Godot-specific formatting
func formatVariable(variable dap.Variable) map[string]interface{} {
	result := map[string]interface{}{
		"name":  variable.Name,
		"value": variable.Value,
		"type":  variable.Type,
	}

	// Detect and format Godot-specific types
	if formatted := formatGodotType(variable.Type, variable.Value); formatted != "" {
		result["formatted"] = formatted
	}

	// Mark if expandable (has children)
	if variable.VariablesReference > 0 {
		result["expandable"] = true
		result["variables_reference"] = variable.VariablesReference
	}

	// Add evaluate name if available (for setting variables later)
	if variable.EvaluateName != "" {
		result["evaluate_name"] = variable.EvaluateName
	}

	return result
}

// formatGodotType returns a formatted representation of common Godot types
// Returns empty string if no special formatting is needed
func formatGodotType(typeName string, value string) string {
	switch typeName {
	case "Vector2", "Vector2i":
		return formatVector2(value)
	case "Vector3", "Vector3i":
		return formatVector3(value)
	case "Vector4", "Vector4i":
		return formatVector4(value)
	case "Color":
		return formatColor(value)
	case "Rect2", "Rect2i":
		return formatRect2(value)
	case "AABB":
		return formatAABB(value)
	case "Plane":
		return formatPlane(value)
	case "Quaternion":
		return formatQuaternion(value)
	case "Transform2D":
		return formatTransform2D(value)
	case "Transform3D":
		return formatTransform3D(value)
	case "Basis":
		return formatBasis(value)
	case "Array":
		return formatArray(value)
	case "Dictionary":
		return formatDictionary(value)
	default:
		// Check for Node types (Node, Node2D, Node3D, Control, etc.)
		if isNodeType(typeName) {
			return formatNode(typeName, value)
		}
		return ""
	}
}

// formatVector2 formats Vector2/Vector2i as "Vector2(x=10, y=20)"
func formatVector2(value string) string {
	// Parse "(10, 20)" format
	coords := parseParenthesizedValues(value, 2)
	if len(coords) == 2 {
		return fmt.Sprintf("Vector2(x=%s, y=%s)", coords[0], coords[1])
	}
	return ""
}

// formatVector3 formats Vector3/Vector3i as "Vector3(x=1, y=2, z=3)"
func formatVector3(value string) string {
	coords := parseParenthesizedValues(value, 3)
	if len(coords) == 3 {
		return fmt.Sprintf("Vector3(x=%s, y=%s, z=%s)", coords[0], coords[1], coords[2])
	}
	return ""
}

// formatVector4 formats Vector4/Vector4i as "Vector4(x=1, y=2, z=3, w=4)"
func formatVector4(value string) string {
	coords := parseParenthesizedValues(value, 4)
	if len(coords) == 4 {
		return fmt.Sprintf("Vector4(x=%s, y=%s, z=%s, w=%s)", coords[0], coords[1], coords[2], coords[3])
	}
	return ""
}

// formatColor formats Color as "Color(r=1.0, g=0.5, b=0.0, a=1.0)"
func formatColor(value string) string {
	components := parseParenthesizedValues(value, 4)
	if len(components) == 4 {
		return fmt.Sprintf("Color(r=%s, g=%s, b=%s, a=%s)", components[0], components[1], components[2], components[3])
	}
	return ""
}

// formatRect2 formats Rect2 as "Rect2(pos=(x, y), size=(w, h))"
func formatRect2(value string) string {
	// Godot format: "[P: (x, y), S: (w, h)]"
	re := regexp.MustCompile(`\[P:\s*\(([^,]+),\s*([^)]+)\),\s*S:\s*\(([^,]+),\s*([^)]+)\)\]`)
	matches := re.FindStringSubmatch(value)
	if len(matches) == 5 {
		return fmt.Sprintf("Rect2(pos=(%s, %s), size=(%s, %s))",
			strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2]),
			strings.TrimSpace(matches[3]), strings.TrimSpace(matches[4]))
	}
	return ""
}

// formatAABB formats AABB (3D bounding box) as "AABB(pos=(x, y, z), size=(w, h, d))"
func formatAABB(value string) string {
	// Similar to Rect2 but 3D
	re := regexp.MustCompile(`\[P:\s*\(([^,]+),\s*([^,]+),\s*([^)]+)\),\s*S:\s*\(([^,]+),\s*([^,]+),\s*([^)]+)\)\]`)
	matches := re.FindStringSubmatch(value)
	if len(matches) == 7 {
		return fmt.Sprintf("AABB(pos=(%s, %s, %s), size=(%s, %s, %s))",
			strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2]), strings.TrimSpace(matches[3]),
			strings.TrimSpace(matches[4]), strings.TrimSpace(matches[5]), strings.TrimSpace(matches[6]))
	}
	return ""
}

// formatPlane formats Plane as "Plane(normal=(x, y, z), d=distance)"
func formatPlane(value string) string {
	coords := parseParenthesizedValues(value, 4)
	if len(coords) == 4 {
		return fmt.Sprintf("Plane(normal=(%s, %s, %s), d=%s)", coords[0], coords[1], coords[2], coords[3])
	}
	return ""
}

// formatQuaternion formats Quaternion as "Quat(x, y, z, w)"
func formatQuaternion(value string) string {
	coords := parseParenthesizedValues(value, 4)
	if len(coords) == 4 {
		return fmt.Sprintf("Quat(x=%s, y=%s, z=%s, w=%s)", coords[0], coords[1], coords[2], coords[3])
	}
	return ""
}

// formatTransform2D formats Transform2D with origin and rotation hint
func formatTransform2D(value string) string {
	// Godot format: "[X: (xx, xy), Y: (yx, yy), O: (ox, oy)]"
	re := regexp.MustCompile(`\[X:\s*\(([^)]+)\),\s*Y:\s*\(([^)]+)\),\s*O:\s*\(([^)]+)\)\]`)
	matches := re.FindStringSubmatch(value)
	if len(matches) == 4 {
		return fmt.Sprintf("Transform2D(x=%s, y=%s, origin=%s)",
			strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2]), strings.TrimSpace(matches[3]))
	}
	return ""
}

// formatTransform3D formats Transform3D
func formatTransform3D(value string) string {
	// Complex format - just indicate it's a transform
	if strings.Contains(value, "[X:") && strings.Contains(value, "[O:") {
		return "Transform3D(...)"
	}
	return ""
}

// formatBasis formats Basis (3x3 rotation matrix)
func formatBasis(value string) string {
	if strings.Contains(value, "[X:") && strings.Contains(value, "Y:") && strings.Contains(value, "Z:") {
		return "Basis(...)"
	}
	return ""
}

// formatArray formats Array with element count hint
func formatArray(value string) string {
	// Try to count elements
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		content := strings.TrimSpace(value[1 : len(value)-1])
		if content == "" {
			return "Array(empty)"
		}
		// Simple element count (not perfect but good enough)
		elements := strings.Split(content, ",")
		if len(elements) <= 3 {
			return fmt.Sprintf("Array(%d): %s", len(elements), value)
		}
		// Show first 3 elements
		preview := strings.Join(elements[:3], ", ")
		return fmt.Sprintf("Array(%d): [%s, ...]", len(elements), preview)
	}
	return ""
}

// formatDictionary formats Dictionary with key count hint
func formatDictionary(value string) string {
	// Try to count key-value pairs
	if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
		content := strings.TrimSpace(value[1 : len(value)-1])
		if content == "" {
			return "Dictionary(empty)"
		}
		// Count colons as rough key count
		keyCount := strings.Count(content, ":")
		if keyCount == 0 {
			return "Dictionary(...)"
		}
		if len(content) <= 50 {
			return fmt.Sprintf("Dictionary(%d): %s", keyCount, value)
		}
		return fmt.Sprintf("Dictionary(%d keys)", keyCount)
	}
	return ""
}

// formatNode formats Node objects with class name
func formatNode(typeName string, value string) string {
	// Godot shows class name for nodes
	// Value often looks like: "<Node#123>" or "Node:<CharacterBody2D#456>"
	if value == "<null>" || value == "null" {
		return fmt.Sprintf("%s(null)", typeName)
	}

	// Extract instance ID if present
	re := regexp.MustCompile(`<([^#]+)#(\d+)>`)
	matches := re.FindStringSubmatch(value)
	if len(matches) == 3 {
		className := matches[1]
		instanceID := matches[2]
		return fmt.Sprintf("%s (ID:%s)", className, instanceID)
	}

	// Fallback: show type and value
	return fmt.Sprintf("%s: %s", typeName, value)
}

// isNodeType checks if a type is a Godot Node or inherits from Node
func isNodeType(typeName string) bool {
	// Common Node types
	nodeTypes := []string{
		"Node", "Node2D", "Node3D",
		"Control", "CanvasItem", "Spatial",
		"Sprite2D", "Sprite3D",
		"CharacterBody2D", "CharacterBody3D",
		"RigidBody2D", "RigidBody3D",
		"StaticBody2D", "StaticBody3D",
		"Area2D", "Area3D",
		"Camera2D", "Camera3D",
		"Label", "Button", "Panel",
		"CollisionShape2D", "CollisionShape3D",
	}

	for _, nodeType := range nodeTypes {
		if typeName == nodeType {
			return true
		}
	}

	// If it contains "Node", "Body", "Area", or "Control", it's likely a Node
	return strings.Contains(typeName, "Node") ||
		strings.Contains(typeName, "Body") ||
		strings.Contains(typeName, "Area") ||
		strings.Contains(typeName, "Control")
}

// parseParenthesizedValues extracts comma-separated values from "(val1, val2, ...)"
func parseParenthesizedValues(value string, expectedCount int) []string {
	// Remove parentheses
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmed, "(") || !strings.HasSuffix(trimmed, ")") {
		return nil
	}

	content := trimmed[1 : len(trimmed)-1]
	parts := strings.Split(content, ",")

	if len(parts) != expectedCount {
		return nil
	}

	// Trim whitespace from each part
	result := make([]string, expectedCount)
	for i, part := range parts {
		result[i] = strings.TrimSpace(part)
	}

	return result
}

// formatVariableList formats a list of DAP variables with Godot-specific formatting
func formatVariableList(variables []dap.Variable) []map[string]interface{} {
	result := make([]map[string]interface{}, len(variables))
	for i, variable := range variables {
		result[i] = formatVariable(variable)
	}
	return result
}
