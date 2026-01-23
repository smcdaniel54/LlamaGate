// Package registry manages the installed extensions and modules registry.
// It stores metadata about installed items in a JSON file.
package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/llamagate/llamagate/internal/homedir"
)

const (
	// RegistryFileName is the name of the registry JSON file
	RegistryFileName = "installed.json"
)

// ItemType represents the type of installed item
type ItemType string

const (
	ItemTypeExtension     ItemType = "extension"
	ItemTypeAgenticModule ItemType = "agentic-module"
)

// InstalledItem represents a single installed extension or module
type InstalledItem struct {
	ID           string    `json:"id"`
	Type         ItemType  `json:"type"`
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	Enabled      bool      `json:"enabled"`
	InstalledAt  time.Time `json:"installed_at"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
	SourcePath   string    `json:"source_path"` // Path to the installed directory
}

// Registry represents the installed items registry
type Registry struct {
	mu     sync.RWMutex
	items  map[string]*InstalledItem // key: id
	path   string
}

// NewRegistry creates a new registry instance
func NewRegistry() (*Registry, error) {
	registryDir, err := homedir.GetRegistryDir()
	if err != nil {
		return nil, err
	}

	registryPath := filepath.Join(registryDir, RegistryFileName)

	reg := &Registry{
		items: make(map[string]*InstalledItem),
		path:  registryPath,
	}

	// Load existing registry if it exists
	if err := reg.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load registry: %w", err)
	}

	// Note: Self-heal is performed by startup package after discovery
	// to avoid import cycles

	return reg, nil
}

// Load loads the registry from disk
func (r *Registry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			// Registry doesn't exist yet, start with empty
			r.items = make(map[string]*InstalledItem)
			return nil
		}
		return err
	}

	var items []*InstalledItem
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("failed to parse registry: %w", err)
	}

	// Convert slice to map
	r.items = make(map[string]*InstalledItem)
	for _, item := range items {
		r.items[item.ID] = item
	}

	return nil
}

// Save saves the registry to disk
// Assumes caller already holds the appropriate lock (read or write)
func (r *Registry) Save() error {
	// Convert map to slice for JSON
	items := make([]*InstalledItem, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(r.path), 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	// Write atomically using temp file
	tmpPath := r.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write registry: %w", err)
	}

	if err := os.Rename(tmpPath, r.path); err != nil {
		return fmt.Errorf("failed to rename registry: %w", err)
	}

	return nil
}

// Register adds or updates an item in the registry
func (r *Registry) Register(item *InstalledItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// If item exists, preserve InstalledAt and update last_updated_at
	if existing, exists := r.items[item.ID]; exists {
		// Preserve InstalledAt from existing item
		item.InstalledAt = existing.InstalledAt
		item.LastUpdatedAt = time.Now()
	} else {
		// New item - set both timestamps
		if item.InstalledAt.IsZero() {
			item.InstalledAt = time.Now()
		}
		if item.LastUpdatedAt.IsZero() {
			item.LastUpdatedAt = item.InstalledAt
		}
	}

	r.items[item.ID] = item
	return r.Save()
}

// Unregister removes an item from the registry
func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.items, id)
	return r.Save()
}

// Get retrieves an item by ID
func (r *Registry) Get(id string) (*InstalledItem, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, exists := r.items[id]
	return item, exists
}

// List returns all items, optionally filtered by type
func (r *Registry) List(itemType ItemType) []*InstalledItem {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var items []*InstalledItem
	for _, item := range r.items {
		if itemType == "" || item.Type == itemType {
			items = append(items, item)
		}
	}

	return items
}

// SetEnabled sets the enabled status of an item
func (r *Registry) SetEnabled(id string, enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.items[id]
	if !exists {
		return fmt.Errorf("item not found: %s", id)
	}

	item.Enabled = enabled
	item.LastUpdatedAt = time.Now()
	return r.Save() // Save() assumes lock is already held
}

// Exists checks if an item exists in the registry
func (r *Registry) Exists(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.items[id]
	return exists
}
