package packaging

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/llamagate/llamagate/internal/homedir"
	"github.com/llamagate/llamagate/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImport_Extension(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test-extension.zip")

	// Create a test extension zip
	createTestExtensionZip(t, zipPath, "test-ext", "1.0.0")

	// Import
	result, err := Import(zipPath)
	require.NoError(t, err)
	assert.Equal(t, "test-ext", result.ID)
	assert.Equal(t, registry.ItemTypeExtension, result.Type)
	assert.Equal(t, "test-ext", result.Name)
	assert.Equal(t, "1.0.0", result.Version)
	assert.True(t, result.Enabled)

	// Verify installed
	extDir, err := homedir.GetExtensionsDir()
	require.NoError(t, err)
	installDir := filepath.Join(extDir, "test-ext")
	assert.DirExists(t, installDir)

	// Verify registry
	reg, err := registry.NewRegistry()
	require.NoError(t, err)
	item, exists := reg.Get("test-ext")
	require.True(t, exists)
	assert.Equal(t, "test-ext", item.ID)
	assert.Equal(t, registry.ItemTypeExtension, item.Type)

	// Cleanup
	_ = Remove("test-ext", registry.ItemTypeExtension)
}

func TestExport_Extension(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test-extension.zip")
	exportPath := filepath.Join(tmpDir, "exported.zip")

	// Create and import test extension
	createTestExtensionZip(t, zipPath, "test-ext-export", "1.0.0")
	_, err := Import(zipPath)
	require.NoError(t, err)

	// Export
	err = Export("test-ext-export", exportPath)
	require.NoError(t, err)
	assert.FileExists(t, exportPath)

	// Verify zip contains manifest
	reader, err := zip.OpenReader(exportPath)
	require.NoError(t, err)
	defer func() { _ = reader.Close() }()

	foundManifest := false
	for _, file := range reader.File {
		if file.Name == "manifest.yaml" || file.Name == "extension.yaml" {
			foundManifest = true
			break
		}
	}
	assert.True(t, foundManifest, "Zip should contain manifest file")

	// Cleanup
	_ = Remove("test-ext-export", registry.ItemTypeExtension)
}

func TestRemove_Extension(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test-extension.zip")

	// Create and import
	createTestExtensionZip(t, zipPath, "test-ext-remove", "1.0.0")
	_, err := Import(zipPath)
	require.NoError(t, err)

	// Verify exists
	reg, err := registry.NewRegistry()
	require.NoError(t, err)
	assert.True(t, reg.Exists("test-ext-remove"))

	// Remove
	err = Remove("test-ext-remove", registry.ItemTypeExtension)
	require.NoError(t, err)

	// Verify removed (get fresh registry instance)
	reg2, err := registry.NewRegistry()
	require.NoError(t, err)
	assert.False(t, reg2.Exists("test-ext-remove"))

	extDir, err := homedir.GetExtensionsDir()
	require.NoError(t, err)
	installDir := filepath.Join(extDir, "test-ext-remove")
	_, err = os.Stat(installDir)
	assert.True(t, os.IsNotExist(err))
}

func TestDetectPackageType(t *testing.T) {
	tmpDir := t.TempDir()

	// Test extension detection
	extDir := filepath.Join(tmpDir, "extension")
	require.NoError(t, os.MkdirAll(extDir, 0755))
	manifestPath := filepath.Join(extDir, "manifest.yaml")
	require.NoError(t, os.WriteFile(manifestPath, []byte(`name: test-ext
version: 1.0.0
description: Test extension
type: workflow
enabled: true
steps:
  - uses: llm.chat
`), 0644))

	pkgType, path, err := DetectPackageType(extDir)
	require.NoError(t, err)
	assert.Equal(t, PackageTypeExtension, pkgType)
	assert.Equal(t, manifestPath, path)
}

// Helper function to create a test extension zip
func createTestExtensionZip(t *testing.T, zipPath, name, version string) {
	zipFile, err := os.Create(zipPath)
	require.NoError(t, err)
	defer func() { _ = zipFile.Close() }()

	zipWriter := zip.NewWriter(zipFile)
	defer func() { _ = zipWriter.Close() }()

	// Add manifest.yaml
	manifestContent := `name: ` + name + `
version: ` + version + `
description: Test extension
type: workflow
enabled: true
steps:
  - uses: llm.chat
`
	manifestEntry, err := zipWriter.Create("manifest.yaml")
	require.NoError(t, err)
	_, err = manifestEntry.Write([]byte(manifestContent))
	require.NoError(t, err)
}
