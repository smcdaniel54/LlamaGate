package proxy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/llamagate/llamagate/internal/cache"
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
			// Encode to buffer first to check for errors before writing headers
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(response); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"failed to encode response"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(buf.Bytes())
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

func TestProxy_HandleChatCompletions_Success(t *testing.T) {
	// Create a mock Ollama server
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			response := map[string]interface{}{
				"model":   req["model"],
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Hello! How can I help you?",
				},
				"done": true,
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
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		proxyInstance.HandleChatCompletions(c)
	})

	// Make request
	reqBody := map[string]interface{}{
		"model": "llama2",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
}

func TestProxy_HandleChatCompletions_MissingMessages(t *testing.T) {
	cacheInstance := cache.New()
	proxyInstance := New("http://localhost:11434", cacheInstance)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		proxyInstance.HandleChatCompletions(c)
	})

	reqBody := map[string]interface{}{
		"model": "llama2",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "error")
}

func TestProxy_HandleChatCompletions_InvalidJSON(t *testing.T) {
	cacheInstance := cache.New()
	proxyInstance := New("http://localhost:11434", cacheInstance)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		proxyInstance.HandleChatCompletions(c)
	})

	req := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProxy_HandleChatCompletions_WithTemperature(t *testing.T) {
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			// Verify temperature is passed
			assert.Contains(t, req, "options")
			options := req["options"].(map[string]interface{})
			assert.Equal(t, 0.7, options["temperature"])

			response := map[string]interface{}{
				"model":   req["model"],
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Response",
				},
				"done": true,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer mockOllama.Close()

	cacheInstance := cache.New()
	proxyInstance := New(mockOllama.URL, cacheInstance)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		proxyInstance.HandleChatCompletions(c)
	})

	reqBody := map[string]interface{}{
		"model":       "llama2",
		"temperature": 0.7,
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProxy_HandleChatCompletions_CacheHit(t *testing.T) {
	// Create a mock Ollama server that returns a specific response
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"model":"llama2","message":{"role":"assistant","content":"First response"}}`))
		}
	}))
	defer mockOllama.Close()

	cacheInstance := cache.New()
	proxyInstance := New(mockOllama.URL, cacheInstance)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		proxyInstance.HandleChatCompletions(c)
	})

	reqBody := map[string]interface{}{
		"model": "llama2",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// First request - should hit Ollama and cache the response
	req1 := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)
	firstResponse := w1.Body.Bytes()

	// Second request - should hit cache (mock server won't be called)
	// Update mock to fail if called (proving cache was used)
	mockOllama.Close() // Close the server to prove cache is used

	req2 := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Should still succeed from cache
	assert.Equal(t, http.StatusOK, w2.Code)
	// Response should match the first response (from cache)
	assert.Equal(t, firstResponse, w2.Body.Bytes())
}

func TestProxy_HandleChatCompletions_OllamaError(t *testing.T) {
	// Create a mock Ollama server that returns an error
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Internal server error"}`))
		}
	}))
	defer mockOllama.Close()

	cacheInstance := cache.New()
	proxyInstance := New(mockOllama.URL, cacheInstance)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		proxyInstance.HandleChatCompletions(c)
	})

	reqBody := map[string]interface{}{
		"model": "llama2",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return error from Ollama (proxy forwards the status code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestProxy_HandleModels_OllamaError(t *testing.T) {
	// Create a mock Ollama server that returns an error
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal server error"}`))
	}))
	defer mockOllama.Close()

	cacheInstance := cache.New()
	proxyInstance := New(mockOllama.URL, cacheInstance)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/models", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		proxyInstance.HandleModels(c)
	})

	req := httptest.NewRequest("GET", "/v1/models", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// When Ollama returns 500, proxy converts to empty models list and returns 200
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data, ok := response["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 0) // Empty models list
}

func TestProxy_HandleModels_InvalidResponse(t *testing.T) {
	// Create a mock Ollama server that returns invalid JSON
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`invalid json`))
		}
	}))
	defer mockOllama.Close()

	cacheInstance := cache.New()
	proxyInstance := New(mockOllama.URL, cacheInstance)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/models", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		proxyInstance.HandleModels(c)
	})

	req := httptest.NewRequest("GET", "/v1/models", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)
}