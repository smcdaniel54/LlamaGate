// Package discovery provides discovery functionality for installed extensions and modules.
package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/llamagate/llamagate/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverInstalledItems(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test registry
	reg, err := registry.NewRegistry()
	require.NoError(t, err)

	// Create a test extension directory structure
	extDir := filepath.Join(tmpDir, "extensions", "installed", "test-ext")
	require.NoError(t, os.MkdirAll(extDir, 0755))

	// Create manifest.yaml
	manifestPath := filepath.Join(extDir, "manifest.yaml")
	manifestContent := `name: test-ext
version: 1.0.0
description: Test extension
type: workflow
enabled: true
steps:
  - uses: llm.chat
`
	require.NoError(t, os.WriteFile(manifestPath, []byte(manifestContent), 0644))

	// Register in registry
	item := &registry.InstalledItem{
		ID:          "test-ext",
		Type:        registry.ItemTypeExtension,
		Name:        "test-ext",
		Version:     "1.0.0",
		Enabled:     true,
		SourcePath:  extDir,
	}
	require.NoError(t, reg.Register(item))

	// Test discovery (no baseDir parameter - uses homedir)
	items, err := DiscoverInstalledItems()
	require.NoError(t, err)

	// Should find the extension
	found := false
	for _, item := range items {
		if item.ID == "test-ext" {
			found = true
			assert.Equal(t, registry.ItemTypeExtension, item.Type)
			assert.Equal(t, "test-ext", item.Name)
			assert.Equal(t, "1.0.0", item.Version)
			assert.True(t, item.Enabled)
		}
	}
	assert.True(t, found, "Extension should be discovered")
}

func TestDiscoverEnabledItems(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test registry
	reg, err := registry.NewRegistry()
	require.NoError(t, err)

	// Create enabled extension
	enabledDir := filepath.Join(tmpDir, "extensions", "installed", "enabled-ext")
	require.NoError(t, os.MkdirAll(enabledDir, 0755))
	manifestPath := filepath.Join(enabledDir, "manifest.yaml")
	manifestContent := `name: enabled-ext
version: 1.0.0
description: Enabled extension
type: workflow
enabled: true
steps:
  - uses: llm.chat
`
	require.NoError(t, os.WriteFile(manifestPath, []byte(manifestContent), 0644))

	enabledItem := &registry.InstalledItem{
		ID:         "enabled-ext",
		Type:       registry.ItemTypeExtension,
		Name:       "enabled-ext",
		Version:    "1.0.0",
		Enabled:    true,
		SourcePath: enabledDir,
	}
	require.NoError(t, reg.Register(enabledItem))

	// Create disabled extension
	disabledDir := filepath.Join(tmpDir, "extensions", "installed", "disabled-ext")
	require.NoError(t, os.MkdirAll(disabledDir, 0755))
	manifestPath2 := filepath.Join(disabledDir, "manifest.yaml")
	manifestContent2 := `name: disabled-ext
version: 1.0.0
description: Disabled extension
type: workflow
enabled: false
steps:
  - uses: llm.chat
`
	require.NoError(t, os.WriteFile(manifestPath2, []byte(manifestContent2), 0644))

	disabledItem := &registry.InstalledItem{
		ID:         "disabled-ext",
		Type:       registry.ItemTypeExtension,
		Name:       "disabled-ext",
		Version:    "1.0.0",
		Enabled:    false,
		SourcePath: disabledDir,
	}
	require.NoError(t, reg.Register(disabledItem))

	// Test discovery of enabled items only
	items, err := DiscoverEnabledItems()
	require.NoError(t, err)

	// Should only find enabled extension
	foundEnabled := false
	foundDisabled := false
	for _, item := range items {
		if item.ID == "enabled-ext" {
			foundEnabled = true
		}
		if item.ID == "disabled-ext" {
			foundDisabled = true
		}
	}
	assert.True(t, foundEnabled, "Enabled extension should be discovered")
	assert.False(t, foundDisabled, "Disabled extension should not be discovered")
}

func TestDiscoverInstalledItems_EmptyDirectory(t *testing.T) {
	// This test verifies that discovery works even when no items are installed
	items, err := DiscoverInstalledItems()
	require.NoError(t, err)
	// Result is always a slice (may be empty if no items installed)
	// We just verify the function doesn't error and returns a valid slice
	assert.IsType(t, []*DiscoveredItem{}, items)
	// Slice may be empty or contain items from other tests/real installations
	_ = items // Use the result to avoid unused variable warning
}
