package plugins

import (
	"fmt"
	"sync"
)

// Registry manages plugin registration and lookup
type Registry struct {
	mu       sync.RWMutex
	plugins  map[string]Plugin
	contexts map[string]*PluginContext // Plugin-specific contexts
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins:  make(map[string]Plugin),
		contexts: make(map[string]*PluginContext),
	}
}

// Register registers a plugin with the registry
func (r *Registry) Register(plugin Plugin) error {
	return r.RegisterWithContext(plugin, nil)
}

// RegisterWithContext registers a plugin with the registry and optional context
func (r *Registry) RegisterWithContext(plugin Plugin, ctx *PluginContext) error {
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
	if ctx != nil {
		r.contexts[metadata.Name] = ctx
	}

	return nil
}

// GetContext retrieves the context for a plugin
func (r *Registry) GetContext(pluginName string) *PluginContext {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.contexts[pluginName]
}

// SetContext sets the context for a plugin
func (r *Registry) SetContext(pluginName string, ctx *PluginContext) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.contexts[pluginName] = ctx
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
