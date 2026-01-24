# Extension Foundation Quick Start

## Overview

This guide will help you get started with the LlamaGate Extension Foundation system. You'll learn how to use **builtin extensions** and create custom extensions to build agentic workflows.

**Note:** This guide describes the **Go-based Extension Foundation** system. LlamaGate also has a **YAML-based extension system** (v0.9.1+) with its own builtin extensions. See [Extensions Quick Start](../EXTENSIONS_QUICKSTART.md) for YAML extensions.

## Builtin Extensions (Go Code)

LlamaGate includes 10 builtin extensions (Go code) that are always available:

- **Agent Caller** - Make local LLM calls
- **Tool Framework** - Execute MCP tools
- **Decision Engine** - Make routing decisions
- **State Manager** - Track workflow state
- **Event System** - Publish/subscribe to events
- **Validation Framework** - Validate outputs
- **Human Interaction** - Handle human input
- **Transformation Pipeline** - Transform data
- **Debug Logger** - Comprehensive logging
- **Visual Debugger** - Visual workflow execution

See [Builtin Extensions Guide](../BUILTIN_EXTENSIONS.md) for complete documentation.

**Extension Types in LlamaGate:**
- **Builtin Extensions (Go Code)**: This Extension Foundation system - `internal/extensions/builtin/` - Core functionality compiled into binary
- **Builtin Extensions (YAML-based)**: YAML extension system - `extensions/builtin/` - Core workflow capabilities, `builtin: true` flag
- **Default Extensions (YAML-based)**: YAML extension system - `extensions/` - Regular workflow extensions

See [Extensions Quick Start](../EXTENSIONS_QUICKSTART.md) for information about YAML-based extensions.

## Prerequisites

- Go 1.21 or later
- LlamaGate installed and running
- Basic understanding of Go

## Installation

The extension foundation is included in LlamaGate. No additional installation is required.

## Basic Usage

### 1. Use Builtin Extensions

Builtin extensions are already available - just get them from the registry:

```go
package main

import (
    "context"
    "github.com/llamagate/llamagate/internal/extensions/builtin/core"
    "github.com/llamagate/llamagate/internal/extensions/builtin/agent"
)

func main() {
    ctx := context.Background()
    registry := core.GetRegistry()
    
    // Register builtin agent caller (if not already registered)
    caller := agent.NewDefaultCaller("default", "1.0.0", "http://localhost:11434", "")
    registry.Register(ctx, caller, map[string]interface{}{
        "base_url": "http://localhost:11434",
        "timeout":  "30s",
    })
    
    // Or just get it if already registered
    caller, _ = registry.GetAgentCaller("default")
}
```

### 2. Use Extensions

```go
// Get agent caller
caller, err := registry.GetAgentCaller("default")
if err != nil {
    log.Fatal(err)
}

// Make a call
response, err := caller.Call(ctx, &core.AgentRequest{
    Model: "mistral",
    Messages: []core.Message{
        {Role: "user", Content: "Hello, world!"},
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(response.Content)
```

### 3. Create a Custom Extension

```go
package myextension

import (
    "context"
    "github.com/llamagate/llamagate/internal/extensions/builtin/core"
)

type MyExtension struct {
    name string
}

func (e *MyExtension) Name() string {
    return e.name
}

func (e *MyExtension) Version() string {
    return "1.0.0"
}

func (e *MyExtension) Initialize(ctx context.Context, config map[string]interface{}) error {
    // Initialize your extension
    return nil
}

func (e *MyExtension) Shutdown(ctx context.Context) error {
    // Cleanup
    return nil
}

// Register it
func init() {
    registry := core.GetRegistry()
    ext := &MyExtension{name: "my-extension"}
    registry.Register(context.Background(), ext, nil)
}
```

## Common Patterns

### Making Agent Calls

```go
caller, _ := registry.GetAgentCaller("default")

// Synchronous call
response, err := caller.Call(ctx, &core.AgentRequest{
    Model: "mistral",
    Messages: []core.Message{
        {Role: "user", Content: "What is 2+2?"},
    },
})

// Streaming call
ch, err := caller.CallStream(ctx, &core.AgentRequest{
    Model: "mistral",
    Messages: []core.Message{
        {Role: "user", Content: "Count to 10"},
    },
    Stream: true,
})

for chunk := range ch {
    if chunk.Error != nil {
        log.Fatal(chunk.Error)
    }
    fmt.Print(chunk.Content)
    if chunk.Done {
        break
    }
}
```

### Executing Tools

```go
framework, _ := registry.GetToolExecutor("default")

// Register a tool
framework.RegisterTool(&tools.ToolHandler{
    Definition: &core.ToolDefinition{
        Name:        "add",
        Description: "Add two numbers",
        Parameters: map[string]interface{}{
            "a": map[string]interface{}{"type": "number"},
            "b": map[string]interface{}{"type": "number"},
        },
    },
    Execute: func(ctx context.Context, params map[string]interface{}) (*core.ToolResult, error) {
        a := params["a"].(float64)
        b := params["b"].(float64)
        return &core.ToolResult{
            Success: true,
            Output:  a + b,
        }, nil
    },
})

// Execute the tool
result, err := framework.Execute(ctx, "add", map[string]interface{}{
    "a": 5,
    "b": 3,
})
```

### Managing State

```go
manager, _ := registry.GetStateManager("default")

// Save state
manager.SaveState(ctx, "workflow-123", &core.WorkflowState{
    Status: "running",
    Step:   "step1",
    Context: map[string]interface{}{
        "data": "value",
    },
})

// Load state
state, err := manager.LoadState(ctx, "workflow-123")

// Update context
manager.UpdateContext(ctx, "workflow-123", map[string]interface{}{
    "new_data": "new_value",
})
```

### Publishing Events

```go
publisher, _ := registry.GetEventPublisher("default")

// Publish an event
publisher.Publish(ctx, &core.Event{
    Type: "workflow.started",
    Source: "my-extension",
    Data: map[string]interface{}{
        "workflow_id": "123",
    },
})
```

### Subscribing to Events

```go
publisher, _ := registry.GetEventPublisher("default")

// Subscribe to events
subscription, _ := publisher.Subscribe(ctx, &core.EventFilter{
    Types: []string{"workflow.*"},
}, func(ctx context.Context, event *core.Event) error {
    fmt.Printf("Received event: %s\n", event.Type)
    return nil
})

// Later, unsubscribe
defer subscription.Unsubscribe(ctx)
```

### Making Decisions

```go
engine, _ := registry.GetDecisionEvaluator("default")

// Evaluate a condition
decision, err := engine.Evaluate(ctx, &core.Condition{
    Type: "llm",
    Expression: "Is this email urgent?",
    Context: map[string]interface{}{
        "email": "Customer complaint about service",
    },
})

if decision.Result {
    // Handle urgent case
}

// Evaluate branches
branch, err := engine.EvaluateBranch(ctx, []*core.Branch{
    {
        ID:       "urgent",
        Name:     "Urgent",
        Priority: 10,
        Condition: &core.Condition{
            Type:       "llm",
            Expression: "Is this urgent?",
        },
    },
    {
        ID:       "normal",
        Name:     "Normal",
        Priority: 5,
        Condition: &core.Condition{
            Type:       "llm",
            Expression: "Is this normal?",
        },
    },
})
```

### Requesting Human Approval

```go
interaction, _ := registry.GetHumanInteraction("default")

// Request approval
response, err := interaction.RequestApproval(ctx, &core.ApprovalRequest{
    Title:       "Approve Transaction",
    Description: "Transaction amount: $1000",
    Options:     []string{"approve", "reject"},
    Timeout:     func() *time.Duration { d := 1 * time.Hour; return &d }(),
})

if response.Approved {
    // Proceed
}
```

## Example: Simple Workflow

```go
package main

import (
    "context"
    "fmt"
    "github.com/llamagate/llamagate/internal/extensions/builtin/core"
    "github.com/llamagate/llamagate/internal/extensions/builtin/agent"
    "github.com/llamagate/llamagate/internal/extensions/builtin/decision"
    "github.com/llamagate/llamagate/internal/extensions/builtin/state"
)

func main() {
    ctx := context.Background()
    registry := core.GetRegistry()
    
    // Register extensions
    caller := agent.NewDefaultCaller("default", "1.0.0", "http://localhost:11434", "")
    registry.Register(ctx, caller, nil)
    
    engine := decision.NewEngine("default", "1.0.0")
    registry.Register(ctx, engine, nil)
    
    manager := state.NewManager("default", "1.0.0")
    registry.Register(ctx, manager, nil)
    
    // Run workflow
    workflowID := "workflow-123"
    
    // Step 1: Classify input
    caller, _ := registry.GetAgentCaller("default")
    response, _ := caller.Call(ctx, &core.AgentRequest{
        Model: "mistral",
        Messages: []core.Message{
            {Role: "user", Content: "Classify: Customer complaint"},
        },
    })
    
    // Step 2: Make decision
    engine, _ := registry.GetDecisionEvaluator("default")
    decision, _ := engine.Evaluate(ctx, &core.Condition{
        Type:       "llm",
        Expression: fmt.Sprintf("Is '%s' urgent?", response.Content),
    })
    
    // Step 3: Save state
    manager, _ := registry.GetStateManager("default")
    manager.SaveState(ctx, workflowID, &core.WorkflowState{
        Status: "completed",
        Context: map[string]interface{}{
            "classification": response.Content,
            "urgent":         decision.Result,
        },
    })
    
    fmt.Printf("Workflow completed: %s\n", workflowID)
}
```

## Next Steps

- Read the [Architecture Guide](ARCHITECTURE.md) for detailed information
- Check out [Examples](EXAMPLES.md) for more complex use cases
- Explore the API documentation
- Build your own extensions

## Troubleshooting

### Extension Not Found

Make sure the extension is registered before use:

```go
registry.Register(ctx, extension, config)
```

### Context Cancellation

Always respect context cancellation:

```go
select {
case <-ctx.Done():
    return ctx.Err()
default:
    // Continue
}
```

### Event Not Received

Check that:
1. Event publisher is registered
2. Event filter matches
3. Handler doesn't return errors

## Support

For questions or issues:
- Check the documentation
- Open an issue on GitHub
- Contact the maintainers
