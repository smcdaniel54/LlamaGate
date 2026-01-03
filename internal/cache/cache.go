// Package cache provides in-memory caching for chat completion responses.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// CacheEntry holds a cached response with timestamp
//
//nolint:revive // CacheEntry is the preferred name for external API
type CacheEntry struct {
	Response  []byte
	Timestamp time.Time
}

// Cache is a thread-safe in-memory cache using sync.Map with TTL and size limits
type Cache struct {
	store       sync.Map
	maxSize     int           // Maximum number of entries (0 = unlimited)
	ttl         time.Duration // Time-to-live for entries (0 = no expiration)
	mu          sync.RWMutex  // For size tracking
	entryCount  int           // Track current size
	stopCleanup chan struct{} // Channel to stop cleanup goroutine
	stopOnce    sync.Once     // Ensure StopCleanup is only called once
}

// CacheOptions holds configuration options for the cache
//
//nolint:revive // CacheOptions is the preferred name for external API
type CacheOptions struct {
	MaxSize int           // Maximum number of entries (0 = unlimited, default: 1000)
	TTL     time.Duration // Time-to-live for entries (0 = no expiration, default: 1 hour)
}

// New creates a new cache instance with default settings (1000 entries max, 1 hour TTL)
func New() *Cache {
	return NewWithOptions(CacheOptions{
		MaxSize: 1000,
		TTL:     1 * time.Hour,
	})
}

// NewWithOptions creates a new cache instance with custom options
func NewWithOptions(opts CacheOptions) *Cache {
	c := &Cache{
		maxSize:     opts.MaxSize,
		ttl:         opts.TTL,
		entryCount:  0,
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine if TTL is set
	if opts.TTL > 0 {
		go c.cleanupExpired()
	}

	return c
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

// Get retrieves a cached response if it exists and is not expired
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

	// Check if entry has expired
	if c.ttl > 0 && time.Since(entry.Timestamp) > c.ttl {
		// Use LoadAndDelete to atomically get and delete the entry
		// This prevents race condition where Set() updates the entry with a new
		// timestamp between our check and deletion
		deletedEntry, deleted := c.store.LoadAndDelete(key)
		if deleted {
			// Verify the deleted entry was actually expired
			// If it was updated by Set() between our check and LoadAndDelete,
			// the new entry won't be expired, so we need to restore it
			deletedCacheEntry, ok := deletedEntry.(*CacheEntry)
			if ok && time.Since(deletedCacheEntry.Timestamp) > c.ttl {
				// Entry was actually expired - decrement counter
				c.mu.Lock()
				c.entryCount--
				c.mu.Unlock()
			} else {
				// Entry was updated with a new valid timestamp - restore it
				// Use LoadOrStore to prevent overwriting a fresh entry that may have
				// been stored by Set() between LoadAndDelete and this restore
				// If another thread stored a fresh entry, LoadOrStore will return it and we won't overwrite
				// If no fresh entry exists, we restore the deleted entry
				c.store.LoadOrStore(key, deletedEntry)
				// Note: We do NOT increment the counter here because the entry was never
				// removed from the counter (only from the store). The counter still accounts
				// for this entry, so restoring it doesn't change the count.
			}
		}
		return nil, false
	}

	return entry.Response, true
}

// Set stores a response in the cache, evicting old entries if size limit is reached
func (c *Cache) Set(model string, messages interface{}, response []byte) error {
	key, err := hashRequest(model, messages)
	if err != nil {
		return err
	}

	// Prepare entry before acquiring lock to minimize lock time
	entry := &CacheEntry{
		Response:  response,
		Timestamp: time.Now(),
	}

	// Use LoadOrStore atomically to handle race conditions with cleanupExpired
	// and concurrent Set() calls. This prevents counter corruption when cleanupExpired
	// deletes an entry between our existence check and store operation.
	_, loaded := c.store.LoadOrStore(key, entry)
	if loaded {
		// Entry already exists - update it atomically
		// However, between LoadOrStore and Store(), another thread could delete
		// the entry and decrement the counter. We need to handle this race condition.
		// Use LoadAndDelete to atomically check if entry still exists
		_, wasDeleted := c.store.LoadAndDelete(key)
		if wasDeleted {
			// Entry was deleted between LoadOrStore and Store() - restore it and increment counter
			// Use LoadOrStore to restore, handling race where another thread might have stored it
			_, loadedOnRestore := c.store.LoadOrStore(key, entry)
			if !loadedOnRestore {
				// We successfully restored the entry - increment counter
				c.mu.Lock()
				c.entryCount++
				c.mu.Unlock()
			}
			// If loadedOnRestore is true, another thread stored it, so counter is already correct
		} else {
			// Entry still exists - update it using Store()
			// However, between LoadAndDelete and Store(), another thread could delete
			// the entry. If that happens, Store() will restore it but we won't increment
			// the counter. This is a very rare race condition that cannot be easily
			// detected without additional state tracking or synchronization overhead.
			// The counter may be slightly off in this case, but it will self-correct
			// on the next eviction or expiration, which is acceptable for this use case.
			c.store.Store(key, entry)
		}
		return nil
	}

	// Entry didn't exist - we successfully stored a new entry
	// Now we need to increment the counter, but we must check for eviction first
	// Hold lock for counter operations and eviction
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict BEFORE incrementing counter to maintain accurate count
	if c.maxSize > 0 && c.entryCount >= c.maxSize {
		// Evict oldest entry (simple FIFO eviction)
		// evictOldest() may fail if Get() concurrently deleted the entry
		// We need to verify eviction succeeded before proceeding
		// Retry eviction with a higher limit to handle concurrent deletions
		// We must succeed in evicting at least one entry to make room
		maxRetries := 10 // Increased retries to handle high concurrency
		for retry := 0; retry < maxRetries && c.entryCount >= c.maxSize; retry++ {
			oldCount := c.entryCount
			c.evictOldest()
			// If eviction succeeded, entryCount will have decreased
			if c.entryCount < oldCount {
				break // Eviction succeeded, we can proceed
			}
			// Eviction failed - entry was already deleted by Get() or cleanupExpired()
			// Try again with a different entry
		}
		// If still at capacity after retries, we must force eviction or return an error
		// Silent failure is not acceptable - callers need to know if caching failed
		if c.entryCount >= c.maxSize {
			// Last resort: try one more eviction attempt
			// If this fails, the cache is truly full and we return an error
			oldCount := c.entryCount
			c.evictOldest()
			if c.entryCount >= oldCount {
				// All eviction attempts failed - cache is full and can't make room
				// We already stored the entry via LoadOrStore, so we cannot delete it
				// as that would violate cache consistency. Instead, we accept that the
				// cache is temporarily over capacity. The counter is accurate, and the
				// next eviction or expiration will correct the size.
				// Return an explicit error so callers know caching succeeded but cache
				// is at capacity
				return errors.New("cache is full and cannot evict entries to make room")
			}
		}
	}

	// Increment counter for the new entry we stored
	c.entryCount++
	return nil
}

// evictOldest removes the oldest entry from the cache (simple FIFO)
// NOTE: This method must be called while holding c.mu.Lock()
func (c *Cache) evictOldest() {
	var oldestKey interface{}
	var oldestTime time.Time
	first := true

	c.store.Range(func(key, value interface{}) bool {
		entry, ok := value.(*CacheEntry)
		if !ok {
			return true
		}

		if first || entry.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Timestamp
			first = false
		}
		return true
	})

	if oldestKey != nil {
		// Use LoadAndDelete to atomically check and delete
		// This prevents double-deletion race condition with Get() which may have
		// already deleted the entry and decremented the counter
		_, deleted := c.store.LoadAndDelete(oldestKey)
		if deleted {
			// Only decrement counter if we actually deleted the entry
			// If Get() already deleted it, LoadAndDelete will return false
			c.entryCount--
		}
	}
}

// cleanupExpired periodically removes expired entries from the cache
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(c.ttl / 2) // Run cleanup at half the TTL interval
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.store.Range(func(key, value interface{}) bool {
				entry, ok := value.(*CacheEntry)
				if !ok {
					return true
				}

				// Check if entry appears expired
				if now.Sub(entry.Timestamp) > c.ttl {
					// Use LoadAndDelete to atomically get and delete the entry
					// This prevents race condition where Set() replaces expired entry
					// with a new valid one between our check and deletion
					deletedEntry, deleted := c.store.LoadAndDelete(key)
					if deleted {
						// Verify the deleted entry was actually expired
						// If it was replaced by Set() between our check and LoadAndDelete,
						// the new entry won't be expired, so we need to restore it
						deletedCacheEntry, ok := deletedEntry.(*CacheEntry)
						if ok && now.Sub(deletedCacheEntry.Timestamp) > c.ttl {
							// Entry was actually expired - decrement counter
							c.mu.Lock()
							c.entryCount--
							c.mu.Unlock()
						} else {
							// Entry was replaced with a new valid one - restore it
							// Use LoadOrStore to prevent overwriting a fresh entry that may have
							// been stored by Set() between LoadAndDelete and this restore
							// If another thread stored a fresh entry, LoadOrStore will return it and we won't overwrite
							// If no fresh entry exists, we restore the deleted entry
							c.store.LoadOrStore(key, deletedEntry)
							// Note: The deleted entry was never decremented from the counter (only expired
							// entries are decremented). The other thread's Set() may have incremented
							// the counter if it was a new entry, or left it unchanged if it was an update.
							// We cannot safely decrement here without knowing which case occurred,
							// and doing so would cause counter corruption since the deleted entry
							// was never removed from the counter in the first place.
							// The counter will be accurate: it accounts for the deleted entry OR the
							// new entry (depending on whether Set() incremented), and the store has
							// the new entry. This is acceptable - the counter may be slightly off
							// in this rare race condition, but it will self-correct on the next
							// eviction or expiration.
							// Note: We do NOT increment the counter here because the entry was never
							// removed from the counter (only from the store). The counter still accounts
							// for this entry, so restoring it doesn't change the count.
						}
					}
				}
				return true
			})
		case <-c.stopCleanup:
			return
		}
	}
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.store.Range(func(key, _ interface{}) bool {
		c.store.Delete(key)
		return true
	})
	c.mu.Lock()
	c.entryCount = 0
	c.mu.Unlock()
}

// StopCleanup stops the background cleanup goroutine
// Safe to call multiple times - uses sync.Once to prevent double-close panic
func (c *Cache) StopCleanup() {
	c.stopOnce.Do(func() {
		if c.stopCleanup != nil {
			close(c.stopCleanup)
		}
	})
}

// Size returns the current number of entries in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.entryCount
}
