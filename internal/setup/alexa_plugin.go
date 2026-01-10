package setup

import (
	"os"

	"github.com/llamagate/llamagate/internal/config"
	"github.com/llamagate/llamagate/internal/plugins"
	"github.com/llamagate/llamagate/internal/proxy"
	alexaplugin "github.com/llamagate/llamagate/plugins"
	"github.com/rs/zerolog/log"
)

// RegisterAlexaPlugin registers the Alexa Skill plugin with context
func RegisterAlexaPlugin(registry *plugins.Registry, proxyInstance *proxy.Proxy, llmHandler plugins.LLMHandlerFunc, cfg *config.Config) error {
	// Get plugin configuration
	var pluginConfig map[string]interface{}
	if cfg.Plugins != nil && cfg.Plugins.Configs != nil {
		if config, ok := cfg.Plugins.Configs["alexa_skill"]; ok {
			pluginConfig = config
		}
	}
	
	// If no config from file, create from environment variables
	if pluginConfig == nil {
		pluginConfig = make(map[string]interface{})
	}
	
	// Load from environment variables (with defaults)
	// These can be overridden by config file
	if pluginConfig["wake_word"] == nil {
		pluginConfig["wake_word"] = getEnvOrDefault("ALEXA_WAKE_WORD", "Smart Voice")
	}
	if pluginConfig["case_sensitive"] == nil {
		pluginConfig["case_sensitive"] = getEnvBoolOrDefault("ALEXA_CASE_SENSITIVE", false)
	}
	if pluginConfig["remove_from_query"] == nil {
		pluginConfig["remove_from_query"] = getEnvBoolOrDefault("ALEXA_REMOVE_FROM_QUERY", true)
	}
	if pluginConfig["default_model"] == nil {
		pluginConfig["default_model"] = getEnvOrDefault("ALEXA_DEFAULT_MODEL", "llama3.2")
	}
	
	// Create plugin with configuration
	alexaPlugin := alexaplugin.NewAlexaSkillPluginWithConfig(pluginConfig)
	
	// Set registry reference so plugin can access its context
	alexaPlugin.SetRegistry(registry)
	
	// Create plugin context with plugin name
	pluginLogger := log.With().Str("plugin", "alexa_skill").Logger()
	pluginCtx := plugins.NewPluginContextWithName("alexa_skill", llmHandler, pluginLogger, pluginConfig)
	
	// Register plugin with context
	if err := registry.RegisterWithContext(alexaPlugin, pluginCtx); err != nil {
		return err
	}
	
	log.Info().Msg("Alexa Skill plugin registered with LLM handler and configuration")
	return nil
}

// Helper functions for environment variable loading
func getEnvOrDefault(key string, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1" || val == "yes"
	}
	return defaultValue
}
