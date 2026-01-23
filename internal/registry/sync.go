// Package registry provides registry synchronization and self-healing functionality.
package registry

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Added   []string
	Removed []string
	Updated []string
	Errors  []string
}

// DiscoveredItemInfo represents discovered item information for sync
type DiscoveredItemInfo struct {
	ID          string
	Type        ItemType
	Name        string
	Version     string
	Enabled     bool
	SourcePath  string
}

// Sync synchronizes the registry with the actual filesystem state
// This self-heals the registry if it's out of sync with disk
// discoveredItems should be provided by the caller to avoid import cycles
func (r *Registry) Sync(discoveredItems []DiscoveredItemInfo) (*SyncResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := &SyncResult{
		Added:   []string{},
		Removed: []string{},
		Updated: []string{},
		Errors:  []string{},
	}

	// Build map of discovered items by ID
	discoveredMap := make(map[string]*DiscoveredItemInfo)
	for i := range discoveredItems {
		discoveredMap[discoveredItems[i].ID] = &discoveredItems[i]
	}

	// Check registry items against disk
	for id, regItem := range r.items {
		discovered, exists := discoveredMap[id]
		if !exists {
			// Registry says installed, but not on disk - remove from registry
			if _, err := os.Stat(regItem.SourcePath); os.IsNotExist(err) {
				log.Warn().
					Str("id", id).
					Str("type", string(regItem.Type)).
					Msg("Removing from registry: item not found on disk")
				delete(r.items, id)
				result.Removed = append(result.Removed, id)
			}
		} else {
			// Item exists on disk - verify path matches and update if needed
			if discovered.SourcePath != regItem.SourcePath {
				regItem.SourcePath = discovered.SourcePath
				result.Updated = append(result.Updated, id)
			}
			// Remove from discovered map (already processed)
			delete(discoveredMap, id)
		}
	}

	// Add items that are on disk but not in registry
	for id, discovered := range discoveredMap {
		// Check if item was in registry (for preserving InstalledAt)
		var installedAt time.Time
		if existingItem, exists := r.items[id]; exists {
			installedAt = existingItem.InstalledAt
		} else {
			installedAt = time.Now()
		}

		item := &InstalledItem{
			ID:            discovered.ID,
			Type:          discovered.Type,
			Name:          discovered.Name,
			Version:       discovered.Version,
			Enabled:       discovered.Enabled,
			SourcePath:    discovered.SourcePath,
			InstalledAt:   installedAt,
			LastUpdatedAt: time.Now(),
		}
		r.items[id] = item
		result.Added = append(result.Added, id)
		log.Info().
			Str("id", id).
			Str("type", string(discovered.Type)).
			Msg("Added to registry: item found on disk but not in registry")
	}

	// Save updated registry
	if err := r.Save(); err != nil {
		return nil, fmt.Errorf("failed to save synced registry: %w", err)
	}

	return result, nil
}

// SelfHeal performs a sync operation and logs the results
// discoveredItems should be provided by the caller to avoid import cycles
func (r *Registry) SelfHeal(discoveredItems []DiscoveredItemInfo) error {
	result, err := r.Sync(discoveredItems)
	if err != nil {
		return err
	}

	if len(result.Added) > 0 || len(result.Removed) > 0 || len(result.Updated) > 0 {
		log.Info().
			Int("added", len(result.Added)).
			Int("removed", len(result.Removed)).
			Int("updated", len(result.Updated)).
			Msg("Registry self-healed: synchronized with filesystem")
	}

	return nil
}
