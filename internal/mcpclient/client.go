package mcpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Client represents an MCP client connection
type Client struct {
	name         string
	transport    Transport
	mu           sync.RWMutex
	initialized  bool
	serverInfo   *ServerInfo
	capabilities *ServerCapabilities
	tools        []Tool
	toolsMap     map[string]*Tool // For quick lookup
}

// NewClient creates a new MCP client with stdio transport
func NewClient(name, command string, args []string, env map[string]string) (*Client, error) {
	transport, err := NewStdioTransport(command, args, env)
	if err != nil {
		return nil, fmt.Errorf("failed to create stdio transport: %w", err)
	}

	client := &Client{
		name:      name,
		transport: transport,
		toolsMap:  make(map[string]*Tool),
	}

	// Initialize the connection
	if err := client.initialize(context.Background()); err != nil {
		transport.Close()
		return nil, fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	// Discover tools
	if err := client.discoverTools(context.Background()); err != nil {
		log.Warn().
			Str("server", name).
			Err(err).
			Msg("Failed to discover tools, continuing without tools")
	}

	return client, nil
}

// NewClientWithSSE creates a new MCP client with SSE transport
// Note: This is a stub for future implementation
func NewClientWithSSE(name, url string, headers map[string]string) (*Client, error) {
	transport, err := NewSSETransport(url, headers)
	if err != nil {
		return nil, err
	}

	return &Client{
		name:      name,
		transport: transport,
		toolsMap:  make(map[string]*Tool),
	}, nil
}

// initialize performs the MCP initialization handshake
func (c *Client) initialize(ctx context.Context) error {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05", // MCP protocol version
		Capabilities:    ClientCapabilities{},
		ClientInfo: ClientInfo{
			Name:    "llamagate",
			Version: "1.1.0",
		},
	}

	resp, err := c.transport.SendRequest(ctx, "initialize", params)
	if err != nil {
		return fmt.Errorf("initialize request failed: %w", err)
	}

	var result InitializeResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return fmt.Errorf("failed to unmarshal initialize result: %w", err)
	}

	// Send initialized notification
	_, err = c.transport.SendRequest(ctx, "notifications/initialized", nil)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to send initialized notification")
	}

	c.mu.Lock()
	c.initialized = true
	c.serverInfo = &result.ServerInfo
	c.capabilities = &result.Capabilities
	c.mu.Unlock()

	log.Info().
		Str("server", c.name).
		Str("server_name", result.ServerInfo.Name).
		Str("server_version", result.ServerInfo.Version).
		Msg("MCP client initialized")

	return nil
}

// discoverTools discovers available tools from the MCP server
func (c *Client) discoverTools(ctx context.Context) error {
	resp, err := c.transport.SendRequest(ctx, "tools/list", nil)
	if err != nil {
		return fmt.Errorf("tools/list request failed: %w", err)
	}

	var result ToolsListResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return fmt.Errorf("failed to unmarshal tools/list result: %w", err)
	}

	c.mu.Lock()
	c.tools = result.Tools
	c.toolsMap = make(map[string]*Tool, len(result.Tools))
	for i := range result.Tools {
		tool := &result.Tools[i]
		c.toolsMap[tool.Name] = tool
	}
	c.mu.Unlock()

	log.Info().
		Str("server", c.name).
		Int("tool_count", len(result.Tools)).
		Msg("Discovered tools from MCP server")

	return nil
}

// GetTools returns all available tools
func (c *Client) GetTools() []Tool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tools := make([]Tool, len(c.tools))
	copy(tools, c.tools)
	return tools
}

// GetTool returns a tool by name
func (c *Client) GetTool(name string) (*Tool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tool, ok := c.toolsMap[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrToolNotFound, name)
	}

	return tool, nil
}

// CallTool executes a tool with the given arguments
func (c *Client) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*ToolCallResult, error) {
	c.mu.RLock()
	if !c.initialized {
		c.mu.RUnlock()
		return nil, ErrNotInitialized
	}
	c.mu.RUnlock()

	params := ToolCallParams{
		Name:      toolName,
		Arguments: arguments,
	}

	resp, err := c.transport.SendRequest(ctx, "tools/call", params)
	if err != nil {
		return nil, fmt.Errorf("tools/call request failed: %w", err)
	}

	var result ToolCallResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tools/call result: %w", err)
	}

	return &result, nil
}

// CallToolWithTimeout executes a tool with a timeout
func (c *Client) CallToolWithTimeout(toolName string, arguments map[string]interface{}, timeout time.Duration) (*ToolCallResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return c.CallTool(ctx, toolName, arguments)
}

// RefreshTools refreshes the tool list from the server
func (c *Client) RefreshTools(ctx context.Context) error {
	return c.discoverTools(ctx)
}

// GetServerInfo returns server information
func (c *Client) GetServerInfo() *ServerInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serverInfo
}

// GetName returns the client name
func (c *Client) GetName() string {
	return c.name
}

// IsInitialized returns whether the client is initialized
func (c *Client) IsInitialized() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.initialized
}

// Close closes the MCP client connection
func (c *Client) Close() error {
	return c.transport.Close()
}

// IsClosed returns whether the client is closed
func (c *Client) IsClosed() bool {
	return c.transport.IsClosed()
}
