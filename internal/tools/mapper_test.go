package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTool_ToOpenAIFormat(t *testing.T) {
	tool := &Tool{
		NamespacedName: "mcp.server.read_file",
		ServerName:     "server",
		OriginalName:   "read_file",
		Description:    "Read a file",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	openAITool := tool.ToOpenAIFormat()

	assert.Equal(t, "mcp.server.read_file", openAITool.Name)
	assert.Equal(t, "Read a file", openAITool.Description)
	assert.NotNil(t, openAITool.Parameters)
	assert.Equal(t, "object", openAITool.Parameters["type"])
}

func TestToolsToOpenAIFormat(t *testing.T) {
	tools := []*Tool{
		{
			NamespacedName: "mcp.server.tool1",
			Description:    "Tool 1",
			InputSchema:    map[string]interface{}{"type": "object"},
		},
		{
			NamespacedName: "mcp.server.tool2",
			Description:    "Tool 2",
			InputSchema:    map[string]interface{}{"type": "object"},
		},
	}

	openAITools := ToolsToOpenAIFormat(tools)

	assert.Len(t, openAITools, 2)
	assert.Equal(t, "mcp.server.tool1", openAITools[0].Name)
	assert.Equal(t, "mcp.server.tool2", openAITools[1].Name)
}

