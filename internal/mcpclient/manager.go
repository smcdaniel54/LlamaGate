package mcpclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// ServerManager manages MCP server connections with pooling, health monitoring, and caching
type ServerManager struct {
	servers           map[string]*ManagedServer
	pools             map[string]*ConnectionPool
	activeConnections map[*Client]*PooledConnection // Track active pooled connections
	healthMonitor     *HealthMonitor
	cache             *Cache
	mu                sync.RWMutex
	config            ManagerConfig
}

// ManagerConfig holds configuration for the server manager
type ManagerConfig struct {
	PoolSize       int
	PoolIdleTime   time.Duration
	HealthInterval time.Duration
	HealthTimeout  time.Duration
	CacheTTL       time.Duration
}

// ManagedServer holds information about a managed server
type ManagedServer struct {
	Name      string
	Client    *Client
	Pool      *ConnectionPool
	Transport string
}

// NewServerManager creates a new server manager
func NewServerManager(config ManagerConfig) *ServerManager {
	if config.PoolSize <= 0 {
		config.PoolSize = 10
	}
	if config.PoolIdleTime <= 0 {
		config.PoolIdleTime = 5 * time.Minute
	}
	if config.HealthInterval <= 0 {
		config.HealthInterval = 60 * time.Second
	}
	if config.HealthTimeout <= 0 {
		config.HealthTimeout = 5 * time.Second
	}
	if config.CacheTTL <= 0 {
		config.CacheTTL = 5 * time.Minute
	}

	manager := &ServerManager{
		servers:           make(map[string]*ManagedServer),
		pools:             make(map[string]*ConnectionPool),
		activeConnections: make(map[*Client]*PooledConnection),
		healthMonitor:     NewHealthMonitor(config.HealthInterval, config.HealthTimeout),
		cache:             NewCache(config.CacheTTL),
		config:            config,
	}

	// Start health monitoring
	manager.healthMonitor.Start()

	return manager
}

// AddServer adds a server to the manager
func (sm *ServerManager) AddServer(name string, client *Client, transport string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.servers[name]; exists {
		return fmt.Errorf("server %s already registered", name)
	}

	// Create connection pool for HTTP transport (stdio doesn't need pooling)
	var pool *ConnectionPool
	if transport == "http" {
		poolConfig := PoolConfig{
			MaxConnections: sm.config.PoolSize,
			MaxIdleTime:    sm.config.PoolIdleTime,
			AcquireTimeout: 30 * time.Second,
		}
		pool = NewConnectionPool(poolConfig)
	}

	serverInfo := &ManagedServer{
		Name:      name,
		Client:    client,
		Pool:      pool,
		Transport: transport,
	}

	sm.servers[name] = serverInfo
	if pool != nil {
		sm.pools[name] = pool
	}

	// Register with health monitor
	sm.healthMonitor.RegisterClient(name, client)

	log.Info().
		Str("server", name).
		Str("transport", transport).
		Msg("Server added to manager")

	return nil
}

// RemoveServer removes a server from the manager
func (sm *ServerManager) RemoveServer(name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	serverInfo, exists := sm.servers[name]
	if !exists {
		return fmt.Errorf("server %s not found", name)
	}

	// Unregister from health monitor
	sm.healthMonitor.UnregisterClient(name)

	// Close pool if exists
	if serverInfo.Pool != nil {
		serverInfo.Pool.Close()
		delete(sm.pools, name)
	}

	// Close client
	if err := serverInfo.Client.Close(); err != nil {
		log.Warn().
			Str("server", name).
			Err(err).
			Msg("Error closing server client")
	}

	// Invalidate cache
	sm.cache.InvalidateAll(name)

	delete(sm.servers, name)

	log.Info().
		Str("server", name).
		Msg("Server removed from manager")

	return nil
}

// GetServer gets a server by name
func (sm *ServerManager) GetServer(name string) (*ManagedServer, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	serverInfo, exists := sm.servers[name]
	if !exists {
		return nil, fmt.Errorf("server %s not found", name)
	}

	return serverInfo, nil
}

// ListServers returns all server names
func (sm *ServerManager) ListServers() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	names := make([]string, 0, len(sm.servers))
	for name := range sm.servers {
		names = append(names, name)
	}

	return names
}

// GetClient gets a client for a server (with pooling for HTTP)
func (sm *ServerManager) GetClient(ctx context.Context, name string) (*Client, error) {
	sm.mu.RLock()
	serverInfo, exists := sm.servers[name]
	sm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("server %s not found", name)
	}

	// For HTTP transport, use connection pool
	if serverInfo.Pool != nil {
		pooledConn, err := serverInfo.Pool.Acquire(ctx, func() (*Client, error) {
			// Factory function to create new client
			// For now, we'll reuse the existing client
			// In a real implementation, you'd create a new connection here
			return serverInfo.Client, nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to acquire connection from pool: %w", err)
		}
		client := pooledConn.GetClient()

		// Track the active connection so we can release it later
		sm.mu.Lock()
		sm.activeConnections[client] = pooledConn
		sm.mu.Unlock()

		return client, nil
	}

	// For stdio transport, return the client directly
	return serverInfo.Client, nil
}

// ReleaseClient releases a pooled connection
func (sm *ServerManager) ReleaseClient(name string, client *Client) {
	if client == nil {
		return
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	serverInfo, exists := sm.servers[name]
	if !exists || serverInfo.Pool == nil {
		return
	}

	// Find and release the pooled connection
	if pooledConn, ok := sm.activeConnections[client]; ok {
		serverInfo.Pool.Release(pooledConn)
		delete(sm.activeConnections, client)
		log.Debug().
			Str("server", name).
			Msg("Released pooled connection")
	}
}

// GetHealth returns health status for a server
func (sm *ServerManager) GetHealth(name string) (*HealthCheckResult, bool) {
	return sm.healthMonitor.GetHealth(name)
}

// GetAllHealth returns health status for all servers
func (sm *ServerManager) GetAllHealth() map[string]*HealthCheckResult {
	return sm.healthMonitor.GetAllHealth()
}

// CheckHealth performs a health check on a server
func (sm *ServerManager) CheckHealth(ctx context.Context, name string) (*HealthCheckResult, error) {
	return sm.healthMonitor.CheckHealth(ctx, name)
}

// GetCache returns the cache instance
func (sm *ServerManager) GetCache() *Cache {
	return sm.cache
}

// Close closes all servers and stops monitoring
func (sm *ServerManager) Close() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Stop health monitoring
	sm.healthMonitor.Stop()

	// Stop cache cleanup goroutine
	sm.cache.Stop()

	// Close all pools
	for name, pool := range sm.pools {
		if err := pool.Close(); err != nil {
			log.Warn().
				Str("server", name).
				Err(err).
				Msg("Error closing connection pool")
		}
	}

	// Close all clients
	var firstErr error
	for name, serverInfo := range sm.servers {
		if err := serverInfo.Client.Close(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			log.Warn().
				Str("server", name).
				Err(err).
				Msg("Error closing server client")
		}
	}

	// Clear active connections map
	sm.activeConnections = nil
	sm.servers = nil
	sm.pools = nil

	return firstErr
}
