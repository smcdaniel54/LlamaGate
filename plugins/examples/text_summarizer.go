package examples

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/llamagate/llamagate/internal/plugins"
)

// TextSummarizerPlugin is an example plugin that summarizes text
// This demonstrates a simple, practical use case
type TextSummarizerPlugin struct {
	// Plugin-specific configuration can be added here
	maxLength int
}

// Metadata returns plugin metadata
func (p *TextSummarizerPlugin) Metadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "text_summarizer",
		Version:     "1.0.0",
		Description: "Summarizes text content to a specified maximum length",
		Author:      "LlamaGate Team",

		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text": map[string]interface{}{
					"type":        "string",
					"description": "The text content to summarize",
				},
				"max_length": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum length of the summary (in characters)",
					"default":     200,
					"minimum":     50,
					"maximum":     1000,
				},
				"style": map[string]interface{}{
					"type":        "string",
					"description": "Summary style: 'brief', 'detailed', or 'bullet'",
					"enum":        []string{"brief", "detailed", "bullet"},
					"default":     "brief",
				},
			},
			"required": []string{"text"},
		},

		OutputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"summary": map[string]interface{}{
					"type":        "string",
					"description": "The generated summary",
				},
				"original_length": map[string]interface{}{
					"type":        "integer",
					"description": "Length of the original text",
				},
				"summary_length": map[string]interface{}{
					"type":        "integer",
					"description": "Length of the generated summary",
				},
				"compression_ratio": map[string]interface{}{
					"type":        "number",
					"description": "Ratio of summary length to original length",
				},
			},
		},

		RequiredInputs: []string{"text"},

		OptionalInputs: map[string]interface{}{
			"max_length": 200,
			"style":      "brief",
		},
	}
}

// ValidateInput validates the input parameters
func (p *TextSummarizerPlugin) ValidateInput(input map[string]interface{}) error {
	// Check required inputs
	if text, exists := input["text"]; !exists {
		return fmt.Errorf("required input 'text' is missing")
	} else if textStr, ok := text.(string); !ok {
		return fmt.Errorf("input 'text' must be a string")
	} else if len(textStr) == 0 {
		return fmt.Errorf("input 'text' cannot be empty")
	}

	// Validate max_length if provided
	if maxLength, exists := input["max_length"]; exists {
		if maxLengthInt, ok := maxLength.(float64); ok {
			if maxLengthInt < 50 || maxLengthInt > 1000 {
				return fmt.Errorf("max_length must be between 50 and 1000")
			}
		} else {
			return fmt.Errorf("max_length must be a number")
		}
	}

	// Validate style if provided
	if style, exists := input["style"]; exists {
		styleStr, ok := style.(string)
		if !ok {
			return fmt.Errorf("style must be a string")
		}
		validStyles := []string{"brief", "detailed", "bullet"}
		valid := false
		for _, vs := range validStyles {
			if vs == styleStr {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("style must be one of: %v", validStyles)
		}
	}

	return nil
}

// Execute runs the text summarization workflow
func (p *TextSummarizerPlugin) Execute(_ context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
	startTime := time.Now()

	// Extract and validate inputs
	text, _ := input["text"].(string)
	maxLength := 200
	if ml, exists := input["max_length"]; exists {
		if mlFloat, ok := ml.(float64); ok {
			maxLength = int(mlFloat)
		}
	}
	style := "brief"
	if s, exists := input["style"]; exists {
		style, _ = s.(string)
	}

	// Workflow Step 1: Preprocess text
	processedText := p.preprocessText(text)

	// Workflow Step 2: Extract key sentences
	keySentences := p.extractKeySentences(processedText, maxLength, style)

	// Workflow Step 3: Format summary based on style
	summary := p.formatSummary(keySentences, style, maxLength)

	// Calculate metrics
	originalLength := len(text)
	summaryLength := len(summary)
	compressionRatio := float64(summaryLength) / float64(originalLength)

	// Build result
	result := map[string]interface{}{
		"summary":           summary,
		"original_length":   originalLength,
		"summary_length":    summaryLength,
		"compression_ratio": compressionRatio,
	}

	executionTime := time.Since(startTime)

	return &plugins.PluginResult{
		Success: true,
		Data:    result,
		Metadata: plugins.ExecutionMetadata{
			ExecutionTime: executionTime,
			StepsExecuted: 3, // Preprocess, Extract, Format
			Timestamp:     time.Now(),
		},
	}, nil
}

// preprocessText cleans and prepares text for summarization
func (p *TextSummarizerPlugin) preprocessText(text string) string {
	// Remove extra whitespace
	text = strings.TrimSpace(text)
	text = strings.Join(strings.Fields(text), " ")
	return text
}

// extractKeySentences extracts the most important sentences
func (p *TextSummarizerPlugin) extractKeySentences(text string, maxLength int, _ string) []string {
	// Split into sentences (simple approach - can be enhanced)
	sentences := strings.Split(text, ".")

	// Filter out empty sentences
	filtered := make([]string, 0)
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if len(s) > 0 {
			filtered = append(filtered, s)
		}
	}

	// Simple extraction: take first sentences that fit within maxLength
	// In a real implementation, this would use more sophisticated algorithms
	selected := make([]string, 0)
	currentLength := 0

	for _, sentence := range filtered {
		sentenceLength := len(sentence) + 1 // +1 for period
		if currentLength+sentenceLength <= maxLength {
			selected = append(selected, sentence)
			currentLength += sentenceLength
		} else {
			break
		}
	}

	return selected
}

// formatSummary formats the summary based on the requested style
func (p *TextSummarizerPlugin) formatSummary(sentences []string, style string, maxLength int) string {
	switch style {
	case "bullet":
		// Format as bullet points
		var builder strings.Builder
		for _, sentence := range sentences {
			builder.WriteString("• ")
			builder.WriteString(sentence)
			builder.WriteString("\n")
		}
		return strings.TrimSpace(builder.String())

	case "detailed":
		// Join with periods and proper spacing
		return strings.Join(sentences, ". ") + "."

	case "brief":
		fallthrough
	default:
		// Join with periods, more compact
		result := strings.Join(sentences, ". ") + "."
		// Truncate if still too long
		if len(result) > maxLength {
			result = result[:maxLength-3] + "..."
		}
		return result
	}
}

// NewTextSummarizerPlugin creates a new text summarizer plugin
func NewTextSummarizerPlugin() plugins.Plugin {
	return &TextSummarizerPlugin{
		maxLength: 200,
	}
}
