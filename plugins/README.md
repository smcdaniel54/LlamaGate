# LlamaGate Plugins

This directory contains the plugin system for LlamaGate, including templates, examples, and documentation.

## Directory Structure

```
plugins/
├── templates/          # Plugin templates for creating new plugins
│   └── plugin_template.go
├── examples/           # Example plugins demonstrating various use cases
│   ├── text_summarizer.go
│   └── workflow_example.go
└── README.md          # This file
```

## Quick Start

### 1. Create a New Plugin

Copy the template and customize it:

```bash
cp plugins/templates/plugin_template.go plugins/my_plugin.go
```

Then edit `my_plugin.go` to implement your plugin logic.

### 2. Use an Example

Review the examples to understand plugin patterns:

- **Text Summarizer** (`examples/text_summarizer.go`): Simple plugin with multi-step workflow
- **Workflow Example** (`examples/workflow_example.go`): Complex agentic workflow with LLM and tool integration

### 3. Register Your Plugin

In your application code:

```go
import (
    "github.com/llamagate/llamagate/internal/plugins"
    "github.com/llamagate/llamagate/plugins/examples"
)

func main() {
    registry := plugins.NewRegistry()
    
    // Register example plugins
    registry.Register(examples.NewTextSummarizerPlugin())
    
    // Register your custom plugin
    registry.Register(NewMyPlugin())
}
```

## Templates

### Simple Plugin Template (Recommended for Beginners)

The `simple_plugin.go` provides the minimal code needed:

- ✅ Only 3 methods required
- ✅ ~50 lines of code
- ✅ Perfect for learning
- ✅ Easy to understand

**Start here if you're new to plugins!**

### Full Plugin Template

The `plugin_template.go` provides a complete starting point with:

- ✅ Metadata structure
- ✅ Input/output schema definitions
- ✅ Validation skeleton
- ✅ Execution framework
- ✅ Error handling
- ✅ Result formatting

**Key Sections to Customize:**

1. **Metadata**: Update name, version, description, author
2. **InputSchema**: Define your input parameters (JSON Schema)
3. **OutputSchema**: Define your output structure (JSON Schema)
4. **RequiredInputs**: List required parameter names
5. **OptionalInputs**: Define optional parameters with defaults
6. **ValidateInput**: Add your validation logic
7. **Execute**: Implement your workflow

## Examples

### Text Summarizer

A practical example that:

- Takes text input
- Summarizes to specified length
- Supports multiple styles (brief, detailed, bullet)
- Returns metrics (compression ratio, lengths)

**Key Features:**
- Input validation
- Multi-step processing
- Configurable parameters
- Structured output

### Workflow Example

Demonstrates agentic workflows with:

- LLM integration
- Tool calling
- Data transformation
- Step dependencies
- Error handling

**Workflow Steps:**
1. Analyze query (LLM)
2. Extract information (Transform)
3. Execute tool (MCP tool)
4. Synthesize result (LLM)

## Documentation

See [docs/PLUGINS.md](../docs/PLUGINS.md) for comprehensive documentation including:

- Plugin architecture
- Creating plugins
- Agentic workflows
- Best practices
- API reference

## Plugin System Architecture

The plugin system consists of:

1. **Types** (`internal/plugins/types.go`): Core types and interfaces
2. **Registry** (`internal/plugins/registry.go`): Plugin registration and lookup
3. **Workflow Executor** (`internal/plugins/workflow.go`): Executes agentic workflows

## Contributing

When creating new plugins:

1. Follow the template structure
2. Include comprehensive validation
3. Document input/output schemas
4. Add error handling
5. Include usage examples
6. Write tests

## License

Same as LlamaGate project (MIT License)
