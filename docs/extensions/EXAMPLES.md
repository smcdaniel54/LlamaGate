# Extension Examples

This document provides detailed examples of using the Extension Foundation system.

## Example 1: Logging Extension

A simple extension that logs all events:

```go
package logging

import (
    "context"
    "log"
    "github.com/smcdaniel54/LlamaGate/internal/extensions/core"
)

type Extension struct {
    name string
}

func (e *Extension) Name() string {
    return e.name
}

func (e *Extension) Initialize(ctx context.Context, config map[string]interface{}) error {
    registry := core.GetRegistry()
    publisher, _ := registry.GetEventPublisher("default")
    
    filter := &core.EventFilter{Types: []string{"*"}}
    publisher.Subscribe(ctx, filter, func(ctx context.Context, event *core.Event) error {
        log.Printf("Event: %s from %s", event.Type, event.Source)
        return nil
    })
    
    return nil
}
```

## Example 2: Text Classification Workflow (Local Model + MCP)

A complete workflow that classifies text using local Ollama model and routes using MCP tools:

```go
package main

import (
    "context"
    "github.com/smcdaniel54/LlamaGate/internal/extensions/core"
    "github.com/smcdaniel54/LlamaGate/internal/extensions/agent"
    "github.com/smcdaniel54/LlamaGate/internal/extensions/decision"
    "github.com/smcdaniel54/LlamaGate/internal/extensions/state"
    "github.com/smcdaniel54/LlamaGate/internal/extensions/tools"
)

func classifyAndRoute(ctx context.Context, text string) error {
    registry := core.GetRegistry()
    
    // Get extensions
    caller, _ := registry.GetAgentCaller("default")  // Uses local Ollama
    engine, _ := registry.GetDecisionEvaluator("default")
    manager, _ := registry.GetStateManager("default")
    toolFramework, _ := registry.GetToolExecutor("default")  // MCP tools
    
    workflowID := "classification-" + generateID()
    
    // Step 1: Classify text using LOCAL MODEL (Ollama)
    response, err := caller.Call(ctx, &core.AgentRequest{
        Model: "mistral",  // Local Ollama model
        Messages: []core.Message{
            {Role: "user", Content: "Classify this text: " + text},
        },
        Tools: []*core.ToolDefinition{
            // MCP tools available for routing
            {Name: "mcp://routing/escalate", Description: "Escalate to urgent queue"},
            {Name: "mcp://routing/standard", Description: "Route to standard queue"},
        },
    })
    if err != nil {
        return err
    }
    
    classification := response.Content
    
    // Step 2: Route based on classification using LOCAL MODEL decision
    branch, err := engine.EvaluateBranch(ctx, []*core.Branch{
        {
            ID:   "urgent",
            Name: "Urgent",
            Priority: 10,
            Condition: &core.Condition{
                Type:       "llm",  // Uses local Ollama model
                Expression: "Is this urgent?",
                Context: map[string]interface{}{
                    "classification": classification,
                },
            },
        },
        {
            ID:   "normal",
            Name: "Normal",
            Priority: 5,
        },
    })
    if err != nil {
        return err
    }
    
    // Step 3: Save state
    manager.SaveState(ctx, workflowID, &core.WorkflowState{
        Status: "completed",
        Context: map[string]interface{}{
            "text":           text,
            "classification": classification,
            "route":          branch.ID,
        },
    })
    
    // Step 4: Handle routing using MCP TOOLS
    switch branch.ID {
    case "urgent":
        // Use MCP escalation tool
        toolFramework.Execute(ctx, "mcp://routing/escalate", map[string]interface{}{
            "text": text,
            "classification": classification,
        })
    case "normal":
        // Use MCP standard routing tool
        toolFramework.Execute(ctx, "mcp://routing/standard", map[string]interface{}{
            "text": text,
            "classification": classification,
        })
    }
    
    return nil
}
```

## Example 3: Database Query Workflow (Local Model + MCP)

A workflow that generates SQL using local Ollama model and executes using MCP database tools:

```go
func queryDatabase(ctx context.Context, question string) (interface{}, error) {
    registry := core.GetRegistry()
    
    caller, _ := registry.GetAgentCaller("default")  // Local Ollama
    framework, _ := registry.GetToolExecutor("default")  // MCP tools
    validator, _ := registry.GetValidator("default")
    
    // Step 1: Generate SQL query using LOCAL MODEL (Ollama)
    response, err := caller.Call(ctx, &core.AgentRequest{
        Model: "mistral",  // Local Ollama model
        Messages: []core.Message{
            {Role: "user", Content: "Generate SQL query for: " + question},
        },
        Tools: []*core.ToolDefinition{
            {
                Name:        "mcp://database/execute",
                Description: "Execute a SQL query via MCP",
            },
        },
    })
    if err != nil {
        return nil, err
    }
    
    // Step 2: Extract SQL from response
    sqlQuery := extractSQL(response.Content)
    
    // Step 3: Validate SQL using LOCAL MODEL
    validation, err := validator.Validate(ctx, sqlQuery, &core.ValidationRules{
        Type: "llm",  // Uses local Ollama model for validation
        Rules: []string{"no_drop", "no_delete", "read_only"},
    })
    if err != nil || !validation.Valid {
        return nil, fmt.Errorf("SQL validation failed: %v", validation.Errors)
    }
    
    // Step 4: Execute query using MCP DATABASE TOOL
    result, err := framework.Execute(ctx, "mcp://database/execute", map[string]interface{}{
        "query": sqlQuery,
    })
    if err != nil {
        return nil, err
    }
    
    // Step 5: Format results using MCP FORMATTING TOOL
    formatted, err := framework.Execute(ctx, "mcp://formatting/json", map[string]interface{}{
        "data": result.Output,
    })
    if err != nil {
        return nil, err
    }
    
    return formatted.Output, nil
}
```

## Example 4: Human Approval Workflow

A workflow that requires human approval:

```go
func processTransaction(ctx context.Context, amount float64) error {
    registry := core.GetRegistry()
    
    interaction, _ := registry.GetHumanInteraction("default")
    manager, _ := registry.GetStateManager("default")
    
    workflowID := "transaction-" + generateID()
    
    // Save initial state
    manager.SaveState(ctx, workflowID, &core.WorkflowState{
        Status: "pending_approval",
        Context: map[string]interface{}{
            "amount": amount,
        },
    })
    
    // Request approval if amount is high
    if amount > 1000 {
        response, err := interaction.RequestApproval(ctx, &core.ApprovalRequest{
            Title:       "Approve Transaction",
            Description: fmt.Sprintf("Transaction amount: $%.2f", amount),
            Options:     []string{"approve", "reject"},
            Timeout:     func() *time.Duration { d := 1 * time.Hour; return &d }(),
        })
        if err != nil {
            return err
        }
        
        if !response.Approved {
            manager.UpdateStatus(ctx, workflowID, "rejected")
            return fmt.Errorf("transaction rejected")
        }
    }
    
    // Process transaction
    manager.UpdateStatus(ctx, workflowID, "processing")
    // ... process transaction ...
    manager.UpdateStatus(ctx, workflowID, "completed")
    
    return nil
}
```

## Example 5: Multi-Step Plan Generation

A workflow that generates and executes a multi-step plan:

```go
func executePlan(ctx context.Context, goal string) error {
    registry := core.GetRegistry()
    
    caller, _ := registry.GetAgentCaller("default")
    manager, _ := registry.GetStateManager("default")
    framework, _ := registry.GetToolExecutor("default")
    
    workflowID := "plan-" + generateID()
    
    // Step 1: Generate plan
    response, err := caller.Call(ctx, &core.AgentRequest{
        Model: "mistral",
        Messages: []core.Message{
            {Role: "user", Content: "Create a step-by-step plan for: " + goal},
        },
    })
    if err != nil {
        return err
    }
    
    // Step 2: Parse plan into steps
    steps := parsePlan(response.Content)
    
    // Step 3: Execute each step
    for i, step := range steps {
        manager.UpdateContext(ctx, workflowID, map[string]interface{}{
            "current_step": i,
            "total_steps": len(steps),
        })
        
        // Execute step
        result, err := framework.Execute(ctx, step.Tool, step.Params)
        if err != nil {
            manager.UpdateStatus(ctx, workflowID, "failed")
            return err
        }
        
        // Update context with result
        manager.UpdateContext(ctx, workflowID, map[string]interface{}{
            fmt.Sprintf("step_%d_result", i): result.Output,
        })
    }
    
    manager.UpdateStatus(ctx, workflowID, "completed")
    return nil
}
```

## Example 6: Event-Driven Workflow

A workflow that reacts to events:

```go
func setupEventDrivenWorkflow(ctx context.Context) error {
    registry := core.GetRegistry()
    
    publisher, _ := registry.GetEventPublisher("default")
    manager, _ := registry.GetStateManager("default")
    
    // Subscribe to workflow events
    publisher.Subscribe(ctx, &core.EventFilter{
        Types: []string{"workflow.started", "workflow.completed"},
    }, func(ctx context.Context, event *core.Event) error {
        workflowID, _ := event.Data["workflow_id"].(string)
        
        switch event.Type {
        case "workflow.started":
            manager.SaveState(ctx, workflowID, &core.WorkflowState{
                Status: "running",
            })
        case "workflow.completed":
            manager.UpdateStatus(ctx, workflowID, "completed")
        }
        
        return nil
    })
    
    return nil
}
```

## Example 7: Data Transformation Pipeline

A workflow that transforms data through multiple steps:

```go
func transformData(ctx context.Context, data interface{}) (interface{}, error) {
    registry := core.GetRegistry()
    
    pipeline, _ := registry.GetTransformer("default")
    
    // Apply multiple transformations
    transformations := []*core.Transformation{
        {
            Type: "map",
            Config: map[string]interface{}{
                "function": "uppercase",
            },
        },
        {
            Type: "filter",
            Config: map[string]interface{}{
                "condition": "length > 10",
            },
        },
        {
            Type: "reduce",
            Config: map[string]interface{}{
                "function": "sum",
            },
        },
    }
    
    result, err := pipeline.TransformMany(ctx, data, transformations)
    if err != nil {
        return nil, err
    }
    
    return result, nil
}
```

## Best Practices

1. **Error Handling**: Always check errors and handle them appropriately
2. **Context**: Respect context cancellation
3. **State**: Save state at important checkpoints
4. **Events**: Publish events for important actions
5. **Validation**: Validate inputs and outputs
6. **Timeouts**: Set appropriate timeouts for operations

## Next Steps

- Explore the API documentation
- Build your own extensions
- Integrate with existing systems
- Share your examples
