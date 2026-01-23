// Package migration provides migration functionality for legacy extensions.
package migration

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/llamagate/llamagate/internal/discovery"
	"github.com/llamagate/llamagate/internal/homedir"
	"github.com/llamagate/llamagate/internal/packaging"
	"github.com/llamagate/llamagate/internal/registry"
	"github.com/rs/zerolog/log"
)

const (
	// MigrationMarkerFile indicates migration has been completed
	MigrationMarkerFile = ".migration_complete"
)

// MigrationResult represents the result of a migration operation
type MigrationResult struct {
	MigratedExtensions int
	MigratedModules    int
	Failed             []string
	LegacyPath         string
}

// HasMigrated checks if migration has already been performed
func HasMigrated() (bool, error) {
	homeDir, err := homedir.GetHomeDir()
	if err != nil {
		return false, err
	}

	markerPath := filepath.Join(homeDir, MigrationMarkerFile)
	_, err = os.Stat(markerPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// MarkMigrated creates a marker file indicating migration is complete
func MarkMigrated() error {
	homeDir, err := homedir.GetHomeDir()
	if err != nil {
		return err
	}

	markerPath := filepath.Join(homeDir, MigrationMarkerFile)
	return os.WriteFile(markerPath, []byte("migration completed"), 0644)
}

// MigrateLegacyExtensions migrates extensions from legacy directory to installed directory
func MigrateLegacyExtensions(legacyExtDir string) (*MigrationResult, error) {
	result := &MigrationResult{
		LegacyPath:         legacyExtDir,
		MigratedExtensions: 0,
		MigratedModules:    0,
		Failed:             []string{},
	}

	// Check if already migrated
	alreadyMigrated, err := HasMigrated()
	if err != nil {
		return nil, fmt.Errorf("failed to check migration status: %w", err)
	}
	if alreadyMigrated {
		log.Info().Msg("Migration already completed, skipping")
		return result, nil
	}

	// Check if legacy directory exists
	if _, err := os.Stat(legacyExtDir); os.IsNotExist(err) {
		log.Info().Str("path", legacyExtDir).Msg("Legacy extensions directory not found, nothing to migrate")
		_ = MarkMigrated() // Mark as complete even if nothing to migrate
		return result, nil
	}

	// Discover legacy extensions
	legacyItems, err := discovery.DiscoverLegacyExtensions(legacyExtDir)
	if err != nil {
		return nil, fmt.Errorf("failed to discover legacy extensions: %w", err)
	}

	if len(legacyItems) == 0 {
		log.Info().Msg("No legacy extensions found to migrate")
		_ = MarkMigrated()
		return result, nil
	}

	log.Info().
		Int("count", len(legacyItems)).
		Str("legacy_path", legacyExtDir).
		Msg("Starting migration of legacy extensions")

	// Get installed extensions directory
	extDir, err := homedir.GetExtensionsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get extensions directory: %w", err)
	}

	// Get registry
	reg, err := registry.NewRegistry()
	if err != nil {
		return nil, fmt.Errorf("failed to open registry: %w", err)
	}

	// Migrate each extension
	for _, item := range legacyItems {
		// Check if already installed
		if reg.Exists(item.ID) {
			log.Debug().
				Str("extension", item.ID).
				Msg("Skipping migration: extension already installed")
			continue
		}

		// Copy to installed directory
		installDir := filepath.Join(extDir, item.ID)
		if err := copyDirectory(item.Path, installDir); err != nil {
			result.Failed = append(result.Failed, fmt.Sprintf("%s: %v", item.ID, err))
			log.Warn().
				Str("extension", item.ID).
				Err(err).
				Msg("Failed to migrate extension")
			continue
		}

		// Register in registry
		regItem := &registry.InstalledItem{
			ID:          item.ID,
			Type:        item.Type,
			Name:        item.Name,
			Version:     item.Version,
			Enabled:     item.Enabled,
			SourcePath:  installDir,
		}

		if err := reg.Register(regItem); err != nil {
			result.Failed = append(result.Failed, fmt.Sprintf("%s (registry): %v", item.ID, err))
			log.Warn().
				Str("extension", item.ID).
				Err(err).
				Msg("Failed to register migrated extension")
			// Clean up copied directory
			_ = os.RemoveAll(installDir)
			continue
		}

		result.MigratedExtensions++
		log.Info().
			Str("extension", item.ID).
			Str("version", item.Version).
			Msg("Migrated extension")
	}

	// Mark migration as complete
	if err := MarkMigrated(); err != nil {
		log.Warn().Err(err).Msg("Failed to create migration marker file")
	}

	log.Info().
		Int("migrated", result.MigratedExtensions).
		Int("failed", len(result.Failed)).
		Msg("Migration complete")

	// Trigger automatic discovery if extensions were migrated
	if result.MigratedExtensions > 0 {
		packaging.AttemptHotReload()
	}

	return result, nil
}

// copyDirectory copies a directory recursively
func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = srcFile.Close() }()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer func() { _ = dstFile.Close() }()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}
