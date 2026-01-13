//nolint:revive // "api" is a meaningful package name for API handlers
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/plugins"
	"github.com/rs/zerolog/log"
)

// RegisterPluginRoutes registers custom API endpoints for plugins
func RegisterPluginRoutes(router *gin.RouterGroup, registry *plugins.Registry) {
	if registry == nil {
		return
	}

	// Get all plugins
	pluginList := registry.List()

	// Register custom endpoints for each plugin
	for _, metadata := range pluginList {
		plugin, err := registry.Get(metadata.Name)
		if err != nil {
			log.Warn().
				Str("plugin", metadata.Name).
				Err(err).
				Msg("Failed to get plugin for route registration")
			continue
		}

		// Check if plugin implements ExtendedPlugin
		extendedPlugin, ok := plugin.(plugins.ExtendedPlugin)
		if !ok {
			continue // Plugin doesn't expose custom endpoints
		}

		// Get API endpoints
		endpoints := extendedPlugin.GetAPIEndpoints()
		if len(endpoints) == 0 {
			continue
		}

		// Create plugin route group
		pluginGroup := router.Group("/plugins/" + metadata.Name)

		// Register each endpoint
		for _, endpoint := range endpoints {
			registerEndpoint(pluginGroup, endpoint, metadata.Name)
		}

		log.Info().
			Str("plugin", metadata.Name).
			Int("endpoints", len(endpoints)).
			Msg("Registered plugin API endpoints")
	}
}

// registerEndpoint registers a single plugin endpoint
func registerEndpoint(group *gin.RouterGroup, endpoint plugins.APIEndpoint, pluginName string) {
	// Build full path
	fullPath := endpoint.Path
	if fullPath[0] != '/' {
		fullPath = "/" + fullPath
	}

	// Register based on method
	switch endpoint.Method {
	case "GET":
		group.GET(fullPath, endpoint.Handler)
	case "POST":
		group.POST(fullPath, endpoint.Handler)
	case "PUT":
		group.PUT(fullPath, endpoint.Handler)
	case "DELETE":
		group.DELETE(fullPath, endpoint.Handler)
	case "PATCH":
		group.PATCH(fullPath, endpoint.Handler)
	default:
		log.Warn().
			Str("plugin", pluginName).
			Str("method", endpoint.Method).
			Str("path", fullPath).
			Msg("Unsupported HTTP method for plugin endpoint")
		return
	}

	log.Debug().
		Str("plugin", pluginName).
		Str("method", endpoint.Method).
		Str("path", fullPath).
		Msg("Registered plugin endpoint")
}
