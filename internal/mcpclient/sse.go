package mcpclient

import (
	"context"
	"fmt"
)

// SSETransport implements MCP communication over Server-Sent Events (SSE)
// This is a stub implementation for future use
type SSETransport struct {
	url     string
	headers map[string]string
	closed  bool
}

// NewSSETransport creates a new SSE transport
// Note: This is a stub implementation. Full SSE support can be added in a future version.
func NewSSETransport(url string, headers map[string]string) (*SSETransport, error) {
	return &SSETransport{
		url:     url,
		headers: headers,
		closed:  false,
	}, fmt.Errorf("SSE transport not yet implemented - use stdio transport for now")
}

// SendRequest sends a JSON-RPC request over SSE
func (t *SSETransport) SendRequest(ctx context.Context, method string, params interface{}) (*JSONRPCResponse, error) {
	return nil, fmt.Errorf("SSE transport not yet implemented")
}

// Close closes the SSE transport
func (t *SSETransport) Close() error {
	t.closed = true
	return nil
}

// IsClosed returns whether the transport is closed
func (t *SSETransport) IsClosed() bool {
	return t.closed
}

// Transport interface for abstraction
type Transport interface {
	SendRequest(ctx context.Context, method string, params interface{}) (*JSONRPCResponse, error)
	Close() error
	IsClosed() bool
}

