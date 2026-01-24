package extensions

import (
	"testing"

	"github.com/stretchr/testify/require"
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

	require.NoError(t, registry.Register(manifest))

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
		require.NoError(t, registry.Register(m))
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

	require.NoError(t, registry.Register(enabled))
	require.NoError(t, registry.Register(disabled))

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

	require.NoError(t, registry.Register(manifest))

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

func TestRegistry_SetEnabled_BuiltinCannotBeDisabled(t *testing.T) {
	registry := NewRegistry()

	builtin := &Manifest{
		Name:        "builtin-ext",
		Version:     "1.0.0",
		Description: "Builtin extension",
		Builtin:     true,
		Enabled:     boolPtr(true),
	}

	require.NoError(t, registry.Register(builtin))

	// Verify it's enabled
	if !registry.IsEnabled("builtin-ext") {
		t.Error("Expected builtin extension to be enabled")
	}

	// Try to disable it - should fail
	err := registry.SetEnabled("builtin-ext", false)
	if err == nil {
		t.Error("Expected error when trying to disable builtin extension")
	}
	if err != nil && err.Error() != "cannot disable builtin extension builtin-ext" {
		t.Errorf("Expected specific error message, got: %v", err)
	}

	// Verify it's still enabled
	if !registry.IsEnabled("builtin-ext") {
		t.Error("Expected builtin extension to still be enabled after failed disable attempt")
	}

	// Can still enable it (no-op but should work)
	err = registry.SetEnabled("builtin-ext", true)
	if err != nil {
		t.Errorf("Should be able to enable builtin extension: %v", err)
	}
}

func TestRegistry_Unregister_BuiltinCannotBeUnregistered(t *testing.T) {
	registry := NewRegistry()

	builtin := &Manifest{
		Name:        "builtin-ext",
		Version:     "1.0.0",
		Description: "Builtin extension",
		Builtin:     true,
	}

	require.NoError(t, registry.Register(builtin))

	// Try to unregister it - should fail
	err := registry.Unregister("builtin-ext")
	if err == nil {
		t.Error("Expected error when trying to unregister builtin extension")
	}
	if err != nil && err.Error() != "cannot unregister builtin extension builtin-ext" {
		t.Errorf("Expected specific error message, got: %v", err)
	}

	// Verify it's still registered
	_, err = registry.Get("builtin-ext")
	if err != nil {
		t.Error("Expected builtin extension to still be registered after failed unregister attempt")
	}
}

func TestRegistry_Register_BuiltinAlwaysEnabled(t *testing.T) {
	registry := NewRegistry()

	// Test that builtin extension is always enabled, even if manifest says disabled
	builtin := &Manifest{
		Name:        "builtin-ext",
		Version:     "1.0.0",
		Description: "Builtin extension",
		Builtin:     true,
		Enabled:     boolPtr(false), // Try to set as disabled in manifest
	}

	require.NoError(t, registry.Register(builtin))

	// Should be enabled despite manifest saying disabled
	if !registry.IsEnabled("builtin-ext") {
		t.Error("Expected builtin extension to be enabled even if manifest says disabled")
	}
}

func TestRegistry_RegisterOrUpdate_BuiltinAlwaysEnabled(t *testing.T) {
	registry := NewRegistry()

	// Test RegisterOrUpdate with builtin extension
	builtin := &Manifest{
		Name:        "builtin-ext",
		Version:     "1.0.0",
		Description: "Builtin extension",
		Builtin:     true,
		Enabled:     boolPtr(false), // Try to set as disabled
	}

	require.NoError(t, registry.RegisterOrUpdate(builtin))

	// Should be enabled despite manifest saying disabled
	if !registry.IsEnabled("builtin-ext") {
		t.Error("Expected builtin extension to be enabled via RegisterOrUpdate even if manifest says disabled")
	}

	// Update the extension - should still be enabled
	builtin.Version = "2.0.0"
	require.NoError(t, registry.RegisterOrUpdate(builtin))
	if !registry.IsEnabled("builtin-ext") {
		t.Error("Expected builtin extension to remain enabled after update")
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

	require.NoError(t, registry.Register(workflow))
	require.NoError(t, registry.Register(middleware))

	workflows := registry.GetByType("workflow")
	if len(workflows) != 1 {
		t.Errorf("Expected 1 workflow extension, got %d", len(workflows))
	}

	middlewares := registry.GetByType("middleware")
	if len(middlewares) != 1 {
		t.Errorf("Expected 1 middleware extension, got %d", len(middlewares))
	}
}
