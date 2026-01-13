package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/llamagate/llamagate/internal/plugins"
)

// ExampleWorkflowPlugin demonstrates an agentic workflow with multiple steps
// This shows how to use LLM calls, tool calls, and data transformations
type ExampleWorkflowPlugin struct {
	workflowExecutor *plugins.WorkflowExecutor
}

// Metadata returns plugin metadata
func (p *ExampleWorkflowPlugin) Metadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "workflow_example",
		Version:     "1.0.0",
		Description: "Example plugin demonstrating agentic workflows with LLM and tool integration",
		Author:      "LlamaGate Team",

		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The user query to process",
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
				"final_result": map[string]interface{}{
					"type":        "string",
					"description": "The final processed result",
				},
				"workflow_steps": map[string]interface{}{
					"type":        "array",
					"description": "Results from each workflow step",
				},
			},
		},

		RequiredInputs: []string{"query"},

		OptionalInputs: map[string]interface{}{
			"model": "llama3.2",
		},
	}
}

// ValidateInput validates the input parameters
func (p *ExampleWorkflowPlugin) ValidateInput(input map[string]interface{}) error {
	if query, exists := input["query"]; !exists {
		return fmt.Errorf("required input 'query' is missing")
	} else if queryStr, ok := query.(string); !ok || len(queryStr) == 0 {
		return fmt.Errorf("input 'query' must be a non-empty string")
	}
	return nil
}

// Execute runs the agentic workflow
func (p *ExampleWorkflowPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
	startTime := time.Now()

	// Build the workflow
	workflow := p.buildWorkflow(input)

	// Execute the workflow
	stepResults, err := p.workflowExecutor.Execute(ctx, workflow, input)
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

	// Extract final result from last step
	finalResult := ""
	if len(stepResults) > 0 {
		lastStep := stepResults[len(stepResults)-1]
		if lastStep.Success && lastStep.Output != nil {
			if result, ok := lastStep.Output["final_result"].(string); ok {
				finalResult = result
			}
		}
	}

	// Build result
	result := map[string]interface{}{
		"final_result":   finalResult,
		"workflow_steps": stepResults,
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

// buildWorkflow creates the agentic workflow definition
func (p *ExampleWorkflowPlugin) buildWorkflow(input map[string]interface{}) *plugins.Workflow {
	model := "llama3.2"
	if m, exists := input["model"]; exists {
		model, _ = m.(string)
	}

	return &plugins.Workflow{
		ID:          "example_workflow",
		Name:        "Example Agentic Workflow",
		Description: "Demonstrates multi-step workflow with LLM and tool calls",
		MaxRetries:  2,
		Timeout:     30 * time.Second,
		Steps: []plugins.WorkflowStep{
			{
				ID:          "step1_analyze",
				Name:        "Analyze Query",
				Description: "Use LLM to analyze the user query",
				Type:        "llm_call",
				Config: map[string]interface{}{
					"model":  model,
					"prompt": fmt.Sprintf("Analyze the following query and determine what action is needed: %s", input["query"]),
				},
			},
			{
				ID:          "step2_extract",
				Name:        "Extract Information",
				Description: "Extract key information from the analysis",
				Type:        "data_transform",
				Config: map[string]interface{}{
					"transform": "extract",
					"input_key": "llm_response",
					"fields":    []interface{}{"action", "parameters"},
				},
				Dependencies: []string{"step1_analyze"},
			},
			{
				ID:          "step3_execute",
				Name:        "Execute Tool",
				Description: "Execute a tool based on the extracted action",
				Type:        "tool_call",
				Config: map[string]interface{}{
					"tool_name":   "mcp.filesystem.read_file",
					"merge_state": true,
				},
				Dependencies: []string{"step2_extract"},
				OnError:      "continue", // Continue even if tool call fails
			},
			{
				ID:          "step4_synthesize",
				Name:        "Synthesize Result",
				Description: "Use LLM to synthesize the final result",
				Type:        "llm_call",
				Config: map[string]interface{}{
					"model":  model,
					"prompt": "Based on the analysis and tool results, provide a comprehensive answer.",
				},
				Dependencies: []string{"step3_execute"},
			},
		},
	}
}

// NewExampleWorkflowPlugin creates a new workflow example plugin
// Note: The workflow executor must be set separately using SetWorkflowExecutor
func NewExampleWorkflowPlugin() plugins.Plugin {
	return &ExampleWorkflowPlugin{}
}

// SetWorkflowExecutor sets the workflow executor for the plugin
func (p *ExampleWorkflowPlugin) SetWorkflowExecutor(executor *plugins.WorkflowExecutor) {
	p.workflowExecutor = executor
}
