// Package homedir provides home directory resolution for LlamaGate.
// It handles platform-specific paths for storing installed extensions and modules.
package homedir

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	// HomeDirName is the name of the LlamaGate home directory
	HomeDirName = ".llamagate"
)

// GetHomeDir returns the LlamaGate home directory path.
// Creates the directory if it doesn't exist.
func GetHomeDir() (string, error) {
	var homeDir string

	// Get user home directory
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Build LlamaGate home directory path
	homeDir = filepath.Join(userHome, HomeDirName)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create LlamaGate home directory: %w", err)
	}

	return homeDir, nil
}

// GetExtensionsDir returns the path to the installed extensions directory.
func GetExtensionsDir() (string, error) {
	homeDir, err := GetHomeDir()
	if err != nil {
		return "", err
	}
	extDir := filepath.Join(homeDir, "extensions", "installed")
	if err := os.MkdirAll(extDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create extensions directory: %w", err)
	}
	return extDir, nil
}

// GetAgenticModulesDir returns the path to the installed agentic modules directory.
func GetAgenticModulesDir() (string, error) {
	homeDir, err := GetHomeDir()
	if err != nil {
		return "", err
	}
	modulesDir := filepath.Join(homeDir, "agentic-modules", "installed")
	if err := os.MkdirAll(modulesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create agentic-modules directory: %w", err)
	}
	return modulesDir, nil
}

// GetRegistryDir returns the path to the registry directory.
func GetRegistryDir() (string, error) {
	homeDir, err := GetHomeDir()
	if err != nil {
		return "", err
	}
	registryDir := filepath.Join(homeDir, "registry")
	if err := os.MkdirAll(registryDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create registry directory: %w", err)
	}
	return registryDir, nil
}

// GetTempDir returns the path to the temporary staging directory.
func GetTempDir() (string, error) {
	homeDir, err := GetHomeDir()
	if err != nil {
		return "", err
	}
	tempDir := filepath.Join(homeDir, "tmp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	return tempDir, nil
}

// GetImportStagingDir returns a temporary directory for staging imports.
func GetImportStagingDir() (string, error) {
	tempDir, err := GetTempDir()
	if err != nil {
		return "", err
	}
	importDir := filepath.Join(tempDir, "import")
	if err := os.MkdirAll(importDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create import staging directory: %w", err)
	}
	return importDir, nil
}

// GetInstallStagingDir returns a temporary directory for staging installs.
func GetInstallStagingDir() (string, error) {
	tempDir, err := GetTempDir()
	if err != nil {
		return "", err
	}
	installDir := filepath.Join(tempDir, "install")
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create install staging directory: %w", err)
	}
	return installDir, nil
}

// Platform returns the current platform (windows, linux, darwin).
func Platform() string {
	return runtime.GOOS
}
