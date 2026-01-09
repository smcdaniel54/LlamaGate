package middleware

import (
	"testing"
)

func TestIsHealthEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "exact match",
			path:     "/health",
			expected: true,
		},
		{
			name:     "single trailing slash",
			path:     "/health/",
			expected: true,
		},
		{
			name:     "multiple trailing slashes",
			path:     "/health//",
			expected: true,
		},
		{
			name:     "many trailing slashes",
			path:     "/health///",
			expected: true,
		},
		{
			name:     "not health endpoint",
			path:     "/v1/models",
			expected: false,
		},
		{
			name:     "health in path but not exact",
			path:     "/v1/health",
			expected: false,
		},
		{
			name:     "health with query string",
			path:     "/health?check=true",
			expected: false, // Query string should be handled separately
		},
		{
			name:     "empty path",
			path:     "",
			expected: false,
		},
		{
			name:     "root path",
			path:     "/",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHealthEndpoint(tt.path)
			if result != tt.expected {
				t.Errorf("isHealthEndpoint(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}
