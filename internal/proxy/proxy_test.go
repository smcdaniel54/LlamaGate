package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/cache"
	"github.com/stretchr/testify/assert"
)

func TestProxy_HandleModels(t *testing.T) {
	// Create a mock Ollama server
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			response := map[string]interface{}{
				"models": []map[string]interface{}{
					{"name": "llama2"},
					{"name": "mistral"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockOllama.Close()

	// Create proxy
	cacheInstance := cache.New()
	proxyInstance := New(mockOllama.URL, cacheInstance)

	// Setup Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/models", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		proxyInstance.HandleModels(c)
	})

	// Make request
	req := httptest.NewRequest("GET", "/v1/models", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "list", response["object"])

	data, ok := response["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 2)
}

func TestProxy_HandleChatCompletions_Validation(t *testing.T) {
	// Create proxy
	cacheInstance := cache.New()
	proxyInstance := New("http://localhost:11434", cacheInstance)

	// Setup Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		proxyInstance.HandleChatCompletions(c)
	})

	// Test missing model
	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
