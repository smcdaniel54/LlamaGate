package plugins

import (
	"context"
	"fmt"
)

// Agent represents an agentic AI agent with capabilities and workflows
type Agent interface {
	// GetDefinition returns the agent definition
	GetDefinition() *AgentDefinition

	// Execute executes the agent with given input and configuration
	Execute(ctx context.Context, input map[string]interface{}, config map[string]interface{}) (*AgentResult, error)
}

// AgentResult represents the result of agent execution
type AgentResult struct {
	// Success indicates if execution was successful
	Success bool `json:"success"`

	// Response contains the agent's response
	Response string `json:"response"`

	// Steps contains the workflow steps executed
	Steps []StepResult `json:"steps,omitempty"`

	// Iterations is the number of iterations performed
	Iterations int `json:"iterations"`

	// ToolsUsed lists the tools used during execution
	ToolsUsed []string `json:"tools_used,omitempty"`

	// Error contains error information if execution failed
	Error string `json:"error,omitempty"`

	// Metadata contains execution metadata
	Metadata ExecutionMetadata `json:"metadata"`
}

// BaseAgent provides a base implementation for agents
type BaseAgent struct {
	definition *AgentDefinition
	executor   *AgentExecutor
}

// NewBaseAgent creates a new base agent
func NewBaseAgent(definition *AgentDefinition, executor *AgentExecutor) *BaseAgent {
	return &BaseAgent{
		definition: definition,
		executor:   executor,
	}
}

// GetDefinition returns the agent definition
func (a *BaseAgent) GetDefinition() *AgentDefinition {
	return a.definition
}

// Execute executes the agent
func (a *BaseAgent) Execute(ctx context.Context, input map[string]interface{}, config map[string]interface{}) (*AgentResult, error) {
	// Execute using agent executor
	stepResults, err := a.executor.ExecuteAgent(ctx, input, config)
	if err != nil {
		return &AgentResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Extract response from last step
	response := ""
	toolsUsed := make(map[string]bool)

	if len(stepResults) > 0 {
		lastStep := stepResults[len(stepResults)-1]
		if lastStep.Success && lastStep.Output != nil {
			if resp, ok := lastStep.Output["response"].(string); ok {
				response = resp
			} else if resp, ok := lastStep.Output["llm_response"].(string); ok {
				response = resp
			}

			// Collect tools used
			if toolName, ok := lastStep.Output["tool_name"].(string); ok {
				toolsUsed[toolName] = true
			}
		}
	}

	// Build tools used list
	toolsList := make([]string, 0, len(toolsUsed))
	for tool := range toolsUsed {
		toolsList = append(toolsList, tool)
	}

	return &AgentResult{
		Success:    true,
		Response:   response,
		Steps:      stepResults,
		Iterations: len(stepResults),
		ToolsUsed:  toolsList,
	}, nil
}

// AgentRegistry manages agent registration
type AgentRegistry struct {
	agents map[string]Agent
}

// NewAgentRegistry creates a new agent registry
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[string]Agent),
	}
}

// Register registers an agent
func (r *AgentRegistry) Register(agent Agent) error {
	def := agent.GetDefinition()
	if def == nil || def.Name == "" {
		return fmt.Errorf("agent definition name cannot be empty")
	}

	if _, exists := r.agents[def.Name]; exists {
		return fmt.Errorf("agent %s is already registered", def.Name)
	}

	r.agents[def.Name] = agent
	return nil
}

// Get retrieves an agent by name
func (r *AgentRegistry) Get(name string) (Agent, error) {
	agent, exists := r.agents[name]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", name)
	}

	return agent, nil
}

// List returns all registered agents
func (r *AgentRegistry) List() []*AgentDefinition {
	definitions := make([]*AgentDefinition, 0, len(r.agents))
	for _, agent := range r.agents {
		definitions = append(definitions, agent.GetDefinition())
	}
	return definitions
}
