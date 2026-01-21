package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/smcdaniel54/LlamaGate/internal/extensions/builtin/core"
)

// Framework provides validation capabilities
type Framework struct {
	name     string
	version  string
	registry *core.Registry
	caller   core.AgentCaller
}

// NewFramework creates a new validation framework
func NewFramework(name, version string) *Framework {
	return &Framework{
		name:     name,
		version:  version,
		registry: core.GetRegistry(),
	}
}

// Name returns the name of the extension
func (f *Framework) Name() string {
	return f.name
}

// Version returns the version of the extension
func (f *Framework) Version() string {
	return f.version
}

// Initialize initializes the framework
func (f *Framework) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Get agent caller for LLM-based validation
	if callerName, ok := config["agent_caller"].(string); ok {
		caller, err := f.registry.GetAgentCaller(callerName)
		if err == nil {
			f.caller = caller
		}
	} else {
		// Try default caller
		caller, err := f.registry.GetAgentCaller("default")
		if err == nil {
			f.caller = caller
		}
	}
	return nil
}

// Shutdown shuts down the framework
func (f *Framework) Shutdown(ctx context.Context) error {
	return nil
}

// Validate validates data against rules
func (f *Framework) Validate(ctx context.Context, data interface{}, rules *core.ValidationRules) (*core.ValidationResult, error) {
	if rules == nil {
		return &core.ValidationResult{
			Valid: true,
		}, nil
	}

	switch rules.Type {
	case "schema":
		return f.validateSchema(ctx, data, rules)
	case "llm":
		return f.validateLLM(ctx, data, rules)
	case "custom":
		return f.validateCustom(ctx, data, rules)
	default:
		return &core.ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("unknown validation type: %s", rules.Type)},
		}, fmt.Errorf("unknown validation type: %s", rules.Type)
	}
}

// validateSchema validates data against a schema
func (f *Framework) validateSchema(ctx context.Context, data interface{}, rules *core.ValidationRules) (*core.ValidationResult, error) {
	// This is a placeholder - in production, use a proper schema validator like JSON Schema
	result := &core.ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
		Score:    1.0,
	}

	// Basic validation - check if data is not nil
	if data == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "data cannot be nil")
		result.Score = 0.0
	}

	return result, nil
}

// validateLLM validates data using an LLM
func (f *Framework) validateLLM(ctx context.Context, data interface{}, rules *core.ValidationRules) (*core.ValidationResult, error) {
	if f.caller == nil {
		return &core.ValidationResult{
			Valid:  false,
			Errors: []string{"no agent caller available"},
		}, fmt.Errorf("no agent caller available")
	}

	// Build validation prompt
	prompt := "Validate the following data and respond with JSON: {\"valid\": true/false, \"errors\": [], \"warnings\": [], \"score\": 0.0-1.0}\n\n"
	prompt += fmt.Sprintf("Data: %v\n", data)

	if len(rules.Rules) > 0 {
		prompt += "\nRules:\n"
		for _, rule := range rules.Rules {
			prompt += fmt.Sprintf("- %s\n", rule)
		}
	}

	req := &core.AgentRequest{
		Model: "default",
		Messages: []core.Message{
			{Role: "user", Content: prompt},
		},
		Temperature: func() *float64 { t := 0.0; return &t }(),
	}

	resp, err := f.caller.Call(ctx, req)
	if err != nil {
		return &core.ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("LLM validation failed: %v", err)},
		}, err
	}

	// Parse response (simplified - in production, parse JSON)
	result := &core.ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
		Score:    0.8,
	}

	// Simple heuristic: if response contains "invalid" or "error", mark as invalid
	if strings.Contains(resp.Content, "invalid") || strings.Contains(resp.Content, "error") {
		result.Valid = false
		result.Errors = append(result.Errors, resp.Content)
		result.Score = 0.3
	}

	return result, nil
}

// validateCustom validates data using custom validators
func (f *Framework) validateCustom(ctx context.Context, data interface{}, rules *core.ValidationRules) (*core.ValidationResult, error) {
	// Custom validators would be registered and called here
	return &core.ValidationResult{
		Valid:  false,
		Errors: []string{"custom validator not implemented"},
	}, fmt.Errorf("custom validator not implemented")
}
