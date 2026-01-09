package examples

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/llamagate/llamagate/internal/plugins"
)

// DynamicConfigPlugin demonstrates dynamic configuration patterns
// This example shows how to adapt plugin behavior based on configuration
type DynamicConfigPlugin struct {
	executor *plugins.WorkflowExecutor
}

// Metadata returns plugin metadata
func (p *DynamicConfigPlugin) Metadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "dynamic_config_example",
		Version:     "1.0.0",
		Description: "Example plugin demonstrating dynamic configuration patterns",
		Author:      "LlamaGate Team",

		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The query to process",
				},
				"environment": map[string]interface{}{
					"type":        "string",
					"description": "Environment (development, staging, production)",
					"enum":        []string{"development", "staging", "production"},
					"default":     "development",
				},
				"max_depth": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum processing depth",
					"default":     3,
					"minimum":     1,
					"maximum":     10,
				},
				"use_cache": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether to use caching",
					"default":     true,
				},
				"model": map[string]interface{}{
					"type":        "string",
					"description": "LLM model to use",
					"default":     "llama3.2",
				},
			},
			"required": []string{"query"},
		},

		OutputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"result": map[string]interface{}{
					"type":        "string",
					"description": "The processed result",
				},
				"config_used": map[string]interface{}{
					"type":        "object",
					"description": "Configuration values used",
				},
				"execution_metadata": map[string]interface{}{
					"type":        "object",
					"description": "Execution metadata",
				},
			},
		},

		RequiredInputs: []string{"query"},

		OptionalInputs: map[string]interface{}{
			"environment": "development",
			"max_depth":   3,
			"use_cache":   true,
			"model":       "llama3.2",
		},
	}
}

// ValidateInput validates the input parameters
func (p *DynamicConfigPlugin) ValidateInput(input map[string]interface{}) error {
	if query, exists := input["query"]; !exists {
		return fmt.Errorf("required input 'query' is missing")
	} else if queryStr, ok := query.(string); !ok || len(queryStr) == 0 {
		return fmt.Errorf("input 'query' must be a non-empty string")
	}

	// Validate environment
	if env, exists := input["environment"]; exists {
		envStr, ok := env.(string)
		if !ok {
			return fmt.Errorf("environment must be a string")
		}
		validEnvs := []string{"development", "staging", "production"}
		valid := false
		for _, ve := range validEnvs {
			if ve == envStr {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("environment must be one of: %v", validEnvs)
		}
	}

	// Validate max_depth
	if maxDepth, exists := input["max_depth"]; exists {
		if maxDepthFloat, ok := maxDepth.(float64); ok {
			if maxDepthFloat < 1 || maxDepthFloat > 10 {
				return fmt.Errorf("max_depth must be between 1 and 10")
			}
		} else {
			return fmt.Errorf("max_depth must be a number")
		}
	}

	return nil
}

// Execute runs the plugin with dynamic configuration
func (p *DynamicConfigPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
	startTime := time.Now()

	// Get configuration values (with defaults)
	config := p.getConfiguration(input)

	// Build workflow dynamically based on configuration
	workflow := p.buildWorkflow(config, input)

	// Execute workflow
	stepResults, err := p.executor.Execute(ctx, workflow, input)
	if err != nil {
		return &plugins.PluginResult{
			Success: false,
			Error:   err.Error(),
			Metadata: plugins.ExecutionMetadata{
				ExecutionTime: time.Since(startTime),
				StepsExecuted: len(stepResults),
				Timestamp:     time.Now(),
			},
		}, nil
	}

	// Extract final result
	finalResult := ""
	if len(stepResults) > 0 {
		lastStep := stepResults[len(stepResults)-1]
		if lastStep.Success && lastStep.Output != nil {
			if result, ok := lastStep.Output["result"].(string); ok {
				finalResult = result
			}
		}
	}

	// Build result with configuration used
	result := map[string]interface{}{
		"result":     finalResult,
		"config_used": config,
		"execution_metadata": map[string]interface{}{
			"steps_executed": len(stepResults),
			"environment":   config["environment"],
		},
	}

	return &plugins.PluginResult{
		Success: true,
		Data:    result,
		Metadata: plugins.ExecutionMetadata{
			ExecutionTime: time.Since(startTime),
			StepsExecuted: len(stepResults),
			Timestamp:     time.Now(),
		},
	}, nil
}

// getConfiguration extracts and applies dynamic configuration
func (p *DynamicConfigPlugin) getConfiguration(input map[string]interface{}) map[string]interface{} {
	config := make(map[string]interface{})

	// Get environment (from input or environment variable)
	env := input["environment"].(string)
	if env == "" {
		env = os.Getenv("ENVIRONMENT")
		if env == "" {
			env = "development"
		}
	}
	config["environment"] = env

	// Get model
	model := "llama3.2"
	if m, exists := input["model"]; exists {
		model, _ = m.(string)
	}
	config["model"] = model

	// Get max_depth
	maxDepth := 3
	if md, exists := input["max_depth"]; exists {
		if mdFloat, ok := md.(float64); ok {
			maxDepth = int(mdFloat)
		}
	}
	config["max_depth"] = maxDepth

	// Get use_cache
	useCache := true
	if uc, exists := input["use_cache"]; exists {
		useCache, _ = uc.(bool)
	}
	config["use_cache"] = useCache

	// Calculate adaptive timeout based on environment and depth
	timeout := p.calculateTimeout(env, maxDepth)
	config["timeout"] = timeout

	// Calculate max retries based on environment
	maxRetries := p.calculateMaxRetries(env)
	config["max_retries"] = maxRetries

	return config
}

// calculateTimeout calculates timeout based on configuration
func (p *DynamicConfigPlugin) calculateTimeout(env string, maxDepth int) time.Duration {
	baseTimeout := 10 * time.Second

	// Adjust based on environment
	switch env {
	case "production":
		baseTimeout = 60 * time.Second
	case "staging":
		baseTimeout = 30 * time.Second
	default: // development
		baseTimeout = 10 * time.Second
	}

	// Adjust based on depth
	timeout := baseTimeout * time.Duration(maxDepth)

	// Cap at maximum
	maxTimeout := 5 * time.Minute
	if timeout > maxTimeout {
		timeout = maxTimeout
	}

	return timeout
}

// calculateMaxRetries calculates max retries based on environment
func (p *DynamicConfigPlugin) calculateMaxRetries(env string) int {
	switch env {
	case "production":
		return 3
	case "staging":
		return 2
	default: // development
		return 1
	}
}

// buildWorkflow builds a workflow dynamically based on configuration
func (p *DynamicConfigPlugin) buildWorkflow(config map[string]interface{}, input map[string]interface{}) *plugins.Workflow {
	steps := []plugins.WorkflowStep{}

	// Step 1: Analyze query
	steps = append(steps, plugins.WorkflowStep{
		ID:          "analyze",
		Name:        "Analyze Query",
		Description: "Analyze the user query",
		Type:        "llm_call",
		Config: map[string]interface{}{
			"model":  config["model"],
			"prompt": fmt.Sprintf("Analyze this query: %s", input["query"]),
		},
	})

	// Step 2: Check cache (if enabled)
	if config["use_cache"].(bool) {
		steps = append(steps, plugins.WorkflowStep{
			ID:          "check_cache",
			Name:        "Check Cache",
			Description: "Check if result is cached",
			Type:        "data_transform",
			Config: map[string]interface{}{
				"transform": "cache_lookup",
				"input_key": "llm_response",
			},
			Dependencies: []string{"analyze"},
			OnError:      "continue", // Continue even if cache check fails
		})
	}

	// Step 3-N: Process with depth
	maxDepth := config["max_depth"].(int)
	prevStepID := "analyze"
	if config["use_cache"].(bool) {
		prevStepID = "check_cache"
	}

	for i := 0; i < maxDepth; i++ {
		stepID := fmt.Sprintf("process_depth_%d", i)
		steps = append(steps, plugins.WorkflowStep{
			ID:          stepID,
			Name:        fmt.Sprintf("Process Depth %d", i),
			Description: fmt.Sprintf("Process at depth %d", i),
			Type:        "llm_call",
			Config: map[string]interface{}{
				"model":  config["model"],
				"prompt": fmt.Sprintf("Process the query at depth %d", i),
			},
			Dependencies: []string{prevStepID},
		})
		prevStepID = stepID
	}

	// Final step: Synthesize
	steps = append(steps, plugins.WorkflowStep{
		ID:          "synthesize",
		Name:        "Synthesize Result",
		Description: "Synthesize final result",
		Type:        "llm_call",
		Config: map[string]interface{}{
			"model":  config["model"],
			"prompt": "Synthesize a final answer based on all processing steps",
		},
		Dependencies: []string{prevStepID},
	})

	return &plugins.Workflow{
		ID:          "dynamic_config_workflow",
		Name:        "Dynamic Configuration Workflow",
		Description: "Workflow with dynamic configuration",
		Steps:       steps,
		MaxRetries:  config["max_retries"].(int),
		Timeout:     config["timeout"].(time.Duration),
	}
}

// SetWorkflowExecutor sets the workflow executor for the plugin
func (p *DynamicConfigPlugin) SetWorkflowExecutor(executor *plugins.WorkflowExecutor) {
	p.executor = executor
}

// NewDynamicConfigPlugin creates a new dynamic config example plugin
func NewDynamicConfigPlugin() plugins.Plugin {
	return &DynamicConfigPlugin{}
}
