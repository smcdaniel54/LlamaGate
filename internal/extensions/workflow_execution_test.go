package extensions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteExtensionInternal_Guardrails(t *testing.T) {
	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": "test",
					},
				},
			},
		}, nil
	}

	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	// Register a simple extension
	manifest := &Manifest{
		Name:        "test-ext",
		Version:     "1.0.0",
		Description: "Test",
		Type:        "workflow",
		Steps: []WorkflowStep{
			{
				Uses: "llm.chat",
				With: map[string]interface{}{
					"prompt": "test prompt",
				},
			},
		},
	}
	require.NoError(t, registry.Register(manifest))

	// Test max depth guard
	execCtx := NewExecutionContext(context.Background(), "test", "test.yaml")
	execCtx.MaxDepth = 1 // Set low limit

	// First call should work
	result1, err := executor.ExecuteExtensionInternal(execCtx, "test-ext", map[string]interface{}{}, "caller")
	require.NoError(t, err)
	assert.NotNil(t, result1)

	// Create child context (depth 1)
	childCtx, err := execCtx.WithChild("test.yaml")
	require.NoError(t, err)

	// Register extension that calls itself (would cause recursion)
	recursiveManifest := &Manifest{
		Name:        "recursive-ext",
		Version:     "1.0.0",
		Description: "Recursive",
		Type:        "workflow",
		Steps: []WorkflowStep{
			{
				Uses: "extension.call",
				With: map[string]interface{}{
					"extension": "recursive-ext",
				},
			},
		},
	}
	require.NoError(t, registry.Register(recursiveManifest))

	// Try to call recursive extension - should hit depth limit
	_, err = executor.ExecuteExtensionInternal(childCtx, "recursive-ext", map[string]interface{}{}, "caller")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum call depth exceeded")
}

func TestExecuteExtensionInternal_CallBudget(t *testing.T) {
	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": "test",
					},
				},
			},
		}, nil
	}

	executor := NewWorkflowExecutor(llmHandler, "test")
	executor.SetRegistry(registry)

	manifest := &Manifest{
		Name:        "test-ext",
		Version:     "1.0.0",
		Description: "Test",
		Type:        "workflow",
		Steps: []WorkflowStep{
			{
				Uses: "llm.chat",
				With: map[string]interface{}{
					"prompt": "test prompt",
				},
			},
		},
	}
	require.NoError(t, registry.Register(manifest))

	// Test call budget exhaustion
	execCtx := NewExecutionContext(context.Background(), "test", "test.yaml")
	execCtx.CallBudget = 1 // Very low budget

	// First call should work
	_, err := executor.ExecuteExtensionInternal(execCtx, "test-ext", map[string]interface{}{}, "caller")
	require.NoError(t, err)

	// Budget should be exhausted
	childCtx, err := execCtx.WithChild("test.yaml")
	require.NoError(t, err)
	assert.Equal(t, 0, childCtx.CallBudget)

	// Next call should fail
	_, err = executor.ExecuteExtensionInternal(childCtx, "test-ext", map[string]interface{}{}, "caller")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "call budget exhausted")
}
