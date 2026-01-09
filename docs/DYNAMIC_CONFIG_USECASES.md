# Dynamic Configuration Use Cases

This document provides practical use cases for using LlamaGate's plugin system with dynamic configuration to create flexible, adaptable workflows.

## Overview

Dynamic configuration allows plugins and workflows to adapt their behavior based on:
- Runtime parameters
- Environment variables
- Configuration files
- User input
- Context from previous steps

## Use Case 1: Environment-Aware Plugin Configuration

### Scenario
A plugin needs to behave differently in development, staging, and production environments.

### Solution
Use environment variables to configure plugin behavior dynamically.

**Example Plugin:**

```go
type EnvironmentAwarePlugin struct {
    environment string
}

func (p *EnvironmentAwarePlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    // Get environment from config or environment variable
    env := os.Getenv("ENVIRONMENT")
    if env == "" {
        env = "development" // Default
    }
    
    // Adjust behavior based on environment
    var maxRetries int
    var timeout time.Duration
    
    switch env {
    case "production":
        maxRetries = 3
        timeout = 30 * time.Second
    case "staging":
        maxRetries = 2
        timeout = 20 * time.Second
    default: // development
        maxRetries = 1
        timeout = 10 * time.Second
    }
    
    // Use dynamic configuration in workflow
    workflow := &plugins.Workflow{
        MaxRetries: maxRetries,
        Timeout: timeout,
        Steps: []plugins.WorkflowStep{
            // ... steps configured with dynamic values
        },
    }
    
    // Execute workflow
    // ...
}
```

## Use Case 2: User-Configurable Workflow Parameters

### Scenario
A workflow needs to accept user-defined parameters that affect its execution path.

### Solution
Use input parameters to dynamically configure workflow steps.

**Example: Dynamic Query Processing**

```go
type DynamicQueryPlugin struct {
    executor *plugins.WorkflowExecutor
}

func (p *DynamicQueryPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    // Extract user configuration
    queryType := input["query_type"].(string)
    maxDepth := input["max_depth"].(int)
    useCache := input["use_cache"].(bool)
    
    // Build workflow dynamically based on configuration
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
    
    // Add conditional steps based on configuration
    if useCache {
        steps = append(steps, plugins.WorkflowStep{
            ID:   "check_cache",
            Type: "data_transform",
            Config: map[string]interface{}{
                "transform": "cache_lookup",
            },
            Dependencies: []string{"analyze"},
        })
    }
    
    // Add depth-based steps
    for i := 0; i < maxDepth; i++ {
        steps = append(steps, plugins.WorkflowStep{
            ID:   fmt.Sprintf("process_depth_%d", i),
            Type: "llm_call",
            Config: map[string]interface{}{
                "model": input["model"].(string),
            },
            Dependencies: []string{steps[len(steps)-1].ID},
        })
    }
    
    workflow := &plugins.Workflow{
        Steps: steps,
    }
    
    return p.executeWorkflow(ctx, workflow, input)
}
```

## Use Case 3: Configuration-Driven Tool Selection

### Scenario
A plugin needs to select which tools to use based on configuration.

### Solution
Dynamically build tool calls based on configuration.

**Example: Multi-Tool Plugin**

```go
type MultiToolPlugin struct {
    executor *plugins.WorkflowExecutor
}

func (p *MultiToolPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    // Get enabled tools from configuration
    enabledTools := input["enabled_tools"].([]interface{})
    
    // Build workflow steps for each enabled tool
    steps := []plugins.WorkflowStep{
        {
            ID:   "analyze",
            Type: "llm_call",
            Config: map[string]interface{}{
                "model": "llama3.2",
                "prompt": "Determine which tools are needed",
            },
        },
    }
    
    // Add tool execution steps dynamically
    for i, toolName := range enabledTools {
        stepID := fmt.Sprintf("tool_%d", i)
        steps = append(steps, plugins.WorkflowStep{
            ID:   stepID,
            Type: "tool_call",
            Config: map[string]interface{}{
                "tool_name": toolName.(string),
                "merge_state": true,
            },
            Dependencies: []string{"analyze"},
        })
    }
    
    // Add synthesis step
    steps = append(steps, plugins.WorkflowStep{
        ID:   "synthesize",
        Type: "llm_call",
        Config: map[string]interface{}{
            "model": "llama3.2",
            "prompt": "Synthesize results from all tools",
        },
        Dependencies: func() []string {
            deps := make([]string, len(enabledTools))
            for i := range enabledTools {
                deps[i] = fmt.Sprintf("tool_%d", i)
            }
            return deps
        }(),
    })
    
    workflow := &plugins.Workflow{
        Steps: steps,
    }
    
    return p.executeWorkflow(ctx, workflow, input)
}
```

## Use Case 4: Adaptive Timeout Configuration

### Scenario
A plugin needs to adjust timeouts based on the complexity of the task.

### Solution
Calculate timeouts dynamically based on input parameters.

**Example: Adaptive Timeout Plugin**

```go
type AdaptiveTimeoutPlugin struct {
    executor *plugins.WorkflowExecutor
}

func (p *AdaptiveTimeoutPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    // Calculate timeout based on input complexity
    baseTimeout := 10 * time.Second
    
    // Adjust based on input size
    if text, ok := input["text"].(string); ok {
        textLength := len(text)
        if textLength > 10000 {
            baseTimeout = 60 * time.Second
        } else if textLength > 1000 {
            baseTimeout = 30 * time.Second
        }
    }
    
    // Adjust based on number of steps
    numSteps := input["num_steps"].(int)
    if numSteps > 5 {
        baseTimeout *= 2
    }
    
    // Create workflow with adaptive timeout
    workflow := &plugins.Workflow{
        Timeout: baseTimeout,
        Steps:   p.buildSteps(input),
    }
    
    return p.executeWorkflow(ctx, workflow, input)
}
```

## Use Case 5: Configuration File-Based Plugin Setup

### Scenario
Plugins need to be configured via YAML/JSON configuration files.

### Solution
Load plugin configuration from files and apply to plugins.

**Example Configuration File (`plugin-config.yaml`):**

```yaml
plugins:
  text_summarizer:
    enabled: true
    config:
      default_max_length: 200
      default_style: "brief"
      supported_styles:
        - "brief"
        - "detailed"
        - "bullet"
  
  workflow_example:
    enabled: true
    config:
      default_model: "llama3.2"
      max_retries: 3
      timeout: 30s
      enabled_tools:
        - "mcp.filesystem.read_file"
        - "mcp.fetch.fetch"
```

**Plugin Configuration Loader:**

```go
type PluginConfig struct {
    Plugins map[string]PluginInstanceConfig `yaml:"plugins"`
}

type PluginInstanceConfig struct {
    Enabled bool                   `yaml:"enabled"`
    Config  map[string]interface{} `yaml:"config"`
}

func LoadPluginConfig(path string) (*PluginConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var config PluginConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}

func ConfigurePlugin(plugin plugins.Plugin, config PluginInstanceConfig) {
    // Apply configuration to plugin
    // This would require plugins to support configuration injection
    // or use a factory pattern
}
```

## Use Case 6: Runtime Configuration Updates

### Scenario
Plugin behavior needs to change at runtime without restarting the service.

### Solution
Use a configuration watcher to reload plugin configuration dynamically.

**Example: Dynamic Configuration Watcher**

```go
type ConfigWatcher struct {
    configPath string
    registry   *plugins.Registry
    lastMod    time.Time
}

func (w *ConfigWatcher) Watch(ctx context.Context) error {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if w.shouldReload() {
                if err := w.reloadConfig(); err != nil {
                    log.Error().Err(err).Msg("Failed to reload config")
                }
            }
        }
    }
}

func (w *ConfigWatcher) shouldReload() bool {
    info, err := os.Stat(w.configPath)
    if err != nil {
        return false
    }
    
    if info.ModTime().After(w.lastMod) {
        w.lastMod = info.ModTime()
        return true
    }
    
    return false
}

func (w *ConfigWatcher) reloadConfig() error {
    config, err := LoadPluginConfig(w.configPath)
    if err != nil {
        return err
    }
    
    // Re-register plugins with new configuration
    for name, pluginConfig := range config.Plugins {
        if pluginConfig.Enabled {
            // Recreate plugin with new config
            plugin := createPluginWithConfig(name, pluginConfig)
            w.registry.Unregister(name)
            w.registry.Register(plugin)
        }
    }
    
    return nil
}
```

## Use Case 7: Context-Aware Configuration

### Scenario
Plugin behavior should adapt based on context from previous workflow steps.

### Solution
Use workflow state to dynamically configure subsequent steps.

**Example: Context-Aware Workflow**

```go
func (p *ContextAwarePlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    workflow := &plugins.Workflow{
        Steps: []plugins.WorkflowStep{
            {
                ID:   "analyze_context",
                Type: "llm_call",
                Config: map[string]interface{}{
                    "model": "llama3.2",
                    "prompt": "Analyze the context and determine next steps",
                },
            },
            {
                ID:   "configure_next",
                Type: "data_transform",
                Config: map[string]interface{}{
                    "transform": "extract",
                    "input_key": "llm_response",
                    "fields": []interface{}{"action", "complexity"},
                },
                Dependencies: []string{"analyze_context"},
            },
            {
                ID:   "execute_adaptive",
                Type: "llm_call",
                Config: map[string]interface{}{
                    // Configuration will be set dynamically based on previous step
                    "model": "llama3.2",
                },
                Dependencies: []string{"configure_next"},
            },
        },
    }
    
    // Custom executor that adapts step config based on state
    executor := NewAdaptiveWorkflowExecutor(p.llmHandler, p.toolHandler)
    results, err := executor.Execute(ctx, workflow, input)
    
    // ...
}
```

## Use Case 8: Multi-Tenant Configuration

### Scenario
Different users/tenants need different plugin configurations.

### Solution
Use tenant-specific configuration that's loaded per request.

**Example: Tenant-Aware Plugin**

```go
type TenantAwarePlugin struct {
    configLoader func(tenantID string) map[string]interface{}
}

func (p *TenantAwarePlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    // Get tenant ID from context or input
    tenantID := input["tenant_id"].(string)
    
    // Load tenant-specific configuration
    tenantConfig := p.configLoader(tenantID)
    
    // Merge with input
    for k, v := range tenantConfig {
        if _, exists := input[k]; !exists {
            input[k] = v
        }
    }
    
    // Execute with tenant-specific config
    // ...
}
```

## Best Practices

### 1. Configuration Validation
Always validate dynamic configuration before use:

```go
func validateConfig(config map[string]interface{}) error {
    if timeout, ok := config["timeout"].(time.Duration); ok {
        if timeout < 1*time.Second || timeout > 5*time.Minute {
            return fmt.Errorf("timeout must be between 1s and 5m")
        }
    }
    return nil
}
```

### 2. Default Values
Always provide sensible defaults:

```go
timeout := 30 * time.Second // Default
if t, ok := config["timeout"].(time.Duration); ok {
    timeout = t
}
```

### 3. Configuration Caching
Cache configuration to avoid repeated file reads:

```go
type ConfigCache struct {
    config map[string]interface{}
    mu     sync.RWMutex
    ttl    time.Duration
    lastUpdate time.Time
}

func (c *ConfigCache) Get() map[string]interface{} {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.config
}
```

### 4. Error Handling
Handle configuration errors gracefully:

```go
config, err := loadConfig()
if err != nil {
    log.Warn().Err(err).Msg("Failed to load config, using defaults")
    config = getDefaultConfig()
}
```

## Summary

Dynamic configuration enables:
- ✅ Environment-aware behavior
- ✅ User-customizable workflows
- ✅ Runtime configuration updates
- ✅ Context-aware adaptation
- ✅ Multi-tenant support
- ✅ Flexible tool selection

Use these patterns to create adaptable, flexible plugins that can adjust their behavior based on configuration and context.
