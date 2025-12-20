package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/cache"
	"github.com/rs/zerolog/log"
)

// Proxy handles forwarding requests to Ollama
type Proxy struct {
	ollamaHost string
	cache      *cache.Cache
	client     *http.Client
}

// New creates a new proxy instance
func New(ollamaHost string, cache *cache.Cache) *Proxy {
	return &Proxy{
		ollamaHost: ollamaHost,
		cache:      cache,
		client: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for LLM requests
		},
	}
}

// ChatCompletionRequest represents an OpenAI-compatible chat completion request
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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
				"message": "Invalid request body",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	// Validate required fields
	if req.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Model is required",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Messages are required",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	// Check cache for non-streaming requests
	if !req.Stream {
		if cached, found := p.cache.Get(req.Model, req.Messages); found {
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

	// Convert to Ollama format
	ollamaReq := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
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
				"message": "Internal server error",
				"type":    "internal_error",
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
				"message": "Internal server error",
				"type":    "internal_error",
			},
		})
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Handle streaming vs non-streaming
	if req.Stream {
		p.handleStreamingResponse(c, httpReq, requestID, startTime)
	} else {
		p.handleNonStreamingResponse(c, httpReq, requestID, startTime, req.Model, req.Messages)
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
				"message": "Failed to connect to Ollama",
				"type":    "server_error",
			},
		})
		return
	}
	defer resp.Body.Close()

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
				"message": "Failed to connect to Ollama",
				"type":    "server_error",
			},
		})
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to read response from Ollama")
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message": "Failed to read response from Ollama",
				"type":    "server_error",
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
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), "GET", ollamaURL, nil)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to create request")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": "Internal server error",
				"type":    "internal_error",
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
				"message": "Failed to connect to Ollama",
				"type":    "server_error",
			},
		})
		return
	}
	defer resp.Body.Close()

	// Read Ollama response
	var ollamaResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Failed to decode Ollama response")
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{
				"message": "Failed to parse response from Ollama",
				"type":    "server_error",
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
						"id":      name,
						"object":  "model",
						"created": 0,
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

