package plugins

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// ExtendedPlugin extends the base Plugin interface with additional capabilities
type ExtendedPlugin interface {
	Plugin

	// GetAPIEndpoints returns API endpoint definitions that this plugin exposes
	// Returns nil if the plugin doesn't expose any custom endpoints
	GetAPIEndpoints() []APIEndpoint

	// GetAgentDefinition returns the agent definition if this plugin represents an agent
	// Returns nil if the plugin is not an agent
	GetAgentDefinition() *AgentDefinition
}

// APIEndpoint defines a custom HTTP endpoint exposed by a plugin
type APIEndpoint struct {
	// Path is the endpoint path (e.g., "/custom/endpoint")
	// Will be prefixed with /v1/plugins/{plugin_name}
	Path string `json:"path"`

	// Method is the HTTP method (GET, POST, PUT, DELETE, etc.)
	Method string `json:"method"`

	// Handler is the Gin handler function for this endpoint
	Handler gin.HandlerFunc `json:"-"`

	// Description describes what this endpoint does
	Description string `json:"description"`

	// RequestSchema defines the expected request body schema (JSON Schema)
	// Only applicable for POST, PUT, PATCH methods
	RequestSchema map[string]interface{} `json:"request_schema,omitempty"`

	// ResponseSchema defines the expected response schema (JSON Schema)
	ResponseSchema map[string]interface{} `json:"response_schema,omitempty"`

	// RequiresAuth indicates if this endpoint requires authentication
	RequiresAuth bool `json:"requires_auth"`

	// RequiresRateLimit indicates if this endpoint should be rate limited
	RequiresRateLimit bool `json:"requires_rate_limit"`
}

// AgentDefinition defines an agent with its capabilities and configuration
type AgentDefinition struct {
	// Name is the agent name
	Name string `json:"name"`

	// Description describes what the agent does
	Description string `json:"description"`

	// Capabilities lists what the agent can do
	Capabilities []string `json:"capabilities"`

	// Tools lists the tools available to this agent
	Tools []string `json:"tools"`

	// Model is the default LLM model for this agent
	Model string `json:"model"`

	// SystemPrompt is the system prompt for this agent
	SystemPrompt string `json:"system_prompt,omitempty"`

	// MaxIterations is the maximum number of iterations for agentic loops
	MaxIterations int `json:"max_iterations,omitempty"`

	// Temperature is the default temperature for LLM calls
	Temperature float64 `json:"temperature,omitempty"`

	// Workflow is the default workflow for this agent
	Workflow *Workflow `json:"workflow,omitempty"`

	// ConfigSchema defines the configuration schema for this agent (JSON Schema)
	ConfigSchema map[string]interface{} `json:"config_schema,omitempty"`
}

// AgentExecutor executes agentic workflows with agent-specific configuration
type AgentExecutor struct {
	*WorkflowExecutor
	agent *AgentDefinition
}

// NewAgentExecutor creates a new agent executor
func NewAgentExecutor(
	agent *AgentDefinition,
	llmHandler func(ctx context.Context, model string, messages []map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error),
	toolHandler func(ctx context.Context, toolName string, arguments map[string]interface{}) (map[string]interface{}, error),
) *AgentExecutor {
	return &AgentExecutor{
		WorkflowExecutor: NewWorkflowExecutor(llmHandler, toolHandler),
		agent:            agent,
	}
}

// ExecuteAgent executes an agent with the given input and configuration
func (e *AgentExecutor) ExecuteAgent(ctx context.Context, input map[string]interface{}, config map[string]interface{}) ([]StepResult, error) {
	// Merge agent configuration with provided config
	mergedConfig := make(map[string]interface{})
	if e.agent != nil {
		// Apply agent defaults
		if e.agent.Model != "" {
			mergedConfig["model"] = e.agent.Model
		}
		if e.agent.Temperature > 0 {
			mergedConfig["temperature"] = e.agent.Temperature
		}
		if e.agent.MaxIterations > 0 {
			mergedConfig["max_iterations"] = e.agent.MaxIterations
		}
	}
	// Override with provided config
	for k, v := range config {
		mergedConfig[k] = v
	}

	// Use agent's workflow or create default workflow
	workflow := e.agent.Workflow
	if workflow == nil {
		// Create a default agentic workflow
		workflow = e.createDefaultAgentWorkflow(mergedConfig)
	}

	// Merge input with agent context
	agentInput := make(map[string]interface{})
	for k, v := range input {
		agentInput[k] = v
	}
	if e.agent.SystemPrompt != "" {
		agentInput["system_prompt"] = e.agent.SystemPrompt
	}

	return e.Execute(ctx, workflow, agentInput)
}

// createDefaultAgentWorkflow creates a default agentic workflow
func (e *AgentExecutor) createDefaultAgentWorkflow(config map[string]interface{}) *Workflow {
	model := "llama3.2"
	if m, ok := config["model"].(string); ok {
		model = m
	}

	steps := []WorkflowStep{
		{
			ID:          "analyze",
			Name:        "Analyze Request",
			Description: "Analyze the user request and determine actions",
			Type:        "llm_call",
			Config: map[string]interface{}{
				"model": model,
			},
		},
	}

	// Add iterative agentic loop (simplified - actual implementation would be more sophisticated)
	// For now, just add one iteration step
	steps = append(steps, WorkflowStep{
		ID:          "iteration_0",
		Name:        "Agent Iteration",
		Description: "Agent processing iteration",
		Type:        "llm_call",
		Config: map[string]interface{}{
			"model": model,
		},
		Dependencies: []string{"analyze"},
	})

	return &Workflow{
		ID:          "default_agent_workflow",
		Name:        "Default Agent Workflow",
		Description: "Default iterative agentic workflow",
		Steps:       steps,
		MaxRetries:  2,
		Timeout:     60 * time.Second,
	}
}
