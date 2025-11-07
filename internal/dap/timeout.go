package dap

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-dap"
)

const (
	// Default timeouts for different operation types
	DefaultConnectTimeout  = 10 * time.Second
	DefaultCommandTimeout  = 30 * time.Second
	DefaultReadTimeout     = 5 * time.Second
)

// ReadWithTimeout reads a DAP message with a timeout
// This prevents hangs when the server stops responding
func (c *Client) ReadWithTimeout(ctx context.Context) (dap.Message, error) {
	type result struct {
		msg dap.Message
		err error
	}

	resultChan := make(chan result, 1)

	// Read in goroutine
	go func() {
		msg, err := c.read()
		resultChan <- result{msg, err}
	}()

	// Wait for result or timeout
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("read timeout: %w", ctx.Err())
	case res := <-resultChan:
		return res.msg, res.err
	}
}

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
