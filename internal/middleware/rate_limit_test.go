package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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

