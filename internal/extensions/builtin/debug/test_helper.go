package debug

import (
	"context"
	"fmt"
	"time"

	"github.com/llamagate/llamagate/internal/extensions/builtin/core"
)

// TestHelper provides utilities for testing and debugging extensions
type TestHelper struct {
	registry   *core.Registry
	logger     *Logger
	visualizer *Visualizer
}

// NewTestHelper creates a new test helper
func NewTestHelper() *TestHelper {
	return &TestHelper{
		registry: core.GetRegistry(),
	}
}

// SetupDebugging sets up logging and visualization for testing
func (th *TestHelper) SetupDebugging(ctx context.Context, _ map[string]interface{}) error {
	// Setup logger
	th.logger = NewLogger("test-logger", "1.0.0")
	if err := th.registry.Register(ctx, th.logger, map[string]interface{}{
		"level":      "debug",
		"structured": false,
		"colors":     true,
	}); err != nil {
		return fmt.Errorf("failed to register logger: %w", err)
	}

	// Setup visualizer
	th.visualizer = NewVisualizer("test-visualizer", "1.0.0")
	if err := th.registry.Register(ctx, th.visualizer, map[string]interface{}{
		"enabled": true,
	}); err != nil {
		return fmt.Errorf("failed to register visualizer: %w", err)
	}

	return nil
}

// TeardownDebugging tears down debugging setup
func (th *TestHelper) TeardownDebugging(ctx context.Context) error {
	if th.logger != nil {
		_ = th.registry.Unregister(ctx, th.logger.Name()) // Ignore errors during test teardown
	}
	if th.visualizer != nil {
		_ = th.registry.Unregister(ctx, th.visualizer.Name()) // Ignore errors during test teardown
	}
	return nil
}

// TraceWorkflow traces a workflow execution
func (th *TestHelper) TraceWorkflow(ctx context.Context, workflowID string, fn func() error) error {
	start := time.Now()

	// Publish workflow started event
	publisher, _ := th.registry.GetEventPublisher("default")
	if publisher != nil {
		_ = publisher.Publish(ctx, &core.Event{ // Ignore publish errors in test helpers
			Type:      "workflow.started",
			Source:    "test-helper",
			Timestamp: start,
			Data: map[string]interface{}{
				"workflow_id": workflowID,
			},
		})
	}

	// Execute workflow
	err := fn()

	duration := time.Since(start)

	// Publish workflow completed event
	if publisher != nil {
		eventType := "workflow.completed"
		if err != nil {
			eventType = "workflow.failed"
		}
		_ = publisher.Publish(ctx, &core.Event{ // Ignore publish errors in test helpers
			Type:      eventType,
			Source:    "test-helper",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"workflow_id": workflowID,
				"duration":    duration.String(),
				"error":       err != nil,
			},
		})
	}

	return err
}

// TraceStep traces a step execution
func (th *TestHelper) TraceStep(ctx context.Context, workflowID, stepName string, fn func() error) error {
	start := time.Now()

	// Publish step started event
	publisher, _ := th.registry.GetEventPublisher("default")
	if publisher != nil {
		_ = publisher.Publish(ctx, &core.Event{ // Ignore publish errors in test helpers
			Type:      "workflow.step.started",
			Source:    "test-helper",
			Timestamp: start,
			Data: map[string]interface{}{
				"workflow_id": workflowID,
				"step":        stepName,
			},
		})
	}

	// Execute step
	err := fn()

	duration := time.Since(start)

	// Publish step completed event
	if publisher != nil {
		eventType := "workflow.step.completed"
		if err != nil {
			eventType = "workflow.step.failed"
		}
		_ = publisher.Publish(ctx, &core.Event{ // Ignore publish errors in test helpers
			Type:      eventType,
			Source:    "test-helper",
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"workflow_id": workflowID,
				"step":        stepName,
				"duration":    duration.String(),
				"error":       err != nil,
			},
		})
	}

	return err
}

// AssertState asserts workflow state
func (th *TestHelper) AssertState(ctx context.Context, workflowID, expectedStatus string) error {
	manager, err := th.registry.GetStateManager("default")
	if err != nil {
		return fmt.Errorf("state manager not available: %w", err)
	}

	state, err := manager.LoadState(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if state.Status != expectedStatus {
		return fmt.Errorf("expected status %s, got %s", expectedStatus, state.Status)
	}

	return nil
}

// DumpState dumps workflow state for debugging
func (th *TestHelper) DumpState(ctx context.Context, workflowID string) error {
	manager, err := th.registry.GetStateManager("default")
	if err != nil {
		return fmt.Errorf("state manager not available: %w", err)
	}

	state, err := manager.LoadState(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	fmt.Printf("\n=== Workflow State: %s ===\n", workflowID)
	fmt.Printf("Status: %s\n", state.Status)
	fmt.Printf("Step: %s\n", state.Step)
	fmt.Printf("Context: %+v\n", state.Context)
	fmt.Printf("History: %d entries\n", len(state.History))
	fmt.Println()

	return nil
}

// WaitForEvent waits for a specific event (for testing)
func (th *TestHelper) WaitForEvent(ctx context.Context, eventType string, timeout time.Duration) (*core.Event, error) {
	publisher, err := th.registry.GetEventPublisher("default")
	if err != nil {
		return nil, fmt.Errorf("event publisher not available: %w", err)
	}

	eventCh := make(chan *core.Event, 1)
	errorCh := make(chan error, 1)

	subscription, err := publisher.Subscribe(ctx, &core.EventFilter{
		Types: []string{eventType},
	}, func(ctx context.Context, event *core.Event) error {
		select {
		case eventCh <- event:
		default:
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = subscription.Unsubscribe(ctx) // Ignore unsubscribe errors in test helpers
	}()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case event := <-eventCh:
		return event, nil
	case err := <-errorCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
