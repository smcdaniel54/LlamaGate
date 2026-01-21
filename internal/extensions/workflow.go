package extensions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/rs/zerolog/log"
)

// WorkflowExecutor executes extension workflows
type WorkflowExecutor struct {
	llmHandler LLMHandlerFunc
	baseDir    string
	registry   *Registry // For extension-to-extension calls
}

// NewWorkflowExecutor creates a new workflow executor
func NewWorkflowExecutor(llmHandler LLMHandlerFunc, baseDir string) *WorkflowExecutor {
	return &WorkflowExecutor{
		llmHandler: llmHandler,
		baseDir:    baseDir,
		registry:   nil, // Set via SetRegistry
	}
}

// SetRegistry sets the registry for extension-to-extension calls
func (e *WorkflowExecutor) SetRegistry(registry *Registry) {
	e.registry = registry
}

// Execute executes a workflow extension
// ctx can be a regular context.Context or *ExecutionContext
func (e *WorkflowExecutor) Execute(ctx context.Context, manifest *Manifest, input map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	state := make(map[string]interface{})

	// Copy input to state
	for k, v := range input {
		state[k] = v
	}

	// Convert to ExecutionContext if needed
	var execCtx *ExecutionContext
	if ec, ok := ctx.(*ExecutionContext); ok {
		execCtx = ec
	} else {
		// Create a basic execution context
		execCtx = NewExecutionContext(ctx, "", GetExtensionDir(e.baseDir, manifest.Name))
	}

	// Add trace ID to state if available
	if execCtx.TraceID != "" {
		state["trace_id"] = execCtx.TraceID
	}

	// Execute each step
	for i, step := range manifest.Steps {
		stepResult, err := e.executeStep(execCtx, step, state, manifest)
		if err != nil {
			return nil, fmt.Errorf("step %d (%s) failed: %w", i, step.Uses, err)
		}

		// Merge step result into state
		for k, v := range stepResult {
			state[k] = v
		}
	}

	// Outputs are handled by individual steps (file.write)
	return result, nil
}

// executeStep executes a single workflow step
func (e *WorkflowExecutor) executeStep(ctx *ExecutionContext, step WorkflowStep, state map[string]interface{}, manifest *Manifest) (map[string]interface{}, error) {
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
	case "extension.call":
		return e.callExtension(ctx, step, resolvedWith, state, manifest)
	case "module.load", "module.validate", "module.execute", "module.record":
		return e.executeModuleStep(ctx, step, resolvedWith, state, manifest)
	case "summary.parse":
		return e.parseSummary(ctx, step, resolvedWith, state, manifest)
	case "rules.evaluate":
		return e.evaluateRules(ctx, step, resolvedWith, state, manifest)
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
func (e *WorkflowExecutor) loadTemplate(_ context.Context, _ WorkflowStep, resolvedWith map[string]interface{}, state map[string]interface{}, manifest *Manifest) (map[string]interface{}, error) {
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
func (e *WorkflowExecutor) renderTemplate(_ *ExecutionContext, _ WorkflowStep, resolvedWith map[string]interface{}, state map[string]interface{}, _ *Manifest) (map[string]interface{}, error) {
	// Get template content from resolvedWith, state, or template_content
	templateContent, ok := resolvedWith["template_content"].(string)
	if !ok {
		if tc, ok := state["template_content"].(string); ok {
			templateContent = tc
		} else {
			return nil, fmt.Errorf("template_content not found in state or step configuration")
		}
	}

	// Get variables from resolvedWith or state
	variables := make(map[string]interface{})
	if vars, ok := resolvedWith["variables"].(map[string]interface{}); ok {
		variables = vars
	} else if vars, ok := state["variables"].(map[string]interface{}); ok {
		variables = vars
	}

	// Use Go template parsing for variable substitution
	// First, try to resolve simple {{var}} references from state if not in variables
	resolvedVars := make(map[string]interface{})
	for k, v := range variables {
		resolvedVars[k] = v
	}
	// Also check state for any missing variables
	for k, v := range state {
		if _, exists := resolvedVars[k]; !exists {
			resolvedVars[k] = v
		}
	}

	// Parse and execute template
	tmpl, err := template.New("workflow").Parse(templateContent)
	if err != nil {
		// If template parsing fails, try simple string replacement as fallback
		resolvedTemplate := templateContent
		for k, v := range resolvedVars {
			placeholder := fmt.Sprintf("{{%s}}", k)
			if strVal, ok := v.(string); ok {
				resolvedTemplate = strings.ReplaceAll(resolvedTemplate, placeholder, strVal)
			} else {
				resolvedTemplate = strings.ReplaceAll(resolvedTemplate, placeholder, fmt.Sprintf("%v", v))
			}
		}
		return map[string]interface{}{
			"rendered_prompt": resolvedTemplate,
			"prompt":          resolvedTemplate,
		}, nil
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, resolvedVars); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	resolvedTemplate := buf.String()

	return map[string]interface{}{
		"rendered_prompt": resolvedTemplate,
		"prompt":          resolvedTemplate, // Also set as prompt for llm.chat
	}, nil
}

// callLLM calls the LLM with the rendered prompt
func (e *WorkflowExecutor) callLLM(ctx *ExecutionContext, _ WorkflowStep, resolvedWith map[string]interface{}, state map[string]interface{}) (map[string]interface{}, error) {
	if e.llmHandler == nil {
		return nil, fmt.Errorf("LLM handler not available")
	}

	// Get model from resolvedWith or state
	model := "mistral" // default
	if m, ok := resolvedWith["model"].(string); ok {
		model = m
	} else if m, ok := state["model"].(string); ok {
		model = m
	}

	// Get prompt from state (can be rendered_prompt or prompt)
	prompt, ok := state["rendered_prompt"].(string)
	if !ok {
		if p, ok := state["prompt"].(string); ok {
			prompt = p
		} else {
			return nil, fmt.Errorf("rendered_prompt or prompt not found in state")
		}
	}

	// Get prompt from resolvedWith if provided directly (for direct prompt passing)
	if promptFromWith, ok := resolvedWith["prompt"].(string); ok {
		prompt = promptFromWith
	} else if promptFromState, ok := state["prompt"].(string); ok {
		prompt = promptFromState
	}

	// Call LLM
	messages := []map[string]interface{}{
		{"role": "user", "content": prompt},
	}

	response, err := e.llmHandler(ctx.Context, model, messages, nil)
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
func (e *WorkflowExecutor) writeFile(_ context.Context, _ WorkflowStep, resolvedWith map[string]interface{}, state map[string]interface{}, manifest *Manifest) (map[string]interface{}, error) {
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

// callExtension calls another extension (extension-to-extension invocation)
func (e *WorkflowExecutor) callExtension(ctx *ExecutionContext, _ WorkflowStep, resolvedWith map[string]interface{}, state map[string]interface{}, callingManifest *Manifest) (map[string]interface{}, error) {
	if e.registry == nil {
		return nil, fmt.Errorf("extension registry not available for extension-to-extension calls")
	}

	// Get extension name from step
	extensionName, ok := resolvedWith["extension"].(string)
	if !ok {
		if ext, ok := state["extension"].(string); ok {
			extensionName = ext
		} else {
			return nil, fmt.Errorf("extension name not specified in step")
		}
	}

	// Get input for the called extension
	var extensionInput map[string]interface{}
	if input, ok := resolvedWith["input"].(map[string]interface{}); ok {
		extensionInput = input
	} else if input, ok := state["input"].(map[string]interface{}); ok {
		extensionInput = input
	} else {
		// Use state as input if no explicit input provided
		extensionInput = state
	}

	// Use the provided execution context
	return e.ExecuteExtensionInternal(ctx, extensionName, extensionInput, callingManifest.Name)
}

// ExecuteExtensionInternal provides the internal API for extension-to-extension invocation
// This is the core function that enforces guardrails and provides structured logging
func (e *WorkflowExecutor) ExecuteExtensionInternal(execCtx *ExecutionContext, extensionName string, input map[string]interface{}, callerName string) (map[string]interface{}, error) {
	// Get extension manifest
	manifest, err := e.registry.Get(extensionName)
	if err != nil {
		return nil, &ExecutionError{
			Extension:    extensionName,
			Step:         "lookup",
			ManifestPath: "",
			Message:      "extension not found",
			Details:      err.Error(),
		}
	}

	// Check if enabled
	if !e.registry.IsEnabled(manifest.Name) {
		return nil, &ExecutionError{
			Extension:    extensionName,
			Step:         "validation",
			ManifestPath: "",
			Message:      "extension is disabled",
			Details:      "cannot execute disabled extension",
		}
	}

	// Only workflow extensions can be called
	if manifest.Type != "workflow" {
		return nil, &ExecutionError{
			Extension:    extensionName,
			Step:         "validation",
			ManifestPath: "",
			Message:      "only workflow extensions can be called",
			Details:      fmt.Sprintf("extension type is %s, expected workflow", manifest.Type),
		}
	}

	// Create child execution context with guardrails
	childCtx, err := execCtx.WithChild(manifest.Name)
	if err != nil {
		return nil, err
	}

	// Get manifest path for error reporting
	manifestPath := GetExtensionDir(e.baseDir, manifest.Name)
	manifestPath = filepath.Join(manifestPath, "manifest.yaml")

	// Log extension call
	log.Info().
		Str("trace_id", childCtx.TraceID).
		Str("caller", callerName).
		Str("extension", extensionName).
		Int("call_depth", childCtx.CallDepth).
		Int("remaining_budget", childCtx.CallBudget).
		Msg("Executing extension-to-extension call")

	// Validate required inputs
	for _, inputDef := range manifest.Inputs {
		if inputDef.Required {
			if _, exists := input[inputDef.ID]; !exists {
				return nil, &ExecutionError{
					Extension:    extensionName,
					Step:         "validation",
					ManifestPath: manifestPath,
					Message:      fmt.Sprintf("required input '%s' is missing", inputDef.ID),
					Details:      "input validation failed",
				}
			}
		}
	}

	// Execute the extension
	result, err := e.Execute(childCtx, manifest, input)
	if err != nil {
		return nil, &ExecutionError{
			Extension:    extensionName,
			Step:         "execution",
			ManifestPath: manifestPath,
			Message:      "extension execution failed",
			Details:      err.Error(),
		}
	}

	// Log successful completion
	log.Info().
		Str("trace_id", childCtx.TraceID).
		Str("caller", callerName).
		Str("extension", extensionName).
		Int("call_depth", childCtx.CallDepth).
		Msg("Extension-to-extension call completed successfully")

	return result, nil
}

// parseSummary parses LLM response as structured JSON summary
func (e *WorkflowExecutor) parseSummary(_ *ExecutionContext, _ WorkflowStep, _ map[string]interface{}, state map[string]interface{}, _ *Manifest) (map[string]interface{}, error) {
	// Get LLM response from state
	llmResponse, ok := state["llm_response"].(string)
	if !ok {
		return nil, fmt.Errorf("llm_response not found in state")
	}

	// Try to parse as JSON
	var summary map[string]interface{}
	if err := json.Unmarshal([]byte(llmResponse), &summary); err != nil {
		// If not JSON, try to extract JSON from markdown code blocks
		jsonStart := strings.Index(llmResponse, "{")
		jsonEnd := strings.LastIndex(llmResponse, "}")
		if jsonStart >= 0 && jsonEnd > jsonStart {
			jsonStr := llmResponse[jsonStart : jsonEnd+1]
			if err := json.Unmarshal([]byte(jsonStr), &summary); err != nil {
				return nil, fmt.Errorf("failed to parse summary as JSON: %w", err)
			}
		} else {
			return nil, fmt.Errorf("llm_response does not contain valid JSON")
		}
	}

	return map[string]interface{}{
		"summary": summary,
	}, nil
}

// evaluateRules evaluates simple if-then rules
func (e *WorkflowExecutor) evaluateRules(_ *ExecutionContext, _ WorkflowStep, resolvedWith map[string]interface{}, state map[string]interface{}, _ *Manifest) (map[string]interface{}, error) {
	// Get rules from step
	rules, ok := resolvedWith["rules"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("rules not found in step configuration")
	}

	// Get summary from state
	summary, ok := state["summary"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("summary not found in state")
	}

	// Evaluate rules (simple implementation - checks urgency_level)
	for _, ruleRaw := range rules {
		rule, ok := ruleRaw.(map[string]interface{})
		if !ok {
			continue
		}

		ifCond, ok := rule["if"].(string)
		if !ok {
			continue
		}

		// Simple rule evaluation - check if urgency_level matches
		if strings.Contains(ifCond, "urgency_level") {
			urgencyLevel, ok := summary["urgency_level"].(string)
			if !ok {
				continue
			}

			// Check if condition matches (handle both 'high' and "high" formats)
			urgencyPattern := fmt.Sprintf("'%s'", urgencyLevel)
			urgencyPattern2 := fmt.Sprintf("\"%s\"", urgencyLevel)
			if strings.Contains(ifCond, urgencyPattern) || strings.Contains(ifCond, urgencyPattern2) {
				then, ok := rule["then"].(map[string]interface{})
				if ok {
					return then, nil
				}
			}
		}
	}

	// Default route if no rules match
	return map[string]interface{}{
		"route":    "queue",
		"priority": 3,
	}, nil
}

// executeModuleStep executes module runner steps (module.load, module.validate, module.execute, module.record)
func (e *WorkflowExecutor) executeModuleStep(ctx *ExecutionContext, step WorkflowStep, resolvedWith map[string]interface{}, state map[string]interface{}, manifest *Manifest) (map[string]interface{}, error) {
	switch step.Uses {
	case "module.load":
		return e.loadModule(ctx, step, resolvedWith, state, manifest)
	case "module.validate":
		return e.validateModule(ctx, step, resolvedWith, state, manifest)
	case "module.execute":
		return e.executeModule(ctx, step, resolvedWith, state, manifest)
	case "module.record":
		return e.createModuleRecord(ctx, step, resolvedWith, state, manifest)
	default:
		return nil, fmt.Errorf("unknown module step type: %s", step.Uses)
	}
}

// loadModule loads an AgenticModule manifest
func (e *WorkflowExecutor) loadModule(_ *ExecutionContext, _ WorkflowStep, resolvedWith map[string]interface{}, state map[string]interface{}, _ *Manifest) (map[string]interface{}, error) {
	moduleName, ok := resolvedWith["module_name"].(string)
	if !ok {
		if mn, ok := state["module_name"].(string); ok {
			moduleName = mn
		} else {
			return nil, fmt.Errorf("module_name is required")
		}
	}

	// Load agenticmodule.yaml from agenticmodules/<name>/ directory
	// Try multiple possible locations
	var manifestPath string
	possiblePaths := []string{
		filepath.Join("agenticmodules", moduleName, "agenticmodule.yaml"),
		filepath.Join(e.baseDir, "..", "agenticmodules", moduleName, "agenticmodule.yaml"),
		filepath.Join(filepath.Dir(e.baseDir), "agenticmodules", moduleName, "agenticmodule.yaml"),
	}
	
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			manifestPath = path
			break
		}
	}
	
	if manifestPath == "" {
		return nil, fmt.Errorf("agenticmodule.yaml not found for module '%s'. Tried: %v", moduleName, possiblePaths)
	}
	
	moduleDir := filepath.Dir(manifestPath)

	moduleManifest, err := LoadAgenticModuleManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load module manifest: %w", err)
	}

	// Store module directory for later use
	return map[string]interface{}{
		"module_manifest": moduleManifest,
		"module_dir":      moduleDir,
		"module_name":     moduleName, // Store for error reporting
	}, nil
}

// validateModule validates an AgenticModule and its referenced extensions
func (e *WorkflowExecutor) validateModule(_ *ExecutionContext, _ WorkflowStep, _ map[string]interface{}, state map[string]interface{}, _ *Manifest) (map[string]interface{}, error) {
	moduleManifestRaw, ok := state["module_manifest"]
	if !ok {
		return nil, fmt.Errorf("module_manifest not found in state")
	}

	moduleManifest, ok := moduleManifestRaw.(*AgenticModuleManifest)
	if !ok {
		return nil, fmt.Errorf("invalid module_manifest type")
	}

	// Validate module manifest
	if err := ValidateAgenticModuleManifest(moduleManifest); err != nil {
		return nil, fmt.Errorf("module validation failed: %w", err)
	}

	// Check that all referenced extensions exist
	if e.registry == nil {
		return nil, fmt.Errorf("registry not available for extension validation")
	}

	for i, step := range moduleManifest.Steps {
		if step.Extension == "" {
			return nil, fmt.Errorf("module step %d is missing 'extension' field", i)
		}

		_, err := e.registry.Get(step.Extension)
		if err != nil {
			return nil, fmt.Errorf("module step %d references extension '%s' which is not found: %w", i, step.Extension, err)
		}
	}

	return map[string]interface{}{
		"module_validated": true,
	}, nil
}

// executeModule executes an AgenticModule's workflow steps
func (e *WorkflowExecutor) executeModule(ctx *ExecutionContext, _ WorkflowStep, resolvedWith map[string]interface{}, state map[string]interface{}, manifest *Manifest) (map[string]interface{}, error) {
	moduleManifestRaw, ok := state["module_manifest"]
	if !ok {
		return nil, fmt.Errorf("module_manifest not found in state")
	}

	moduleManifest, ok := moduleManifestRaw.(*AgenticModuleManifest)
	if !ok {
		return nil, fmt.Errorf("invalid module_manifest type")
	}

	// Get module input
	moduleInput, ok := resolvedWith["module_input"].(map[string]interface{})
	if !ok {
		if mi, ok := state["module_input"].(map[string]interface{}); ok {
			moduleInput = mi
		} else {
			return nil, fmt.Errorf("module_input is required")
		}
	}

	// Get guardrails
	maxRuntime := 5 * time.Minute
	if mrs, ok := resolvedWith["max_runtime_seconds"].(float64); ok {
		maxRuntime = time.Duration(mrs) * time.Second
	} else if mrs, ok := state["max_runtime_seconds"].(float64); ok {
		maxRuntime = time.Duration(mrs) * time.Second
	}

	maxSteps := 100
	if ms, ok := resolvedWith["max_steps"].(float64); ok {
		maxSteps = int(ms)
	} else if ms, ok := state["max_steps"].(float64); ok {
		maxSteps = int(ms)
	}

	// Create module execution context
	moduleCtx := &ExecutionContext{
		Context:      ctx.Context,
		CallDepth:    0,
		MaxDepth:     10,
		CallBudget:   maxSteps,
		StartTime:    time.Now(),
		MaxRuntime:   maxRuntime,
		TraceID:      ctx.TraceID,
		ManifestPath: filepath.Join("agenticmodules", moduleManifest.Name, "agenticmodule.yaml"),
	}

	// Execute module steps sequentially
	moduleState := make(map[string]interface{})
	for k, v := range moduleInput {
		moduleState[k] = v
	}

	stepRecords := make([]map[string]interface{}, 0)
	currentOutput := make(map[string]interface{})

	for i, moduleStep := range moduleManifest.Steps {
		stepStartTime := time.Now()

		// Prepare step input
		stepInput := make(map[string]interface{})
		if moduleStep.Input != nil {
			// Resolve input from module state
			for k, v := range moduleStep.Input {
				stepInput[k] = v
			}
		} else {
			// Use module state as input
			stepInput = moduleState
		}

		// Execute extension
		result, err := e.ExecuteExtensionInternal(moduleCtx, moduleStep.Extension, stepInput, "agenticmodule_runner")
		if err != nil {
			stepRecord := map[string]interface{}{
				"step_index":  i,
				"extension":   moduleStep.Extension,
				"status":      "failed",
				"error":       err.Error(),
				"duration_ms": time.Since(stepStartTime).Milliseconds(),
			}
			stepRecords = append(stepRecords, stepRecord)

			// Handle error based on step configuration
			if moduleStep.OnError == "stop" {
				return nil, fmt.Errorf("module step %d (%s) failed: %w", i, moduleStep.Extension, err)
			}
			// Continue on error if on_error is not "stop"
			continue
		}

		// Merge result into module state
		for k, v := range result {
			moduleState[k] = v
			currentOutput[k] = v
		}

		stepRecord := map[string]interface{}{
			"step_index":   i,
			"extension":    moduleStep.Extension,
			"status":       "success",
			"duration_ms":  time.Since(stepStartTime).Milliseconds(),
			"output":       result,
		}
		stepRecords = append(stepRecords, stepRecord)
	}

	return map[string]interface{}{
		"module_output":      currentOutput,
		"step_records":       stepRecords,
		"total_steps":        len(stepRecords),
		"total_duration_ms":  time.Since(moduleCtx.StartTime).Milliseconds(),
	}, nil
}

// createModuleRecord creates a structured run record for the module execution
func (e *WorkflowExecutor) createModuleRecord(ctx *ExecutionContext, _ WorkflowStep, _ map[string]interface{}, state map[string]interface{}, _ *Manifest) (map[string]interface{}, error) {
	moduleManifestRaw, ok := state["module_manifest"]
	if !ok {
		return nil, fmt.Errorf("module_manifest not found in state")
	}

	moduleManifest, ok := moduleManifestRaw.(*AgenticModuleManifest)
	if !ok {
		return nil, fmt.Errorf("invalid module_manifest type")
	}

	stepRecords, ok := state["step_records"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("step_records not found in state")
	}

	moduleOutput, ok := state["module_output"].(map[string]interface{})
	if !ok {
		moduleOutput = make(map[string]interface{})
	}

	totalDuration, ok := state["total_duration_ms"].(int64)
	if !ok {
		totalDuration = 0
	}

	// Create run record
	runRecord := map[string]interface{}{
		"module_name":        moduleManifest.Name,
		"module_version":     moduleManifest.Version,
		"trace_id":           ctx.TraceID,
		"started_at":         ctx.StartTime.Format(time.RFC3339),
		"completed_at":       time.Now().Format(time.RFC3339),
		"total_duration_ms":  totalDuration,
		"steps":              stepRecords,
		"final_output":       moduleOutput,
	}

	return map[string]interface{}{
		"run_record": runRecord,
	}, nil
}
