package tools

import (
	"fmt"
	"sync"

	"github.com/llamagate/llamagate/internal/mcpclient"
	"github.com/rs/zerolog/log"
)

// Manager manages tools from multiple MCP servers
type Manager struct {
	mu     sync.RWMutex
	tools  map[string]*Tool // keyed by namespaced name
	clients map[string]*mcpclient.Client // keyed by server name
}

// NewManager creates a new tool manager
func NewManager() *Manager {
	return &Manager{
		tools:   make(map[string]*Tool),
		clients: make(map[string]*mcpclient.Client),
	}
}

// AddClient adds an MCP client and registers its tools
func (m *Manager) AddClient(client *mcpclient.Client) error {
	serverName := client.GetName()

	mcpTools := client.GetTools()
	if len(mcpTools) == 0 {
		log.Info().
			Str("server", serverName).
			Msg("MCP server has no tools")
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if client already exists
	if _, exists := m.clients[serverName]; exists {
		return fmt.Errorf("client %s already registered", serverName)
	}

	m.clients[serverName] = client

	// Register tools with namespace
	for _, mcpTool := range mcpTools {
		namespacedName := fmt.Sprintf("mcp.%s.%s", serverName, mcpTool.Name)
		
		tool := &Tool{
			NamespacedName: namespacedName,
			ServerName:     serverName,
			OriginalName:   mcpTool.Name,
			Description:   mcpTool.Description,
			InputSchema:   mcpTool.InputSchema,
		}

		m.tools[namespacedName] = tool

		log.Debug().
			Str("server", serverName).
			Str("original_name", mcpTool.Name).
			Str("namespaced_name", namespacedName).
			Msg("Registered tool")
	}

	log.Info().
		Str("server", serverName).
		Int("tool_count", len(mcpTools)).
		Msg("Registered tools from MCP server")

	return nil
}

// RemoveClient removes an MCP client and its tools
func (m *Manager) RemoveClient(serverName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.clients[serverName]
	if !exists {
		return fmt.Errorf("client %s not found", serverName)
	}

	// Remove all tools from this server
	for namespacedName, tool := range m.tools {
		if tool.ServerName == serverName {
			delete(m.tools, namespacedName)
		}
	}

	delete(m.clients, serverName)

	if err := client.Close(); err != nil {
		log.Warn().
			Str("server", serverName).
			Err(err).
			Msg("Error closing MCP client")
	}

	log.Info().
		Str("server", serverName).
		Msg("Removed MCP client and its tools")

	return nil
}

// GetTool returns a tool by its namespaced name
func (m *Manager) GetTool(namespacedName string) (*Tool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tool, ok := m.tools[namespacedName]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", namespacedName)
	}

	return tool, nil
}

// GetAllTools returns all registered tools
func (m *Manager) GetAllTools() []*Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools := make([]*Tool, 0, len(m.tools))
	for _, tool := range m.tools {
		tools = append(tools, tool)
	}

	return tools
}

// GetClient returns the MCP client for a server
func (m *Manager) GetClient(serverName string) (*mcpclient.Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, ok := m.clients[serverName]
	if !ok {
		return nil, fmt.Errorf("client not found: %s", serverName)
	}

	return client, nil
}

// GetClientForTool returns the MCP client for a tool's server
func (m *Manager) GetClientForTool(namespacedName string) (*mcpclient.Client, error) {
	tool, err := m.GetTool(namespacedName)
	if err != nil {
		return nil, err
	}

	return m.GetClient(tool.ServerName)
}

// CloseAll closes all MCP clients
func (m *Manager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error
	for serverName, client := range m.clients {
		if err := client.Close(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			log.Warn().
				Str("server", serverName).
				Err(err).
				Msg("Error closing MCP client")
		}
	}

	m.tools = make(map[string]*Tool)
	m.clients = make(map[string]*mcpclient.Client)

	return firstErr
}

