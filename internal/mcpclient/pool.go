package mcpclient

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	ErrPoolExhausted = errors.New("connection pool exhausted")
	ErrPoolClosed    = errors.New("connection pool closed")
)

// PoolConfig holds configuration for a connection pool
type PoolConfig struct {
	MaxConnections int           // Maximum number of connections in the pool
	MaxIdleTime    time.Duration // Maximum time a connection can be idle
	AcquireTimeout time.Duration // Timeout for acquiring a connection
}

// DefaultPoolConfig returns a default pool configuration
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxConnections: 10,
		MaxIdleTime:    5 * time.Minute,
		AcquireTimeout: 30 * time.Second,
	}
}

// PooledConnection represents a connection in the pool
type PooledConnection struct {
	client      *Client
	lastUsed    time.Time
	inUse       bool
	mu          sync.Mutex
}

// ConnectionPool manages a pool of MCP client connections
type ConnectionPool struct {
	config      PoolConfig
	connections []*PooledConnection
	mu          sync.RWMutex
	closed      bool
	cleanupOnce sync.Once
	cleanupStop chan struct{}
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config PoolConfig) *ConnectionPool {
	if config.MaxConnections <= 0 {
		config.MaxConnections = 10
	}
	if config.MaxIdleTime <= 0 {
		config.MaxIdleTime = 5 * time.Minute
	}
	if config.AcquireTimeout <= 0 {
		config.AcquireTimeout = 30 * time.Second
	}

	pool := &ConnectionPool{
		config:      config,
		connections: make([]*PooledConnection, 0),
		cleanupStop: make(chan struct{}),
	}

	// Start cleanup goroutine
	pool.cleanupOnce.Do(func() {
		go pool.cleanupIdleConnections()
	})

	return pool
}

// Acquire acquires a connection from the pool
// If no connection is available and pool is not full, creates a new one
// If pool is full, waits for a connection to become available
func (p *ConnectionPool) Acquire(ctx context.Context, factory func() (*Client, error)) (*PooledConnection, error) {
	deadline := time.Now().Add(p.config.AcquireTimeout)
	if deadlineCtx, ok := ctx.Deadline(); ok && deadlineCtx.Before(deadline) {
		deadline = deadlineCtx
	}

	for {
		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			return nil, ErrPoolClosed
		}

		// Try to find an available connection
		for _, conn := range p.connections {
			conn.mu.Lock()
			if !conn.inUse && !conn.client.IsClosed() {
				conn.inUse = true
				conn.lastUsed = time.Now()
				conn.mu.Unlock()
				p.mu.Unlock()
				return conn, nil
			}
			conn.mu.Unlock()
		}

		// If pool is not full, create a new connection
		// Check must be done while holding the lock to prevent race conditions
		if len(p.connections) < p.config.MaxConnections {
			// Reserve a slot by checking the condition while holding the lock
			// Unlock only for the expensive factory() call
			p.mu.Unlock()
			client, err := factory()
			if err != nil {
				return nil, err
			}

			conn := &PooledConnection{
				client:   client,
				lastUsed: time.Now(),
				inUse:    true,
			}

			// Re-acquire lock and verify we can still add the connection
			p.mu.Lock()
			if p.closed {
				p.mu.Unlock()
				client.Close()
				return nil, ErrPoolClosed
			}
			// Double-check: another goroutine might have filled the pool while we were creating the connection
			if len(p.connections) >= p.config.MaxConnections {
				p.mu.Unlock()
				client.Close()
				// Retry the loop to wait for an available connection
				continue
			}
			p.connections = append(p.connections, conn)
			p.mu.Unlock()

			return conn, nil
		}

		p.mu.Unlock()

		// Pool is full, wait a bit and retry
		if time.Now().After(deadline) {
			return nil, ErrPoolExhausted
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			// Retry
		}
	}
}

// Release releases a connection back to the pool
func (p *ConnectionPool) Release(conn *PooledConnection) {
	if conn == nil {
		return
	}

	conn.mu.Lock()
	conn.inUse = false
	conn.lastUsed = time.Now()
	conn.mu.Unlock()
}

// Remove removes a connection from the pool (e.g., if it's broken)
func (p *ConnectionPool) Remove(conn *PooledConnection) {
	if conn == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for i, c := range p.connections {
		if c == conn {
			// Close the connection
			conn.client.Close()

			// Remove from slice
			p.connections = append(p.connections[:i], p.connections[i+1:]...)
			return
		}
	}
}

// Close closes all connections in the pool
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	close(p.cleanupStop)

	var firstErr error
	for _, conn := range p.connections {
		if err := conn.client.Close(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	p.connections = nil
	return firstErr
}

// Stats returns pool statistics
func (p *ConnectionPool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := PoolStats{
		Total:      len(p.connections),
		InUse:      0,
		Idle:       0,
		MaxAllowed: p.config.MaxConnections,
	}

	for _, conn := range p.connections {
		conn.mu.Lock()
		if conn.inUse {
			stats.InUse++
		} else {
			stats.Idle++
		}
		conn.mu.Unlock()
	}

	return stats
}

// PoolStats represents pool statistics
type PoolStats struct {
	Total      int
	InUse      int
	Idle       int
	MaxAllowed int
}

// cleanupIdleConnections periodically removes idle connections
func (p *ConnectionPool) cleanupIdleConnections() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-p.cleanupStop:
			return
		case <-ticker.C:
			p.cleanup()
		}
	}
}

// cleanup removes idle connections that have exceeded MaxIdleTime
func (p *ConnectionPool) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	now := time.Now()
	toRemove := make([]*PooledConnection, 0)

	for _, conn := range p.connections {
		conn.mu.Lock()
		if !conn.inUse && now.Sub(conn.lastUsed) > p.config.MaxIdleTime {
			toRemove = append(toRemove, conn)
		}
		conn.mu.Unlock()
	}

	// Remove idle connections
	for _, conn := range toRemove {
		for i, c := range p.connections {
			if c == conn {
				conn.client.Close()
				p.connections = append(p.connections[:i], p.connections[i+1:]...)
				log.Debug().
					Str("server", conn.client.GetName()).
					Msg("Removed idle connection from pool")
				break
			}
		}
	}
}

// GetClient returns the underlying client from a pooled connection
func (pc *PooledConnection) GetClient() *Client {
	return pc.client
}

