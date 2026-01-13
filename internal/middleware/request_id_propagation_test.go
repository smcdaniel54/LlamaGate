package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRequestIDPropagation verifies that request IDs are consistently propagated
func TestRequestIDPropagation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())

	// Track request IDs through the request
	var capturedRequestID string
	router.GET("/test", func(c *gin.Context) {
		capturedRequestID = GetRequestID(c)
		c.JSON(http.StatusOK, gin.H{"request_id": capturedRequestID})
	})

	// Test 1: Request without X-Request-ID header (should generate one)
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	require.Equal(t, http.StatusOK, w1.Code)
	assert.NotEmpty(t, capturedRequestID, "Request ID should be generated")
	assert.Equal(t, capturedRequestID, w1.Header().Get("X-Request-ID"), "Request ID should be in response header")

	// Test 2: Request with X-Request-ID header (should use provided one)
	providedRequestID := "test-request-id-12345"
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Request-ID", providedRequestID)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	require.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, providedRequestID, capturedRequestID, "Should use provided request ID")
	assert.Equal(t, providedRequestID, w2.Header().Get("X-Request-ID"), "Request ID should be in response header")
}

// TestRequestIDContextPropagation verifies request ID propagation through context
func TestRequestIDContextPropagation(t *testing.T) {
	// Test WithRequestID and GetRequestIDFromContext
	ctx := context.Background()
	requestID := "test-context-request-id"

	ctxWithID := WithRequestID(ctx, requestID)
	retrievedID := GetRequestIDFromContext(ctxWithID)

	assert.Equal(t, requestID, retrievedID, "Request ID should be retrievable from context")

	// Test with context without request ID
	emptyID := GetRequestIDFromContext(ctx)
	assert.Empty(t, emptyID, "Context without request ID should return empty string")
}

// TestRequestIDConsistency verifies that a single request produces logs with the same request ID
func TestRequestIDConsistency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())

	var requestIDs []string
	router.GET("/test", func(c *gin.Context) {
		// Simulate multiple log calls that should all have the same request ID
		requestIDs = append(requestIDs, GetRequestID(c))
		requestIDs = append(requestIDs, GetRequestID(c))
		requestIDs = append(requestIDs, GetRequestID(c))
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Len(t, requestIDs, 3, "Should capture 3 request IDs")

	// All request IDs should be the same
	for i := 1; i < len(requestIDs); i++ {
		assert.Equal(t, requestIDs[0], requestIDs[i], "All request IDs should be consistent")
	}
}
