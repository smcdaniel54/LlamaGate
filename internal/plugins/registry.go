package plugins

import (
	"fmt"
	"sync"
)

// Registry manages plugin registration and lookup
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
	}
}

// Register registers a plugin with the registry
func (r *Registry) Register(plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	metadata := plugin.Metadata()
	if metadata.Name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	if _, exists := r.plugins[metadata.Name]; exists {
		return fmt.Errorf("plugin %s is already registered", metadata.Name)
	}

	r.plugins[metadata.Name] = plugin
	return nil
}

// Get retrieves a plugin by name
func (r *Registry) Get(name string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return plugin, nil
}

// List returns all registered plugins
func (r *Registry) List() []PluginMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadatas := make([]PluginMetadata, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		metadatas = append(metadatas, plugin.Metadata())
	}

	return metadatas
}

// Unregister removes a plugin from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	delete(r.plugins, name)
	return nil
}
