package logger

import (
	"io"
	"os"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	logFileHandle *os.File
	logFileMutex  sync.Mutex
	loggerClosed  bool // Track if logger has been closed
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
	// Use 0600 permissions (owner read/write only) for security
	if logFile != "" {
		// Close any previously opened log file to prevent file descriptor leak
		logFileMutex.Lock()
		if logFileHandle != nil {
			if err := logFileHandle.Close(); err != nil {
				log.Warn().Err(err).Msg("Failed to close log file handle")
			}
			logFileHandle = nil
		}
		logFileMutex.Unlock()

		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			log.Warn().Err(err).Str("log_file", logFile).Msg("Failed to open log file, logging to stdout only")
		} else {
			logFileMutex.Lock()
			logFileHandle = file
			loggerClosed = false // Reset closed flag when opening new file
			logFileMutex.Unlock()
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

// Close closes the log file handle if one was opened and updates the logger
// to only write to stdout. Safe to call multiple times.
func Close() {
	logFileMutex.Lock()
	defer logFileMutex.Unlock()
	if logFileHandle != nil {
		if err := logFileHandle.Close(); err != nil {
			log.Warn().Err(err).Msg("Failed to close log file handle")
		}
		logFileHandle = nil
	}
	// Update the global logger to only write to stdout after file is closed
	// This prevents writes to a closed file handle
	// Preserve the current log level when recreating the logger
	if !loggerClosed {
		loggerClosed = true
		currentLevel := zerolog.GlobalLevel()
		// Create new logger with stdout output
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
		// Set the global level to preserve log filtering
		// Using .Level() on the logger instance doesn't affect global filtering
		zerolog.SetGlobalLevel(currentLevel)
	}
}

// Get returns the global logger instance
func Get() zerolog.Logger {
	return log.Logger
}
