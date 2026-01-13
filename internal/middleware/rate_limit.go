// Package middleware provides HTTP middleware for authentication, rate limiting, and request tracking.
package middleware

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/response"
	"github.com/rs/zerolog/log"
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

		// Check if request is allowed using Reserve to get retry time
		reservation := rl.limiter.Reserve()
		if !reservation.OK() {
			// Limiter is closed or invalid - should not happen in normal operation
			requestID := GetRequestID(c)
			log.Error().
				Str("request_id", requestID).
				Str("ip", c.ClientIP()).
				Str("path", c.Request.URL.Path).
				Msg("Rate limiter returned invalid reservation")
			response.RateLimitExceeded(c, "Rate limit exceeded", requestID)
			c.Abort()
			return
		}

		// Check if reservation allows immediate access
		delay := reservation.Delay()
		if delay > 0 {
			// Request is rate limited
			reservation.Cancel() // Cancel the reservation since we're rejecting the request
			requestID := GetRequestID(c)

			// Calculate Retry-After header value (in seconds, rounded up)
			retryAfter := int(math.Ceil(delay.Seconds()))
			if retryAfter < 1 {
				retryAfter = 1 // Minimum 1 second
			}
			c.Header("Retry-After", strconv.Itoa(retryAfter))

			// Log rate limit decision with structured fields
			log.Warn().
				Str("request_id", requestID).
				Str("ip", c.ClientIP()).
				Str("path", c.Request.URL.Path).
				Dur("retry_after", delay).
				Int("retry_after_seconds", retryAfter).
				Str("limiter_decision", "rate_limited").
				Msg("Rate limit exceeded")

			response.RateLimitExceeded(c, "Rate limit exceeded", requestID)
			c.Abort()
			return
		}

		// Request is allowed - reservation will be consumed when request completes
		// We don't need to explicitly cancel it as it will be consumed naturally

		c.Next()
	}
}
