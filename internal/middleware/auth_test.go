package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupTestRouter creates a test router with auth middleware
func setupTestRouter(apiKey string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware(apiKey))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return router
}

// TestAuthMiddleware_NoAPIKey tests behavior when no API key is configured
func TestAuthMiddleware_NoAPIKey(t *testing.T) {
	router := setupTestRouter("")
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAuthMiddleware_ValidHeaders tests successful authentication with valid headers
func TestAuthMiddleware_ValidHeaders(t *testing.T) {
	router := setupTestRouter("test-key-123")

	testCases := []struct {
		name        string
		setupHeader func(*http.Request)
	}{
		{
			name: "X-API-Key",
			setupHeader: func(req *http.Request) {
				req.Header.Set("X-API-Key", "test-key-123")
			},
		},
		{
			name: "Authorization_Bearer",
			setupHeader: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer test-key-123")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			tc.setupHeader(req)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// TestAuthMiddleware_InvalidHeaders tests authentication failures
func TestAuthMiddleware_InvalidHeaders(t *testing.T) {
	router := setupTestRouter("test-key-123")

	testCases := []struct {
		name        string
		setupHeader func(*http.Request)
	}{
		{
			name:        "missing_key",
			setupHeader: func(_ *http.Request) {}, // No header set
		},
		{
			name: "invalid_key",
			setupHeader: func(req *http.Request) {
				req.Header.Set("X-API-Key", "wrong-key")
			},
		},
		{
			name: "invalid_bearer_token",
			setupHeader: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer wrong-key")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			tc.setupHeader(req)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

// TestAuthMiddleware_HealthEndpoint tests that health endpoint bypasses authentication
func TestAuthMiddleware_HealthEndpoint(t *testing.T) {
	router := setupTestRouter("test-key-123")
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAuthMiddleware_BearerCaseInsensitive tests case-insensitivity of Bearer scheme
func TestAuthMiddleware_BearerCaseInsensitive(t *testing.T) {
	router := setupTestRouter("test-key-123")

	testCases := []string{
		"Bearer test-key-123",
		"bearer test-key-123",
		"BEARER test-key-123",
		"BeArEr test-key-123",
	}

	for _, authHeader := range testCases {
		t.Run(authHeader, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", authHeader)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Failed for header: %s", authHeader)
		})
	}
}

// TestAuthMiddleware_XAPIKeyCaseInsensitive tests case-insensitivity of x-api-key header
func TestAuthMiddleware_XAPIKeyCaseInsensitive(t *testing.T) {
	router := setupTestRouter("test-key-123")

	testCases := []struct {
		headerName string
		key        string
	}{
		{"X-API-Key", "test-key-123"},
		{"x-api-key", "test-key-123"},
		{"X-Api-Key", "test-key-123"},
		{"x-Api-Key", "test-key-123"},
	}

	for _, tc := range testCases {
		t.Run(tc.headerName, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set(tc.headerName, tc.key)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Failed for header: %s", tc.headerName)
		})
	}
}

// TestAuthMiddleware_InvalidBearerFormat tests invalid Authorization header formats
func TestAuthMiddleware_InvalidBearerFormat(t *testing.T) {
	router := setupTestRouter("test-key-123")

	testCases := []struct {
		name        string
		authHeader  string
		description string
	}{
		{
			name:        "missing_token",
			authHeader:  "Bearer",
			description: "Missing token",
		},
		{
			name:        "empty_token",
			authHeader:  "Bearer ",
			description: "Empty token",
		},
		{
			name:        "wrong_scheme",
			authHeader:  "Basic test-key-123",
			description: "Wrong scheme",
		},
		{
			name:        "no_scheme",
			authHeader:  "test-key-123",
			description: "No scheme",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tc.authHeader)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code, "Should fail for: %s", tc.description)
		})
	}
}

// TestAuthMiddleware_ErrorResponseFormat verifies that error responses match OpenAI-compatible format
func TestAuthMiddleware_ErrorResponseFormat(t *testing.T) {
	router := setupTestRouter("test-key-123")

	testCases := []struct {
		name        string
		setupHeader func(*http.Request)
	}{
		{
			name:        "missing_key",
			setupHeader: func(_ *http.Request) {}, // No header set
		},
		{
			name: "invalid_key",
			setupHeader: func(req *http.Request) {
				req.Header.Set("X-API-Key", "wrong-key")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			tc.setupHeader(req)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err, "Error response should be valid JSON")

			// Verify OpenAI-compatible error structure
			assert.Contains(t, response, "error", "Response must contain 'error' field")
			errorObj, ok := response["error"].(map[string]interface{})
			assert.True(t, ok, "Error should be an object")

			assert.Contains(t, errorObj, "message", "Error must contain 'message' field")
			assert.Contains(t, errorObj, "type", "Error must contain 'type' field")
			assert.Equal(t, "invalid_request_error", errorObj["type"], "Error type should be 'invalid_request_error'")
			assert.NotEmpty(t, errorObj["message"], "Error message should not be empty")
		})
	}
}
