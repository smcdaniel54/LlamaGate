// Package mcpclient provides MCP (Model Context Protocol) client functionality.
package mcpclient

import "fmt"

// MCP-specific errors
var (
	ErrNotInitialized      = fmt.Errorf("MCP client not initialized")
	ErrConnectionClosed    = fmt.Errorf("MCP connection closed")
	ErrToolNotFound        = fmt.Errorf("tool not found")
	ErrToolExecutionFailed = fmt.Errorf("tool execution failed")
	ErrTimeout             = fmt.Errorf("operation timed out")
	ErrInvalidTransport    = fmt.Errorf("invalid transport type")
	ErrClientNotFound      = fmt.Errorf("client not found")
)

// ConnectionError represents a connection-related error
type ConnectionError struct {
	ServerName string
	Err        error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error for server %s: %v", e.ServerName, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// ToolExecutionError represents a tool execution error
type ToolExecutionError struct {
	ToolName string
	Err      error
}

func (e *ToolExecutionError) Error() string {
	return fmt.Sprintf("tool execution error for %s: %v", e.ToolName, e.Err)
}

func (e *ToolExecutionError) Unwrap() error {
	return e.Err
}
