// Package tools provides tool management, mapping, and guardrails for MCP tools.
package tools

// OpenAIFunction represents an OpenAI function definition
type OpenAIFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
}

// ToOpenAIFormat converts an MCP tool to OpenAI function format
func (t *Tool) ToOpenAIFormat() OpenAIFunction {
	// OpenAI expects parameters to be a JSON Schema object
	// MCP's inputSchema is already a JSON Schema, so we can use it directly
	// However, we need to ensure it's properly structured

	parameters := make(map[string]interface{})
	if t.InputSchema != nil {
		// Copy the schema
		for k, v := range t.InputSchema {
			parameters[k] = v
		}
	} else {
		// Default empty schema
		parameters["type"] = "object"
		parameters["properties"] = make(map[string]interface{})
	}

	return OpenAIFunction{
		Name:        t.NamespacedName,
		Description: t.Description,
		Parameters:  parameters,
	}
}

// ToolsToOpenAIFormat converts multiple tools to OpenAI format
//
//nolint:revive // ToolsToOpenAIFormat is the preferred name for external API
func ToolsToOpenAIFormat(tools []*Tool) []OpenAIFunction {
	result := make([]OpenAIFunction, len(tools))
	for i, tool := range tools {
		result[i] = tool.ToOpenAIFormat()
	}
	return result
}
