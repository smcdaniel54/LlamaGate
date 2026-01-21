# Extension Foundation Architecture

## Overview

The Extension Foundation provides a standardized way to extend LlamaGate with custom functionality for building agentic workflows. It consists of core interfaces, a registration system, and **builtin extensions** that are always available.

## Builtin Extensions

LlamaGate includes **10 builtin extensions** that provide core functionality:

1. **Agent Caller** - Local LLM calls (Ollama)
2. **Tool Framework** - MCP tool execution
3. **Decision Engine** - Condition evaluation and routing
4. **State Manager** - Workflow state tracking
5. **Event System** - Event publishing and subscription
6. **Validation Framework** - Output validation
7. **Human Interaction** - Human-in-the-loop workflows
8. **Transformation Pipeline** - Data transformation
9. **Debug Logger** - Comprehensive logging
10. **Visual Debugger** - Visual workflow execution

All builtin extensions are part of LlamaGate core and are always available. See [Builtin Extensions Guide](../BUILTIN_EXTENSIONS.md) for complete documentation.

## Core Components

### Extension Registry

The `Registry` is the central component that manages all extensions. It provides:

- **Registration**: Register extensions with initialization
- **Discovery**: Find extensions by name or type
- **Lifecycle**: Manage extension initialization and shutdown
- **Type Categorization**: Automatically categorize extensions by interface

```go
registry := core.GetRegistry()
registry.Register(ctx, extension, config)
caller, _ := registry.GetAgentCaller("default")
```

### Extension Interfaces

All extensions implement the base `Extension` interface:

```go
type Extension interface {
    Name() string
    Version() string
    Initialize(ctx context.Context, config map[string]interface{}) error
    Shutdown(ctx context.Context) error
}
```

Specialized interfaces extend this base:

- `AgentCaller`: Make LLM calls
- `ToolExecutor`: Execute tools
- `DecisionEvaluator`: Evaluate decisions
- `StateManager`: Manage workflow state
- `EventPublisher`: Publish/subscribe to events
- `Validator`: Validate outputs
- `HumanInteraction`: Handle human input
- `Transformer`: Transform data

## Extension Points

### 1. Agent Caller

Provides a standardized interface for making LLM calls:

```go
caller, _ := registry.GetAgentCaller("default")
response, _ := caller.Call(ctx, &core.AgentRequest{
    Model: "mistral",
    Messages: []core.Message{
        {Role: "user", Content: "Hello"},
    },
})
```

**Features:**
- Synchronous and streaming calls
- Context passing
- Error handling
- Event publishing

### 2. Tool Framework

Manages tool registration and execution:

```go
framework, _ := registry.GetToolExecutor("default")
framework.RegisterTool(&tools.ToolHandler{
    Definition: &core.ToolDefinition{
        Name: "calculate",
        Description: "Perform calculations",
    },
    Execute: func(ctx context.Context, params map[string]interface{}) (*core.ToolResult, error) {
        // Tool implementation
    },
})
```

**Features:**
- Tool registration
- Allow/deny lists
- Timeout management
- Result handling

### 3. Decision Engine

Evaluates conditions and makes routing decisions:

```go
engine, _ := registry.GetDecisionEvaluator("default")
decision, _ := engine.Evaluate(ctx, &core.Condition{
    Type: "llm",
    Expression: "Is this urgent?",
    Context: map[string]interface{}{
        "context": "Customer complaint",
    },
})
```

**Features:**
- LLM-based decisions
- Expression evaluation
- Branch selection
- Confidence scoring

### 4. State Manager

Tracks workflow execution state:

```go
manager, _ := registry.GetStateManager("default")
manager.SaveState(ctx, "workflow-123", &core.WorkflowState{
    Status: "running",
    Context: map[string]interface{}{
        "step": "processing",
    },
})
```

**Features:**
- State persistence
- Context accumulation
- History tracking
- Resume capability

### 5. Event System

Publishes and subscribes to events:

```go
publisher, _ := registry.GetEventPublisher("default")
publisher.Publish(ctx, &core.Event{
    Type: "workflow.started",
    Data: map[string]interface{}{
        "workflow_id": "123",
    },
})

subscription, _ := publisher.Subscribe(ctx, &core.EventFilter{
    Types: []string{"workflow.*"},
}, func(ctx context.Context, event *core.Event) error {
    // Handle event
    return nil
})
```

**Features:**
- Event publishing
- Filtered subscriptions
- Async handling
- Type-based routing

### 6. Validation Framework

Validates outputs and enforces quality gates:

```go
validator, _ := registry.GetValidator("default")
result, _ := validator.Validate(ctx, data, &core.ValidationRules{
    Type: "schema",
    Schema: schema,
})
```

**Features:**
- Schema validation
- LLM-based validation
- Quality scoring
- Error reporting

### 7. Human Interaction

Handles human-in-the-loop interactions:

```go
interaction, _ := registry.GetHumanInteraction("default")
response, _ := interaction.RequestApproval(ctx, &core.ApprovalRequest{
    Title: "Approve transaction?",
    Description: "Transaction amount: $1000",
})
```

**Features:**
- Approval requests
- Input prompts
- Timeout handling
- Response management

### 8. Transformation Pipeline

Transforms data between workflow steps:

```go
pipeline, _ := registry.GetTransformer("default")
result, _ := pipeline.Transform(ctx, data, &core.Transformation{
    Type: "map",
    Config: map[string]interface{}{
        "function": "uppercase",
    },
})
```

**Features:**
- Map transformations
- Filter operations
- Reduce operations
- Pipeline composition

## Extension Lifecycle

1. **Registration**: Extension is registered with the registry
2. **Initialization**: `Initialize()` is called with configuration
3. **Active**: Extension is available for use
4. **Shutdown**: `Shutdown()` is called when unloading

## Event Flow

Extensions can publish and subscribe to events:

```
Extension A → Publish Event → Event System → Subscribers (Extension B, C)
```

Events enable loose coupling between extensions.

## State Management

Workflow state is managed centrally:

```
Workflow Step 1 → Update State → State Manager → Save
Workflow Step 2 → Load State → State Manager → Continue
```

## Best Practices

1. **Idempotency**: Extensions should be idempotent where possible
2. **Error Handling**: Always return errors, never panic
3. **Context**: Respect context cancellation
4. **Events**: Use events for cross-extension communication
5. **State**: Use state manager for persistent data
6. **Configuration**: Accept configuration via Initialize()

## Extension Development

### Creating a New Extension

1. Implement the appropriate interface(s)
2. Register with the registry
3. Handle initialization and shutdown
4. Publish events for important actions
5. Use other extensions via the registry

### Example Extension

```go
type MyExtension struct {
    name string
}

func (e *MyExtension) Name() string {
    return e.name
}

func (e *MyExtension) Initialize(ctx context.Context, config map[string]interface{}) error {
    // Initialize
    return nil
}

func (e *MyExtension) Shutdown(ctx context.Context) error {
    // Cleanup
    return nil
}

// Register
registry := core.GetRegistry()
ext := &MyExtension{name: "my-extension"}
registry.Register(ctx, ext, config)
```

## Integration with LlamaGate

Extensions integrate with LlamaGate's existing components:

- **Proxy Layer**: Agent calls use LlamaGate's proxy
- **MCP Client**: Tool framework can use MCP tools
- **Configuration**: Uses LlamaGate's config system
- **Logging**: Uses LlamaGate's logger

## Performance Considerations

- Extensions are loaded on startup
- Event subscriptions are async
- State operations are in-memory (can be extended)
- Tool execution has timeouts

## Security

- Tool allow/deny lists
- Timeout enforcement
- Size limits
- Input validation

## Future Enhancements

- Persistent state storage
- Extension marketplace
- Hot reloading
- Performance metrics
- Distributed events
