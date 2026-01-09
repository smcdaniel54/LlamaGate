# Plugin System and Workflows - Complete Explanation

This document provides a comprehensive explanation of LlamaGate's plugin system, workflows, and validation mechanisms.

## Table of Contents

1. [Plugin System Overview](#plugin-system-overview)
2. [Plugin Architecture](#plugin-architecture)
3. [Workflow System](#workflow-system)
4. [Validation System](#validation-system)
5. [Complete Example](#complete-example)
6. [Best Practices](#best-practices)

## Plugin System Overview

### What is a Plugin?

A plugin in LlamaGate is a self-contained, reusable component that:
- Processes inputs through defined workflows
- Integrates with LLM models and MCP tools
- Returns structured outputs
- Can expose custom API endpoints
- Can define agentic behaviors

### Core Concepts

**Plugin**: A component that implements the `Plugin` interface
**Workflow**: A sequence of steps that process data
**Agent**: A specialized plugin with structured capabilities
**Validation**: Input/output validation and error handling

## Plugin Architecture

### Plugin Interface

Every plugin must implement three core methods:

```go
type Plugin interface {
    // Metadata returns information about the plugin
    Metadata() PluginMetadata
    
    // ValidateInput validates input parameters
    ValidateInput(input map[string]interface{}) error
    
    // Execute runs the plugin workflow
    Execute(ctx context.Context, input map[string]interface{}) (*PluginResult, error)
}
```

### Plugin Lifecycle

```
1. Registration → Plugin registered with registry
2. Input Validation → ValidateInput() called
3. Execution → Execute() called with validated input
4. Workflow Execution → Workflow steps executed
5. Result → PluginResult returned
```

### Plugin Metadata

Metadata defines the plugin's contract:

```go
type PluginMetadata struct {
    Name           string                 // Unique identifier
    Version        string                 // Version number
    Description    string                 // What it does
    Author         string                 // Plugin author
    InputSchema    map[string]interface{} // JSON Schema for inputs
    OutputSchema   map[string]interface{} // JSON Schema for outputs
    RequiredInputs []string               // Required parameter names
    OptionalInputs map[string]interface{} // Optional params with defaults
}
```

**Example:**

```go
func (p *TextSummarizerPlugin) Metadata() plugins.PluginMetadata {
    return plugins.PluginMetadata{
        Name:        "text_summarizer",
        Version:     "1.0.0",
        Description: "Summarizes text content",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "text": map[string]interface{}{
                    "type":        "string",
                    "description": "Text to summarize",
                },
                "max_length": map[string]interface{}{
                    "type":        "integer",
                    "description": "Maximum summary length",
                    "default":    200,
                },
            },
            "required": []string{"text"},
        },
        RequiredInputs: []string{"text"},
        OptionalInputs: map[string]interface{}{
            "max_length": 200,
        },
    }
}
```

## Workflow System

### What is a Workflow?

A workflow is a sequence of steps that process data. Each step can:
- Call LLM models
- Execute tools
- Transform data
- Evaluate conditions
- Handle errors

### Workflow Structure

```go
type Workflow struct {
    ID          string        // Unique identifier
    Name        string        // Human-readable name
    Description string        // What it does
    Steps       []WorkflowStep // Steps in execution order
    MaxRetries  int           // Max retries for failed steps
    Timeout     time.Duration // Maximum execution time
}
```

### Workflow Steps

Each step defines an operation:

```go
type WorkflowStep struct {
    ID          string                 // Unique step identifier
    Name        string                 // Human-readable name
    Description string                 // What this step does
    Type        string                 // Step type (see below)
    Config      map[string]interface{} // Step configuration
    Dependencies []string              // Step IDs that must complete first
    OnError     string                 // Error handling: "stop", "continue", "retry"
}
```

### Step Types

#### 1. LLM Call (`llm_call`)

Calls a language model:

```go
{
    ID:   "analyze",
    Type: "llm_call",
    Config: map[string]interface{}{
        "model":  "llama3.2",
        "prompt": "Analyze this query...",
        // OR
        "messages": []map[string]interface{}{
            {"role": "user", "content": "..."},
        },
    },
}
```

**How it works:**
1. Extracts model and messages from config
2. Calls LLM handler function
3. Returns LLM response in state

#### 2. Tool Call (`tool_call`)

Executes an MCP tool:

```go
{
    ID:   "read_file",
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

**How it works:**
1. Extracts tool name and arguments
2. Optionally merges workflow state
3. Calls tool handler function
4. Returns tool result in state

#### 3. Data Transform (`data_transform`)

Transforms data between steps:

```go
{
    ID:   "extract",
    Type: "data_transform",
    Config: map[string]interface{}{
        "transform": "extract",
        "input_key": "llm_response",
        "fields": []interface{}{"key1", "key2"},
    },
}
```

**Transform Types:**
- `extract`: Extract specific fields from data
- `format`: Format data using template

**How it works:**
1. Gets input data from workflow state
2. Applies transformation
3. Returns transformed data

#### 4. Condition (`condition`)

Evaluates conditional logic:

```go
{
    ID:   "check",
    Type: "condition",
    Config: map[string]interface{}{
        "condition": "some_state_key",
    },
}
```

**How it works:**
1. Checks if state key exists and is truthy
2. Returns condition result
3. Used for conditional workflow execution

### Workflow Execution Flow

```
1. Initialize State
   ↓
2. For each step (in order):
   a. Check dependencies (all must succeed)
   b. Execute step
   c. Update state with step output
   d. Handle errors based on OnError policy
   ↓
3. Return step results
```

### Step Dependencies

Steps can depend on previous steps:

```go
{
    ID:   "step2",
    Dependencies: []string{"step1"},
    // step2 only executes after step1 succeeds
}
```

**Dependency Rules:**
- All dependencies must complete successfully
- Steps execute in dependency order
- Circular dependencies are invalid

### Error Handling

Each step can define error handling:

- **"stop"** (default): Stop workflow on error
- **"continue"**: Continue to next step
- **"retry"**: Retry step up to MaxRetries

```go
{
    ID:      "optional_step",
    OnError: "continue", // Continue even if this step fails
}
```

### Workflow Executor

The `WorkflowExecutor` executes workflows:

```go
executor := plugins.NewWorkflowExecutor(
    llmHandler,  // Function to handle LLM calls
    toolHandler, // Function to handle tool calls
)

results, err := executor.Execute(ctx, workflow, initialInput)
```

**Execution Process:**
1. Validates workflow
2. Creates context with timeout
3. Executes steps in order
4. Tracks state between steps
5. Handles errors and retries
6. Returns step results

## Validation System

### Three Levels of Validation

#### 1. Input Validation (Plugin Level)

**When:** Before plugin execution
**Where:** `ValidateInput()` method
**Purpose:** Ensure inputs meet plugin requirements

```go
func (p *MyPlugin) ValidateInput(input map[string]interface{}) error {
    // Check required inputs
    if _, exists := input["required_param"]; !exists {
        return fmt.Errorf("required input 'required_param' is missing")
    }
    
    // Validate types
    if text, ok := input["text"].(string); !ok {
        return fmt.Errorf("input 'text' must be a string")
    } else if len(text) == 0 {
        return fmt.Errorf("input 'text' cannot be empty")
    }
    
    // Validate ranges
    if maxLength, exists := input["max_length"]; exists {
        if maxLengthFloat, ok := maxLength.(float64); ok {
            if maxLengthFloat < 50 || maxLengthFloat > 1000 {
                return fmt.Errorf("max_length must be between 50 and 1000")
            }
        }
    }
    
    // Validate enums
    if style, exists := input["style"]; exists {
        validStyles := []string{"brief", "detailed", "bullet"}
        styleStr, ok := style.(string)
        if !ok || !contains(validStyles, styleStr) {
            return fmt.Errorf("style must be one of: %v", validStyles)
        }
    }
    
    return nil
}
```

**Validation Checks:**
- ✅ Required parameters present
- ✅ Parameter types correct
- ✅ Value ranges valid
- ✅ Enum values valid
- ✅ Format validation (if needed)

#### 2. Schema Validation (Metadata Level)

**When:** During plugin registration/query
**Where:** `InputSchema` and `OutputSchema` in metadata
**Purpose:** Document and validate structure

```go
InputSchema: map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "text": map[string]interface{}{
            "type":        "string",
            "description": "Text to process",
            "minLength":   1,
        },
        "max_length": map[string]interface{}{
            "type":        "integer",
            "description": "Maximum length",
            "minimum":     50,
            "maximum":     1000,
            "default":     200,
        },
    },
    "required": []string{"text"},
}
```

**Schema Benefits:**
- ✅ API documentation
- ✅ Client-side validation
- ✅ Type safety
- ✅ Default values

#### 3. Workflow Step Validation (Execution Level)

**When:** During workflow execution
**Where:** Workflow executor
**Purpose:** Ensure step configuration is valid

```go
// Executor validates:
// - Step type is valid
// - Required config fields present
// - Dependencies exist and succeeded
// - Timeout not exceeded
```

**Step Validation:**
- ✅ Step type recognized
- ✅ Config fields valid
- ✅ Dependencies resolved
- ✅ Timeout not exceeded

### Validation Best Practices

#### 1. Validate Early

```go
func (p *MyPlugin) ValidateInput(input map[string]interface{}) error {
    // Validate immediately, fail fast
    if err := p.validateRequired(input); err != nil {
        return err
    }
    if err := p.validateTypes(input); err != nil {
        return err
    }
    return nil
}
```

#### 2. Provide Clear Error Messages

```go
// Bad
return fmt.Errorf("invalid input")

// Good
return fmt.Errorf("input 'max_length' must be between 50 and 1000, got %v", maxLength)
```

#### 3. Validate Types Explicitly

```go
// Check type before use
if text, ok := input["text"].(string); !ok {
    return fmt.Errorf("input 'text' must be a string")
}
// Now safe to use text
```

#### 4. Use Defaults for Optional Inputs

```go
// In Execute(), apply defaults
maxLength := 200 // Default
if ml, exists := input["max_length"]; exists {
    if mlFloat, ok := ml.(float64); ok {
        maxLength = int(mlFloat)
    }
}
```

## Complete Example

### Full Plugin with Validation and Workflow

```go
package examples

import (
    "context"
    "fmt"
    "time"
    "github.com/llamagate/llamagate/internal/plugins"
)

type CompleteExamplePlugin struct {
    executor *plugins.WorkflowExecutor
}

// Metadata defines the plugin contract
func (p *CompleteExamplePlugin) Metadata() plugins.PluginMetadata {
    return plugins.PluginMetadata{
        Name:        "complete_example",
        Version:     "1.0.0",
        Description: "Complete example with validation and workflow",
        
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "query": map[string]interface{}{
                    "type":        "string",
                    "description": "Query to process",
                    "minLength":   1,
                },
                "model": map[string]interface{}{
                    "type":        "string",
                    "description": "LLM model to use",
                    "default":     "llama3.2",
                },
                "use_tools": map[string]interface{}{
                    "type":        "boolean",
                    "description": "Whether to use tools",
                    "default":     true,
                },
            },
            "required": []string{"query"},
        },
        
        OutputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "result": map[string]interface{}{
                    "type":        "string",
                    "description": "Processed result",
                },
            },
        },
        
        RequiredInputs: []string{"query"},
        OptionalInputs: map[string]interface{}{
            "model":     "llama3.2",
            "use_tools": true,
        },
    }
}

// ValidateInput validates inputs before execution
func (p *CompleteExamplePlugin) ValidateInput(input map[string]interface{}) error {
    // 1. Check required inputs
    if query, exists := input["query"]; !exists {
        return fmt.Errorf("required input 'query' is missing")
    } else if queryStr, ok := query.(string); !ok {
        return fmt.Errorf("input 'query' must be a string")
    } else if len(queryStr) == 0 {
        return fmt.Errorf("input 'query' cannot be empty")
    }
    
    // 2. Validate optional inputs if provided
    if model, exists := input["model"]; exists {
        if modelStr, ok := model.(string); !ok {
            return fmt.Errorf("input 'model' must be a string")
        } else if len(modelStr) == 0 {
            return fmt.Errorf("input 'model' cannot be empty")
        }
    }
    
    if useTools, exists := input["use_tools"]; exists {
        if _, ok := useTools.(bool); !ok {
            return fmt.Errorf("input 'use_tools' must be a boolean")
        }
    }
    
    return nil
}

// Execute runs the plugin workflow
func (p *CompleteExamplePlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    startTime := time.Now()
    
    // Apply defaults for optional inputs
    model := "llama3.2"
    if m, exists := input["model"]; exists {
        model, _ = m.(string)
    }
    
    useTools := true
    if ut, exists := input["use_tools"]; exists {
        useTools, _ = ut.(bool)
    }
    
    // Build workflow dynamically
    workflow := p.buildWorkflow(model, useTools)
    
    // Execute workflow
    stepResults, err := p.executor.Execute(ctx, workflow, input)
    if err != nil {
        return &plugins.PluginResult{
            Success: false,
            Error:   err.Error(),
            Metadata: plugins.ExecutionMetadata{
                ExecutionTime: time.Since(startTime),
                StepsExecuted: len(stepResults),
                Timestamp:     time.Now(),
            },
        }, nil
    }
    
    // Extract result
    result := ""
    if len(stepResults) > 0 {
        lastStep := stepResults[len(stepResults)-1]
        if lastStep.Success && lastStep.Output != nil {
            if r, ok := lastStep.Output["result"].(string); ok {
                result = r
            }
        }
    }
    
    return &plugins.PluginResult{
        Success: true,
        Data: map[string]interface{}{
            "result": result,
        },
        Metadata: plugins.ExecutionMetadata{
            ExecutionTime: time.Since(startTime),
            StepsExecuted: len(stepResults),
            Timestamp:     time.Now(),
        },
    }, nil
}

// buildWorkflow creates the workflow based on configuration
func (p *CompleteExamplePlugin) buildWorkflow(model string, useTools bool) *plugins.Workflow {
    steps := []plugins.WorkflowStep{
        {
            ID:          "analyze",
            Name:        "Analyze Query",
            Description: "Analyze the user query",
            Type:        "llm_call",
            Config: map[string]interface{}{
                "model":  model,
                "prompt": fmt.Sprintf("Analyze this query: {{query}}"),
            },
        },
    }
    
    // Conditionally add tool step
    if useTools {
        steps = append(steps, plugins.WorkflowStep{
            ID:          "use_tool",
            Name:        "Use Tool",
            Description: "Execute a tool",
            Type:        "tool_call",
            Config: map[string]interface{}{
                "tool_name": "mcp.filesystem.read_file",
                "merge_state": true,
            },
            Dependencies: []string{"analyze"},
            OnError:      "continue", // Continue even if tool fails
        })
    }
    
    // Add synthesis step
    prevStep := "analyze"
    if useTools {
        prevStep = "use_tool"
    }
    
    steps = append(steps, plugins.WorkflowStep{
        ID:          "synthesize",
        Name:        "Synthesize Result",
        Description: "Synthesize final result",
        Type:        "llm_call",
        Config: map[string]interface{}{
            "model":  model,
            "prompt": "Synthesize a final answer",
        },
        Dependencies: []string{prevStep},
    })
    
    return &plugins.Workflow{
        ID:          "complete_example_workflow",
        Name:        "Complete Example Workflow",
        Description: "Example workflow with validation",
        Steps:       steps,
        MaxRetries:  2,
        Timeout:     30 * time.Second,
    }
}

func NewCompleteExamplePlugin(executor *plugins.WorkflowExecutor) plugins.Plugin {
    return &CompleteExamplePlugin{
        executor: executor,
    }
}
```

### Execution Flow

```
1. Plugin Registration
   registry.Register(NewCompleteExamplePlugin(executor))
   
2. API Request
   POST /v1/plugins/complete_example/execute
   {
     "query": "Process this",
     "model": "llama3.2",
     "use_tools": true
   }
   
3. Input Validation
   ValidateInput() called
   - Checks query is present and non-empty
   - Validates model is string
   - Validates use_tools is boolean
   
4. Execution
   Execute() called
   - Applies defaults
   - Builds workflow
   - Executes workflow steps
   
5. Workflow Execution
   - Step 1: analyze (LLM call)
   - Step 2: use_tool (tool call, if enabled)
   - Step 3: synthesize (LLM call)
   
6. Result
   {
     "success": true,
     "data": {
       "result": "..."
     },
     "metadata": {
       "execution_time": "2.5s",
       "steps_executed": 3
     }
   }
```

## Best Practices

### Validation

1. **Validate Early**: Check inputs before processing
2. **Clear Errors**: Provide specific, actionable error messages
3. **Type Safety**: Always check types before use
4. **Defaults**: Apply defaults for optional inputs
5. **Schema**: Define JSON Schema for documentation

### Workflows

1. **Single Purpose**: Each step should do one thing
2. **Dependencies**: Use dependencies to ensure order
3. **Error Handling**: Configure OnError appropriately
4. **Timeouts**: Set reasonable timeouts
5. **Retries**: Use retries for transient failures

### Plugins

1. **Metadata**: Complete, accurate metadata
2. **Validation**: Comprehensive input validation
3. **Error Handling**: Graceful error handling
4. **Documentation**: Clear documentation
5. **Testing**: Test with various inputs

## Summary

The plugin system provides:

- ✅ **Structured Plugins**: Clear interface and metadata
- ✅ **Flexible Workflows**: Multi-step processing with dependencies
- ✅ **Comprehensive Validation**: Input, schema, and execution validation
- ✅ **Error Handling**: Configurable error strategies
- ✅ **Type Safety**: Strong typing throughout
- ✅ **Extensibility**: Easy to extend and customize

All components work together to provide a robust, flexible system for building agentic workflows and plugins.
