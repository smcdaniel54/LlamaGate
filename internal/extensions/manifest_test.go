package extensions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifest(t *testing.T) {
	// Create a temporary manifest file
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.yaml")
	manifestContent := `name: test-extension
version: 1.0.0
description: Test extension
type: workflow
enabled: true
`

	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if manifest.Name != "test-extension" {
		t.Errorf("Expected name 'test-extension', got '%s'", manifest.Name)
	}
	if manifest.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", manifest.Version)
	}
	if manifest.Type != "workflow" {
		t.Errorf("Expected type 'workflow', got '%s'", manifest.Type)
	}
	if !manifest.IsEnabled() {
		t.Error("Expected extension to be enabled")
	}
}

func TestValidateManifest(t *testing.T) {
	tests := []struct {
		name    string
		manifest *Manifest
		wantErr bool
	}{
		{
			name: "valid manifest",
			manifest: &Manifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test description",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			manifest: &Manifest{
				Version:     "1.0.0",
				Description: "Test description",
			},
			wantErr: true,
		},
		{
			name: "missing version",
			manifest: &Manifest{
				Name:        "test",
				Description: "Test description",
			},
			wantErr: true,
		},
		{
			name: "missing description",
			manifest: &Manifest{
				Name:    "test",
				Version: "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "invalid name format",
			manifest: &Manifest{
				Name:        "test@invalid",
				Version:     "1.0.0",
				Description: "Test description",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			manifest: &Manifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test description",
				Type:        "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateManifest(tt.manifest)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateManifest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDiscoverExtensions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create extension directories
	ext1Dir := filepath.Join(tmpDir, "extension1")
	ext2Dir := filepath.Join(tmpDir, "extension2")
	os.MkdirAll(ext1Dir, 0755)
	os.MkdirAll(ext2Dir, 0755)

	// Create valid manifests
	manifest1 := `name: extension1
version: 1.0.0
description: First extension
type: workflow
`
	manifest2 := `name: extension2
version: 2.0.0
description: Second extension
type: middleware
`

	os.WriteFile(filepath.Join(ext1Dir, "manifest.yaml"), []byte(manifest1), 0644)
	os.WriteFile(filepath.Join(ext2Dir, "manifest.yaml"), []byte(manifest2), 0644)

	manifests, err := DiscoverExtensions(tmpDir)
	if err != nil {
		t.Fatalf("Failed to discover extensions: %v", err)
	}

	if len(manifests) != 2 {
		t.Errorf("Expected 2 extensions, got %d", len(manifests))
	}

	// Check extension names
	names := make(map[string]bool)
	for _, m := range manifests {
		names[m.Name] = true
	}
	if !names["extension1"] || !names["extension2"] {
		t.Error("Expected both extensions to be discovered")
	}
}

func TestDiscoverExtensions_InvalidManifest(t *testing.T) {
	tmpDir := t.TempDir()

	extDir := filepath.Join(tmpDir, "invalid")
	os.MkdirAll(extDir, 0755)

	// Create invalid manifest (missing required fields)
	invalidManifest := `name: invalid
`
	os.WriteFile(filepath.Join(extDir, "manifest.yaml"), []byte(invalidManifest), 0644)

	manifests, err := DiscoverExtensions(tmpDir)
	// Should return error for invalid manifest
	if err == nil {
		t.Error("Expected error for invalid manifest")
	}
	if len(manifests) > 0 {
		t.Error("Expected no valid manifests")
	}
}

func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		enabled  *bool
		expected bool
	}{
		{
			name:     "nil enabled (default true)",
			enabled:  nil,
			expected: true,
		},
		{
			name:     "explicitly enabled",
			enabled:  boolPtr(true),
			expected: true,
		},
		{
			name:     "explicitly disabled",
			enabled:  boolPtr(false),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manifest{Enabled: tt.enabled}
			if m.IsEnabled() != tt.expected {
				t.Errorf("IsEnabled() = %v, want %v", m.IsEnabled(), tt.expected)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
