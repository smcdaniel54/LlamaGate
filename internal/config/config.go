package config

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

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
	Timeout      time.Duration // HTTP client timeout
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
	viper.SetDefault("TIMEOUT", "5m") // 5 minutes default

	cfg := &Config{
		OllamaHost:   viper.GetString("OLLAMA_HOST"),
		APIKey:       viper.GetString("API_KEY"),
		RateLimitRPS: viper.GetFloat64("RATE_LIMIT_RPS"),
		Debug:        viper.GetBool("DEBUG"),
		Port:         viper.GetString("PORT"),
		LogFile:      viper.GetString("LOG_FILE"),
	}

	// Parse timeout duration
	timeoutStr := viper.GetString("TIMEOUT")
	if timeoutStr == "" {
		timeoutStr = "5m" // Default to 5 minutes
	}
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid TIMEOUT format: %w (expected duration like '5m', '30s', '1h')", err)
	}
	cfg.Timeout = timeout

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates all configuration values
func (c *Config) Validate() error {
	// Validate OLLAMA_HOST
	if c.OllamaHost == "" {
		return fmt.Errorf("OLLAMA_HOST is required")
	}

	// Validate OLLAMA_HOST is a valid URL
	parsedURL, err := url.Parse(c.OllamaHost)
	if err != nil {
		return fmt.Errorf("invalid OLLAMA_HOST URL: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("invalid OLLAMA_HOST scheme: must be http or https")
	}

	// Validate PORT
	port, err := strconv.Atoi(c.Port)
	if err != nil {
		return fmt.Errorf("invalid PORT: %s (must be a number)", c.Port)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid PORT: %d (must be between 1 and 65535)", port)
	}

	// Validate RATE_LIMIT_RPS
	if c.RateLimitRPS <= 0 {
		return fmt.Errorf("invalid RATE_LIMIT_RPS: %f (must be greater than 0)", c.RateLimitRPS)
	}

	// Validate TIMEOUT
	if c.Timeout <= 0 {
		return fmt.Errorf("invalid TIMEOUT: %v (must be greater than 0)", c.Timeout)
	}
	if c.Timeout > 30*time.Minute {
		return fmt.Errorf("invalid TIMEOUT: %v (must be less than 30 minutes)", c.Timeout)
	}

	return nil
}
