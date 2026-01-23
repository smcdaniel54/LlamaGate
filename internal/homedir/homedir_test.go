package homedir

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHomeDir(t *testing.T) {
	homeDir, err := GetHomeDir()
	require.NoError(t, err)
	assert.NotEmpty(t, homeDir)
	assert.Contains(t, homeDir, HomeDirName)

	// Verify directory exists
	info, err := os.Stat(homeDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestGetExtensionsDir(t *testing.T) {
	extDir, err := GetExtensionsDir()
	require.NoError(t, err)
	assert.Contains(t, extDir, "extensions")
	assert.Contains(t, extDir, "installed")

	// Verify directory exists
	info, err := os.Stat(extDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestGetAgenticModulesDir(t *testing.T) {
	modulesDir, err := GetAgenticModulesDir()
	require.NoError(t, err)
	assert.Contains(t, modulesDir, "agentic-modules")
	assert.Contains(t, modulesDir, "installed")

	// Verify directory exists
	info, err := os.Stat(modulesDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestGetRegistryDir(t *testing.T) {
	registryDir, err := GetRegistryDir()
	require.NoError(t, err)
	assert.Contains(t, registryDir, "registry")

	// Verify directory exists
	info, err := os.Stat(registryDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestGetTempDir(t *testing.T) {
	tempDir, err := GetTempDir()
	require.NoError(t, err)
	assert.Contains(t, tempDir, "tmp")

	// Verify directory exists
	info, err := os.Stat(tempDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestGetImportStagingDir(t *testing.T) {
	importDir, err := GetImportStagingDir()
	require.NoError(t, err)
	assert.Contains(t, importDir, "import")

	// Verify directory exists
	info, err := os.Stat(importDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestGetInstallStagingDir(t *testing.T) {
	installDir, err := GetInstallStagingDir()
	require.NoError(t, err)
	assert.Contains(t, installDir, "install")

	// Verify directory exists
	info, err := os.Stat(installDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestPlatform(t *testing.T) {
	platform := Platform()
	assert.NotEmpty(t, platform)
	assert.Contains(t, []string{"windows", "linux", "darwin"}, platform)
}

func TestDirectoryStructure(t *testing.T) {
	homeDir, err := GetHomeDir()
	require.NoError(t, err)

	// Verify all expected subdirectories exist
	expectedDirs := []string{
		filepath.Join(homeDir, "extensions", "installed"),
		filepath.Join(homeDir, "agentic-modules", "installed"),
		filepath.Join(homeDir, "registry"),
		filepath.Join(homeDir, "tmp", "import"),
		filepath.Join(homeDir, "tmp", "install"),
	}

	for _, dir := range expectedDirs {
		info, err := os.Stat(dir)
		require.NoError(t, err, "Directory should exist: %s", dir)
		assert.True(t, info.IsDir(), "Should be a directory: %s", dir)
	}
}
