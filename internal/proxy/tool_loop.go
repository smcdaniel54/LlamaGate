package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/llamagate/llamagate/internal/tools"
	"github.com/rs/zerolog/log"
)

// executeToolLoop executes tools in a loop until no more tool calls or max rounds reached
func (p *Proxy) executeToolLoop(ctx context.Context, requestID string, model string, initialMessages []Message, ollamaHost string, stream bool, temperature float64) ([]byte, error) {
	_ = stream // stream parameter is reserved for future use
	if p.toolManager == nil || p.guardrails == nil {
		// No tool support, return error or handle normally
		return nil, fmt.Errorf("tool execution not available")
	}

	messages := make([]Message, len(initialMessages))
	copy(messages, initialMessages)

	round := 0
	maxRounds := p.guardrails.MaxToolRounds()

	for round < maxRounds {
		// Get available tools and inject them into the request
		availableTools := p.toolManager.GetAllTools()
		if len(availableTools) == 0 {
			// No tools available, proceed with normal request
			break
		}

		// Convert tools to OpenAI format
		openAITools := tools.ToolsToOpenAIFormat(availableTools)
		toolDefs := make([]OpenAITool, len(openAITools))
		for i, tool := range openAITools {
			toolDefs[i] = OpenAITool{
				Type:     "function",
				Function: OpenAIFunction(tool),
			}
		}

		// Make request to Ollama with tools
		ollamaReq := map[string]interface{}{
			"model":    model,
			"messages": messages,
			"stream":   false, // Tool execution only supports non-streaming
		}

		if temperature > 0 {
			ollamaReq["options"] = map[string]interface{}{
				"temperature": temperature,
			}
		}

		// Convert tools to Ollama format (if Ollama supports it)
		// Note: Ollama may not support tools directly, so we might need to handle this differently
		// For now, we'll inject tools as part of the system message or use a different approach
		// This is a placeholder - actual implementation depends on Ollama's tool support

		reqBody, err := json.Marshal(ollamaReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		// Forward to Ollama
		ollamaURL := fmt.Sprintf("%s/api/chat", ollamaHost)
		httpReq, err := http.NewRequestWithContext(ctx, "POST", ollamaURL, bytes.NewReader(reqBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := p.client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("failed to forward request to Ollama: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Warn().
				Str("request_id", requestID).
				Err(closeErr).
				Msg("Failed to close response body")
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
		}

		// Parse response
		var ollamaResp map[string]interface{}
		if err := json.Unmarshal(body, &ollamaResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		// Extract message from response
		var assistantMessage Message
		switch messageData := ollamaResp["message"].(type) {
		case map[string]interface{}:
			// Convert to our Message type
			if role, ok := messageData["role"].(string); ok {
				assistantMessage.Role = role
			}
			if content, ok := messageData["content"].(string); ok {
				assistantMessage.Content = content
			}

			// Check for tool calls (Ollama format may vary)
			// This is a placeholder - actual format depends on Ollama's implementation
			if toolCallsData, ok := messageData["tool_calls"].([]interface{}); ok {
				assistantMessage.ToolCalls = parseToolCalls(toolCallsData)
			}
		default:
			// Fallback: create default assistant message
			assistantMessage = Message{
				Role:    "assistant",
				Content: "",
			}
		}

		// Add assistant message to conversation
		messages = append(messages, assistantMessage)

		// Check if there are tool calls
		if len(assistantMessage.ToolCalls) == 0 {
			// No more tool calls, return the final response
			// Convert back to OpenAI format
			lastMsg := messages[len(messages)-1]
			return convertToOpenAIResponse(&lastMsg, model), nil
		}

		// Validate tool calls count
		if err := p.guardrails.ValidateToolCallsPerRound(len(assistantMessage.ToolCalls)); err != nil {
			log.Warn().
				Str("request_id", requestID).
				Err(err).
				Msg("Tool calls per round limit exceeded")
			// Return error response
			return createErrorResponse("too_many_tool_calls", err.Error(), requestID), nil
		}

		// Execute tool calls
		toolMessages := make([]Message, 0, len(assistantMessage.ToolCalls))
		for _, toolCall := range assistantMessage.ToolCalls {
			// Validate tool is allowed
			if err := p.guardrails.ValidateToolCall(toolCall.Function.Name); err != nil {
				log.Warn().
					Str("request_id", requestID).
					Str("tool", toolCall.Function.Name).
					Err(err).
					Msg("Tool call denied")
				// Add error message
				toolMessages = append(toolMessages, Message{
					Role:       "tool",
					ToolCallID: toolCall.ID,
					Content:    fmt.Sprintf(`{"error": "Tool call denied: %s"}`, err.Error()),
				})
				continue
			}

			// Execute tool
			result, err := p.executeTool(ctx, requestID, toolCall)
			if err != nil {
				log.Error().
					Str("request_id", requestID).
					Str("tool", toolCall.Function.Name).
					Err(err).
					Msg("Tool execution failed")
				// Add error message
				toolMessages = append(toolMessages, Message{
					Role:       "tool",
					ToolCallID: toolCall.ID,
					Content:    fmt.Sprintf(`{"error": "Tool execution failed: %s"}`, err.Error()),
				})
				continue
			}

			// Truncate result if needed
			result = p.guardrails.TruncateResult(result)

			// Add tool result message
			toolMessages = append(toolMessages, Message{
				Role:       "tool",
				ToolCallID: toolCall.ID,
				Content:    result,
			})
		}

		// Add tool messages to conversation
		messages = append(messages, toolMessages...)

		round++
		log.Debug().
			Str("request_id", requestID).
			Int("round", round).
			Int("tool_calls", len(assistantMessage.ToolCalls)).
			Msg("Tool execution round completed")
	}

	if round >= maxRounds {
		log.Warn().
			Str("request_id", requestID).
			Int("rounds", round).
			Msg("Maximum tool rounds reached")
		return createErrorResponse("max_tool_rounds_exceeded", fmt.Sprintf("Maximum tool rounds (%d) exceeded", maxRounds), requestID), nil
	}

	// Final response
	lastMsg := messages[len(messages)-1]
	return convertToOpenAIResponse(&lastMsg, model), nil
}

// executeTool executes a single tool call
func (p *Proxy) executeTool(ctx context.Context, requestID string, toolCall ToolCall) (string, error) {
	// Parse tool name to extract server and tool name
	// Format: mcp.<serverName>.<toolName>
	parts := strings.Split(toolCall.Function.Name, ".")
	if len(parts) != 3 || parts[0] != "mcp" {
		return "", fmt.Errorf("invalid tool name format: %s", toolCall.Function.Name)
	}

	serverName := parts[1]
	originalToolName := parts[2]

	// Get MCP client
	client, err := p.toolManager.GetClient(serverName)
	if err != nil {
		return "", fmt.Errorf("failed to get MCP client: %w", err)
	}

	// Parse arguments
	var arguments map[string]interface{}
	if toolCall.Function.Arguments != "" {
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &arguments); err != nil {
			return "", fmt.Errorf("failed to parse tool arguments: %w", err)
		}
	} else {
		arguments = make(map[string]interface{})
	}

	// Execute tool with timeout
	timeout := p.guardrails.GetTimeout()
	toolCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	startTime := time.Now()
	result, err := client.CallTool(toolCtx, originalToolName, arguments)
	duration := time.Since(startTime)

	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Str("tool", toolCall.Function.Name).
			Dur("duration", duration).
			Err(err).
			Msg("Tool execution failed")
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	// Convert result to string
	// MCP returns content as an array, we'll extract text content
	var resultText strings.Builder
	for _, content := range result.Content {
		if content.Type == "text" {
			resultText.WriteString(content.Text)
		}
	}

	log.Info().
		Str("request_id", requestID).
		Str("tool", toolCall.Function.Name).
		Dur("duration", duration).
		Bool("is_error", result.IsError).
		Msg("Tool executed successfully")

	return resultText.String(), nil
}

// parseToolCalls parses tool calls from Ollama response
func parseToolCalls(data []interface{}) []ToolCall {
	toolCalls := make([]ToolCall, 0, len(data))
	for _, item := range data {
		if itemMap, ok := item.(map[string]interface{}); ok {
			toolCall := ToolCall{
				Type: "function",
			}
			if id, ok := itemMap["id"].(string); ok {
				toolCall.ID = id
			}
			if functionData, ok := itemMap["function"].(map[string]interface{}); ok {
				if name, ok := functionData["name"].(string); ok {
					toolCall.Function.Name = name
				}
				if args, ok := functionData["arguments"].(string); ok {
					toolCall.Function.Arguments = args
				} else if argsObj, ok := functionData["arguments"].(map[string]interface{}); ok {
					// Convert object to JSON string
					if argsJSON, marshalErr := json.Marshal(argsObj); marshalErr == nil {
						toolCall.Function.Arguments = string(argsJSON)
					}
				}
			}
			toolCalls = append(toolCalls, toolCall)
		}
	}
	return toolCalls
}

// convertToOpenAIResponse converts a message to OpenAI response format
func convertToOpenAIResponse(message *Message, model string) []byte {
	response := map[string]interface{}{
		"id":      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role":    message.Role,
					"content": message.Content,
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

	// Add tool calls if present
	if len(message.ToolCalls) > 0 {
		choice := response["choices"].([]map[string]interface{})[0]
		choice["message"].(map[string]interface{})["tool_calls"] = message.ToolCalls
		choice["finish_reason"] = "tool_calls"
	}

	jsonData, _ := json.Marshal(response)
	return jsonData
}

// createErrorResponse creates an error response in OpenAI format
func createErrorResponse(errorType, message, requestID string) []byte {
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"message":    message,
			"type":       errorType,
			"request_id": requestID,
		},
	}
	jsonData, _ := json.Marshal(response)
	return jsonData
}
