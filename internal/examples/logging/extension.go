package logging

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/llamagate/llamagate/internal/extensions/builtin/core"
)

// Extension is a simple logging extension example
// For production use, see internal/extensions/debug/logger.go for comprehensive logging
type Extension struct {
	name    string
	version string
	level   string
}

// NewExtension creates a new logging extension
func NewExtension() *Extension {
	return &Extension{
		name:    "logging",
		version: "1.0.0",
		level:   "info",
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
	if level, ok := config["level"].(string); ok {
		e.level = level
	}

	// Subscribe to events
	registry := core.GetRegistry()
	publisher, err := registry.GetEventPublisher("default")
	if err == nil {
		filter := &core.EventFilter{
			Types: []string{"agent.call.started", "agent.call.completed", "tool.execution.started", "tool.execution.completed"},
		}
		_, _ = publisher.Subscribe(ctx, filter, e.handleEvent)
	}

	log.Printf("[%s] Logging extension initialized with level: %s", e.name, e.level)
	return nil
}

// Shutdown shuts down the extension
func (e *Extension) Shutdown(ctx context.Context) error {
	log.Printf("[%s] Logging extension shutting down", e.name)
	return nil
}

// handleEvent handles events for logging
func (e *Extension) handleEvent(ctx context.Context, event *core.Event) error {
	if event == nil {
		return nil
	}

	// Format log message
	msg := fmt.Sprintf("[%s] Event: %s from %s at %s",
		e.name,
		event.Type,
		event.Source,
		event.Timestamp.Format(time.RFC3339),
	)

	// Add data if present
	if len(event.Data) > 0 {
		msg += fmt.Sprintf(" - Data: %v", event.Data)
	}

	log.Println(msg)
	return nil
}
