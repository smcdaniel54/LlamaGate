// Package human provides human-in-the-loop interaction capabilities for extensions.
package human

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/llamagate/llamagate/internal/extensions/builtin/core"
)

// Interface provides human-in-the-loop interaction capabilities
type Interface struct {
	name           string
	version        string
	mu             sync.RWMutex
	pending        map[string]*pendingRequest
	registry       *core.Registry
	defaultTimeout time.Duration
}

type pendingRequest struct {
	request   *core.ApprovalRequest
	response  chan *core.ApprovalResponse
	timeout   *time.Timer
	completed bool
}

// NewInterface creates a new human interaction interface
func NewInterface(name, version string) *Interface {
	return &Interface{
		name:           name,
		version:        version,
		pending:        make(map[string]*pendingRequest),
		registry:       core.GetRegistry(),
		defaultTimeout: 24 * time.Hour, // Default 24 hour timeout
	}
}

// Name returns the name of the extension
func (h *Interface) Name() string {
	return h.name
}

// Version returns the version of the extension
func (h *Interface) Version() string {
	return h.version
}

// Initialize initializes the interface
func (h *Interface) Initialize(ctx context.Context, config map[string]interface{}) error {
	if timeout, ok := config["default_timeout"].(string); ok {
		if d, err := time.ParseDuration(timeout); err == nil {
			h.defaultTimeout = d
		}
	}
	return nil
}

// Shutdown shuts down the interface
func (h *Interface) Shutdown(_ context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Cancel all pending requests
	for _, req := range h.pending {
		if req.timeout != nil {
			req.timeout.Stop()
		}
		close(req.response)
	}

	return nil
}

// RequestApproval requests approval from a human
func (h *Interface) RequestApproval(ctx context.Context, request *core.ApprovalRequest) (*core.ApprovalResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("approval request cannot be nil")
	}

	if request.RequestID == "" {
		request.RequestID = fmt.Sprintf("approval-%d", time.Now().UnixNano())
	}

	// Set default timeout if not specified
	timeout := h.defaultTimeout
	if request.Timeout != nil {
		timeout = *request.Timeout
	}

	// Create pending request
	pending := &pendingRequest{
		request:  request,
		response: make(chan *core.ApprovalResponse, 1),
		timeout:  time.NewTimer(timeout),
	}

	h.mu.Lock()
	h.pending[request.RequestID] = pending
	h.mu.Unlock()

	// Publish event
	if publisher := h.getEventPublisher(); publisher != nil {
		_ = publisher.Publish(ctx, &core.Event{
			Type:      "human.approval.requested",
			Source:    h.name,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"request_id":  request.RequestID,
				"workflow_id": request.WorkflowID,
				"title":       request.Title,
			},
		})
	}

	// Wait for response or timeout
	select {
	case resp := <-pending.response:
		h.mu.Lock()
		pending.completed = true
		delete(h.pending, request.RequestID)
		h.mu.Unlock()
		return resp, nil
	case <-pending.timeout.C:
		h.mu.Lock()
		pending.completed = true
		delete(h.pending, request.RequestID)
		h.mu.Unlock()

		// Return timeout response
		return &core.ApprovalResponse{
			RequestID: request.RequestID,
			Approved:  false,
			Comment:   "Request timed out",
			Timestamp: time.Now(),
		}, fmt.Errorf("approval request timed out")
	case <-ctx.Done():
		h.mu.Lock()
		pending.completed = true
		delete(h.pending, request.RequestID)
		h.mu.Unlock()
		return nil, ctx.Err()
	}
}

// WaitForInput waits for human input
func (h *Interface) WaitForInput(ctx context.Context, prompt *core.InputPrompt) (*core.InputResponse, error) {
	if prompt == nil {
		return nil, fmt.Errorf("input prompt cannot be nil")
	}

	if prompt.PromptID == "" {
		prompt.PromptID = fmt.Sprintf("input-%d", time.Now().UnixNano())
	}

	// Set default timeout if not specified
	timeout := h.defaultTimeout
	if prompt.Timeout != nil {
		timeout = *prompt.Timeout
	}

	// Create pending request
	pending := &pendingRequest{
		request: &core.ApprovalRequest{
			RequestID:  prompt.PromptID,
			WorkflowID: prompt.WorkflowID,
			Title:      prompt.Prompt,
		},
		response: make(chan *core.ApprovalResponse, 1),
		timeout:  time.NewTimer(timeout),
	}

	h.mu.Lock()
	h.pending[prompt.PromptID] = pending
	h.mu.Unlock()

	// Publish event
	if publisher := h.getEventPublisher(); publisher != nil {
		_ = publisher.Publish(ctx, &core.Event{
			Type:      "human.input.requested",
			Source:    h.name,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"prompt_id":   prompt.PromptID,
				"workflow_id": prompt.WorkflowID,
			},
		})
	}

	// Wait for response or timeout
	select {
	case resp := <-pending.response:
		h.mu.Lock()
		pending.completed = true
		delete(h.pending, prompt.PromptID)
		h.mu.Unlock()

		// Convert to InputResponse
		return &core.InputResponse{
			PromptID:  prompt.PromptID,
			Value:     resp.Choice,
			Timestamp: resp.Timestamp,
		}, nil
	case <-pending.timeout.C:
		h.mu.Lock()
		pending.completed = true
		delete(h.pending, prompt.PromptID)
		h.mu.Unlock()

		if prompt.Required {
			return nil, fmt.Errorf("required input timed out")
		}
		return &core.InputResponse{
			PromptID:  prompt.PromptID,
			Value:     nil,
			Timestamp: time.Now(),
		}, nil
	case <-ctx.Done():
		h.mu.Lock()
		pending.completed = true
		delete(h.pending, prompt.PromptID)
		h.mu.Unlock()
		return nil, ctx.Err()
	}
}

// RespondApproval responds to an approval request (called by external system)
func (h *Interface) RespondApproval(ctx context.Context, requestID string, approved bool, choice, comment string) error {
	h.mu.Lock()
	pending, exists := h.pending[requestID]
	h.mu.Unlock()

	if !exists || pending.completed {
		return fmt.Errorf("approval request %s not found or already completed", requestID)
	}

	response := &core.ApprovalResponse{
		RequestID: requestID,
		Approved:  approved,
		Choice:    choice,
		Comment:   comment,
		Timestamp: time.Now(),
	}

	// Send response
	select {
	case pending.response <- response:
		// Publish event
		if publisher := h.getEventPublisher(); publisher != nil {
			_ = publisher.Publish(ctx, &core.Event{
				Type:      "human.approval.responded",
				Source:    h.name,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"request_id": requestID,
					"approved":   approved,
				},
			})
		}
		return nil
	default:
		return fmt.Errorf("failed to send approval response")
	}
}

// getEventPublisher gets the event publisher if available
func (h *Interface) getEventPublisher() core.EventPublisher {
	if h.registry == nil {
		return nil
	}
	publisher, err := h.registry.GetEventPublisher("default")
	if err != nil {
		return nil
	}
	return publisher
}
