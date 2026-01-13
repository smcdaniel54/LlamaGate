package mcpclient

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHealthMonitor(t *testing.T) {
	monitor := NewHealthMonitor(60*time.Second, 5*time.Second)
	assert.NotNil(t, monitor)

	monitor.Stop()
}

func TestHealthMonitor_RegisterUnregister(t *testing.T) {
	monitor := NewHealthMonitor(60*time.Second, 5*time.Second)
	defer monitor.Stop()

	transport := newMockTransport()
	client := &Client{
		name:         "test",
		transport:    transport,
		toolsMap:     make(map[string]*Tool),
		resourcesMap: make(map[string]*Resource),
		promptsMap:   make(map[string]*Prompt),
	}

	monitor.RegisterClient("test", client)

	health, exists := monitor.GetHealth("test")
	assert.True(t, exists)
	assert.Equal(t, HealthStatusUnknown, health.Status)

	monitor.UnregisterClient("test")

	_, exists = monitor.GetHealth("test")
	assert.False(t, exists)
}

func TestHealthMonitor_CheckHealth(t *testing.T) {
	monitor := NewHealthMonitor(60*time.Second, 5*time.Second)
	defer monitor.Stop()

	transport := newMockTransport()
	// Set up transport to return successful response
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
		name:         "test",
		transport:    transport,
		toolsMap:     make(map[string]*Tool),
		resourcesMap: make(map[string]*Resource),
		promptsMap:   make(map[string]*Prompt),
	}

	monitor.RegisterClient("test", client)

	ctx := context.Background()
	result, err := monitor.CheckHealth(ctx, "test")
	require.NoError(t, err)
	assert.Equal(t, HealthStatusHealthy, result.Status)
	assert.GreaterOrEqual(t, result.Latency, time.Duration(0))
	assert.NotZero(t, result.LastCheck)
}

func TestHealthMonitor_GetAllHealth(t *testing.T) {
	monitor := NewHealthMonitor(60*time.Second, 5*time.Second)
	defer monitor.Stop()

	transport1 := newMockTransport()
	client1 := &Client{
		name:         "test1",
		transport:    transport1,
		toolsMap:     make(map[string]*Tool),
		resourcesMap: make(map[string]*Resource),
		promptsMap:   make(map[string]*Prompt),
	}

	transport2 := newMockTransport()
	client2 := &Client{
		name:         "test2",
		transport:    transport2,
		toolsMap:     make(map[string]*Tool),
		resourcesMap: make(map[string]*Resource),
		promptsMap:   make(map[string]*Prompt),
	}

	monitor.RegisterClient("test1", client1)
	monitor.RegisterClient("test2", client2)

	allHealth := monitor.GetAllHealth()
	assert.Len(t, allHealth, 2)
	assert.Contains(t, allHealth, "test1")
	assert.Contains(t, allHealth, "test2")
}
