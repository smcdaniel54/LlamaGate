package mcpclient

import (
	"sync"
	"time"
)

// CacheEntry represents a cached entry with expiration
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Cache provides TTL-based caching for MCP metadata
type Cache struct {
	tools     map[string]*CacheEntry
	resources map[string]*CacheEntry
	prompts   map[string]*CacheEntry
	mu        sync.RWMutex
	ttl       time.Duration
}

// NewCache creates a new cache with the specified TTL
func NewCache(ttl time.Duration) *Cache {
	if ttl <= 0 {
		ttl = 5 * time.Minute // Default TTL
	}

	cache := &Cache{
		tools:     make(map[string]*CacheEntry),
		resources: make(map[string]*CacheEntry),
		prompts:   make(map[string]*CacheEntry),
		ttl:       ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// GetTool gets a cached tool list for a server
func (c *Cache) GetTool(serverName string) ([]Tool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.tools[serverName]
	if !exists || entry.IsExpired() {
		return nil, false
	}

	tools, ok := entry.Data.([]Tool)
	if !ok {
		return nil, false
	}

	// Return a copy to avoid race conditions
	result := make([]Tool, len(tools))
	copy(result, tools)
	return result, true
}

// SetTool caches a tool list for a server
func (c *Cache) SetTool(serverName string, tools []Tool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a copy to avoid race conditions
	toolsCopy := make([]Tool, len(tools))
	copy(toolsCopy, tools)

	c.tools[serverName] = &CacheEntry{
		Data:      toolsCopy,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// GetResources gets cached resources for a server
func (c *Cache) GetResources(serverName string) ([]Resource, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.resources[serverName]
	if !exists || entry.IsExpired() {
		return nil, false
	}

	resources, ok := entry.Data.([]Resource)
	if !ok {
		return nil, false
	}

	// Return a copy to avoid race conditions
	result := make([]Resource, len(resources))
	copy(result, resources)
	return result, true
}

// SetResources caches resources for a server
func (c *Cache) SetResources(serverName string, resources []Resource) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a copy to avoid race conditions
	resourcesCopy := make([]Resource, len(resources))
	copy(resourcesCopy, resources)

	c.resources[serverName] = &CacheEntry{
		Data:      resourcesCopy,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// GetPrompts gets cached prompts for a server
func (c *Cache) GetPrompts(serverName string) ([]Prompt, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.prompts[serverName]
	if !exists || entry.IsExpired() {
		return nil, false
	}

	prompts, ok := entry.Data.([]Prompt)
	if !ok {
		return nil, false
	}

	// Return a copy to avoid race conditions
	result := make([]Prompt, len(prompts))
	copy(result, prompts)
	return result, true
}

// SetPrompts caches prompts for a server
func (c *Cache) SetPrompts(serverName string, prompts []Prompt) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a copy to avoid race conditions
	promptsCopy := make([]Prompt, len(prompts))
	copy(promptsCopy, prompts)

	c.prompts[serverName] = &CacheEntry{
		Data:      promptsCopy,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// InvalidateTool invalidates cached tools for a server
func (c *Cache) InvalidateTool(serverName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.tools, serverName)
}

// InvalidateResources invalidates cached resources for a server
func (c *Cache) InvalidateResources(serverName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.resources, serverName)
}

// InvalidatePrompts invalidates cached prompts for a server
func (c *Cache) InvalidatePrompts(serverName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.prompts, serverName)
}

// InvalidateAll invalidates all cached data for a server
func (c *Cache) InvalidateAll(serverName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.tools, serverName)
	delete(c.resources, serverName)
	delete(c.prompts, serverName)
}

// Clear clears all cached data
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tools = make(map[string]*CacheEntry)
	c.resources = make(map[string]*CacheEntry)
	c.prompts = make(map[string]*CacheEntry)
}

// cleanup periodically removes expired entries
func (c *Cache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()

		// Clean up expired tools
		for key, entry := range c.tools {
			if entry.IsExpired() {
				delete(c.tools, key)
			}
		}

		// Clean up expired resources
		for key, entry := range c.resources {
			if entry.IsExpired() {
				delete(c.resources, key)
			}
		}

		// Clean up expired prompts
		for key, entry := range c.prompts {
			if entry.IsExpired() {
				delete(c.prompts, key)
			}
		}

		c.mu.Unlock()
		_ = now // Avoid unused variable warning
	}
}

