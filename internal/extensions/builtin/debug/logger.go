package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/smcdaniel54/LlamaGate/internal/extensions/builtin/core"
)

// Logger provides comprehensive debugging and logging capabilities
type Logger struct {
	name        string
	version     string
	level       string
	registry    *core.Registry
	logFile     *os.File
	structured  bool
	colors      bool
	eventCounts map[string]int
}

// LogLevel represents logging levels
const (
	LogLevelDebug   = "debug"
	LogLevelInfo    = "info"
	LogLevelWarning = "warning"
	LogLevelError   = "error"
)

// NewLogger creates a new debug logger
func NewLogger(name, version string) *Logger {
	return &Logger{
		name:        name,
		version:     version,
		level:       LogLevelInfo,
		registry:    core.GetRegistry(),
		structured:  true,
		colors:      true,
		eventCounts: make(map[string]int),
	}
}

// Name returns the name of the extension
func (l *Logger) Name() string {
	return l.name
}

// Version returns the version of the extension
func (l *Logger) Version() string {
	return l.version
}

// Initialize initializes the logger
func (l *Logger) Initialize(ctx context.Context, config map[string]interface{}) error {
	if level, ok := config["level"].(string); ok {
		l.level = level
	}
	if structured, ok := config["structured"].(bool); ok {
		l.structured = structured
	}
	if colors, ok := config["colors"].(bool); ok {
		l.colors = colors
	}
	if logFile, ok := config["log_file"].(string); ok && logFile != "" {
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		l.logFile = file
	}

	// Subscribe to all events
	publisher, err := l.registry.GetEventPublisher("default")
	if err == nil {
		filter := &core.EventFilter{} // All events
		_, _ = publisher.Subscribe(ctx, filter, l.handleEvent)
	}

	l.Log(LogLevelInfo, "Debug logger initialized", map[string]interface{}{
		"level":      l.level,
		"structured": l.structured,
		"colors":     l.colors,
	})

	return nil
}

// Shutdown shuts down the logger
func (l *Logger) Shutdown(ctx context.Context) error {
	if l.logFile != nil {
		l.logFile.Close()
	}

	// Print event counts summary
	summary := make(map[string]interface{})
	for k, v := range l.eventCounts {
		summary[k] = v
	}
	l.Log(LogLevelInfo, "Event counts summary", summary)
	return nil
}

// Log logs a message at the specified level
func (l *Logger) Log(level, message string, data map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	timestamp := time.Now().Format(time.RFC3339)
	entry := map[string]interface{}{
		"timestamp": timestamp,
		"level":     level,
		"message":   message,
		"source":    l.name,
	}
	if data != nil {
		entry["data"] = data
	}

	if l.structured {
		l.logStructured(entry)
	} else {
		l.logPlain(level, message, data)
	}

	if l.logFile != nil {
		l.logToFile(entry)
	}
}

// shouldLog checks if a log level should be logged
func (l *Logger) shouldLog(level string) bool {
	levels := map[string]int{
		LogLevelDebug:   0,
		LogLevelInfo:    1,
		LogLevelWarning: 2,
		LogLevelError:   3,
	}
	return levels[level] >= levels[l.level]
}

// logStructured logs in structured JSON format
func (l *Logger) logStructured(entry map[string]interface{}) {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}
	fmt.Println(string(jsonData))
}

// logPlain logs in plain text format with colors
func (l *Logger) logPlain(level, message string, data map[string]interface{}) {
	timestamp := time.Now().Format("15:04:05")
	prefix := l.getColorPrefix(level)
	reset := ""
	if l.colors {
		reset = "\033[0m"
	}

	fmt.Printf("%s[%s] %s%s %s", prefix, timestamp, level, reset, message)
	if data != nil && len(data) > 0 {
		fmt.Printf(" %+v", data)
	}
	fmt.Println()
}

// logToFile logs to file
func (l *Logger) logToFile(entry map[string]interface{}) {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return
	}
	l.logFile.WriteString(string(jsonData) + "\n")
}

// getColorPrefix returns color prefix for log level
func (l *Logger) getColorPrefix(level string) string {
	if !l.colors {
		return ""
	}
	colors := map[string]string{
		LogLevelDebug:   "\033[36m", // Cyan
		LogLevelInfo:   "\033[32m",  // Green
		LogLevelWarning: "\033[33m",  // Yellow
		LogLevelError:   "\033[31m",  // Red
	}
	return colors[level]
}

// handleEvent handles events for logging
func (l *Logger) handleEvent(ctx context.Context, event *core.Event) error {
	if event == nil {
		return nil
	}

	// Count events
	l.eventCounts[event.Type]++

	// Log event
	l.Log(LogLevelDebug, fmt.Sprintf("Event: %s", event.Type), map[string]interface{}{
		"source":    event.Source,
		"event_id":  event.ID,
		"event_data": event.Data,
	})

	return nil
}

// Debug logs a debug message
func (l *Logger) Debug(message string, data map[string]interface{}) {
	l.Log(LogLevelDebug, message, data)
}

// Info logs an info message
func (l *Logger) Info(message string, data map[string]interface{}) {
	l.Log(LogLevelInfo, message, data)
}

// Warning logs a warning message
func (l *Logger) Warning(message string, data map[string]interface{}) {
	l.Log(LogLevelWarning, message, data)
}

// Error logs an error message
func (l *Logger) Error(message string, data map[string]interface{}) {
	l.Log(LogLevelError, message, data)
}
