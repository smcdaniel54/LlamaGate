// Package plugins provides plugin system functionality
package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// PluginContext provides plugins with access to LlamaGate services
type PluginContext struct {
	// LLMHandler is a function that plugins can use to make LLM calls
	LLMHandler LLMHandlerFunc

	// Logger is a plugin-specific logger instance
	Logger zerolog.Logger

	// Config is plugin-specific configuration
	Config map[string]interface{}

	// HTTPClient is an HTTP client for making external requests
	HTTPClient *http.Client

	// PluginName is the name of the plugin (for logging context)
	PluginName string
}

// LLMHandlerFunc is a function type for making LLM calls
// Returns the LLM response as a map, or an error
type LLMHandlerFunc func(ctx context.Context, model string, messages []map[string]interface{}, options map[string]interface{}) (map[string]interface{}, error)

// NewPluginContext creates a new plugin context
func NewPluginContext(llmHandler LLMHandlerFunc, logger zerolog.Logger, config map[string]interface{}) *PluginContext {
	return &PluginContext{
		LLMHandler: llmHandler,
		Logger:     logger,
		Config:     config,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		PluginName: "",
	}
}

// NewPluginContextWithName creates a new plugin context with plugin name
func NewPluginContextWithName(pluginName string, llmHandler LLMHandlerFunc, logger zerolog.Logger, config map[string]interface{}) *PluginContext {
	// Add plugin name to logger context
	logger = logger.With().Str("plugin", pluginName).Logger()

	return &PluginContext{
		LLMHandler: llmHandler,
		Logger:     logger,
		Config:     config,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		PluginName: pluginName,
	}
}

// CallLLM is a convenience method for plugins to make LLM calls
func (ctx *PluginContext) CallLLM(pluginCtx context.Context, model string, messages []map[string]interface{}, options map[string]interface{}) (string, error) {
	if ctx.LLMHandler == nil {
		return "", fmt.Errorf("LLM handler not available in plugin context")
	}

	result, err := ctx.LLMHandler(pluginCtx, model, messages, options)
	if err != nil {
		return "", fmt.Errorf("LLM call failed: %w", err)
	}

	// Extract the content from the response
	// Response format: {"choices": [{"message": {"content": "..."}}]}
	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					return content, nil
				}
			}
		}
	}

	// Try alternative response format
	if content, ok := result["content"].(string); ok {
		return content, nil
	}

	// If we can't extract content, return the whole result as JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal LLM response: %w", err)
	}

	return string(jsonBytes), nil
}

// GetConfig retrieves a configuration value
func (ctx *PluginContext) GetConfig(key string, defaultValue interface{}) interface{} {
	if ctx.Config == nil {
		return defaultValue
	}
	if value, ok := ctx.Config[key]; ok {
		return value
	}
	return defaultValue
}

// GetConfigString retrieves a string configuration value
func (ctx *PluginContext) GetConfigString(key string, defaultValue string) string {
	value := ctx.GetConfig(key, defaultValue)
	if str, ok := value.(string); ok {
		return str
	}
	return defaultValue
}

// GetConfigBool retrieves a boolean configuration value
func (ctx *PluginContext) GetConfigBool(key string, defaultValue bool) bool {
	value := ctx.GetConfig(key, defaultValue)
	if b, ok := value.(bool); ok {
		return b
	}
	return defaultValue
}

// GetConfigInt retrieves an integer configuration value
func (ctx *PluginContext) GetConfigInt(key string, defaultValue int) int {
	value := ctx.GetConfig(key, defaultValue)
	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case int64:
		return int(v)
	}
	return defaultValue
}

// LogInfo returns an info-level logger event with plugin context
func (ctx *PluginContext) LogInfo() *zerolog.Event {
	return ctx.Logger.Info().Str("plugin", ctx.PluginName)
}

// LogError logs an error message with plugin context
func (ctx *PluginContext) LogError(err error) *zerolog.Event {
	return ctx.Logger.Error().Str("plugin", ctx.PluginName).Err(err)
}

// LogWarn logs a warning message with plugin context
func (ctx *PluginContext) LogWarn() *zerolog.Event {
	return ctx.Logger.Warn().Str("plugin", ctx.PluginName)
}

// LogDebug logs a debug message with plugin context
func (ctx *PluginContext) LogDebug() *zerolog.Event {
	return ctx.Logger.Debug().Str("plugin", ctx.PluginName)
}
