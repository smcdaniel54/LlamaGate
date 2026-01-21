package tools

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/llamagate/llamagate/internal/extensions/builtin/core"
)

// Framework manages tool registration and execution
type Framework struct {
	name      string
	version   string
	mu        sync.RWMutex
	tools     map[string]*ToolHandler
	registry  *core.Registry
	timeout   time.Duration
	maxSize   int64
	allowList []string
	denyList  []string
}

// ToolHandler defines how a tool is executed
type ToolHandler struct {
	Definition *core.ToolDefinition
	Execute    func(ctx context.Context, params map[string]interface{}) (*core.ToolResult, error)
}

// NewFramework creates a new tool framework
func NewFramework(name, version string) *Framework {
	return &Framework{
		name:     name,
		version:  version,
		tools:    make(map[string]*ToolHandler),
		registry: core.GetRegistry(),
		timeout:  30 * time.Second,
		maxSize:  10 * 1024 * 1024, // 10MB default
	}
}

// Name returns the name of the extension
func (f *Framework) Name() string {
	return f.name
}

// Version returns the version of the extension
func (f *Framework) Version() string {
	return f.version
}

// Initialize initializes the framework
func (f *Framework) Initialize(ctx context.Context, config map[string]interface{}) error {
	if timeout, ok := config["timeout"].(string); ok {
		if d, err := time.ParseDuration(timeout); err == nil {
			f.timeout = d
		}
	}
	if maxSize, ok := config["max_size"].(int64); ok {
		f.maxSize = maxSize
	}
	if allowList, ok := config["allow_list"].([]interface{}); ok {
		f.allowList = make([]string, len(allowList))
		for i, v := range allowList {
			if s, ok := v.(string); ok {
				f.allowList[i] = s
			}
		}
	}
	if denyList, ok := config["deny_list"].([]interface{}); ok {
		f.denyList = make([]string, len(denyList))
		for i, v := range denyList {
			if s, ok := v.(string); ok {
				f.denyList[i] = s
			}
		}
	}
	return nil
}

// Shutdown shuts down the framework
func (f *Framework) Shutdown(ctx context.Context) error {
	return nil
}

// RegisterTool registers a tool
func (f *Framework) RegisterTool(handler *ToolHandler) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if handler.Definition == nil {
		return fmt.Errorf("tool definition is required")
	}
	if handler.Execute == nil {
		return fmt.Errorf("tool execute function is required")
	}

	name := handler.Definition.Name
	if _, exists := f.tools[name]; exists {
		return fmt.Errorf("tool %s already registered", name)
	}

	f.tools[name] = handler
	return nil
}

// UnregisterTool unregisters a tool
func (f *Framework) UnregisterTool(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.tools, name)
}

// Execute executes a tool with the given parameters
func (f *Framework) Execute(ctx context.Context, toolName string, params map[string]interface{}) (*core.ToolResult, error) {
	// Check allow/deny lists
	if !f.isAllowed(toolName) {
		return &core.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("tool %s is not allowed", toolName),
		}, fmt.Errorf("tool %s is not allowed", toolName)
	}

	// Get tool handler
	f.mu.RLock()
	handler, exists := f.tools[toolName]
	f.mu.RUnlock()

	if !exists {
		return &core.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("tool %s not found", toolName),
		}, fmt.Errorf("tool %s not found", toolName)
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	// Publish event
	if publisher := f.getEventPublisher(); publisher != nil {
		_ = publisher.Publish(ctx, &core.Event{
			Type:      "tool.execution.started",
			Source:    f.name,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"tool": toolName,
			},
		})
	}

	// Execute tool
	start := time.Now()
	result, err := handler.Execute(ctx, params)
	duration := time.Since(start)

	if result == nil {
		result = &core.ToolResult{}
	}
	result.Duration = duration

	if err != nil {
		result.Success = false
		result.Error = err.Error()

		// Publish error event
		if publisher := f.getEventPublisher(); publisher != nil {
			_ = publisher.Publish(ctx, &core.Event{
				Type:      "tool.execution.failed",
				Source:    f.name,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"tool":  toolName,
					"error": err.Error(),
				},
			})
		}
		return result, err
	}

	result.Success = true

	// Publish success event
	if publisher := f.getEventPublisher(); publisher != nil {
		_ = publisher.Publish(ctx, &core.Event{
			Type:      "tool.execution.completed",
			Source:    f.name,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"tool":     toolName,
				"duration": duration.String(),
			},
		})
	}

	return result, nil
}

// ListTools returns all available tools
func (f *Framework) ListTools(_ context.Context) ([]*core.ToolDefinition, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	tools := make([]*core.ToolDefinition, 0, len(f.tools))
	for _, handler := range f.tools {
		// Filter by allow/deny lists
		if f.isAllowed(handler.Definition.Name) {
			tools = append(tools, handler.Definition)
		}
	}

	return tools, nil
}

// isAllowed checks if a tool is allowed based on allow/deny lists
func (f *Framework) isAllowed(toolName string) bool {
	// Check deny list first
	for _, denied := range f.denyList {
		if denied == toolName || denied == "*" {
			return false
		}
	}

	// If allow list is empty, allow all (except denied)
	if len(f.allowList) == 0 {
		return true
	}

	// Check allow list
	for _, allowed := range f.allowList {
		if allowed == toolName || allowed == "*" {
			return true
		}
	}

	return false
}

// getEventPublisher gets the event publisher if available
func (f *Framework) getEventPublisher() core.EventPublisher {
	if f.registry == nil {
		return nil
	}
	publisher, err := f.registry.GetEventPublisher("default")
	if err != nil {
		return nil
	}
	return publisher
}
