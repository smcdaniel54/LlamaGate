package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/llamagate/llamagate/internal/cache"
)

// TestStreaming_SSEFormat validates SSE format conformance
func TestStreaming_SSEFormat(t *testing.T) {
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)

			chunks := []string{
				`data: {"model":"llama2","message":{"role":"assistant","content":"Hello"},"done":false}` + "\n\n",
				`data: {"model":"llama2","message":{"role":"assistant","content":" world"},"done":false}` + "\n\n",
				`data: {"model":"llama2","message":{"role":"assistant","content":""},"done":true}` + "\n\n",
			}
			for _, chunk := range chunks {
				_, _ = w.Write([]byte(chunk))
			}
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
		"stream": true,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", w.Header().Get("Connection"))

	// Validate SSE format
	body := w.Body.String()
	lines := strings.Split(body, "\n")

	chunkCount := 0
	doneFound := false
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			chunkCount++
			jsonStr := strings.TrimPrefix(line, "data: ")
			if jsonStr == "[DONE]" {
				doneFound = true
				continue
			}
			if jsonStr != "" {
				var chunk map[string]interface{}
				err := json.Unmarshal([]byte(jsonStr), &chunk)
				require.NoError(t, err, "SSE chunk should be valid JSON")

				// Validate OpenAI-compatible structure
				assert.Equal(t, "chat.completion.chunk", chunk["object"])
				assert.Contains(t, chunk, "id")
				assert.Contains(t, chunk, "created")
				assert.Contains(t, chunk, "model")
				assert.Contains(t, chunk, "choices")

				choices, ok := chunk["choices"].([]interface{})
				require.True(t, ok)
				require.Len(t, choices, 1)

				choice, ok := choices[0].(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, choice, "index")
				assert.Contains(t, choice, "delta")
				assert.Contains(t, choice, "finish_reason")

				delta, ok := choice["delta"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "assistant", delta["role"])
			}
		}
	}

	assert.Greater(t, chunkCount, 0, "Should receive at least one chunk")
	assert.True(t, doneFound, "Should receive [DONE] marker")
}

// TestStreaming_OpenAICompatibleChunks validates chunks match OpenAI format
func TestStreaming_OpenAICompatibleChunks(t *testing.T) {
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)

			chunks := []string{
				`data: {"model":"llama2","message":{"role":"assistant","content":"Test"},"done":false}` + "\n\n",
				`data: {"model":"llama2","message":{"role":"assistant","content":""},"done":true}` + "\n\n",
			}
			for _, chunk := range chunks {
				_, _ = w.Write([]byte(chunk))
			}
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
			{"role": "user", "content": "Test"},
		},
		"stream": true,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	// Parse chunks
	body := w.Body.String()
	scanner := bufio.NewScanner(strings.NewReader(body))

	var chunks []map[string]interface{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			jsonStr := strings.TrimPrefix(line, "data: ")
			if jsonStr == "[DONE]" {
				break
			}
			if jsonStr != "" {
				var chunk map[string]interface{}
				if err := json.Unmarshal([]byte(jsonStr), &chunk); err == nil {
					chunks = append(chunks, chunk)
				}
			}
		}
	}

	require.Greater(t, len(chunks), 0, "Should have at least one chunk")

	// Validate first chunk
	firstChunk := chunks[0]
	assert.Equal(t, "chat.completion.chunk", firstChunk["object"])
	assert.Equal(t, "llama2", firstChunk["model"])

	choices, ok := firstChunk["choices"].([]interface{})
	require.True(t, ok)
	require.Len(t, choices, 1)

	choice, ok := choices[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(0), choice["index"])

	delta, ok := choice["delta"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "assistant", delta["role"])
	assert.Contains(t, delta, "content")

	// Validate final chunk has finish_reason
	if len(chunks) > 1 {
		lastChunk := chunks[len(chunks)-1]
		lastChoices, ok := lastChunk["choices"].([]interface{})
		if ok && len(lastChoices) > 0 {
			lastChoice, ok := lastChoices[0].(map[string]interface{})
			if ok {
				assert.Equal(t, "stop", lastChoice["finish_reason"])
			}
		}
	}
}

// TestStreaming_ClientDisconnect tests that client disconnects don't leak resources
func TestStreaming_ClientDisconnect(t *testing.T) {
	// Create a mock Ollama that sends many chunks slowly
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)

			// Send chunks slowly to allow client disconnect
			for i := 0; i < 10; i++ {
				chunk := `data: {"model":"llama2","message":{"role":"assistant","content":"chunk"},"done":false}` + "\n\n"
				_, _ = w.Write([]byte(chunk))
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				time.Sleep(50 * time.Millisecond)
			}
			chunk := `data: {"model":"llama2","message":{"role":"assistant","content":""},"done":true}` + "\n\n"
			_, _ = w.Write([]byte(chunk))
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
		"stream": true,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Create request with cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Start request in goroutine
	done := make(chan bool)
	go func() {
		router.ServeHTTP(w, req)
		done <- true
	}()

	// Cancel context after a short delay (simulating client disconnect)
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for handler to complete (should handle disconnect gracefully)
	select {
	case <-done:
		// Handler completed, which is fine
	case <-time.After(2 * time.Second):
		t.Fatal("Handler did not complete after client disconnect - potential goroutine leak")
	}
}

// TestStreaming_ErrorHandling tests error handling in streaming
func TestStreaming_ErrorHandling(t *testing.T) {
	// Test 1: Ollama returns error
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"Internal server error"}`))
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
		"stream": true,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Error responses should be forwarded (not converted)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestStreaming_FinalDoneMarker validates [DONE] marker is sent
func TestStreaming_FinalDoneMarker(t *testing.T) {
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)

			chunks := []string{
				`data: {"model":"llama2","message":{"role":"assistant","content":"Hello"},"done":false}` + "\n\n",
				`data: {"model":"llama2","message":{"role":"assistant","content":""},"done":true}` + "\n\n",
			}
			for _, chunk := range chunks {
				_, _ = w.Write([]byte(chunk))
			}
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
		"stream": true,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()
	// Should end with [DONE] marker
	assert.Contains(t, body, "data: [DONE]", "Should contain [DONE] marker")

	// [DONE] should be the last data line
	lines := strings.Split(body, "\n")
	doneFound := false
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.HasPrefix(lines[i], "data: ") {
			if strings.Contains(lines[i], "[DONE]") {
				doneFound = true
			}
			break
		}
	}
	assert.True(t, doneFound, "[DONE] marker should be the last data line")
}
