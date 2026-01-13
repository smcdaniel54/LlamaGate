// Package response provides standardized HTTP error response utilities.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Error types
const (
	ErrorTypeInvalidRequest     = "invalid_request_error"
	ErrorTypeInternalError      = "internal_error"
	ErrorTypeServerError        = "server_error"
	ErrorTypeNotFound           = "not_found"
	ErrorTypeServiceUnavailable = "service_unavailable"
	ErrorTypeRateLimit          = "rate_limit_error"
)

// ErrorResponse sends a standardized error response
func ErrorResponse(c *gin.Context, statusCode int, errorType, message, requestID string) {
	errorObj := gin.H{
		"message": message,
		"type":    errorType,
	}
	if requestID != "" {
		errorObj["request_id"] = requestID
	}
	c.JSON(statusCode, gin.H{
		"error": errorObj,
	})
}

// ErrorResponseWithDetails sends a standardized error response with additional details
func ErrorResponseWithDetails(c *gin.Context, statusCode int, errorType, message, requestID, details string) {
	errorObj := gin.H{
		"message": message,
		"type":    errorType,
	}
	if requestID != "" {
		errorObj["request_id"] = requestID
	}
	if details != "" {
		errorObj["details"] = details
	}
	c.JSON(statusCode, gin.H{
		"error": errorObj,
	})
}

// BadRequest sends a 400 Bad Request error
func BadRequest(c *gin.Context, message, requestID string) {
	ErrorResponse(c, http.StatusBadRequest, ErrorTypeInvalidRequest, message, requestID)
}

// InternalError sends a 500 Internal Server Error
func InternalError(c *gin.Context, message, requestID string) {
	ErrorResponse(c, http.StatusInternalServerError, ErrorTypeInternalError, message, requestID)
}

// ServerError sends a 502 Bad Gateway error (for upstream server issues)
func ServerError(c *gin.Context, message, requestID string) {
	ErrorResponse(c, http.StatusBadGateway, ErrorTypeServerError, message, requestID)
}

// NotFound sends a 404 Not Found error
func NotFound(c *gin.Context, message, requestID string) {
	ErrorResponse(c, http.StatusNotFound, ErrorTypeNotFound, message, requestID)
}

// ServiceUnavailable sends a 503 Service Unavailable error
func ServiceUnavailable(c *gin.Context, message, requestID string) {
	ErrorResponse(c, http.StatusServiceUnavailable, ErrorTypeServiceUnavailable, message, requestID)
}

// RateLimitExceeded sends a 429 Too Many Requests error
func RateLimitExceeded(c *gin.Context, message, requestID string) {
	ErrorResponse(c, http.StatusTooManyRequests, ErrorTypeRateLimit, message, requestID)
}
