// Package decision provides decision evaluation capabilities for extensions.
package decision

import (
	"context"
	"fmt"

	"github.com/llamagate/llamagate/internal/extensions/builtin/core"
)

// Engine provides decision evaluation capabilities
type Engine struct {
	name     string
	version  string
	registry *core.Registry
	caller   core.AgentCaller
}

// NewEngine creates a new decision engine
func NewEngine(name, version string) *Engine {
	return &Engine{
		name:     name,
		version:  version,
		registry: core.GetRegistry(),
	}
}

// Name returns the name of the extension
func (e *Engine) Name() string {
	return e.name
}

// Version returns the version of the extension
func (e *Engine) Version() string {
	return e.version
}

// Initialize initializes the engine
func (e *Engine) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Get agent caller for LLM-based decisions
	if callerName, ok := config["agent_caller"].(string); ok {
		caller, err := e.registry.GetAgentCaller(callerName)
		if err == nil {
			e.caller = caller
		}
	} else {
		// Try default caller
		caller, err := e.registry.GetAgentCaller("default")
		if err == nil {
			e.caller = caller
		}
	}
	return nil
}

// Shutdown shuts down the engine
func (e *Engine) Shutdown(ctx context.Context) error {
	return nil
}

// Evaluate evaluates a condition and returns a decision
func (e *Engine) Evaluate(ctx context.Context, condition *core.Condition) (*core.Decision, error) {
	switch condition.Type {
	case "llm":
		return e.evaluateLLM(ctx, condition)
	case "expression":
		return e.evaluateExpression(ctx, condition)
	case "custom":
		return e.evaluateCustom(ctx, condition)
	default:
		return &core.Decision{
			Result:     false,
			Confidence: 0.0,
			Reason:     fmt.Sprintf("unknown condition type: %s", condition.Type),
		}, fmt.Errorf("unknown condition type: %s", condition.Type)
	}
}

// EvaluateBranch evaluates multiple branches and selects one
func (e *Engine) EvaluateBranch(ctx context.Context, branches []*core.Branch) (*core.Branch, error) {
	if len(branches) == 0 {
		return nil, fmt.Errorf("no branches provided")
	}

	// Sort by priority (higher priority first)
	sortedBranches := make([]*core.Branch, len(branches))
	copy(sortedBranches, branches)

	// Simple sort by priority (in production, use sort.Slice)
	for i := 0; i < len(sortedBranches)-1; i++ {
		for j := i + 1; j < len(sortedBranches); j++ {
			if sortedBranches[i].Priority < sortedBranches[j].Priority {
				sortedBranches[i], sortedBranches[j] = sortedBranches[j], sortedBranches[i]
			}
		}
	}

	// Evaluate branches in priority order
	for _, branch := range sortedBranches {
		if branch.Condition == nil {
			// No condition means always match
			return branch, nil
		}

		decision, err := e.Evaluate(ctx, branch.Condition)
		if err != nil {
			continue
		}

		if decision.Result {
			return branch, nil
		}
	}

	// No branch matched, return first branch as default
	return sortedBranches[0], nil
}

// evaluateLLM evaluates a condition using an LLM
func (e *Engine) evaluateLLM(ctx context.Context, condition *core.Condition) (*core.Decision, error) {
	if e.caller == nil {
		return &core.Decision{
			Result:     false,
			Confidence: 0.0,
			Reason:     "no agent caller available",
		}, fmt.Errorf("no agent caller available")
	}

	// Build prompt for decision
	prompt := fmt.Sprintf("Evaluate the following condition and respond with only 'true' or 'false':\n\n%s", condition.Expression)
	if contextStr, ok := condition.Context["context"].(string); ok {
		prompt = fmt.Sprintf("%s\n\nContext: %s", prompt, contextStr)
	}

	req := &core.AgentRequest{
		Model: "default", // Would come from config
		Messages: []core.Message{
			{Role: "user", Content: prompt},
		},
		Temperature: func() *float64 { t := 0.0; return &t }(),
		MaxTokens:   func() *int { t := 10; return &t }(),
	}

	resp, err := e.caller.Call(ctx, req)
	if err != nil {
		return &core.Decision{
			Result:     false,
			Confidence: 0.0,
			Reason:     fmt.Sprintf("LLM evaluation failed: %v", err),
		}, err
	}

	// Parse response
	result := false
	content := resp.Content
	if content == "true" || content == "True" || content == "TRUE" {
		result = true
	}

	return &core.Decision{
		Result:     result,
		Confidence: 0.8, // Would be extracted from LLM response if available
		Reason:     content,
	}, nil
}

// evaluateExpression evaluates a condition using expression evaluation
func (e *Engine) evaluateExpression(ctx context.Context, condition *core.Condition) (*core.Decision, error) {
	// This is a placeholder - in production, use an expression evaluator
	// For now, simple string matching
	result := condition.Expression != ""

	return &core.Decision{
		Result:     result,
		Confidence: 1.0,
		Reason:     "expression evaluated",
	}, nil
}

// evaluateCustom evaluates a condition using a custom evaluator
func (e *Engine) evaluateCustom(ctx context.Context, condition *core.Condition) (*core.Decision, error) {
	// Custom evaluators would be registered and called here
	// For now, return a default decision
	return &core.Decision{
		Result:     false,
		Confidence: 0.0,
		Reason:     "custom evaluator not implemented",
	}, fmt.Errorf("custom evaluator not implemented")
}
