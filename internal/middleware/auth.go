package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware validates API key from X-API-Key header or Authorization Bearer token
func AuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
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
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					providedKey = parts[1]
				}
			}
		}

		// Validate the key using constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(providedKey), []byte(apiKey)) != 1 {
			log.Warn().
				Str("request_id", c.GetString("request_id")).
				Str("ip", c.ClientIP()).
				Msg("Authentication failed")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "Invalid API key",
					"type":    "invalid_request_error",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

