package visual

import (
	"context"
	"fmt"

	"github.com/smcdaniel54/LlamaGate/internal/extensions/builtin/core"
)

// Extension is a visual output extension example
type Extension struct {
	name     string
	version  string
	enabled  bool
	registry *core.Registry
}

// NewExtension creates a new visual extension
func NewExtension() *Extension {
	return &Extension{
		name:     "visual",
		version:  "1.0.0",
		enabled:  true,
		registry: core.GetRegistry(),
	}
}

// Name returns the name of the extension
func (e *Extension) Name() string {
	return e.name
}

// Version returns the version of the extension
func (e *Extension) Version() string {
	return e.version
}

// Initialize initializes the extension
func (e *Extension) Initialize(ctx context.Context, config map[string]interface{}) error {
	if enabled, ok := config["enabled"].(bool); ok {
		e.enabled = enabled
	}

	if !e.enabled {
		return nil
	}

	// Subscribe to workflow events
	publisher, err := e.registry.GetEventPublisher("default")
	if err == nil {
		filter := &core.EventFilter{
			Types: []string{"state.saved", "state.status_changed", "agent.call.completed", "tool.execution.completed"},
		}
		_, _ = publisher.Subscribe(ctx, filter, e.handleEvent)
	}

	fmt.Printf("✨ Visual extension initialized\n")
	return nil
}

// Shutdown shuts down the extension
func (e *Extension) Shutdown(ctx context.Context) error {
	if e.enabled {
		fmt.Printf("👋 Visual extension shutting down\n")
	}
	return nil
}

// handleEvent handles events for visual output
func (e *Extension) handleEvent(ctx context.Context, event *core.Event) error {
	if !e.enabled || event == nil {
		return nil
	}

	// Format visual output based on event type
	switch event.Type {
	case "state.saved":
		if workflowID, ok := event.Data["workflow_id"].(string); ok {
			fmt.Printf("📝 Workflow state saved: %s\n", workflowID)
		}
	case "state.status_changed":
		if workflowID, ok := event.Data["workflow_id"].(string); ok {
			if status, ok := event.Data["status"].(string); ok {
				icon := getStatusIcon(status)
				fmt.Printf("%s Workflow %s status: %s\n", icon, workflowID, status)
			}
		}
	case "agent.call.completed":
		if responseID, ok := event.Data["response_id"].(string); ok {
			fmt.Printf("🤖 Agent call completed: %s\n", responseID)
		}
	case "tool.execution.completed":
		if tool, ok := event.Data["tool"].(string); ok {
			if duration, ok := event.Data["duration"].(string); ok {
				fmt.Printf("🔧 Tool %s completed in %s\n", tool, duration)
			}
		}
	}

	return nil
}

// getStatusIcon returns an icon for a status
func getStatusIcon(status string) string {
	switch status {
	case "running":
		return "▶️"
	case "paused":
		return "⏸️"
	case "completed":
		return "✅"
	case "failed":
		return "❌"
	default:
		return "⚪"
	}
}
