package extensions

import (
	"context"
	"fmt"
	"time"
)

// ExecutionContext carries execution metadata for extension-to-extension calls
type ExecutionContext struct {
	context.Context
	CallDepth    int           // Current call depth (0 = top level)
	MaxDepth     int           // Maximum allowed depth (default: 10)
	CallBudget   int           // Remaining call budget (default: 100)
	StartTime    time.Time     // Execution start time
	MaxRuntime   time.Duration // Maximum total runtime (default: 5 minutes)
	TraceID      string        // Correlation/trace ID
	ManifestPath string        // Path to YAML manifest for error reporting
}

// NewExecutionContext creates a new execution context with guardrails
func NewExecutionContext(ctx context.Context, traceID, manifestPath string) *ExecutionContext {
	return &ExecutionContext{
		Context:      ctx,
		CallDepth:    0,
		MaxDepth:     10,
		CallBudget:   100,
		StartTime:    time.Now(),
		MaxRuntime:   5 * time.Minute,
		TraceID:      traceID,
		ManifestPath: manifestPath,
	}
}

// WithChild creates a child context for nested extension calls
func (ec *ExecutionContext) WithChild(manifestPath string) (*ExecutionContext, error) {
	// Check depth limit
	if ec.CallDepth >= ec.MaxDepth {
		return nil, &ExecutionError{
			Extension:    "",
			Step:         "extension_call",
			ManifestPath: manifestPath,
			Message:      "maximum call depth exceeded",
			Details:      "recursive extension calls exceeded maximum depth",
		}
	}

	// Check call budget
	if ec.CallBudget <= 0 {
		return nil, &ExecutionError{
			Extension:    "",
			Step:         "extension_call",
			ManifestPath: manifestPath,
			Message:      "call budget exhausted",
			Details:      "maximum number of extension calls reached",
		}
	}

	// Check runtime limit
	if time.Since(ec.StartTime) > ec.MaxRuntime {
		return nil, &ExecutionError{
			Extension:    "",
			Step:         "extension_call",
			ManifestPath: manifestPath,
			Message:      "maximum runtime exceeded",
			Details:      "total execution time exceeded limit",
		}
	}

	return &ExecutionContext{
		Context:      ec.Context,
		CallDepth:    ec.CallDepth + 1,
		MaxDepth:     ec.MaxDepth,
		CallBudget:   ec.CallBudget - 1,
		StartTime:    ec.StartTime,
		MaxRuntime:   ec.MaxRuntime,
		TraceID:      ec.TraceID,
		ManifestPath: manifestPath,
	}, nil
}

// ExecutionError provides actionable error information
type ExecutionError struct {
	Extension    string
	Step         string
	ManifestPath string
	Message      string
	Details      string
}

func (e *ExecutionError) Error() string {
	if e.ManifestPath != "" {
		return fmt.Sprintf("extension %s (step: %s, manifest: %s): %s - %s", e.Extension, e.Step, e.ManifestPath, e.Message, e.Details)
	}
	return fmt.Sprintf("extension %s (step: %s): %s - %s", e.Extension, e.Step, e.Message, e.Details)
}
