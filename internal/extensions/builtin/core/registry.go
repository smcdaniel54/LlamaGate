package core

import (
	"context"
	"fmt"
	"sync"
)

// Registry manages extension registration and discovery
type Registry struct {
	mu           sync.RWMutex
	extensions   map[string]Extension
	callers      map[string]AgentCaller
	executors    map[string]ToolExecutor
	evaluators   map[string]DecisionEvaluator
	managers     map[string]StateManager
	publishers   map[string]EventPublisher
	validators   map[string]Validator
	interactions map[string]HumanInteraction
	transformers map[string]Transformer
}

var (
	globalRegistry *Registry
	once           sync.Once
)

// GetRegistry returns the global extension registry
func GetRegistry() *Registry {
	once.Do(func() {
		globalRegistry = &Registry{
			extensions:   make(map[string]Extension),
			callers:      make(map[string]AgentCaller),
			executors:    make(map[string]ToolExecutor),
			evaluators:   make(map[string]DecisionEvaluator),
			managers:     make(map[string]StateManager),
			publishers:   make(map[string]EventPublisher),
			validators:   make(map[string]Validator),
			interactions: make(map[string]HumanInteraction),
			transformers: make(map[string]Transformer),
		}
	})
	return globalRegistry
}

// Register registers an extension
func (r *Registry) Register(ctx context.Context, ext Extension, config map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := ext.Name()
	if _, exists := r.extensions[name]; exists {
		return fmt.Errorf("extension %s already registered", name)
	}

	// Initialize the extension
	if err := ext.Initialize(ctx, config); err != nil {
		return fmt.Errorf("failed to initialize extension %s: %w", name, err)
	}

	// Register in appropriate category
	r.extensions[name] = ext

	if caller, ok := ext.(AgentCaller); ok {
		r.callers[name] = caller
	}
	if executor, ok := ext.(ToolExecutor); ok {
		r.executors[name] = executor
	}
	if evaluator, ok := ext.(DecisionEvaluator); ok {
		r.evaluators[name] = evaluator
	}
	if manager, ok := ext.(StateManager); ok {
		r.managers[name] = manager
	}
	if publisher, ok := ext.(EventPublisher); ok {
		r.publishers[name] = publisher
	}
	if validator, ok := ext.(Validator); ok {
		r.validators[name] = validator
	}
	if interaction, ok := ext.(HumanInteraction); ok {
		r.interactions[name] = interaction
	}
	if transformer, ok := ext.(Transformer); ok {
		r.transformers[name] = transformer
	}

	return nil
}

// Unregister unregisters an extension
func (r *Registry) Unregister(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ext, exists := r.extensions[name]
	if !exists {
		return fmt.Errorf("extension %s not found", name)
	}

	// Shutdown the extension
	if err := ext.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown extension %s: %w", name, err)
	}

	// Remove from all categories
	delete(r.extensions, name)
	delete(r.callers, name)
	delete(r.executors, name)
	delete(r.evaluators, name)
	delete(r.managers, name)
	delete(r.publishers, name)
	delete(r.validators, name)
	delete(r.interactions, name)
	delete(r.transformers, name)

	return nil
}

// GetExtension retrieves an extension by name
func (r *Registry) GetExtension(name string) (Extension, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ext, exists := r.extensions[name]
	if !exists {
		return nil, fmt.Errorf("extension %s not found", name)
	}

	return ext, nil
}

// GetAgentCaller retrieves an agent caller by name
func (r *Registry) GetAgentCaller(name string) (AgentCaller, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	caller, exists := r.callers[name]
	if !exists {
		return nil, fmt.Errorf("agent caller %s not found", name)
	}

	return caller, nil
}

// GetToolExecutor retrieves a tool executor by name
func (r *Registry) GetToolExecutor(name string) (ToolExecutor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	executor, exists := r.executors[name]
	if !exists {
		return nil, fmt.Errorf("tool executor %s not found", name)
	}

	return executor, nil
}

// GetDecisionEvaluator retrieves a decision evaluator by name
func (r *Registry) GetDecisionEvaluator(name string) (DecisionEvaluator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	evaluator, exists := r.evaluators[name]
	if !exists {
		return nil, fmt.Errorf("decision evaluator %s not found", name)
	}

	return evaluator, nil
}

// GetStateManager retrieves a state manager by name
func (r *Registry) GetStateManager(name string) (StateManager, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	manager, exists := r.managers[name]
	if !exists {
		return nil, fmt.Errorf("state manager %s not found", name)
	}

	return manager, nil
}

// GetEventPublisher retrieves an event publisher by name
func (r *Registry) GetEventPublisher(name string) (EventPublisher, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	publisher, exists := r.publishers[name]
	if !exists {
		return nil, fmt.Errorf("event publisher %s not found", name)
	}

	return publisher, nil
}

// GetValidator retrieves a validator by name
func (r *Registry) GetValidator(name string) (Validator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	validator, exists := r.validators[name]
	if !exists {
		return nil, fmt.Errorf("validator %s not found", name)
	}

	return validator, nil
}

// GetHumanInteraction retrieves a human interaction handler by name
func (r *Registry) GetHumanInteraction(name string) (HumanInteraction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	interaction, exists := r.interactions[name]
	if !exists {
		return nil, fmt.Errorf("human interaction %s not found", name)
	}

	return interaction, nil
}

// GetTransformer retrieves a transformer by name
func (r *Registry) GetTransformer(name string) (Transformer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	transformer, exists := r.transformers[name]
	if !exists {
		return nil, fmt.Errorf("transformer %s not found", name)
	}

	return transformer, nil
}

// ListExtensions returns all registered extensions
func (r *Registry) ListExtensions() []Extension {
	r.mu.RLock()
	defer r.mu.RUnlock()

	extensions := make([]Extension, 0, len(r.extensions))
	for _, ext := range r.extensions {
		extensions = append(extensions, ext)
	}

	return extensions
}

// ListByType returns extensions of a specific type
func (r *Registry) ListByType(typ string) []Extension {
	r.mu.RLock()
	defer r.mu.RUnlock()

	extensions := make([]Extension, 0)

	switch typ {
	case "agent_caller":
		for _, caller := range r.callers {
			extensions = append(extensions, caller)
		}
	case "tool_executor":
		for _, executor := range r.executors {
			extensions = append(extensions, executor)
		}
	case "decision_evaluator":
		for _, evaluator := range r.evaluators {
			extensions = append(extensions, evaluator)
		}
	case "state_manager":
		for _, manager := range r.managers {
			extensions = append(extensions, manager)
		}
	case "event_publisher":
		for _, publisher := range r.publishers {
			extensions = append(extensions, publisher)
		}
	case "validator":
		for _, validator := range r.validators {
			extensions = append(extensions, validator)
		}
	case "human_interaction":
		for _, interaction := range r.interactions {
			extensions = append(extensions, interaction)
		}
	case "transformer":
		for _, transformer := range r.transformers {
			extensions = append(extensions, transformer)
		}
	}

	return extensions
}
