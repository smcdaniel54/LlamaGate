// Package discovery provides discovery functionality for installed extensions and modules.
// It can scan disk directories and fallback if registry is missing.
package discovery

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/llamagate/llamagate/internal/extensions"
	"github.com/llamagate/llamagate/internal/homedir"
	"github.com/llamagate/llamagate/internal/packaging"
	"github.com/llamagate/llamagate/internal/registry"
)

// DiscoveredItem represents a discovered extension or module
type DiscoveredItem struct {
	ID          string
	Type        registry.ItemType
	Name        string
	Version     string
	Path        string
	Enabled     bool
	FromRegistry bool
}

// DiscoverInstalledItems discovers all installed extensions and modules.
// First tries registry, then falls back to disk scan if registry is missing or incomplete.
func DiscoverInstalledItems() ([]*DiscoveredItem, error) {
	var items []*DiscoveredItem

	// Try registry first
	reg, err := registry.NewRegistry()
	if err == nil {
		// Load from registry
		regItems := reg.List("")
		for _, item := range regItems {
			// Verify item still exists on disk
			if _, err := os.Stat(item.SourcePath); err == nil {
				items = append(items, &DiscoveredItem{
					ID:           item.ID,
					Type:         item.Type,
					Name:         item.Name,
					Version:      item.Version,
					Path:         item.SourcePath,
					Enabled:      item.Enabled,
					FromRegistry: true,
				})
			}
		}
	}

	// Fallback: scan disk for items not in registry
	extItems, err := scanExtensions()
	if err == nil {
		for _, item := range extItems {
			// Only add if not already in registry
			found := false
			for _, existing := range items {
				if existing.ID == item.ID && existing.Type == item.Type {
					found = true
					break
				}
			}
			if !found {
				items = append(items, item)
			}
		}
	}

	moduleItems, err := scanModules()
	if err == nil {
		for _, item := range moduleItems {
			// Only add if not already in registry
			found := false
			for _, existing := range items {
				if existing.ID == item.ID && existing.Type == item.Type {
					found = true
					break
				}
			}
			if !found {
				items = append(items, item)
			}
		}
	}

	return items, nil
}

// DiscoverEnabledItems discovers only enabled items
func DiscoverEnabledItems() ([]*DiscoveredItem, error) {
	allItems, err := DiscoverInstalledItems()
	if err != nil {
		return nil, err
	}

	var enabledItems []*DiscoveredItem
	for _, item := range allItems {
		if item.Enabled {
			enabledItems = append(enabledItems, item)
		}
	}

	return enabledItems, nil
}

// scanExtensions scans the extensions directory for installed extensions
func scanExtensions() ([]*DiscoveredItem, error) {
	extDir, err := homedir.GetExtensionsDir()
	if err != nil {
		return nil, err
	}

	var items []*DiscoveredItem

	// Scan each subdirectory
	entries, err := os.ReadDir(extDir)
	if err != nil {
		if os.IsNotExist(err) {
			return items, nil // Directory doesn't exist, return empty
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		extPath := filepath.Join(extDir, entry.Name())

		// Try to load as extension
		pkgManifest, err := packaging.LoadExtensionPackageManifest(extPath)
		if err != nil {
			continue // Skip if not a valid extension
		}

		// Determine ID (use name if ID not set)
		id := pkgManifest.ID
		if id == "" {
			id = pkgManifest.Name
		}

		// Check registry for enabled status
		enabled := true
		reg, err := registry.NewRegistry()
		if err == nil {
			if regItem, exists := reg.Get(id); exists {
				enabled = regItem.Enabled
			} else if pkgManifest.EnabledDefault != nil {
				enabled = *pkgManifest.EnabledDefault
			}
		} else if pkgManifest.EnabledDefault != nil {
			enabled = *pkgManifest.EnabledDefault
		}

		items = append(items, &DiscoveredItem{
			ID:           id,
			Type:         registry.ItemTypeExtension,
			Name:         pkgManifest.Name,
			Version:      pkgManifest.Version,
			Path:         extPath,
			Enabled:      enabled,
			FromRegistry: false,
		})
	}

	return items, nil
}

// scanModules scans the modules directory for installed modules
func scanModules() ([]*DiscoveredItem, error) {
	modulesDir, err := homedir.GetAgenticModulesDir()
	if err != nil {
		return nil, err
	}

	var items []*DiscoveredItem

	// Scan each subdirectory
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return items, nil // Directory doesn't exist, return empty
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		modulePath := filepath.Join(modulesDir, entry.Name())

		// Try to load as module
		pkgManifest, err := packaging.LoadModulePackageManifest(modulePath)
		if err != nil {
			continue // Skip if not a valid module
		}

		// Determine ID (use name if ID not set)
		id := pkgManifest.ID
		if id == "" {
			id = pkgManifest.Name
		}

		// Check registry for enabled status
		enabled := true
		reg, err := registry.NewRegistry()
		if err == nil {
			if regItem, exists := reg.Get(id); exists {
				enabled = regItem.Enabled
			} else if pkgManifest.EnabledDefault != nil {
				enabled = *pkgManifest.EnabledDefault
			}
		} else if pkgManifest.EnabledDefault != nil {
			enabled = *pkgManifest.EnabledDefault
		}

		items = append(items, &DiscoveredItem{
			ID:           id,
			Type:         registry.ItemTypeAgenticModule,
			Name:         pkgManifest.Name,
			Version:      pkgManifest.Version,
			Path:         modulePath,
			Enabled:      enabled,
			FromRegistry: false,
		})
	}

	return items, nil
}

// DiscoverLegacyExtensions scans the legacy extensions directory (repo-based)
// and returns items that can be migrated
func DiscoverLegacyExtensions(legacyDir string) ([]*DiscoveredItem, error) {
	if legacyDir == "" {
		return nil, nil
	}

	var items []*DiscoveredItem

	manifests, err := extensions.DiscoverExtensions(legacyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to discover legacy extensions: %w", err)
	}

	for _, manifest := range manifests {
		// Try to find the directory containing this manifest
		manifestPath := filepath.Join(legacyDir, manifest.Name, "manifest.yaml")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			// Try nested structure
			continue
		}

		extPath := filepath.Dir(manifestPath)

		items = append(items, &DiscoveredItem{
			ID:           manifest.Name,
			Type:         registry.ItemTypeExtension,
			Name:         manifest.Name,
			Version:      manifest.Version,
			Path:         extPath,
			Enabled:      manifest.Enabled == nil || *manifest.Enabled,
			FromRegistry: false,
		})
	}

	return items, nil
}
