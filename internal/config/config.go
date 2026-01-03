// Package config provides configuration management for LlamaGate.
package config

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
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
	MCP          *MCPConfig    // MCP configuration (optional)
}

// MCPConfig holds MCP client configuration
type MCPConfig struct {
	Enabled              bool
	MaxToolRounds        int
	MaxToolCallsPerRound int
	DefaultToolTimeout   time.Duration
	MaxToolResultSize    int64
	AllowTools           []string
	DenyTools            []string
	Servers              []MCPServerConfig
}

// MCPServerConfig holds configuration for a single MCP server
type MCPServerConfig struct {
	Name      string
	Enabled   bool
	Transport string // "stdio" or "sse"
	// stdio fields
	Command string
	Args    []string
	Env     map[string]string
	// SSE fields
	URL     string
	Headers map[string]string
	// common
	Timeout time.Duration
}

// Load reads configuration from .env file (if present) and environment variables using Viper
// Also supports YAML/JSON config files for MCP configuration
// Environment variables take precedence over .env file values
func Load() (*Config, error) {
	// Try to load .env file (ignore error if it doesn't exist)
	_ = godotenv.Load()

	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	// Try to load config file (yaml or json)
	viper.SetConfigName("llamagate")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.llamagate")
	_ = viper.ReadInConfig() // Ignore error if config file doesn't exist

	// Set defaults
	viper.SetDefault("OLLAMA_HOST", "http://localhost:11434")
	viper.SetDefault("API_KEY", "")
	viper.SetDefault("RATE_LIMIT_RPS", 10.0)
	viper.SetDefault("DEBUG", false)
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("LOG_FILE", "")
	viper.SetDefault("TIMEOUT", "5m") // 5 minutes default

	// MCP defaults
	viper.SetDefault("MCP_ENABLED", false)
	viper.SetDefault("MCP_MAX_TOOL_ROUNDS", 10)
	viper.SetDefault("MCP_MAX_TOOL_CALLS_PER_ROUND", 10)
	viper.SetDefault("MCP_DEFAULT_TOOL_TIMEOUT", "30s")
	viper.SetDefault("MCP_MAX_TOOL_RESULT_SIZE", 1024*1024) // 1MB

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

	// Load MCP configuration
	mcpConfig, err := loadMCPConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load MCP config: %w", err)
	}
	cfg.MCP = mcpConfig

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadMCPConfig loads MCP configuration from viper
func loadMCPConfig() (*MCPConfig, error) {
	enabled := viper.GetBool("MCP_ENABLED")
	if !enabled {
		// Check if MCP config exists in YAML/JSON
		if viper.IsSet("mcp") {
			enabled = true
		} else {
			return nil, nil // MCP not enabled
		}
	}

	mcp := &MCPConfig{
		Enabled:              enabled,
		MaxToolRounds:        viper.GetInt("MCP_MAX_TOOL_ROUNDS"),
		MaxToolCallsPerRound: viper.GetInt("MCP_MAX_TOOL_CALLS_PER_ROUND"),
		MaxToolResultSize:    viper.GetInt64("MCP_MAX_TOOL_RESULT_SIZE"),
	}

	// Parse default tool timeout
	timeoutStr := viper.GetString("MCP_DEFAULT_TOOL_TIMEOUT")
	if timeoutStr == "" {
		timeoutStr = "30s"
	}
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid MCP_DEFAULT_TOOL_TIMEOUT format: %w", err)
	}
	mcp.DefaultToolTimeout = timeout

	// Load allow/deny tools from env vars (comma-separated)
	if allowStr := viper.GetString("MCP_ALLOW_TOOLS"); allowStr != "" {
		mcp.AllowTools = strings.FieldsFunc(allowStr, func(c rune) bool {
			return c == ',' || c == ' '
		})
	}
	if denyStr := viper.GetString("MCP_DENY_TOOLS"); denyStr != "" {
		mcp.DenyTools = strings.FieldsFunc(denyStr, func(c rune) bool {
			return c == ',' || c == ' '
		})
	}

	// Load servers from config file (YAML/JSON)
	if viper.IsSet("mcp.servers") {
		var servers []MCPServerConfig
		if err := viper.UnmarshalKey("mcp.servers", &servers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal MCP servers: %w", err)
		}
		mcp.Servers = servers
	}

	// Also support environment variable format for simple cases
	// MCP_SERVER_<NAME>_COMMAND, MCP_SERVER_<NAME>_ARGS, etc.
	// This is more complex, so we'll focus on YAML/JSON for now

	return mcp, nil
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

	// Validate MCP configuration if enabled
	if c.MCP != nil && c.MCP.Enabled {
		if err := c.MCP.Validate(); err != nil {
			return fmt.Errorf("invalid MCP configuration: %w", err)
		}
	}

	return nil
}

// Validate validates MCP configuration
func (m *MCPConfig) Validate() error {
	if m.MaxToolRounds <= 0 {
		return fmt.Errorf("MCP_MAX_TOOL_ROUNDS must be greater than 0")
	}
	if m.MaxToolCallsPerRound <= 0 {
		return fmt.Errorf("MCP_MAX_TOOL_CALLS_PER_ROUND must be greater than 0")
	}
	if m.DefaultToolTimeout <= 0 {
		return fmt.Errorf("MCP_DEFAULT_TOOL_TIMEOUT must be greater than 0")
	}
	if m.MaxToolResultSize <= 0 {
		return fmt.Errorf("MCP_MAX_TOOL_RESULT_SIZE must be greater than 0")
	}

	// Validate servers
	for i, server := range m.Servers {
		if err := server.Validate(); err != nil {
			return fmt.Errorf("server %d (%s): %w", i, server.Name, err)
		}
	}

	return nil
}

// Validate validates MCP server configuration
func (s *MCPServerConfig) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("server name is required")
	}
	if s.Transport == "" {
		return fmt.Errorf("transport is required (stdio or sse)")
	}
	if s.Transport != "stdio" && s.Transport != "sse" {
		return fmt.Errorf("invalid transport: %s (must be stdio or sse)", s.Transport)
	}

	if s.Transport == "stdio" {
		if s.Command == "" {
			return fmt.Errorf("command is required for stdio transport")
		}
	} else if s.Transport == "sse" {
		if s.URL == "" {
			return fmt.Errorf("URL is required for SSE transport")
		}
	}

	if s.Timeout <= 0 {
		s.Timeout = 30 * time.Second // Default timeout
	}

	return nil
}
