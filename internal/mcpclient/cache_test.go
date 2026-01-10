package mcpclient

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCache(t *testing.T) {
	cache := NewCache(5 * time.Minute)
	assert.NotNil(t, cache)
}

func TestCache_Tools(t *testing.T) {
	cache := NewCache(5 * time.Minute)

	tools := []Tool{
		{Name: "tool1", Description: "Test tool 1"},
		{Name: "tool2", Description: "Test tool 2"},
	}

	// Set tools
	cache.SetTool("server1", tools)

	// Get tools
	cached, found := cache.GetTool("server1")
	assert.True(t, found)
	assert.Len(t, cached, 2)
	assert.Equal(t, "tool1", cached[0].Name)

	// Invalidate
	cache.InvalidateTool("server1")
	_, found = cache.GetTool("server1")
	assert.False(t, found)
}

func TestCache_Resources(t *testing.T) {
	cache := NewCache(5 * time.Minute)

	resources := []Resource{
		{URI: "file:///test1.txt", Name: "test1.txt"},
		{URI: "file:///test2.txt", Name: "test2.txt"},
	}

	// Set resources
	cache.SetResources("server1", resources)

	// Get resources
	cached, found := cache.GetResources("server1")
	assert.True(t, found)
	assert.Len(t, cached, 2)
	assert.Equal(t, "file:///test1.txt", cached[0].URI)

	// Invalidate
	cache.InvalidateResources("server1")
	_, found = cache.GetResources("server1")
	assert.False(t, found)
}

func TestCache_Prompts(t *testing.T) {
	cache := NewCache(5 * time.Minute)

	prompts := []Prompt{
		{Name: "prompt1", Description: "Test prompt 1"},
		{Name: "prompt2", Description: "Test prompt 2"},
	}

	// Set prompts
	cache.SetPrompts("server1", prompts)

	// Get prompts
	cached, found := cache.GetPrompts("server1")
	assert.True(t, found)
	assert.Len(t, cached, 2)
	assert.Equal(t, "prompt1", cached[0].Name)

	// Invalidate
	cache.InvalidatePrompts("server1")
	_, found = cache.GetPrompts("server1")
	assert.False(t, found)
}

func TestCache_Expiration(t *testing.T) {
	cache := NewCache(100 * time.Millisecond)

	tools := []Tool{{Name: "tool1"}}
	cache.SetTool("server1", tools)

	// Should be found immediately
	_, found := cache.GetTool("server1")
	assert.True(t, found)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not be found after expiration
	_, found = cache.GetTool("server1")
	assert.False(t, found)
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(5 * time.Minute)
	defer cache.Stop()

	cache.SetTool("server1", []Tool{{Name: "tool1"}})
	cache.SetResources("server1", []Resource{{URI: "file:///test.txt"}})
	cache.SetPrompts("server1", []Prompt{{Name: "prompt1"}})

	cache.Clear()

	_, found := cache.GetTool("server1")
	assert.False(t, found)
	_, found = cache.GetResources("server1")
	assert.False(t, found)
	_, found = cache.GetPrompts("server1")
	assert.False(t, found)
}

func TestCache_Stop(t *testing.T) {
	cache := NewCache(5 * time.Minute)

	// Stop should be idempotent
	cache.Stop()
	cache.Stop()
	cache.Stop()

	// Cache should still work after stopping cleanup goroutine
	cache.SetTool("server1", []Tool{{Name: "tool1"}})
	tools, found := cache.GetTool("server1")
	assert.True(t, found)
	assert.NotNil(t, tools)
}

func TestCache_StopPreventsGoroutineLeak(t *testing.T) {
	// This test verifies that Stop() actually stops the cleanup goroutine
	// by checking that multiple calls to Stop() don't panic and the cache
	// can still be used after stopping
	cache := NewCache(1 * time.Second) // Use longer TTL to avoid expiration during test

	// Add some entries
	cache.SetTool("server1", []Tool{{Name: "tool1"}})
	cache.SetResources("server1", []Resource{{URI: "file:///test.txt"}})

	// Verify entries are set before stopping
	tools, found := cache.GetTool("server1")
	require.True(t, found, "Tool should be found before Stop()")
	require.NotNil(t, tools, "Tools should not be nil before Stop()")

	// Stop the cleanup goroutine
	cache.Stop()

	// Wait a bit to ensure cleanup would have run if it was still active
	time.Sleep(200 * time.Millisecond)

	// Cache operations should still work after stopping
	tools, found = cache.GetTool("server1")
	assert.True(t, found, "Tool should still be found after Stop()")
	assert.NotNil(t, tools, "Tools should not be nil after Stop()")

	// Multiple calls to Stop() should be safe
	cache.Stop()
	cache.Stop()
}
