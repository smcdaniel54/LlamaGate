// Package extensions provides the extension system for LlamaGate.
// Extensions are declarative, YAML-based modules that extend LlamaGate
// functionality through workflows, middleware hooks, and observer hooks.
package extensions

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/middleware"
	"github.com/llamagate/llamagate/internal/response"
	"github.com/rs/zerolog/log"
)

// Handler handles extension-related HTTP requests
type Handler struct {
	registry         *Registry
	workflowExecutor *WorkflowExecutor
	baseDir          string
	routeManager     *RouteManager // Can be nil if not set
}

// NewHandler creates a new extension handler
func NewHandler(registry *Registry, llmHandler LLMHandlerFunc, baseDir string) *Handler {
	executor := NewWorkflowExecutor(llmHandler, baseDir)
	executor.SetRegistry(registry) // Enable extension-to-extension calls
	return &Handler{
		registry:         registry,
		workflowExecutor: executor,
		baseDir:          baseDir,
		routeManager:     nil, // Set via SetRouteManager
	}
}

// SetRouteManager sets the route manager for refreshing routes
func (h *Handler) SetRouteManager(rm *RouteManager) {
	h.routeManager = rm
}

// ListExtensions lists all registered extensions
// GET /v1/extensions
func (h *Handler) ListExtensions(c *gin.Context) {
	_ = middleware.GetRequestID(c) // For logging context

	manifests := h.registry.List()

	// Convert to response format
	extensions := make([]map[string]interface{}, 0, len(manifests))
	for _, manifest := range manifests {
		extensions = append(extensions, map[string]interface{}{
			"name":        manifest.Name,
			"version":     manifest.Version,
			"description": manifest.Description,
			"type":        manifest.Type,
			"enabled":     h.registry.IsEnabled(manifest.Name),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"extensions": extensions,
		"count":      len(extensions),
	})
}

// GetExtension gets information about a specific extension
// GET /v1/extensions/:name
func (h *Handler) GetExtension(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	extensionName := c.Param("name")

	manifest, err := h.registry.Get(extensionName)
	if err != nil {
		response.NotFound(c, "Extension not found", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":        manifest.Name,
		"version":     manifest.Version,
		"description": manifest.Description,
		"type":        manifest.Type,
		"enabled":     h.registry.IsEnabled(manifest.Name),
		"inputs":      manifest.Inputs,
		"outputs":     manifest.Outputs,
	})
}

// ExecuteExtension executes a workflow extension
// POST /v1/extensions/:name/execute
func (h *Handler) ExecuteExtension(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	extensionName := c.Param("name")

	manifest, err := h.registry.Get(extensionName)
	if err != nil {
		response.NotFound(c, "Extension not found", requestID)
		return
	}

	// Check if enabled
	if !h.registry.IsEnabled(manifest.Name) {
		response.ServiceUnavailable(c, "Extension is disabled", requestID)
		return
	}

	// Only workflow extensions can be executed via API
	if manifest.Type != "workflow" {
		response.BadRequest(c, "Only workflow extensions can be executed via API", requestID)
		return
	}

	// Parse request body
	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("extension", extensionName).
			Err(err).
			Msg("Failed to parse request body")
		response.BadRequest(c, "Invalid request body", requestID)
		return
	}

	// Validate required inputs
	for _, inputDef := range manifest.Inputs {
		if inputDef.Required {
			if _, exists := input[inputDef.ID]; !exists {
				response.BadRequest(c, fmt.Sprintf("Required input '%s' is missing", inputDef.ID), requestID)
				return
			}
		}
	}

	// Create execution context with guardrails
	execCtx := NewExecutionContext(c.Request.Context(), requestID, GetExtensionDir("extensions", extensionName))

	// Execute workflow
	result, err := h.workflowExecutor.Execute(execCtx, manifest, input)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("extension", extensionName).
			Err(err).
			Msg("Extension execution failed")
		response.InternalError(c, fmt.Sprintf("Extension execution failed: %v", err), requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// RefreshExtensions re-discovers extensions from the directory and updates the registry
// POST /v1/extensions/refresh
func (h *Handler) RefreshExtensions(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	
	// Ensure we always send a response, even if something goes wrong
	defer func() {
		if r := recover(); r != nil {
			log.Error().
				Str("request_id", requestID).
				Interface("panic", r).
				Msg("Panic in RefreshExtensions handler")
			// Only send error if response hasn't been written yet
			if !c.Writer.Written() {
				response.InternalError(c, fmt.Sprintf("Internal error during refresh: %v", r), requestID)
			}
		}
	}()

	// Get current extensions before refresh
	currentManifests := h.registry.List()
	currentNames := make(map[string]bool)
	for _, manifest := range currentManifests {
		currentNames[manifest.Name] = true
	}

	// Re-discover extensions with panic recovery
	var manifests []*Manifest
	var discoverErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Str("request_id", requestID).
					Interface("panic", r).
					Msg("Panic during extension discovery")
				discoverErr = fmt.Errorf("panic during discovery: %v", r)
			}
		}()
		manifests, discoverErr = DiscoverExtensions(h.baseDir)
	}()
	
	if discoverErr != nil {
		log.Warn().
			Str("request_id", requestID).
			Err(discoverErr).
			Msg("Failed to discover extensions during refresh")
		response.InternalError(c, fmt.Sprintf("Failed to discover extensions: %v", discoverErr), requestID)
		return
	}

	// Track changes
	added := []string{}
	updated := []string{}
	removed := []string{}
	errors := []string{}

	// Build map of discovered extensions
	discoveredNames := make(map[string]bool)
	discoveredManifests := make(map[string]*Manifest)
	for _, manifest := range manifests {
		discoveredNames[manifest.Name] = true
		discoveredManifests[manifest.Name] = manifest
	}

	// Add new extensions and update existing ones
	for _, manifest := range manifests {
		if _, exists := currentNames[manifest.Name]; exists {
			// Extension exists - update it
			if err := h.registry.RegisterOrUpdate(manifest); err != nil {
				log.Warn().
					Str("request_id", requestID).
					Str("extension", manifest.Name).
					Err(err).
					Msg("Failed to update extension during refresh")
				errors = append(errors, fmt.Sprintf("%s: %v", manifest.Name, err))
			} else {
				updated = append(updated, manifest.Name)
				log.Info().
					Str("request_id", requestID).
					Str("extension", manifest.Name).
					Str("version", manifest.Version).
					Msg("Extension updated during refresh")

				// Update routes if route manager is available
				if h.routeManager != nil {
					// Unregister old routes first
					h.routeManager.UnregisterExtensionRoutes(manifest.Name)
					// Register new routes
					if len(manifest.Endpoints) > 0 {
						// Use recover to catch any panics during route registration
						func() {
							defer func() {
								if r := recover(); r != nil {
									log.Error().
										Str("request_id", requestID).
										Str("extension", manifest.Name).
										Interface("panic", r).
										Msg("Panic during route registration, continuing with refresh")
									errors = append(errors, fmt.Sprintf("%s (routes): panic during registration: %v", manifest.Name, r))
								}
							}()
							if err := h.routeManager.RegisterExtensionRoutes(manifest); err != nil {
								log.Warn().
									Str("request_id", requestID).
									Str("extension", manifest.Name).
									Err(err).
									Msg("Failed to register extension routes during refresh")
								errors = append(errors, fmt.Sprintf("%s (routes): %v", manifest.Name, err))
							}
						}()
					}
				}
			}
		} else {
			// New extension - register it
			if err := h.registry.RegisterOrUpdate(manifest); err != nil {
				log.Warn().
					Str("request_id", requestID).
					Str("extension", manifest.Name).
					Err(err).
					Msg("Failed to register extension during refresh")
				errors = append(errors, fmt.Sprintf("%s: %v", manifest.Name, err))
			} else {
				added = append(added, manifest.Name)
				log.Info().
					Str("request_id", requestID).
					Str("extension", manifest.Name).
					Str("version", manifest.Version).
					Str("type", manifest.Type).
					Bool("enabled", manifest.IsEnabled()).
					Msg("Extension registered during refresh")

				// Register routes if route manager is available
				if h.routeManager != nil && len(manifest.Endpoints) > 0 {
					// Use recover to catch any panics during route registration
					func() {
						defer func() {
							if r := recover(); r != nil {
								log.Error().
									Str("request_id", requestID).
									Str("extension", manifest.Name).
									Interface("panic", r).
									Msg("Panic during route registration, continuing with refresh")
								errors = append(errors, fmt.Sprintf("%s (routes): panic during registration: %v", manifest.Name, r))
							}
						}()
						if err := h.routeManager.RegisterExtensionRoutes(manifest); err != nil {
							log.Warn().
								Str("request_id", requestID).
								Str("extension", manifest.Name).
								Err(err).
								Msg("Failed to register extension routes during refresh")
							errors = append(errors, fmt.Sprintf("%s (routes): %v", manifest.Name, err))
						}
					}()
				}
			}
		}
	}

	// Remove extensions that no longer exist
	for name := range currentNames {
		if !discoveredNames[name] {
			if err := h.registry.Unregister(name); err != nil {
				log.Warn().
					Str("request_id", requestID).
					Str("extension", name).
					Err(err).
					Msg("Failed to unregister extension during refresh")
				errors = append(errors, fmt.Sprintf("%s (unregister): %v", name, err))
			} else {
				removed = append(removed, name)
				log.Info().
					Str("request_id", requestID).
					Str("extension", name).
					Msg("Extension removed during refresh")

				// Unregister routes if route manager is available
				if h.routeManager != nil {
					h.routeManager.UnregisterExtensionRoutes(name)
				}
			}
		}
	}

	// Build response
	responseData := gin.H{
		"status":  "refreshed",
		"added":   added,
		"updated": updated,
		"removed": removed,
		"total":   len(manifests),
	}

	if len(errors) > 0 {
		responseData["errors"] = errors
		responseData["status"] = "partial"
		c.JSON(http.StatusPartialContent, responseData)
		return
	}

	log.Info().
		Str("request_id", requestID).
		Int("added", len(added)).
		Int("updated", len(updated)).
		Int("removed", len(removed)).
		Int("total", len(manifests)).
		Msg("Extension refresh complete")

	c.JSON(http.StatusOK, responseData)
}
