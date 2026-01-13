package plugins

import (
	"context"
	"fmt"
	"time"
)

// WorkflowExecutor executes agentic workflows
type WorkflowExecutor struct {
	// LLMCallHandler handles LLM calls in workflow steps
	LLMCallHandler func(ctx context.Context, model string, messages []map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error)

	// ToolCallHandler handles tool calls in workflow steps
	ToolCallHandler func(ctx context.Context, toolName string, arguments map[string]interface{}) (map[string]interface{}, error)
}

// NewWorkflowExecutor creates a new workflow executor
func NewWorkflowExecutor(
	llmHandler func(ctx context.Context, model string, messages []map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error),
	toolHandler func(ctx context.Context, toolName string, arguments map[string]interface{}) (map[string]interface{}, error),
) *WorkflowExecutor {
	return &WorkflowExecutor{
		LLMCallHandler:  llmHandler,
		ToolCallHandler: toolHandler,
	}
}

// Execute runs a workflow and returns the results
func (e *WorkflowExecutor) Execute(ctx context.Context, workflow *Workflow, initialInput map[string]interface{}) ([]StepResult, error) {
	if workflow == nil {
		return nil, fmt.Errorf("workflow cannot be nil")
	}

	// Create context with timeout if specified
	if workflow.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, workflow.Timeout)
		defer cancel()
	}

	// Track step results
	stepResults := make(map[string]*StepResult)
	results := make([]StepResult, 0, len(workflow.Steps))

	// Track execution state
	state := make(map[string]interface{})
	for k, v := range initialInput {
		state[k] = v
	}

	// Execute steps in order, respecting dependencies
	for _, step := range workflow.Steps {
		// Check dependencies
		if err := e.checkDependencies(step, stepResults); err != nil {
			return results, fmt.Errorf("step %s dependency check failed: %w", step.ID, err)
		}

		// Execute step
		result := e.executeStep(ctx, step, state, workflow.MaxRetries)
		stepResults[step.ID] = result
		results = append(results, *result)

		// Update state with step output
		if result.Success && result.Output != nil {
			for k, v := range result.Output {
				state[k] = v
			}
		}

		// Handle step errors based on OnError policy
		if !result.Success {
			switch step.OnError {
			case "stop":
				return results, fmt.Errorf("step %s failed: %s", step.ID, result.Error)
			case "continue":
				// Continue to next step
				continue
			case "retry":
				// Already handled in executeStep with MaxRetries
				if !result.Success {
					return results, fmt.Errorf("step %s failed after retries: %s", step.ID, result.Error)
				}
			default:
				// Default to stop
				return results, fmt.Errorf("step %s failed: %s", step.ID, result.Error)
			}
		}
	}

	return results, nil
}

// checkDependencies verifies that all step dependencies have completed successfully
func (e *WorkflowExecutor) checkDependencies(step WorkflowStep, stepResults map[string]*StepResult) error {
	for _, depID := range step.Dependencies {
		depResult, exists := stepResults[depID]
		if !exists {
			return fmt.Errorf("dependency %s not found", depID)
		}
		if !depResult.Success {
			return fmt.Errorf("dependency %s failed", depID)
		}
	}
	return nil
}

// executeStep executes a single workflow step
func (e *WorkflowExecutor) executeStep(ctx context.Context, step WorkflowStep, state map[string]interface{}, maxRetries int) *StepResult {
	startTime := time.Now()
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		var output map[string]interface{}
		var err error

		switch step.Type {
		case "llm_call":
			output, err = e.executeLLMCall(ctx, step, state)
		case "tool_call":
			output, err = e.executeToolCall(ctx, step, state)
		case "data_transform":
			output, err = e.executeDataTransform(step, state)
		case "condition":
			output, err = e.executeCondition(step, state)
		default:
			err = fmt.Errorf("unknown step type: %s", step.Type)
		}

		if err == nil {
			return &StepResult{
				StepID:        step.ID,
				Success:       true,
				Output:        output,
				ExecutionTime: time.Since(startTime),
			}
		}

		lastErr = err
		if attempt < maxRetries {
			// Wait before retry (exponential backoff)
			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
		}
	}

	return &StepResult{
		StepID:        step.ID,
		Success:       false,
		Error:         lastErr.Error(),
		ExecutionTime: time.Since(startTime),
	}
}

// executeLLMCall executes an LLM call step
func (e *WorkflowExecutor) executeLLMCall(ctx context.Context, step WorkflowStep, state map[string]interface{}) (map[string]interface{}, error) {
	if e.LLMCallHandler == nil {
		return nil, fmt.Errorf("LLM call handler not configured")
	}

	// Extract model from config or state
	model, ok := step.Config["model"].(string)
	if !ok {
		model = "llama3.2" // Default model
	}

	// Build messages from config or state
	messages, ok := step.Config["messages"].([]map[string]interface{})
	if !ok {
		// Try to get from state
		if stateMessages, exists := state["messages"]; exists {
			messages, _ = stateMessages.([]map[string]interface{})
		}
	}

	if messages == nil {
		messages = []map[string]interface{}{
			{
				"role":    "user",
				"content": step.Config["prompt"].(string),
			},
		}
	}

	// Execute LLM call
	result, err := e.LLMCallHandler(ctx, model, messages, step.Config)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	return map[string]interface{}{
		"llm_response": result,
	}, nil
}

// executeToolCall executes a tool call step
func (e *WorkflowExecutor) executeToolCall(ctx context.Context, step WorkflowStep, state map[string]interface{}) (map[string]interface{}, error) {
	if e.ToolCallHandler == nil {
		return nil, fmt.Errorf("tool call handler not configured")
	}

	// Extract tool name from config
	toolName, ok := step.Config["tool_name"].(string)
	if !ok {
		return nil, fmt.Errorf("tool_name not specified in step config")
	}

	// Extract arguments from config or state
	arguments, ok := step.Config["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	// Merge with state if specified
	if mergeState, ok := step.Config["merge_state"].(bool); ok && mergeState {
		for k, v := range state {
			if _, exists := arguments[k]; !exists {
				arguments[k] = v
			}
		}
	}

	// Execute tool call
	result, err := e.ToolCallHandler(ctx, toolName, arguments)
	if err != nil {
		return nil, fmt.Errorf("tool call failed: %w", err)
	}

	return map[string]interface{}{
		"tool_result": result,
	}, nil
}

// executeDataTransform executes a data transformation step
func (e *WorkflowExecutor) executeDataTransform(step WorkflowStep, state map[string]interface{}) (map[string]interface{}, error) {
	// Extract transformation function from config
	transform, ok := step.Config["transform"].(string)
	if !ok {
		return nil, fmt.Errorf("transform function not specified")
	}

	// Get input data from state
	inputKey, ok := step.Config["input_key"].(string)
	if !ok {
		inputKey = "input"
	}

	inputData, exists := state[inputKey]
	if !exists {
		return nil, fmt.Errorf("input data not found in state: %s", inputKey)
	}

	// Apply transformation based on type
	switch transform {
	case "extract":
		// Extract specific fields
		fields, ok := step.Config["fields"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("fields not specified for extract transform")
		}
		result := make(map[string]interface{})
		inputMap, ok := inputData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("input data is not a map")
		}
		for _, field := range fields {
			fieldStr := field.(string)
			if val, exists := inputMap[fieldStr]; exists {
				result[fieldStr] = val
			}
		}
		return result, nil

	case "format":
		// Format data using template
		template, ok := step.Config["template"].(string)
		if !ok {
			return nil, fmt.Errorf("template not specified for format transform")
		}
		// Simple template replacement (can be enhanced)
		return map[string]interface{}{
			"formatted": template, // Placeholder - would need actual templating
		}, nil

	default:
		return nil, fmt.Errorf("unknown transform type: %s", transform)
	}
}

// executeCondition executes a conditional step
func (e *WorkflowExecutor) executeCondition(step WorkflowStep, state map[string]interface{}) (map[string]interface{}, error) {
	// Extract condition from config
	condition, ok := step.Config["condition"].(string)
	if !ok {
		return nil, fmt.Errorf("condition not specified")
	}

	// Evaluate condition (simplified - would need actual expression evaluator)
	// For now, check if a state key exists and is truthy
	value, exists := state[condition]
	if !exists {
		return map[string]interface{}{
			"condition_met": false,
		}, nil
	}

	// Check if value is truthy
	conditionMet := false
	switch v := value.(type) {
	case bool:
		conditionMet = v
	case string:
		conditionMet = v != ""
	case int, int64, float64:
		conditionMet = v != 0
	default:
		conditionMet = v != nil
	}

	return map[string]interface{}{
		"condition_met": conditionMet,
	}, nil
}
