package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestInit_DebugMode(t *testing.T) {
	Init(true, "")

	level := zerolog.GlobalLevel()
	assert.Equal(t, zerolog.DebugLevel, level)
}

func TestInit_InfoMode(t *testing.T) {
	Init(false, "")

	level := zerolog.GlobalLevel()
	assert.Equal(t, zerolog.InfoLevel, level)
}

func TestInit_WithLogFile(t *testing.T) {
	// Create a temporary log file
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	Init(false, logFile)
	defer Close()

	// Verify file was created
	_, err := os.Stat(logFile)
	assert.NoError(t, err)

	// Verify logger is initialized
	logger := Get()
	assert.NotNil(t, logger)
}

func TestInit_WithInvalidLogFile(t *testing.T) {
	// Try to write to an invalid path (directory that doesn't exist)
	invalidPath := "/nonexistent/directory/test.log"

	// Should not panic, just log warning
	Init(false, invalidPath)
	defer Close()

	// Logger should still be initialized (fallback to stdout)
	logger := Get()
	assert.NotNil(t, logger)
}

func TestClose(t *testing.T) {
	// Create a temporary log file
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	Init(false, logFile)

	// Close should be safe to call multiple times
	Close()
	Close()
	Close()

	// Logger should still work (writes to stdout)
	logger := Get()
	assert.NotNil(t, logger)
}

func TestClose_NoLogFile(t *testing.T) {
	Init(false, "")

	// Should be safe to close even when no log file was opened
	Close()

	logger := Get()
	assert.NotNil(t, logger)
}

func TestGet(t *testing.T) {
	Init(false, "")
	defer Close()

	logger := Get()
	assert.NotNil(t, logger)

	// Verify it returns a valid logger
	assert.IsType(t, zerolog.Logger{}, logger)
}

func TestInit_ReinitWithNewLogFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize with first log file
	logFile1 := filepath.Join(tmpDir, "test1.log")
	Init(false, logFile1)

	// Verify first file exists
	_, err := os.Stat(logFile1)
	assert.NoError(t, err)

	// Reinitialize with second log file
	logFile2 := filepath.Join(tmpDir, "test2.log")
	Init(false, logFile2)
	defer Close()

	// Verify second file exists
	_, err = os.Stat(logFile2)
	assert.NoError(t, err)

	// First file should still exist (not deleted, just closed)
	_, err = os.Stat(logFile1)
	assert.NoError(t, err)
}

func TestInit_ReinitWithoutLogFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Initialize with log file
	Init(false, logFile)

	// Reinitialize without log file
	Init(false, "")
	defer Close()

	// Logger should still work
	logger := Get()
	assert.NotNil(t, logger)
}
