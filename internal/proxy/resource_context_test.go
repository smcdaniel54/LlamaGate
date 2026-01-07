package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/cache"
	"github.com/llamagate/llamagate/internal/mcpclient"
	"github.com/llamagate/llamagate/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createMockMCPServerForResources creates a mock MCP server that supports resources
func createMockMCPServerForResources() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req mcpclient.JSONRPCRequest
		json.NewDecoder(r.Body).Decode(&req)

		var resp mcpclient.JSONRPCResponse
		resp.JSONRPC = mcpclient.JSONRPCVersion
		resp.ID = req.ID

		switch req.Method {
		case "initialize":
			result := mcpclient.InitializeResult{
				ProtocolVersion: "2024-11-05",
				Capabilities:    mcpclient.ServerCapabilities{},
				ServerInfo: mcpclient.ServerInfo{
					Name:    "test-server",
					Version: "1.0.0",
				},
			}
			resultJSON, _ := json.Marshal(result)
			resp.Result = resultJSON

		case "notifications/initialized":
			resp.Result = json.RawMessage(`{}`)

		case "resources/list":
			result := mcpclient.ResourcesListResult{
				Resources: []mcpclient.Resource{
					{URI: "file:///test.txt", Name: "test.txt", MimeType: "text/plain"},
				},
			}
			resultJSON, _ := json.Marshal(result)
			resp.Result = resultJSON

		case "resources/read":
			var params mcpclient.ResourceReadParams
			json.Unmarshal(req.Params, &params)
			if params.URI == "file:///test.txt" {
				result := mcpclient.ResourceReadResult{
					Contents: []mcpclient.ResourceContent{
						{URI: "file:///test.txt", MimeType: "text/plain", Text: "This is test file content"},
					},
				}
				resultJSON, _ := json.Marshal(result)
				resp.Result = resultJSON
			} else {
				resp.Error = &mcpclient.JSONRPCError{Code: -32000, Message: "Resource not found"}
			}

		default:
			resp.Result = json.RawMessage(`{}`)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestProxy_injectMCPResourceContext(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	cacheInstance := cache.New()
	proxy := New("http://localhost:11434", cacheInstance)

	// Create server manager and mock MCP server
	managerConfig := mcpclient.ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}
	serverManager := mcpclient.NewServerManager(managerConfig)
	defer serverManager.Close()

	mcpServer := createMockMCPServerForResources()
	defer mcpServer.Close()

	// Create MCP client
	client, err := mcpclient.NewClientWithHTTP("test-server", mcpServer.URL, nil, 30*time.Second)
	require.NoError(t, err)
	defer client.Close()

	// Add client to server manager
	err = serverManager.AddServer("test-server", client, "http")
	require.NoError(t, err)

	// Set server manager on proxy
	proxy.SetServerManager(serverManager)

	// Create tool manager (required for resource injection check)
	toolManager := tools.NewManager()
	toolManager.AddClient(client)
	proxy.SetToolManager(toolManager, nil)

	// Test messages with MCP URI
	messages := []Message{
		{
			Role:    "user",
			Content: "Please summarize mcp://test-server/file:///test.txt",
		},
	}

	// Inject resource context
	ctx := context.Background()
	enhancedMessages, err := proxy.injectMCPResourceContext(ctx, "test-request-id", messages)

	// Verify
	require.NoError(t, err)
	assert.Greater(t, len(enhancedMessages), len(messages), "Should have added resource context")

	// Check that first message is system message with resource content
	assert.Equal(t, "system", enhancedMessages[0].Role)
	assert.Contains(t, enhancedMessages[0].Content.(string), "This is test file content")
	assert.Contains(t, enhancedMessages[0].Content.(string), "mcp://test-server/file:///test.txt")
}

func TestProxy_injectMCPResourceContext_NoURIs(t *testing.T) {
	// Setup
	cacheInstance := cache.New()
	proxy := New("http://localhost:11434", cacheInstance)

	// Create server manager
	managerConfig := mcpclient.ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}
	serverManager := mcpclient.NewServerManager(managerConfig)
	defer serverManager.Close()

	proxy.SetServerManager(serverManager)

	toolManager := tools.NewManager()
	proxy.SetToolManager(toolManager, nil)

	// Test messages without MCP URI
	messages := []Message{
		{
			Role:    "user",
			Content: "This is a regular message without any MCP URIs",
		},
	}

	// Inject resource context
	ctx := context.Background()
	enhancedMessages, err := proxy.injectMCPResourceContext(ctx, "test-request-id", messages)

	// Verify - should return messages unchanged
	require.NoError(t, err)
	assert.Equal(t, len(messages), len(enhancedMessages), "Should not add context for messages without URIs")
	assert.Equal(t, messages[0].Content, enhancedMessages[0].Content)
}

func TestProxy_injectMCPResourceContext_MultipleURIs(t *testing.T) {
	// Setup
	cacheInstance := cache.New()
	proxy := New("http://localhost:11434", cacheInstance)

	// Create server manager and mock MCP server
	managerConfig := mcpclient.ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}
	serverManager := mcpclient.NewServerManager(managerConfig)
	defer serverManager.Close()

	mcpServer := createMockMCPServerForResources()
	defer mcpServer.Close()

	// Create MCP client
	client, err := mcpclient.NewClientWithHTTP("test-server", mcpServer.URL, nil, 30*time.Second)
	require.NoError(t, err)
	defer client.Close()

	// Add client to server manager
	err = serverManager.AddServer("test-server", client, "http")
	require.NoError(t, err)

	proxy.SetServerManager(serverManager)

	toolManager := tools.NewManager()
	toolManager.AddClient(client)
	proxy.SetToolManager(toolManager, nil)

	// Test messages with multiple MCP URIs
	messages := []Message{
		{
			Role:    "user",
			Content: "Compare mcp://test-server/file:///test.txt and mcp://test-server/file:///test.txt",
		},
	}

	// Inject resource context
	ctx := context.Background()
	enhancedMessages, err := proxy.injectMCPResourceContext(ctx, "test-request-id", messages)

	// Verify
	require.NoError(t, err)
	assert.Greater(t, len(enhancedMessages), len(messages))

	// Should have resource content (may appear once or twice depending on deduplication)
	systemContent := enhancedMessages[0].Content.(string)
	assert.Contains(t, systemContent, "This is test file content")
}

func TestProxy_injectMCPResourceContext_InvalidServer(t *testing.T) {
	// Setup
	cacheInstance := cache.New()
	proxy := New("http://localhost:11434", cacheInstance)

	// Create server manager (empty - no servers)
	managerConfig := mcpclient.ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}
	serverManager := mcpclient.NewServerManager(managerConfig)
	defer serverManager.Close()

	proxy.SetServerManager(serverManager)

	toolManager := tools.NewManager()
	proxy.SetToolManager(toolManager, nil)

	// Test messages with MCP URI pointing to non-existent server
	messages := []Message{
		{
			Role:    "user",
			Content: "Read mcp://nonexistent-server/file.txt",
		},
	}

	// Inject resource context
	ctx := context.Background()
	enhancedMessages, err := proxy.injectMCPResourceContext(ctx, "test-request-id", messages)

	// Verify - should handle error gracefully and return original messages
	require.NoError(t, err) // Function should not return error, just log warning
	assert.Equal(t, len(messages), len(enhancedMessages), "Should return original messages on error")
}
