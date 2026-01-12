package tools

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGuardrails_ValidateToolCall(t *testing.T) {
	tests := []struct {
		name        string
		allowTools  []string
		denyTools   []string
		toolName    string
		expectError bool
	}{
		{
			name:        "no restrictions",
			allowTools:  []string{},
			denyTools:   []string{},
			toolName:    "mcp.server.tool",
			expectError: false,
		},
		{
			name:        "allowed by pattern",
			allowTools:  []string{"mcp.server.*"},
			denyTools:   []string{},
			toolName:    "mcp.server.tool",
			expectError: false,
		},
		{
			name:        "denied by pattern",
			allowTools:  []string{},
			denyTools:   []string{"mcp.server.*"},
			toolName:    "mcp.server.tool",
			expectError: true,
		},
		{
			name:        "not in allow list",
			allowTools:  []string{"mcp.other.*"},
			denyTools:   []string{},
			toolName:    "mcp.server.tool",
			expectError: true,
		},
		{
			name:        "deny takes precedence",
			allowTools:  []string{"mcp.server.*"},
			denyTools:   []string{"mcp.server.dangerous"},
			toolName:    "mcp.server.dangerous",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guardrails, err := NewGuardrails(
				tt.allowTools,
				tt.denyTools,
				10,
				10,
				50,
				30*time.Second,
				1024*1024,
			)
			require.NoError(t, err)

			err = guardrails.ValidateToolCall(tt.toolName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGuardrails_ValidateToolRounds(t *testing.T) {
	guardrails, err := NewGuardrails(
		[]string{},
		[]string{},
		5,
		10,
		50,
		30*time.Second,
		1024*1024,
	)
	require.NoError(t, err)

	assert.NoError(t, guardrails.ValidateToolRounds(4))
	assert.Error(t, guardrails.ValidateToolRounds(5))
	assert.Error(t, guardrails.ValidateToolRounds(10))
}

func TestGuardrails_ValidateToolCallsPerRound(t *testing.T) {
	guardrails, err := NewGuardrails(
		[]string{},
		[]string{},
		10,
		5,
		50,
		30*time.Second,
		1024*1024,
	)
	require.NoError(t, err)

	assert.NoError(t, guardrails.ValidateToolCallsPerRound(5))
	assert.Error(t, guardrails.ValidateToolCallsPerRound(6))
}

func TestGuardrails_TruncateResult(t *testing.T) {
	guardrails, err := NewGuardrails(
		[]string{},
		[]string{},
		10,
		10,
		50,
		30*time.Second,
		100, // 100 bytes max
	)
	require.NoError(t, err)

	smallResult := "small result"
	assert.Equal(t, smallResult, guardrails.TruncateResult(smallResult))

	largeResult := make([]byte, 200)
	for i := range largeResult {
		largeResult[i] = 'a'
	}
	truncated := guardrails.TruncateResult(string(largeResult))
	assert.LessOrEqual(t, len(truncated), 120) // 100 + some overhead for truncation marker
	assert.Contains(t, truncated, "[truncated]")
}

func TestGuardrails_GetTimeout(t *testing.T) {
	timeout := 45 * time.Second
	guardrails, err := NewGuardrails(
		[]string{},
		[]string{},
		10,
		10,
		50,
		timeout,
		1024*1024,
	)
	require.NoError(t, err)

	assert.Equal(t, timeout, guardrails.GetTimeout())
}

func TestGuardrails_ValidateTotalToolCalls(t *testing.T) {
	guardrails, err := NewGuardrails(
		[]string{},
		[]string{},
		10,
		10,
		50, // max total tool calls
		30*time.Second,
		1024*1024,
	)
	require.NoError(t, err)

	// Should pass for values below limit
	assert.NoError(t, guardrails.ValidateTotalToolCalls(49))
	assert.NoError(t, guardrails.ValidateTotalToolCalls(0))
	assert.NoError(t, guardrails.ValidateTotalToolCalls(25))

	// Should fail for values at or above limit
	assert.Error(t, guardrails.ValidateTotalToolCalls(50))
	assert.Error(t, guardrails.ValidateTotalToolCalls(51))
	assert.Error(t, guardrails.ValidateTotalToolCalls(100))

	// Verify error message
	err = guardrails.ValidateTotalToolCalls(50)
	assert.Contains(t, err.Error(), "maximum total tool calls")
	assert.Contains(t, err.Error(), "50")
}

func TestGuardrails_MaxTotalToolCalls(t *testing.T) {
	guardrails, err := NewGuardrails(
		[]string{},
		[]string{},
		10,
		10,
		75, // max total tool calls
		30*time.Second,
		1024*1024,
	)
	require.NoError(t, err)

	assert.Equal(t, 75, guardrails.MaxTotalToolCalls())
}
