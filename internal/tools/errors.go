package tools

import (
	"fmt"
	"strings"
)

// FormatError creates a structured error message following the Problem + Context + Solution pattern.
//
// problem: A clear statement of what went wrong (e.g., "Failed to connect to Godot DAP server")
// context: Relevant details (e.g., "localhost:6006") or ""
// solutions: A list of actionable steps the user can take
// cause: The underlying error (optional, can be nil)
func FormatError(problem string, context string, solutions []string, cause error) error {
	var sb strings.Builder

	// 1. Problem
	sb.WriteString(problem)
	if context != "" {
		sb.WriteString(fmt.Sprintf(" (%s)", context))
	}
	sb.WriteString("\n\n")

	// 2. Context (Optional - if detailed context is needed, it can be part of problem or a separate arg)
	// For now, we assume 'context' arg matches the simple "host:port" style.
	// If we wanted a detailed "Context:" section, we could add it.

	// 3. Possible Causes (inferred from typical scenarios, but passed as solutions usually imply causes)
	// Actually, let's stick to the requested pattern:
	// "Possible causes:" section?
	// The existing example in connect.go uses "Possible causes" and "Solutions".

	// Let's make it flexible. If solutions are provided:
	if len(solutions) > 0 {
		sb.WriteString("Suggestions:\n")
		for i, sol := range solutions {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, sol))
		}
		sb.WriteString("\n")
	}

	// 4. Underlying Error
	if cause != nil {
		sb.WriteString(fmt.Sprintf("Error details: %v", cause))
	}

	return fmt.Errorf("%s", sb.String())
}

// Common error templates

func ErrNotConnected() error {
	return FormatError(
		"Not connected to Godot DAP server",
		"",
		[]string{
			"Call godot_connect() to establish a connection",
			"Ensure Godot editor is running",
		},
		nil,
	)
}

func ErrSessionNotReady(state string) error {
	return FormatError(
		"DAP Session is not ready for this operation",
		fmt.Sprintf("Current state: %s", state),
		[]string{
			"If state is 'initialized', call godot_launch_main_scene() (or other launch tool) to start the game",
			"If state is 'disconnected', call godot_connect()",
		},
		nil,
	)
}
