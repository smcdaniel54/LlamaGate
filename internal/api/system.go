package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/config"
	"github.com/llamagate/llamagate/internal/introspection"
	"github.com/llamagate/llamagate/internal/memory"
	"github.com/llamagate/llamagate/internal/middleware"
)

// SystemHandler handles GET /v1/system/* introspection endpoints.
type SystemHandler struct {
	cfg          *config.Config
	memoryStore  memory.Store
	ollamaHost   string
	healthTimeout time.Duration
}

// NewSystemHandler creates a SystemHandler.
func NewSystemHandler(cfg *config.Config, memoryStore memory.Store) *SystemHandler {
	return &SystemHandler{
		cfg:           cfg,
		memoryStore:   memoryStore,
		ollamaHost:    cfg.OllamaHost,
		healthTimeout: cfg.HealthCheckTimeout,
	}
}

func (h *SystemHandler) ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": data})
}

func (h *SystemHandler) fail(c *gin.Context, code string, message string, status int) {
	c.JSON(status, gin.H{
		"ok": false,
		"error": gin.H{"code": code, "message": message},
	})
}

// GetInfo returns runtime snapshot + capability flags.
func (h *SystemHandler) GetInfo(c *gin.Context) {
	_ = middleware.GetRequestID(c)
	rt := introspection.CaptureRuntime()
	cap := gin.H{
		"endpoints":             []string{"/v1/chat/completions", "/v1/models", "/v1/system/*"},
		"mcp_enabled":           h.cfg.MCP != nil && h.cfg.MCP.Enabled,
		"memory_enabled":        h.cfg.Memory != nil && h.cfg.Memory.Enabled,
		"introspection_enabled": true,
	}
	h.ok(c, gin.H{
		"runtime":      rt,
		"capabilities": cap,
	})
}

// GetHardware returns hardware snapshot (sanitized per config).
func (h *SystemHandler) GetHardware(c *gin.Context) {
	_ = middleware.GetRequestID(c)
	if h.cfg.Introspection == nil || !h.cfg.Introspection.Hardware.Enabled {
		h.ok(c, gin.H{"message": "hardware introspection disabled"})
		return
	}
	level := introspection.DetailLevel(h.cfg.Introspection.Hardware.DetailLevel)
	if level != introspection.DetailStandard && level != introspection.DetailFull {
		level = introspection.DetailMinimal
	}
	dataDir := ""
	if h.cfg.Memory != nil && h.cfg.Memory.Dir != "" {
		dataDir = h.cfg.Memory.Dir
	}
	snap := introspection.CaptureHardwareWithTimeout(dataDir, level, 5*time.Second)
	h.ok(c, snap)
}

// GetModels returns models snapshot from backend.
func (h *SystemHandler) GetModels(c *gin.Context) {
	_ = middleware.GetRequestID(c)
	snap, err := introspection.CaptureModels(c.Request.Context(), h.ollamaHost, h.healthTimeout)
	if err != nil {
		h.fail(c, "backend_error", err.Error(), http.StatusServiceUnavailable)
		return
	}
	h.ok(c, snap)
}

// GetHealth returns backend health snapshot.
func (h *SystemHandler) GetHealth(c *gin.Context) {
	_ = middleware.GetRequestID(c)
	snap := introspection.CaptureHealth(c.Request.Context(), h.ollamaHost, h.healthTimeout)
	h.ok(c, snap)
}

// GetConfig returns sanitized config (no secrets).
func (h *SystemHandler) GetConfig(c *gin.Context) {
	_ = middleware.GetRequestID(c)
	m, err := introspection.ConfigToMap(h.cfg)
	if err != nil {
		h.fail(c, "internal_error", "failed to serialize config", http.StatusInternalServerError)
		return
	}
	sanitized := introspection.SanitizeConfig(m)
	h.ok(c, sanitized)
}

// GetMemory returns memory status summary only.
func (h *SystemHandler) GetMemory(c *gin.Context) {
	_ = middleware.GetRequestID(c)
	if h.cfg.Memory == nil || !h.cfg.Memory.Enabled || h.memoryStore == nil {
		h.fail(c, "disabled", "memory is disabled", http.StatusNotFound)
		return
	}
	status, err := h.memoryStore.MemoryStatus()
	if err != nil {
		h.fail(c, "internal_error", err.Error(), http.StatusInternalServerError)
		return
	}
	h.ok(c, status)
}
