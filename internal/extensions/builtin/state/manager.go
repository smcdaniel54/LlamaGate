package state

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/llamagate/llamagate/internal/extensions/builtin/core"
)

// Manager manages workflow state
type Manager struct {
	name     string
	version  string
	mu       sync.RWMutex
	states   map[string]*core.WorkflowState
	registry *core.Registry
}

// NewManager creates a new state manager
func NewManager(name, version string) *Manager {
	return &Manager{
		name:     name,
		version:  version,
		states:   make(map[string]*core.WorkflowState),
		registry: core.GetRegistry(),
	}
}

// Name returns the name of the extension
func (m *Manager) Name() string {
	return m.name
}

// Version returns the version of the extension
func (m *Manager) Version() string {
	return m.version
}

// Initialize initializes the manager
func (m *Manager) Initialize(ctx context.Context, config map[string]interface{}) error {
	return nil
}

// Shutdown shuts down the manager
func (m *Manager) Shutdown(ctx context.Context) error {
	return nil
}

// SaveState saves workflow state
func (m *Manager) SaveState(ctx context.Context, workflowID string, state *core.WorkflowState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}

	state.WorkflowID = workflowID
	state.UpdatedAt = time.Now()
	if state.CreatedAt.IsZero() {
		state.CreatedAt = time.Now()
	}

	m.states[workflowID] = state

	// Publish event
	if publisher := m.getEventPublisher(); publisher != nil {
		_ = publisher.Publish(ctx, &core.Event{
			Type:      "state.saved",
			Source:    m.name,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"workflow_id": workflowID,
				"status":      state.Status,
			},
		})
	}

	return nil
}

// LoadState loads workflow state
func (m *Manager) LoadState(ctx context.Context, workflowID string) (*core.WorkflowState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.states[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow state %s not found", workflowID)
	}

	return state, nil
}

// UpdateContext updates the context for a workflow
func (m *Manager) UpdateContext(ctx context.Context, workflowID string, updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.states[workflowID]
	if !exists {
		// Create new state if it doesn't exist
		state = &core.WorkflowState{
			WorkflowID: workflowID,
			Status:     "running",
			Context:    make(map[string]interface{}),
			History:    []*core.StateHistory{},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		m.states[workflowID] = state
	}

	// Update context
	if state.Context == nil {
		state.Context = make(map[string]interface{})
	}
	for k, v := range updates {
		state.Context[k] = v
	}

	state.UpdatedAt = time.Now()

	// Add to history
	state.History = append(state.History, &core.StateHistory{
		Timestamp: time.Now(),
		Action:    "context_updated",
		Data:      updates,
	})

	return nil
}

// AddHistory adds a history entry
func (m *Manager) AddHistory(ctx context.Context, workflowID, step, action string, data map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.states[workflowID]
	if !exists {
		return fmt.Errorf("workflow state %s not found", workflowID)
	}

	state.History = append(state.History, &core.StateHistory{
		Timestamp: time.Now(),
		Step:      step,
		Action:    action,
		Data:      data,
	})

	state.UpdatedAt = time.Now()
	return nil
}

// UpdateStatus updates the status of a workflow
func (m *Manager) UpdateStatus(ctx context.Context, workflowID, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.states[workflowID]
	if !exists {
		return fmt.Errorf("workflow state %s not found", workflowID)
	}

	state.Status = status
	state.UpdatedAt = time.Now()

	// Publish event
	if publisher := m.getEventPublisher(); publisher != nil {
		_ = publisher.Publish(ctx, &core.Event{
			Type:      "state.status_changed",
			Source:    m.name,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"workflow_id": workflowID,
				"status":      status,
			},
		})
	}

	return nil
}

// getEventPublisher gets the event publisher if available
func (m *Manager) getEventPublisher() core.EventPublisher {
	if m.registry == nil {
		return nil
	}
	publisher, err := m.registry.GetEventPublisher("default")
	if err != nil {
		return nil
	}
	return publisher
}
