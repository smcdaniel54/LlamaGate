// Package middleware provides HTTP middleware for authentication, rate limiting, and request tracking.
package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/response"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware validates API key from X-API-Key header or Authorization Bearer token
func AuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health check endpoint
		if isHealthEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Skip auth if API key is not configured
		if apiKey == "" {
			c.Next()
			return
		}

		// Check X-API-Key header first
		providedKey := c.GetHeader("X-API-Key")
		if providedKey == "" {
			// Check Authorization Bearer header
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
					providedKey = parts[1]
				}
			}
		}

		// Validate the key using constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(providedKey), []byte(apiKey)) != 1 {
			requestID := GetRequestID(c)
			log.Warn().
				Str("request_id", requestID).
				Str("ip", c.ClientIP()).
				Msg("Authentication failed")
			response.ErrorResponse(c, http.StatusUnauthorized, response.ErrorTypeInvalidRequest, "Invalid API key", requestID)
			c.Abort()
			return
		}

		c.Next()
	}
}
