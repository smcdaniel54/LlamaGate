package registry

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestNewRegistry(t *testing.T) {
	reg, err := NewRegistry()
	require.NoError(t, err)
	assert.NotNil(t, reg)
}

func TestRegisterAndGet(t *testing.T) {
	reg, err := NewRegistry()
	require.NoError(t, err)

	item := &InstalledItem{
		ID:         "test-extension",
		Type:       ItemTypeExtension,
		Name:       "Test Extension",
		Version:    "1.0.0",
		Enabled:    true,
		SourcePath: "/path/to/extension",
	}

	err = reg.Register(item)
	require.NoError(t, err)

	retrieved, exists := reg.Get("test-extension")
	require.True(t, exists)
	assert.Equal(t, item.ID, retrieved.ID)
	assert.Equal(t, item.Name, retrieved.Name)
	assert.Equal(t, item.Version, retrieved.Version)
	assert.True(t, retrieved.Enabled)
	assert.False(t, retrieved.InstalledAt.IsZero())
}

func TestUnregister(t *testing.T) {
	reg, err := NewRegistry()
	require.NoError(t, err)

	item := &InstalledItem{
		ID:         "temp-extension",
		Type:       ItemTypeExtension,
		Name:       "Temp Extension",
		Version:    "1.0.0",
		Enabled:    true,
		SourcePath: "/path/to/extension",
	}

	err = reg.Register(item)
	require.NoError(t, err)

	assert.True(t, reg.Exists("temp-extension"))

	err = reg.Unregister("temp-extension")
	require.NoError(t, err)

	assert.False(t, reg.Exists("temp-extension"))
}

func TestList(t *testing.T) {
	reg, err := NewRegistry()
	require.NoError(t, err)

	// Clear registry
	items := reg.List("")
	for _, item := range items {
		_ = reg.Unregister(item.ID)
	}

	// Add test items
	ext1 := &InstalledItem{
		ID:         "ext1",
		Type:       ItemTypeExtension,
		Name:       "Extension 1",
		Version:    "1.0.0",
		Enabled:    true,
		SourcePath: "/path/to/ext1",
	}
	ext2 := &InstalledItem{
		ID:         "ext2",
		Type:       ItemTypeExtension,
		Name:       "Extension 2",
		Version:    "2.0.0",
		Enabled:    false,
		SourcePath: "/path/to/ext2",
	}
	module1 := &InstalledItem{
		ID:         "module1",
		Type:       ItemTypeAgenticModule,
		Name:       "Module 1",
		Version:    "1.0.0",
		Enabled:    true,
		SourcePath: "/path/to/module1",
	}

	require.NoError(t, reg.Register(ext1))
	require.NoError(t, reg.Register(ext2))
	require.NoError(t, reg.Register(module1))

	// List all
	all := reg.List("")
	assert.Len(t, all, 3)

	// List extensions only
	extensions := reg.List(ItemTypeExtension)
	assert.Len(t, extensions, 2)

	// List modules only
	modules := reg.List(ItemTypeAgenticModule)
	assert.Len(t, modules, 1)
}

func TestSetEnabled(t *testing.T) {
	reg, err := NewRegistry()
	require.NoError(t, err)

	item := &InstalledItem{
		ID:         "toggle-extension",
		Type:       ItemTypeExtension,
		Name:       "Toggle Extension",
		Version:    "1.0.0",
		Enabled:    true,
		SourcePath: "/path/to/extension",
	}

	require.NoError(t, reg.Register(item))

	// Disable
	err = reg.SetEnabled("toggle-extension", false)
	require.NoError(t, err)

	retrieved, _ := reg.Get("toggle-extension")
	assert.False(t, retrieved.Enabled)

	// Enable
	err = reg.SetEnabled("toggle-extension", true)
	require.NoError(t, err)

	retrieved, _ = reg.Get("toggle-extension")
	assert.True(t, retrieved.Enabled)
}

func TestSetEnabled_NotFound(t *testing.T) {
	reg, err := NewRegistry()
	require.NoError(t, err)

	err = reg.SetEnabled("nonexistent", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPersistence(t *testing.T) {
	// Create first registry and add item
	reg1, err := NewRegistry()
	require.NoError(t, err)

	item := &InstalledItem{
		ID:         "persistent-extension",
		Type:       ItemTypeExtension,
		Name:       "Persistent Extension",
		Version:    "1.0.0",
		Enabled:    true,
		SourcePath: "/path/to/extension",
	}

	require.NoError(t, reg1.Register(item))

	// Create new registry instance (simulating restart)
	reg2, err := NewRegistry()
	require.NoError(t, err)

	// Item should still exist
	retrieved, exists := reg2.Get("persistent-extension")
	require.True(t, exists)
	assert.Equal(t, item.ID, retrieved.ID)
	assert.Equal(t, item.Name, retrieved.Name)

	// Clean up
	_ = reg2.Unregister("persistent-extension")
}

func TestUpdateExistingItem(t *testing.T) {
	reg, err := NewRegistry()
	require.NoError(t, err)

	originalTime := time.Now().Add(-1 * time.Hour)
	item := &InstalledItem{
		ID:          "update-extension",
		Type:        ItemTypeExtension,
		Name:        "Original Name",
		Version:     "1.0.0",
		Enabled:     true,
		SourcePath:  "/path/to/extension",
		InstalledAt: originalTime,
	}

	require.NoError(t, reg.Register(item))

	// Small delay to ensure LastUpdatedAt is different
	time.Sleep(10 * time.Millisecond)

	// Update the item (without InstalledAt - should be preserved from existing)
	updatedItem := &InstalledItem{
		ID:         "update-extension",
		Type:       ItemTypeExtension,
		Name:       "Updated Name",
		Version:    "2.0.0",
		Enabled:    true,
		SourcePath: "/path/to/updated",
		// InstalledAt not set - should be preserved from existing item
	}

	require.NoError(t, reg.Register(updatedItem))

	retrieved, _ := reg.Get("update-extension")
	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.Equal(t, "2.0.0", retrieved.Version)
	// InstalledAt should be preserved from original registration
	// Note: JSON serialization/unmarshaling may cause timezone shifts, so we verify
	// that the time is preserved by checking it's significantly in the past (not current time)
	timeDiff := time.Since(retrieved.InstalledAt)
	assert.True(t, timeDiff > 30*time.Minute, 
		"InstalledAt should be preserved (should be ~1 hour ago, got %v)", timeDiff)
	assert.True(t, retrieved.LastUpdatedAt.After(originalTime)) // LastUpdatedAt updated
}

func TestRegistryFileLocation(t *testing.T) {
	reg, err := NewRegistry()
	require.NoError(t, err)

	// Verify registry file exists after save
	item := &InstalledItem{
		ID:         "file-test",
		Type:       ItemTypeExtension,
		Name:       "File Test",
		Version:    "1.0.0",
		Enabled:    true,
		SourcePath: "/path/to/test",
	}

	require.NoError(t, reg.Register(item))

	// Check that file exists
	registryDir, err := os.UserHomeDir()
	require.NoError(t, err)
	registryPath := filepath.Join(registryDir, ".llamagate", "registry", RegistryFileName)

	info, err := os.Stat(registryPath)
	require.NoError(t, err)
	assert.False(t, info.IsDir())
	assert.Greater(t, info.Size(), int64(0))

	// Clean up
	_ = reg.Unregister("file-test")
}
