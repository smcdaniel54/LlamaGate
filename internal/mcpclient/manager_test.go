package mcpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServerManager(t *testing.T) {
	config := ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}

	manager := NewServerManager(config)
	assert.NotNil(t, manager)
	assert.Len(t, manager.ListServers(), 0)

	// Cleanup
	manager.Close()
}

func TestServerManager_AddRemoveServer(t *testing.T) {
	config := ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}

	manager := NewServerManager(config)
	defer manager.Close()

	// Create a mock client
	transport := newMockTransport()
	client := &Client{
		name:         "test-server",
		transport:    transport,
		toolsMap:     make(map[string]*Tool),
		resourcesMap: make(map[string]*Resource),
		promptsMap:   make(map[string]*Prompt),
	}

	// Add server
	err := manager.AddServer("test-server", client, "stdio")
	require.NoError(t, err)

	servers := manager.ListServers()
	assert.Len(t, servers, 1)
	assert.Contains(t, servers, "test-server")

	// Get server
	serverInfo, err := manager.GetServer("test-server")
	require.NoError(t, err)
	assert.Equal(t, "test-server", serverInfo.Name)
	assert.Equal(t, "stdio", serverInfo.Transport)
	assert.Nil(t, serverInfo.Pool) // stdio doesn't use pooling

	// Remove server
	err = manager.RemoveServer("test-server")
	require.NoError(t, err)

	servers = manager.ListServers()
	assert.Len(t, servers, 0)
}

func TestServerManager_HTTPTransportWithPool(t *testing.T) {
	config := ManagerConfig{
		PoolSize:       2,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}

	manager := NewServerManager(config)
	defer manager.Close()

	// Create a mock HTTP server for testing
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req JSONRPCRequest
		json.NewDecoder(r.Body).Decode(&req)
		resp := JSONRPCResponse{
			JSONRPC: JSONRPCVersion,
			ID:      req.ID,
			Result:  json.RawMessage(`{"protocolVersion":"2024-11-05","capabilities":{},"serverInfo":{"name":"test-server","version":"1.0.0"}}`),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create HTTP client with real HTTPTransport
	client, err := NewClientWithHTTP("http-server", server.URL, nil, 30*time.Second)
	require.NoError(t, err)

	// Add HTTP server (should create pool)
	err = manager.AddServer("http-server", client, "http")
	require.NoError(t, err)

	serverInfo, err := manager.GetServer("http-server")
	require.NoError(t, err)
	assert.NotNil(t, serverInfo.Pool) // HTTP should have a pool

	// Test getting client (should use pool and create new client instances)
	ctx := context.Background()
	acquiredClient1, err := manager.GetClient(ctx, "http-server")
	require.NoError(t, err)
	assert.NotNil(t, acquiredClient1)
	// Verify it's a different instance (not the same pointer) - pool should create new clients
	// Compare pointers explicitly to ensure they're different instances
	assert.NotSame(t, client, acquiredClient1, "Pool should return a new client instance, not reuse the same one")

	// Release the first client
	manager.ReleaseClient("http-server", acquiredClient1)

	// Get another client - pool might reuse the released one or create a new one
	acquiredClient2, err := manager.GetClient(ctx, "http-server")
	require.NoError(t, err)
	assert.NotNil(t, acquiredClient2)
	// Should be different from original client (the one passed to AddServer)
	assert.NotSame(t, client, acquiredClient2, "Pool should not return the original client instance")

	// Get multiple clients concurrently to verify pool creates multiple instances
	acquiredClient3, err := manager.GetClient(ctx, "http-server")
	require.NoError(t, err)
	assert.NotNil(t, acquiredClient3)
	assert.NotSame(t, client, acquiredClient3, "Pool should not return the original client instance")

	// Verify at least one of the acquired clients is different from the others
	// (proving the pool is creating new instances, not always returning the same one)
	allDifferent := acquiredClient1 != acquiredClient2 || acquiredClient1 != acquiredClient3 || acquiredClient2 != acquiredClient3
	assert.True(t, allDifferent, "Pool should create multiple different client instances when needed")
}

func TestServerManager_HealthMonitoring(t *testing.T) {
	config := ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 100 * time.Millisecond, // Fast for testing
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}

	manager := NewServerManager(config)
	defer manager.Close()

	// Create a mock client with successful health check
	transport := newMockTransport()
	transport.responseFunc = func(method string, _ interface{}) (*JSONRPCResponse, error) {
		if method == "tools/list" {
			result := ToolsListResult{Tools: []Tool{}}
			resultJSON, _ := json.Marshal(result)
			return &JSONRPCResponse{
				JSONRPC: JSONRPCVersion,
				ID:      1,
				Result:  resultJSON,
			}, nil
		}
		return &JSONRPCResponse{
			JSONRPC: JSONRPCVersion,
			ID:      1,
			Result:  json.RawMessage(`{}`),
		}, nil
	}

	client := &Client{
		name:         "test-server",
		transport:    transport,
		toolsMap:     make(map[string]*Tool),
		resourcesMap: make(map[string]*Resource),
		promptsMap:   make(map[string]*Prompt),
	}

	err := manager.AddServer("test-server", client, "stdio")
	require.NoError(t, err)

	// Wait a bit for health check to run
	time.Sleep(150 * time.Millisecond)

	// Check health
	health, exists := manager.GetHealth("test-server")
	require.True(t, exists)
	assert.NotZero(t, health.LastCheck)

	// Get all health
	allHealth := manager.GetAllHealth()
	assert.Contains(t, allHealth, "test-server")
}

func TestServerManager_Caching(t *testing.T) {
	config := ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}

	manager := NewServerManager(config)
	defer manager.Close()

	cache := manager.GetCache()
	assert.NotNil(t, cache)

	// Test caching tools
	tools := []Tool{
		{Name: "tool1", Description: "Test tool 1"},
	}
	cache.SetTool("test-server", tools)

	cached, found := cache.GetTool("test-server")
	assert.True(t, found)
	assert.Len(t, cached, 1)

	// Invalidate
	cache.InvalidateAll("test-server")
	_, found = cache.GetTool("test-server")
	assert.False(t, found)
}

func TestServerManager_DuplicateServer(t *testing.T) {
	config := ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}

	manager := NewServerManager(config)
	defer manager.Close()

	transport := newMockTransport()
	client := &Client{
		name:         "test-server",
		transport:    transport,
		toolsMap:     make(map[string]*Tool),
		resourcesMap: make(map[string]*Resource),
		promptsMap:   make(map[string]*Prompt),
	}

	err := manager.AddServer("test-server", client, "stdio")
	require.NoError(t, err)

	// Try to add duplicate
	err = manager.AddServer("test-server", client, "stdio")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestServerManager_GetNonExistentServer(t *testing.T) {
	config := ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}

	manager := NewServerManager(config)
	defer manager.Close()

	_, err := manager.GetServer("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestServerManager_Close(t *testing.T) {
	config := ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}

	manager := NewServerManager(config)

	transport := newMockTransport()
	client := &Client{
		name:         "test-server",
		transport:    transport,
		toolsMap:     make(map[string]*Tool),
		resourcesMap: make(map[string]*Resource),
		promptsMap:   make(map[string]*Prompt),
	}

	err := manager.AddServer("test-server", client, "stdio")
	require.NoError(t, err)

	// Close manager
	err = manager.Close()
	require.NoError(t, err)

	// Should not be able to get server after close
	servers := manager.ListServers()
	assert.Len(t, servers, 0)
}
