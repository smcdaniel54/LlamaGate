package extensions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtensionDiscovery_EndToEnd tests the complete extension discovery flow
func TestExtensionDiscovery_EndToEnd(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the three example extensions
	setupExampleExtensions(t, tmpDir)

	// Discover extensions
	manifests, err := DiscoverExtensions(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, 3, len(manifests))

	// Verify all three extensions are discovered
	names := make(map[string]bool)
	for _, m := range manifests {
		names[m.Name] = true
	}
	assert.True(t, names["prompt-template-executor"])
	assert.True(t, names["request-inspector"])
	assert.True(t, names["cost-usage-reporter"])
}

// TestPromptTemplateExecutor_EndToEnd tests the complete workflow for prompt-template-executor
func TestPromptTemplateExecutor_EndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	extDir := filepath.Join(tmpDir, "prompt-template-executor")
	setupPromptTemplateExecutor(t, extDir)

	// Load manifest
	manifest, err := LoadManifest(filepath.Join(extDir, "manifest.yaml"))
	require.NoError(t, err)

	// Create registry and register
	registry := NewRegistry()
	err = registry.Register(manifest)
	require.NoError(t, err)

	// Mock LLM handler
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": "Generated executive summary",
					},
				},
			},
		}, nil
	}

	// Execute workflow
	executor := NewWorkflowExecutor(llmHandler, tmpDir)
	input := map[string]interface{}{
		"template_id": "example",
		"variables": map[string]interface{}{
			"document_type": "executive summary",
			"format":        "markdown",
		},
		"model": "llama3.2",
	}

	result, err := executor.Execute(context.Background(), manifest, input)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify output file was created
	outputPath := filepath.Join(extDir, "output", "result.md")
	assert.FileExists(t, outputPath)

	// Verify file content
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Generated executive summary")
}

// TestRequestInspector_EndToEnd tests the complete flow for request-inspector
func TestRequestInspector_EndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	extDir := filepath.Join(tmpDir, "request-inspector")
	setupRequestInspector(t, extDir)

	// Load manifest
	manifest, err := LoadManifest(filepath.Join(extDir, "manifest.yaml"))
	require.NoError(t, err)

	// Create registry and register
	registry := NewRegistry()
	err = registry.Register(manifest)
	require.NoError(t, err)

	// Create hook manager
	hookManager := NewHookManager(registry, tmpDir)

	// Create test server with middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(hookManager.CreateMiddlewareHook())
	router.POST("/v1/chat/completions", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request
	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify audit log was created (relative to extension directory)
	auditDir := filepath.Join(extDir, "var", "audit")
	auditFiles, err := os.ReadDir(auditDir)
	require.NoError(t, err)
	assert.Greater(t, len(auditFiles), 0)
}

// TestCostUsageReporter_EndToEnd tests the complete flow for cost-usage-reporter
func TestCostUsageReporter_EndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	extDir := filepath.Join(tmpDir, "cost-usage-reporter")
	setupCostUsageReporter(t, extDir)

	// Load manifest
	manifest, err := LoadManifest(filepath.Join(extDir, "manifest.yaml"))
	require.NoError(t, err)

	// Create registry and register
	registry := NewRegistry()
	err = registry.Register(manifest)
	require.NoError(t, err)

	// Create hook manager
	hookManager := NewHookManager(registry, tmpDir)

	// Create mock Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	c.Set("request_id", "test-request-123")

	// Simulate LLM response
	responseData := map[string]interface{}{
		"model": "llama3.2",
		"usage": map[string]interface{}{
			"prompt_tokens":     100,
			"completion_tokens": 200,
			"total_tokens":      300,
		},
	}

	// Execute response hook
	hookManager.ExecuteResponseHooks(c, responseData)

	// Verify usage report was created
	reportPath := filepath.Join(extDir, "output", "usage_report.json")
	assert.FileExists(t, reportPath)

	// Verify report content
	data, err := os.ReadFile(reportPath)
	require.NoError(t, err)

	var report []map[string]interface{}
	err = json.Unmarshal(data, &report)
	require.NoError(t, err)

	assert.Greater(t, len(report), 0)
	lastEntry := report[len(report)-1]
	assert.Equal(t, "test-request-123", lastEntry["request_id"])
	assert.Equal(t, "llama3.2", lastEntry["model"])
	assert.Equal(t, float64(100), lastEntry["prompt_tokens"])
	assert.Equal(t, float64(200), lastEntry["completion_tokens"])
}

// TestExtensionEnableDisable tests enable/disable functionality
func TestExtensionEnableDisable(t *testing.T) {
	registry := NewRegistry()

	enabled := &Manifest{
		Name:        "enabled-ext",
		Version:     "1.0.0",
		Description: "Enabled extension",
		Enabled:     boolPtr(true),
	}

	disabled := &Manifest{
		Name:        "disabled-ext",
		Version:     "1.0.0",
		Description: "Disabled extension",
		Enabled:     boolPtr(false),
	}

	require.NoError(t, registry.Register(enabled))
	require.NoError(t, registry.Register(disabled))

	// Test enabled extension
	assert.True(t, registry.IsEnabled("enabled-ext"))

	// Test disabled extension
	assert.False(t, registry.IsEnabled("disabled-ext"))

	// Test toggle
	err := registry.SetEnabled("enabled-ext", false)
	require.NoError(t, err)
	assert.False(t, registry.IsEnabled("enabled-ext"))

	err = registry.SetEnabled("disabled-ext", true)
	require.NoError(t, err)
	assert.True(t, registry.IsEnabled("disabled-ext"))
}

// Helper functions to set up example extensions

func setupExampleExtensions(t *testing.T, baseDir string) {
	setupPromptTemplateExecutor(t, filepath.Join(baseDir, "prompt-template-executor"))
	setupRequestInspector(t, filepath.Join(baseDir, "request-inspector"))
	setupCostUsageReporter(t, filepath.Join(baseDir, "cost-usage-reporter"))
}

func setupPromptTemplateExecutor(t *testing.T, extDir string) {
	require.NoError(t, os.MkdirAll(filepath.Join(extDir, "templates"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(extDir, "output"), 0755))

	manifest := `name: prompt-template-executor
version: 1.0.0
description: Execute approved prompt templates with structured inputs.
type: workflow
enabled: true

inputs:
  - id: template_id
    type: string
    required: true
  - id: variables
    type: object
    required: true

outputs:
  - id: result
    type: file
    path: ./output/result.md

steps:
  - uses: template.load
  - uses: template.render
  - uses: llm.chat
  - uses: file.write
`

	template := `You are a helpful assistant. Please generate a {{.document_type}} based on the following information:

{{range $key, $value := .variables}}
- {{$key}}: {{$value}}
{{end}}

Please format the output as {{.format}}.
`

	require.NoError(t, os.WriteFile(filepath.Join(extDir, "manifest.yaml"), []byte(manifest), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(extDir, "templates", "example.txt"), []byte(template), 0644))
}

func setupRequestInspector(t *testing.T, extDir string) {
	require.NoError(t, os.MkdirAll(extDir, 0755))

	manifest := `name: request-inspector
version: 1.0.0
description: Redacted audit logging for inbound and outbound requests.
type: middleware
enabled: true

config:
  enabled: true
  sample_rate: 1.0
  audit_dir: ./var/audit
  redact:
    - path: $.messages[*].content
      mode: truncate
      max_len: 120

hooks:
  - on: http.request
    match:
      path_prefix: /v1/
    action: audit.log
`

	require.NoError(t, os.WriteFile(filepath.Join(extDir, "manifest.yaml"), []byte(manifest), 0644))
}

func setupCostUsageReporter(t *testing.T, extDir string) {
	require.NoError(t, os.MkdirAll(filepath.Join(extDir, "output"), 0755))

	manifest := `name: cost-usage-reporter
version: 1.0.0
description: Track token usage and estimated cost per request.
type: observer
enabled: true

config:
  report_interval: per_run

outputs:
  - id: usage_report
    type: file
    path: ./output/usage_report.json

hooks:
  - on: llm.response
    action: usage.track
`

	require.NoError(t, os.WriteFile(filepath.Join(extDir, "manifest.yaml"), []byte(manifest), 0644))
}
