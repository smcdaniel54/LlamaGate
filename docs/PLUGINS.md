# LlamaGate Plugin System

The LlamaGate plugin system allows you to create reusable, agentic workflows that integrate with local LLM models and MCP tools. This guide will help you understand how to create, use, and extend plugins.

## Table of Contents

- [Overview](#overview)
- [Plugin Architecture](#plugin-architecture)
- [Creating a Plugin](#creating-a-plugin)
- [Plugin Template](#plugin-template)
- [Example Plugins](#example-plugins)
- [Agentic Workflows](#agentic-workflows)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)

## Overview

Plugins in LlamaGate are self-contained modules that:

- Define clear input/output schemas
- Implement validation logic
- Execute agentic workflows
- Integrate with LLM models and MCP tools
- Return structured results

### Key Concepts

**Plugin**: A reusable component that processes inputs and produces outputs through a defined workflow.

**Workflow**: A sequence of steps that can include:
- LLM calls (chat completions)
- Tool calls (MCP tools)
- Data transformations
- Conditional logic

**Agentic**: The workflow can make decisions, call tools, and iterate based on results.

## Plugin Architecture

### Core Components

1. **Plugin Interface**: Defines the contract all plugins must implement
2. **Plugin Registry**: Manages plugin registration and lookup
3. **Workflow Executor**: Executes multi-step agentic workflows
4. **Type System**: Defines input/output schemas and validation

### Plugin Lifecycle

```
1. Registration → Plugin is registered with the registry
2. Validation → Input parameters are validated
3. Execution → Workflow is executed
4. Result → Structured output is returned
```

## Creating a Plugin

### Step 1: Use the Template

Start by copying the plugin template:

```bash
cp plugins/templates/plugin_template.go plugins/my_plugin.go
```

### Step 2: Define Metadata

Update the `Metadata()` method with your plugin information:

```go
func (p *MyPlugin) Metadata() plugins.PluginMetadata {
    return plugins.PluginMetadata{
        Name:        "my_plugin",
        Version:     "1.0.0",
        Description: "What my plugin does",
        Author:      "Your Name",
        
        InputSchema: map[string]interface{}{
            // Define your input schema (JSON Schema)
        },
        
        OutputSchema: map[string]interface{}{
            // Define your output schema (JSON Schema)
        },
        
        RequiredInputs: []string{"required_param"},
        OptionalInputs: map[string]interface{}{
            "optional_param": "default_value",
        },
    }
}
```

### Step 3: Implement Validation

Add validation logic in `ValidateInput()`:

```go
func (p *MyPlugin) ValidateInput(input map[string]interface{}) error {
    // Check required inputs
    if _, exists := input["required_param"]; !exists {
        return fmt.Errorf("required input 'required_param' is missing")
    }
    
    // Add custom validation
    // ...
    
    return nil
}
```

### Step 4: Implement Execution

Add your workflow logic in `Execute()`:

```go
func (p *MyPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    startTime := time.Now()
    
    // Your workflow logic here
    // ...
    
    return &plugins.PluginResult{
        Success: true,
        Data: map[string]interface{}{
            "result": "your result",
        },
        Metadata: plugins.ExecutionMetadata{
            ExecutionTime: time.Since(startTime),
            StepsExecuted: 1,
            Timestamp: time.Now(),
        },
    }, nil
}
```

### Step 5: Register Your Plugin

Register your plugin in your application:

```go
registry := plugins.NewRegistry()
registry.Register(NewMyPlugin())
```

## Plugin Template

The plugin template (`plugins/templates/plugin_template.go`) provides a complete starting point with:

- ✅ Metadata structure
- ✅ Input/output schema definitions
- ✅ Validation skeleton
- ✅ Execution framework
- ✅ Error handling
- ✅ Result formatting

### Template Structure

```go
type TemplatePlugin struct {
    // Add plugin-specific fields
}

func (p *TemplatePlugin) Metadata() plugins.PluginMetadata {
    // Define metadata, schemas, required/optional inputs
}

func (p *TemplatePlugin) ValidateInput(input map[string]interface{}) error {
    // Validate inputs
}

func (p *TemplatePlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    // Execute workflow
}
```

## Example Plugins

### Text Summarizer Plugin

The `TextSummarizerPlugin` (`plugins/examples/text_summarizer.go`) demonstrates:

- ✅ Input validation with type checking
- ✅ Multi-step workflow (preprocess → extract → format)
- ✅ Configurable parameters (max_length, style)
- ✅ Structured output with metrics

**Usage Example:**

```go
plugin := NewTextSummarizerPlugin()
result, err := plugin.Execute(ctx, map[string]interface{}{
    "text": "Long text to summarize...",
    "max_length": 200,
    "style": "brief",
})
```

### Workflow Example Plugin

The `ExampleWorkflowPlugin` (`plugins/examples/workflow_example.go`) demonstrates:

- ✅ Multi-step agentic workflow
- ✅ LLM integration
- ✅ Tool calling
- ✅ Data transformation
- ✅ Step dependencies
- ✅ Error handling strategies

**Workflow Steps:**

1. **Analyze Query**: LLM analyzes the user query
2. **Extract Information**: Extract key information from analysis
3. **Execute Tool**: Call MCP tools based on extracted info
4. **Synthesize Result**: LLM synthesizes final result

## Agentic Workflows

Agentic workflows allow plugins to:

- Make decisions based on context
- Call LLM models iteratively
- Execute tools and process results
- Transform data between steps
- Handle errors and retries

### Workflow Step Types

#### 1. LLM Call (`llm_call`)

Calls a language model with a prompt:

```go
{
    Type: "llm_call",
    Config: map[string]interface{}{
        "model": "llama3.2",
        "prompt": "Your prompt here",
        // or
        "messages": []map[string]interface{}{
            {"role": "user", "content": "..."},
        },
    },
}
```

#### 2. Tool Call (`tool_call`)

Executes an MCP tool:

```go
{
    Type: "tool_call",
    Config: map[string]interface{}{
        "tool_name": "mcp.filesystem.read_file",
        "arguments": map[string]interface{}{
            "path": "/path/to/file",
        },
        "merge_state": true, // Merge state into arguments
    },
}
```

#### 3. Data Transform (`data_transform`)

Transforms data between steps:

```go
{
    Type: "data_transform",
    Config: map[string]interface{}{
        "transform": "extract",
        "input_key": "llm_response",
        "fields": []interface{}{"key1", "key2"},
    },
}
```

#### 4. Condition (`condition`)

Evaluates conditional logic:

```go
{
    Type: "condition",
    Config: map[string]interface{}{
        "condition": "some_state_key",
    },
}
```

### Step Dependencies

Steps can depend on previous steps:

```go
{
    ID: "step2",
    Dependencies: []string{"step1"},
    // ...
}
```

### Error Handling

Configure error handling per step:

- `"stop"`: Stop workflow on error (default)
- `"continue"`: Continue to next step
- `"retry"`: Retry step up to MaxRetries

```go
{
    OnError: "continue",
    // ...
}
```

### Creating a Workflow

```go
workflow := &plugins.Workflow{
    ID:          "my_workflow",
    Name:        "My Workflow",
    Description: "Description of what it does",
    MaxRetries:  2,
    Timeout:     30 * time.Second,
    Steps: []plugins.WorkflowStep{
        // Define your steps
    },
}

executor := plugins.NewWorkflowExecutor(llmHandler, toolHandler)
results, err := executor.Execute(ctx, workflow, initialInput)
```

## Best Practices

### 1. Input Validation

- ✅ Always validate required inputs
- ✅ Check input types
- ✅ Validate ranges and formats
- ✅ Provide clear error messages

### 2. Error Handling

- ✅ Return structured errors
- ✅ Use context for cancellation
- ✅ Handle timeouts gracefully
- ✅ Log errors with context

### 3. Workflow Design

- ✅ Keep steps focused and single-purpose
- ✅ Use dependencies to ensure order
- ✅ Configure appropriate timeouts
- ✅ Handle errors at each step

### 4. Output Structure

- ✅ Use consistent output formats
- ✅ Include execution metadata
- ✅ Provide meaningful error messages
- ✅ Document output schema

### 5. Testing

- ✅ Test with valid inputs
- ✅ Test with invalid inputs
- ✅ Test error scenarios
- ✅ Test workflow edge cases

### 6. Documentation

- ✅ Document input/output schemas
- ✅ Explain workflow steps
- ✅ Provide usage examples
- ✅ Document error conditions

## API Reference

### Plugin Interface

```go
type Plugin interface {
    Metadata() PluginMetadata
    ValidateInput(input map[string]interface{}) error
    Execute(ctx context.Context, input map[string]interface{}) (*PluginResult, error)
}
```

### PluginMetadata

```go
type PluginMetadata struct {
    Name           string
    Version        string
    Description    string
    Author         string
    InputSchema    map[string]interface{} // JSON Schema
    OutputSchema   map[string]interface{} // JSON Schema
    RequiredInputs []string
    OptionalInputs map[string]interface{}
}
```

### PluginResult

```go
type PluginResult struct {
    Success  bool
    Data     map[string]interface{}
    Error    string
    Metadata ExecutionMetadata
}
```

### Workflow

```go
type Workflow struct {
    ID          string
    Name        string
    Description string
    Steps       []WorkflowStep
    MaxRetries  int
    Timeout     time.Duration
}
```

### Registry

```go
registry := plugins.NewRegistry()
registry.Register(plugin)
plugin, err := registry.Get("plugin_name")
plugins := registry.List()
```

## Next Steps

1. **Explore Examples**: Review `plugins/examples/` for working examples
2. **Use Template**: Start with `plugins/templates/plugin_template.go`
3. **Read API Docs**: See `internal/plugins/` for detailed API documentation
4. **Build Your Plugin**: Create your own plugin following the guide
5. **Dynamic Configuration**: See [Dynamic Config Use Cases](DYNAMIC_CONFIG_USECASES.md) for advanced patterns
6. **Deep Dive**: See [Plugin System Explained](PLUGIN_SYSTEM_EXPLAINED.md) for comprehensive explanation of workflows and validations

## Support

For questions, issues, or contributions:

- Check existing examples in `plugins/examples/`
- Review the template in `plugins/templates/`
- See API documentation in `internal/plugins/`
- Open an issue on GitHub

---

Happy plugin development! 🚀
