package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimitMiddleware_AllowsRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rl := NewRateLimitMiddleware(10.0) // 10 requests per second
	router.Use(rl.Handler())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make a request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitMiddleware_ExceedsLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rl := NewRateLimitMiddleware(1.0) // 1 request per second
	router.Use(rl.Handler())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request immediately after should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRateLimitMiddleware_HealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rl := NewRateLimitMiddleware(0.1) // Very low rate limit
	router.Use(rl.Handler())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Health endpoint should not be rate limited
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}
}

func TestRateLimitMiddleware_ZeroRPS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rl := NewRateLimitMiddleware(0.0) // Zero RPS should default to 1
	router.Use(rl.Handler())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Should still allow at least one request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitMiddleware_AllowsAfterTime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rl := NewRateLimitMiddleware(2.0) // 2 requests per second
	router.Use(rl.Handler())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make 2 requests quickly
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Third request should be rate limited
	req3 := httptest.NewRequest("GET", "/test", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)

	// Wait a bit and try again
	time.Sleep(600 * time.Millisecond)
	req4 := httptest.NewRequest("GET", "/test", nil)
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)
}

func TestNewRateLimitMiddleware(t *testing.T) {
	// Test with various RPS values
	testCases := []struct {
		rps      float64
		expected bool
	}{
		{0.0, true},   // Should handle zero
		{1.0, true},   // Normal value
		{10.0, true},  // Normal value
		{100.0, true}, // High value
		{-1.0, true},   // Negative should still work (burst defaults to 1)
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			rl := NewRateLimitMiddleware(tc.rps)
			assert.NotNil(t, rl)
			assert.NotNil(t, rl.limiter)
		})
	}
}

// TestRateLimitMiddleware_429ResponseFormat verifies that 429 responses have consistent OpenAI-compatible format
func TestRateLimitMiddleware_429ResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware()) // Add request ID middleware for proper request IDs
	rl := NewRateLimitMiddleware(1.0) // 1 request per second
	router.Use(rl.Handler())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusOK, w1.Code, "First request should succeed")

	// Second request immediately after should be rate limited (429)
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Verify HTTP status code
	assert.Equal(t, http.StatusTooManyRequests, w2.Code, "Should return 429 Too Many Requests")

	// Verify Retry-After header is present
	retryAfterHeader := w2.Header().Get("Retry-After")
	assert.NotEmpty(t, retryAfterHeader, "Retry-After header should be present")
	retryAfterSeconds, err := strconv.Atoi(retryAfterHeader)
	require.NoError(t, err, "Retry-After should be a valid integer")
	assert.GreaterOrEqual(t, retryAfterSeconds, 1, "Retry-After should be at least 1 second")
	assert.LessOrEqual(t, retryAfterSeconds, 60, "Retry-After should be reasonable (<= 60 seconds)")

	// Verify response body is valid JSON with OpenAI-compatible error structure
	var response map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify OpenAI-compatible error structure
	assert.Contains(t, response, "error", "Response must contain 'error' field")
	errorObj, ok := response["error"].(map[string]interface{})
	require.True(t, ok, "Error should be an object")

	assert.Contains(t, errorObj, "message", "Error must contain 'message' field")
	assert.Contains(t, errorObj, "type", "Error must contain 'type' field")
	assert.Equal(t, "rate_limit_error", errorObj["type"], "Error type should be 'rate_limit_error'")
	assert.NotEmpty(t, errorObj["message"], "Error message should not be empty")
	assert.Contains(t, errorObj["message"], "Rate limit exceeded", "Error message should mention rate limit")

	// Verify request_id is present in error object
	assert.Contains(t, errorObj, "request_id", "Error should contain 'request_id' field")
	requestID, ok := errorObj["request_id"].(string)
	assert.True(t, ok, "request_id should be a string")
	assert.NotEmpty(t, requestID, "request_id should not be empty")
}
