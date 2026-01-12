package config

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestEnv() {
	// Clear viper
	viper.Reset()
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	// Clear environment variables
	os.Clearenv()
}

func TestLoad_Defaults(t *testing.T) {
	setupTestEnv()

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "http://localhost:11434", cfg.OllamaHost)
	assert.Equal(t, "", cfg.APIKey)
	assert.Equal(t, 50.0, cfg.RateLimitRPS)
	assert.Equal(t, false, cfg.Debug)
	assert.Equal(t, "11435", cfg.Port)
	assert.Equal(t, "", cfg.LogFile)
	assert.Equal(t, 5*time.Minute, cfg.Timeout)
	assert.Nil(t, cfg.MCP) // MCP disabled by default
}

func TestLoad_FromEnvVars(t *testing.T) {
	setupTestEnv()

	os.Setenv("OLLAMA_HOST", "http://example.com:11434")
	os.Setenv("API_KEY", "test-key-123")
	os.Setenv("RATE_LIMIT_RPS", "20.5")
	os.Setenv("DEBUG", "true")
	os.Setenv("PORT", "9090")
	os.Setenv("LOG_FILE", "/tmp/llamagate.log")
	os.Setenv("TIMEOUT", "10m")

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "http://example.com:11434", cfg.OllamaHost)
	assert.Equal(t, "test-key-123", cfg.APIKey)
	assert.Equal(t, 20.5, cfg.RateLimitRPS)
	assert.Equal(t, true, cfg.Debug)
	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "/tmp/llamagate.log", cfg.LogFile)
	assert.Equal(t, 10*time.Minute, cfg.Timeout)
}

func TestLoad_APIKeyTrimsWhitespace(t *testing.T) {
	setupTestEnv()

	os.Setenv("API_KEY", "  test-key-123  ")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "test-key-123", cfg.APIKey)
}

func TestLoad_InvalidTimeout(t *testing.T) {
	setupTestEnv()

	os.Setenv("TIMEOUT", "invalid-duration")

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "invalid TIMEOUT format")
}

func TestLoad_ValidTimeoutFormats(t *testing.T) {
	setupTestEnv()

	testCases := []struct {
		timeoutStr string
		expected   time.Duration
	}{
		{"5m", 5 * time.Minute},
		{"30s", 30 * time.Second},
		{"10m", 10 * time.Minute},
		{"29m", 29 * time.Minute},
	}

	for _, tc := range testCases {
		t.Run(tc.timeoutStr, func(t *testing.T) {
			setupTestEnv()
			os.Setenv("TIMEOUT", tc.timeoutStr)

			cfg, err := Load()
			require.NoError(t, err)
			assert.Equal(t, tc.expected, cfg.Timeout)
		})
	}
}

func TestLoad_MCPEnabled(t *testing.T) {
	setupTestEnv()

	os.Setenv("MCP_ENABLED", "true")
	os.Setenv("MCP_MAX_TOOL_ROUNDS", "5")
	os.Setenv("MCP_MAX_TOOL_CALLS_PER_ROUND", "3")
	os.Setenv("MCP_MAX_TOTAL_TOOL_CALLS", "20")
	os.Setenv("MCP_DEFAULT_TOOL_TIMEOUT", "15s")
	os.Setenv("MCP_MAX_TOOL_RESULT_SIZE", "2048")

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg.MCP)
	assert.True(t, cfg.MCP.Enabled)
	assert.Equal(t, 5, cfg.MCP.MaxToolRounds)
	assert.Equal(t, 3, cfg.MCP.MaxToolCallsPerRound)
	assert.Equal(t, 20, cfg.MCP.MaxTotalToolCalls)
	assert.Equal(t, int64(2048), cfg.MCP.MaxToolResultSize)
	assert.Equal(t, 15*time.Second, cfg.MCP.DefaultToolTimeout)
}

func TestLoad_MCPDefaults(t *testing.T) {
	setupTestEnv()

	os.Setenv("MCP_ENABLED", "true")

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg.MCP)

	assert.Equal(t, 10, cfg.MCP.MaxToolRounds)
	assert.Equal(t, 10, cfg.MCP.MaxToolCallsPerRound)
	assert.Equal(t, 50, cfg.MCP.MaxTotalToolCalls) // Default max total tool calls
	assert.Equal(t, int64(1024*1024), cfg.MCP.MaxToolResultSize) // 1MB
	assert.Equal(t, 30*time.Second, cfg.MCP.DefaultToolTimeout)
	assert.Equal(t, 5*time.Minute, cfg.MCP.ConnectionIdleTime)
	assert.Equal(t, 60*time.Second, cfg.MCP.HealthCheckInterval)
	assert.Equal(t, 5*time.Second, cfg.MCP.HealthCheckTimeout)
	assert.Equal(t, 5*time.Minute, cfg.MCP.CacheTTL)
}

func TestLoad_MCPAllowDenyTools(t *testing.T) {
	setupTestEnv()

	os.Setenv("MCP_ENABLED", "true")
	os.Setenv("MCP_ALLOW_TOOLS", "tool1,tool2,tool3")
	os.Setenv("MCP_DENY_TOOLS", "tool4, tool5")

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg.MCP)

	assert.Equal(t, []string{"tool1", "tool2", "tool3"}, cfg.MCP.AllowTools)
	assert.Equal(t, []string{"tool4", "tool5"}, cfg.MCP.DenyTools)
}

func TestLoad_MCPInvalidDuration(t *testing.T) {
	setupTestEnv()

	os.Setenv("MCP_ENABLED", "true")
	os.Setenv("MCP_CONNECTION_IDLE_TIME", "invalid")

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "invalid MCP_CONNECTION_IDLE_TIME format")
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				OllamaHost:   "http://localhost:11434",
				Port:         "11435",
				RateLimitRPS: 10.0,
				Timeout:      5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "missing OllamaHost",
			config: &Config{
				OllamaHost:   "",
				Port:         "11435",
				RateLimitRPS: 10.0,
				Timeout:      5 * time.Minute,
			},
			wantErr: true,
			errMsg:  "OLLAMA_HOST is required",
		},
		{
			name: "invalid OllamaHost URL",
			config: &Config{
				OllamaHost:   "not-a-url",
				Port:         "11435",
				RateLimitRPS: 10.0,
				Timeout:      5 * time.Minute,
			},
			wantErr: true,
			errMsg:  "invalid OLLAMA_HOST",
		},
		{
			name: "invalid OllamaHost scheme",
			config: &Config{
				OllamaHost:   "ftp://localhost:11434",
				Port:         "11435",
				RateLimitRPS: 10.0,
				Timeout:      5 * time.Minute,
			},
			wantErr: true,
			errMsg:  "invalid OLLAMA_HOST scheme",
		},
		{
			name: "negative rate limit",
			config: &Config{
				OllamaHost:   "http://localhost:11434",
				Port:         "11435",
				RateLimitRPS: -1.0,
				Timeout:      5 * time.Minute,
			},
			wantErr: true,
			errMsg:  "invalid RATE_LIMIT_RPS",
		},
		{
			name: "zero rate limit",
			config: &Config{
				OllamaHost:   "http://localhost:11434",
				Port:         "11435",
				RateLimitRPS: 0.0,
				Timeout:      5 * time.Minute,
			},
			wantErr: true,
			errMsg:  "invalid RATE_LIMIT_RPS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMCPServerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  MCPServerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid stdio config",
			config: MCPServerConfig{
				Name:      "test-server",
				Enabled:   true,
				Transport: "stdio",
				Command:   "npx",
				Args:      []string{"-y", "@modelcontextprotocol/server-filesystem"},
			},
			wantErr: false,
		},
		{
			name: "valid http config",
			config: MCPServerConfig{
				Name:      "test-server",
				Enabled:   true,
				Transport: "http",
				URL:       "http://localhost:3000",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			config: MCPServerConfig{
				Name:      "",
				Enabled:   true,
				Transport: "stdio",
				Command:   "npx",
			},
			wantErr: true,
			errMsg:  "server name is required",
		},
		{
			name: "stdio missing command",
			config: MCPServerConfig{
				Name:      "test-server",
				Enabled:   true,
				Transport: "stdio",
				Command:   "",
			},
			wantErr: true,
			errMsg:  "command is required for stdio transport",
		},
		{
			name: "http missing URL",
			config: MCPServerConfig{
				Name:      "test-server",
				Enabled:   true,
				Transport: "http",
				URL:       "",
			},
			wantErr: true,
			errMsg:  "URL is required for http transport",
		},
		{
			name: "invalid transport",
			config: MCPServerConfig{
				Name:      "test-server",
				Enabled:   true,
				Transport: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid transport",
		},
		{
			name: "empty URL for http",
			config: MCPServerConfig{
				Name:      "test-server",
				Enabled:   true,
				Transport: "http",
				URL:       "",
			},
			wantErr: true,
			errMsg:  "URL is required for http transport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
