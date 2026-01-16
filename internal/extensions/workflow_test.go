package extensions

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkflowExecutor_Execute(t *testing.T) {
	tmpDir := t.TempDir()
	extDir := filepath.Join(tmpDir, "test-extension")
	require.NoError(t, os.MkdirAll(extDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(extDir, "templates"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(extDir, "output"), 0755))

	// Create a simple template
	templateContent := "Hello {{.name}}"
	templatePath := filepath.Join(extDir, "templates", "test.txt")
	require.NoError(t, os.WriteFile(templatePath, []byte(templateContent), 0644))

	// Create manifest
	manifest := &Manifest{
		Name:        "test-extension",
		Version:     "1.0.0",
		Description: "Test extension",
		Type:        "workflow",
		Steps: []WorkflowStep{
			{Uses: "template.load"},
			{Uses: "template.render"},
			{Uses: "llm.chat"},
			{Uses: "file.write"},
		},
		Outputs: []OutputDefinition{
			{ID: "result", Type: "file", Path: "./output/result.md"},
		},
	}

	// Mock LLM handler
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": "Generated response",
					},
				},
			},
		}, nil
	}

	executor := NewWorkflowExecutor(llmHandler, tmpDir)

	input := map[string]interface{}{
		"template_id": "test",
		"variables": map[string]interface{}{
			"name": "World",
		},
		"model": "llama3.2",
	}

	result, err := executor.Execute(context.Background(), manifest, input)
	if err != nil {
		t.Fatalf("Failed to execute workflow: %v", err)
	}

	// Check that output file was created
	outputPath := filepath.Join(extDir, "output", "result.md")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Expected output file to be created")
	}

	// Check file content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if string(content) != "Generated response" {
		t.Errorf("Expected output file to contain 'Generated response', got '%s'", string(content))
	}

	_ = result // Result may be empty, that's okay
}

func TestWorkflowExecutor_TemplateLoad(t *testing.T) {
	tmpDir := t.TempDir()
	extDir := filepath.Join(tmpDir, "test-extension")
	require.NoError(t, os.MkdirAll(filepath.Join(extDir, "templates"), 0755))

	templateContent := "Test template"
	templatePath := filepath.Join(extDir, "templates", "test.txt")
	require.NoError(t, os.WriteFile(templatePath, []byte(templateContent), 0644))

	manifest := &Manifest{Name: "test-extension"}
	executor := NewWorkflowExecutor(nil, tmpDir)

	state := map[string]interface{}{
		"template_id": "test",
	}

	result, err := executor.loadTemplate(context.Background(), WorkflowStep{}, map[string]interface{}{}, state, manifest)
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	if content, ok := result["template_content"].(string); !ok || content != templateContent {
		t.Errorf("Expected template content '%s', got '%v'", templateContent, result["template_content"])
	}
}

func TestWorkflowExecutor_TemplateRender(t *testing.T) {
	executor := NewWorkflowExecutor(nil, "")

	state := map[string]interface{}{
		"template_content": "Hello {{.name}}",
		"variables": map[string]interface{}{
			"name": "World",
		},
	}

	execCtx := NewExecutionContext(context.Background(), "", "")
	result, err := executor.renderTemplate(execCtx, WorkflowStep{}, map[string]interface{}{}, state, &Manifest{})
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if rendered, ok := result["rendered_prompt"].(string); !ok || rendered != "Hello World" {
		t.Errorf("Expected rendered prompt 'Hello World', got '%v'", result["rendered_prompt"])
	}
}

func TestWorkflowExecutor_CallLLM(t *testing.T) {
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": "LLM response",
					},
				},
			},
		}, nil
	}

	executor := NewWorkflowExecutor(llmHandler, "")

	state := map[string]interface{}{
		"rendered_prompt": "Test prompt",
		"model":           "llama3.2",
	}

	execCtx := NewExecutionContext(context.Background(), "", "")
	result, err := executor.callLLM(execCtx, WorkflowStep{}, map[string]interface{}{}, state)
	if err != nil {
		t.Fatalf("Failed to call LLM: %v", err)
	}

	if response, ok := result["llm_response"].(string); !ok || response != "LLM response" {
		t.Errorf("Expected LLM response 'LLM response', got '%v'", result["llm_response"])
	}
}

func TestWorkflowExecutor_WriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	extDir := filepath.Join(tmpDir, "test-extension")
	require.NoError(t, os.MkdirAll(extDir, 0755))

	manifest := &Manifest{
		Name: "test-extension",
		Outputs: []OutputDefinition{
			{ID: "result", Type: "file", Path: "./output/result.md"},
		},
	}

	executor := NewWorkflowExecutor(nil, tmpDir)

	state := map[string]interface{}{
		"llm_response": "Test content",
	}

	result, err := executor.writeFile(context.Background(), WorkflowStep{}, map[string]interface{}{}, state, manifest)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	outputPath := filepath.Join(extDir, "output", "result.md")
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Expected output file to be created")
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if string(content) != "Test content" {
		t.Errorf("Expected file content 'Test content', got '%s'", string(content))
	}

	_ = result
}
