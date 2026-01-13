// Package plugins provides plugin implementations for LlamaGate.
package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/plugins"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// AlexaSkillPlugin handles Alexa Skill requests and integrates with LlamaGate LLM
type AlexaSkillPlugin struct {
	wakeWord        string
	caseSensitive   bool
	removeFromQuery bool
	defaultModel    string
	registry        *plugins.Registry // Reference to registry to get context
}

// NewAlexaSkillPlugin creates a new Alexa Skill plugin
func NewAlexaSkillPlugin() *AlexaSkillPlugin {
	return &AlexaSkillPlugin{
		wakeWord:        "Smart Voice",
		caseSensitive:   false,
		removeFromQuery: true,
		defaultModel:    "llama3.2",
	}
}

// NewAlexaSkillPluginWithConfig creates a new Alexa Skill plugin with configuration
func NewAlexaSkillPluginWithConfig(config map[string]interface{}) *AlexaSkillPlugin {
	plugin := &AlexaSkillPlugin{
		wakeWord:        "Smart Voice",
		caseSensitive:   false,
		removeFromQuery: true,
		defaultModel:    "llama3.2",
	}

	// Apply configuration
	if wakeWord, ok := config["wake_word"].(string); ok && wakeWord != "" {
		plugin.wakeWord = wakeWord
	}
	if caseSensitive, ok := config["case_sensitive"].(bool); ok {
		plugin.caseSensitive = caseSensitive
	}
	if removeFromQuery, ok := config["remove_from_query"].(bool); ok {
		plugin.removeFromQuery = removeFromQuery
	}
	if defaultModel, ok := config["default_model"].(string); ok && defaultModel != "" {
		plugin.defaultModel = defaultModel
	}

	return plugin
}

// SetRegistry sets the registry reference (called during registration)
func (p *AlexaSkillPlugin) SetRegistry(registry *plugins.Registry) {
	p.registry = registry
}

// Metadata returns plugin metadata
func (p *AlexaSkillPlugin) Metadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "alexa_skill",
		Version:     "1.0.0",
		Description: "Handles Alexa Skill requests with wake word detection and LLM integration",
		Author:      "Smart Voice Team",

		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"version": map[string]interface{}{
					"type":        "string",
					"description": "Alexa request version",
				},
				"session": map[string]interface{}{
					"type":        "object",
					"description": "Alexa session information",
				},
				"request": map[string]interface{}{
					"type":        "object",
					"description": "Alexa request body",
				},
			},
			"required": []string{"version", "session", "request"},
		},

		OutputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"version": map[string]interface{}{
					"type":        "string",
					"description": "Alexa response version",
				},
				"response": map[string]interface{}{
					"type":        "object",
					"description": "Alexa response body",
				},
			},
		},

		RequiredInputs: []string{"version", "session", "request"},
	}
}

// ValidateInput validates the Alexa request
func (p *AlexaSkillPlugin) ValidateInput(input map[string]interface{}) error {
	if _, exists := input["version"]; !exists {
		return fmt.Errorf("required input 'version' is missing")
	}
	if _, exists := input["session"]; !exists {
		return fmt.Errorf("required input 'session' is missing")
	}
	if _, exists := input["request"]; !exists {
		return fmt.Errorf("required input 'request' is missing")
	}
	return nil
}

// Execute processes the Alexa request (used for standard plugin execution)
func (p *AlexaSkillPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
	startTime := time.Now()

	// Convert input to Alexa request format
	alexaReq, err := p.parseAlexaRequest(input)
	if err != nil {
		return &plugins.PluginResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse Alexa request: %v", err),
			Metadata: plugins.ExecutionMetadata{
				ExecutionTime: time.Since(startTime),
				StepsExecuted: 0,
				Timestamp:     time.Now(),
			},
		}, nil
	}

	// Process the request
	alexaResp, err := p.processAlexaRequest(ctx, alexaReq)
	if err != nil {
		return &plugins.PluginResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to process Alexa request: %v", err),
			Metadata: plugins.ExecutionMetadata{
				ExecutionTime: time.Since(startTime),
				StepsExecuted: 1,
				Timestamp:     time.Now(),
			},
		}, nil
	}

	// Convert response to map
	responseMap := map[string]interface{}{
		"version":  alexaResp.Version,
		"response": alexaResp.Response,
	}
	if alexaResp.SessionAttributes != nil {
		responseMap["sessionAttributes"] = alexaResp.SessionAttributes
	}

	return &plugins.PluginResult{
		Success: true,
		Data:    responseMap,
		Metadata: plugins.ExecutionMetadata{
			ExecutionTime: time.Since(startTime),
			StepsExecuted: 1,
			Timestamp:     time.Now(),
		},
	}, nil
}

// GetAgentDefinition returns nil (this plugin is not an agent)
func (p *AlexaSkillPlugin) GetAgentDefinition() *plugins.AgentDefinition {
	return nil
}

// GetAPIEndpoints returns custom API endpoints for this plugin
func (p *AlexaSkillPlugin) GetAPIEndpoints() []plugins.APIEndpoint {
	return []plugins.APIEndpoint{
		{
			Path:        "/alexa",
			Method:      "POST",
			Description: "Handle Alexa Skill requests",
			Handler:     p.handleAlexaEndpoint,
			RequestSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version": map[string]interface{}{
						"type": "string",
					},
					"session": map[string]interface{}{
						"type": "object",
					},
					"request": map[string]interface{}{
						"type": "object",
					},
				},
				"required": []string{"version", "session", "request"},
			},
			RequiresAuth:      false, // Alexa doesn't use API keys
			RequiresRateLimit: true,
		},
	}
}

// handleAlexaEndpoint handles POST /v1/plugins/alexa_skill/alexa requests
func (p *AlexaSkillPlugin) handleAlexaEndpoint(c *gin.Context) {
	// Get plugin context for logging
	var pluginCtx *plugins.PluginContext
	if p.registry != nil {
		pluginCtx = p.registry.GetContext("alexa_skill")
	}

	var logger *zerolog.Logger
	if pluginCtx != nil {
		logger = &pluginCtx.Logger
	} else {
		baseLogger := log.With().Str("plugin", "alexa_skill").Logger()
		logger = &baseLogger
	}

	var alexaReq AlexaRequest
	if err := c.ShouldBindJSON(&alexaReq); err != nil {
		logger.Error().Err(err).Msg("Failed to parse Alexa request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Process the request
	alexaResp, err := p.processAlexaRequest(c.Request.Context(), &alexaReq)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to process Alexa request")
		// Return error response in Alexa format
		alexaResp = p.createErrorResponse("I'm sorry, I encountered an error processing your request.")
	}

	c.JSON(http.StatusOK, alexaResp)
}

// AlexaRequest represents an incoming Alexa request
type AlexaRequest struct {
	Version string                 `json:"version"`
	Session map[string]interface{} `json:"session"`
	Request map[string]interface{} `json:"request"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// AlexaResponse represents the response to Alexa
type AlexaResponse struct {
	Version           string                 `json:"version"`
	SessionAttributes map[string]interface{} `json:"sessionAttributes,omitempty"`
	Response          AlexaResponseBody      `json:"response"`
}

// AlexaResponseBody contains the response details
type AlexaResponseBody struct {
	OutputSpeech     *AlexaOutputSpeech `json:"outputSpeech,omitempty"`
	Reprompt         *AlexaReprompt     `json:"reprompt,omitempty"`
	Directives       []interface{}      `json:"directives,omitempty"`
	ShouldEndSession bool               `json:"shouldEndSession"`
	Card             *AlexaCard         `json:"card,omitempty"`
}

// AlexaOutputSpeech contains speech output
type AlexaOutputSpeech struct {
	Type string `json:"type"` // PlainText or SSML
	Text string `json:"text,omitempty"`
	SSML string `json:"ssml,omitempty"`
}

// AlexaReprompt contains reprompt information
type AlexaReprompt struct {
	OutputSpeech AlexaOutputSpeech `json:"outputSpeech"`
}

// AlexaCard contains card information
type AlexaCard struct {
	Type    string `json:"type"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Text    string `json:"text,omitempty"`
}

// parseAlexaRequest converts input map to AlexaRequest
func (p *AlexaSkillPlugin) parseAlexaRequest(input map[string]interface{}) (*AlexaRequest, error) {
	// Convert to JSON and back to ensure proper structure
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var alexaReq AlexaRequest
	if err := json.Unmarshal(jsonData, &alexaReq); err != nil {
		return nil, err
	}

	return &alexaReq, nil
}

// processAlexaRequest processes an Alexa request and returns a response.
// Errors are handled internally and converted to error responses, so this always returns nil error.
func (p *AlexaSkillPlugin) processAlexaRequest(ctx context.Context, alexaReq *AlexaRequest) (*AlexaResponse, error) { //nolint:unparam // Error is always nil by design
	// Extract query text from Alexa request
	query := p.extractQueryText(alexaReq)

	if query == "" {
		return p.createErrorResponse("I didn't understand your request."), nil
	}

	// Detect wake word
	found, processedQuery := p.detectWakeWord(query)

	if !found {
		// No wake word detected - return default response
		return p.createResponse("I heard your request, but I'm not configured to handle it yet.", true), nil
	}

	// Process through LLM
	responseText, err := p.processWithLLM(ctx, processedQuery)
	if err != nil {
		// Log error using plugin context if available
		if p.registry != nil {
			if pluginCtx := p.registry.GetContext("alexa_skill"); pluginCtx != nil {
				pluginCtx.Logger.Error().Err(err).Str("query", processedQuery).Msg("LLM processing failed")
			} else {
				log.Error().Err(err).Str("query", processedQuery).Msg("LLM processing failed")
			}
		} else {
			log.Error().Err(err).Str("query", processedQuery).Msg("LLM processing failed")
		}
		return p.createErrorResponse("I'm sorry, I encountered an error processing your request."), nil
	}

	// Format Alexa response
	return p.createResponse(responseText, true), nil
}

// extractQueryText extracts the query text from an Alexa request
func (p *AlexaSkillPlugin) extractQueryText(alexaReq *AlexaRequest) string {
	request, ok := alexaReq.Request["intent"].(map[string]interface{})
	if !ok {
		return ""
	}

	intentName, _ := request["name"].(string)
	if intentName == "" {
		return ""
	}

	// Try to get query from intent slots
	slots, ok := request["slots"].(map[string]interface{})
	if ok {
		// Try "query" slot
		if querySlot, ok := slots["query"].(map[string]interface{}); ok {
			if value, ok := querySlot["value"].(string); ok && value != "" {
				return value
			}
		}
		// Try "Query" slot (capitalized)
		if querySlot, ok := slots["Query"].(map[string]interface{}); ok {
			if value, ok := querySlot["value"].(string); ok && value != "" {
				return value
			}
		}
	}

	// Fall back to intent name
	return intentName
}

// detectWakeWord detects the wake word in query text
func (p *AlexaSkillPlugin) detectWakeWord(query string) (bool, string) {
	if query == "" {
		return false, query
	}

	// Normalize for comparison
	normalizedQuery := query
	normalizedWakeWord := p.wakeWord

	if !p.caseSensitive {
		normalizedQuery = strings.ToLower(query)
		normalizedWakeWord = strings.ToLower(p.wakeWord)
	}

	// Generate variations of the wake word
	variations := p.generateWakeWordVariations(normalizedWakeWord)

	// Check if any variation is present
	var matchedVariation string
	for _, variation := range variations {
		if strings.Contains(normalizedQuery, variation) {
			matchedVariation = variation
			break
		}
	}

	if matchedVariation == "" {
		return false, query
	}

	// Remove wake word from query if configured
	if p.removeFromQuery {
		idx := strings.Index(normalizedQuery, matchedVariation)
		if idx >= 0 {
			// Remove wake word and clean up spaces
			before := strings.TrimSpace(query[:idx])
			after := strings.TrimSpace(query[idx+len(matchedVariation):])

			// Combine remaining parts
			var processed string
			if before != "" && after != "" {
				processed = before + " " + after
			} else if before != "" {
				processed = before
			} else if after != "" {
				processed = after
			}

			// Clean up multiple spaces
			processed = strings.Join(strings.Fields(processed), " ")
			return true, processed
		}
	}

	return true, query
}

// generateWakeWordVariations creates variations of the wake word
func (p *AlexaSkillPlugin) generateWakeWordVariations(wakeWord string) []string {
	variations := []string{wakeWord} // Always include the original

	// If wake word contains space, add version without space
	if strings.Contains(wakeWord, " ") {
		noSpace := strings.ReplaceAll(wakeWord, " ", "")
		variations = append(variations, noSpace)
	}

	// If wake word doesn't contain space, try to add version with space
	if !strings.Contains(wakeWord, " ") && len(wakeWord) > 0 {
		var withSpace strings.Builder
		for i, r := range wakeWord {
			if i > 0 && r >= 'A' && r <= 'Z' {
				withSpace.WriteRune(' ')
			}
			withSpace.WriteRune(r)
		}
		if withSpace.String() != wakeWord {
			variations = append(variations, withSpace.String())
		}
	}

	return variations
}

// processWithLLM processes the query through LlamaGate's LLM
func (p *AlexaSkillPlugin) processWithLLM(ctx context.Context, query string) (string, error) {
	// Get plugin context from registry
	if p.registry == nil {
		return "", fmt.Errorf("plugin registry not available")
	}

	pluginCtx := p.registry.GetContext("alexa_skill")
	if pluginCtx == nil {
		return "", fmt.Errorf("plugin context not available")
	}

	// Use LLM handler from context
	messages := []map[string]interface{}{
		{
			"role":    "user",
			"content": query,
		},
	}

	options := map[string]interface{}{
		"temperature": 0.7,
	}

	response, err := pluginCtx.CallLLM(ctx, p.defaultModel, messages, options)
	if err != nil {
		pluginCtx.Logger.Error().Err(err).Str("query", query).Msg("LLM call failed")
		return "", fmt.Errorf("LLM processing failed: %w", err)
	}

	pluginCtx.Logger.Info().Str("query", query).Str("model", p.defaultModel).Msg("LLM call successful")
	return response, nil
}

// createResponse creates an Alexa response
func (p *AlexaSkillPlugin) createResponse(text string, shouldEndSession bool) *AlexaResponse {
	return &AlexaResponse{
		Version: "1.0",
		Response: AlexaResponseBody{
			OutputSpeech: &AlexaOutputSpeech{
				Type: "PlainText",
				Text: text,
			},
			ShouldEndSession: shouldEndSession,
		},
	}
}

// createErrorResponse creates an error response for Alexa
func (p *AlexaSkillPlugin) createErrorResponse(message string) *AlexaResponse {
	return p.createResponse(message, true)
}
