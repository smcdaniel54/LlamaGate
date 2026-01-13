package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// parseDurationWithDefault parses a duration from viper with a default value
// Returns an error if the duration format is invalid
func parseDurationWithDefault(key, defaultValue string) (time.Duration, error) {
	durationStr := viper.GetString(key)
	if durationStr == "" {
		durationStr = defaultValue
	}
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, fmt.Errorf("invalid %s format: %w (expected duration like '5m', '30s', '1h')", key, err)
	}
	return duration, nil
}
