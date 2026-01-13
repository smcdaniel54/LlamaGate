package plugins

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPlugin is a simple plugin for testing
type mockPlugin struct {
	name string
}

func (m *mockPlugin) Metadata() PluginMetadata {
	return PluginMetadata{
		Name:           m.name,
		Version:        "1.0.0",
		Description:    "Test plugin",
		RequiredInputs: []string{"input"},
	}
}

func (m *mockPlugin) ValidateInput(input map[string]interface{}) error {
	if _, ok := input["input"]; !ok {
		return assert.AnError
	}
	return nil
}

func (m *mockPlugin) Execute(_ context.Context, input map[string]interface{}) (*PluginResult, error) {
	return &PluginResult{
		Success: true,
		Data:    input,
	}, nil
}

func TestRegistry_RegisterWithContext(t *testing.T) {
	registry := NewRegistry()
	logger := zerolog.Nop()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{}, nil
	}
	config := map[string]interface{}{"key": "value"}
	pluginCtx := NewPluginContext(llmHandler, logger, config)

	plugin := &mockPlugin{name: "test_plugin"}

	err := registry.RegisterWithContext(plugin, pluginCtx)
	require.NoError(t, err)

	// Verify plugin is registered
	retrieved, err := registry.Get("test_plugin")
	assert.NoError(t, err)
	assert.Equal(t, plugin, retrieved)

	// Verify context is stored
	retrievedCtx := registry.GetContext("test_plugin")
	assert.NotNil(t, retrievedCtx)
	assert.Equal(t, pluginCtx, retrievedCtx)
}

func TestRegistry_GetContext(t *testing.T) {
	registry := NewRegistry()
	logger := zerolog.Nop()
	pluginCtx := NewPluginContext(nil, logger, nil)

	plugin := &mockPlugin{name: "test_plugin"}
	err := registry.RegisterWithContext(plugin, pluginCtx)
	require.NoError(t, err)

	// Test getting context
	ctx := registry.GetContext("test_plugin")
	assert.NotNil(t, ctx)
	assert.Equal(t, pluginCtx, ctx)

	// Test getting context for non-existent plugin
	ctx = registry.GetContext("non_existent")
	assert.Nil(t, ctx)
}

func TestRegistry_SetContext(t *testing.T) {
	registry := NewRegistry()
	logger := zerolog.Nop()
	pluginCtx1 := NewPluginContext(nil, logger, nil)
	pluginCtx2 := NewPluginContext(nil, logger, map[string]interface{}{"updated": true})

	plugin := &mockPlugin{name: "test_plugin"}
	err := registry.RegisterWithContext(plugin, pluginCtx1)
	require.NoError(t, err)

	// Update context
	registry.SetContext("test_plugin", pluginCtx2)

	// Verify context was updated
	ctx := registry.GetContext("test_plugin")
	assert.NotNil(t, ctx)
	assert.Equal(t, pluginCtx2, ctx)
}
