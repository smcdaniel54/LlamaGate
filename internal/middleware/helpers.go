// Package middleware provides HTTP middleware for authentication, rate limiting, and request tracking.
package middleware

import "strings"

const (
	// HealthEndpointPath is the path for the health check endpoint that should be excluded from authentication and rate limiting
	HealthEndpointPath = "/health"
)

// isHealthEndpoint checks if the request path is the health check endpoint
// It normalizes the path by removing trailing slashes to handle edge cases
func isHealthEndpoint(path string) bool {
	// Normalize path: remove trailing slash and compare
	normalized := strings.TrimSuffix(path, "/")
	return normalized == HealthEndpointPath
}

