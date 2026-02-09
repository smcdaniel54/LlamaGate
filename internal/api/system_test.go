package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/config"
	"github.com/llamagate/llamagate/internal/memory"
	"github.com/llamagate/llamagate/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemHandler_GetInfo_returns200WhenEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	cfg := &config.Config{
		OllamaHost:          "http://localhost:11434",
		HealthCheckTimeout:  5 * time.Second,
		Introspection:       &config.IntrospectionConfig{Enabled: true},
		Memory:              &config.MemoryConfig{Enabled: false},
	}
	handler := NewSystemHandler(cfg, nil)
	router.GET("/v1/system/info", handler.GetInfo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/system/info", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body struct {
		OK   bool `json:"ok"`
		Data struct {
			Runtime      interface{} `json:"runtime"`
			Capabilities interface{} `json:"capabilities"`
		} `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.True(t, body.OK)
	assert.NotNil(t, body.Data.Runtime)
	assert.NotNil(t, body.Data.Capabilities)
}

func TestSystemHandler_GetMemory_returns404WhenMemoryDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	cfg := &config.Config{
		Introspection: &config.IntrospectionConfig{Enabled: true},
		Memory:        &config.MemoryConfig{Enabled: false},
	}
	handler := NewSystemHandler(cfg, nil)
	router.GET("/v1/system/memory", handler.GetMemory)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/system/memory", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.False(t, body["ok"].(bool))
}

func TestSystemHandler_GetMemory_returns200WhenEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	dir := t.TempDir()
	store, err := memory.NewFileStore(dir, memory.DefaultLimits())
	require.NoError(t, err)

	cfg := &config.Config{
		Introspection: &config.IntrospectionConfig{Enabled: true},
		Memory:        &config.MemoryConfig{Enabled: true},
	}
	handler := NewSystemHandler(cfg, store)
	router.GET("/v1/system/memory", handler.GetMemory)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/system/memory", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body struct {
		OK   bool `json:"ok"`
		Data struct {
			UserCount int `json:"user_count"`
		} `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.True(t, body.OK)
}
