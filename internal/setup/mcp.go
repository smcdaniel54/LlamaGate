package setup

import (
	"fmt"
	"time"

	"github.com/llamagate/llamagate/internal/config"
	"github.com/llamagate/llamagate/internal/mcpclient"
	"github.com/llamagate/llamagate/internal/proxy"
	"github.com/llamagate/llamagate/internal/tools"
	"github.com/rs/zerolog/log"
)

// MCPComponents holds initialized MCP components
type MCPComponents struct {
	ToolManager   *tools.Manager
	ServerManager *mcpclient.ServerManager
	Guardrails    *tools.Guardrails
}

// InitializeMCP initializes MCP clients, managers, and guardrails based on configuration
func InitializeMCP(cfg *config.MCPConfig) (*MCPComponents, error) {
	if cfg == nil || !cfg.Enabled {
		return nil, nil
	}

	log.Info().Msg("Initializing MCP clients...")

	toolManager := tools.NewManager()

	// Create server manager with configuration
	managerConfig := mcpclient.ManagerConfig{
		PoolSize:       cfg.ConnectionPoolSize,
		PoolIdleTime:   cfg.ConnectionIdleTime,
		HealthInterval: cfg.HealthCheckInterval,
		HealthTimeout:  cfg.HealthCheckTimeout,
		CacheTTL:       cfg.CacheTTL,
	}
	serverManager := mcpclient.NewServerManager(managerConfig)

	// Create guardrails
	guardrails, err := tools.NewGuardrails(
		cfg.AllowTools,
		cfg.DenyTools,
		cfg.MaxToolRounds,
		cfg.MaxToolCallsPerRound,
		cfg.MaxTotalToolCalls,
		cfg.DefaultToolTimeout,
		cfg.MaxToolResultSize,
	)
	if err != nil {
		serverManager.Close()
		return nil, fmt.Errorf("failed to create MCP guardrails: %w", err)
	}

	// Initialize MCP clients for each configured server
	for _, serverCfg := range cfg.Servers {
		if !serverCfg.Enabled {
			log.Debug().
				Str("server", serverCfg.Name).
				Msg("MCP server disabled, skipping")
			continue
		}

		client, err := createMCPClient(serverCfg)
		if err != nil {
			log.Error().
				Str("server", serverCfg.Name).
				Err(err).
				Msg("Failed to initialize MCP client")
			continue
		}

		// Add client to tool manager
		if err := toolManager.AddClient(client); err != nil {
			log.Error().
				Str("server", serverCfg.Name).
				Err(err).
				Msg("Failed to add MCP client to tool manager")
			client.Close()
			continue
		}

		// Add server to server manager
		if err := serverManager.AddServer(serverCfg.Name, client, serverCfg.Transport); err != nil {
			log.Error().
				Str("server", serverCfg.Name).
				Err(err).
				Msg("Failed to add server to server manager")
			// Continue anyway - tool manager has the client
		}

		log.Info().
			Str("server", serverCfg.Name).
			Str("transport", serverCfg.Transport).
			Msg("MCP client initialized successfully")
	}

	toolCount := len(toolManager.GetAllTools())
	log.Info().
		Int("total_tools", toolCount).
		Msg("MCP initialization complete")

	return &MCPComponents{
		ToolManager:   toolManager,
		ServerManager: serverManager,
		Guardrails:    guardrails,
	}, nil
}

// createMCPClient creates an MCP client based on server configuration
func createMCPClient(serverCfg config.MCPServerConfig) (*mcpclient.Client, error) {
	// Use server timeout or default
	timeout := serverCfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second // Default timeout
	}

	switch serverCfg.Transport {
	case "stdio":
		return mcpclient.NewClientWithTimeout(
			serverCfg.Name,
			serverCfg.Command,
			serverCfg.Args,
			serverCfg.Env,
			timeout,
		)
	case "http":
		return mcpclient.NewClientWithHTTP(
			serverCfg.Name,
			serverCfg.URL,
			serverCfg.Headers,
			timeout,
		)
	case "sse":
		return nil, fmt.Errorf("SSE transport not yet implemented")
	default:
		return nil, fmt.Errorf("unknown transport type: %s", serverCfg.Transport)
	}
}

// ConfigureProxy configures the proxy with MCP components
func ConfigureProxy(p *proxy.Proxy, components *MCPComponents, mcpConfig *config.MCPConfig) {
	if components == nil {
		return
	}

	// Set tool manager and guardrails on proxy
	p.SetToolManager(components.ToolManager, components.Guardrails)

	// Set server manager on proxy for MCP resource context injection
	p.SetServerManager(components.ServerManager)

	// Set resource fetch timeout
	if mcpConfig != nil && mcpConfig.ResourceFetchTimeout > 0 {
		p.SetResourceFetchTimeout(mcpConfig.ResourceFetchTimeout)
	}
}
