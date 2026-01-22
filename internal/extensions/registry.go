package extensions

import (
	"fmt"
	"sync"
)

// Registry manages extension registration and lookup
type Registry struct {
	mu         sync.RWMutex
	extensions map[string]*Manifest
	enabled    map[string]bool // Track enabled/disabled state
}

// NewRegistry creates a new extension registry
func NewRegistry() *Registry {
	return &Registry{
		extensions: make(map[string]*Manifest),
		enabled:    make(map[string]bool),
	}
}

// Register registers an extension manifest
func (r *Registry) Register(manifest *Manifest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if manifest.Name == "" {
		return fmt.Errorf("extension name cannot be empty")
	}

	if _, exists := r.extensions[manifest.Name]; exists {
		return fmt.Errorf("extension %s is already registered", manifest.Name)
	}

	r.extensions[manifest.Name] = manifest
	r.enabled[manifest.Name] = manifest.IsEnabled()

	return nil
}

// Get retrieves an extension by name
func (r *Registry) Get(name string) (*Manifest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	manifest, exists := r.extensions[name]
	if !exists {
		return nil, fmt.Errorf("extension %s not found", name)
	}

	return manifest, nil
}

// List returns all registered extensions
func (r *Registry) List() []*Manifest {
	r.mu.RLock()
	defer r.mu.RUnlock()

	manifests := make([]*Manifest, 0, len(r.extensions))
	for _, manifest := range r.extensions {
		manifests = append(manifests, manifest)
	}

	return manifests
}

// IsEnabled checks if an extension is enabled
func (r *Registry) IsEnabled(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.enabled[name]
}

// SetEnabled sets the enabled state of an extension
func (r *Registry) SetEnabled(name string, enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.extensions[name]; !exists {
		return fmt.Errorf("extension %s not found", name)
	}

	r.enabled[name] = enabled
	return nil
}

// GetByType returns all extensions of a specific type
func (r *Registry) GetByType(extType string) []*Manifest {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var manifests []*Manifest
	for _, manifest := range r.extensions {
		if manifest.Type == extType && r.enabled[manifest.Name] {
			manifests = append(manifests, manifest)
		}
	}

	return manifests
}

// Unregister removes an extension from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.extensions[name]; !exists {
		return fmt.Errorf("extension %s not found", name)
	}

	delete(r.extensions, name)
	delete(r.enabled, name)
	return nil
}

// RegisterOrUpdate registers a new extension or updates an existing one
func (r *Registry) RegisterOrUpdate(manifest *Manifest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if manifest.Name == "" {
		return fmt.Errorf("extension name cannot be empty")
	}

	r.extensions[manifest.Name] = manifest
	r.enabled[manifest.Name] = manifest.IsEnabled()

	return nil
}
