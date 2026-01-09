# Plugin System Simplicity Guide

This document explains how simple and lightweight the plugin system is to use.

## Is It Simple? Yes! ✅

### Minimal Plugin (3 Methods, ~50 Lines)

The simplest plugin only needs 3 methods:

```go
type MyPlugin struct{}

// 1. Metadata - What is this plugin?
func (p *MyPlugin) Metadata() plugins.PluginMetadata {
    return plugins.PluginMetadata{
        Name:        "my_plugin",
        Description: "Does something useful",
        RequiredInputs: []string{"input"},
    }
}

// 2. ValidateInput - Check inputs (optional)
func (p *MyPlugin) ValidateInput(input map[string]interface{}) error {
    if _, exists := input["input"]; !exists {
        return fmt.Errorf("input is required")
    }
    return nil
}

// 3. Execute - Do the work
func (p *MyPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    // Your logic here
    return &plugins.PluginResult{
        Success: true,
        Data: map[string]interface{}{"result": "done"},
    }, nil
}
```

**That's it!** Just 3 methods, ~50 lines of code.

## Is It Lightweight? Yes! ✅

### Comparison

| Aspect | Simple Plugin | Full Plugin |
|--------|---------------|-------------|
| **Methods** | 3 | 3 |
| **Lines of Code** | ~50 | ~150 |
| **Required Knowledge** | Basic Go | Go + JSON Schema |
| **Time to Create** | 5 minutes | 15-30 minutes |

### What Makes It Lightweight?

1. **No Required Schemas**: Schemas are optional for documentation
2. **Simple Types**: Just `map[string]interface{}` for inputs/outputs
3. **No Dependencies**: Core plugin has zero external dependencies
4. **Optional Features**: Workflows, agents, custom endpoints are optional

### Minimal Example

```go
// Simplest possible plugin
type EchoPlugin struct{}

func (p *EchoPlugin) Metadata() plugins.PluginMetadata {
    return plugins.PluginMetadata{
        Name: "echo",
        RequiredInputs: []string{"message"},
    }
}

func (p *EchoPlugin) ValidateInput(input map[string]interface{}) error {
    if _, ok := input["message"]; !ok {
        return fmt.Errorf("message required")
    }
    return nil
}

func (p *EchoPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    return &plugins.PluginResult{
        Success: true,
        Data: map[string]interface{}{"echo": input["message"]},
    }, nil
}
```

**Only 25 lines!**

## Is It Easy to Understand? Yes! ✅

### Three Simple Concepts

1. **Metadata**: What your plugin is
2. **Validation**: Check inputs are OK
3. **Execution**: Do the work

### Clear Flow

```
User calls plugin
    ↓
ValidateInput() checks inputs
    ↓
Execute() does the work
    ↓
Returns result
```

### No Complex Concepts

- ❌ No dependency injection
- ❌ No complex configuration
- ❌ No framework magic
- ❌ No code generation
- ✅ Just 3 methods
- ✅ Plain Go code
- ✅ Standard types

## Progressive Complexity

The system supports simple to complex use cases:

### Level 1: Simple (5 minutes)
```go
// Just 3 methods, basic validation
type SimplePlugin struct{}
func (p *SimplePlugin) Metadata() ...
func (p *SimplePlugin) ValidateInput() ...
func (p *SimplePlugin) Execute() ...
```

### Level 2: With Schemas (10 minutes)
```go
// Add JSON Schema for documentation
InputSchema: map[string]interface{}{...}
OutputSchema: map[string]interface{}{...}
```

### Level 3: With Workflows (20 minutes)
```go
// Add workflow executor for multi-step processing
executor := plugins.NewWorkflowExecutor(...)
workflow := &plugins.Workflow{Steps: [...]}
```

### Level 4: Advanced (30+ minutes)
```go
// Custom endpoints, agents, etc.
func (p *Plugin) GetAPIEndpoints() ...
func (p *Plugin) GetAgentDefinition() ...
```

## Quick Start (2 Minutes)

1. **Copy template**:
   ```bash
   cp plugins/templates/simple_plugin.go my_plugin.go
   ```

2. **Edit 3 methods**:
   - Change name in `Metadata()`
   - Add validation in `ValidateInput()`
   - Add logic in `Execute()`

3. **Register**:
   ```go
   registry.Register(&MyPlugin{})
   ```

4. **Done!** Use via API:
   ```bash
   POST /v1/plugins/my_plugin/execute
   {"input": "hello"}
   ```

## Comparison with Other Systems

| Feature | LlamaGate | Other Systems |
|---------|-----------|---------------|
| **Lines for Basic Plugin** | ~50 | 200-500+ |
| **Required Methods** | 3 | 5-10+ |
| **Configuration Files** | None | Often required |
| **Learning Curve** | Low | Medium-High |
| **Boilerplate** | Minimal | Significant |

## Real Examples

### Example 1: Echo Plugin (25 lines)
```go
type EchoPlugin struct{}
func (p *EchoPlugin) Metadata() plugins.PluginMetadata {
    return plugins.PluginMetadata{Name: "echo", RequiredInputs: []string{"message"}}
}
func (p *EchoPlugin) ValidateInput(input map[string]interface{}) error {
    if _, ok := input["message"]; !ok { return fmt.Errorf("message required") }
    return nil
}
func (p *EchoPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    return &plugins.PluginResult{Success: true, Data: map[string]interface{}{"echo": input["message"]}}, nil
}
```

### Example 2: Calculator Plugin (40 lines)
```go
type CalculatorPlugin struct{}
func (p *CalculatorPlugin) Metadata() plugins.PluginMetadata {
    return plugins.PluginMetadata{
        Name: "calculator",
        RequiredInputs: []string{"a", "b", "op"},
    }
}
func (p *CalculatorPlugin) ValidateInput(input map[string]interface{}) error {
    // Validate a, b are numbers, op is +, -, *, /
    return nil
}
func (p *CalculatorPlugin) Execute(ctx context.Context, input map[string]interface{}) (*plugins.PluginResult, error) {
    a := input["a"].(float64)
    b := input["b"].(float64)
    op := input["op"].(string)
    // Calculate result
    return &plugins.PluginResult{Success: true, Data: map[string]interface{}{"result": result}}, nil
}
```

## Summary

### ✅ Simple to Use
- Only 3 methods required
- Clear, straightforward interface
- No magic or hidden behavior

### ✅ Lightweight to Define
- ~50 lines for basic plugin
- ~25 lines for minimal plugin
- No required configuration files
- No code generation needed

### ✅ Easy to Understand
- Three clear concepts (metadata, validation, execution)
- Plain Go code
- Standard types
- Progressive complexity

**The plugin system is designed to be as simple as possible while still being powerful enough for complex use cases.**
