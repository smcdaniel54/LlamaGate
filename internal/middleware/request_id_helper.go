package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

// RequestIDContextKey is the key used to store request ID in context
// Using a string constant allows cross-package access
const RequestIDContextKey = "request_id"

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

// WithRequestID adds a request ID to a context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDContextKey, requestID)
}

// GetRequestIDFromContext extracts the request ID from a context
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDContextKey).(string); ok {
		return requestID
	}
	return ""
}
