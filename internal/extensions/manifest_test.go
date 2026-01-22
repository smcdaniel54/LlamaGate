package extensions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverExtensions_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create flat structure
	flatExtDir := filepath.Join(tmpDir, "flat", "ext1")
	require.NoError(t, os.MkdirAll(flatExtDir, 0755))
	writeTestManifest(t, flatExtDir, "ext1", "1.0.0")

	// Create nested structure (simulating agenticmodules)
	nestedExtDir := filepath.Join(tmpDir, "nested", "module1", "extensions", "ext2")
	require.NoError(t, os.MkdirAll(nestedExtDir, 0755))
	writeTestManifest(t, nestedExtDir, "ext2", "1.0.0")

	// Discover from flat directory
	flatManifests, err := DiscoverExtensions(filepath.Join(tmpDir, "flat"))
	require.NoError(t, err)
	assert.Len(t, flatManifests, 1)
	assert.Equal(t, "ext1", flatManifests[0].Name)

	// Discover from nested directory
	nestedManifests, err := DiscoverExtensions(filepath.Join(tmpDir, "nested"))
	require.NoError(t, err)
	assert.Len(t, nestedManifests, 1)
	assert.Equal(t, "ext2", nestedManifests[0].Name)
}

func writeTestManifest(t *testing.T, dir, name, version string) {
	manifest := `name: ` + name + `
version: ` + version + `
description: Test extension
type: workflow
enabled: true
steps:
  - uses: llm.chat
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "manifest.yaml"), []byte(manifest), 0644))
}

func TestValidateManifest_ActionableErrors(t *testing.T) {
	tests := []struct {
		name     string
		manifest *Manifest
		wantErr  bool
		errMsg   string
	}{
		{
			name: "missing name",
			manifest: &Manifest{
				Version:     "1.0.0",
				Description: "Test",
				Type:        "workflow",
			},
			wantErr: true,
			errMsg:  "name' field is required",
		},
		{
			name: "invalid name format",
			manifest: &Manifest{
				Name:        "invalid name!",
				Version:     "1.0.0",
				Description: "Test",
				Type:        "workflow",
			},
			wantErr: true,
			errMsg:  "contains invalid characters",
		},
		{
			name: "missing version",
			manifest: &Manifest{
				Name:        "test",
				Description: "Test",
				Type:        "workflow",
			},
			wantErr: true,
			errMsg:  "version' field is required",
		},
		{
			name: "invalid type",
			manifest: &Manifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test",
				Type:        "invalid",
			},
			wantErr: true,
			errMsg:  "type' field 'invalid' is invalid",
		},
		{
			name: "workflow without steps",
			manifest: &Manifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test",
				Type:        "workflow",
				Steps:       []WorkflowStep{},
			},
			wantErr: true,
			errMsg:  "must have at least one step",
		},
		{
			name: "valid manifest",
			manifest: &Manifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test",
				Type:        "workflow",
				Steps: []WorkflowStep{
					{Uses: "llm.chat"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateManifest(tt.manifest)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateManifest_Endpoints(t *testing.T) {
	tests := []struct {
		name     string
		manifest *Manifest
		wantErr  bool
		errMsg   string
	}{
		{
			name: "non-workflow with endpoints",
			manifest: &Manifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test",
				Type:        "middleware",
				Hooks: []HookDefinition{
					{On: "http.request", Action: "log"},
				},
				Endpoints: []EndpointDefinition{
					{Path: "/test", Method: "GET", Description: "Test"},
				},
			},
			wantErr: true,
			errMsg:  "only workflow extensions can define endpoints",
		},
		{
			name: "endpoint missing path",
			manifest: &Manifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test",
				Type:        "workflow",
				Steps: []WorkflowStep{
					{Uses: "llm.chat"},
				},
				Endpoints: []EndpointDefinition{
					{Method: "GET", Description: "Test"},
				},
			},
			wantErr: true,
			errMsg:  "missing 'path' field",
		},
		{
			name: "endpoint path without leading slash",
			manifest: &Manifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test",
				Type:        "workflow",
				Steps: []WorkflowStep{
					{Uses: "llm.chat"},
				},
				Endpoints: []EndpointDefinition{
					{Path: "test", Method: "GET", Description: "Test"},
				},
			},
			wantErr: true,
			errMsg:  "Path must start with '/'",
		},
		{
			name: "endpoint missing method",
			manifest: &Manifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test",
				Type:        "workflow",
				Steps: []WorkflowStep{
					{Uses: "llm.chat"},
				},
				Endpoints: []EndpointDefinition{
					{Path: "/test", Description: "Test"},
				},
			},
			wantErr: true,
			errMsg:  "missing 'method' field",
		},
		{
			name: "endpoint invalid method",
			manifest: &Manifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test",
				Type:        "workflow",
				Steps: []WorkflowStep{
					{Uses: "llm.chat"},
				},
				Endpoints: []EndpointDefinition{
					{Path: "/test", Method: "INVALID", Description: "Test"},
				},
			},
			wantErr: true,
			errMsg:  "invalid method",
		},
		{
			name: "valid endpoints",
			manifest: &Manifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test",
				Type:        "workflow",
				Steps: []WorkflowStep{
					{Uses: "llm.chat"},
				},
				Endpoints: []EndpointDefinition{
					{Path: "/test", Method: "GET", Description: "Test endpoint"},
					{Path: "/post", Method: "POST", Description: "POST endpoint"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateManifest(tt.manifest)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
