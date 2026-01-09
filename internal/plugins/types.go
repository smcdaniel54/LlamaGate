package plugins

import (
	"context"
	"time"
)

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	// Metadata returns information about the plugin
	Metadata() PluginMetadata

	// ValidateInput validates the input parameters before execution
	ValidateInput(input map[string]interface{}) error

	// Execute runs the plugin workflow and returns results
	Execute(ctx context.Context, input map[string]interface{}) (*PluginResult, error)
}

// PluginMetadata contains information about a plugin
type PluginMetadata struct {
	// Name is the unique identifier for the plugin
	Name string `json:"name"`

	// Version is the plugin version
	Version string `json:"version"`

	// Description describes what the plugin does
	Description string `json:"description"`

	// Author is the plugin author
	Author string `json:"author,omitempty"`

	// InputSchema defines the expected input parameters (JSON Schema)
	InputSchema map[string]interface{} `json:"input_schema"`

	// OutputSchema defines the expected output structure (JSON Schema)
	OutputSchema map[string]interface{} `json:"output_schema"`

	// RequiredInputs lists required input parameter names
	RequiredInputs []string `json:"required_inputs"`

	// OptionalInputs lists optional input parameter names with defaults
	OptionalInputs map[string]interface{} `json:"optional_inputs,omitempty"`
}

// PluginResult represents the output of a plugin execution
type PluginResult struct {
	// Success indicates if the execution was successful
	Success bool `json:"success"`

	// Data contains the output data
	Data map[string]interface{} `json:"data,omitempty"`

	// Error contains error information if execution failed
	Error string `json:"error,omitempty"`

	// Metadata contains execution metadata
	Metadata ExecutionMetadata `json:"metadata"`
}

// ExecutionMetadata contains information about plugin execution
type ExecutionMetadata struct {
	// ExecutionTime is how long the plugin took to execute
	ExecutionTime time.Duration `json:"execution_time"`

	// StepsExecuted is the number of workflow steps executed
	StepsExecuted int `json:"steps_executed"`

	// Timestamp is when the execution completed
	Timestamp time.Time `json:"timestamp"`
}

// WorkflowStep represents a single step in an agentic workflow
type WorkflowStep struct {
	// ID is a unique identifier for this step
	ID string `json:"id"`

	// Name is a human-readable name for the step
	Name string `json:"name"`

	// Description describes what this step does
	Description string `json:"description"`

	// Type is the type of step (e.g., "llm_call", "tool_call", "data_transform", "condition")
	Type string `json:"type"`

	// Config contains step-specific configuration
	Config map[string]interface{} `json:"config"`

	// Dependencies lists step IDs that must complete before this step
	Dependencies []string `json:"dependencies,omitempty"`

	// OnError defines error handling behavior ("continue", "stop", "retry")
	OnError string `json:"on_error,omitempty"`
}

// Workflow defines an agentic workflow for a plugin
type Workflow struct {
	// ID is a unique identifier for the workflow
	ID string `json:"id"`

	// Name is a human-readable name
	Name string `json:"name"`

	// Description describes the workflow
	Description string `json:"description"`

	// Steps are the workflow steps in execution order
	Steps []WorkflowStep `json:"steps"`

	// MaxRetries is the maximum number of retries for failed steps
	MaxRetries int `json:"max_retries,omitempty"`

	// Timeout is the maximum execution time for the workflow
	Timeout time.Duration `json:"timeout,omitempty"`
}

// StepResult represents the result of executing a workflow step
type StepResult struct {
	// StepID is the ID of the step that was executed
	StepID string `json:"step_id"`

	// Success indicates if the step succeeded
	Success bool `json:"success"`

	// Output contains the step output data
	Output map[string]interface{} `json:"output,omitempty"`

	// Error contains error information if the step failed
	Error string `json:"error,omitempty"`

	// ExecutionTime is how long the step took
	ExecutionTime time.Duration `json:"execution_time"`
}
