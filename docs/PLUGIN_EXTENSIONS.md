# Plugin System Extensions

This document describes the extended capabilities of the LlamaGate plugin system, including custom API endpoints, agent definitions, and advanced agentic workflows.

## Overview

The plugin system has been extended to support:

1. **Custom API Endpoints**: Plugins can define their own HTTP endpoints
2. **Agent Definitions**: Structured agent configurations with capabilities
3. **Advanced Agentic Workflows**: Enhanced workflow patterns for agents

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
        {
            Path:        "/status",
            Method:      "GET",
            Description: "Get plugin status",
            Handler:     p.handleStatus,
            RequiresAuth: false,
        },
    }
}

func (p *MyPlugin) handleCustomAction(c *gin.Context) {
    // Handle the custom endpoint
    var input map[string]interface{}
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // Execute plugin logic
    result, err := p.Execute(c.Request.Context(), input)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, result)
}
```

### Endpoint Registration

Endpoints are automatically registered at:
- `GET /v1/plugins/{plugin_name}/custom/action`
- `GET /v1/plugins/{plugin_name}/status`

### Endpoint Configuration

- **Path**: Relative path (will be prefixed with `/v1/plugins/{plugin_name}`)
- **Method**: HTTP method (GET, POST, PUT, DELETE, PATCH)
- **Handler**: Gin handler function
- **RequestSchema**: JSON Schema for request validation
- **ResponseSchema**: JSON Schema for response documentation
- **RequiresAuth**: Whether authentication is required
- **RequiresRateLimit**: Whether rate limiting applies

## Agent Definitions

Agents are specialized plugins with structured capabilities and workflows.

### Defining an Agent

```go
type MyAgent struct {
    definition *plugins.AgentDefinition
    executor   *plugins.AgentExecutor
}

func NewMyAgent() *MyAgent {
    definition := &plugins.AgentDefinition{
        Name:        "my_agent",
        Description: "An intelligent agent for specific tasks",
        Capabilities: []string{
            "text_analysis",
            "data_processing",
            "decision_making",
        },
        Tools: []string{
            "mcp.filesystem.read_file",
            "mcp.fetch.fetch",
        },
        Model:         "llama3.2",
        SystemPrompt:  "You are a helpful assistant...",
        MaxIterations: 10,
        Temperature:   0.7,
        Workflow: &plugins.Workflow{
            Steps: []plugins.WorkflowStep{
                // Define agent workflow
            },
        },
    }
    
    executor := plugins.NewAgentExecutor(
        definition,
        llmHandler,
        toolHandler,
    )
    
    return &MyAgent{
        definition: definition,
        executor:   executor,
    }
}

func (a *MyAgent) GetDefinition() *plugins.AgentDefinition {
    return a.definition
}

func (a *MyAgent) Execute(ctx context.Context, input map[string]interface{}, config map[string]interface{}) (*plugins.AgentResult, error) {
    stepResults, err := a.executor.ExecuteAgent(ctx, input, config)
    // Process results and return AgentResult
}
```

### Agent Definition Fields

- **Name**: Unique agent identifier
- **Description**: What the agent does
- **Capabilities**: List of agent capabilities
- **Tools**: Available tools for the agent
- **Model**: Default LLM model
- **SystemPrompt**: System prompt for the agent
- **MaxIterations**: Maximum iterations for agentic loops
- **Temperature**: Default temperature for LLM calls
- **Workflow**: Default workflow for the agent
- **ConfigSchema**: Configuration schema (JSON Schema)

### Agent Execution

```go
agent := NewMyAgent()
result, err := agent.Execute(ctx, map[string]interface{}{
    "query": "Process this request",
}, map[string]interface{}{
    "temperature": 0.8,
    "max_iterations": 5,
})
```

## Advanced Agentic Workflows

### Iterative Agent Loops

Agents can perform iterative reasoning:

```go
workflow := &plugins.Workflow{
    Steps: []plugins.WorkflowStep{
        {
            ID:   "analyze",
            Type: "llm_call",
            Config: map[string]interface{}{
                "model": "llama3.2",
                "prompt": "Analyze the problem",
            },
        },
        {
            ID:   "reason",
            Type: "llm_call",
            Config: map[string]interface{}{
                "model": "llama3.2",
                "prompt": "Reason about the solution",
            },
            Dependencies: []string{"analyze"},
        },
        {
            ID:   "execute",
            Type: "tool_call",
            Config: map[string]interface{}{
                "tool_name": "mcp.filesystem.read_file",
            },
            Dependencies: []string{"reason"},
        },
        {
            ID:   "synthesize",
            Type: "llm_call",
            Config: map[string]interface{}{
                "model": "llama3.2",
                "prompt": "Synthesize the final answer",
            },
            Dependencies: []string{"execute"},
        },
    },
    MaxRetries: 3,
    Timeout:    60 * time.Second,
}
```

### Conditional Workflows

Use conditions to create branching workflows:

```go
workflow := &plugins.Workflow{
    Steps: []plugins.WorkflowStep{
        {
            ID:   "check_condition",
            Type: "condition",
            Config: map[string]interface{}{
                "condition": "needs_tool",
            },
        },
        {
            ID:   "use_tool",
            Type: "tool_call",
            Config: map[string]interface{}{
                "tool_name": "mcp.filesystem.read_file",
            },
            Dependencies: []string{"check_condition"},
            OnError: "continue", // Continue even if tool fails
        },
        {
            ID:   "finalize",
            Type: "llm_call",
            Config: map[string]interface{}{
                "model": "llama3.2",
            },
            Dependencies: []string{"use_tool"},
        },
    },
}
```

## Use Cases

### Use Case 1: Plugin with Custom API

A plugin that exposes a specialized endpoint:

```go
func (p *DataProcessorPlugin) GetAPIEndpoints() []plugins.APIEndpoint {
    return []plugins.APIEndpoint{
        {
            Path:        "/process",
            Method:      "POST",
            Description: "Process data with custom logic",
            Handler:     p.handleProcess,
            RequestSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "data": map[string]interface{}{
                        "type": "string",
                    },
                },
            },
        },
    }
}
```

### Use Case 2: Agent with Workflow

An agent that uses a complex workflow:

```go
agent := &plugins.AgentDefinition{
    Name: "research_agent",
    Capabilities: []string{"research", "analysis", "synthesis"},
    Tools: []string{"mcp.fetch.fetch", "mcp.filesystem.read_file"},
    Model: "llama3.2",
    Workflow: &plugins.Workflow{
        Steps: []plugins.WorkflowStep{
            // Research workflow steps
        },
    },
}
```

### Use Case 3: Multi-Step Agentic Process

An agent that performs iterative reasoning:

```go
workflow := &plugins.Workflow{
    Steps: []plugins.WorkflowStep{
        {ID: "understand", Type: "llm_call", ...},
        {ID: "plan", Type: "llm_call", Dependencies: []string{"understand"}},
        {ID: "execute", Type: "tool_call", Dependencies: []string{"plan"}},
        {ID: "reflect", Type: "llm_call", Dependencies: []string{"execute"}},
        {ID: "refine", Type: "llm_call", Dependencies: []string{"reflect"}},
    },
}
```

## Integration

### Registering Extended Plugins

```go
registry := plugins.NewRegistry()

// Register plugin (automatically detects ExtendedPlugin)
plugin := NewMyExtendedPlugin()
registry.Register(plugin)

// Register routes (in main.go)
api.RegisterPluginRoutes(v1, registry)
```

### Registering Agents

```go
agentRegistry := plugins.NewAgentRegistry()
agent := NewMyAgent()
agentRegistry.Register(agent)
```

## Best Practices

1. **API Endpoints**
   - Use descriptive paths
   - Document request/response schemas
   - Set appropriate auth/rate limit flags

2. **Agent Definitions**
   - Clearly define capabilities
   - List all available tools
   - Provide meaningful system prompts

3. **Workflows**
   - Keep steps focused
   - Use dependencies correctly
   - Handle errors appropriately

## Summary

The extended plugin system enables:

- ✅ **Custom API Endpoints**: Plugins can expose specialized endpoints
- ✅ **Agent Definitions**: Structured agent configurations
- ✅ **Advanced Workflows**: Complex agentic processing patterns
- ✅ **Flexible Integration**: Easy to extend and customize

All use cases from the dynamic configuration guide can now be implemented as plugins with custom endpoints, agent definitions, and advanced workflows.
