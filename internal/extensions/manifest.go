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
	Type        string                 `yaml:"type"`    // "workflow", "middleware", "observer"
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

// ValidateManifest validates a manifest with actionable error messages
func ValidateManifest(m *Manifest) error {
	if m.Name == "" {
		return fmt.Errorf("validation error: 'name' field is required in manifest.yaml")
	}

	// Validate name format (alphanumeric + underscore + hyphen)
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !nameRegex.MatchString(m.Name) {
		return fmt.Errorf("validation error: 'name' field '%s' contains invalid characters. Only alphanumeric characters, underscores, and hyphens are allowed", m.Name)
	}

	if m.Version == "" {
		return fmt.Errorf("validation error: 'version' field is required in manifest.yaml")
	}

	// Validate version format (semantic versioning recommended)
	versionRegex := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+`)
	if !versionRegex.MatchString(m.Version) {
		return fmt.Errorf("validation warning: 'version' field '%s' should follow semantic versioning (e.g., '1.0.0')", m.Version)
	}

	if m.Description == "" {
		return fmt.Errorf("validation error: 'description' field is required in manifest.yaml")
	}

	// Validate type
	validTypes := map[string]bool{
		"workflow":   true,
		"middleware": true,
		"observer":   true,
	}
	if m.Type == "" {
		return fmt.Errorf("validation error: 'type' field is required. Must be one of: workflow, middleware, observer")
	}
	if !validTypes[m.Type] {
		return fmt.Errorf("validation error: 'type' field '%s' is invalid. Must be one of: workflow, middleware, observer", m.Type)
	}

	// Validate workflow-specific fields
	if m.Type == "workflow" {
		if len(m.Steps) == 0 {
			return fmt.Errorf("validation error: workflow extension '%s' must have at least one step in 'steps' field", m.Name)
		}
		for i, step := range m.Steps {
			if step.Uses == "" {
				return fmt.Errorf("validation error: workflow step %d in extension '%s' is missing 'uses' field", i, m.Name)
			}
		}
	}

	// Validate middleware/observer-specific fields
	if m.Type == "middleware" || m.Type == "observer" {
		if len(m.Hooks) == 0 {
			return fmt.Errorf("validation error: %s extension '%s' must have at least one hook in 'hooks' field", m.Type, m.Name)
		}
		for i, hook := range m.Hooks {
			if hook.On == "" {
				return fmt.Errorf("validation error: hook %d in extension '%s' is missing 'on' field", i, m.Name)
			}
		}
	}

	// Validate inputs
	for i, input := range m.Inputs {
		if input.ID == "" {
			return fmt.Errorf("validation error: input %d in extension '%s' is missing 'id' field", i, m.Name)
		}
		if input.Type == "" {
			return fmt.Errorf("validation error: input '%s' in extension '%s' is missing 'type' field", input.ID, m.Name)
		}
		validInputTypes := map[string]bool{
			"string":  true,
			"number":  true,
			"boolean": true,
			"object":  true,
			"array":   true,
		}
		if !validInputTypes[input.Type] {
			return fmt.Errorf("validation error: input '%s' in extension '%s' has invalid type '%s'. Must be one of: string, number, boolean, object, array", input.ID, m.Name, input.Type)
		}
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
// Supports both flat structure (extensions/<name>/manifest.yaml) and nested structure
// (e.g., agenticmodules/<module>/extensions/<name>/manifest.yaml)
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
	// This supports nested directories (e.g., agenticmodules/<module>/extensions/<name>/manifest.yaml)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for manifest.yaml files (both in extension directories and nested in modules)
		if info.Name() == "manifest.yaml" {
			manifest, err := LoadManifest(path)
			if err != nil {
				// Log error but continue discovering other extensions
				// Return nil to continue walking, but log the error
				return nil // Continue discovering other extensions
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
