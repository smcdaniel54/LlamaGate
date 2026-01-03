// Package tools provides tool management, mapping, and guardrails for MCP tools.
package tools

// Tool represents a namespaced tool definition
type Tool struct {
	// NamespacedName is the full name including namespace (e.g., "mcp.filesystem.read_file")
	NamespacedName string
	// ServerName is the MCP server name
	ServerName string
	// OriginalName is the original tool name from the MCP server
	OriginalName string
	// Description is the tool description
	Description string
	// InputSchema is the JSON Schema for the tool input
	InputSchema map[string]interface{}
}

// ToolCall represents a tool call request
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"` // Namespaced name
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolCallID string
	Content    string
	IsError    bool
}
