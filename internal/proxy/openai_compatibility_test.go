package proxy

import (
	"bufio"
	"bytes"
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
	"github.com/llamagate/llamagate/internal/tools"
)

// TestOpenAICompatibility_BasicChatCompletion validates basic non-streaming chat completion
// matches OpenAI API response schema
func TestOpenAICompatibility_BasicChatCompletion(t *testing.T) {
	// Create mock Ollama server that returns OpenAI-compatible response
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			// Ollama returns its own format, but proxy should convert to OpenAI format
			ollamaResp := map[string]interface{}{
				"model": "llama2",
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Hello! How can I help you?",
				},
				"done": true,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ollamaResp)
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

	// Make OpenAI-compatible request
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

	// Validate response
	require.Equal(t, http.StatusOK, w.Code, "Response should be 200 OK")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")

	// Validate OpenAI-compatible schema - required fields
	assert.Contains(t, response, "id", "Response must contain 'id' field")
	assert.Contains(t, response, "object", "Response must contain 'object' field")
	assert.Contains(t, response, "created", "Response must contain 'created' field")
	assert.Contains(t, response, "model", "Response must contain 'model' field")
	assert.Contains(t, response, "choices", "Response must contain 'choices' field")
	assert.Contains(t, response, "usage", "Response must contain 'usage' field")

	// Validate object type
	assert.Equal(t, "chat.completion", response["object"], "Object should be 'chat.completion'")

	// Validate model matches request
	assert.Equal(t, "llama2", response["model"], "Model should match request")

	// Validate choices array
	choices, ok := response["choices"].([]interface{})
	require.True(t, ok, "Choices should be an array")
	require.Len(t, choices, 1, "Should have exactly one choice")

	choice, ok := choices[0].(map[string]interface{})
	require.True(t, ok, "Choice should be an object")

	// Validate choice fields
	assert.Contains(t, choice, "index", "Choice must contain 'index' field")
	assert.Contains(t, choice, "message", "Choice must contain 'message' field")
	assert.Contains(t, choice, "finish_reason", "Choice must contain 'finish_reason' field")

	assert.Equal(t, float64(0), choice["index"], "Index should be 0")

	// Validate message
	message, ok := choice["message"].(map[string]interface{})
	require.True(t, ok, "Message should be an object")
	assert.Contains(t, message, "role", "Message must contain 'role' field")
	assert.Contains(t, message, "content", "Message must contain 'content' field")
	assert.Equal(t, "assistant", message["role"], "Role should be 'assistant'")
	assert.NotEmpty(t, message["content"], "Content should not be empty")

	// Validate usage object
	usage, ok := response["usage"].(map[string]interface{})
	require.True(t, ok, "Usage should be an object")
	assert.Contains(t, usage, "prompt_tokens", "Usage must contain 'prompt_tokens' field")
	assert.Contains(t, usage, "completion_tokens", "Usage must contain 'completion_tokens' field")
	assert.Contains(t, usage, "total_tokens", "Usage must contain 'total_tokens' field")
}

// TestOpenAICompatibility_SystemUserRoleHandling validates system and user role handling
func TestOpenAICompatibility_SystemUserRoleHandling(t *testing.T) {
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			// Verify request contains system and user messages
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			messages, ok := req["messages"].([]interface{})
			require.True(t, ok, "Request should contain messages array")

			// Verify system message is present
			foundSystem := false
			foundUser := false
			for _, msg := range messages {
				msgMap, ok := msg.(map[string]interface{})
				if !ok {
					continue
				}
				if role, ok := msgMap["role"].(string); ok {
					if role == "system" {
						foundSystem = true
					}
					if role == "user" {
						foundUser = true
					}
				}
			}
			assert.True(t, foundSystem, "System message should be in request")
			assert.True(t, foundUser, "User message should be in request")

			ollamaResp := map[string]interface{}{
				"model": "llama2",
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Response with system context",
				},
				"done": true,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ollamaResp)
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

	// Request with system and user roles
	reqBody := map[string]interface{}{
		"model": "llama2",
		"messages": []map[string]interface{}{
			{"role": "system", "content": "You are a helpful assistant."},
			{"role": "user", "content": "Hello"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Validate response structure
	choices, ok := response["choices"].([]interface{})
	require.True(t, ok)
	require.Len(t, choices, 1)

	choice, ok := choices[0].(map[string]interface{})
	require.True(t, ok)

	message, ok := choice["message"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "assistant", message["role"], "Response role should be 'assistant'")
}

// TestOpenAICompatibility_StreamingSSE validates streaming SSE response format
func TestOpenAICompatibility_StreamingSSE(t *testing.T) {
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			// Ollama streaming format (SSE)
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)

			// Send multiple SSE chunks
			chunks := []string{
				`data: {"model":"llama2","message":{"role":"assistant","content":"Hello"},"done":false}` + "\n\n",
				`data: {"model":"llama2","message":{"role":"assistant","content":" world"},"done":false}` + "\n\n",
				`data: {"model":"llama2","message":{"role":"assistant","content":""},"done":true}` + "\n\n",
			}
			for _, chunk := range chunks {
				w.Write([]byte(chunk))
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

	// Request with stream=true
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

	// Validate streaming response
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"), "Content-Type should be text/event-stream")
	assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"), "Cache-Control should be no-cache")
	assert.Equal(t, "keep-alive", w.Header().Get("Connection"), "Connection should be keep-alive")

	// Validate SSE format - should contain "data: " prefix
	body := w.Body.String()
	assert.Contains(t, body, "data: ", "Response should contain SSE 'data: ' prefix")

	// Parse SSE chunks
	scanner := bufio.NewScanner(strings.NewReader(body))
	chunkCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			chunkCount++
			// Each data line should be valid JSON (after "data: " prefix)
			jsonStr := strings.TrimPrefix(line, "data: ")
			if jsonStr != "" && jsonStr != "[DONE]" {
				var chunk map[string]interface{}
				err := json.Unmarshal([]byte(jsonStr), &chunk)
				assert.NoError(t, err, "SSE chunk should be valid JSON")
			}
		}
	}
	assert.Greater(t, chunkCount, 0, "Should receive at least one SSE chunk")
}

// TestOpenAICompatibility_ToolFunctionCalling validates tool/function calling response format
func TestOpenAICompatibility_ToolFunctionCalling(t *testing.T) {
	// This test validates that when tools are provided, the response format
	// matches OpenAI's tool calling format
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			// Ollama response with tool calls (if supported)
			ollamaResp := map[string]interface{}{
				"model": "llama2",
				"message": map[string]interface{}{
					"role": "assistant",
					"content": "",
					"tool_calls": []interface{}{
						map[string]interface{}{
							"id":   "call_123",
							"type": "function",
							"function": map[string]interface{}{
								"name":      "get_weather",
								"arguments": `{"location": "San Francisco"}`,
							},
						},
					},
				},
				"done": true,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ollamaResp)
		}
	}))
	defer mockOllama.Close()

	cacheInstance := cache.New()
	proxyInstance := New(mockOllama.URL, cacheInstance)

	// Create tool manager and guardrails for tool calling
	toolManager := tools.NewManager()
	guardrails, err := tools.NewGuardrails(
		[]string{}, // allow all
		[]string{}, // deny none
		10,         // max tool rounds
		5,          // max calls per round
		50,         // max total tool calls
		30*time.Second, // default timeout
		1024*1024,  // max result size 1MB
	)
	require.NoError(t, err, "Should create guardrails")
	proxyInstance.SetToolManager(toolManager, guardrails)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		proxyInstance.HandleChatCompletions(c)
	})

	// Request with tools
	reqBody := map[string]interface{}{
		"model": "llama2",
		"messages": []map[string]interface{}{
			{"role": "user", "content": "What's the weather in San Francisco?"},
		},
		"tools": []map[string]interface{}{
			{
				"type": "function",
				"function": map[string]interface{}{
					"name":        "get_weather",
					"description": "Get weather for a location",
					"parameters": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "The city and state",
							},
						},
						"required": []string{"location"},
					},
				},
			},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Note: Tool calling may use tool loop which returns different format
	// This test validates the response is valid JSON and contains expected structure
	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")

	// Response should have OpenAI-compatible structure
	assert.Contains(t, response, "choices", "Response should contain 'choices' field")

	// If tool calls are present, validate format
	if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
		choice, ok := choices[0].(map[string]interface{})
		if ok {
			message, ok := choice["message"].(map[string]interface{})
			if ok && message["tool_calls"] != nil {
				toolCalls, ok := message["tool_calls"].([]interface{})
				if ok {
					for _, tc := range toolCalls {
						toolCall, ok := tc.(map[string]interface{})
						require.True(t, ok, "Tool call should be an object")
						assert.Contains(t, toolCall, "id", "Tool call must contain 'id' field")
						assert.Contains(t, toolCall, "type", "Tool call must contain 'type' field")
						assert.Contains(t, toolCall, "function", "Tool call must contain 'function' field")

						if function, ok := toolCall["function"].(map[string]interface{}); ok {
							assert.Contains(t, function, "name", "Function must contain 'name' field")
							assert.Contains(t, function, "arguments", "Function must contain 'arguments' field")
						}
					}
				}
			}
		}
	}
}

// TestOpenAICompatibility_ErrorHandling validates error response format matches OpenAI
func TestOpenAICompatibility_ErrorHandling(t *testing.T) {
	cacheInstance := cache.New()
	proxyInstance := New("http://localhost:11434", cacheInstance)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		proxyInstance.HandleChatCompletions(c)
	})

	// Test 1: Missing model field
	reqBody := map[string]interface{}{
		"messages": []map[string]interface{}{
			{"role": "user", "content": "Hello"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code, "Should return 400 for missing model")

	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err, "Error response should be valid JSON")

	// OpenAI error format: {"error": {"message": "...", "type": "...", ...}}
	assert.Contains(t, errorResponse, "error", "Error response must contain 'error' field")

	errorObj, ok := errorResponse["error"].(map[string]interface{})
	if ok {
		assert.Contains(t, errorObj, "message", "Error object must contain 'message' field")
		assert.NotEmpty(t, errorObj["message"], "Error message should not be empty")
	}

	// Test 2: Missing messages field
	reqBody2 := map[string]interface{}{
		"model": "llama2",
	}
	bodyBytes2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	require.Equal(t, http.StatusBadRequest, w2.Code, "Should return 400 for missing messages")

	var errorResponse2 map[string]interface{}
	err2 := json.Unmarshal(w2.Body.Bytes(), &errorResponse2)
	require.NoError(t, err2, "Error response should be valid JSON")
	assert.Contains(t, errorResponse2, "error", "Error response must contain 'error' field")

	// Test 3: Invalid JSON
	req3 := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader("invalid json"))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	require.Equal(t, http.StatusBadRequest, w3.Code, "Should return 400 for invalid JSON")
}

// TestOpenAICompatibility_UsageFields validates usage fields are present and numeric
func TestOpenAICompatibility_UsageFields(t *testing.T) {
	mockOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			ollamaResp := map[string]interface{}{
				"model": "llama2",
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Test response",
				},
				"done": true,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ollamaResp)
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

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Validate usage fields
	usage, ok := response["usage"].(map[string]interface{})
	require.True(t, ok, "Usage should be an object")

	// Validate all three usage fields are present and numeric
	promptTokens, ok := usage["prompt_tokens"]
	require.True(t, ok, "Usage must contain 'prompt_tokens'")
	_, isFloat := promptTokens.(float64)
	_, isInt := promptTokens.(int)
	assert.True(t, isFloat || isInt, "prompt_tokens should be numeric")

	completionTokens, ok := usage["completion_tokens"]
	require.True(t, ok, "Usage must contain 'completion_tokens'")
	_, isFloat = completionTokens.(float64)
	_, isInt = completionTokens.(int)
	assert.True(t, isFloat || isInt, "completion_tokens should be numeric")

	totalTokens, ok := usage["total_tokens"]
	require.True(t, ok, "Usage must contain 'total_tokens'")
	_, isFloat = totalTokens.(float64)
	_, isInt = totalTokens.(int)
	assert.True(t, isFloat || isInt, "total_tokens should be numeric")
}
