package core

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockExtension is a simple mock implementation of Extension
type mockExtension struct {
	name    string
	version string
	initErr error
	shutErr error
}

func (m *mockExtension) Name() string {
	return m.name
}

func (m *mockExtension) Version() string {
	return m.version
}

func (m *mockExtension) Initialize(_ context.Context, _ map[string]interface{}) error {
	return m.initErr
}

func (m *mockExtension) Shutdown(_ context.Context) error {
	return m.shutErr
}

// mockAgentCaller is a mock implementation of AgentCaller
type mockAgentCaller struct {
	*mockExtension
	callErr      error
	callStreamErr error
}

func (m *mockAgentCaller) Call(ctx context.Context, req *AgentRequest) (*AgentResponse, error) {
	if m.callErr != nil {
		return nil, m.callErr
	}
	return &AgentResponse{
		ID:      "test-id",
		Model:   req.Model,
		Content: "test response",
	}, nil
}

func (m *mockAgentCaller) CallStream(_ context.Context, req *AgentRequest) (<-chan *StreamChunk, error) {
	if m.callStreamErr != nil {
		return nil, m.callStreamErr
	}
	ch := make(chan *StreamChunk, 1)
	ch <- &StreamChunk{Content: "test", Done: true}
	close(ch)
	return ch, nil
}

func TestRegistry_Register(t *testing.T) {
	ctx := context.Background()
	registry := GetRegistry()

	t.Run("register basic extension", func(t *testing.T) {
		ext := &mockExtension{
			name:    "test-extension",
			version: "1.0.0",
		}

		err := registry.Register(ctx, ext, nil)
		require.NoError(t, err)

		// Verify it was registered
		retrieved, err := registry.GetExtension("test-extension")
		require.NoError(t, err)
		assert.Equal(t, "test-extension", retrieved.Name())
		assert.Equal(t, "1.0.0", retrieved.Version())
	})

	t.Run("register duplicate extension", func(t *testing.T) {
		ext1 := &mockExtension{name: "duplicate", version: "1.0.0"}
		ext2 := &mockExtension{name: "duplicate", version: "2.0.0"}

		err := registry.Register(ctx, ext1, nil)
		require.NoError(t, err)

		err = registry.Register(ctx, ext2, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("register extension with initialization error", func(t *testing.T) {
		ext := &mockExtension{
			name:    "init-error",
			version: "1.0.0",
			initErr: assert.AnError,
		}

		err := registry.Register(ctx, ext, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to initialize")
	})

	t.Run("register agent caller", func(t *testing.T) {
		caller := &mockAgentCaller{
			mockExtension: &mockExtension{
				name:    "test-caller",
				version: "1.0.0",
			},
		}

		err := registry.Register(ctx, caller, nil)
		require.NoError(t, err)

		// Verify it's registered as both extension and caller
		retrieved, err := registry.GetExtension("test-caller")
		require.NoError(t, err)
		assert.Equal(t, "test-caller", retrieved.Name())

		retrievedCaller, err := registry.GetAgentCaller("test-caller")
		require.NoError(t, err)
		assert.NotNil(t, retrievedCaller)
	})
}

func TestRegistry_Unregister(t *testing.T) {
	ctx := context.Background()
	registry := GetRegistry()

	t.Run("unregister existing extension", func(t *testing.T) {
		ext := &mockExtension{name: "to-unregister", version: "1.0.0"}
		err := registry.Register(ctx, ext, nil)
		require.NoError(t, err)

		err = registry.Unregister(ctx, "to-unregister")
		require.NoError(t, err)

		// Verify it's gone
		_, err = registry.GetExtension("to-unregister")
		assert.Error(t, err)
	})

	t.Run("unregister non-existent extension", func(t *testing.T) {
		err := registry.Unregister(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("unregister with shutdown error", func(t *testing.T) {
		ext := &mockExtension{
			name:    "shutdown-error",
			version: "1.0.0",
			shutErr: assert.AnError,
		}
		err := registry.Register(ctx, ext, nil)
		require.NoError(t, err)

		err = registry.Unregister(ctx, "shutdown-error")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to shutdown")
	})
}

func TestRegistry_GetExtension(t *testing.T) {
	ctx := context.Background()
	registry := GetRegistry()

	ext := &mockExtension{name: "get-test", version: "1.0.0"}
	err := registry.Register(ctx, ext, nil)
	require.NoError(t, err)

	t.Run("get existing extension", func(t *testing.T) {
		retrieved, err := registry.GetExtension("get-test")
		require.NoError(t, err)
		assert.Equal(t, "get-test", retrieved.Name())
	})

	t.Run("get non-existent extension", func(t *testing.T) {
		_, err := registry.GetExtension("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestRegistry_GetAgentCaller(t *testing.T) {
	ctx := context.Background()
	registry := GetRegistry()

	caller := &mockAgentCaller{
		mockExtension: &mockExtension{name: "get-caller-test", version: "1.0.0"},
	}
	err := registry.Register(ctx, caller, nil)
	require.NoError(t, err)

	t.Run("get existing caller", func(t *testing.T) {
		retrieved, err := registry.GetAgentCaller("get-caller-test")
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	t.Run("get non-existent caller", func(t *testing.T) {
		_, err := registry.GetAgentCaller("non-existent-caller")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestRegistry_ListExtensions(t *testing.T) {
	ctx := context.Background()
	registry := GetRegistry()

	// Register multiple extensions
	ext1 := &mockExtension{name: "list-1", version: "1.0.0"}
	ext2 := &mockExtension{name: "list-2", version: "1.0.0"}
	ext3 := &mockAgentCaller{
		mockExtension: &mockExtension{name: "list-3", version: "1.0.0"},
	}

	require.NoError(t, registry.Register(ctx, ext1, nil))
	require.NoError(t, registry.Register(ctx, ext2, nil))
	require.NoError(t, registry.Register(ctx, ext3, nil))

	t.Run("list all extensions", func(t *testing.T) {
		extensions := registry.ListExtensions()
		assert.GreaterOrEqual(t, len(extensions), 3)

		names := make(map[string]bool)
		for _, ext := range extensions {
			names[ext.Name()] = true
		}

		assert.True(t, names["list-1"])
		assert.True(t, names["list-2"])
		assert.True(t, names["list-3"])
	})
}

func TestRegistry_ListByType(t *testing.T) {
	ctx := context.Background()
	registry := GetRegistry()

	caller := &mockAgentCaller{
		mockExtension: &mockExtension{name: "type-caller", version: "1.0.0"},
	}
	require.NoError(t, registry.Register(ctx, caller, nil))

	t.Run("list by agent_caller type", func(t *testing.T) {
		extensions := registry.ListByType("agent_caller")
		assert.GreaterOrEqual(t, len(extensions), 1)

		found := false
		for _, ext := range extensions {
			if ext.Name() == "type-caller" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("list by unknown type", func(t *testing.T) {
		extensions := registry.ListByType("unknown_type")
		assert.Empty(t, extensions)
	})
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	registry := GetRegistry()

	// Test concurrent registration
	t.Run("concurrent registration", func(t *testing.T) {
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(id int) {
				ext := &mockExtension{
					name:    "concurrent-" + string(rune(id)),
					version: "1.0.0",
				}
				_ = registry.Register(ctx, ext, nil)
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify no panics occurred
		extensions := registry.ListExtensions()
		assert.GreaterOrEqual(t, len(extensions), 10)
	})
}

func TestRegistry_GetRegistry(t *testing.T) {
	// Test that GetRegistry returns the same instance
	reg1 := GetRegistry()
	reg2 := GetRegistry()
	assert.Equal(t, reg1, reg2)
}

func TestAgentCaller_Call(t *testing.T) {
	ctx := context.Background()
	caller := &mockAgentCaller{
		mockExtension: &mockExtension{name: "caller", version: "1.0.0"},
	}

	req := &AgentRequest{
		Model:    "test-model",
		Messages: []Message{{Role: "user", Content: "test"}},
		Stream:   false,
	}

	t.Run("successful call", func(t *testing.T) {
		resp, err := caller.Call(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "test-model", resp.Model)
		assert.Equal(t, "test response", resp.Content)
	})

	t.Run("call with error", func(t *testing.T) {
		caller.callErr = assert.AnError
		resp, err := caller.Call(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		caller.callErr = nil // Reset
	})
}

func TestAgentCaller_CallStream(t *testing.T) {
	ctx := context.Background()
	caller := &mockAgentCaller{
		mockExtension: &mockExtension{name: "caller", version: "1.0.0"},
	}

	req := &AgentRequest{
		Model:    "test-model",
		Messages: []Message{{Role: "user", Content: "test"}},
		Stream:   true,
	}

	t.Run("successful stream", func(t *testing.T) {
		ch, err := caller.CallStream(ctx, req)
		require.NoError(t, err)

		chunk := <-ch
		assert.Equal(t, "test", chunk.Content)
		assert.True(t, chunk.Done)
	})

	t.Run("stream with error", func(t *testing.T) {
		caller.callStreamErr = assert.AnError
		ch, err := caller.CallStream(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, ch)
		caller.callStreamErr = nil // Reset
	})
}

func TestRegistry_ConfigHandling(t *testing.T) {
	ctx := context.Background()
	registry := GetRegistry()

	t.Run("register with config", func(t *testing.T) {
		ext := &mockExtension{name: "config-test", version: "1.0.0"}
		config := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}

		err := registry.Register(ctx, ext, config)
		require.NoError(t, err)
		// Config is passed to Initialize, which our mock handles
	})
}

func TestRegistry_TypeCategorization(t *testing.T) {
	ctx := context.Background()
	registry := GetRegistry()

	caller := &mockAgentCaller{
		mockExtension: &mockExtension{name: "categorize-test", version: "1.0.0"},
	}
	require.NoError(t, registry.Register(ctx, caller, nil))

	// Verify it's in the callers map
	retrieved, err := registry.GetAgentCaller("categorize-test")
	require.NoError(t, err)
	assert.NotNil(t, retrieved)

	// Verify it's also in extensions
	ext, err := registry.GetExtension("categorize-test")
	require.NoError(t, err)
	assert.NotNil(t, ext)
}

func TestRegistry_UnregisterRemovesFromAllCategories(t *testing.T) {
	ctx := context.Background()
	registry := GetRegistry()

	caller := &mockAgentCaller{
		mockExtension: &mockExtension{name: "remove-test", version: "1.0.0"},
	}
	require.NoError(t, registry.Register(ctx, caller, nil))

	// Verify it exists
	_, err := registry.GetAgentCaller("remove-test")
	require.NoError(t, err)

	// Unregister
	require.NoError(t, registry.Unregister(ctx, "remove-test"))

	// Verify it's removed from all categories
	_, err = registry.GetExtension("remove-test")
	assert.Error(t, err)

	_, err = registry.GetAgentCaller("remove-test")
	assert.Error(t, err)
}

func TestRegistry_ContextPropagation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	registry := GetRegistry()
	ext := &mockExtension{name: "ctx-test", version: "1.0.0"}

	err := registry.Register(ctx, ext, nil)
	require.NoError(t, err)

	// Context should be passed through to Initialize
	// (mock doesn't use it, but real implementations would)
}
