// Package agent provides agent/LLM calling capabilities for extensions.
package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/llamagate/llamagate/internal/extensions/builtin/core"
)

// DefaultCaller is the default implementation of AgentCaller
type DefaultCaller struct {
	name     string
	version  string
	baseURL  string
	apiKey   string
	timeout  time.Duration
	registry *core.Registry
}

// NewDefaultCaller creates a new default agent caller
func NewDefaultCaller(name, version, baseURL, apiKey string) *DefaultCaller {
	return &DefaultCaller{
		name:     name,
		version:  version,
		baseURL:  baseURL,
		apiKey:   apiKey,
		timeout:  30 * time.Second,
		registry: core.GetRegistry(),
	}
}

// Name returns the name of the extension
func (c *DefaultCaller) Name() string {
	return c.name
}

// Version returns the version of the extension
func (c *DefaultCaller) Version() string {
	return c.version
}

// Initialize initializes the caller
func (c *DefaultCaller) Initialize(_ context.Context, config map[string]interface{}) error {
	if baseURL, ok := config["base_url"].(string); ok {
		c.baseURL = baseURL
	}
	if apiKey, ok := config["api_key"].(string); ok {
		c.apiKey = apiKey
	}
	if timeout, ok := config["timeout"].(string); ok {
		if d, err := time.ParseDuration(timeout); err == nil {
			c.timeout = d
		}
	}
	return nil
}

// Shutdown shuts down the caller
func (c *DefaultCaller) Shutdown(_ context.Context) error {
	return nil
}

// Call makes a synchronous agent call
func (c *DefaultCaller) Call(ctx context.Context, req *core.AgentRequest) (*core.AgentResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Publish event before call
	if publisher := c.getEventPublisher(); publisher != nil {
		_ = publisher.Publish(ctx, &core.Event{
			Type:      "agent.call.started",
			Source:    c.name,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"model":    req.Model,
				"messages": len(req.Messages),
			},
		})
	}

	// Make the actual API call (this would integrate with LlamaGate's proxy)
	// For now, this is a placeholder that would call the underlying Ollama API
	response, err := c.makeAPICall(ctx, req)
	if err != nil {
		// Publish error event
		if publisher := c.getEventPublisher(); publisher != nil {
			_ = publisher.Publish(ctx, &core.Event{
				Type:      "agent.call.failed",
				Source:    c.name,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"error": err.Error(),
				},
			})
		}
		return nil, fmt.Errorf("agent call failed: %w", err)
	}

	// Publish success event
	if publisher := c.getEventPublisher(); publisher != nil {
		_ = publisher.Publish(ctx, &core.Event{
			Type:      "agent.call.completed",
			Source:    c.name,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"response_id": response.ID,
			},
		})
	}

	return response, nil
}

// CallStream makes a streaming agent call
func (c *DefaultCaller) CallStream(ctx context.Context, req *core.AgentRequest) (<-chan *core.StreamChunk, error) {
	// Create timeout context that will be managed by the goroutine
	streamCtx, cancel := context.WithTimeout(ctx, c.timeout)

	ch := make(chan *core.StreamChunk, 10)

	go func() {
		// Ensure cleanup happens when goroutine exits
		defer cancel()
		defer close(ch)

		// Publish event before stream
		if publisher := c.getEventPublisher(); publisher != nil {
			_ = publisher.Publish(streamCtx, &core.Event{
				Type:      "agent.stream.started",
				Source:    c.name,
				Timestamp: time.Now(),
			})
		}

		// Make streaming API call (placeholder)
		// The context will be cancelled if timeout expires or parent context is cancelled
		err := c.makeStreamingAPICall(streamCtx, req, ch)
		if err != nil {
			// Send error chunk before closing
			select {
			case ch <- &core.StreamChunk{
				Done:  true,
				Error: err,
			}:
			case <-streamCtx.Done():
				// Context cancelled, channel may be closed
			}
		}
	}()

	return ch, nil
}

// makeAPICall makes the actual HTTP API call
// This would integrate with LlamaGate's proxy layer
func (c *DefaultCaller) makeAPICall(_ context.Context, req *core.AgentRequest) (*core.AgentResponse, error) {
	// TODO: Integrate with LlamaGate's proxy/proxy.go
	// This is a placeholder implementation
	return &core.AgentResponse{
		ID:      fmt.Sprintf("call-%d", time.Now().UnixNano()),
		Model:   req.Model,
		Content: "Response placeholder",
		Usage: &core.Usage{
			TotalTokens: 100,
		},
	}, nil
}

// makeStreamingAPICall makes a streaming HTTP API call
func (c *DefaultCaller) makeStreamingAPICall(ctx context.Context, _ *core.AgentRequest, ch chan<- *core.StreamChunk) error {
	// TODO: Integrate with LlamaGate's streaming proxy
	// This is a placeholder implementation
	// Check context before sending chunks
	select {
	case <-ctx.Done():
		return ctx.Err()
	case ch <- &core.StreamChunk{
		Content: "Streaming response",
		Done:    false,
	}:
	}

	// Check context before sending final chunk
	select {
	case <-ctx.Done():
		return ctx.Err()
	case ch <- &core.StreamChunk{
		Content: "",
		Done:    true,
	}:
	}
	return nil
}

// getEventPublisher gets the event publisher if available
func (c *DefaultCaller) getEventPublisher() core.EventPublisher {
	if c.registry == nil {
		return nil
	}
	publisher, err := c.registry.GetEventPublisher("default")
	if err != nil {
		return nil
	}
	return publisher
}
