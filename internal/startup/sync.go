// Package startup provides registry synchronization functionality.
package startup

import (
	"github.com/llamagate/llamagate/internal/discovery"
	"github.com/llamagate/llamagate/internal/registry"
	"github.com/rs/zerolog/log"
)

// SyncRegistry synchronizes the registry with filesystem state
func SyncRegistry(reg *registry.Registry) error {
	// Discover installed items
	discoveredItems, err := discovery.DiscoverInstalledItems()
	if err != nil {
		return err
	}

	// Convert to registry format
	regItems := make([]registry.DiscoveredItemInfo, 0, len(discoveredItems))
	for _, item := range discoveredItems {
		regItems = append(regItems, registry.DiscoveredItemInfo{
			ID:         item.ID,
			Type:       item.Type,
			Name:       item.Name,
			Version:    item.Version,
			Enabled:    item.Enabled,
			SourcePath: item.Path,
		})
	}

	// Perform sync
	result, err := reg.Sync(regItems)
	if err != nil {
		return err
	}

	if len(result.Added) > 0 || len(result.Removed) > 0 || len(result.Updated) > 0 {
		log.Info().
			Int("added", len(result.Added)).
			Int("removed", len(result.Removed)).
			Int("updated", len(result.Updated)).
			Msg("Registry synchronized with filesystem")
	}

	return nil
}
