package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/llamagate/llamagate/internal/extensions"
)

// CreateExtensionLLMHandler creates an LLM handler function for extensions
// This allows extensions to make LLM calls without circular dependencies
func (p *Proxy) CreateExtensionLLMHandler() extensions.LLMHandlerFunc {
	return func(ctx context.Context, model string, messages []map[string]interface{}, options map[string]interface{}) (map[string]interface{}, error) {
		// Convert messages to proxy.Message format
		proxyMessages := make([]Message, 0, len(messages))
		for _, msg := range messages {
			proxyMsg := Message{
				Role: getString(msg, "role", "user"),
			}

			// Handle content (can be string or array)
			if content, ok := msg["content"]; ok {
				proxyMsg.Content = content
			}

			proxyMessages = append(proxyMessages, proxyMsg)
		}

		// Build Ollama request
		ollamaReq := map[string]interface{}{
			"model":    model,
			"messages": convertMessagesToOllamaFormat(proxyMessages),
			"stream":   false,
		}

		// Add temperature if provided
		if temp, ok := options["temperature"].(float64); ok && temp > 0 {
			ollamaReq["options"] = map[string]interface{}{
				"temperature": temp,
			}
		}

		// Marshal request
		reqBody, err := json.Marshal(ollamaReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		// Forward to Ollama
		ollamaURL := fmt.Sprintf("%s/api/chat", p.ollamaHost)
		httpReq, err := http.NewRequestWithContext(ctx, "POST", ollamaURL, bytes.NewReader(reqBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, err := p.client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
		}

		// Read response
		var ollamaResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		// Convert Ollama response to OpenAI format
		return convertOllamaToOpenAIFormat(ollamaResp, model), nil
	}
}

// convertMessagesToOllamaFormat converts proxy messages to Ollama format
func convertMessagesToOllamaFormat(messages []Message) []map[string]interface{} {
	ollamaMessages := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		ollamaMsg := map[string]interface{}{
			"role": msg.Role,
		}

		// Handle content
		if msg.Content != nil {
			ollamaMsg["content"] = msg.Content
		}

		ollamaMessages = append(ollamaMessages, ollamaMsg)
	}
	return ollamaMessages
}

// convertOllamaToOpenAIFormat converts Ollama response to OpenAI format
func convertOllamaToOpenAIFormat(ollamaResp map[string]interface{}, model string) map[string]interface{} {
	// Extract message content
	var content string
	if message, ok := ollamaResp["message"].(map[string]interface{}); ok {
		if msgContent, ok := message["content"].(string); ok {
			content = msgContent
		}
	}

	// Build OpenAI-format response
	response := map[string]interface{}{
		"id":      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": content,
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     0,
			"completion_tokens": 0,
			"total_tokens":      0,
		},
	}

	return response
}

// getString safely extracts a string value from a map
func getString(m map[string]interface{}, key string, defaultValue string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}
