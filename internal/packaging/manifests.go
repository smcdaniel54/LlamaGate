// Package packaging provides import/export functionality for extensions and modules.
package packaging

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/llamagate/llamagate/internal/extensions"
	"gopkg.in/yaml.v3"
)

var (
	// idRegex validates ID format (alphanumeric + underscore + hyphen, same as name)
	idRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// PackageManifest represents the packaging-specific manifest fields
// This extends the base extension/manifest with packaging metadata
type PackageManifest struct {
	// Common fields
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description,omitempty"`

	// Extension-specific fields (for extension.yaml)
	EnabledDefault *bool  `yaml:"enabled_default,omitempty"`
	HotReload      *bool  `yaml:"hot_reload,omitempty"`
	LoadMode       string `yaml:"load_mode,omitempty"` // "eager" or "lazy"
	Autostart      *bool  `yaml:"autostart,omitempty"`

	// Module-specific fields (for module.yaml)
	EntryWorkflow string `yaml:"entry_workflow,omitempty"`
}

// ExtensionPackageManifest combines packaging manifest with full extension manifest
type ExtensionPackageManifest struct {
	PackageManifest
	ExtensionManifest *extensions.Manifest
}

// ModulePackageManifest combines packaging manifest with full module manifest
type ModulePackageManifest struct {
	PackageManifest
	ModuleManifest *extensions.AgenticModuleManifest
}

// DetectPackageType detects whether a directory contains an extension or module
// by checking for manifest.yaml (extension) or agenticmodule.yaml (module)
// Note: extension.yaml and module.yaml are packaging metadata only, not full manifests
func DetectPackageType(dir string) (PackageType, string, error) {
	// Check for manifest.yaml (extension manifest - primary)
	manifestPath := filepath.Join(dir, "manifest.yaml")
	if _, err := os.Stat(manifestPath); err == nil {
		// Try to load as extension manifest to verify
		if _, err := extensions.LoadManifest(manifestPath); err == nil {
			return PackageTypeExtension, manifestPath, nil
		}
	}

	// Check for agenticmodule.yaml (module manifest - primary)
	agenticModulePath := filepath.Join(dir, "agenticmodule.yaml")
	if _, err := os.Stat(agenticModulePath); err == nil {
		// Try to load as module manifest to verify
		if _, err := extensions.LoadAgenticModuleManifest(agenticModulePath); err == nil {
			return PackageTypeAgenticModule, agenticModulePath, nil
		}
	}

	// Fallback: Check for extension.yaml (packaging metadata only, but indicates extension type)
	// This is a weak signal - we still need manifest.yaml to actually load
	extPath := filepath.Join(dir, "extension.yaml")
	if _, err := os.Stat(extPath); err == nil {
		// If manifest.yaml also exists, it's definitely an extension
		if _, err := os.Stat(manifestPath); err == nil {
			return PackageTypeExtension, manifestPath, nil
		}
		// Otherwise, we can't determine without manifest.yaml
	}

	// Fallback: Check for module.yaml (packaging metadata only, but indicates module type)
	modulePath := filepath.Join(dir, "module.yaml")
	if _, err := os.Stat(modulePath); err == nil {
		// If agenticmodule.yaml also exists, it's definitely a module
		if _, err := os.Stat(agenticModulePath); err == nil {
			return PackageTypeAgenticModule, agenticModulePath, nil
		}
		// Otherwise, we can't determine without agenticmodule.yaml
	}

	return PackageTypeUnknown, "", fmt.Errorf("no valid manifest found in directory (need manifest.yaml for extension or agenticmodule.yaml for module)")
}

// LoadExtensionPackageManifest loads an extension package manifest
func LoadExtensionPackageManifest(dir string) (*ExtensionPackageManifest, error) {
	// Always load manifest.yaml as the primary extension definition
	manifestPath := filepath.Join(dir, "manifest.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no extension manifest found: manifest.yaml is required")
	}

	// Load the extension manifest (the actual definition)
	extManifest, err := extensions.LoadManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load extension manifest: %w", err)
	}

	// Create package manifest from extension manifest
	// ID defaults to name (must be filesystem-safe and match name validation)
	pkgManifest := PackageManifest{
		ID:          extManifest.Name, // Use name as ID if not specified separately
		Name:        extManifest.Name,
		Version:     extManifest.Version,
		Description: extManifest.Description,
	}

	// Optionally load packaging-specific fields from extension.yaml if it exists
	extPath := filepath.Join(dir, "extension.yaml")
	if data, err := os.ReadFile(extPath); err == nil {
		var pkgOnly PackageManifest
		if err := yaml.Unmarshal(data, &pkgOnly); err == nil {
			// Merge packaging fields (extension.yaml contains only packaging metadata)
			if pkgOnly.ID != "" {
				// Validate ID format matches name validation rules
				if !idRegex.MatchString(pkgOnly.ID) {
					return nil, fmt.Errorf("invalid id format: '%s'. ID must contain only alphanumeric characters, underscores, and hyphens", pkgOnly.ID)
				}
				pkgManifest.ID = pkgOnly.ID
			}
			if pkgOnly.EnabledDefault != nil {
				pkgManifest.EnabledDefault = pkgOnly.EnabledDefault
			}
			if pkgOnly.HotReload != nil {
				pkgManifest.HotReload = pkgOnly.HotReload
			}
			if pkgOnly.LoadMode != "" {
				pkgManifest.LoadMode = pkgOnly.LoadMode
			}
			if pkgOnly.Autostart != nil {
				pkgManifest.Autostart = pkgOnly.Autostart
			}
		}
	}

	// Set defaults
	if pkgManifest.EnabledDefault == nil {
		enabled := true
		pkgManifest.EnabledDefault = &enabled
	}
	if pkgManifest.HotReload == nil {
		hotReload := true
		pkgManifest.HotReload = &hotReload
	}

	return &ExtensionPackageManifest{
		PackageManifest:  pkgManifest,
		ExtensionManifest: extManifest,
	}, nil
}

// LoadModulePackageManifest loads a module package manifest
func LoadModulePackageManifest(dir string) (*ModulePackageManifest, error) {
	// Always load agenticmodule.yaml as the primary module definition
	manifestPath := filepath.Join(dir, "agenticmodule.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no module manifest found: agenticmodule.yaml is required")
	}

	// Load the module manifest (the actual definition)
	moduleManifest, err := extensions.LoadAgenticModuleManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load module manifest: %w", err)
	}

	// Create package manifest from module manifest
	// ID defaults to name (must be filesystem-safe and match name validation)
	pkgManifest := PackageManifest{
		ID:          moduleManifest.Name, // Use name as ID if not specified separately
		Name:        moduleManifest.Name,
		Version:     moduleManifest.Version,
		Description: moduleManifest.Description,
	}

	// Optionally load packaging-specific fields from module.yaml if it exists
	modulePath := filepath.Join(dir, "module.yaml")
	if data, err := os.ReadFile(modulePath); err == nil {
		var pkgOnly PackageManifest
		if err := yaml.Unmarshal(data, &pkgOnly); err == nil {
			// Merge packaging fields (module.yaml contains only packaging metadata)
			if pkgOnly.ID != "" {
				// Validate ID format matches name validation rules
				if !idRegex.MatchString(pkgOnly.ID) {
					return nil, fmt.Errorf("invalid id format: '%s'. ID must contain only alphanumeric characters, underscores, and hyphens", pkgOnly.ID)
				}
				pkgManifest.ID = pkgOnly.ID
			}
			if pkgOnly.EntryWorkflow != "" {
				pkgManifest.EntryWorkflow = pkgOnly.EntryWorkflow
			}
			if pkgOnly.EnabledDefault != nil {
				pkgManifest.EnabledDefault = pkgOnly.EnabledDefault
			}
			if pkgOnly.HotReload != nil {
				pkgManifest.HotReload = pkgOnly.HotReload
			}
		}
	}

	// Set defaults
	if pkgManifest.EnabledDefault == nil {
		enabled := true
		pkgManifest.EnabledDefault = &enabled
	}
	if pkgManifest.HotReload == nil {
		hotReload := true
		pkgManifest.HotReload = &hotReload
	}

	return &ModulePackageManifest{
		PackageManifest: pkgManifest,
		ModuleManifest:  moduleManifest,
	}, nil
}

// ValidatePackageManifest validates a package manifest
func ValidatePackageManifest(pkg *PackageManifest, pkgType PackageType) error {
	if pkg.ID == "" {
		return fmt.Errorf("validation error: 'id' field is required")
	}

	// Validate ID format (must be filesystem-safe, same as name validation)
	if !idRegex.MatchString(pkg.ID) {
		return fmt.Errorf("validation error: 'id' field '%s' contains invalid characters. Only alphanumeric characters, underscores, and hyphens are allowed", pkg.ID)
	}

	if pkg.Name == "" {
		return fmt.Errorf("validation error: 'name' field is required")
	}

	if pkg.Version == "" {
		return fmt.Errorf("validation error: 'version' field is required")
	}

	if pkgType == PackageTypeExtension {
		if pkg.LoadMode != "" && pkg.LoadMode != "eager" && pkg.LoadMode != "lazy" {
			return fmt.Errorf("validation error: 'load_mode' must be 'eager' or 'lazy'")
		}
	}

	return nil
}

// PackageType represents the type of package
type PackageType string

const (
	PackageTypeExtension     PackageType = "extension"
	PackageTypeAgenticModule PackageType = "agentic-module"
	PackageTypeUnknown       PackageType = "unknown"
)
