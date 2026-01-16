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
	registry        *Registry
	workflowExecutor *WorkflowExecutor
}

// NewHandler creates a new extension handler
func NewHandler(registry *Registry, llmHandler LLMHandlerFunc, baseDir string) *Handler {
	executor := NewWorkflowExecutor(llmHandler, baseDir)
	executor.SetRegistry(registry) // Enable extension-to-extension calls
	return &Handler{
		registry:        registry,
		workflowExecutor: executor,
	}
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
