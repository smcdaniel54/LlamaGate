package plugins

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestPluginContext_CallLLM(t *testing.T) {
	tests := []struct {
		name       string
		llmHandler LLMHandlerFunc
		expectErr  bool
	}{
		{
			name: "successful LLM call",
			llmHandler: func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
				return map[string]interface{}{
					"choices": []interface{}{
						map[string]interface{}{
							"message": map[string]interface{}{
								"content": "Hello, world!",
							},
						},
					},
				}, nil
			},
			expectErr: false,
		},
		{
			name:       "no LLM handler",
			llmHandler: nil,
			expectErr:  true,
		},
		{
			name: "LLM handler error",
			llmHandler: func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
				return nil, assert.AnError
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zerolog.Nop()
			ctx := NewPluginContext(tt.llmHandler, logger, nil)

			messages := []map[string]interface{}{
				{"role": "user", "content": "test"},
			}

			result, err := ctx.CallLLM(context.Background(), "llama3.2", messages, nil)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "Hello, world!", result)
			}
		})
	}
}

func TestPluginContext_GetConfig(t *testing.T) {
	config := map[string]interface{}{
		"string_val": "test",
		"int_val":    42,
		"bool_val":   true,
	}

	logger := zerolog.Nop()
	ctx := NewPluginContext(nil, logger, config)

	t.Run("GetConfig", func(t *testing.T) {
		assert.Equal(t, "test", ctx.GetConfig("string_val", "default"))
		assert.Equal(t, "default", ctx.GetConfig("missing", "default"))
	})

	t.Run("GetConfigString", func(t *testing.T) {
		assert.Equal(t, "test", ctx.GetConfigString("string_val", "default"))
		assert.Equal(t, "default", ctx.GetConfigString("missing", "default"))
	})

	t.Run("GetConfigInt", func(t *testing.T) {
		assert.Equal(t, 42, ctx.GetConfigInt("int_val", 0))
		assert.Equal(t, 0, ctx.GetConfigInt("missing", 0))
	})

	t.Run("GetConfigBool", func(t *testing.T) {
		assert.True(t, ctx.GetConfigBool("bool_val", false))
		assert.False(t, ctx.GetConfigBool("missing", false))
	})
}
