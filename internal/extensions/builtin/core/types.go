package core

import (
	"time"
)

// AgentRequest represents a request to an agent/LLM
type AgentRequest struct {
	Model       string                 `json:"model"`
	Messages    []Message              `json:"messages"`
	Tools       []*ToolDefinition      `json:"tools,omitempty"`
	Temperature *float64               `json:"temperature,omitempty"`
	MaxTokens   *int                   `json:"max_tokens,omitempty"`
	Stream      bool                   `json:"stream"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AgentResponse represents a response from an agent/LLM
type AgentResponse struct {
	ID        string                 `json:"id"`
	Model     string                 `json:"model"`
	Content   string                 `json:"content"`
	ToolCalls []*ToolCall            `json:"tool_calls,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Usage     *Usage                 `json:"usage,omitempty"`
}

// StreamChunk represents a chunk in a streaming response
type StreamChunk struct {
	Content   string                 `json:"content"`
	Done      bool                   `json:"done"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Error     error                  `json:"error,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ToolDefinition represents a tool that can be called
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCall represents a call to a tool
type ToolCall struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Success   bool                   `json:"success"`
	Output    interface{}            `json:"output"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Duration  time.Duration           `json:"duration"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Condition represents a condition to evaluate
type Condition struct {
	Type      string                 `json:"type"` // "llm", "expression", "custom"
	Expression string                `json:"expression,omitempty"`
	Context   map[string]interface{} `json:"context"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Decision represents the result of a decision evaluation
type Decision struct {
	Result    bool                   `json:"result"`
	Confidence float64               `json:"confidence,omitempty"`
	Reason    string                 `json:"reason,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Branch represents a branch in a decision tree
type Branch struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Condition   *Condition             `json:"condition"`
	Priority    int                    `json:"priority"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowState represents the state of a workflow
type WorkflowState struct {
	WorkflowID   string                 `json:"workflow_id"`
	Status       string                 `json:"status"` // "running", "paused", "completed", "failed"
	Step         string                 `json:"step"`
	Context      map[string]interface{} `json:"context"`
	History      []*StateHistory        `json:"history"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// StateHistory represents a history entry in workflow state
type StateHistory struct {
	Timestamp time.Time              `json:"timestamp"`
	Step      string                 `json:"step"`
	Action    string                 `json:"action"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Event represents an event in the system
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EventFilter defines criteria for filtering events
type EventFilter struct {
	Types   []string               `json:"types,omitempty"`
	Sources []string               `json:"sources,omitempty"`
	Match   map[string]interface{} `json:"match,omitempty"`
}

// ValidationRules defines rules for validation
type ValidationRules struct {
	Type      string                 `json:"type"` // "schema", "custom", "llm"
	Schema    map[string]interface{} `json:"schema,omitempty"`
	Rules     []string               `json:"rules,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid     bool                   `json:"valid"`
	Errors    []string               `json:"errors,omitempty"`
	Warnings  []string               `json:"warnings,omitempty"`
	Score     float64                `json:"score,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ApprovalRequest represents a request for human approval
type ApprovalRequest struct {
	RequestID   string                 `json:"request_id"`
	WorkflowID  string                 `json:"workflow_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
	Options     []string               `json:"options,omitempty"`
	Timeout     *time.Duration         `json:"timeout,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ApprovalResponse represents a response to an approval request
type ApprovalResponse struct {
	RequestID  string                 `json:"request_id"`
	Approved   bool                   `json:"approved"`
	Choice     string                 `json:"choice,omitempty"`
	Comment    string                 `json:"comment,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// InputPrompt represents a prompt for human input
type InputPrompt struct {
	PromptID   string                 `json:"prompt_id"`
	WorkflowID string                 `json:"workflow_id"`
	Type       string                 `json:"type"` // "text", "choice", "file", "custom"
	Prompt     string                 `json:"prompt"`
	Options    []string               `json:"options,omitempty"`
	Required   bool                   `json:"required"`
	Timeout    *time.Duration         `json:"timeout,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// InputResponse represents a response to an input prompt
type InputResponse struct {
	PromptID  string                 `json:"prompt_id"`
	Value     interface{}            `json:"value"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Transformation represents a data transformation
type Transformation struct {
	Type       string                 `json:"type"` // "map", "filter", "reduce", "custom"
	Config     map[string]interface{} `json:"config"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}
