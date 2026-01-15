package extensions

import (
	"bytes"
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

func TestHandler_ListExtensions(t *testing.T) {
	registry := NewRegistry()

	manifest1 := &Manifest{
		Name:        "ext1",
		Version:     "1.0.0",
		Description: "First extension",
		Type:        "workflow",
	}
	manifest2 := &Manifest{
		Name:        "ext2",
		Version:     "2.0.0",
		Description: "Second extension",
		Type:        "middleware",
	}

	require.NoError(t, registry.Register(manifest1))
	require.NoError(t, registry.Register(manifest2))

	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return nil, nil
	}

	handler := NewHandler(registry, llmHandler, "")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/extensions", handler.ListExtensions)

	req := httptest.NewRequest("GET", "/extensions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(2), response["count"])
	extensions := response["extensions"].([]interface{})
	assert.Equal(t, 2, len(extensions))
}

func TestHandler_GetExtension(t *testing.T) {
	registry := NewRegistry()

	manifest := &Manifest{
		Name:        "test-ext",
		Version:     "1.0.0",
		Description: "Test extension",
		Type:        "workflow",
		Inputs: []InputDefinition{
			{ID: "input1", Type: "string", Required: true},
		},
		Outputs: []OutputDefinition{
			{ID: "output1", Type: "file", Path: "./output/result.md"},
		},
	}

	require.NoError(t, registry.Register(manifest))

	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return nil, nil
	}

	handler := NewHandler(registry, llmHandler, "")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/extensions/:name", handler.GetExtension)

	req := httptest.NewRequest("GET", "/extensions/test-ext", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test-ext", response["name"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.Equal(t, "workflow", response["type"])
}

func TestHandler_GetExtension_NotFound(t *testing.T) {
	registry := NewRegistry()
	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return nil, nil
	}

	handler := NewHandler(registry, llmHandler, "")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/extensions/:name", handler.GetExtension)

	req := httptest.NewRequest("GET", "/extensions/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_ExecuteExtension(t *testing.T) {
	tmpDir := t.TempDir()
	extDir := filepath.Join(tmpDir, "test-extension")
	require.NoError(t, os.MkdirAll(filepath.Join(extDir, "templates"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(extDir, "output"), 0755))

	// Create template
	templatePath := filepath.Join(extDir, "templates", "test.txt")
	require.NoError(t, os.WriteFile(templatePath, []byte("Test template"), 0644))

	registry := NewRegistry()

	manifest := &Manifest{
		Name:        "test-extension",
		Version:     "1.0.0",
		Description: "Test extension",
		Type:        "workflow",
		Inputs: []InputDefinition{
			{ID: "template_id", Type: "string", Required: true},
			{ID: "variables", Type: "object", Required: true},
		},
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

	require.NoError(t, registry.Register(manifest))

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

	handler := NewHandler(registry, llmHandler, tmpDir)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/extensions/:name/execute", handler.ExecuteExtension)

	requestBody := map[string]interface{}{
		"template_id": "test",
		"variables": map[string]interface{}{
			"name": "World",
		},
		"model": "llama3.2",
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/extensions/test-extension/execute", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
}

func TestHandler_ExecuteExtension_Disabled(t *testing.T) {
	registry := NewRegistry()

	manifest := &Manifest{
		Name:        "disabled-ext",
		Version:     "1.0.0",
		Description: "Disabled extension",
		Type:        "workflow",
		Enabled:     boolPtr(false),
	}

	require.NoError(t, registry.Register(manifest))

	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return nil, nil
	}

	handler := NewHandler(registry, llmHandler, "")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/extensions/:name/execute", handler.ExecuteExtension)

	requestBody := map[string]interface{}{}
	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/extensions/disabled-ext/execute", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestHandler_ExecuteExtension_NonWorkflow(t *testing.T) {
	registry := NewRegistry()

	manifest := &Manifest{
		Name:        "middleware-ext",
		Version:     "1.0.0",
		Description: "Middleware extension",
		Type:        "middleware",
	}

	require.NoError(t, registry.Register(manifest))

	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return nil, nil
	}

	handler := NewHandler(registry, llmHandler, "")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/extensions/:name/execute", handler.ExecuteExtension)

	requestBody := map[string]interface{}{}
	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/extensions/middleware-ext/execute", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_ExecuteExtension_MissingRequiredInput(t *testing.T) {
	registry := NewRegistry()

	manifest := &Manifest{
		Name:        "test-ext",
		Version:     "1.0.0",
		Description: "Test extension",
		Type:        "workflow",
		Inputs: []InputDefinition{
			{ID: "required_input", Type: "string", Required: true},
		},
	}

	require.NoError(t, registry.Register(manifest))

	llmHandler := func(_ context.Context, _ string, _ []map[string]interface{}, _ map[string]interface{}) (map[string]interface{}, error) {
		return nil, nil
	}

	handler := NewHandler(registry, llmHandler, "")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/extensions/:name/execute", handler.ExecuteExtension)

	requestBody := map[string]interface{}{}
	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/extensions/test-ext/execute", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
