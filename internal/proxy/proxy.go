// Package proxy provides HTTP proxy handlers for OpenAI-compatible chat completions.
package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/cache"
	"github.com/llamagate/llamagate/internal/tools"
	"github.com/rs/zerolog/log"
)

// Proxy handles forwarding requests to Ollama
type Proxy struct {
	ollamaHost  string
	cache       *cache.Cache
	client      *http.Client
	toolManager *tools.Manager    // Optional tool manager for MCP
	guardrails  *tools.Guardrails // Optional guardrails for tool execution
	serverManager interface{}     // Optional server manager for MCP resource access (to avoid circular import)
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
// serverManager should be *mcpclient.ServerManager (using interface{} to avoid circular import)
func (p *Proxy) SetServerManager(serverManager interface{}) {
	p.serverManager = serverManager
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
	requestID := c.GetString("request_id")
	startTime := time.Now()

	// Parse request body
	var req ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to parse request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message":    "Invalid request body",
				"type":       "invalid_request_error",
				"request_id": requestID,
			},
		})
		return
	}

	// Validate required fields
	if req.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message":    "Model is required",
				"type":       "invalid_request_error",
				"request_id": requestID,
			},
		})
		return
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message":    "Messages are required",
				"type":       "invalid_request_error",
				"request_id": requestID,
			},
		})
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
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"message":    "Tool execution failed",
						"type":       "internal_error",
						"request_id": requestID,
					},
				})
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
	if !req.Stream {
		if cached, found := p.cache.Get(req.Model, finalMessages); found {
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message":    "Internal server error",
				"type":       "internal_error",
				"request_id": requestID,
			},
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message":    "Internal server error",
				"type":       "internal_error",
				"request_id": requestID,
			},
		})
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Handle streaming vs non-streaming
	if req.Stream {
		p.handleStreamingResponse(c, httpReq, requestID, startTime)
	} else {
		p.handleNonStreamingResponse(c, httpReq, requestID, startTime, req.Model, finalMessages)
	}
}

// handleStreamingResponse handles streaming responses from Ollama
func (p *Proxy) handleStreamingResponse(c *gin.Context, httpReq *http.Request, requestID string, startTime time.Time) {
	resp, err := p.client.Do(httpReq)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to forward request to Ollama")
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message":    "Failed to connect to Ollama",
				"type":       "server_error",
				"request_id": requestID,
			},
		})
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

	// Set headers for streaming
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(resp.StatusCode)

	// Stream the response
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to stream response")
		return
	}

	log.Info().
		Str("request_id", requestID).
		Dur("duration", time.Since(startTime)).
		Int("status", resp.StatusCode).
		Msg("Streaming response completed")
}

// handleNonStreamingResponse handles non-streaming responses from Ollama
func (p *Proxy) handleNonStreamingResponse(c *gin.Context, httpReq *http.Request, requestID string, startTime time.Time, model string, messages []Message) {
	resp, err := p.client.Do(httpReq)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to forward request to Ollama")
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message":    "Failed to connect to Ollama",
				"type":       "server_error",
				"request_id": requestID,
			},
		})
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
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message":    "Failed to read response from Ollama",
				"type":       "server_error",
				"request_id": requestID,
			},
		})
		return
	}

	// Cache successful responses
	if resp.StatusCode == http.StatusOK {
		if err := p.cache.Set(model, messages, body); err != nil {
			log.Warn().
				Str("request_id", requestID).
				Err(err).
				Msg("Failed to cache response")
		}
	}

	// Forward response to client
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)

	log.Info().
		Str("request_id", requestID).
		Dur("duration", time.Since(startTime)).
		Int("status", resp.StatusCode).
		Msg("Request completed")
}

// HandleModels handles /v1/models endpoint
func (p *Proxy) HandleModels(c *gin.Context) {
	requestID := c.GetString("request_id")
	startTime := time.Now()

	// Forward to Ollama /api/tags
	ollamaURL := fmt.Sprintf("%s/api/tags", p.ollamaHost)
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), "GET", ollamaURL, http.NoBody)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to create request")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message":    "Internal server error",
				"type":       "internal_error",
				"request_id": requestID,
			},
		})
		return
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to forward request to Ollama")
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message":    "Failed to connect to Ollama",
				"type":       "server_error",
				"request_id": requestID,
			},
		})
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
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message":    "Failed to parse response from Ollama",
				"type":       "server_error",
				"request_id": requestID,
			},
		})
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
