package packaging

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePackageManifest(t *testing.T) {
	tests := []struct {
		name    string
		manifest PackageManifest
		pkgType PackageType
		wantErr bool
	}{
		{
			name: "valid extension manifest",
			manifest: PackageManifest{
				ID:      "test-ext",
				Name:    "Test Extension",
				Version: "1.0.0",
			},
			pkgType: PackageTypeExtension,
			wantErr: false,
		},
		{
			name: "missing id",
			manifest: PackageManifest{
				Name:    "Test Extension",
				Version: "1.0.0",
			},
			pkgType: PackageTypeExtension,
			wantErr: true,
		},
		{
			name: "invalid id format",
			manifest: PackageManifest{
				ID:      "test.ext", // Contains dot
				Name:    "Test Extension",
				Version: "1.0.0",
			},
			pkgType: PackageTypeExtension,
			wantErr: true,
		},
		{
			name: "invalid load_mode",
			manifest: PackageManifest{
				ID:       "test-ext",
				Name:     "Test Extension",
				Version:  "1.0.0",
				LoadMode: "invalid",
			},
			pkgType: PackageTypeExtension,
			wantErr: true,
		},
		{
			name: "valid load_mode eager",
			manifest: PackageManifest{
				ID:       "test-ext",
				Name:     "Test Extension",
				Version:  "1.0.0",
				LoadMode: "eager",
			},
			pkgType: PackageTypeExtension,
			wantErr: false,
		},
		{
			name: "valid load_mode lazy",
			manifest: PackageManifest{
				ID:       "test-ext",
				Name:     "Test Extension",
				Version:  "1.0.0",
				LoadMode: "lazy",
			},
			pkgType: PackageTypeExtension,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePackageManifest(&tt.manifest, tt.pkgType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadExtensionPackageManifest(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manifest.yaml
	manifestPath := filepath.Join(tmpDir, "manifest.yaml")
	manifestContent := `name: test-ext
version: 1.0.0
description: Test extension
type: workflow
enabled: true
steps:
  - uses: llm.chat
`
	require.NoError(t, os.WriteFile(manifestPath, []byte(manifestContent), 0644))

	// Load
	pkgManifest, err := LoadExtensionPackageManifest(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, "test-ext", pkgManifest.ID)
	assert.Equal(t, "test-ext", pkgManifest.Name)
	assert.Equal(t, "1.0.0", pkgManifest.Version)
	assert.NotNil(t, pkgManifest.ExtensionManifest)
}
