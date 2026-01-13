// Package templates provides plugin templates and examples for LlamaGate.
package templates

import (
	"context"
	"fmt"
	"time"

	"github.com/llamagate/llamagate/internal/plugins"
)

// TemplatePlugin is a template for creating new plugins
// Copy this file and modify it to create your own plugin
type TemplatePlugin struct {
	// Add any plugin-specific fields here
}

// Metadata returns plugin metadata
func (p *TemplatePlugin) Metadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "template_plugin",                            // TODO: Change to your plugin name
		Version:     "1.0.0",                                      // TODO: Set your version
		Description: "A template plugin for creating new plugins", // TODO: Describe your plugin
		Author:      "Your Name",                                  // TODO: Set your name

		// Define your input schema (JSON Schema format)
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				// TODO: Define your required inputs
				"required_input": map[string]interface{}{
					"type":        "string",
					"description": "A required input parameter",
				},
				// TODO: Define your optional inputs
				"optional_input": map[string]interface{}{
					"type":        "string",
					"description": "An optional input parameter",
					"default":     "default_value",
				},
			},
			"required": []string{"required_input"},
		},

		// Define your output schema (JSON Schema format)
		OutputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"result": map[string]interface{}{
					"type":        "string",
					"description": "The result of the plugin execution",
				},
			},
		},

		// List required input parameter names
		RequiredInputs: []string{"required_input"},

		// Define optional inputs with default values
		OptionalInputs: map[string]interface{}{
			"optional_input": "default_value",
		},
	}
}

// ValidateInput validates the input parameters
func (p *TemplatePlugin) ValidateInput(input map[string]interface{}) error {
	metadata := p.Metadata()

	// Check required inputs
	for _, required := range metadata.RequiredInputs {
		if _, exists := input[required]; !exists {
			return fmt.Errorf("required input '%s' is missing", required)
		}
	}

	// TODO: Add custom validation logic here
	// Example: Check input types, ranges, formats, etc.

	return nil
}

// Execute runs the plugin workflow
func (p *TemplatePlugin) Execute(_ context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
	startTime := time.Now()

	// Apply default values for optional inputs
	metadata := p.Metadata()
	processedInput := make(map[string]interface{})
	for k, v := range input {
		processedInput[k] = v
	}
	for k, defaultValue := range metadata.OptionalInputs {
		if _, exists := processedInput[k]; !exists {
			processedInput[k] = defaultValue
		}
	}

	// TODO: Implement your plugin logic here
	// This is where you would:
	// 1. Process the inputs
	// 2. Execute your workflow
	// 3. Generate outputs

	// Example: Simple processing
	result := map[string]interface{}{
		"result": fmt.Sprintf("Processed: %v", processedInput),
	}

	executionTime := time.Since(startTime)

	return &plugins.PluginResult{
		Success: true,
		Data:    result,
		Metadata: plugins.ExecutionMetadata{
			ExecutionTime: executionTime,
			StepsExecuted: 1, // TODO: Update based on actual steps
			Timestamp:     time.Now(),
		},
	}, nil
}

// NewTemplatePlugin creates a new instance of the template plugin
func NewTemplatePlugin() plugins.Plugin {
	return &TemplatePlugin{}
}

// Register this plugin (call this from your main application)
func init() {
	// Uncomment to auto-register:
	// registry := plugins.NewRegistry()
	// registry.Register(NewTemplatePlugin())
}
