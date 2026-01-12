package tools

import (
	"fmt"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"github.com/rs/zerolog/log"
)

// Guardrails enforces security and performance constraints on tool usage
type Guardrails struct {
	allowPatterns    []glob.Glob
	denyPatterns     []glob.Glob
	maxToolRounds    int
	maxCallsPerRound int
	maxTotalCalls    int // Maximum total tool calls across all rounds
	defaultTimeout   time.Duration
	maxResultSize    int64
}

// NewGuardrails creates a new guardrails instance
func NewGuardrails(
	allowTools []string,
	denyTools []string,
	maxToolRounds int,
	maxCallsPerRound int,
	maxTotalCalls int,
	defaultTimeout time.Duration,
	maxResultSize int64,
) (*Guardrails, error) {
	g := &Guardrails{
		maxToolRounds:    maxToolRounds,
		maxCallsPerRound: maxCallsPerRound,
		maxTotalCalls:    maxTotalCalls,
		defaultTimeout:   defaultTimeout,
		maxResultSize:    maxResultSize,
	}

	// Compile allow patterns
	for _, pattern := range allowTools {
		compiled, err := glob.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid allow pattern %s: %w", pattern, err)
		}
		g.allowPatterns = append(g.allowPatterns, compiled)
	}

	// Compile deny patterns
	for _, pattern := range denyTools {
		compiled, err := glob.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid deny pattern %s: %w", pattern, err)
		}
		g.denyPatterns = append(g.denyPatterns, compiled)
	}

	return g, nil
}

// ValidateToolCall validates whether a tool call is allowed
func (g *Guardrails) ValidateToolCall(toolName string) error {
	// Check deny list first (deny takes precedence)
	for _, pattern := range g.denyPatterns {
		if pattern.Match(toolName) {
			return fmt.Errorf("tool %s is denied by pattern", toolName)
		}
	}

	// If allow list is empty, allow all (except denied)
	if len(g.allowPatterns) == 0 {
		return nil
	}

	// Check allow list
	for _, pattern := range g.allowPatterns {
		if pattern.Match(toolName) {
			return nil
		}
	}

	return fmt.Errorf("tool %s is not in allow list", toolName)
}

// ValidateToolRounds validates the number of tool execution rounds
func (g *Guardrails) ValidateToolRounds(rounds int) error {
	if rounds >= g.maxToolRounds {
		return fmt.Errorf("maximum tool rounds (%d) exceeded", g.maxToolRounds)
	}
	return nil
}

// ValidateToolCallsPerRound validates the number of tool calls in a round
func (g *Guardrails) ValidateToolCallsPerRound(count int) error {
	if count > g.maxCallsPerRound {
		return fmt.Errorf("maximum tool calls per round (%d) exceeded", g.maxCallsPerRound)
	}
	return nil
}

// ValidateTotalToolCalls validates the total number of tool calls across all rounds
func (g *Guardrails) ValidateTotalToolCalls(totalCount int) error {
	if totalCount >= g.maxTotalCalls {
		return fmt.Errorf("maximum total tool calls (%d) exceeded", g.maxTotalCalls)
	}
	return nil
}

// MaxTotalToolCalls returns the maximum total tool calls allowed
func (g *Guardrails) MaxTotalToolCalls() int {
	return g.maxTotalCalls
}

// GetTimeout returns the timeout for tool execution
func (g *Guardrails) GetTimeout() time.Duration {
	return g.defaultTimeout
}

// MaxToolRounds returns the maximum number of tool execution rounds
func (g *Guardrails) MaxToolRounds() int {
	return g.maxToolRounds
}

// MaxCallsPerRound returns the maximum number of tool calls per round
func (g *Guardrails) MaxCallsPerRound() int {
	return g.maxCallsPerRound
}

// TruncateResult safely truncates a tool result to the maximum size
func (g *Guardrails) TruncateResult(result string) string {
	if int64(len(result)) <= g.maxResultSize {
		return result
	}

	// Truncate and add indicator
	truncated := result[:g.maxResultSize]

	// Try to truncate at a safe boundary (end of a line or JSON boundary)
	// For simplicity, we'll just truncate and add a marker
	// In production, you might want to be smarter about JSON truncation
	lastNewline := strings.LastIndex(truncated, "\n")
	if lastNewline > int(g.maxResultSize*9/10) {
		truncated = truncated[:lastNewline]
	}

	truncated += "\n... [truncated]"

	log.Warn().
		Int64("original_size", int64(len(result))).
		Int64("max_size", g.maxResultSize).
		Msg("Tool result truncated")

	return truncated
}

// RedactSensitiveData redacts potentially sensitive data from logs
func RedactSensitiveData(data interface{}) interface{} {
	// Simple redaction - in production, you might want more sophisticated logic
	// This is a placeholder that can be enhanced
	switch v := data.(type) {
	case string:
		// Redact if it looks like a token, API key, or password
		if len(v) > 20 && (strings.Contains(v, "sk-") || strings.Contains(v, "token") || strings.Contains(v, "password")) {
			return "[REDACTED]"
		}
		return v
	case map[string]interface{}:
		redacted := make(map[string]interface{})
		for k, val := range v {
			// Redact common sensitive keys
			if strings.Contains(strings.ToLower(k), "password") ||
				strings.Contains(strings.ToLower(k), "token") ||
				strings.Contains(strings.ToLower(k), "secret") ||
				strings.Contains(strings.ToLower(k), "key") {
				redacted[k] = "[REDACTED]"
			} else {
				redacted[k] = RedactSensitiveData(val)
			}
		}
		return redacted
	case []interface{}:
		redacted := make([]interface{}, len(v))
		for i, item := range v {
			redacted[i] = RedactSensitiveData(item)
		}
		return redacted
	default:
		return v
	}
}
