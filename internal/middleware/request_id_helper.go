package middleware

import (
	"github.com/gin-gonic/gin"
)

// GetRequestID extracts the request ID from the Gin context with a fallback
// Returns the request ID if present, or an empty string if not found
func GetRequestID(c *gin.Context) string {
	requestID := c.GetString("request_id")
	if requestID == "" {
		// Fallback: try to get from header (for external requests)
		requestID = c.GetHeader("X-Request-ID")
	}
	return requestID
}

