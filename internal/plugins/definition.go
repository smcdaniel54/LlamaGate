package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// PluginDefinition represents a plugin defined in JSON/YAML
// This allows models to define plugins declaratively
type PluginDefinition struct {
	Name        string `json:"name"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description"`
	Author      string `json:"author,omitempty"`

	InputSchema  map[string]interface{} `json:"input_schema,omitempty"`
	OutputSchema map[string]interface{} `json:"output_schema,omitempty"`

	RequiredInputs []string               `json:"required_inputs,omitempty"`
	OptionalInputs map[string]interface{} `json:"optional_inputs,omitempty"`

	Workflow *WorkflowDefinition `json:"workflow,omitempty"`
}

// WorkflowDefinition represents a workflow in JSON/YAML format
type WorkflowDefinition struct {
	ID          string                   `json:"id,omitempty"`
	Name        string                   `json:"name,omitempty"`
	Description string                   `json:"description,omitempty"`
	Steps       []WorkflowStepDefinition `json:"steps"`
	MaxRetries  int                      `json:"max_retries,omitempty"`
	Timeout     string                   `json:"timeout,omitempty"` // Duration string like "30s"
}

// WorkflowStepDefinition represents a workflow step in JSON/YAML format
type WorkflowStepDefinition struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Type         string                 `json:"type"`
	Config       map[string]interface{} `json:"config"`
	Dependencies []string               `json:"dependencies,omitempty"`
	OnError      string                 `json:"on_error,omitempty"`
}

// CreatePluginFromDefinition creates a plugin from a JSON/YAML definition
// This allows models to define plugins declaratively
func CreatePluginFromDefinition(def *PluginDefinition) (Plugin, error) {
	if def == nil {
		return nil, fmt.Errorf("plugin definition cannot be nil")
	}

	if def.Name == "" {
		return nil, fmt.Errorf("plugin name is required")
	}

	// Create a simple plugin that uses the definition
	plugin := &DefinitionBasedPlugin{
		definition: def,
	}

	return plugin, nil
}

// DefinitionBasedPlugin is a plugin created from a JSON/YAML definition
type DefinitionBasedPlugin struct {
	definition *PluginDefinition
}

// Metadata returns the plugin metadata from the definition
func (p *DefinitionBasedPlugin) Metadata() PluginMetadata {
	metadata := PluginMetadata{
		Name:        p.definition.Name,
		Version:     p.definition.Version,
		Description: p.definition.Description,
		Author:      p.definition.Author,
	}

	if p.definition.InputSchema != nil {
		metadata.InputSchema = p.definition.InputSchema
	}

	if p.definition.OutputSchema != nil {
		metadata.OutputSchema = p.definition.OutputSchema
	}

	if len(p.definition.RequiredInputs) > 0 {
		metadata.RequiredInputs = p.definition.RequiredInputs
	}

	if len(p.definition.OptionalInputs) > 0 {
		metadata.OptionalInputs = p.definition.OptionalInputs
	}

	return metadata
}

// ValidateInput validates input based on the definition
func (p *DefinitionBasedPlugin) ValidateInput(input map[string]interface{}) error {
	// Check required inputs
	for _, required := range p.definition.RequiredInputs {
		if _, exists := input[required]; !exists {
			return fmt.Errorf("required input '%s' is missing", required)
		}
	}

	// TODO: Add JSON Schema validation if InputSchema is provided
	// This would use a JSON Schema validator library

	return nil
}

// Execute executes the plugin workflow
func (p *DefinitionBasedPlugin) Execute(_ context.Context, input map[string]interface{}) (*PluginResult, error) {
	// Apply optional input defaults
	processedInput := make(map[string]interface{})
	for k, v := range input {
		processedInput[k] = v
	}
	for k, defaultValue := range p.definition.OptionalInputs {
		if _, exists := processedInput[k]; !exists {
			processedInput[k] = defaultValue
		}
	}

	// If workflow is defined, it would be executed here
	// For now, return a simple result
	return &PluginResult{
		Success: true,
		Data: map[string]interface{}{
			"message": "Plugin executed",
			"input":   processedInput,
		},
		Metadata: ExecutionMetadata{
			ExecutionTime: 0,
			StepsExecuted: 0,
			Timestamp:     time.Now(),
		},
	}, nil
}

// ParsePluginDefinition parses a plugin definition from JSON
func ParsePluginDefinition(jsonData []byte) (*PluginDefinition, error) {
	var def PluginDefinition
	if err := json.Unmarshal(jsonData, &def); err != nil {
		return nil, fmt.Errorf("failed to parse plugin definition: %w", err)
	}

	return &def, nil
}

// PluginDefinitionToJSON converts a plugin definition to JSON
func PluginDefinitionToJSON(def *PluginDefinition) ([]byte, error) {
	return json.MarshalIndent(def, "", "  ")
}
