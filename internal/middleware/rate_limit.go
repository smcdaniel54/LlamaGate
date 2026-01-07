// Package middleware provides HTTP middleware for authentication, rate limiting, and request tracking.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/response"
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
		// Skip rate limiting for health check endpoint
		if isHealthEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Check if request is allowed
		if !rl.limiter.Allow() {
			requestID := GetRequestID(c)
			response.RateLimitExceeded(c, "Rate limit exceeded", requestID)
			c.Abort()
			return
		}

		c.Next()
	}
}
