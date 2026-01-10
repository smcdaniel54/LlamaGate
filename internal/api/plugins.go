package api

import (
	"fmt"
	"net/http"
	"time"

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

	// Get plugin context for enhanced logging
	pluginCtx := h.registry.GetContext(pluginName)
	
	// Validate input
	if err := plugin.ValidateInput(input); err != nil {
		// Use plugin context logger if available
		if pluginCtx != nil {
			pluginCtx.LogWarn().
				Str("request_id", requestID).
				Err(err).
				Msg("Plugin input validation failed")
		} else {
			log.Warn().
				Str("request_id", requestID).
				Str("plugin", pluginName).
				Err(err).
				Msg("Plugin input validation failed")
		}
		response.BadRequest(c, err.Error(), requestID)
		return
	}

	// Execute plugin with timing
	startTime := time.Now()
	result, err := plugin.Execute(c.Request.Context(), input)
	executionTime := time.Since(startTime)
	
	if err != nil {
		// Use plugin context logger if available, otherwise use default logger
		if pluginCtx != nil {
			pluginCtx.LogError(err).
				Str("request_id", requestID).
				Dur("execution_time", executionTime).
				Msg("Plugin execution failed")
		} else {
			log.Error().
				Str("request_id", requestID).
				Str("plugin", pluginName).
				Dur("execution_time", executionTime).
				Err(err).
				Msg("Plugin execution failed")
		}
		
		// Return structured error response
		response.InternalError(c, fmt.Sprintf("Plugin execution failed: %v", err), requestID)
		return
	}
	
	// Log successful execution with timing
	if pluginCtx != nil {
		pluginCtx.LogInfo().
			Str("request_id", requestID).
			Dur("execution_time", executionTime).
			Msg("Plugin executed successfully")
	} else {
		log.Info().
			Str("request_id", requestID).
			Str("plugin", pluginName).
			Dur("execution_time", executionTime).
			Msg("Plugin executed successfully")
	}

	c.JSON(http.StatusOK, result)
}
