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
	resources    []Resource
	resourcesMap map[string]*Resource // For quick lookup by URI
	prompts      []Prompt
	promptsMap   map[string]*Prompt // For quick lookup by name
}

// NewClient creates a new MCP client with stdio transport
func NewClient(name, command string, args []string, env map[string]string) (*Client, error) {
	return NewClientWithTimeout(name, command, args, env, 0)
}

// NewClientWithTimeout creates a new MCP client with stdio transport and a timeout
// If timeout is 0, it uses context.Background() (no timeout)
func NewClientWithTimeout(name, command string, args []string, env map[string]string, timeout time.Duration) (*Client, error) {
	transport, err := NewStdioTransport(command, args, env)
	if err != nil {
		return nil, fmt.Errorf("failed to create stdio transport: %w", err)
	}

	client := &Client{
		name:         name,
		transport:    transport,
		toolsMap:     make(map[string]*Tool),
		resourcesMap: make(map[string]*Resource),
		promptsMap:   make(map[string]*Prompt),
	}

	// Create context with timeout if specified
	var ctx context.Context
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	// Initialize the connection
	if err := client.initialize(ctx); err != nil {
		transport.Close()
		return nil, fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	// Discover tools
	if err := client.discoverTools(ctx); err != nil {
		log.Warn().
			Str("server", name).
			Err(err).
			Msg("Failed to discover tools, continuing without tools")
	}

	// Discover resources
	if err := client.discoverResources(ctx); err != nil {
		log.Warn().
			Str("server", name).
			Err(err).
			Msg("Failed to discover resources, continuing without resources")
	}

	// Discover prompts
	if err := client.discoverPrompts(ctx); err != nil {
		log.Warn().
			Str("server", name).
			Err(err).
			Msg("Failed to discover prompts, continuing without prompts")
	}

	return client, nil
}

// NewClientWithSSE creates a new MCP client with SSE transport
func NewClientWithSSE(name, url string, headers map[string]string) (*Client, error) {
	transport, err := NewSSETransport(url, headers)
	if err != nil {
		return nil, err
	}

	client := &Client{
		name:         name,
		transport:    transport,
		toolsMap:     make(map[string]*Tool),
		resourcesMap: make(map[string]*Resource),
		promptsMap:   make(map[string]*Prompt),
	}

	// Initialize the client (MCP handshake)
	ctx := context.Background()
	if err := client.initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	// Discover available tools
	if err := client.discoverTools(ctx); err != nil {
		log.Warn().
			Str("server", name).
			Err(err).
			Msg("Failed to discover tools, continuing without tools")
	}

	// Discover available resources
	if err := client.discoverResources(ctx); err != nil {
		log.Warn().
			Str("server", name).
			Err(err).
			Msg("Failed to discover resources, continuing without resources")
	}

	// Discover available prompts
	if err := client.discoverPrompts(ctx); err != nil {
		log.Warn().
			Str("server", name).
			Err(err).
			Msg("Failed to discover prompts, continuing without prompts")
	}

	return client, nil
}

// initialize performs the MCP initialization handshake
func (c *Client) initialize(ctx context.Context) error {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05", // MCP protocol version
		Capabilities:    ClientCapabilities{},
		ClientInfo: ClientInfo{
			Name:    "llamagate",
			Version: "0.9.0",
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

// NewClientWithHTTP creates a new MCP client with HTTP transport
func NewClientWithHTTP(name, url string, headers map[string]string, timeout time.Duration) (*Client, error) {
	transport, err := NewHTTPTransport(url, headers, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP transport: %w", err)
	}

	client := &Client{
		name:         name,
		transport:    transport,
		toolsMap:     make(map[string]*Tool),
		resourcesMap: make(map[string]*Resource),
		promptsMap:   make(map[string]*Prompt),
	}

	// Create context with timeout if specified
	var ctx context.Context
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	// Initialize the connection
	if err := client.initialize(ctx); err != nil {
		transport.Close()
		return nil, fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	// Discover tools
	if err := client.discoverTools(ctx); err != nil {
		log.Warn().
			Str("server", name).
			Err(err).
			Msg("Failed to discover tools, continuing without tools")
	}

	// Discover resources
	if err := client.discoverResources(ctx); err != nil {
		log.Warn().
			Str("server", name).
			Err(err).
			Msg("Failed to discover resources, continuing without resources")
	}

	// Discover prompts
	if err := client.discoverPrompts(ctx); err != nil {
		log.Warn().
			Str("server", name).
			Err(err).
			Msg("Failed to discover prompts, continuing without prompts")
	}

	return client, nil
}

// discoverResources discovers available resources from the MCP server
func (c *Client) discoverResources(ctx context.Context) error {
	resp, err := c.transport.SendRequest(ctx, "resources/list", nil)
	if err != nil {
		return fmt.Errorf("resources/list request failed: %w", err)
	}

	var result ResourcesListResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return fmt.Errorf("failed to unmarshal resources/list result: %w", err)
	}

	c.mu.Lock()
	c.resources = result.Resources
	c.resourcesMap = make(map[string]*Resource, len(result.Resources))
	for i := range result.Resources {
		resource := &result.Resources[i]
		c.resourcesMap[resource.URI] = resource
	}
	c.mu.Unlock()

	log.Info().
		Str("server", c.name).
		Int("resource_count", len(result.Resources)).
		Msg("Discovered resources from MCP server")

	return nil
}

// GetResources returns all available resources
func (c *Client) GetResources() []Resource {
	c.mu.RLock()
	defer c.mu.RUnlock()

	resources := make([]Resource, len(c.resources))
	copy(resources, c.resources)
	return resources
}

// GetResource returns a resource by URI
func (c *Client) GetResource(uri string) (*Resource, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	resource, ok := c.resourcesMap[uri]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", uri)
	}

	return resource, nil
}

// ReadResource reads a resource by URI
func (c *Client) ReadResource(ctx context.Context, uri string) (*ResourceReadResult, error) {
	c.mu.RLock()
	if !c.initialized {
		c.mu.RUnlock()
		return nil, ErrNotInitialized
	}
	c.mu.RUnlock()

	params := ResourceReadParams{
		URI: uri,
	}

	resp, err := c.transport.SendRequest(ctx, "resources/read", params)
	if err != nil {
		return nil, fmt.Errorf("resources/read request failed: %w", err)
	}

	var result ResourceReadResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resources/read result: %w", err)
	}

	return &result, nil
}

// RefreshResources refreshes the resource list from the server
func (c *Client) RefreshResources(ctx context.Context) error {
	return c.discoverResources(ctx)
}

// discoverPrompts discovers available prompts from the MCP server
func (c *Client) discoverPrompts(ctx context.Context) error {
	resp, err := c.transport.SendRequest(ctx, "prompts/list", nil)
	if err != nil {
		return fmt.Errorf("prompts/list request failed: %w", err)
	}

	var result PromptsListResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return fmt.Errorf("failed to unmarshal prompts/list result: %w", err)
	}

	c.mu.Lock()
	c.prompts = result.Prompts
	c.promptsMap = make(map[string]*Prompt, len(result.Prompts))
	for i := range result.Prompts {
		prompt := &result.Prompts[i]
		c.promptsMap[prompt.Name] = prompt
	}
	c.mu.Unlock()

	log.Info().
		Str("server", c.name).
		Int("prompt_count", len(result.Prompts)).
		Msg("Discovered prompts from MCP server")

	return nil
}

// GetPrompts returns all available prompts
func (c *Client) GetPrompts() []Prompt {
	c.mu.RLock()
	defer c.mu.RUnlock()

	prompts := make([]Prompt, len(c.prompts))
	copy(prompts, c.prompts)
	return prompts
}

// GetPrompt returns a prompt by name
func (c *Client) GetPrompt(name string) (*Prompt, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	prompt, ok := c.promptsMap[name]
	if !ok {
		return nil, fmt.Errorf("prompt not found: %s", name)
	}

	return prompt, nil
}

// GetPromptTemplate gets a prompt template with arguments
func (c *Client) GetPromptTemplate(ctx context.Context, name string, arguments map[string]interface{}) (*PromptGetResult, error) {
	c.mu.RLock()
	if !c.initialized {
		c.mu.RUnlock()
		return nil, ErrNotInitialized
	}
	c.mu.RUnlock()

	params := PromptGetParams{
		Name:      name,
		Arguments: arguments,
	}

	resp, err := c.transport.SendRequest(ctx, "prompts/get", params)
	if err != nil {
		return nil, fmt.Errorf("prompts/get request failed: %w", err)
	}

	var result PromptGetResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prompts/get result: %w", err)
	}

	return &result, nil
}

// RefreshPrompts refreshes the prompt list from the server
func (c *Client) RefreshPrompts(ctx context.Context) error {
	return c.discoverPrompts(ctx)
}
