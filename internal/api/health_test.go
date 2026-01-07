package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_CheckHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		ollamaStatus   int
		ollamaError    bool
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "healthy",
			ollamaStatus:   http.StatusOK,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"status":      "healthy",
				"ollama":      "connected",
				"ollama_host": "http://localhost:11434",
			},
		},
		{
			name:           "ollama unreachable",
			ollamaError:    true,
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody: map[string]interface{}{
				"status":      "unhealthy",
				"error":       "Ollama unreachable",
				"ollama_host": "http://localhost:11434",
			},
		},
		{
			name:           "ollama non-OK status",
			ollamaStatus:   http.StatusInternalServerError,
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody: map[string]interface{}{
				"status":      "unhealthy",
				"error":       "Ollama returned status 500",
				"ollama_host": "http://localhost:11434",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config
			cfg := &config.Config{
				HealthCheckTimeout: 1 * time.Second, // Short timeout for faster tests
			}

			// Create mock Ollama server
			if !tt.ollamaError {
				ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.ollamaStatus)
					w.Write([]byte(`{"models":[]}`))
				}))
				defer ollamaServer.Close()
				cfg.OllamaHost = ollamaServer.URL
			} else {
				// Use an invalid/unreachable URL to simulate connection error
				cfg.OllamaHost = "http://127.0.0.1:1" // Port 1 is typically not in use
			}

			// Create handler
			handler := NewHealthHandler(cfg)

			// Create request
			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Call handler
			handler.CheckHealth(c)

			// Assert status
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert body
			var body map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &body)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody["status"], body["status"])
			if tt.expectedBody["error"] != nil {
				assert.Equal(t, tt.expectedBody["error"], body["error"])
			}
			if tt.expectedBody["ollama"] != nil {
				assert.Equal(t, tt.expectedBody["ollama"], body["ollama"])
			}
		})
	}
}
