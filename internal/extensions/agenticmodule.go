package extensions

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AgenticModuleManifest represents an AgenticModule manifest
type AgenticModuleManifest struct {
	Name        string                 `yaml:"name"`
	Version     string                 `yaml:"version"`
	Description string                 `yaml:"description"`
	Steps       []AgenticModuleStep    `yaml:"steps"`
	Config      map[string]interface{} `yaml:"config,omitempty"`
}

// AgenticModuleStep represents a step in an AgenticModule workflow
type AgenticModuleStep struct {
	Extension string                 `yaml:"extension"`
	Input     map[string]interface{} `yaml:"input,omitempty"`
	OnError   string                 `yaml:"on_error,omitempty"` // "stop" or "continue"
}

// LoadAgenticModuleManifest loads an AgenticModule manifest from a file
func LoadAgenticModuleManifest(path string) (*AgenticModuleManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read agenticmodule.yaml: %w", err)
	}

	var manifest AgenticModuleManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := ValidateAgenticModuleManifest(&manifest); err != nil {
		return nil, fmt.Errorf("manifest validation failed: %w", err)
	}

	return &manifest, nil
}

// ValidateAgenticModuleManifest validates an AgenticModule manifest
func ValidateAgenticModuleManifest(m *AgenticModuleManifest) error {
	if m.Name == "" {
		return fmt.Errorf("validation error: 'name' field is required in agenticmodule.yaml")
	}

	if m.Version == "" {
		return fmt.Errorf("validation error: 'version' field is required in agenticmodule.yaml")
	}

	if m.Description == "" {
		return fmt.Errorf("validation error: 'description' field is required in agenticmodule.yaml")
	}

	if len(m.Steps) == 0 {
		return fmt.Errorf("validation error: module '%s' must have at least one step", m.Name)
	}

	for i, step := range m.Steps {
		if step.Extension == "" {
			return fmt.Errorf("validation error: module step %d is missing 'extension' field", i)
		}

		if step.OnError != "" && step.OnError != "stop" && step.OnError != "continue" {
			return fmt.Errorf("validation error: module step %d has invalid 'on_error' value '%s'. Must be 'stop' or 'continue'", i, step.OnError)
		}
	}

	return nil
}

// DiscoverAgenticModules discovers all AgenticModules in a directory
func DiscoverAgenticModules(dir string) ([]*AgenticModuleManifest, error) {
	var manifests []*AgenticModuleManifest

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return manifests, nil // Return empty list
	}

	// Walk directory looking for agenticmodule.yaml files
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for agenticmodule.yaml files
		if info.Name() == "agenticmodule.yaml" {
			manifest, err := LoadAgenticModuleManifest(path)
			if err != nil {
				// Log error but continue discovering other modules
				return fmt.Errorf("failed to load agenticmodule.yaml at %s: %w", path, err)
			}
			manifests = append(manifests, manifest)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover agenticmodules: %w", err)
	}

	return manifests, nil
}
