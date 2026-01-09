// Package middleware provides HTTP middleware for authentication, rate limiting, and request tracking.
package middleware

import "strings"

const (
	// HealthEndpointPath is the path for the health check endpoint that should be excluded from authentication and rate limiting
	HealthEndpointPath = "/health"
)

// isHealthEndpoint checks if the request path is the health check endpoint
// It normalizes the path by removing all trailing slashes to handle edge cases
func isHealthEndpoint(path string) bool {
	// Normalize path: remove all trailing slashes and compare
	normalized := strings.TrimRight(path, "/")
	return normalized == HealthEndpointPath
}

