package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the global zerolog logger with appropriate settings
func Init(debug bool, logFile string) {
	// Set global log level
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Configure output format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Set up output writers
	var writers []io.Writer

	// Always write to stdout
	writers = append(writers, os.Stdout)

	// If log file is specified, also write to file
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Warn().Err(err).Str("log_file", logFile).Msg("Failed to open log file, logging to stdout only")
		} else {
			writers = append(writers, file)
		}
	}

	// Create multi-writer if we have multiple outputs
	var output io.Writer
	if len(writers) > 1 {
		output = io.MultiWriter(writers...)
	} else {
		output = writers[0]
	}

	log.Logger = zerolog.New(output).With().Timestamp().Logger()
}

// Get returns the global logger instance
func Get() zerolog.Logger {
	return log.Logger
}
