package extensions

import (
	"testing"
)

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	manifest := &Manifest{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test extension",
	}

	err := registry.Register(manifest)
	if err != nil {
		t.Fatalf("Failed to register extension: %v", err)
	}

	// Try to register duplicate
	err = registry.Register(manifest)
	if err == nil {
		t.Error("Expected error when registering duplicate extension")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()

	manifest := &Manifest{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test extension",
	}

	registry.Register(manifest)

	retrieved, err := registry.Get("test")
	if err != nil {
		t.Fatalf("Failed to get extension: %v", err)
	}

	if retrieved.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", retrieved.Name)
	}

	// Try to get non-existent extension
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent extension")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	manifests := []*Manifest{
		{Name: "ext1", Version: "1.0.0", Description: "First"},
		{Name: "ext2", Version: "2.0.0", Description: "Second"},
	}

	for _, m := range manifests {
		registry.Register(m)
	}

	list := registry.List()
	if len(list) != 2 {
		t.Errorf("Expected 2 extensions, got %d", len(list))
	}
}

func TestRegistry_IsEnabled(t *testing.T) {
	registry := NewRegistry()

	enabled := &Manifest{
		Name:        "enabled",
		Version:     "1.0.0",
		Description: "Enabled extension",
		Enabled:     boolPtr(true),
	}

	disabled := &Manifest{
		Name:        "disabled",
		Version:     "1.0.0",
		Description: "Disabled extension",
		Enabled:     boolPtr(false),
	}

	registry.Register(enabled)
	registry.Register(disabled)

	if !registry.IsEnabled("enabled") {
		t.Error("Expected 'enabled' extension to be enabled")
	}
	if registry.IsEnabled("disabled") {
		t.Error("Expected 'disabled' extension to be disabled")
	}
}

func TestRegistry_SetEnabled(t *testing.T) {
	registry := NewRegistry()

	manifest := &Manifest{
		Name:        "test",
		Version:     "1.0.0",
		Description: "Test extension",
		Enabled:     boolPtr(true),
	}

	registry.Register(manifest)

	// Disable it
	err := registry.SetEnabled("test", false)
	if err != nil {
		t.Fatalf("Failed to disable extension: %v", err)
	}

	if registry.IsEnabled("test") {
		t.Error("Expected extension to be disabled")
	}

	// Try to set enabled for non-existent extension
	err = registry.SetEnabled("nonexistent", true)
	if err == nil {
		t.Error("Expected error when setting enabled for non-existent extension")
	}
}

func TestRegistry_GetByType(t *testing.T) {
	registry := NewRegistry()

	workflow := &Manifest{
		Name:        "workflow-ext",
		Version:     "1.0.0",
		Description: "Workflow extension",
		Type:        "workflow",
		Enabled:     boolPtr(true),
	}

	middleware := &Manifest{
		Name:        "middleware-ext",
		Version:     "1.0.0",
		Description: "Middleware extension",
		Type:        "middleware",
		Enabled:     boolPtr(true),
	}

	registry.Register(workflow)
	registry.Register(middleware)

	workflows := registry.GetByType("workflow")
	if len(workflows) != 1 {
		t.Errorf("Expected 1 workflow extension, got %d", len(workflows))
	}

	middlewares := registry.GetByType("middleware")
	if len(middlewares) != 1 {
		t.Errorf("Expected 1 middleware extension, got %d", len(middlewares))
	}
}
