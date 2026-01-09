package templates

import (
	"context"
	"fmt"
	"time"

	"github.com/llamagate/llamagate/internal/plugins"
)

// SimplePlugin demonstrates the minimal code needed for a plugin
// This is the simplest possible plugin - just 3 methods!
type SimplePlugin struct{}

// Metadata - Define what your plugin does
func (p *SimplePlugin) Metadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "simple_plugin",
		Version:     "1.0.0",
		Description: "A simple example plugin",
		RequiredInputs: []string{"input"},
		OptionalInputs: map[string]interface{}{
			"option": "default",
		},
	}
}

// ValidateInput - Check inputs are valid (optional, can return nil)
func (p *SimplePlugin) ValidateInput(input map[string]interface{}) error {
	if _, exists := input["input"]; !exists {
		return fmt.Errorf("input is required")
	}
	return nil
}

// Execute - Do the work
func (p *SimplePlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
	startTime := time.Now()

	// Get inputs
	text := input["input"].(string)
	option := "default"
	if opt, ok := input["option"].(string); ok {
		option = opt
	}

	// Do your work here
	result := fmt.Sprintf("Processed: %s with option: %s", text, option)

	// Return result
	return &plugins.PluginResult{
		Success: true,
		Data: map[string]interface{}{
			"result": result,
		},
		Metadata: plugins.ExecutionMetadata{
			ExecutionTime: time.Since(startTime),
			StepsExecuted: 1,
			Timestamp:     time.Now(),
		},
	}, nil
}

// That's it! Just register it:
// registry.Register(&SimplePlugin{})
