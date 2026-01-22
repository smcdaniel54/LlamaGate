package extensions

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/middleware"
	"github.com/llamagate/llamagate/internal/response"
	"github.com/rs/zerolog/log"
)

const (
	// ExtensionRoutePrefix is the base path for extension endpoints
	ExtensionRoutePrefix = "/v1/extensions"
	// DefaultRequiresAuth is the default value for requires_auth
	DefaultRequiresAuth = true
	// DefaultRequiresRateLimit is the default value for requires_rate_limit
	DefaultRequiresRateLimit = true
)

// RouteManager manages dynamic routes for extensions
type RouteManager struct {
	mu        sync.RWMutex
	router    *gin.Engine
	registry  *Registry
	executor  *WorkflowExecutor
	routes    map[string]*RouteInfo // Track registered routes
	baseDir   string
	apiKey    string
	rateLimit *middleware.RateLimitMiddleware
}

// RouteInfo tracks registered route information
type RouteInfo struct {
	ExtensionName string
	Endpoint      EndpointDefinition
	Handler       gin.HandlerFunc
}

// NewRouteManager creates a new route manager
func NewRouteManager(
	router *gin.Engine,
	registry *Registry,
	executor *WorkflowExecutor,
	baseDir string,
	apiKey string,
	rateLimit *middleware.RateLimitMiddleware,
) *RouteManager {
	return &RouteManager{
		router:    router,
		registry:  registry,
		executor:  executor,
		routes:    make(map[string]*RouteInfo),
		baseDir:   baseDir,
		apiKey:    apiKey,
		rateLimit: rateLimit,
	}
}

// RegisterExtensionRoutes registers all endpoints for an extension
func (rm *RouteManager) RegisterExtensionRoutes(manifest *Manifest) error {
	if len(manifest.Endpoints) == 0 {
		return nil // No endpoints to register
	}

	// Only workflow extensions can have endpoints (validated in manifest, but double-check)
	if manifest.Type != "workflow" {
		return fmt.Errorf("only workflow extensions can define endpoints, extension '%s' is type '%s'", manifest.Name, manifest.Type)
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	for _, endpoint := range manifest.Endpoints {
		if err := rm.registerRoute(manifest, endpoint); err != nil {
			return fmt.Errorf("failed to register route for extension %s: %w", manifest.Name, err)
		}
	}

	return nil
}

// registerRoute registers a single endpoint route
func (rm *RouteManager) registerRoute(manifest *Manifest, endpoint EndpointDefinition) error {
	// Build full path: /v1/extensions/{name}{endpoint.path}
	fullPath := fmt.Sprintf("%s/%s%s", ExtensionRoutePrefix, manifest.Name, endpoint.Path)

	// Normalize path (remove trailing slashes, etc.)
	fullPath = normalizePath(fullPath)

	// Create route key for tracking
	method := strings.ToUpper(endpoint.Method)
	routeKey := fmt.Sprintf("%s:%s", method, fullPath)

	// Check for conflicts
	if existing, exists := rm.routes[routeKey]; exists {
		return fmt.Errorf("route conflict: %s already registered by extension %s", routeKey, existing.ExtensionName)
	}

	// Build handler chain
	handlers := rm.buildHandlerChain(manifest, endpoint)

	// Register with router based on method
	switch method {
	case "GET":
		rm.router.GET(fullPath, handlers...)
	case "POST":
		rm.router.POST(fullPath, handlers...)
	case "PUT":
		rm.router.PUT(fullPath, handlers...)
	case "DELETE":
		rm.router.DELETE(fullPath, handlers...)
	case "PATCH":
		rm.router.PATCH(fullPath, handlers...)
	case "HEAD":
		rm.router.HEAD(fullPath, handlers...)
	case "OPTIONS":
		rm.router.OPTIONS(fullPath, handlers...)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}

	// Track route
	rm.routes[routeKey] = &RouteInfo{
		ExtensionName: manifest.Name,
		Endpoint:      endpoint,
		Handler:       handlers[len(handlers)-1], // Last handler is the actual endpoint handler
	}

	log.Info().
		Str("extension", manifest.Name).
		Str("method", method).
		Str("path", fullPath).
		Msg("Registered extension endpoint")

	return nil
}

// buildHandlerChain builds the middleware chain for an endpoint
func (rm *RouteManager) buildHandlerChain(manifest *Manifest, endpoint EndpointDefinition) []gin.HandlerFunc {
	var handlers []gin.HandlerFunc

	// Determine if auth is required (default: true)
	requiresAuth := DefaultRequiresAuth
	if endpoint.RequiresAuth != nil {
		requiresAuth = *endpoint.RequiresAuth
	}

	// Apply auth middleware if required
	if requiresAuth && rm.apiKey != "" {
		handlers = append(handlers, middleware.AuthMiddleware(rm.apiKey))
	}

	// Determine if rate limiting is required (default: true)
	requiresRateLimit := DefaultRequiresRateLimit
	if endpoint.RequiresRateLimit != nil {
		requiresRateLimit = *endpoint.RequiresRateLimit
	}

	// Apply rate limiting if required
	if requiresRateLimit && rm.rateLimit != nil {
		handlers = append(handlers, rm.rateLimit.Handler())
	}

	// Add the actual endpoint handler (always last)
	handlers = append(handlers, rm.createEndpointHandler(manifest, endpoint))

	return handlers
}

// createEndpointHandler creates a Gin handler for an extension endpoint
func (rm *RouteManager) createEndpointHandler(manifest *Manifest, endpoint EndpointDefinition) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := middleware.GetRequestID(c)

		// Check if extension is enabled
		if !rm.registry.IsEnabled(manifest.Name) {
			response.ServiceUnavailable(c, "Extension is disabled", requestID)
			return
		}

		// Only workflow extensions can have custom endpoints (double-check)
		if manifest.Type != "workflow" {
			response.BadRequest(c, "Only workflow extensions can define custom endpoints", requestID)
			return
		}

		// Initialize input map
		input := make(map[string]interface{})

		// Parse request body (for POST, PUT, PATCH)
		if endpoint.Method == "POST" || endpoint.Method == "PUT" || endpoint.Method == "PATCH" {
			// Try to parse JSON body, but don't fail if body is empty
			if c.Request.ContentLength > 0 {
				if err := c.ShouldBindJSON(&input); err != nil {
					// Only fail if there's actual content but it's invalid
					if err.Error() != "EOF" {
						log.Debug().
							Str("request_id", requestID).
							Str("extension", manifest.Name).
							Err(err).
							Msg("Failed to parse request body, continuing with empty input")
						// Continue with empty input rather than failing
					}
				}
			}
		}

		// Add query parameters (for all methods, but especially GET, DELETE)
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				// If key already exists in input (from body), prefer body value
				if _, exists := input[key]; !exists {
					input[key] = values[0] // Take first value
				}
			}
		}

		// Add path parameters (Gin path params like :name, :id, etc.)
		for _, param := range c.Params {
			input[param.Key] = param.Value
		}

		// Create execution context
		execCtx := NewExecutionContext(c.Request.Context(), requestID, GetExtensionDir(rm.baseDir, manifest.Name))

		// Execute workflow
		result, err := rm.executor.Execute(execCtx, manifest, input)
		if err != nil {
			log.Error().
				Str("request_id", requestID).
				Str("extension", manifest.Name).
				Str("endpoint", endpoint.Path).
				Err(err).
				Msg("Extension endpoint execution failed")
			response.InternalError(c, fmt.Sprintf("Extension execution failed: %v", err), requestID)
			return
		}

		// Return response
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	}
}

// normalizePath normalizes a path (removes trailing slashes, etc.)
func normalizePath(path string) string {
	// Remove trailing slash (except root)
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	return path
}

// UnregisterExtensionRoutes removes routes for an extension
// Note: Gin doesn't support route removal at runtime, so this is a no-op for now
// Routes will remain until server restart. This is a known limitation.
func (rm *RouteManager) UnregisterExtensionRoutes(extensionName string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Remove from tracking map
	keysToDelete := []string{}
	for key, info := range rm.routes {
		if info.ExtensionName == extensionName {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(rm.routes, key)
		log.Info().
			Str("extension", extensionName).
			Str("route_key", key).
			Msg("Unregistered extension route (note: route remains active until server restart)")
	}
}

// GetRegisteredRoutes returns all registered routes (for debugging/testing)
func (rm *RouteManager) GetRegisteredRoutes() map[string]*RouteInfo {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	routes := make(map[string]*RouteInfo)
	for k, v := range rm.routes {
		routes[k] = v
	}
	return routes
}
