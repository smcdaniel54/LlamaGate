package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/middleware"
	"github.com/llamagate/llamagate/internal/plugins"
	"github.com/llamagate/llamagate/internal/response"
	"github.com/rs/zerolog/log"
)

// PluginHandler handles plugin-related HTTP requests
type PluginHandler struct {
	registry *plugins.Registry
}

// NewPluginHandler creates a new plugin handler
func NewPluginHandler(registry *plugins.Registry) *PluginHandler {
	return &PluginHandler{
		registry: registry,
	}
}

// ListPlugins lists all registered plugins
// GET /v1/plugins
func (h *PluginHandler) ListPlugins(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	if h.registry == nil {
		response.ServiceUnavailable(c, "Plugin system not available", requestID)
		return
	}

	metadatas := h.registry.List()
	c.JSON(http.StatusOK, gin.H{
		"plugins": metadatas,
		"count":   len(metadatas),
	})
}

// GetPlugin gets information about a specific plugin
// GET /v1/plugins/:name
func (h *PluginHandler) GetPlugin(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	pluginName := c.Param("name")

	if h.registry == nil {
		response.ServiceUnavailable(c, "Plugin system not available", requestID)
		return
	}

	plugin, err := h.registry.Get(pluginName)
	if err != nil {
		response.NotFound(c, "Plugin not found", requestID)
		return
	}

	c.JSON(http.StatusOK, plugin.Metadata())
}

// ExecutePlugin executes a plugin with the provided input
// POST /v1/plugins/:name/execute
func (h *PluginHandler) ExecutePlugin(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	pluginName := c.Param("name")

	if h.registry == nil {
		response.ServiceUnavailable(c, "Plugin system not available", requestID)
		return
	}

	plugin, err := h.registry.Get(pluginName)
	if err != nil {
		response.NotFound(c, "Plugin not found", requestID)
		return
	}

	// Parse request body
	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("plugin", pluginName).
			Err(err).
			Msg("Failed to parse request body")
		response.BadRequest(c, "Invalid request body", requestID)
		return
	}

	// Validate input
	if err := plugin.ValidateInput(input); err != nil {
		log.Warn().
			Str("request_id", requestID).
			Str("plugin", pluginName).
			Err(err).
			Msg("Plugin input validation failed")
		response.BadRequest(c, err.Error(), requestID)
		return
	}

	// Execute plugin
	result, err := plugin.Execute(c.Request.Context(), input)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("plugin", pluginName).
			Err(err).
			Msg("Plugin execution failed")
		response.InternalError(c, "Plugin execution failed", requestID)
		return
	}

	c.JSON(http.StatusOK, result)
}
