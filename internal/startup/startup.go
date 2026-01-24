// Package startup provides startup integration for loading installed extensions and modules.
package startup

import (
	"fmt"
	"path/filepath"

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

	// 1. Load builtin YAML extensions first (priority loading)
	builtinDir := filepath.Join(legacyBaseDir, "builtin")
	builtinManifests, err := extensions.DiscoverExtensions(builtinDir)
	if err != nil {
		log.Debug().Err(err).Msg("Builtin extension discovery failed or directory doesn't exist")
	} else {
		for _, manifest := range builtinManifests {
			// Set builtin flag
			manifest.Builtin = true
			// Register with extension registry
			if err := extRegistry.RegisterOrUpdate(manifest); err != nil {
				failures = append(failures, fmt.Sprintf("%s (builtin): %v", manifest.Name, err))
				log.Warn().
					Str("extension", manifest.Name).
					Err(err).
					Msg("Failed to register builtin extension")
				continue
			}

			loadedCount++
			log.Info().
				Str("extension", manifest.Name).
				Str("version", manifest.Version).
				Msg("Loaded builtin extension")
		}
	}

	// 2. Discover enabled installed items
	installedItems, err := discovery.DiscoverEnabledItems()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to discover installed items, falling back to legacy directory")
		// Fall through to legacy discovery
	}

	// 3. Load installed extensions
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

	// 4. Discover legacy extensions (only if not already installed)
	// Note: DiscoverExtensions will walk the directory recursively, so we need to filter out builtin/
	legacyManifests, err := extensions.DiscoverExtensions(legacyBaseDir)
	if err != nil {
		log.Debug().Err(err).Msg("Legacy extension discovery failed or directory doesn't exist")
	} else {
		builtinDirPath := filepath.Join(legacyBaseDir, "builtin")
		builtinDirNormalized := filepath.Clean(builtinDirPath)
		
		for _, manifest := range legacyManifests {
			// Check if already loaded from installed directory or builtin directory
			if _, err := extRegistry.Get(manifest.Name); err == nil {
				log.Debug().
					Str("extension", manifest.Name).
					Msg("Skipping legacy extension (already loaded)")
				continue
			}

			// Explicitly filter out extensions from builtin/ subdirectory
			// Since DiscoverExtensions doesn't return file paths, we check:
			// 1. If manifest has builtin flag set (shouldn't happen for legacy, but defensive)
			// 2. If extension is already registered as builtin (from step 1)
			if manifest.Builtin {
				log.Debug().
					Str("extension", manifest.Name).
					Msg("Skipping extension marked as builtin (already loaded from builtin directory)")
				continue
			}
			
			// Additional safety check: verify this isn't a builtin extension by checking registry
			// This handles the case where DiscoverExtensions recursively walks and finds builtin extensions
			if existing, err := extRegistry.Get(manifest.Name); err == nil && existing.Builtin {
				log.Debug().
					Str("extension", manifest.Name).
					Msg("Skipping extension already registered as builtin")
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
		_ = builtinDirNormalized // Suppress unused variable warning (kept for potential future path checking)
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
