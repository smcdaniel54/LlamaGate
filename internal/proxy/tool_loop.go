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

	"github.com/llamagate/llamagate/internal/middleware"
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
	totalToolCalls := 0 // Track total tool calls across all rounds

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

		// Add tools to the request (OpenAI format)
		// Ollama may support tools directly, or we may need to handle them differently
		// For now, we add them in OpenAI format and also inject as system message for compatibility
		if len(toolDefs) > 0 {
			ollamaReq["tools"] = toolDefs
		}

		if temperature > 0 {
			ollamaReq["options"] = map[string]interface{}{
				"temperature": temperature,
			}
		}

		// Also inject tools as system message for models that don't support tools directly
		// This ensures tool information is available to the model
		if len(availableTools) > 0 {
			toolDescriptions := make([]string, 0, len(availableTools))
			for _, tool := range availableTools {
				toolDescriptions = append(toolDescriptions, fmt.Sprintf("- %s: %s", tool.NamespacedName, tool.Description))
			}
			systemMsg := Message{
				Role:    "system",
				Content: fmt.Sprintf("Available tools:\n%s\n\nYou can call these tools by responding with tool_calls in your response.", strings.Join(toolDescriptions, "\n")),
			}
			// Prepend system message if not already present
			if len(messages) == 0 || messages[0].Role != "system" {
				messages = append([]Message{systemMsg}, messages...)
				ollamaReq["messages"] = messages
			}
		}

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
		// Propagate request ID to Ollama for correlation
		httpReq.Header.Set("X-Request-ID", requestID)

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

		// Validate tool calls count per round
		if err := p.guardrails.ValidateToolCallsPerRound(len(assistantMessage.ToolCalls)); err != nil {
			log.Warn().
				Str("request_id", requestID).
				Int("round", round+1).
				Int("tool_calls", len(assistantMessage.ToolCalls)).
				Err(err).
				Msg("Tool calls per round limit exceeded")
			// Return clear, user-facing error
			return createErrorResponse(
				"too_many_tool_calls_per_round",
				fmt.Sprintf("Maximum tool calls per round (%d) exceeded. This request attempted %d tool calls in round %d.", p.guardrails.MaxCallsPerRound(), len(assistantMessage.ToolCalls), round+1),
				requestID,
			), nil
		}

		// Check total tool calls limit before executing
		if err := p.guardrails.ValidateTotalToolCalls(totalToolCalls + len(assistantMessage.ToolCalls)); err != nil {
			log.Warn().
				Str("request_id", requestID).
				Int("round", round+1).
				Int("total_tool_calls", totalToolCalls).
				Int("pending_calls", len(assistantMessage.ToolCalls)).
				Err(err).
				Msg("Total tool calls limit exceeded")
			// Return clear, user-facing error
			return createErrorResponse(
				"max_total_tool_calls_exceeded",
				fmt.Sprintf("Maximum total tool calls (%d) exceeded. This request has made %d tool calls and attempted %d more.", p.guardrails.MaxTotalToolCalls(), totalToolCalls, len(assistantMessage.ToolCalls)),
				requestID,
			), nil
		}

		// Execute tool calls
		toolMessages := make([]Message, 0, len(assistantMessage.ToolCalls))
		for _, toolCall := range assistantMessage.ToolCalls {
			// Increment total tool calls counter
			totalToolCalls++
			
			// Check total limit again after increment (defensive check)
			if err := p.guardrails.ValidateTotalToolCalls(totalToolCalls); err != nil {
				log.Warn().
					Str("request_id", requestID).
					Int("round", round+1).
					Int("total_tool_calls", totalToolCalls).
					Err(err).
					Msg("Total tool calls limit exceeded during execution")
				// Return clear, user-facing error
				return createErrorResponse(
					"max_total_tool_calls_exceeded",
					fmt.Sprintf("Maximum total tool calls (%d) exceeded. This request has executed %d tool calls.", p.guardrails.MaxTotalToolCalls(), totalToolCalls),
					requestID,
				), nil
			}
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
			Int("tool_calls_this_round", len(assistantMessage.ToolCalls)).
			Int("total_tool_calls", totalToolCalls).
			Msg("Tool execution round completed")

		// Check if we've hit the max rounds limit
		// If we have tool calls and we're at or past the limit, we need to exit with an error
		// Note: round is now the number of completed rounds (1-indexed), so round == maxRounds means we've done maxRounds rounds
		if round >= maxRounds {
			// Check if there are pending tool calls that we couldn't execute
			// If assistantMessage still has tool calls, it means we hit the limit with pending work
			if len(assistantMessage.ToolCalls) > 0 {
				log.Warn().
					Str("request_id", requestID).
					Int("rounds", round).
					Int("pending_tool_calls", len(assistantMessage.ToolCalls)).
					Msg("Maximum tool rounds reached with pending tool calls")
				// Return clear, user-facing error
				return createErrorResponse(
					"max_tool_rounds_exceeded",
					fmt.Sprintf("Maximum tool execution rounds (%d) exceeded. This request completed %d rounds with %d pending tool calls.", maxRounds, round, len(assistantMessage.ToolCalls)),
					requestID,
				), nil
			}
			// If no tool calls, we completed naturally at the limit - this is fine
			// The loop will exit naturally and we'll return the final response below
		}
	}

	// If we exit the loop naturally (no more tool calls), return the final response
	// This handles the case where we completed exactly maxRounds rounds with no pending tool calls

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

	// Execute tool with timeout and propagate request ID
	timeout := p.guardrails.GetTimeout()
	toolCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	toolCtx = middleware.WithRequestID(toolCtx, requestID)

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
