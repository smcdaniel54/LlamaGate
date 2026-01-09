package setup

import (
	"github.com/llamagate/llamagate/internal/plugins"
	"github.com/rs/zerolog/log"
)

// RegisterTestPlugins registers test plugins for use case testing
// This is only called when ENABLE_TEST_PLUGINS=true
func RegisterTestPlugins(registry *plugins.Registry) error {
	// Note: Test plugins are in tests/plugins/test_plugins.go
	// They need to be imported and registered here
	// For now, this is a placeholder that logs the intent
	
	log.Info().Msg("Test plugins registration requested")
	log.Warn().Msg("Test plugins must be imported and registered manually")
	log.Warn().Msg("See tests/plugins/test_plugins.go for test plugin definitions")
	
	// Example of how to register (when testplugins package is available):
	// testPlugins := testplugins.CreateTestPlugins()
	// for _, plugin := range testPlugins {
	//     if err := registry.Register(plugin); err != nil {
	//         return fmt.Errorf("failed to register test plugin: %w", err)
	//     }
	//     log.Info().Str("plugin", plugin.Metadata().Name).Msg("Registered test plugin")
	// }
	
	return nil
}
