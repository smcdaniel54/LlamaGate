package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/mcpclient"
	"github.com/llamagate/llamagate/internal/response"
	"github.com/llamagate/llamagate/internal/tools"
	"github.com/rs/zerolog/log"
)

// MCPHandler handles MCP-related HTTP API endpoints
type MCPHandler struct {
	toolManager          *tools.Manager
	serverManager        *mcpclient.ServerManager
	toolExecutionTimeout time.Duration
}

// NewMCPHandler creates a new MCP API handler
func NewMCPHandler(toolManager *tools.Manager, serverManager *mcpclient.ServerManager, toolExecutionTimeout time.Duration) *MCPHandler {
	return &MCPHandler{
		toolManager:          toolManager,
		serverManager:        serverManager,
		toolExecutionTimeout: toolExecutionTimeout,
	}
}

// ServerInfo represents information about an MCP server
type ServerInfo struct {
	Name      string                       `json:"name"`
	Transport string                       `json:"transport"`
	Status    string                       `json:"status"`
	Health    *mcpclient.HealthCheckResult `json:"health,omitempty"`
	Tools     int                          `json:"tools"`
	Resources int                          `json:"resources"`
	Prompts   int                          `json:"prompts"`
}

// ListServers lists all MCP servers
// GET /v1/mcp/servers
func (h *MCPHandler) ListServers(c *gin.Context) {
	if h.serverManager == nil {
		response.ServiceUnavailable(c, "MCP is not enabled", "")
		return
	}

	serverNames := h.serverManager.ListServers()
	servers := make([]ServerInfo, 0, len(serverNames))

	for _, name := range serverNames {
		serverInfo, err := h.serverManager.GetServer(name)
		if err != nil {
			log.Warn().
				Str("server", name).
				Err(err).
				Msg("Failed to get server info")
			continue
		}

		// Get health status
		health, _ := h.serverManager.GetHealth(name)

		// Get client to count tools/resources/prompts
		client, err := h.toolManager.GetClient(name)
		var toolCount, resourceCount, promptCount int
		if err == nil {
			toolCount = len(client.GetTools())
			resourceCount = len(client.GetResources())
			promptCount = len(client.GetPrompts())
		}

		status := "unknown"
		if health != nil {
			status = health.Status.String()
		}

		servers = append(servers, ServerInfo{
			Name:      name,
			Transport: serverInfo.Transport,
			Status:    status,
			Health:    health,
			Tools:     toolCount,
			Resources: resourceCount,
			Prompts:   promptCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"servers": servers,
		"count":   len(servers),
	})
}

// GetServer gets information about a specific MCP server
// GET /v1/mcp/servers/:name
func (h *MCPHandler) GetServer(c *gin.Context) {
	if h.serverManager == nil {
		response.ServiceUnavailable(c, "MCP is not enabled", "")
		return
	}

	name := c.Param("name")
	serverInfo, err := h.serverManager.GetServer(name)
	if err != nil {
		response.NotFound(c, "Server not found", "")
		return
	}

	// Get health status
	health, _ := h.serverManager.GetHealth(name)

	// Get client from server info
	client := serverInfo.Client
	toolCount := len(client.GetTools())
	resourceCount := len(client.GetResources())
	promptCount := len(client.GetPrompts())

	status := "unknown"
	if health != nil {
		status = health.Status.String()
	}

	c.JSON(http.StatusOK, ServerInfo{
		Name:      name,
		Transport: serverInfo.Transport,
		Status:    status,
		Health:    health,
		Tools:     toolCount,
		Resources: resourceCount,
		Prompts:   promptCount,
	})
}

// GetServerHealth gets health status for a specific server
// GET /v1/mcp/servers/:name/health
func (h *MCPHandler) GetServerHealth(c *gin.Context) {
	if h.serverManager == nil {
		response.ServiceUnavailable(c, "MCP is not enabled", "")
		return
	}

	name := c.Param("name")

	// Perform immediate health check
	health, err := h.serverManager.CheckHealth(c.Request.Context(), name)
	if err != nil {
		response.NotFound(c, "Server not found", "")
		return
	}

	status := health.Status.String()

	c.JSON(http.StatusOK, gin.H{
		"server": name,
		"status": status,
		"health": health,
	})
}

// GetAllHealth gets health status for all servers
// GET /v1/mcp/servers/health
func (h *MCPHandler) GetAllHealth(c *gin.Context) {
	if h.serverManager == nil {
		response.ServiceUnavailable(c, "MCP is not enabled", "")
		return
	}

	allHealth := h.serverManager.GetAllHealth()

	// Convert to response format
	healthMap := make(map[string]interface{})
	for name, health := range allHealth {
		status := health.Status.String()

		healthMap[name] = gin.H{
			"status": status,
			"health": health,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"servers": healthMap,
		"count":   len(healthMap),
	})
}

// ServerStats represents pool statistics for a server
type ServerStats struct {
	Pool *mcpclient.PoolStats `json:"pool,omitempty"`
}

// GetServerStats gets statistics for a specific server
// GET /v1/mcp/servers/:name/stats
func (h *MCPHandler) GetServerStats(c *gin.Context) {
	if h.serverManager == nil {
		response.ServiceUnavailable(c, "MCP is not enabled", "")
		return
	}

	name := c.Param("name")
	serverInfo, err := h.serverManager.GetServer(name)
	if err != nil {
		response.NotFound(c, "Server not found", "")
		return
	}

	stats := ServerStats{}
	if serverInfo.Pool != nil {
		poolStats := serverInfo.Pool.Stats()
		stats.Pool = &poolStats
	}

	c.JSON(http.StatusOK, gin.H{
		"server": name,
		"stats":  stats,
	})
}

// ListServerTools lists tools for a specific server
// GET /v1/mcp/servers/:name/tools
func (h *MCPHandler) ListServerTools(c *gin.Context) {
	if h.serverManager == nil {
		response.ServiceUnavailable(c, "MCP is not enabled", "")
		return
	}

	name := c.Param("name")
	serverInfo, err := h.serverManager.GetServer(name)
	if err != nil {
		response.NotFound(c, "Server not found", "")
		return
	}

	client := serverInfo.Client

	tools := client.GetTools()
	c.JSON(http.StatusOK, gin.H{
		"server": name,
		"tools":  tools,
		"count":  len(tools),
	})
}

// ListServerResources lists resources for a specific server
// GET /v1/mcp/servers/:name/resources
func (h *MCPHandler) ListServerResources(c *gin.Context) {
	if h.serverManager == nil {
		response.ServiceUnavailable(c, "MCP is not enabled", "")
		return
	}

	name := c.Param("name")
	serverInfo, err := h.serverManager.GetServer(name)
	if err != nil {
		response.NotFound(c, "Server not found", "")
		return
	}

	client := serverInfo.Client

	resources := client.GetResources()
	c.JSON(http.StatusOK, gin.H{
		"server":    name,
		"resources": resources,
		"count":     len(resources),
	})
}

// ReadServerResource reads a resource from a server
// GET /v1/mcp/servers/:name/resources/*uri
func (h *MCPHandler) ReadServerResource(c *gin.Context) {
	if h.serverManager == nil {
		response.ServiceUnavailable(c, "MCP is not enabled", "")
		return
	}

	name := c.Param("name")
	uri := c.Param("uri")

	// Remove leading slash if present
	if len(uri) > 0 && uri[0] == '/' {
		uri = uri[1:]
	}

	// If URI param is empty, try query param
	if uri == "" {
		uri = c.Query("uri")
	}

	if uri == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "URI parameter is required",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	serverInfo, err := h.serverManager.GetServer(name)
	if err != nil {
		response.NotFound(c, "Server not found", "")
		return
	}

	client := serverInfo.Client

	result, err := client.ReadResource(c.Request.Context(), uri)
	if err != nil {
		response.InternalError(c, err.Error(), "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"server":   name,
		"uri":      uri,
		"contents": result.Contents, // Note: JSON tag is "contents" but struct field is "Contents"
	})
}

// ListServerPrompts lists prompts for a specific server
// GET /v1/mcp/servers/:name/prompts
func (h *MCPHandler) ListServerPrompts(c *gin.Context) {
	if h.serverManager == nil {
		response.ServiceUnavailable(c, "MCP is not enabled", "")
		return
	}

	name := c.Param("name")
	serverInfo, err := h.serverManager.GetServer(name)
	if err != nil {
		response.NotFound(c, "Server not found", "")
		return
	}

	client := serverInfo.Client

	prompts := client.GetPrompts()
	c.JSON(http.StatusOK, gin.H{
		"server":  name,
		"prompts": prompts,
		"count":   len(prompts),
	})
}

// GetServerPrompt gets a prompt template from a server
// POST /v1/mcp/servers/:name/prompts/:promptName
func (h *MCPHandler) GetServerPrompt(c *gin.Context) {
	if h.serverManager == nil {
		response.ServiceUnavailable(c, "MCP is not enabled", "")
		return
	}

	name := c.Param("name")
	promptName := c.Param("promptName")

	serverInfo, err := h.serverManager.GetServer(name)
	if err != nil {
		response.NotFound(c, "Server not found", "")
		return
	}

	client := serverInfo.Client

	// Parse arguments from request body
	var requestBody struct {
		Arguments map[string]interface{} `json:"arguments,omitempty"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil && err.Error() != "EOF" {
		// EOF is fine - arguments are optional
		log.Debug().Err(err).Msg("Failed to parse prompt arguments, using empty")
	}

	result, err := client.GetPromptTemplate(c.Request.Context(), promptName, requestBody.Arguments)
	if err != nil {
		response.InternalError(c, err.Error(), "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"server":   name,
		"prompt":   promptName,
		"messages": result.Messages,
	})
}

// ExecuteToolRequest represents a request to execute a tool
type ExecuteToolRequest struct {
	Server    string                 `json:"server" binding:"required"`
	Tool      string                 `json:"tool" binding:"required"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ExecuteToolResponse represents the result of tool execution
type ExecuteToolResponse struct {
	Server   string      `json:"server"`
	Tool     string      `json:"tool"`
	Result   interface{} `json:"result,omitempty"`
	Error    string      `json:"error,omitempty"`
	IsError  bool        `json:"is_error"`
	Duration string      `json:"duration,omitempty"`
}

// ExecuteTool executes a tool on a specific server
// POST /v1/mcp/execute
func (h *MCPHandler) ExecuteTool(c *gin.Context) {
	if h.serverManager == nil {
		response.ServiceUnavailable(c, "MCP is not enabled", "")
		return
	}

	var req ExecuteToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", "")
		return
	}

	// Get server
	serverInfo, err := h.serverManager.GetServer(req.Server)
	if err != nil {
		response.NotFound(c, "Server not found", "")
		return
	}

	client := serverInfo.Client

	// Execute tool with timeout
	timeout := h.toolExecutionTimeout
	if timeout == 0 {
		timeout = 30 * time.Second // Default fallback
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	startTime := time.Now()
	result, err := client.CallTool(ctx, req.Tool, req.Arguments)
	duration := time.Since(startTime)

	response := ExecuteToolResponse{
		Server:   req.Server,
		Tool:     req.Tool,
		Duration: duration.String(),
	}

	if err != nil {
		response.Error = err.Error()
		response.IsError = true
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Convert result to interface
	var resultData interface{}
	if len(result.Content) > 0 {
		if result.Content[0].Type == "text" {
			resultData = result.Content[0].Text
		} else {
			resultData = result.Content
		}
	}

	response.Result = resultData
	response.IsError = result.IsError

	statusCode := http.StatusOK
	if result.IsError {
		statusCode = http.StatusInternalServerError
	}

	c.JSON(statusCode, response)
}

// RefreshServerMetadata refreshes tools, resources, and prompts for a server
// POST /v1/mcp/servers/:name/refresh
func (h *MCPHandler) RefreshServerMetadata(c *gin.Context) {
	if h.serverManager == nil {
		response.ServiceUnavailable(c, "MCP is not enabled", "")
		return
	}

	name := c.Param("name")
	serverInfo, err := h.serverManager.GetServer(name)
	if err != nil {
		response.NotFound(c, "Server not found", "")
		return
	}

	client := serverInfo.Client
	ctx := c.Request.Context()

	// Refresh all metadata
	var errors []string

	if err := client.RefreshTools(ctx); err != nil {
		errors = append(errors, fmt.Sprintf("tools: %v", err))
	}

	if err := client.RefreshResources(ctx); err != nil {
		errors = append(errors, fmt.Sprintf("resources: %v", err))
	}

	if err := client.RefreshPrompts(ctx); err != nil {
		errors = append(errors, fmt.Sprintf("prompts: %v", err))
	}

	// Invalidate cache
	if h.serverManager.GetCache() != nil {
		h.serverManager.GetCache().InvalidateAll(name)
	}

	if len(errors) > 0 {
		c.JSON(http.StatusPartialContent, gin.H{
			"server": name,
			"status": "partial",
			"errors": errors,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"server": name,
		"status": "refreshed",
	})
}
