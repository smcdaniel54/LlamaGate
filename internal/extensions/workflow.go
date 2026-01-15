package extensions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/llamagate/llamagate/internal/plugins"
)

// WorkflowExecutor executes extension workflows
type WorkflowExecutor struct {
	llmHandler plugins.LLMHandlerFunc
	baseDir    string
}

// NewWorkflowExecutor creates a new workflow executor
func NewWorkflowExecutor(llmHandler plugins.LLMHandlerFunc, baseDir string) *WorkflowExecutor {
	return &WorkflowExecutor{
		llmHandler: llmHandler,
		baseDir:    baseDir,
	}
}

// Execute executes a workflow extension
func (e *WorkflowExecutor) Execute(ctx context.Context, manifest *Manifest, input map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	state := make(map[string]interface{})

	// Copy input to state
	for k, v := range input {
		state[k] = v
	}

	// Execute each step
	for i, step := range manifest.Steps {
		stepResult, err := e.executeStep(ctx, step, state, manifest)
		if err != nil {
			return nil, fmt.Errorf("step %d (%s) failed: %w", i, step.Uses, err)
		}

		// Merge step result into state
		if stepResult != nil {
			for k, v := range stepResult {
				state[k] = v
			}
		}
	}

	// Outputs are handled by individual steps (file.write)
	return result, nil
}

// executeStep executes a single workflow step
func (e *WorkflowExecutor) executeStep(ctx context.Context, step WorkflowStep, state map[string]interface{}, manifest *Manifest) (map[string]interface{}, error) {
	// Resolve "with" values from state
	resolvedWith := e.resolveStepWith(step.With, state)

	switch step.Uses {
	case "template.load":
		return e.loadTemplate(ctx, step, resolvedWith, state, manifest)
	case "template.render":
		return e.renderTemplate(ctx, step, resolvedWith, state, manifest)
	case "llm.chat":
		return e.callLLM(ctx, step, resolvedWith, state)
	case "file.write":
		return e.writeFile(ctx, step, resolvedWith, state, manifest)
	default:
		return nil, fmt.Errorf("unknown step type: %s", step.Uses)
	}
}

// resolveStepWith resolves step "with" values from state
func (e *WorkflowExecutor) resolveStepWith(with map[string]interface{}, state map[string]interface{}) map[string]interface{} {
	resolved := make(map[string]interface{})
	for k, v := range with {
		// If value is a string that references state, resolve it
		if str, ok := v.(string); ok {
			// Check if it's a state reference (simple implementation)
			if val, exists := state[str]; exists {
				resolved[k] = val
			} else {
				resolved[k] = v
			}
		} else {
			resolved[k] = v
		}
	}
	return resolved
}

// loadTemplate loads a template file
func (e *WorkflowExecutor) loadTemplate(ctx context.Context, step WorkflowStep, resolvedWith map[string]interface{}, state map[string]interface{}, manifest *Manifest) (map[string]interface{}, error) {
	templateID, ok := resolvedWith["template_id"].(string)
	if !ok {
		if tid, ok := state["template_id"].(string); ok {
			templateID = tid
		}
	}
	if templateID == "" {
		return nil, fmt.Errorf("template_id is required")
	}

	// Load template from extension directory
	extDir := GetExtensionDir(e.baseDir, manifest.Name)
	templatePath := filepath.Join(extDir, "templates", templateID+".txt")

	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	return map[string]interface{}{
		"template_content": string(data),
	}, nil
}

// renderTemplate renders a template with variables
func (e *WorkflowExecutor) renderTemplate(ctx context.Context, step WorkflowStep, resolvedWith map[string]interface{}, state map[string]interface{}, manifest *Manifest) (map[string]interface{}, error) {
	templateContent, ok := state["template_content"].(string)
	if !ok {
		return nil, fmt.Errorf("template_content not found in state")
	}

	// Get variables from resolvedWith or state
	variables := make(map[string]interface{})
	if vars, ok := resolvedWith["variables"].(map[string]interface{}); ok {
		variables = vars
	} else if vars, ok := state["variables"].(map[string]interface{}); ok {
		variables = vars
	}

	// Parse and execute template
	tmpl, err := template.New("workflow").Parse(templateContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, variables); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return map[string]interface{}{
		"rendered_prompt": buf.String(),
	}, nil
}

// callLLM calls the LLM with the rendered prompt
func (e *WorkflowExecutor) callLLM(ctx context.Context, step WorkflowStep, resolvedWith map[string]interface{}, state map[string]interface{}) (map[string]interface{}, error) {
	if e.llmHandler == nil {
		return nil, fmt.Errorf("LLM handler not available")
	}

	// Get model from resolvedWith or state
	model := "llama3.2" // default
	if m, ok := resolvedWith["model"].(string); ok {
		model = m
	} else if m, ok := state["model"].(string); ok {
		model = m
	}

	// Get prompt from state
	prompt, ok := state["rendered_prompt"].(string)
	if !ok {
		return nil, fmt.Errorf("rendered_prompt not found in state")
	}

	// Call LLM
	messages := []map[string]interface{}{
		{"role": "user", "content": prompt},
	}

	response, err := e.llmHandler(ctx, model, messages, nil)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Extract content from response
	content := ""
	if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if c, ok := message["content"].(string); ok {
					content = c
				}
			}
		}
	}

	return map[string]interface{}{
		"llm_response": content,
	}, nil
}

// writeFile writes output to a file
func (e *WorkflowExecutor) writeFile(ctx context.Context, step WorkflowStep, resolvedWith map[string]interface{}, state map[string]interface{}, manifest *Manifest) (map[string]interface{}, error) {
	// Get output path from resolvedWith or manifest outputs
	outputPath := ""
	if path, ok := resolvedWith["path"].(string); ok {
		outputPath = path
	} else {
		// Find output definition
		for _, output := range manifest.Outputs {
			if output.Type == "file" && output.Path != "" {
				outputPath = output.Path
				break
			}
		}
	}

	if outputPath == "" {
		return nil, fmt.Errorf("output path not specified")
	}

	// Make path relative to extension directory if relative
	if !filepath.IsAbs(outputPath) {
		extDir := GetExtensionDir(e.baseDir, manifest.Name)
		outputPath = filepath.Join(extDir, outputPath)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get content from state
	content, ok := state["llm_response"].(string)
	if !ok {
		return nil, fmt.Errorf("llm_response not found in state")
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]interface{}{
		"output_file": outputPath,
	}, nil
}
