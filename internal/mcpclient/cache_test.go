package mcpclient

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

