# LlamaGate Plugin System

The LlamaGate plugin system allows you to create reusable, agentic workflows that integrate with local LLM models and MCP tools. This comprehensive guide covers everything you need to know.

## Table of Contents

- [Quick Start](#quick-start)
- [Overview](#overview)
- [Creating a Plugin](#creating-a-plugin)
- [Agentic Workflows](#agentic-workflows)
- [Custom API Endpoints](#custom-api-endpoints)
- [JSON/YAML Definitions](#jsonyaml-definitions)
- [Dynamic Configuration Examples](#dynamic-configuration-examples)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)

## Quick Start

**Get started in 5 minutes:**

1. **Use the simple template** (~50 lines):
   ```bash
   cat plugins/templates/simple_plugin.go
   ```

2. **Copy and customize**:
   ```bash
   cp plugins/templates/simple_plugin.go plugins/my_plugin.go
   ```

3. **Implement 3 methods**:
   - `Metadata()` - Describe your plugin
   - `ValidateInput()` - Validate inputs
   - `Execute()` - Do the work

4. **Register your plugin**:
   ```go
   registry := plugins.NewRegistry()
   registry.Register(NewMyPlugin())
   ```

See [Plugin Quick Start](PLUGIN_QUICKSTART.md) for detailed step-by-step instructions.

## Overview

Plugins in LlamaGate are self-contained modules that:

- ✅ Define clear input/output schemas
- ✅ Implement validation logic
- ✅ Execute agentic workflows
- ✅ Integrate with LLM models and MCP tools
- ✅ Return structured results
- ✅ Can expose custom API endpoints
- ✅ Support JSON/YAML definitions (model-friendly)

### Key Concepts

**Plugin**: A reusable component that processes inputs and produces outputs through a defined workflow.

**Workflow**: A sequence of steps that can include:
- LLM calls (chat completions)
- Tool calls (MCP tools)
- Data transformations
- Conditional logic

**Agentic**: The workflow can make decisions, call tools, and iterate based on results.

### Why Plugins?

- **Simple**: Only 3 methods required (~50 lines for basic plugin)
- **Flexible**: Progressive complexity - start simple, add features as needed
- **Model-Friendly**: Can be defined in JSON/YAML for AI models
- **Powerful**: Support complex multi-step workflows with LLMs and tools

## Creating a Plugin

### Step 1: Use the Template

Start with the simple template (recommended) or full template:

```bash
# Simple template (~50 lines)
cp plugins/templates/simple_plugin.go plugins/my_plugin.go

# Full template (more features)
cp plugins/templates/plugin_template.go plugins/my_plugin.go
```

### Step 2: Define Metadata

Update the `Metadata()` method:

```go
func (p *MyPlugin) Metadata() plugins.PluginMetadata {
    return plugins.PluginMetadata{
        Name:        "my_plugin",
        Version:     "1.0.0",
        Description: "What my plugin does",
        Author:      "Your Name",
        
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "input_param": map[string]interface{}{
                    "type":        "string",
                    "description": "Your input parameter",
                },
            },
            "required": []string{"input_param"},
        },
        
        OutputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "result": map[string]interface{}{
                    "type": "string",
                },
            },
        },
        
        RequiredInputs: []string{"input_param"},
        OptionalInputs: map[string]interface{}{
            "optional_param": "default_value",
        },
    }
}
```

### Step 3: Implement Validation

Add validation logic:

```go
func (p *MyPlugin) ValidateInput(input map[string]interface{}) error {
    // Check required inputs
    if _, exists := input["input_param"]; !exists {
        return fmt.Errorf("required input 'input_param' is missing")
    }
    
    // Type validation
    if val, ok := input["input_param"].(string); !ok {
        return fmt.Errorf("input_param must be a string")
    }
    
    // Custom validation
    if len(val) == 0 {
        return fmt.Errorf("input_param cannot be empty")
    }
    
    return nil
}
```

### Step 4: Implement Execution

Add your workflow logic:

```go
func (p *MyPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    startTime := time.Now()
    
    // Your workflow logic here
    inputParam := input["input_param"].(string)
    result := processInput(inputParam)
    
    return &plugins.PluginResult{
        Success: true,
        Data: map[string]interface{}{
            "result": result,
        },
        Metadata: plugins.ExecutionMetadata{
            ExecutionTime: time.Since(startTime),
            StepsExecuted: 1,
            Timestamp:     time.Now(),
        },
    }, nil
}
```

### Step 5: Register Your Plugin

Register in your application:

```go
registry := plugins.NewRegistry()
if err := registry.Register(NewMyPlugin()); err != nil {
    log.Error().Err(err).Msg("Failed to register plugin")
}
```

## Agentic Workflows

Agentic workflows allow plugins to make decisions, call LLMs iteratively, execute tools, and transform data between steps.

### Workflow Step Types

#### 1. LLM Call (`llm_call`)

Calls a language model:

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
        "merge_state": true, // Merge workflow state into arguments
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

### Creating a Workflow

```go
workflow := &plugins.Workflow{
    ID:          "my_workflow",
    Name:        "My Workflow",
    Description: "Description of what it does",
    MaxRetries:  2,
    Timeout:     30 * time.Second,
    Steps: []plugins.WorkflowStep{
        {
            ID:   "step1",
            Type: "llm_call",
            Config: map[string]interface{}{
                "model":  "llama3.2",
                "prompt": "Analyze this query",
            },
        },
        {
            ID:   "step2",
            Type: "tool_call",
            Config: map[string]interface{}{
                "tool_name": "mcp.filesystem.read_file",
                "arguments": map[string]interface{}{
                    "path": "/path/to/file",
                },
            },
            Dependencies: []string{"step1"},
            OnError:      "continue", // Continue on error
        },
    },
}

executor := plugins.NewWorkflowExecutor(llmHandler, toolHandler)
results, err := executor.Execute(ctx, workflow, initialInput)
```

### Step Dependencies

Steps can depend on previous steps:

```go
{
    ID: "step2",
    Dependencies: []string{"step1"},
    // step2 will only run after step1 completes successfully
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

## Custom API Endpoints

Plugins can expose custom HTTP endpoints beyond the standard `/v1/plugins/:name/execute` endpoint.

### Defining Custom Endpoints

Implement the `ExtendedPlugin` interface:

```go
type MyPlugin struct {
    // ... plugin fields
}

func (p *MyPlugin) GetAPIEndpoints() []plugins.APIEndpoint {
    return []plugins.APIEndpoint{
        {
            Path:        "/custom/action",
            Method:      "POST",
            Description: "Perform a custom action",
            Handler:     p.handleCustomAction,
            RequestSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "action": map[string]interface{}{
                        "type": "string",
                    },
                },
            },
            RequiresAuth:      true,
            RequiresRateLimit: true,
        },
    }
}

func (p *MyPlugin) handleCustomAction(c *gin.Context) {
    var input map[string]interface{}
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    result, err := p.Execute(c.Request.Context(), input)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, result)
}
```

Endpoints are automatically registered at `/v1/plugins/{plugin_name}/custom/action`.

## JSON/YAML Definitions

Plugins can be defined in JSON/YAML, making them model-friendly and easy to generate programmatically.

### Defining a Plugin in JSON

```json
{
  "name": "text_processor",
  "version": "1.0.0",
  "description": "Processes and analyzes text",
  "input_schema": {
    "type": "object",
    "properties": {
      "text": {
        "type": "string",
        "description": "Text to process"
      },
      "operation": {
        "type": "string",
        "enum": ["summarize", "analyze", "extract"],
        "description": "Operation to perform"
      }
    },
    "required": ["text", "operation"]
  },
  "output_schema": {
    "type": "object",
    "properties": {
      "result": {
        "type": "string",
        "description": "Processed result"
      }
    }
  },
  "workflow": {
    "steps": [
      {
        "id": "process",
        "type": "llm_call",
        "config": {
          "model": "llama3.2",
          "prompt": "Process this text: {{text}} with operation: {{operation}}"
        }
      }
    ]
  }
}
```

### Creating Plugin from Definition

```go
import "github.com/llamagate/llamagate/internal/plugins"

// Parse JSON definition
jsonData := []byte(`{...}`)
def, err := plugins.ParsePluginDefinition(jsonData)
if err != nil {
    return err
}

// Create plugin
plugin, err := plugins.CreatePluginFromDefinition(def)
if err != nil {
    return err
}

// Register
registry.Register(plugin)
```

### Model-Friendly Features

- ✅ **Self-Documenting**: Clear metadata and schemas
- ✅ **JSON-Serializable**: All structures are JSON-compatible
- ✅ **Declarative**: Workflows defined declaratively
- ✅ **Minimal Requirements**: Only name and description required

See example: `plugins/examples/model_generated_example.json`

## Dynamic Configuration Examples

Plugins can adapt their behavior based on runtime parameters, environment variables, and user input.

### Example 1: Environment-Aware Configuration

```go
func (p *MyPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    env := os.Getenv("ENVIRONMENT")
    if env == "" {
        env = "development"
    }
    
    var timeout time.Duration
    switch env {
    case "production":
        timeout = 30 * time.Second
    case "staging":
        timeout = 20 * time.Second
    default:
        timeout = 10 * time.Second
    }
    
    workflow := &plugins.Workflow{
        Timeout: timeout,
        Steps:   []plugins.WorkflowStep{...},
    }
    
    // Execute with dynamic timeout
    // ...
}
```

### Example 2: User-Configurable Workflow

```go
func (p *MyPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    queryType := input["query_type"].(string)
    useCache := input["use_cache"].(bool)
    
    steps := []plugins.WorkflowStep{
        {
            ID:   "analyze",
            Type: "llm_call",
            Config: map[string]interface{}{
                "model":  input["model"].(string),
                "prompt": fmt.Sprintf("Analyze this %s query", queryType),
            },
        },
    }
    
    // Add conditional steps
    if useCache {
        steps = append(steps, plugins.WorkflowStep{
            ID:   "check_cache",
            Type: "data_transform",
            // ...
        })
    }
    
    workflow := &plugins.Workflow{Steps: steps}
    // ...
}
```

### Example 3: Configuration-Driven Tool Selection

```go
func (p *MyPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    enabledTools := input["enabled_tools"].([]interface{})
    
    steps := []plugins.WorkflowStep{}
    
    // Add tool steps dynamically
    for i, toolName := range enabledTools {
        steps = append(steps, plugins.WorkflowStep{
            ID:   fmt.Sprintf("tool_%d", i),
            Type: "tool_call",
            Config: map[string]interface{}{
                "tool_name": toolName.(string),
            },
        })
    }
    
    // ...
}
```

See `plugins/examples/dynamic_config_example.go` for complete examples.

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

### Core Plugin Interface

All plugins must implement the `Plugin` interface:

```go
type Plugin interface {
    // Metadata returns information about the plugin
    Metadata() PluginMetadata
    
    // ValidateInput validates the input parameters before execution
    ValidateInput(input map[string]interface{}) error
    
    // Execute runs the plugin workflow and returns results
    Execute(ctx context.Context, input map[string]interface{}) (*PluginResult, error)
}
```

**Example Implementation:**

```go
type MyPlugin struct{}

func (p *MyPlugin) Metadata() plugins.PluginMetadata {
    return plugins.PluginMetadata{
        Name:        "my_plugin",
        Version:     "1.0.0",
        Description: "Does something useful",
        RequiredInputs: []string{"input"},
    }
}

func (p *MyPlugin) ValidateInput(input map[string]interface{}) error {
    if _, exists := input["input"]; !exists {
        return fmt.Errorf("input is required")
    }
    return nil
}

func (p *MyPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    // Your logic here
    return &plugins.PluginResult{
        Success: true,
        Data: map[string]interface{}{"result": "done"},
        Metadata: plugins.ExecutionMetadata{
            ExecutionTime: time.Since(start),
            StepsExecuted: 1,
            Timestamp: time.Now(),
        },
    }, nil
}
```

### PluginMetadata

Complete metadata structure:

```go
type PluginMetadata struct {
    // Name is the unique identifier for the plugin (required)
    Name string `json:"name"`
    
    // Version is the plugin version (e.g., "1.0.0")
    Version string `json:"version"`
    
    // Description describes what the plugin does
    Description string `json:"description"`
    
    // Author is the plugin author (optional)
    Author string `json:"author,omitempty"`
    
    // InputSchema defines the expected input parameters (JSON Schema)
    InputSchema map[string]interface{} `json:"input_schema"`
    
    // OutputSchema defines the expected output structure (JSON Schema)
    OutputSchema map[string]interface{} `json:"output_schema"`
    
    // RequiredInputs lists required input parameter names
    RequiredInputs []string `json:"required_inputs"`
    
    // OptionalInputs lists optional input parameter names with defaults
    OptionalInputs map[string]interface{} `json:"optional_inputs,omitempty"`
}
```

### PluginResult

Result structure returned by `Execute()`:

```go
type PluginResult struct {
    // Success indicates if the execution was successful
    Success bool `json:"success"`
    
    // Data contains the output data
    Data map[string]interface{} `json:"data,omitempty"`
    
    // Error contains error information if execution failed
    Error string `json:"error,omitempty"`
    
    // Metadata contains execution metadata
    Metadata ExecutionMetadata `json:"metadata"`
}

type ExecutionMetadata struct {
    // ExecutionTime is how long the plugin took to execute
    ExecutionTime time.Duration `json:"execution_time"`
    
    // StepsExecuted is the number of workflow steps executed
    StepsExecuted int `json:"steps_executed"`
    
    // Timestamp is when the execution completed
    Timestamp time.Time `json:"timestamp"`
}
```

### PluginContext

The `PluginContext` provides plugins with access to LlamaGate services:

```go
type PluginContext struct {
    // LLMHandler is a function that plugins can use to make LLM calls
    LLMHandler LLMHandlerFunc
    
    // Logger is a plugin-specific logger instance
    Logger zerolog.Logger
    
    // Config is plugin-specific configuration
    Config map[string]interface{}
    
    // HTTPClient is an HTTP client for making external requests
    HTTPClient *http.Client
    
    // PluginName is the name of the plugin (for logging context)
    PluginName string
}
```

**Accessing PluginContext:**

```go
// In your plugin Execute method
func (p *MyPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    // Get plugin context from registry
    pluginCtx := p.registry.GetContext("my_plugin")
    if pluginCtx == nil {
        return nil, fmt.Errorf("plugin context not available")
    }
    
    // Use LLM handler
    response, err := pluginCtx.CallLLM(ctx, "llama3.2", messages, options)
    if err != nil {
        pluginCtx.LogError(err).Msg("LLM call failed")
        return nil, err
    }
    
    // Use logger
    pluginCtx.LogInfo().Msg("Processing complete")
    
    // Access configuration
    timeout := pluginCtx.GetConfigInt("timeout", 30)
    
    return result, nil
}
```

**PluginContext Methods:**

```go
// CallLLM makes an LLM call and returns the response content
func (ctx *PluginContext) CallLLM(
    pluginCtx context.Context,
    model string,
    messages []map[string]interface{},
    options map[string]interface{},
) (string, error)

// GetConfig retrieves a configuration value
func (ctx *PluginContext) GetConfig(key string, defaultValue interface{}) interface{}

// GetConfigString retrieves a string configuration value
func (ctx *PluginContext) GetConfigString(key string, defaultValue string) string

// GetConfigBool retrieves a boolean configuration value
func (ctx *PluginContext) GetConfigBool(key string, defaultValue bool) bool

// GetConfigInt retrieves an integer configuration value
func (ctx *PluginContext) GetConfigInt(key string, defaultValue int) int

// LogInfo returns an info-level logger event with plugin context
func (ctx *PluginContext) LogInfo() *zerolog.Event

// LogError logs an error message with plugin context
func (ctx *PluginContext) LogError(err error) *zerolog.Event

// LogWarn logs a warning message with plugin context
func (ctx *PluginContext) LogWarn() *zerolog.Event

// LogDebug logs a debug message with plugin context
func (ctx *PluginContext) LogDebug() *zerolog.Event
```

### ExtendedPlugin Interface

For plugins that need custom API endpoints or agent definitions:

```go
type ExtendedPlugin interface {
    Plugin
    
    // GetAPIEndpoints returns API endpoint definitions that this plugin exposes
    GetAPIEndpoints() []APIEndpoint
    
    // GetAgentDefinition returns the agent definition if this plugin represents an agent
    GetAgentDefinition() *AgentDefinition
}
```

**APIEndpoint Structure:**

```go
type APIEndpoint struct {
    // Path is the endpoint path (e.g., "/custom/endpoint")
    // Will be prefixed with /v1/plugins/{plugin_name}
    Path string `json:"path"`
    
    // Method is the HTTP method (GET, POST, PUT, DELETE, etc.)
    Method string `json:"method"`
    
    // Handler is the Gin handler function for this endpoint
    Handler gin.HandlerFunc `json:"-"`
    
    // Description describes what this endpoint does
    Description string `json:"description"`
    
    // RequestSchema defines the expected request body schema (JSON Schema)
    RequestSchema map[string]interface{} `json:"request_schema,omitempty"`
    
    // ResponseSchema defines the expected response schema (JSON Schema)
    ResponseSchema map[string]interface{} `json:"response_schema,omitempty"`
    
    // RequiresAuth indicates if this endpoint requires authentication
    RequiresAuth bool `json:"requires_auth"`
    
    // RequiresRateLimit indicates if this endpoint should be rate limited
    RequiresRateLimit bool `json:"requires_rate_limit"`
}
```

### Registry API

The `Registry` manages plugin registration and lookup:

```go
// Create a new registry
registry := plugins.NewRegistry()

// Register a plugin
err := registry.Register(plugin)

// Register a plugin with context
err := registry.RegisterWithContext(plugin, pluginContext)

// Get a plugin by name
plugin, err := registry.Get("plugin_name")

// Get plugin context
context := registry.GetContext("plugin_name")

// Set plugin context
registry.SetContext("plugin_name", context)

// List all registered plugins
plugins := registry.List() // Returns []PluginMetadata
```

### HTTP API Endpoints

#### List All Plugins

**Endpoint:** `GET /v1/plugins`

**Response:**
```json
{
  "plugins": [
    {
      "name": "my_plugin",
      "version": "1.0.0",
      "description": "Does something useful",
      "input_schema": {...},
      "output_schema": {...},
      "required_inputs": ["input"]
    }
  ],
  "count": 1
}
```

**Example:**
```bash
curl -X GET http://localhost:11435/v1/plugins \
  -H "X-API-Key: sk-llamagate"
```

#### Get Plugin Metadata

**Endpoint:** `GET /v1/plugins/:name`

**Response:**
```json
{
  "name": "my_plugin",
  "version": "1.0.0",
  "description": "Does something useful",
  "input_schema": {...},
  "output_schema": {...},
  "required_inputs": ["input"]
}
```

**Example:**
```bash
curl -X GET http://localhost:11435/v1/plugins/my_plugin \
  -H "X-API-Key: sk-llamagate"
```

#### Execute Plugin

**Endpoint:** `POST /v1/plugins/:name/execute`

**Request Body:**
```json
{
  "input": "value",
  "optional_param": "optional_value"
}
```

**Response (Success):**
```json
{
  "success": true,
  "data": {
    "result": "output"
  },
  "metadata": {
    "execution_time": "100ms",
    "steps_executed": 1,
    "timestamp": "2026-01-10T12:00:00Z"
  }
}
```

**Response (Error):**
```json
{
  "success": false,
  "error": "Error message",
  "metadata": {
    "execution_time": "50ms",
    "steps_executed": 0,
    "timestamp": "2026-01-10T12:00:00Z"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:11435/v1/plugins/my_plugin/execute \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{"input": "test"}'
```

### Workflow Types

**WorkflowStep Types:**

- `"llm_call"` - Make an LLM call
- `"tool_call"` - Call an MCP tool
- `"data_transform"` - Transform data
- `"condition"` - Conditional logic

**Error Handling (`OnError`):**

- `"stop"` - Stop workflow on error (default)
- `"continue"` - Continue to next step
- `"retry"` - Retry the step (up to MaxRetries)

## Examples

- **Text Summarizer**: `plugins/examples/text_summarizer.go`
- **Workflow Example**: `plugins/examples/workflow_example.go`
- **Dynamic Config**: `plugins/examples/dynamic_config_example.go`
- **Model-Generated**: `plugins/examples/model_generated_example.json`

## Templates

- **Simple Template**: `plugins/templates/simple_plugin.go` (~50 lines)
- **Full Template**: `plugins/templates/plugin_template.go` (complete features)

## Next Steps

1. **Quick Start**: See [Plugin Quick Start](PLUGIN_QUICKSTART.md) for step-by-step guide
2. **Explore Examples**: Review `plugins/examples/` for working examples
3. **Use Template**: Start with `plugins/templates/simple_plugin.go`
4. **Read Code**: See `internal/plugins/` for detailed API documentation
5. **Build Your Plugin**: Create your own plugin following this guide

## Support

For questions, issues, or contributions:

- Check existing examples in `plugins/examples/`
- Review templates in `plugins/templates/`
- See API documentation in `internal/plugins/`
- Open an issue on GitHub

---

Happy plugin development! 🚀
