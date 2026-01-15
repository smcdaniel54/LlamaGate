package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/llamagate/llamagate/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxy_CreateExtensionLLMHandler(t *testing.T) {
	// Create a mock Ollama server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Return mock Ollama response
		response := map[string]interface{}{
			"message": map[string]interface{}{
				"role":    "assistant",
				"content": "Mock response from Ollama",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Create proxy with mock server
	cacheInstance := cache.New()
	proxyInstance := NewWithTimeout(mockServer.URL, cacheInstance, 0)

	// Create LLM handler
	llmHandler := proxyInstance.CreateExtensionLLMHandler()
	require.NotNil(t, llmHandler)

	// Test LLM call
	messages := []map[string]interface{}{
		{
			"role":    "user",
			"content": "Hello",
		},
	}

	result, err := llmHandler(context.Background(), "llama3.2", messages, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify response format
	assert.Contains(t, result, "choices")
	choices, ok := result["choices"]
	require.True(t, ok)

	// Handle both []interface{} and []map[string]interface{} types
	var choice map[string]interface{}
	switch v := choices.(type) {
	case []interface{}:
		require.Len(t, v, 1)
		choice, ok = v[0].(map[string]interface{})
		require.True(t, ok)
	case []map[string]interface{}:
		require.Len(t, v, 1)
		choice = v[0]
	default:
		t.Fatalf("unexpected choices type: %T", choices)
	}

	assert.Contains(t, choice, "message")
	message, ok := choice["message"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Mock response from Ollama", message["content"])
}

func TestConvertOllamaToOpenAIFormat(t *testing.T) {
	ollamaResp := map[string]interface{}{
		"message": map[string]interface{}{
			"role":    "assistant",
			"content": "Test response",
		},
	}

	result := convertOllamaToOpenAIFormat(ollamaResp, "llama3.2")

	assert.Contains(t, result, "choices")
	assert.Contains(t, result, "model")
	assert.Equal(t, "llama3.2", result["model"])

	choices, ok := result["choices"]
	require.True(t, ok)

	// Handle both []interface{} and []map[string]interface{} types
	var choice map[string]interface{}
	switch v := choices.(type) {
	case []interface{}:
		require.Len(t, v, 1)
		choice, ok = v[0].(map[string]interface{})
		require.True(t, ok)
	case []map[string]interface{}:
		require.Len(t, v, 1)
		choice = v[0]
	default:
		t.Fatalf("unexpected choices type: %T", choices)
	}

	message, ok := choice["message"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Test response", message["content"])
}
