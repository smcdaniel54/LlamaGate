package setup

import (
	"fmt"

	"github.com/llamagate/llamagate/internal/plugins"
	testplugins "github.com/llamagate/llamagate/tests/plugins"
	"github.com/rs/zerolog/log"
)

// RegisterTestPlugins registers test plugins for use case testing
// This is only called when ENABLE_TEST_PLUGINS=true
func RegisterTestPlugins(registry *plugins.Registry) error {
	log.Info().Msg("Registering test plugins for use case testing")

	testPlugins := testplugins.CreateTestPlugins()
	for _, plugin := range testPlugins {
		if err := registry.Register(plugin); err != nil {
			return fmt.Errorf("failed to register test plugin %s: %w", plugin.Metadata().Name, err)
		}
		log.Info().Str("plugin", plugin.Metadata().Name).Msg("Registered test plugin")
	}

	log.Info().Int("count", len(testPlugins)).Msg("All test plugins registered successfully")
	return nil
}
