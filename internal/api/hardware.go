package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/hardware"
	"github.com/llamagate/llamagate/internal/middleware"
	"github.com/llamagate/llamagate/internal/response"
	"github.com/rs/zerolog/log"
)

// HardwareHandler handles hardware-related API requests
type HardwareHandler struct {
	detector    *hardware.Detector
	recommender *hardware.Recommender
}

// NewHardwareHandler creates a new hardware handler
func NewHardwareHandler() *HardwareHandler {
	detector := hardware.NewDetector()
	recommender := hardware.NewRecommender()

	// Load recommendations data (embedded, should always succeed)
	if err := recommender.LoadRecommendations(); err != nil {
		log.Warn().Err(err).Msg("Failed to load embedded hardware recommendations data")
	}

	return &HardwareHandler{
		detector:    detector,
		recommender: recommender,
	}
}

// GetRecommendations returns hardware specs and recommended models
// GET /v1/hardware/recommendations
func (h *HardwareHandler) GetRecommendations(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	// Detect hardware
	specs, err := h.detector.Detect(c.Request.Context())
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("Hardware detection failed")
		response.InternalError(c, "Failed to detect hardware", requestID)
		return
	}

	// Build output with hardware specs
	output := map[string]interface{}{
		"hardware": specs,
	}

	// Get recommendations (data is embedded in binary)
	groupID, err := h.recommender.ClassifyHardwareGroup(specs)
	if err == nil {
		recommendations, err := h.recommender.GetRecommendations(groupID)
		if err == nil {
			output["hardware_group"] = groupID
			output["recommendations"] = recommendations
		} else {
			log.Debug().
				Str("request_id", requestID).
				Err(err).
				Msg("Recommendations not available")
		}
	} else {
		log.Debug().
			Str("request_id", requestID).
			Err(err).
			Msg("Hardware classification not available")
	}

	// Return the result as JSON
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    output,
	})
}
