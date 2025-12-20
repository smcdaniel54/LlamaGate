package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

// CacheEntry holds a cached response with timestamp
type CacheEntry struct {
	Response  []byte
	Timestamp time.Time
}

// Cache is a thread-safe in-memory cache using sync.Map
type Cache struct {
	store sync.Map
}

// New creates a new cache instance
func New() *Cache {
	return &Cache{}
}

// hashRequest creates a SHA256 hash of the model and messages for use as a cache key
func hashRequest(model string, messages interface{}) (string, error) {
	// Create a struct to hash
	keyData := struct {
		Model    string      `json:"model"`
		Messages interface{} `json:"messages"`
	}{
		Model:    model,
		Messages: messages,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(keyData)
	if err != nil {
		return "", err
	}

	// Hash the JSON
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:]), nil
}

// Get retrieves a cached response if it exists
func (c *Cache) Get(model string, messages interface{}) ([]byte, bool) {
	key, err := hashRequest(model, messages)
	if err != nil {
		return nil, false
	}

	value, ok := c.store.Load(key)
	if !ok {
		return nil, false
	}

	entry, ok := value.(*CacheEntry)
	if !ok {
		return nil, false
	}

	return entry.Response, true
}

// Set stores a response in the cache
func (c *Cache) Set(model string, messages interface{}, response []byte) error {
	key, err := hashRequest(model, messages)
	if err != nil {
		return err
	}

	entry := &CacheEntry{
		Response:  response,
		Timestamp: time.Now(),
	}

	c.store.Store(key, entry)
	return nil
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.store.Range(func(key, value interface{}) bool {
		c.store.Delete(key)
		return true
	})
}

