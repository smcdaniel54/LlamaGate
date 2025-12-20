package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	OllamaHost   string
	APIKey       string
	RateLimitRPS float64
	Debug        bool
	Port         string
	LogFile      string
}

// Load reads configuration from .env file (if present) and environment variables using Viper
// Environment variables take precedence over .env file values
func Load() (*Config, error) {
	// Try to load .env file (ignore error if it doesn't exist)
	_ = godotenv.Load()

	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("OLLAMA_HOST", "http://localhost:11434")
	viper.SetDefault("API_KEY", "")
	viper.SetDefault("RATE_LIMIT_RPS", 10.0)
	viper.SetDefault("DEBUG", false)
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("LOG_FILE", "")

	cfg := &Config{
		OllamaHost:   viper.GetString("OLLAMA_HOST"),
		APIKey:       viper.GetString("API_KEY"),
		RateLimitRPS: viper.GetFloat64("RATE_LIMIT_RPS"),
		Debug:        viper.GetBool("DEBUG"),
		Port:         viper.GetString("PORT"),
		LogFile:      viper.GetString("LOG_FILE"),
	}

	// Validate required config
	if cfg.OllamaHost == "" {
		return nil, fmt.Errorf("OLLAMA_HOST is required")
	}

	return cfg, nil
}
