# Plugin System Quick Start

Get started with LlamaGate plugins in 5 minutes!

## What are Plugins?

Plugins are reusable components that:
- Process inputs through defined workflows
- Integrate with LLM models and MCP tools
- Return structured outputs
- Enable agentic (AI-driven) workflows

## Quick Start

### 1. Review the Simple Template (Recommended!)

Start with the simple template - it's only ~50 lines:

```bash
cat plugins/templates/simple_plugin.go
```

Or use the full template for more features:

```bash
cat plugins/templates/plugin_template.go
```

### 2. Try an Example

See a working example:

```bash
cat plugins/examples/text_summarizer.go
```

### 3. Create Your First Plugin

Copy the template:

```bash
cp plugins/templates/plugin_template.go plugins/my_first_plugin.go
```

Edit `my_first_plugin.go`:

1. **Update Metadata**:
   ```go
   Name:        "my_first_plugin",
   Description: "What my plugin does",
   ```

2. **Define Input Schema**:
   ```go
   InputSchema: map[string]interface{}{
       "type": "object",
       "properties": map[string]interface{}{
           "input_param": map[string]interface{}{
               "type": "string",
               "description": "Your input parameter",
           },
       },
       "required": []string{"input_param"},
   },
   ```

3. **Implement Execute**:
   ```go
   func (p *MyFirstPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
       // Your logic here
       return &plugins.PluginResult{
           Success: true,
           Data: map[string]interface{}{
               "result": "your result",
           },
       }, nil
   }
   ```

### 4. Register Your Plugin

In your application:

```go
import (
    "github.com/llamagate/llamagate/internal/plugins"
    myplugin "github.com/llamagate/llamagate/plugins"
)

func main() {
    registry := plugins.NewRegistry()
    registry.Register(myplugin.NewMyFirstPlugin())
}
```

## Example: Text Summarizer

Here's a complete example:

```go
import "github.com/llamagate/llamagate/plugins/examples"

// Create plugin
plugin := examples.NewTextSummarizerPlugin()

// Execute
result, err := plugin.Execute(ctx, map[string]interface{}{
    "text": "Long text to summarize...",
    "max_length": 200,
    "style": "brief",
})

if result.Success {
    fmt.Println("Summary:", result.Data["summary"])
}
```

## Agentic Workflows

Create multi-step workflows that use LLMs and tools:

```go
workflow := &plugins.Workflow{
    Steps: []plugins.WorkflowStep{
        {
            Type: "llm_call",
            Config: map[string]interface{}{
                "model": "llama3.2",
                "prompt": "Analyze this query...",
            },
        },
        {
            Type: "tool_call",
            Config: map[string]interface{}{
                "tool_name": "mcp.filesystem.read_file",
                "arguments": map[string]interface{}{
                    "path": "/path/to/file",
                },
            },
        },
    },
}
```

## Next Steps

- 📖 Read [Full Documentation](PLUGINS.md)
- 🔍 Explore [Examples](plugins/examples/)
- 📝 Use [Template](plugins/templates/)
- 🚀 Build Your Plugin!

## Resources

- **Full Guide**: [docs/PLUGINS.md](PLUGINS.md)
- **Examples**: [plugins/examples/](plugins/examples/)
- **Template**: [plugins/templates/](plugins/templates/)
- **API Reference**: [internal/plugins/](internal/plugins/)

---

Happy plugin development! 🎉
