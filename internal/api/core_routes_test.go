package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/cache"
	"github.com/llamagate/llamagate/internal/config"
	"github.com/llamagate/llamagate/internal/middleware"
	"github.com/llamagate/llamagate/internal/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCoreOnly_RemovedExtensionEndpointsReturn404 verifies that when the server
// is run without extension routes (Phase 1 core-only), requests to /v1/extensions
// return 404.
func TestCoreOnly_RemovedExtensionEndpointsReturn404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	cfg := &config.Config{OllamaHost: "http://localhost:11434", HealthCheckTimeout: 5 * time.Second}
	router.GET("/health", NewHealthHandler(cfg).CheckHealth)
	router.GET("/v1/hardware/recommendations", NewHardwareHandler().GetRecommendations)

	v1 := router.Group("/v1")
	{
		cacheInstance := cache.New()
		proxyInstance := proxy.NewWithTimeout("http://localhost:11434", cacheInstance, 5*time.Minute)
		v1.POST("/chat/completions", proxyInstance.HandleChatCompletions)
		v1.GET("/models", proxyInstance.HandleModels)
		// No /extensions routes registered (core-only)
	}

	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/v1/extensions"},
		{"GET", "/v1/extensions/any-name"},
		{"PUT", "/v1/extensions/any-name"},
		{"POST", "/v1/extensions/any-name/execute"},
		{"POST", "/v1/extensions/refresh"},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			var req *http.Request
			var err error
			if tt.method == "GET" {
				req, err = http.NewRequest(tt.method, tt.path, nil)
			} else {
				req, err = http.NewRequest(tt.method, tt.path, http.NoBody)
			}
			require.NoError(t, err)
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code, "removed extension endpoint should return 404")
		})
	}
}

// TestCoreOnly_HealthEndpointReachable verifies that the core health endpoint is reachable.
func TestCoreOnly_HealthEndpointReachable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	cfg := &config.Config{OllamaHost: "http://127.0.0.1:1", HealthCheckTimeout: 10 * time.Millisecond}
	router.GET("/health", NewHealthHandler(cfg).CheckHealth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)
	// Health may be 200 (if something is on port 1) or 503 (Ollama unreachable)
	assert.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, w.Code)
}
