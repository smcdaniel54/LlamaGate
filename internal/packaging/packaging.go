// Package packaging provides import/export functionality for extensions and modules.
package packaging

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/llamagate/llamagate/internal/extensions"
	"github.com/llamagate/llamagate/internal/homedir"
	"github.com/llamagate/llamagate/internal/registry"
	"github.com/rs/zerolog/log"
)

var (
	// importMutex prevents concurrent imports
	importMutex sync.Mutex
)

// ImportResult represents the result of an import operation
type ImportResult struct {
	ID      string
	Type    registry.ItemType
	Name    string
	Version string
	Path    string
	Enabled bool
}

// Import imports a zip file containing an extension or module
func Import(zipPath string) (*ImportResult, error) {
	// Prevent concurrent imports
	importMutex.Lock()
	defer importMutex.Unlock()

	// Get staging directory
	stagingDir, err := homedir.GetImportStagingDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get staging directory: %w", err)
	}

	// Create unique staging directory
	stagingPath, err := os.MkdirTemp(stagingDir, "import-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create staging directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(stagingPath) }() // Clean up staging on exit

	// Extract zip to staging
	if err := extractZip(zipPath, stagingPath); err != nil {
		return nil, fmt.Errorf("failed to extract zip: %w", err)
	}

	// Detect package type
	pkgType, _, err := DetectPackageType(stagingPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect package type: %w", err)
	}

	var result *ImportResult

	// Load and validate manifest based on type
	switch pkgType {
	case PackageTypeExtension:
		pkgManifest, err := LoadExtensionPackageManifest(stagingPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load extension manifest: %w", err)
		}

		if err := ValidatePackageManifest(&pkgManifest.PackageManifest, pkgType); err != nil {
			return nil, fmt.Errorf("manifest validation failed: %w", err)
		}

		// Validate extension manifest
		if err := extensions.ValidateManifest(pkgManifest.ExtensionManifest); err != nil {
			return nil, fmt.Errorf("extension manifest validation failed: %w", err)
		}

		result, err = installExtension(pkgManifest, stagingPath)
		if err != nil {
			return nil, fmt.Errorf("failed to install extension: %w", err)
		}

		// Always trigger automatic discovery after import
		AttemptHotReload()

	case PackageTypeAgenticModule:
		pkgManifest, err := LoadModulePackageManifest(stagingPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load module manifest: %w", err)
		}

		if err := ValidatePackageManifest(&pkgManifest.PackageManifest, pkgType); err != nil {
			return nil, fmt.Errorf("manifest validation failed: %w", err)
		}

		// Validate module manifest
		if err := extensions.ValidateAgenticModuleManifest(pkgManifest.ModuleManifest); err != nil {
			return nil, fmt.Errorf("module manifest validation failed: %w", err)
		}

		result, err = installModule(pkgManifest, stagingPath)
		if err != nil {
			return nil, fmt.Errorf("failed to install module: %w", err)
		}

		// Always trigger automatic discovery after import
		AttemptHotReload()

	default:
		return nil, fmt.Errorf("unknown package type")
	}

	return result, nil
}

// installExtension installs an extension with atomic swap and backup
func installExtension(pkgManifest *ExtensionPackageManifest, stagingPath string) (*ImportResult, error) {
	extDir, err := homedir.GetExtensionsDir()
	if err != nil {
		return nil, err
	}

	installDir := filepath.Join(extDir, pkgManifest.ID)

	// Create backup if existing installation exists
	var backupPath string
	if _, err := os.Stat(installDir); err == nil {
		// Create backup directory structure
		backupBaseDir := filepath.Join(extDir, "backups", pkgManifest.ID)
		if err := os.MkdirAll(backupBaseDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create backup directory: %w", err)
		}

		// Create timestamped backup
		timestamp := time.Now().Format("20060102-150405")
		backupPath = filepath.Join(backupBaseDir, timestamp)

		// Move current installation to backup
		if err := os.Rename(installDir, backupPath); err != nil {
			return nil, fmt.Errorf("failed to create backup: %w", err)
		}

		// Clean up old backups (keep last 2)
		cleanupOldBackups(backupBaseDir, 2)
	}

	// Use atomic install: copy to temp, then rename
	installStagingDir, err := homedir.GetInstallStagingDir()
	if err != nil {
		// Rollback: restore backup if it exists
		if backupPath != "" {
			_ = os.Rename(backupPath, installDir)
		}
		return nil, err
	}

	tempInstallPath := filepath.Join(installStagingDir, pkgManifest.ID)
	
	// Remove temp path if it exists
	_ = os.RemoveAll(tempInstallPath)

	// Copy staging to temp install path
	if err := copyDirectory(stagingPath, tempInstallPath); err != nil {
		// Rollback: restore backup if it exists
		if backupPath != "" {
			_ = os.Rename(backupPath, installDir)
		}
		return nil, fmt.Errorf("failed to copy to staging: %w", err)
	}

	// Atomic swap: rename temp to final location
	if err := os.Rename(tempInstallPath, installDir); err != nil {
		// Rollback: restore backup if it exists
		if backupPath != "" {
			_ = os.Rename(backupPath, installDir)
		}
		return nil, fmt.Errorf("failed to install extension: %w", err)
	}

	// Update registry
	reg, err := registry.NewRegistry()
	if err != nil {
		// Rollback: restore backup and remove new install
		if backupPath != "" {
			_ = os.RemoveAll(installDir)
			_ = os.Rename(backupPath, installDir)
		}
		return nil, fmt.Errorf("failed to open registry: %w", err)
	}

	enabled := true
	if pkgManifest.EnabledDefault != nil {
		enabled = *pkgManifest.EnabledDefault
	}

	item := &registry.InstalledItem{
		ID:          pkgManifest.ID,
		Type:        registry.ItemTypeExtension,
		Name:        pkgManifest.Name,
		Version:     pkgManifest.Version,
		Enabled:     enabled,
		SourcePath:  installDir,
	}

	if err := reg.Register(item); err != nil {
		// Rollback: restore backup and remove new install
		if backupPath != "" {
			_ = os.RemoveAll(installDir)
			_ = os.Rename(backupPath, installDir)
		}
		return nil, fmt.Errorf("failed to register extension: %w", err)
	}

	// Registry update succeeded - remove backup
	if backupPath != "" {
		_ = os.RemoveAll(backupPath)
	}

	return &ImportResult{
		ID:      pkgManifest.ID,
		Type:    registry.ItemTypeExtension,
		Name:    pkgManifest.Name,
		Version: pkgManifest.Version,
		Path:    installDir,
		Enabled: enabled,
	}, nil
}

// installModule installs a module with atomic swap and backup
func installModule(pkgManifest *ModulePackageManifest, stagingPath string) (*ImportResult, error) {
	modulesDir, err := homedir.GetAgenticModulesDir()
	if err != nil {
		return nil, err
	}

	installDir := filepath.Join(modulesDir, pkgManifest.ID)

	// Create backup if existing installation exists
	var backupPath string
	if _, err := os.Stat(installDir); err == nil {
		// Create backup directory structure
		backupBaseDir := filepath.Join(modulesDir, "backups", pkgManifest.ID)
		if err := os.MkdirAll(backupBaseDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create backup directory: %w", err)
		}

		// Create timestamped backup
		timestamp := time.Now().Format("20060102-150405")
		backupPath = filepath.Join(backupBaseDir, timestamp)

		// Move current installation to backup
		if err := os.Rename(installDir, backupPath); err != nil {
			return nil, fmt.Errorf("failed to create backup: %w", err)
		}

		// Clean up old backups (keep last 2)
		cleanupOldBackups(backupBaseDir, 2)
	}

	// Use atomic install: copy to temp, then rename
	installStagingDir, err := homedir.GetInstallStagingDir()
	if err != nil {
		// Rollback: restore backup if it exists
		if backupPath != "" {
			_ = os.Rename(backupPath, installDir)
		}
		return nil, err
	}

	tempInstallPath := filepath.Join(installStagingDir, pkgManifest.ID)
	
	// Remove temp path if it exists
	_ = os.RemoveAll(tempInstallPath)

	// Copy staging to temp install path
	if err := copyDirectory(stagingPath, tempInstallPath); err != nil {
		// Rollback: restore backup if it exists
		if backupPath != "" {
			_ = os.Rename(backupPath, installDir)
		}
		return nil, fmt.Errorf("failed to copy to staging: %w", err)
	}

	// Atomic swap: rename temp to final location
	if err := os.Rename(tempInstallPath, installDir); err != nil {
		// Rollback: restore backup if it exists
		if backupPath != "" {
			_ = os.Rename(backupPath, installDir)
		}
		return nil, fmt.Errorf("failed to install module: %w", err)
	}

	// Update registry
	reg, err := registry.NewRegistry()
	if err != nil {
		// Rollback: restore backup and remove new install
		if backupPath != "" {
			_ = os.RemoveAll(installDir)
			_ = os.Rename(backupPath, installDir)
		}
		return nil, fmt.Errorf("failed to open registry: %w", err)
	}

	enabled := true
	if pkgManifest.EnabledDefault != nil {
		enabled = *pkgManifest.EnabledDefault
	}

	item := &registry.InstalledItem{
		ID:          pkgManifest.ID,
		Type:        registry.ItemTypeAgenticModule,
		Name:        pkgManifest.Name,
		Version:     pkgManifest.Version,
		Enabled:     enabled,
		SourcePath:  installDir,
	}

	if err := reg.Register(item); err != nil {
		// Rollback: restore backup and remove new install
		if backupPath != "" {
			_ = os.RemoveAll(installDir)
			_ = os.Rename(backupPath, installDir)
		}
		return nil, fmt.Errorf("failed to register module: %w", err)
	}

	// Registry update succeeded - remove backup
	if backupPath != "" {
		_ = os.RemoveAll(backupPath)
	}

	return &ImportResult{
		ID:      pkgManifest.ID,
		Type:    registry.ItemTypeAgenticModule,
		Name:    pkgManifest.Name,
		Version: pkgManifest.Version,
		Path:    installDir,
		Enabled: enabled,
	}, nil
}

// Export exports an installed extension or module to a zip file
func Export(id string, outputPath string) error {
	reg, err := registry.NewRegistry()
	if err != nil {
		return fmt.Errorf("failed to open registry: %w", err)
	}

	item, exists := reg.Get(id)
	if !exists {
		return fmt.Errorf("item not found: %s", id)
	}

	// Verify source path exists
	if _, err := os.Stat(item.SourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source path does not exist: %s", item.SourcePath)
	}

	// Create zip file
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer func() { _ = zipFile.Close() }()

	zipWriter := zip.NewWriter(zipFile)
	defer func() { _ = zipWriter.Close() }()

	// Walk directory and add files to zip
	var checksums []string
	err = filepath.Walk(item.SourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path from source
		relPath, err := filepath.Rel(item.SourcePath, path)
		if err != nil {
			return err
		}

		// Use forward slashes in zip (zip standard)
		zipPath := filepath.ToSlash(relPath)

		// Open file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()

		// Read file content
		content, err := io.ReadAll(file)
		if err != nil {
			return err
		}

		// Calculate checksum
		hash := sha256.Sum256(content)
		checksum := hex.EncodeToString(hash[:])
		checksums = append(checksums, fmt.Sprintf("%s  %s", checksum, zipPath))

		// Create zip entry
		zipEntry, err := zipWriter.Create(zipPath)
		if err != nil {
			return err
		}

		// Write content
		_, err = zipEntry.Write(content)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to create zip: %w", err)
	}

	// Add checksums.txt
	if len(checksums) > 0 {
		checksumsEntry, err := zipWriter.Create("checksums.txt")
		if err != nil {
			return fmt.Errorf("failed to create checksums entry: %w", err)
		}
		_, err = checksumsEntry.Write([]byte(strings.Join(checksums, "\n") + "\n"))
		if err != nil {
			return fmt.Errorf("failed to write checksums: %w", err)
		}
	}

	return nil
}

// Remove removes an installed extension or module
func Remove(id string, itemType registry.ItemType) error {
	reg, err := registry.NewRegistry()
	if err != nil {
		return fmt.Errorf("failed to open registry: %w", err)
	}

	item, exists := reg.Get(id)
	if !exists {
		return fmt.Errorf("item not found: %s", id)
	}

	// Verify item type matches
	if item.Type != itemType {
		return fmt.Errorf("item type mismatch: expected %s, got %s", itemType, item.Type)
	}

	// Remove from filesystem
	if _, err := os.Stat(item.SourcePath); err == nil {
		if err := os.RemoveAll(item.SourcePath); err != nil {
			return fmt.Errorf("failed to remove directory: %w", err)
		}
	}

	// Remove backups (if any)
	backupBaseDir := filepath.Join(filepath.Dir(item.SourcePath), "backups", id)
	if _, err := os.Stat(backupBaseDir); err == nil {
		// Clean up backups (non-critical, log but don't fail)
		_ = os.RemoveAll(backupBaseDir)
	}

	// Remove from registry
	if err := reg.Unregister(id); err != nil {
		return fmt.Errorf("failed to remove from registry: %w", err)
	}

	return nil
}

// extractZip extracts a zip file to a directory
func extractZip(zipPath, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer func() { _ = reader.Close() }()

	for _, file := range reader.File {
		path := filepath.Join(destDir, file.Name)

		// Check for zip slip vulnerability
		if !strings.HasPrefix(path, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			_ = os.MkdirAll(path, file.FileInfo().Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			_ = outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		_ = outFile.Close()
		_ = rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
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

// cleanupOldBackups removes old backups, keeping only the last N
func cleanupOldBackups(backupDir string, keepCount int) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return // Non-critical, ignore errors
	}

	// Sort by name (timestamp format ensures chronological order)
	backups := make([]os.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			backups = append(backups, entry)
		}
	}

	// Remove oldest backups if we have more than keepCount
	if len(backups) > keepCount {
		// Sort by name (timestamp) - oldest first
		for i := 0; i < len(backups)-keepCount; i++ {
			oldBackup := filepath.Join(backupDir, backups[i].Name())
			_ = os.RemoveAll(oldBackup) // Non-critical cleanup
		}
	}
}

// AttemptHotReload attempts to trigger automatic discovery by calling the refresh endpoint
// This is a best-effort operation - if the server is not running, it fails silently
func AttemptHotReload() {
	// Default LlamaGate API endpoint
	refreshURL := "http://localhost:11435/v1/extensions/refresh"
	
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("POST", refreshURL, nil)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create hot reload request")
		return
	}

	// Try to get API key from environment (if set)
	if apiKey := os.Getenv("LLAMAGATE_API_KEY"); apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Debug().Err(err).Msg("Hot reload skipped (server not running or unreachable)")
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent {
		log.Info().Msg("Hot reload triggered successfully")
	} else {
		log.Debug().
			Int("status_code", resp.StatusCode).
			Msg("Hot reload request returned non-success status")
	}
}
