// Package testplugins provides test plugin implementations for LlamaGate testing.
package testplugins

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/llamagate/llamagate/internal/plugins"
)

// Test plugins for each use case

// UseCase1Plugin provides environment-aware plugin configuration.
type UseCase1Plugin struct{}

// Metadata returns the plugin metadata.
func (p *UseCase1Plugin) Metadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:           "usecase1_environment_aware",
		Version:        "1.0.0",
		Description:    "Environment-aware plugin that adapts behavior based on environment",
		RequiredInputs: []string{"input"},
		OptionalInputs: map[string]interface{}{
			"environment": "development",
		},
	}
}

// ValidateInput validates the plugin input.
func (p *UseCase1Plugin) ValidateInput(input map[string]interface{}) error {
	if _, exists := input["input"]; !exists {
		return fmt.Errorf("required input 'input' is missing")
	}
	return nil
}

// Execute executes the plugin with the given input.
func (p *UseCase1Plugin) Execute(_ context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}
	if envInput, ok := input["environment"].(string); ok {
		env = envInput
	}

	var timeout time.Duration
	switch env {
	case "production":
		timeout = 30 * time.Second
	case "staging":
		timeout = 20 * time.Second
	default:
		timeout = 10 * time.Second
	}

	return &plugins.PluginResult{
		Success: true,
		Data: map[string]interface{}{
			"environment": env,
			"timeout":     timeout.String(),
			"input":       input["input"],
		},
		Metadata: plugins.ExecutionMetadata{
			ExecutionTime: 0,
			StepsExecuted: 1,
			Timestamp:     time.Now(),
		},
	}, nil
}

// UseCase2Plugin provides user-configurable workflow parameters.
type UseCase2Plugin struct {
	executor *plugins.WorkflowExecutor
}

// Metadata returns the plugin metadata.
func (p *UseCase2Plugin) Metadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:           "usecase2_user_configurable",
		Version:        "1.0.0",
		Description:    "Plugin with user-configurable workflow parameters",
		RequiredInputs: []string{"query"},
		OptionalInputs: map[string]interface{}{
			"max_depth": 3,
			"use_cache": true,
		},
	}
}

// ValidateInput validates the plugin input.
func (p *UseCase2Plugin) ValidateInput(input map[string]interface{}) error {
	if _, exists := input["query"]; !exists {
		return fmt.Errorf("required input 'query' is missing")
	}
	return nil
}

// Execute executes the plugin with the given input.
func (p *UseCase2Plugin) Execute(_ context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
	maxDepth := 3
	if md, ok := input["max_depth"].(float64); ok {
		maxDepth = int(md)
	}
	useCache := true
	if uc, ok := input["use_cache"].(bool); ok {
		useCache = uc
	}

	steps := []plugins.WorkflowStep{
		{
			ID:   "analyze",
			Type: "llm_call",
			Config: map[string]interface{}{
				"model":  "llama3.2",
				"prompt": fmt.Sprintf("Analyze: %s", input["query"]),
			},
		},
	}

	if useCache {
		steps = append(steps, plugins.WorkflowStep{
			ID:   "cache",
			Type: "data_transform",
			Config: map[string]interface{}{
				"transform": "extract",
			},
			Dependencies: []string{"analyze"},
		})
	}

	for i := 0; i < maxDepth; i++ {
		prevStep := "analyze"
		if useCache && i == 0 {
			prevStep = "cache"
		} else if i > 0 {
			prevStep = fmt.Sprintf("depth_%d", i-1)
		}
		steps = append(steps, plugins.WorkflowStep{
			ID:   fmt.Sprintf("depth_%d", i),
			Type: "llm_call",
			Config: map[string]interface{}{
				"model": "llama3.2",
			},
			Dependencies: []string{prevStep},
		})
	}

	return &plugins.PluginResult{
		Success: true,
		Data: map[string]interface{}{
			"max_depth":   maxDepth,
			"use_cache":   useCache,
			"steps_count": len(steps),
		},
		Metadata: plugins.ExecutionMetadata{
			ExecutionTime: 0,
			StepsExecuted: len(steps),
			Timestamp:     time.Now(),
		},
	}, nil
}

// SetExecutor sets the workflow executor for the plugin.
func (p *UseCase2Plugin) SetExecutor(executor *plugins.WorkflowExecutor) {
	p.executor = executor
}

// UseCase3Plugin provides configuration-driven tool selection.
type UseCase3Plugin struct {
	executor *plugins.WorkflowExecutor
}

// Metadata returns the plugin metadata.
func (p *UseCase3Plugin) Metadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:           "usecase3_tool_selection",
		Version:        "1.0.0",
		Description:    "Plugin with configuration-driven tool selection",
		RequiredInputs: []string{"action"},
		OptionalInputs: map[string]interface{}{
			"enabled_tools": []string{},
		},
	}
}

// ValidateInput validates the plugin input.
func (p *UseCase3Plugin) ValidateInput(input map[string]interface{}) error {
	if _, exists := input["action"]; !exists {
		return fmt.Errorf("required input 'action' is missing")
	}
	return nil
}

// Execute executes the plugin with the given input.
func (p *UseCase3Plugin) Execute(_ context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
	enabledTools := []string{}
	if et, ok := input["enabled_tools"].([]interface{}); ok {
		for _, tool := range et {
			if toolStr, ok := tool.(string); ok {
				enabledTools = append(enabledTools, toolStr)
			}
		}
	}

	steps := []plugins.WorkflowStep{
		{
			ID:   "analyze",
			Type: "llm_call",
			Config: map[string]interface{}{
				"model": "llama3.2",
			},
		},
	}

	for i, toolName := range enabledTools {
		steps = append(steps, plugins.WorkflowStep{
			ID:   fmt.Sprintf("tool_%d", i),
			Type: "tool_call",
			Config: map[string]interface{}{
				"tool_name": toolName,
			},
			Dependencies: []string{"analyze"},
		})
	}

	return &plugins.PluginResult{
		Success: true,
		Data: map[string]interface{}{
			"enabled_tools": enabledTools,
			"tools_count":   len(enabledTools),
			"steps_count":   len(steps),
		},
		Metadata: plugins.ExecutionMetadata{
			ExecutionTime: 0,
			StepsExecuted: len(steps),
			Timestamp:     time.Now(),
		},
	}, nil
}

// SetExecutor sets the workflow executor for the plugin.
func (p *UseCase3Plugin) SetExecutor(executor *plugins.WorkflowExecutor) {
	p.executor = executor
}

// UseCase4Plugin provides adaptive timeout configuration.
type UseCase4Plugin struct{}

// Metadata returns the plugin metadata.
func (p *UseCase4Plugin) Metadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:           "usecase4_adaptive_timeout",
		Version:        "1.0.0",
		Description:    "Plugin with adaptive timeout configuration",
		RequiredInputs: []string{"text"},
		OptionalInputs: map[string]interface{}{
			"complexity": "low",
		},
	}
}

// ValidateInput validates the plugin input.
func (p *UseCase4Plugin) ValidateInput(input map[string]interface{}) error {
	if _, exists := input["text"]; !exists {
		return fmt.Errorf("required input 'text' is missing")
	}
	return nil
}

// Execute executes the plugin with the given input.
func (p *UseCase4Plugin) Execute(_ context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
	text := input["text"].(string)
	complexity := "low"
	if c, ok := input["complexity"].(string); ok {
		complexity = c
	}

	baseTimeout := 10 * time.Second
	textLength := len(text)

	if textLength > 10000 {
		baseTimeout = 60 * time.Second
	} else if textLength > 1000 {
		baseTimeout = 30 * time.Second
	}

	if complexity == "high" {
		baseTimeout *= 2
	}

	return &plugins.PluginResult{
		Success: true,
		Data: map[string]interface{}{
			"text_length":        textLength,
			"complexity":         complexity,
			"calculated_timeout": baseTimeout.String(),
		},
		Metadata: plugins.ExecutionMetadata{
			ExecutionTime: 0,
			StepsExecuted: 1,
			Timestamp:     time.Now(),
		},
	}, nil
}

// UseCase5Plugin provides configuration file-based plugin setup.
type UseCase5Plugin struct{}

// Metadata returns the plugin metadata.
func (p *UseCase5Plugin) Metadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:           "usecase5_config_file",
		Version:        "1.0.0",
		Description:    "Plugin that can be configured via config file",
		RequiredInputs: []string{"operation"},
		OptionalInputs: map[string]interface{}{
			"config_file": "default.json",
		},
	}
}

// ValidateInput validates the plugin input.
func (p *UseCase5Plugin) ValidateInput(input map[string]interface{}) error {
	if _, exists := input["operation"]; !exists {
		return fmt.Errorf("required input 'operation' is missing")
	}
	return nil
}

// Execute executes the plugin with the given input.
func (p *UseCase5Plugin) Execute(_ context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
	configFile := "default.json"
	if cf, ok := input["config_file"].(string); ok {
		configFile = cf
	}

	return &plugins.PluginResult{
		Success: true,
		Data: map[string]interface{}{
			"operation":   input["operation"],
			"config_file": configFile,
			"loaded":      true,
		},
		Metadata: plugins.ExecutionMetadata{
			ExecutionTime: 0,
			StepsExecuted: 1,
			Timestamp:     time.Now(),
		},
	}, nil
}

// UseCase6Plugin provides runtime configuration updates.
type UseCase6Plugin struct {
	config map[string]interface{}
}

// Metadata returns the plugin metadata.
func (p *UseCase6Plugin) Metadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:           "usecase6_runtime_config",
		Version:        "1.0.0",
		Description:    "Plugin that supports runtime configuration updates",
		RequiredInputs: []string{"action"},
		OptionalInputs: map[string]interface{}{
			"config": map[string]interface{}{},
		},
	}
}

// ValidateInput validates the plugin input.
func (p *UseCase6Plugin) ValidateInput(input map[string]interface{}) error {
	if _, exists := input["action"]; !exists {
		return fmt.Errorf("required input 'action' is missing")
	}
	return nil
}

// Execute executes the plugin with the given input.
func (p *UseCase6Plugin) Execute(_ context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
	action := input["action"].(string)

	if action == "update_config" {
		if config, ok := input["config"].(map[string]interface{}); ok {
			p.config = config
			return &plugins.PluginResult{
				Success: true,
				Data: map[string]interface{}{
					"action": "config_updated",
					"config": p.config,
				},
			}, nil
		}
	}

	return &plugins.PluginResult{
		Success: true,
		Data: map[string]interface{}{
			"action": action,
			"config": p.config,
		},
		Metadata: plugins.ExecutionMetadata{
			ExecutionTime: 0,
			StepsExecuted: 1,
			Timestamp:     time.Now(),
		},
	}, nil
}

// UseCase7Plugin provides context-aware configuration.
type UseCase7Plugin struct {
	executor *plugins.WorkflowExecutor
}

// Metadata returns the plugin metadata.
func (p *UseCase7Plugin) Metadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:           "usecase7_context_aware",
		Version:        "1.0.0",
		Description:    "Plugin with context-aware configuration",
		RequiredInputs: []string{"query"},
		OptionalInputs: map[string]interface{}{},
	}
}

// ValidateInput validates the plugin input.
func (p *UseCase7Plugin) ValidateInput(input map[string]interface{}) error {
	if _, exists := input["query"]; !exists {
		return fmt.Errorf("required input 'query' is missing")
	}
	return nil
}

// Execute executes the plugin with the given input.
func (p *UseCase7Plugin) Execute(_ context.Context, _ map[string]interface{}) (*plugins.PluginResult, error) {
	steps := []plugins.WorkflowStep{
		{
			ID:   "analyze",
			Type: "llm_call",
			Config: map[string]interface{}{
				"model": "llama3.2",
			},
		},
		{
			ID:   "extract",
			Type: "data_transform",
			Config: map[string]interface{}{
				"transform": "extract",
				"input_key": "llm_response",
			},
			Dependencies: []string{"analyze"},
		},
		{
			ID:   "adapt",
			Type: "llm_call",
			Config: map[string]interface{}{
				"model": "llama3.2",
			},
			Dependencies: []string{"extract"},
		},
	}

	return &plugins.PluginResult{
		Success: true,
		Data: map[string]interface{}{
			"steps_count":  len(steps),
			"context_used": true,
		},
		Metadata: plugins.ExecutionMetadata{
			ExecutionTime: 0,
			StepsExecuted: len(steps),
			Timestamp:     time.Now(),
		},
	}, nil
}

// SetExecutor sets the workflow executor for the plugin.
func (p *UseCase7Plugin) SetExecutor(executor *plugins.WorkflowExecutor) {
	p.executor = executor
}

// UseCase8Plugin provides multi-tenant configuration.
type UseCase8Plugin struct {
	tenantConfigs map[string]map[string]interface{}
}

// Metadata returns the plugin metadata.
func (p *UseCase8Plugin) Metadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:           "usecase8_multi_tenant",
		Version:        "1.0.0",
		Description:    "Plugin with multi-tenant configuration support",
		RequiredInputs: []string{"tenant_id", "operation"},
		OptionalInputs: map[string]interface{}{},
	}
}

// ValidateInput validates the plugin input.
func (p *UseCase8Plugin) ValidateInput(input map[string]interface{}) error {
	if _, exists := input["tenant_id"]; !exists {
		return fmt.Errorf("required input 'tenant_id' is missing")
	}
	if _, exists := input["operation"]; !exists {
		return fmt.Errorf("required input 'operation' is missing")
	}
	return nil
}

// Execute executes the plugin with the given input.
func (p *UseCase8Plugin) Execute(_ context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
	tenantID := input["tenant_id"].(string)
	if p.tenantConfigs == nil {
		p.tenantConfigs = make(map[string]map[string]interface{})
	}

	tenantConfig := p.tenantConfigs[tenantID]
	if tenantConfig == nil {
		tenantConfig = map[string]interface{}{
			"timeout": "30s",
			"retries": 3,
		}
		p.tenantConfigs[tenantID] = tenantConfig
	}

	return &plugins.PluginResult{
		Success: true,
		Data: map[string]interface{}{
			"tenant_id":     tenantID,
			"operation":     input["operation"],
			"tenant_config": tenantConfig,
		},
		Metadata: plugins.ExecutionMetadata{
			ExecutionTime: 0,
			StepsExecuted: 1,
			Timestamp:     time.Now(),
		},
	}, nil
}

// CreateTestPlugins creates all test plugins for testing purposes.
func CreateTestPlugins() []plugins.Plugin {
	return []plugins.Plugin{
		&UseCase1Plugin{},
		&UseCase2Plugin{},
		&UseCase3Plugin{},
		&UseCase4Plugin{},
		&UseCase5Plugin{},
		&UseCase6Plugin{},
		&UseCase7Plugin{},
		&UseCase8Plugin{},
	}
}
