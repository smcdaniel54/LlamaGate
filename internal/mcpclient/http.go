package mcpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// getRequestIDFromContext extracts request ID from context
// Uses the same key as middleware package for consistency
func getRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

// HTTPTransport implements MCP communication over HTTP
type HTTPTransport struct {
	url        string
	headers    map[string]string
	httpClient *http.Client
	mu         sync.RWMutex
	closed     bool
}

// NewHTTPTransport creates a new HTTP transport for MCP communication
func NewHTTPTransport(url string, headers map[string]string, timeout time.Duration) (*HTTPTransport, error) {
	if url == "" {
		return nil, fmt.Errorf("URL is required for HTTP transport")
	}

	// Create HTTP client with timeout
	if timeout == 0 {
		timeout = 30 * time.Second // Default timeout
	}

	httpClient := &http.Client{
		Timeout: timeout,
	}

	transport := &HTTPTransport{
		url:        url,
		headers:    make(map[string]string),
		httpClient: httpClient,
		closed:     false,
	}

	// Copy headers
	for k, v := range headers {
		transport.headers[k] = v
	}

	// Set default Content-Type if not provided
	if _, ok := transport.headers["Content-Type"]; !ok {
		transport.headers["Content-Type"] = "application/json"
	}

	return transport, nil
}

// SendRequest sends a JSON-RPC request over HTTP and waits for a response
func (t *HTTPTransport) SendRequest(ctx context.Context, method string, params interface{}) (*JSONRPCResponse, error) {
	t.mu.RLock()
	if t.closed {
		t.mu.RUnlock()
		return nil, ErrConnectionClosed
	}
	t.mu.RUnlock()

	// Generate request ID
	requestID := time.Now().UnixNano()

	// Marshal params
	var paramsJSON json.RawMessage
	if params != nil {
		var err error
		paramsJSON, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	// Create JSON-RPC request
	req := JSONRPCRequest{
		JSONRPC: JSONRPCVersion,
		ID:      requestID,
		Method:  method,
		Params:  paramsJSON,
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", t.url, bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for k, v := range t.headers {
		httpReq.Header.Set(k, v)
	}

	// Propagate request ID from context if available
	if requestID := getRequestIDFromContext(ctx); requestID != "" {
		httpReq.Header.Set("X-Request-ID", requestID)
	}

	// Send request
	resp, err := t.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON-RPC response
	var jsonRPCResp JSONRPCResponse
	if err := json.Unmarshal(body, &jsonRPCResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON-RPC response: %w", err)
	}

	// Check for JSON-RPC error
	if jsonRPCResp.Error != nil {
		return nil, fmt.Errorf("JSON-RPC error: %w", jsonRPCResp.Error)
	}

	// Verify request ID matches
	// Convert both to float64 for comparison since JSON unmarshalling converts numbers to float64
	var receivedID float64
	switch v := jsonRPCResp.ID.(type) {
	case float64:
		receivedID = v
	case int64:
		receivedID = float64(v)
	case int:
		receivedID = float64(v)
	default:
		// If ID is not a number, log warning
		log.Warn().
			Interface("expected_id", requestID).
			Interface("received_id", jsonRPCResp.ID).
			Msg("Request ID mismatch in HTTP transport response (non-numeric ID)")
		return &jsonRPCResp, nil
	}
	
	expectedID := float64(requestID)
	if receivedID != expectedID {
		log.Warn().
			Float64("expected_id", expectedID).
			Float64("received_id", receivedID).
			Msg("Request ID mismatch in HTTP transport response")
	}

	return &jsonRPCResp, nil
}

// Close closes the HTTP transport
func (t *HTTPTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	return nil
}

// IsClosed returns whether the transport is closed
func (t *HTTPTransport) IsClosed() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.closed
}

// GetURL returns the transport URL
func (t *HTTPTransport) GetURL() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.url
}

// GetHeaders returns a copy of the transport headers
func (t *HTTPTransport) GetHeaders() map[string]string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	headers := make(map[string]string, len(t.headers))
	for k, v := range t.headers {
		headers[k] = v
	}
	return headers
}

// GetTimeout returns the HTTP client timeout
func (t *HTTPTransport) GetTimeout() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.httpClient.Timeout
}
