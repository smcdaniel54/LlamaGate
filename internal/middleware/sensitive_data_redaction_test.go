package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSensitiveDataNotLogged verifies that sensitive headers are never logged
func TestSensitiveDataNotLogged(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Capture log output
	var logOutput strings.Builder
	logger := zerolog.New(&logOutput)
	originalLogger := log.Logger
	log.Logger = logger
	defer func() {
		log.Logger = originalLogger
	}()

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(AuthMiddleware("sk-test-api-key"))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test with X-API-Key header
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-API-Key", "sk-secret-api-key-12345")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	logOutput1 := logOutput.String()
	logOutput.Reset()

	// Verify API key is NOT in logs
	assert.NotContains(t, logOutput1, "sk-secret-api-key-12345", "API key should not be logged")
	assert.NotContains(t, logOutput1, "X-API-Key", "X-API-Key header name should not be logged with value")

	// Test with Authorization Bearer header
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("Authorization", "Bearer sk-bearer-token-67890")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	logOutput2 := logOutput.String()

	// Verify bearer token is NOT in logs
	assert.NotContains(t, logOutput2, "sk-bearer-token-67890", "Bearer token should not be logged")
	assert.NotContains(t, logOutput2, "Bearer sk-bearer-token-67890", "Full Authorization header should not be logged")

	// Verify authentication failure doesn't log the key
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.Header.Set("X-API-Key", "sk-wrong-key")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	require.Equal(t, http.StatusUnauthorized, w3.Code)
	logOutput3 := logOutput.String()

	// Verify wrong API key is NOT in logs
	assert.NotContains(t, logOutput3, "sk-wrong-key", "Wrong API key should not be logged")
	assert.Contains(t, logOutput3, "Authentication failed", "Should log authentication failure message")
}

// TestRequestIDInLogs verifies that request IDs are present in logs
// Note: This test verifies that request IDs are available for logging.
// Actual logging happens in the main application's logging middleware.
func TestRequestIDInLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Capture log output
	var logOutput strings.Builder
	logger := zerolog.New(&logOutput)
	originalLogger := log.Logger
	log.Logger = logger
	defer func() {
		log.Logger = originalLogger
	}()

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(AuthMiddleware("sk-test-api-key"))

	// Add a handler that logs with request ID to verify it's available
	router.GET("/test", func(c *gin.Context) {
		requestID := GetRequestID(c)
		log.Info().
			Str("request_id", requestID).
			Msg("Test log entry")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "sk-test-api-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	logOutputStr := logOutput.String()

	// Verify request ID is in logs
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID, "Request ID should be generated")
	assert.Contains(t, logOutputStr, requestID, "Request ID should be present in logs")
	assert.Contains(t, logOutputStr, "request_id", "Log should contain request_id field")
}
