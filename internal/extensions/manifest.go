package extensions

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manifest represents an extension manifest loaded from YAML
type Manifest struct {
	Name        string                 `yaml:"name"`
	Version     string                 `yaml:"version"`
	Description string                 `yaml:"description"`
	Type        string                 `yaml:"type"` // "workflow", "middleware", "observer"
	Enabled     *bool                  `yaml:"enabled"` // nil means default (true)
	Config      map[string]interface{} `yaml:"config,omitempty"`
	Inputs      []InputDefinition      `yaml:"inputs,omitempty"`
	Outputs     []OutputDefinition     `yaml:"outputs,omitempty"`
	Steps       []WorkflowStep         `yaml:"steps,omitempty"`
	Hooks       []HookDefinition       `yaml:"hooks,omitempty"`
}

// InputDefinition defines an input parameter
type InputDefinition struct {
	ID       string `yaml:"id"`
	Type     string `yaml:"type"`
	Required bool   `yaml:"required,omitempty"`
}

// OutputDefinition defines an output
type OutputDefinition struct {
	ID   string `yaml:"id"`
	Type string `yaml:"type"`
	Path string `yaml:"path,omitempty"`
}

// WorkflowStep represents a workflow step
type WorkflowStep struct {
	Uses    string                 `yaml:"uses"`
	With    map[string]interface{} `yaml:"with,omitempty"`
	ID      string                 `yaml:"id,omitempty"`
	OnError string                 `yaml:"on_error,omitempty"`
}

// HookDefinition defines a hook (middleware or observer)
type HookDefinition struct {
	On     string                 `yaml:"on"` // "http.request", "llm.response"
	Match  map[string]interface{} `yaml:"match,omitempty"`
	Action string                 `yaml:"action"`
}

// LoadManifest loads a manifest from a file path
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := ValidateManifest(&manifest); err != nil {
		return nil, fmt.Errorf("manifest validation failed: %w", err)
	}

	return &manifest, nil
}

// ValidateManifest validates a manifest
func ValidateManifest(m *Manifest) error {
	if m.Name == "" {
		return fmt.Errorf("name is required")
	}

	// Validate name format (alphanumeric + underscore + hyphen)
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !nameRegex.MatchString(m.Name) {
		return fmt.Errorf("name must be alphanumeric with underscores and hyphens only")
	}

	if m.Version == "" {
		return fmt.Errorf("version is required")
	}

	if m.Description == "" {
		return fmt.Errorf("description is required")
	}

	// Validate type
	validTypes := map[string]bool{
		"workflow":  true,
		"middleware": true,
		"observer":  true,
	}
	if m.Type != "" && !validTypes[m.Type] {
		return fmt.Errorf("type must be one of: workflow, middleware, observer")
	}

	return nil
}

// IsEnabled checks if extension is enabled (defaults to true)
func (m *Manifest) IsEnabled() bool {
	if m.Enabled == nil {
		return true
	}
	return *m.Enabled
}

// DiscoverExtensions discovers all extensions in a directory
func DiscoverExtensions(dir string) ([]*Manifest, error) {
	var manifests []*Manifest

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create extensions directory: %w", err)
		}
		return manifests, nil // Return empty list, directory created
	}

	// Walk directory looking for manifest.yaml files
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for manifest.yaml files
		if info.Name() == "manifest.yaml" {
			manifest, err := LoadManifest(path)
			if err != nil {
				// Log error but continue discovering other extensions
				return fmt.Errorf("failed to load manifest at %s: %w", path, err)
			}
			manifests = append(manifests, manifest)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover extensions: %w", err)
	}

	return manifests, nil
}

// GetExtensionDir returns the directory path for an extension
func GetExtensionDir(baseDir, extensionName string) string {
	// Extension name might have path separators, sanitize them
	sanitized := strings.ReplaceAll(extensionName, string(filepath.Separator), "_")
	return filepath.Join(baseDir, sanitized)
}
