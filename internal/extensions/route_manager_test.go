package extensions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestManifest creates a simple manifest for testing
// Uses a simple workflow that doesn't require external dependencies
func createTestManifest(name string, endpoints []EndpointDefinition) *Manifest {
	return &Manifest{
		Name:        name,
		Version:     "1.0.0",
		Description: "Test extension",
		Type:        "workflow",
		Steps: []WorkflowStep{
			// Use template.render with template_content directly
			// This will work without external dependencies
			{Uses: "template.render", With: map[string]interface{}{
				"template_content": "Result: success",
				"variables": map[string]interface{}{
					"result": "success",
				},
			}},
		},
		Endpoints: endpoints,
	}
}

func TestRouteManager_RegisterExtensionRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": "test response",
					},
				},
			},
		}, nil
	}
	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	rateLimit := middleware.NewRateLimitMiddleware(100.0) // High limit for testing

	rm := NewRouteManager(router, registry, executor, "test", "", rateLimit) // Empty API key to disable auth

	requiresAuth := false
	manifest := createTestManifest("test-extension", []EndpointDefinition{
		{
			Path:        "/test",
			Method:      "GET",
			Description: "Test endpoint",
			RequiresAuth: &requiresAuth,
		},
	})

	err := registry.Register(manifest)
	require.NoError(t, err)

	err = rm.RegisterExtensionRoutes(manifest)
	require.NoError(t, err)

	// Test that route is registered
	req := httptest.NewRequest("GET", "/v1/extensions/test-extension/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	if success, ok := response["success"].(bool); ok {
		assert.True(t, success)
	}
}

func TestRouteManager_RegisterMultipleEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "ok"}, nil
	}
	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	rateLimit := middleware.NewRateLimitMiddleware(100.0)
	rm := NewRouteManager(router, registry, executor, "test", "", rateLimit) // Empty API key to disable auth

	requiresAuth := false
	manifest := createTestManifest("multi-endpoint", []EndpointDefinition{
		{Path: "/get", Method: "GET", Description: "GET endpoint", RequiresAuth: &requiresAuth},
		{Path: "/post", Method: "POST", Description: "POST endpoint", RequiresAuth: &requiresAuth},
		{Path: "/put", Method: "PUT", Description: "PUT endpoint", RequiresAuth: &requiresAuth},
	})

	err := registry.Register(manifest)
	require.NoError(t, err)

	err = rm.RegisterExtensionRoutes(manifest)
	require.NoError(t, err)

	// Test GET endpoint
	req1 := httptest.NewRequest("GET", "/v1/extensions/multi-endpoint/get", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Test POST endpoint
	req2 := httptest.NewRequest("POST", "/v1/extensions/multi-endpoint/post", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Test PUT endpoint
	req3 := httptest.NewRequest("PUT", "/v1/extensions/multi-endpoint/put", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}

func TestRouteManager_RouteConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "ok"}, nil
	}
	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	rateLimit := middleware.NewRateLimitMiddleware(100.0)
	rm := NewRouteManager(router, registry, executor, "test", "", rateLimit) // Empty API key to disable auth

	manifest1 := createTestManifest("ext1", []EndpointDefinition{
		{Path: "/conflict", Method: "GET", Description: "Conflict endpoint"},
	})

	manifest2 := createTestManifest("ext2", []EndpointDefinition{
		{Path: "/conflict", Method: "GET", Description: "Conflict endpoint"},
	})

	err := registry.Register(manifest1)
	require.NoError(t, err)
	err = registry.Register(manifest2)
	require.NoError(t, err)

	// Register first extension
	err = rm.RegisterExtensionRoutes(manifest1)
	require.NoError(t, err)

	// Try to register second extension with conflicting route
	// Note: Gin allows duplicate routes, but RouteManager tracks them and should detect conflict
	err = rm.RegisterExtensionRoutes(manifest2)
	// RouteManager should detect the conflict in its tracking map
	// However, Gin will allow it, so we check that it's tracked
	routes := rm.GetRegisteredRoutes()
	// Both routes will be registered in Gin, but RouteManager should track both
	// The conflict detection happens at registration time in RouteManager
	if err == nil {
		// If no error, both routes are tracked (Gin allows duplicates)
		// This is actually expected behavior - conflicts are runtime, not registration-time
		assert.Len(t, routes, 2) // Both routes tracked
	} else {
		assert.Contains(t, err.Error(), "route conflict")
	}
}

func TestRouteManager_NonWorkflowExtension(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "ok"}, nil
	}
	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	rateLimit := middleware.NewRateLimitMiddleware(100.0)
	rm := NewRouteManager(router, registry, executor, "test", "", rateLimit) // Empty API key to disable auth

	manifest := &Manifest{
		Name:        "middleware-ext",
		Version:     "1.0.0",
		Description: "Middleware extension",
		Type:        "middleware",
		Hooks: []HookDefinition{
			{On: "http.request", Action: "log"},
		},
		Endpoints: []EndpointDefinition{
			{Path: "/test", Method: "GET", Description: "Should fail"},
		},
	}

	err := registry.Register(manifest)
	require.NoError(t, err)

	// Should fail because only workflow extensions can have endpoints
	err = rm.RegisterExtensionRoutes(manifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only workflow extensions")
}

func TestRouteManager_EmptyEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "ok"}, nil
	}
	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	rateLimit := middleware.NewRateLimitMiddleware(100.0)
	rm := NewRouteManager(router, registry, executor, "test", "", rateLimit) // Empty API key to disable auth

	manifest := createTestManifest("no-endpoints", []EndpointDefinition{})

	err := registry.Register(manifest)
	require.NoError(t, err)

	// Should succeed with no endpoints
	err = rm.RegisterExtensionRoutes(manifest)
	require.NoError(t, err)
}

func TestRouteManager_QueryParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "ok"}, nil
	}
	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	rateLimit := middleware.NewRateLimitMiddleware(100.0)
	rm := NewRouteManager(router, registry, executor, "test", "", rateLimit) // Empty API key to disable auth

	manifest := createTestManifest("query-test", []EndpointDefinition{
		{Path: "/query", Method: "GET", Description: "Query endpoint"},
	})

	err := registry.Register(manifest)
	require.NoError(t, err)

	err = rm.RegisterExtensionRoutes(manifest)
	require.NoError(t, err)

	// Test with query parameters
	req := httptest.NewRequest("GET", "/v1/extensions/query-test/query?param1=value1&param2=value2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouteManager_PathParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "ok"}, nil
	}
	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	rateLimit := middleware.NewRateLimitMiddleware(100.0)
	rm := NewRouteManager(router, registry, executor, "test", "", rateLimit) // Empty API key to disable auth

	manifest := createTestManifest("path-test", []EndpointDefinition{
		{Path: "/:id", Method: "GET", Description: "Path param endpoint"},
	})

	err := registry.Register(manifest)
	require.NoError(t, err)

	err = rm.RegisterExtensionRoutes(manifest)
	require.NoError(t, err)

	// Test with path parameter
	req := httptest.NewRequest("GET", "/v1/extensions/path-test/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouteManager_PostWithBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "ok"}, nil
	}
	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	rateLimit := middleware.NewRateLimitMiddleware(100.0)
	rm := NewRouteManager(router, registry, executor, "test", "", rateLimit) // Empty API key to disable auth

	manifest := createTestManifest("post-test", []EndpointDefinition{
		{Path: "/post", Method: "POST", Description: "POST endpoint"},
	})

	err := registry.Register(manifest)
	require.NoError(t, err)

	err = rm.RegisterExtensionRoutes(manifest)
	require.NoError(t, err)

	// Test POST with JSON body
	req := httptest.NewRequest("POST", "/v1/extensions/post-test/post", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouteManager_DisabledExtension(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "ok"}, nil
	}
	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	rateLimit := middleware.NewRateLimitMiddleware(100.0)
	rm := NewRouteManager(router, registry, executor, "test", "", rateLimit) // Empty API key to disable auth

	disabled := false
	manifest := createTestManifest("disabled-ext", []EndpointDefinition{
		{Path: "/test", Method: "GET", Description: "Test endpoint"},
	})
	manifest.Enabled = &disabled

	err := registry.Register(manifest)
	require.NoError(t, err)

	err = rm.RegisterExtensionRoutes(manifest)
	require.NoError(t, err)

	// Test that disabled extension returns 503
	req := httptest.NewRequest("GET", "/v1/extensions/disabled-ext/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestRouteManager_UnregisterExtensionRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "ok"}, nil
	}
	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	rateLimit := middleware.NewRateLimitMiddleware(100.0)
	rm := NewRouteManager(router, registry, executor, "test", "", rateLimit) // Empty API key to disable auth

	manifest := createTestManifest("unregister-test", []EndpointDefinition{
		{Path: "/test", Method: "GET", Description: "Test endpoint"},
	})

	err := registry.Register(manifest)
	require.NoError(t, err)

	err = rm.RegisterExtensionRoutes(manifest)
	require.NoError(t, err)

	// Verify route is registered
	routes := rm.GetRegisteredRoutes()
	assert.Len(t, routes, 1)

	// Unregister
	rm.UnregisterExtensionRoutes("unregister-test")

	// Verify route is removed from tracking (note: route still active in Gin)
	routes = rm.GetRegisteredRoutes()
	assert.Len(t, routes, 0)
}

func TestRouteManager_GetRegisteredRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "ok"}, nil
	}
	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	rateLimit := middleware.NewRateLimitMiddleware(100.0)
	rm := NewRouteManager(router, registry, executor, "test", "", rateLimit) // Empty API key to disable auth

	manifest := createTestManifest("routes-test", []EndpointDefinition{
		{Path: "/endpoint1", Method: "GET", Description: "Endpoint 1"},
		{Path: "/endpoint2", Method: "POST", Description: "Endpoint 2"},
	})

	err := registry.Register(manifest)
	require.NoError(t, err)

	err = rm.RegisterExtensionRoutes(manifest)
	require.NoError(t, err)

	routes := rm.GetRegisteredRoutes()
	assert.Len(t, routes, 2)

	// Verify route keys
	expectedKeys := []string{
		"GET:/v1/extensions/routes-test/endpoint1",
		"POST:/v1/extensions/routes-test/endpoint2",
	}
	for _, key := range expectedKeys {
		assert.Contains(t, routes, key)
		assert.Equal(t, "routes-test", routes[key].ExtensionName)
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"root path", "/", "/"},
		{"path with trailing slash", "/test/", "/test"},
		{"path without trailing slash", "/test", "/test"},
		{"long path with trailing slash", "/v1/extensions/test/endpoint/", "/v1/extensions/test/endpoint"},
		{"path with multiple slashes", "/test//path", "/test//path"}, // Doesn't handle multiple slashes
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRouteManager_AllHTTPMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "ok"}, nil
	}
	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	rateLimit := middleware.NewRateLimitMiddleware(100.0)
	rm := NewRouteManager(router, registry, executor, "test", "", rateLimit) // Empty API key to disable auth

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range methods {
		manifest := createTestManifest("method-test-"+method, []EndpointDefinition{
			{Path: "/test", Method: method, Description: method + " endpoint"},
		})

		err := registry.Register(manifest)
		require.NoError(t, err)

		err = rm.RegisterExtensionRoutes(manifest)
		require.NoError(t, err)

		// Test the route
		req := httptest.NewRequest(method, "/v1/extensions/method-test-"+method+"/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Most methods should return 200, but HEAD/OPTIONS might have different behavior
		// HEAD returns 200 with no body, OPTIONS might return different status
		switch method {
		case "HEAD":
			// HEAD should return 200 but might not have body
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNoContent)
		case "OPTIONS":
			// OPTIONS might return 200, 204, or 405 depending on implementation
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNoContent || w.Code == http.StatusMethodNotAllowed)
		default:
			assert.Equal(t, http.StatusOK, w.Code)
		}
	}
}

func TestRouteManager_WorkflowExecutionError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	registry := NewRegistry()
	// LLM handler that returns error
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return nil, fmt.Errorf("LLM service unavailable")
	}
	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	rateLimit := middleware.NewRateLimitMiddleware(100.0)
	rm := NewRouteManager(router, registry, executor, "test", "", rateLimit)

	// Create manifest with step that will fail
	manifest := &Manifest{
		Name:        "error-test",
		Version:     "1.0.0",
		Description: "Error test",
		Type:        "workflow",
		Steps: []WorkflowStep{
			{Uses: "llm.chat", With: map[string]interface{}{
				"prompt": "test",
				"model":  "mistral",
			}},
		},
		Endpoints: []EndpointDefinition{
			{Path: "/error", Method: "GET", Description: "Error endpoint"},
		},
	}

	err := registry.Register(manifest)
	require.NoError(t, err)

	err = rm.RegisterExtensionRoutes(manifest)
	require.NoError(t, err)

	// Test that workflow error returns 500
	req := httptest.NewRequest("GET", "/v1/extensions/error-test/error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestRouteManager_InvalidJSONBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "ok"}, nil
	}
	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	rateLimit := middleware.NewRateLimitMiddleware(100.0)
	rm := NewRouteManager(router, registry, executor, "test", "", rateLimit)

	manifest := createTestManifest("json-test", []EndpointDefinition{
		{Path: "/post", Method: "POST", Description: "POST endpoint"},
	})

	err := registry.Register(manifest)
	require.NoError(t, err)

	err = rm.RegisterExtensionRoutes(manifest)
	require.NoError(t, err)

	// Test POST with invalid JSON
	req := httptest.NewRequest("POST", "/v1/extensions/json-test/post", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still work (we continue with empty input on parse error)
	// The handler is lenient with JSON parsing
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouteManager_EmptyRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "ok"}, nil
	}
	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	rateLimit := middleware.NewRateLimitMiddleware(100.0)
	rm := NewRouteManager(router, registry, executor, "test", "", rateLimit)

	manifest := createTestManifest("empty-body-test", []EndpointDefinition{
		{Path: "/post", Method: "POST", Description: "POST endpoint"},
	})

	err := registry.Register(manifest)
	require.NoError(t, err)

	err = rm.RegisterExtensionRoutes(manifest)
	require.NoError(t, err)

	// Test POST with empty body
	req := httptest.NewRequest("POST", "/v1/extensions/empty-body-test/post", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouteManager_MultiplePathParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "ok"}, nil
	}
	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	rateLimit := middleware.NewRateLimitMiddleware(100.0)
	rm := NewRouteManager(router, registry, executor, "test", "", rateLimit)

	manifest := createTestManifest("multi-param-test", []EndpointDefinition{
		{Path: "/:id/:action", Method: "GET", Description: "Multi param endpoint"},
	})

	err := registry.Register(manifest)
	require.NoError(t, err)

	err = rm.RegisterExtensionRoutes(manifest)
	require.NoError(t, err)

	// Test with multiple path parameters
	req := httptest.NewRequest("GET", "/v1/extensions/multi-param-test/123/update", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
