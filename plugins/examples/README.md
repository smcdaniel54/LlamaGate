# Plugin Examples

This directory contains example plugins that demonstrate various use cases and patterns.

## Examples

### Dynamic Config Example (`dynamic_config_example.go`)

Demonstrates dynamic configuration patterns for adaptable plugins.

**Features:**
- Environment-aware configuration (development/staging/production)
- Adaptive timeouts based on configuration
- Dynamic workflow building
- Configurable retry logic
- Cache integration

**Usage:**

```go
import "github.com/llamagate/llamagate/plugins/examples"

plugin := examples.NewDynamicConfigPlugin()
plugin.SetWorkflowExecutor(executor)

result, err := plugin.Execute(ctx, map[string]interface{}{
    "query": "Process this query",
    "environment": "production",
    "max_depth": 5,
    "use_cache": true,
    "model": "llama3.2",
})
```

**Configuration Options:**
- `environment`: "development", "staging", or "production" (affects timeouts and retries)
- `max_depth`: Processing depth (1-10, affects workflow steps and timeout)
- `use_cache`: Enable/disable caching
- `model`: LLM model to use

**See Also:** [Dynamic Config Use Cases](../../docs/DYNAMIC_CONFIG_USECASES.md)

### Text Summarizer (`text_summarizer.go`)

A practical example plugin that summarizes text content.

**Features:**
- Input validation with type checking
- Multi-step workflow (preprocess → extract → format)
- Configurable parameters (max_length, style)
- Structured output with metrics

**Usage:**

```go
import "github.com/llamagate/llamagate/plugins/examples"

plugin := examples.NewTextSummarizerPlugin()
result, err := plugin.Execute(ctx, map[string]interface{}{
    "text": "Long text to summarize...",
    "max_length": 200,
    "style": "brief",
})
```

**Input Parameters:**
- `text` (required): The text content to summarize
- `max_length` (optional, default: 200): Maximum length of summary
- `style` (optional, default: "brief"): Summary style ("brief", "detailed", "bullet")

**Output:**
- `summary`: The generated summary
- `original_length`: Length of original text
- `summary_length`: Length of generated summary
- `compression_ratio`: Ratio of summary to original length

### Workflow Example (`workflow_example.go`)

Demonstrates an agentic workflow with LLM and tool integration.

**Features:**
- Multi-step agentic workflow
- LLM integration
- Tool calling
- Data transformation
- Step dependencies
- Error handling strategies

**Usage:**

```go
import "github.com/llamagate/llamagate/plugins/examples"

// Create workflow executor with handlers
executor := plugins.NewWorkflowExecutor(llmHandler, toolHandler)

// Create plugin
plugin := examples.NewExampleWorkflowPlugin()
plugin.SetWorkflowExecutor(executor)

// Execute
result, err := plugin.Execute(ctx, map[string]interface{}{
    "query": "What is in the file /path/to/file.txt?",
    "model": "llama3.2",
})
```

**Workflow Steps:**
1. **Analyze Query**: LLM analyzes the user query
2. **Extract Information**: Extract key information from analysis
3. **Execute Tool**: Call MCP tools based on extracted info
4. **Synthesize Result**: LLM synthesizes final result

**Input Parameters:**
- `query` (required): The user query to process
- `model` (optional, default: "llama3.2"): LLM model to use

**Output:**
- `final_result`: The final processed result
- `workflow_steps`: Results from each workflow step

## Learning from Examples

These examples demonstrate:

1. **Simple Plugins**: Text summarizer shows a straightforward plugin with clear input/output
2. **Complex Workflows**: Workflow example shows multi-step agentic processing
3. **Best Practices**: Both examples follow plugin system best practices

## Next Steps

1. Review the examples to understand patterns
2. Use the template (`../templates/plugin_template.go`) to create your own
3. See [../README.md](../README.md) and [../../docs/PLUGINS.md](../../docs/PLUGINS.md) for more information
