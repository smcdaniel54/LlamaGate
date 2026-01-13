package proxy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/llamagate/llamagate/internal/tools"
)

// TestToolLoop_MaxRoundsLimit tests that maximum tool rounds limit validation is called
// Note: Full integration test requires MCP tools to be registered, which is complex to mock.
// This test validates that the guardrails limit exists and can be enforced.
func TestToolLoop_MaxRoundsLimit(t *testing.T) {
	// Test that guardrails has max rounds limit
	guardrails, err := tools.NewGuardrails(
		[]string{},
		[]string{},
		3,  // max tool rounds
		10, // max calls per round
		50, // max total tool calls
		30*time.Second,
		1024*1024,
	)
	require.NoError(t, err)

	// Validate that the limit is enforced
	assert.NoError(t, guardrails.ValidateToolRounds(2))
	assert.Error(t, guardrails.ValidateToolRounds(3))
	assert.Error(t, guardrails.ValidateToolRounds(4))

	// Verify error message is user-facing
	err = guardrails.ValidateToolRounds(3)
	assert.Contains(t, err.Error(), "maximum tool rounds")
	assert.Contains(t, err.Error(), "3")
}

// TestToolLoop_MaxTotalToolCallsLimit tests that maximum total tool calls limit validation is called
func TestToolLoop_MaxTotalToolCallsLimit(t *testing.T) {
	// Test that guardrails has max total tool calls limit
	guardrails, err := tools.NewGuardrails(
		[]string{},
		[]string{},
		10, // max rounds
		10, // max calls per round
		5,  // max total tool calls (low to test limit)
		30*time.Second,
		1024*1024,
	)
	require.NoError(t, err)

	// Validate that the limit is enforced
	assert.NoError(t, guardrails.ValidateTotalToolCalls(4))
	assert.Error(t, guardrails.ValidateTotalToolCalls(5))
	assert.Error(t, guardrails.ValidateTotalToolCalls(6))

	// Verify error message is user-facing
	err = guardrails.ValidateTotalToolCalls(5)
	assert.Contains(t, err.Error(), "maximum total tool calls")
	assert.Contains(t, err.Error(), "5")
}

// TestToolLoop_MaxCallsPerRoundLimit tests that maximum tool calls per round limit validation is called
func TestToolLoop_MaxCallsPerRoundLimit(t *testing.T) {
	// Test that guardrails has max calls per round limit
	guardrails, err := tools.NewGuardrails(
		[]string{},
		[]string{},
		10, // max rounds
		10, // max calls per round
		50, // max total tool calls
		30*time.Second,
		1024*1024,
	)
	require.NoError(t, err)

	// Validate that the limit is enforced
	assert.NoError(t, guardrails.ValidateToolCallsPerRound(10))
	assert.Error(t, guardrails.ValidateToolCallsPerRound(11))
	assert.Error(t, guardrails.ValidateToolCallsPerRound(15))

	// Verify error message is user-facing
	err = guardrails.ValidateToolCallsPerRound(11)
	assert.Contains(t, err.Error(), "maximum tool calls per round")
	assert.Contains(t, err.Error(), "10")
}

// TestToolLoop_ErrorMessagesAreUserFacing tests that error messages are clear and user-facing
func TestToolLoop_ErrorMessagesAreUserFacing(t *testing.T) {
	guardrails, err := tools.NewGuardrails(
		[]string{},
		[]string{},
		2,  // low max rounds
		10, // max calls per round
		5,  // low max total calls
		30*time.Second,
		1024*1024,
	)
	require.NoError(t, err)

	// Test max rounds error message
	err = guardrails.ValidateToolRounds(2)
	require.Error(t, err)
	message := err.Error()
	assert.Contains(t, message, "maximum")
	assert.Contains(t, message, "tool rounds")
	assert.Contains(t, message, "2")
	assert.NotContains(t, message, "failed")
	assert.NotContains(t, message, "internal")
	assert.Greater(t, len(message), 10, "Error message should be descriptive")

	// Test max calls per round error message
	err = guardrails.ValidateToolCallsPerRound(11)
	require.Error(t, err)
	message = err.Error()
	assert.Contains(t, message, "maximum")
	assert.Contains(t, message, "tool calls per round")
	assert.Contains(t, message, "10")
	assert.NotContains(t, message, "failed")
	assert.NotContains(t, message, "internal")
	assert.Greater(t, len(message), 10, "Error message should be descriptive")

	// Test max total tool calls error message
	err = guardrails.ValidateTotalToolCalls(5)
	require.Error(t, err)
	message = err.Error()
	assert.Contains(t, message, "maximum")
	assert.Contains(t, message, "total tool calls")
	assert.Contains(t, message, "5")
	assert.NotContains(t, message, "failed")
	assert.NotContains(t, message, "internal")
	assert.Greater(t, len(message), 10, "Error message should be descriptive")
}
