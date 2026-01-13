package mcpclient

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConnectionPool(t *testing.T) {
	config := DefaultPoolConfig()
	pool := NewConnectionPool(config)

	assert.NotNil(t, pool)
	stats := pool.Stats()
	assert.Equal(t, 0, stats.Total)
	assert.Equal(t, config.MaxConnections, stats.MaxAllowed)

	// Cleanup
	pool.Close()
}

func TestConnectionPool_Acquire(t *testing.T) {
	config := PoolConfig{
		MaxConnections: 2,
		MaxIdleTime:    1 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := NewConnectionPool(config)

	// Create a factory function
	callCount := 0
	factory := func() (*Client, error) {
		callCount++
		transport := newMockTransport()
		client := &Client{
			name:         "test",
			transport:    transport,
			toolsMap:     make(map[string]*Tool),
			resourcesMap: make(map[string]*Resource),
			promptsMap:   make(map[string]*Prompt),
		}
		return client, nil
	}

	ctx := context.Background()

	// Acquire first connection
	conn1, err := pool.Acquire(ctx, factory)
	require.NoError(t, err)
	assert.NotNil(t, conn1)
	assert.Equal(t, 1, callCount)

	stats := pool.Stats()
	assert.Equal(t, 1, stats.Total)
	assert.Equal(t, 1, stats.InUse)
	assert.Equal(t, 0, stats.Idle)

	// Acquire second connection
	conn2, err := pool.Acquire(ctx, factory)
	require.NoError(t, err)
	assert.NotNil(t, conn2)
	assert.Equal(t, 2, callCount)

	stats = pool.Stats()
	assert.Equal(t, 2, stats.Total)
	assert.Equal(t, 2, stats.InUse)

	// Release connections
	pool.Release(conn1)
	pool.Release(conn2)

	stats = pool.Stats()
	assert.Equal(t, 2, stats.Total)
	assert.Equal(t, 0, stats.InUse)
	assert.Equal(t, 2, stats.Idle)

	// Cleanup
	pool.Close()
}

func TestConnectionPool_MaxConnections(t *testing.T) {
	config := PoolConfig{
		MaxConnections: 1,
		MaxIdleTime:    1 * time.Minute,
		AcquireTimeout: 100 * time.Millisecond,
	}
	pool := NewConnectionPool(config)

	callCount := 0
	factory := func() (*Client, error) {
		callCount++
		transport := newMockTransport()
		client := &Client{
			name:         "test",
			transport:    transport,
			toolsMap:     make(map[string]*Tool),
			resourcesMap: make(map[string]*Resource),
			promptsMap:   make(map[string]*Prompt),
		}
		return client, nil
	}

	ctx := context.Background()

	// Acquire first connection
	conn1, err := pool.Acquire(ctx, factory)
	require.NoError(t, err)

	// Try to acquire second connection (should timeout)
	ctx2, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	_, err = pool.Acquire(ctx2, factory)
	assert.Error(t, err)
	assert.Equal(t, ErrPoolExhausted, err)

	// Release first connection
	pool.Release(conn1)

	// Now should be able to acquire
	conn2, err := pool.Acquire(ctx, factory)
	require.NoError(t, err)
	assert.NotNil(t, conn2)

	pool.Release(conn2)
	pool.Close()
}

func TestConnectionPool_Remove(t *testing.T) {
	config := DefaultPoolConfig()
	pool := NewConnectionPool(config)

	factory := func() (*Client, error) {
		transport := newMockTransport()
		client := &Client{
			name:         "test",
			transport:    transport,
			toolsMap:     make(map[string]*Tool),
			resourcesMap: make(map[string]*Resource),
			promptsMap:   make(map[string]*Prompt),
		}
		return client, nil
	}

	ctx := context.Background()
	conn, err := pool.Acquire(ctx, factory)
	require.NoError(t, err)

	stats := pool.Stats()
	assert.Equal(t, 1, stats.Total)

	pool.Remove(conn)

	stats = pool.Stats()
	assert.Equal(t, 0, stats.Total)

	pool.Close()
}

func TestConnectionPool_Close(t *testing.T) {
	config := DefaultPoolConfig()
	pool := NewConnectionPool(config)

	factory := func() (*Client, error) {
		transport := newMockTransport()
		client := &Client{
			name:         "test",
			transport:    transport,
			toolsMap:     make(map[string]*Tool),
			resourcesMap: make(map[string]*Resource),
			promptsMap:   make(map[string]*Prompt),
		}
		return client, nil
	}

	ctx := context.Background()
	conn, err := pool.Acquire(ctx, factory)
	require.NoError(t, err)

	err = pool.Close()
	require.NoError(t, err)

	// Try to acquire after close
	_, err = pool.Acquire(ctx, factory)
	assert.Error(t, err)
	assert.Equal(t, ErrPoolClosed, err)

	_ = conn // Avoid unused variable
}

// TestConnectionPool_Acquire_RaceCondition tests that concurrent Acquire calls
// don't exceed MaxConnections limit
func TestConnectionPool_Acquire_RaceCondition(t *testing.T) {
	config := PoolConfig{
		MaxConnections: 5,
		MaxIdleTime:    1 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := NewConnectionPool(config)
	defer pool.Close()

	callCount := 0
	var callCountMu sync.Mutex
	factory := func() (*Client, error) {
		callCountMu.Lock()
		callCount++
		callCountMu.Unlock()
		// Simulate connection creation delay to increase chance of race condition
		time.Sleep(10 * time.Millisecond)
		transport := newMockTransport()
		client := &Client{
			name:         "test",
			transport:    transport,
			toolsMap:     make(map[string]*Tool),
			resourcesMap: make(map[string]*Resource),
			promptsMap:   make(map[string]*Prompt),
		}
		return client, nil
	}

	ctx := context.Background()

	// Launch many goroutines trying to acquire connections concurrently
	var wg sync.WaitGroup
	numGoroutines := 20 // More than MaxConnections to test the limit
	acquired := make([]*PooledConnection, 0)
	var acquiredMu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := pool.Acquire(ctx, factory)
			if err == nil && conn != nil {
				acquiredMu.Lock()
				acquired = append(acquired, conn)
				acquiredMu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Verify we never exceeded MaxConnections
	stats := pool.Stats()
	assert.LessOrEqual(t, stats.Total, config.MaxConnections, "Pool should not exceed MaxConnections")
	assert.LessOrEqual(t, stats.Total, 5, "Pool should not exceed MaxConnections")

	// Release all acquired connections
	acquiredMu.Lock()
	for _, conn := range acquired {
		pool.Release(conn)
	}
	acquiredMu.Unlock()

	// Verify final state
	stats = pool.Stats()
	assert.LessOrEqual(t, stats.Total, config.MaxConnections)
}
