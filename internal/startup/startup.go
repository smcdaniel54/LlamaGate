// Package startup provides startup integration for loading installed extensions and modules.
package startup

import (
	"fmt"

	"github.com/llamagate/llamagate/internal/discovery"
	"github.com/llamagate/llamagate/internal/extensions"
	"github.com/llamagate/llamagate/internal/migration"
	"github.com/llamagate/llamagate/internal/packaging"
	"github.com/llamagate/llamagate/internal/registry"
	"github.com/rs/zerolog/log"
)

// LoadInstalledExtensions loads all enabled installed extensions into the extension registry
func LoadInstalledExtensions(extRegistry *extensions.Registry, legacyBaseDir string) (int, []string) {
	var loadedCount int
	var failures []string

	// Perform automatic migration on first run (if not already done)
	alreadyMigrated, _ := migration.HasMigrated()
	if !alreadyMigrated {
		log.Info().Msg("First run detected: migrating legacy extensions")
		result, err := migration.MigrateLegacyExtensions(legacyBaseDir)
		if err != nil {
			log.Warn().Err(err).Msg("Migration failed, continuing with legacy discovery")
		} else if result.MigratedExtensions > 0 {
			log.Info().
				Int("migrated", result.MigratedExtensions).
				Msg("Migrated legacy extensions to new layout")
		}
	}

	// Discover enabled installed items
	installedItems, err := discovery.DiscoverEnabledItems()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to discover installed items, falling back to legacy directory")
		// Fall through to legacy discovery
	}

	// Load installed extensions
	for _, item := range installedItems {
		if item.Type != registry.ItemTypeExtension {
			continue // Skip modules for now
		}

		// Load extension manifest
		pkgManifest, err := packaging.LoadExtensionPackageManifest(item.Path)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", item.ID, err))
			log.Warn().
				Str("extension", item.ID).
				Err(err).
				Msg("Failed to load installed extension manifest")
			continue
		}

		// Register with extension registry
		if err := extRegistry.RegisterOrUpdate(pkgManifest.ExtensionManifest); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", item.ID, err))
			log.Warn().
				Str("extension", item.ID).
				Err(err).
				Msg("Failed to register installed extension")
			continue
		}

		loadedCount++
		log.Info().
			Str("extension", item.ID).
			Str("name", item.Name).
			Str("version", item.Version).
			Msg("Loaded installed extension")
	}

	// Discover legacy extensions (only if not already installed)
	legacyManifests, err := extensions.DiscoverExtensions(legacyBaseDir)
	if err != nil {
		log.Debug().Err(err).Msg("Legacy extension discovery failed or directory doesn't exist")
	} else {
		for _, manifest := range legacyManifests {
			// Check if already loaded from installed directory
			if _, err := extRegistry.Get(manifest.Name); err == nil {
				log.Debug().
					Str("extension", manifest.Name).
					Msg("Skipping legacy extension (already loaded from installed directory)")
				continue
			}

			// Register legacy extension
			if err := extRegistry.RegisterOrUpdate(manifest); err != nil {
				failures = append(failures, fmt.Sprintf("%s (legacy): %v", manifest.Name, err))
				log.Warn().
					Str("extension", manifest.Name).
					Err(err).
					Msg("Failed to register legacy extension")
				continue
			}

			loadedCount++
			log.Info().
				Str("extension", manifest.Name).
				Str("version", manifest.Version).
				Msg("Loaded legacy extension")
		}
	}

	return loadedCount, failures
}

// LoadInstalledModules discovers and returns installed modules (for future use)
func LoadInstalledModules() (int, []string) {
	var loadedCount int
	var failures []string

	installedItems, err := discovery.DiscoverEnabledItems()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to discover installed modules")
		return 0, []string{err.Error()}
	}

	for _, item := range installedItems {
		if item.Type != registry.ItemTypeAgenticModule {
			continue
		}

		// Load module manifest
		_, err := packaging.LoadModulePackageManifest(item.Path)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", item.ID, err))
			log.Warn().
				Str("module", item.ID).
				Err(err).
				Msg("Failed to load installed module manifest")
			continue
		}

		loadedCount++
		log.Info().
			Str("module", item.ID).
			Str("name", item.Name).
			Str("version", item.Version).
			Msg("Discovered installed module")
	}

	return loadedCount, failures
}
