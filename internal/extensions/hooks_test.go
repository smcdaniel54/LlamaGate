package extensions

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHookManager_AuditLog(t *testing.T) {
	tmpDir := t.TempDir()
	auditDir := filepath.Join(tmpDir, "audit")

	registry := NewRegistry()
	manifest := &Manifest{
		Name:        "request-inspector",
		Version:     "1.0.0",
		Description: "Request inspector",
		Type:        "middleware",
		Config: map[string]interface{}{
			"audit_dir":  auditDir,
			"sample_rate": 1.0,
		},
		Hooks: []HookDefinition{
			{On: "http.request", Action: "audit.log"},
		},
	}

	registry.Register(manifest)

	hookManager := NewHookManager(registry, tmpDir)

	// Create a mock Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	c.Request = req
	c.Set("request_id", "test-request-id")

	err := hookManager.auditLog(manifest, c, nil)
	assert.NoError(t, err)

	// Check that audit file was created
	auditFile := filepath.Join(auditDir, "audit-"+time.Now().Format("2006-01-02")+".jsonl")
	if _, err := os.Stat(auditFile); os.IsNotExist(err) {
		t.Fatal("Expected audit file to be created")
	}

	// Read and verify audit entry
	data, err := os.ReadFile(auditFile)
	assert.NoError(t, err)

	var entry map[string]interface{}
	lines := string(data)
	// Get last line (JSONL format)
	lastLine := ""
	for i := len(lines) - 1; i >= 0; i-- {
		if lines[i] == '\n' {
			lastLine = lines[i+1:]
			break
		}
	}
	if lastLine == "" {
		lastLine = lines
	}

	err = json.Unmarshal([]byte(lastLine), &entry)
	assert.NoError(t, err)

	assert.Equal(t, "test-request-id", entry["request_id"])
}

func TestHookManager_TrackUsage(t *testing.T) {
	tmpDir := t.TempDir()
	extDir := filepath.Join(tmpDir, "cost-usage-reporter")
	os.MkdirAll(filepath.Join(extDir, "output"), 0755)

	registry := NewRegistry()
	manifest := &Manifest{
		Name:        "cost-usage-reporter",
		Version:     "1.0.0",
		Description: "Cost usage reporter",
		Type:        "observer",
		Outputs: []OutputDefinition{
			{ID: "usage_report", Type: "file", Path: "./output/usage_report.json"},
		},
		Hooks: []HookDefinition{
			{On: "llm.response", Action: "usage.track"},
		},
	}

	registry.Register(manifest)

	hookManager := NewHookManager(registry, tmpDir)

	// Create a mock Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	c.Set("request_id", "test-request-id")

	responseData := map[string]interface{}{
		"model": "llama3.2",
		"usage": map[string]interface{}{
			"prompt_tokens":     10,
			"completion_tokens": 20,
			"total_tokens":      30,
		},
	}

	err := hookManager.trackUsage(manifest, c, responseData)
	assert.NoError(t, err)

	// Check that usage report was created
	reportFile := filepath.Join(extDir, "output", "usage_report.json")
	if _, err := os.Stat(reportFile); os.IsNotExist(err) {
		t.Fatal("Expected usage report file to be created")
	}

	// Read and verify report
	data, err := os.ReadFile(reportFile)
	assert.NoError(t, err)

	var report []map[string]interface{}
	err = json.Unmarshal(data, &report)
	assert.NoError(t, err)

	assert.Greater(t, len(report), 0)
	assert.Equal(t, "test-request-id", report[len(report)-1]["request_id"])
	assert.Equal(t, "llama3.2", report[len(report)-1]["model"])
}

func TestHookManager_MatchesRequest(t *testing.T) {
	registry := NewRegistry()
	hookManager := NewHookManager(registry, "")

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	req := httptest.NewRequest("GET", "/v1/chat/completions", nil)
	c.Request = req

	// Test path prefix matching
	hook := HookDefinition{
		On: "http.request",
		Match: map[string]interface{}{
			"path_prefix": "/v1/",
		},
	}

	c.Request.URL.Path = "/v1/chat/completions"
	assert.True(t, hookManager.matchesRequest(c, hook))

	c.Request.URL.Path = "/health"
	assert.False(t, hookManager.matchesRequest(c, hook))

	// Test no match criteria (matches all)
	hookNoMatch := HookDefinition{
		On: "http.request",
	}
	assert.True(t, hookManager.matchesRequest(c, hookNoMatch))
}
