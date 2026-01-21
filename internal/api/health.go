// Package api provides HTTP API handlers for LlamaGate endpoints.
//
//nolint:revive // Package name 'api' is standard for HTTP API handlers
package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/config"
	"github.com/llamagate/llamagate/internal/middleware"
	"github.com/rs/zerolog/log"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	ollamaHost         string
	healthCheckTimeout time.Duration
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		ollamaHost:         cfg.OllamaHost,
		healthCheckTimeout: cfg.HealthCheckTimeout,
	}
}

// CheckHealth handles the /health endpoint
func (h *HealthHandler) CheckHealth(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	// Check Ollama connectivity with a timeout
	healthClient := &http.Client{
		Timeout: h.healthCheckTimeout,
	}

	ollamaHealthURL := fmt.Sprintf("%s/api/tags", h.ollamaHost)
	resp, err := healthClient.Get(ollamaHealthURL)
	// Register defer immediately after request to ensure body is always closed
	// This handles both success and error cases where resp might be non-nil
	if resp != nil {
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Warn().
					Str("request_id", requestID).
					Err(closeErr).
					Msg("Failed to close health check response body")
			}
		}()
	}
	if err != nil {
		log.Warn().
			Str("request_id", requestID).
			Err(err).
			Str("ollama_host", h.ollamaHost).
			Msg("Health check: Ollama unreachable")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":      "unhealthy",
			"error":       "Ollama unreachable",
			"ollama_host": h.ollamaHost,
		})
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Warn().
			Str("request_id", requestID).
			Int("status", resp.StatusCode).
			Str("ollama_host", h.ollamaHost).
			Msg("Health check: Ollama returned non-OK status")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":      "unhealthy",
			"error":       fmt.Sprintf("Ollama returned status %d", resp.StatusCode),
			"ollama_host": h.ollamaHost,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "healthy",
		"ollama":      "connected",
		"ollama_host": h.ollamaHost,
	})
}
