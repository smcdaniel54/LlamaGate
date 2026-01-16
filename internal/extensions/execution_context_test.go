package extensions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutionContext_WithChild(t *testing.T) {
	ctx := context.Background()
	execCtx := NewExecutionContext(ctx, "test-trace-123", "test.yaml")

	// Test normal child creation
	child, err := execCtx.WithChild("child.yaml")
	require.NoError(t, err)
	assert.Equal(t, 1, child.CallDepth)
	assert.Equal(t, 99, child.CallBudget)
	assert.Equal(t, "test-trace-123", child.TraceID)

	// Test max depth limit
	deepCtx := execCtx
	for i := 0; i < 10; i++ {
		var err error
		deepCtx, err = deepCtx.WithChild("test.yaml")
		require.NoError(t, err)
	}

	// Next call should fail
	_, err = deepCtx.WithChild("test.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum call depth exceeded")

	// Test call budget exhaustion
	budgetCtx := NewExecutionContext(ctx, "test", "test.yaml")
	budgetCtx.CallBudget = 1

	child1, err := budgetCtx.WithChild("test.yaml")
	require.NoError(t, err)
	assert.Equal(t, 0, child1.CallBudget)

	// Next call should fail
	_, err = child1.WithChild("test.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "call budget exhausted")

	// Test runtime limit
	runtimeCtx := NewExecutionContext(ctx, "test", "test.yaml")
	runtimeCtx.StartTime = time.Now().Add(-6 * time.Minute) // Started 6 minutes ago
	runtimeCtx.MaxRuntime = 5 * time.Minute

	_, err = runtimeCtx.WithChild("test.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum runtime exceeded")
}

func TestExecutionError(t *testing.T) {
	err := &ExecutionError{
		Extension:    "test-extension",
		Step:         "validation",
		ManifestPath: "extensions/test/manifest.yaml",
		Message:      "required input missing",
		Details:      "input 'x' is required",
	}

	errMsg := err.Error()
	assert.Contains(t, errMsg, "test-extension")
	assert.Contains(t, errMsg, "validation")
	assert.Contains(t, errMsg, "manifest.yaml")
	assert.Contains(t, errMsg, "required input missing")
}
