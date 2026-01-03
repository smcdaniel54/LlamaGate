// Package middleware provides HTTP middleware for authentication, rate limiting, and request tracking.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitMiddleware implements rate limiting using a leaky bucket algorithm
type RateLimitMiddleware struct {
	limiter *rate.Limiter
}

// NewRateLimitMiddleware creates a new rate limiter with the specified requests per second
func NewRateLimitMiddleware(rps float64) *RateLimitMiddleware {
	// Create a limiter that allows rps requests per second
	// Burst is set to rps to allow some burst traffic
	burst := int(rps)
	if burst < 1 {
		burst = 1
	}
	limiter := rate.NewLimiter(rate.Limit(rps), burst)

	return &RateLimitMiddleware{
		limiter: limiter,
	}
}

// Handler returns a Gin middleware handler for rate limiting
func (rl *RateLimitMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request is allowed
		if !rl.limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"message": "Rate limit exceeded",
					"type":    "rate_limit_error",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
