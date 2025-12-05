package dap

import (
	"context"
	"time"
)

const (
	// Default timeouts for different operation types
	DefaultConnectTimeout = 10 * time.Second
	DefaultCommandTimeout = 30 * time.Second
	DefaultReadTimeout    = 5 * time.Second
)

// WithConnectTimeout creates a context with the default connect timeout
func WithConnectTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, DefaultConnectTimeout)
}

// WithCommandTimeout creates a context with the default command timeout
func WithCommandTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, DefaultCommandTimeout)
}

// WithReadTimeout creates a context with the default read timeout
func WithReadTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, DefaultReadTimeout)
}

// WithTimeout creates a context with a custom timeout
func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, timeout)
}
