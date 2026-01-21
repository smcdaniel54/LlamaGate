// Package core provides core interfaces and types for the extension foundation system.
package core

import (
	"context"
)

// Extension is the base interface that all extensions must implement
type Extension interface {
	// Name returns the unique name of the extension
	Name() string
	
	// Version returns the version of the extension
	Version() string
	
	// Initialize is called when the extension is registered
	Initialize(ctx context.Context, config map[string]interface{}) error
	
	// Shutdown is called when the extension is being unloaded
	Shutdown(ctx context.Context) error
}

// AgentCaller defines the interface for making agent/LLM calls
type AgentCaller interface {
	Extension
	
	// Call makes a synchronous agent call
	Call(ctx context.Context, req *AgentRequest) (*AgentResponse, error)
	
	// CallStream makes a streaming agent call
	CallStream(ctx context.Context, req *AgentRequest) (<-chan *StreamChunk, error)
}

// ToolExecutor defines the interface for executing tools
type ToolExecutor interface {
	Extension
	
	// Execute executes a tool with the given parameters
	Execute(ctx context.Context, toolName string, params map[string]interface{}) (*ToolResult, error)
	
	// ListTools returns all available tools
	ListTools(ctx context.Context) ([]*ToolDefinition, error)
}

// DecisionEvaluator defines the interface for evaluating decisions
type DecisionEvaluator interface {
	Extension
	
	// Evaluate evaluates a condition and returns a decision
	Evaluate(ctx context.Context, condition *Condition) (*Decision, error)
	
	// EvaluateBranch evaluates multiple branches and selects one
	EvaluateBranch(ctx context.Context, branches []*Branch) (*Branch, error)
}

// StateManager defines the interface for managing workflow state
type StateManager interface {
	Extension
	
	// SaveState saves workflow state
	SaveState(ctx context.Context, workflowID string, state *WorkflowState) error
	
	// LoadState loads workflow state
	LoadState(ctx context.Context, workflowID string) (*WorkflowState, error)
	
	// UpdateContext updates the context for a workflow
	UpdateContext(ctx context.Context, workflowID string, updates map[string]interface{}) error
}

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	Extension
	
	// Publish publishes an event
	Publish(ctx context.Context, event *Event) error
	
	// Subscribe subscribes to events matching the filter
	Subscribe(ctx context.Context, filter *EventFilter, handler EventHandler) (Subscription, error)
}

// Validator defines the interface for validating outputs
type Validator interface {
	Extension
	
	// Validate validates data against rules
	Validate(ctx context.Context, data interface{}, rules *ValidationRules) (*ValidationResult, error)
}

// HumanInteraction defines the interface for human-in-the-loop interactions
type HumanInteraction interface {
	Extension
	
	// RequestApproval requests approval from a human
	RequestApproval(ctx context.Context, request *ApprovalRequest) (*ApprovalResponse, error)
	
	// WaitForInput waits for human input
	WaitForInput(ctx context.Context, prompt *InputPrompt) (*InputResponse, error)
}

// Transformer defines the interface for data transformation
type Transformer interface {
	Extension
	
	// Transform transforms data using the specified transformation
	Transform(ctx context.Context, data interface{}, transformation *Transformation) (interface{}, error)
}

// EventHandler is a function type for handling events
type EventHandler func(ctx context.Context, event *Event) error

// Subscription represents an event subscription
type Subscription interface {
	// Unsubscribe unsubscribes from events
	Unsubscribe(ctx context.Context) error
}
