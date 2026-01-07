package mcpclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// HealthStatus represents the health status of an MCP server
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// String returns the string representation of the health status
func (hs HealthStatus) String() string {
	switch hs {
	case HealthStatusHealthy:
		return "healthy"
	case HealthStatusUnhealthy:
		return "unhealthy"
	case HealthStatusUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Status      HealthStatus
	LastCheck   time.Time
	LastSuccess time.Time
	LastError   error
	Latency     time.Duration
}

// HealthMonitor monitors the health of MCP clients
type HealthMonitor struct {
	clients  map[string]*Client
	results  map[string]*HealthCheckResult
	mu       sync.RWMutex
	interval time.Duration
	timeout  time.Duration
	stopChan chan struct{}
	stopped  bool
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(interval, timeout time.Duration) *HealthMonitor {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &HealthMonitor{
		clients:  make(map[string]*Client),
		results:  make(map[string]*HealthCheckResult),
		interval: interval,
		timeout:  timeout,
		stopChan: make(chan struct{}),
	}
}

// RegisterClient registers a client for health monitoring
func (hm *HealthMonitor) RegisterClient(name string, client *Client) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.clients[name] = client
	hm.results[name] = &HealthCheckResult{
		Status:    HealthStatusUnknown,
		LastCheck: time.Time{},
	}
}

// UnregisterClient unregisters a client from health monitoring
func (hm *HealthMonitor) UnregisterClient(name string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	delete(hm.clients, name)
	delete(hm.results, name)
}

// Start starts the health monitoring goroutine
func (hm *HealthMonitor) Start() {
	hm.mu.Lock()
	if hm.stopped {
		hm.mu.Unlock()
		return
	}
	hm.mu.Unlock()

	go hm.monitorLoop()
}

// Stop stops the health monitoring
func (hm *HealthMonitor) Stop() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.stopped {
		return
	}

	hm.stopped = true
	close(hm.stopChan)
}

// CheckHealth performs a health check on a specific client
func (hm *HealthMonitor) CheckHealth(ctx context.Context, name string) (*HealthCheckResult, error) {
	hm.mu.RLock()
	client, exists := hm.clients[name]
	hm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("client not found: %s", name)
	}

	start := time.Now()
	result := &HealthCheckResult{
		LastCheck: time.Now(),
	}

	// Create context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, hm.timeout)
	defer cancel()

	// Try to list tools as a health check
	// This is a lightweight operation that verifies the connection works
	_, err := client.transport.SendRequest(checkCtx, "tools/list", nil)
	result.Latency = time.Since(start)

	if err != nil {
		result.Status = HealthStatusUnhealthy
		result.LastError = err
	} else {
		result.Status = HealthStatusHealthy
		result.LastSuccess = time.Now()
		result.LastError = nil
	}

	// Update result
	hm.mu.Lock()
	hm.results[name] = result
	hm.mu.Unlock()

	return result, nil
}

// GetHealth returns the current health status for a client
func (hm *HealthMonitor) GetHealth(name string) (*HealthCheckResult, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	result, exists := hm.results[name]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	return &HealthCheckResult{
		Status:      result.Status,
		LastCheck:   result.LastCheck,
		LastSuccess: result.LastSuccess,
		LastError:   result.LastError,
		Latency:     result.Latency,
	}, true
}

// GetAllHealth returns health status for all monitored clients
func (hm *HealthMonitor) GetAllHealth() map[string]*HealthCheckResult {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	results := make(map[string]*HealthCheckResult, len(hm.results))
	for name, result := range hm.results {
		// Return a copy to avoid race conditions
		results[name] = &HealthCheckResult{
			Status:      result.Status,
			LastCheck:   result.LastCheck,
			LastSuccess: result.LastSuccess,
			LastError:   result.LastError,
			Latency:     result.Latency,
		}
	}

	return results
}

// monitorLoop runs the periodic health check loop
func (hm *HealthMonitor) monitorLoop() {
	ticker := time.NewTicker(hm.interval)
	defer ticker.Stop()

	// Perform initial health check
	hm.performHealthChecks()

	for {
		select {
		case <-hm.stopChan:
			return
		case <-ticker.C:
			hm.performHealthChecks()
		}
	}
}

// performHealthChecks performs health checks on all registered clients
func (hm *HealthMonitor) performHealthChecks() {
	hm.mu.RLock()
	clients := make(map[string]*Client, len(hm.clients))
	for name, client := range hm.clients {
		clients[name] = client
	}
	hm.mu.RUnlock()

	ctx := context.Background()
	for name, client := range clients {
		// Skip if client is closed
		if client.IsClosed() {
			hm.mu.Lock()
			result := hm.results[name]
			if result != nil {
				result.Status = HealthStatusUnhealthy
				result.LastCheck = time.Now()
				result.LastError = ErrConnectionClosed
			}
			hm.mu.Unlock()
			continue
		}

		_, err := hm.CheckHealth(ctx, name)
		if err != nil {
			log.Debug().
				Str("server", name).
				Err(err).
				Msg("Health check failed")
		}
	}
}
