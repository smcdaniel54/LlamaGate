// Package proxy provides HTTP proxy handlers for OpenAI-compatible chat completions.
package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/cache"
	"github.com/llamagate/llamagate/internal/middleware"
	"github.com/llamagate/llamagate/internal/response"
	"github.com/llamagate/llamagate/internal/tools"
	"github.com/rs/zerolog/log"
)

// Proxy handles forwarding requests to Ollama
type Proxy struct {
	ollamaHost           string
	cache                *cache.Cache
	client               *http.Client
	toolManager          *tools.Manager         // Optional tool manager for MCP
	guardrails           *tools.Guardrails      // Optional guardrails for tool execution
	serverManager        ServerManagerInterface // Optional server manager for MCP resource access
	resourceFetchTimeout time.Duration          // Timeout for fetching MCP resources
}

// New creates a new proxy instance with default timeout (5 minutes)
func New(ollamaHost string, cache *cache.Cache) *Proxy {
	return NewWithTimeout(ollamaHost, cache, 5*time.Minute)
}

// NewWithTimeout creates a new proxy instance with custom timeout
func NewWithTimeout(ollamaHost string, cache *cache.Cache, timeout time.Duration) *Proxy {
	return &Proxy{
		ollamaHost: ollamaHost,
		cache:      cache,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// SetToolManager sets the tool manager for MCP tool support
func (p *Proxy) SetToolManager(toolManager *tools.Manager, guardrails *tools.Guardrails) {
	p.toolManager = toolManager
	p.guardrails = guardrails
}

// SetServerManager sets the server manager for MCP resource access
func (p *Proxy) SetServerManager(serverManager ServerManagerInterface) {
	p.serverManager = serverManager
}

// SetResourceFetchTimeout sets the timeout for fetching MCP resources
func (p *Proxy) SetResourceFetchTimeout(timeout time.Duration) {
	p.resourceFetchTimeout = timeout
}

// Close closes all connections and cleans up resources
// This should be called during graceful shutdown
func (p *Proxy) Close() {
	// Close idle HTTP client connections
	if p.client != nil && p.client.Transport != nil {
		if transport, ok := p.client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
}

// ChatCompletionRequest represents an OpenAI-compatible chat completion request
type ChatCompletionRequest struct {
	Model       string           `json:"model"`
	Messages    []Message        `json:"messages"`
	Stream      bool             `json:"stream,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Tools       []OpenAITool     `json:"tools,omitempty"`       // OpenAI tools format
	Functions   []OpenAIFunction `json:"functions,omitempty"`   // Legacy functions format
	ToolChoice  interface{}      `json:"tool_choice,omitempty"` // "none", "auto", or object
}

// Message represents a chat message
type Message struct {
	Role       string      `json:"role"`
	Content    interface{} `json:"content"` // Can be string or array
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"` // For tool role messages
	Name       string      `json:"name,omitempty"`         // For function role messages (legacy)
}

// ToolCall represents a tool call in a message
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"` // "function"
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction represents the function details in a tool call
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// OpenAITool represents an OpenAI tool definition
type OpenAITool struct {
	Type     string         `json:"type"` // "function"
	Function OpenAIFunction `json:"function"`
}

// OpenAIFunction represents an OpenAI function definition
type OpenAIFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
}

// HandleChatCompletions handles /v1/chat/completions endpoint
func (p *Proxy) HandleChatCompletions(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	startTime := time.Now()

	// Parse request body
	var req ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to parse request body")
		response.BadRequest(c, "Invalid request body", requestID)
		return
	}

	// Validate required fields
	if err := ValidateChatRequest(&req); err != nil {
		var msg string
		if valErr, ok := err.(*ValidationError); ok {
			msg = valErr.Message
		} else {
			msg = err.Error()
		}
		response.BadRequest(c, msg, requestID)
		return
	}

	// Handle tool execution if MCP is enabled and tools are available
	if p.toolManager != nil && p.guardrails != nil && !req.Stream {
		// Check if tools are requested or if we should inject available tools
		hasTools := len(req.Tools) > 0 || len(req.Functions) > 0
		availableTools := p.toolManager.GetAllTools()

		if hasTools || len(availableTools) > 0 {
			// Use tool execution loop
			result, err := p.executeToolLoop(c.Request.Context(), requestID, req.Model, req.Messages, p.ollamaHost, req.Stream, req.Temperature)
			if err != nil {
				log.Error().
					Str("request_id", requestID).
					Err(err).
					Msg("Tool execution loop failed")
				response.InternalError(c, "Tool execution failed", requestID)
				return
			}

			c.Data(http.StatusOK, "application/json", result)
			log.Info().
				Str("request_id", requestID).
				Dur("duration", time.Since(startTime)).
				Msg("Tool execution completed")
			return
		}
	}

	// Inject MCP resource context if MCP URIs are found in messages
	finalMessages, err := p.injectMCPResourceContext(c.Request.Context(), requestID, req.Messages)
	if err != nil {
		log.Warn().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to inject MCP resource context, continuing without it")
		// Continue with original messages if resource injection fails
		finalMessages = req.Messages
	}

	// Build final messages - check if tools will be injected before checking cache
	// This ensures cache keys match the actual query sent to Ollama
	if p.toolManager != nil && (len(req.Tools) > 0 || len(req.Functions) > 0) {
		availableTools := p.toolManager.GetAllTools()
		if len(availableTools) > 0 {
			// Build messages with tool context for cache key
			toolDescriptions := make([]string, 0, len(availableTools))
			for _, tool := range availableTools {
				toolDescriptions = append(toolDescriptions, fmt.Sprintf("- %s: %s", tool.NamespacedName, tool.Description))
			}
			systemMsg := Message{
				Role:    "system",
				Content: fmt.Sprintf("Available tools:\n%s", strings.Join(toolDescriptions, "\n")),
			}
			// Prepend system message to final messages
			finalMessages = append([]Message{systemMsg}, req.Messages...)
		}
	}

	// Check cache for non-streaming requests using final messages (with tool context if applicable)
	// Cache key includes all parameters that affect the response
	if !req.Stream {
		cacheParams := cache.KeyParams{
			Model:       req.Model,
			Messages:    finalMessages,
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
			Tools:       req.Tools,
			Functions:   req.Functions,
			ToolChoice:  req.ToolChoice,
		}
		if cached, found := p.cache.Get(cacheParams); found {
			log.Info().
				Str("request_id", requestID).
				Str("model", req.Model).
				Dur("duration", time.Since(startTime)).
				Msg("Cache hit")
			c.Data(http.StatusOK, "application/json", cached)
			return
		}
		log.Debug().
			Str("request_id", requestID).
			Str("model", req.Model).
			Msg("Cache miss")
	}

	// Convert to Ollama format using final messages
	ollamaReq := map[string]interface{}{
		"model":    req.Model,
		"messages": finalMessages,
		"stream":   req.Stream,
	}

	if req.Temperature > 0 {
		ollamaReq["options"] = map[string]interface{}{
			"temperature": req.Temperature,
		}
	}

	// Marshal request
	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to marshal request")
		response.InternalError(c, "Internal server error", requestID)
		return
	}

	// Forward to Ollama
	ollamaURL := fmt.Sprintf("%s/api/chat", p.ollamaHost)
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), "POST", ollamaURL, bytes.NewReader(reqBody))
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to create request")
		response.InternalError(c, "Internal server error", requestID)
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	// Propagate request ID to Ollama for correlation
	httpReq.Header.Set("X-Request-ID", requestID)

	// Handle streaming vs non-streaming
	if req.Stream {
		p.handleStreamingResponse(c, httpReq, requestID, startTime, req.Model)
	} else {
		// Build cache params for storing response
		cacheParams := cache.KeyParams{
			Model:       req.Model,
			Messages:    finalMessages,
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
			Tools:       req.Tools,
			Functions:   req.Functions,
			ToolChoice:  req.ToolChoice,
		}
		p.handleNonStreamingResponse(c, httpReq, requestID, startTime, cacheParams)
	}
}

// handleStreamingResponse handles streaming responses from Ollama and converts them to OpenAI-compatible SSE format
func (p *Proxy) handleStreamingResponse(c *gin.Context, httpReq *http.Request, requestID string, startTime time.Time, model string) {
	resp, err := p.client.Do(httpReq)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to forward request to Ollama")
		response.ServerError(c, "Failed to connect to Ollama", requestID)
		return
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Warn().
				Str("request_id", requestID).
				Err(closeErr).
				Msg("Failed to close response body")
		}
	}()

	// Set headers for streaming (OpenAI-compatible SSE)
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(resp.StatusCode)

	// Flush headers immediately
	c.Writer.Flush()

	// Handle non-OK responses
	if resp.StatusCode != http.StatusOK {
		// Forward error response as-is
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			log.Error().
				Str("request_id", requestID).
				Err(err).
				Msg("Failed to stream error response")
		}
		return
	}

	// Parse Ollama SSE chunks and convert to OpenAI format
	scanner := bufio.NewScanner(resp.Body)
	chunkID := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	created := time.Now().Unix()
	accumulatedContent := ""

	for scanner.Scan() {
		// Check for client disconnect
		select {
		case <-c.Request.Context().Done():
			log.Info().
				Str("request_id", requestID).
				Msg("Client disconnected, stopping stream")
			return
		default:
		}

		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Extract JSON from SSE line
		jsonStr := strings.TrimPrefix(line, "data: ")
		if jsonStr == "" {
			continue
		}

		// Parse Ollama chunk
		var ollamaChunk map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &ollamaChunk); err != nil {
			log.Warn().
				Str("request_id", requestID).
				Err(err).
				Str("line", line).
				Msg("Failed to parse Ollama SSE chunk, skipping")
			continue
		}

		// Check if stream is done
		done, _ := ollamaChunk["done"].(bool)
		if done {
			// Send final chunk with accumulated content and finish_reason
			finalChunk := convertOllamaStreamChunkToOpenAI(ollamaChunk, model, chunkID, created, true)
			if finalChunk != nil {
				chunkJSON, err := json.Marshal(finalChunk)
				if err == nil {
					_, writeErr := fmt.Fprintf(c.Writer, "data: %s\n\n", chunkJSON)
					if writeErr != nil {
						log.Warn().
							Str("request_id", requestID).
							Err(writeErr).
							Msg("Failed to write final chunk, client may have disconnected")
						return
					}
					c.Writer.Flush()
				}
			}
			// Send OpenAI-compatible [DONE] message
			_, writeErr := fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
			if writeErr != nil {
				log.Warn().
					Str("request_id", requestID).
					Err(writeErr).
					Msg("Failed to write [DONE] marker, client may have disconnected")
				return
			}
			c.Writer.Flush()
			break
		}

		// Convert Ollama chunk to OpenAI format
		openAIChunk := convertOllamaStreamChunkToOpenAI(ollamaChunk, model, chunkID, created, false)
		if openAIChunk == nil {
			continue
		}

		// Marshal and send as SSE
		chunkJSON, err := json.Marshal(openAIChunk)
		if err != nil {
			log.Warn().
				Str("request_id", requestID).
				Err(err).
				Msg("Failed to marshal OpenAI chunk, skipping")
			continue
		}

		// Write SSE-formatted chunk: "data: {...}\n\n"
		_, writeErr := fmt.Fprintf(c.Writer, "data: %s\n\n", chunkJSON)
		if writeErr != nil {
			log.Warn().
				Str("request_id", requestID).
				Err(writeErr).
				Msg("Failed to write chunk, client may have disconnected")
			return
		}

		// Flush to ensure client receives chunk immediately
		c.Writer.Flush()

		// Accumulate content for final chunk
		if message, ok := ollamaChunk["message"].(map[string]interface{}); ok {
			if content, ok := message["content"].(string); ok {
				accumulatedContent += content
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Error reading streaming response")
		// Send error chunk in OpenAI format
		errorChunk := map[string]interface{}{
			"error": map[string]interface{}{
				"message": fmt.Sprintf("Stream error: %v", err),
				"type":    "stream_error",
			},
		}
		errorJSON, _ := json.Marshal(errorChunk)
		fmt.Fprintf(c.Writer, "data: %s\n\n", errorJSON)
		c.Writer.Flush()
		return
	}

	log.Info().
		Str("request_id", requestID).
		Dur("duration", time.Since(startTime)).
		Msg("Streaming response completed")
}

// convertOllamaStreamChunkToOpenAI converts an Ollama streaming chunk to OpenAI-compatible format
func convertOllamaStreamChunkToOpenAI(ollamaChunk map[string]interface{}, model, chunkID string, created int64, isFinal bool) map[string]interface{} {
	var deltaContent string
	var finishReason interface{}

	if message, ok := ollamaChunk["message"].(map[string]interface{}); ok {
		if content, ok := message["content"].(string); ok {
			deltaContent = content
		}
	}

	if isFinal {
		finishReason = "stop"
	}

	// OpenAI streaming chunk format
	chunk := map[string]interface{}{
		"id":      chunkID,
		"object":  "chat.completion.chunk",
		"created": created,
		"model":   model,
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"delta": map[string]interface{}{
					"role":    "assistant",
					"content": deltaContent,
				},
				"finish_reason": finishReason,
			},
		},
	}

	return chunk
}

// handleNonStreamingResponse handles non-streaming responses from Ollama
func (p *Proxy) handleNonStreamingResponse(c *gin.Context, httpReq *http.Request, requestID string, startTime time.Time, cacheParams cache.KeyParams) {
	resp, err := p.client.Do(httpReq)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to forward request to Ollama")
		response.ServerError(c, "Failed to connect to Ollama", requestID)
		return
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Warn().
				Str("request_id", requestID).
				Err(closeErr).
				Msg("Failed to close response body")
		}
	}()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to read response from Ollama")
		response.ServerError(c, "Failed to read response from Ollama", requestID)
		return
	}

	// Convert Ollama response to OpenAI format for successful responses
	var responseBody []byte
	if resp.StatusCode == http.StatusOK {
		// Parse Ollama response
		var ollamaResp map[string]interface{}
		if err := json.Unmarshal(body, &ollamaResp); err != nil {
			log.Error().
				Str("request_id", requestID).
				Err(err).
				Msg("Failed to parse Ollama response")
			response.ServerError(c, "Failed to parse response from Ollama", requestID)
			return
		}

		// Convert to OpenAI format
		openAIResp := convertOllamaToOpenAIFormat(ollamaResp, cacheParams.Model)

		// Marshal to JSON
		responseBody, err = json.Marshal(openAIResp)
		if err != nil {
			log.Error().
				Str("request_id", requestID).
				Err(err).
				Msg("Failed to marshal OpenAI response")
			response.ServerError(c, "Failed to format response", requestID)
			return
		}

		// Cache the converted OpenAI-format response
		if err := p.cache.Set(cacheParams, responseBody); err != nil {
			log.Warn().
				Str("request_id", requestID).
				Err(err).
				Msg("Failed to cache response")
		}
	} else {
		// For non-OK responses, forward as-is
		responseBody = body
	}

	// Return response to client
	c.Data(resp.StatusCode, "application/json", responseBody)

	log.Info().
		Str("request_id", requestID).
		Dur("duration", time.Since(startTime)).
		Int("status", resp.StatusCode).
		Msg("Request completed")
}

// HandleModels handles /v1/models endpoint
func (p *Proxy) HandleModels(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	startTime := time.Now()

	// Forward to Ollama /api/tags
	ollamaURL := fmt.Sprintf("%s/api/tags", p.ollamaHost)
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), "GET", ollamaURL, http.NoBody)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to create request")
		response.InternalError(c, "Internal server error", requestID)
		return
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to forward request to Ollama")
		response.ServerError(c, "Failed to connect to Ollama", requestID)
		return
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Warn().
				Str("request_id", requestID).
				Err(closeErr).
				Msg("Failed to close response body")
		}
	}()

	// Read Ollama response
	var ollamaResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to decode Ollama response")
		response.ServerError(c, "Failed to parse response from Ollama", requestID)
		return
	}

	// Convert Ollama format to OpenAI format
	models := []map[string]interface{}{}
	if modelsData, ok := ollamaResp["models"].([]interface{}); ok {
		for _, modelData := range modelsData {
			if modelMap, ok := modelData.(map[string]interface{}); ok {
				if name, ok := modelMap["name"].(string); ok {
					models = append(models, map[string]interface{}{
						"id":       name,
						"object":   "model",
						"created":  0,
						"owned_by": "ollama",
					})
				}
			}
		}
	}

	openAIResp := map[string]interface{}{
		"object": "list",
		"data":   models,
	}

	c.JSON(http.StatusOK, openAIResp)

	log.Info().
		Str("request_id", requestID).
		Dur("duration", time.Since(startTime)).
		Int("count", len(models)).
		Msg("Models list retrieved")
}
