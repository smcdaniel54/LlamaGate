package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/llamagate/llamagate/internal/mcpclient"
	"github.com/llamagate/llamagate/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter(handler *MCPHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/v1")
	mcp := v1.Group("/mcp")
	{
		mcp.GET("/servers", handler.ListServers)
		mcp.GET("/servers/health", handler.GetAllHealth)
		mcp.GET("/servers/:name", handler.GetServer)
		mcp.GET("/servers/:name/health", handler.GetServerHealth)
		mcp.GET("/servers/:name/stats", handler.GetServerStats)
		mcp.GET("/servers/:name/tools", handler.ListServerTools)
		mcp.GET("/servers/:name/resources", handler.ListServerResources)
		mcp.GET("/servers/:name/resources/*uri", handler.ReadServerResource)
		mcp.GET("/servers/:name/prompts", handler.ListServerPrompts)
		mcp.POST("/servers/:name/prompts/:promptName", handler.GetServerPrompt)
		mcp.POST("/execute", handler.ExecuteTool)
		mcp.POST("/servers/:name/refresh", handler.RefreshServerMetadata)
	}
	return router
}

func createMockMCPServer() *httptest.Server {
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
		case "tools/list":
			resp.Result = json.RawMessage(`{"tools":[]}`)
		case "resources/list":
			resp.Result = json.RawMessage(`{"resources":[]}`)
		case "prompts/list":
			resp.Result = json.RawMessage(`{"prompts":[]}`)
		case "tools/call":
			// Return a proper tool execution result (MCP uses "tools/call" not "tools/execute")
			result := map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "Tool execution result",
					},
				},
				"isError": false,
			}
			resultJSON, _ := json.Marshal(result)
			resp.Result = resultJSON
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestMCPHandler_ListServers_NotEnabled(t *testing.T) {
	handler := NewMCPHandler(nil, nil)
	router := setupTestRouter(handler)

	req := httptest.NewRequest("GET", "/v1/mcp/servers", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestMCPHandler_ListServers(t *testing.T) {
	managerConfig := mcpclient.ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}
	serverManager := mcpclient.NewServerManager(managerConfig)
	defer serverManager.Close()

	toolManager := tools.NewManager()

	// Create mock MCP server
	mcpServer := createMockMCPServer()
	defer mcpServer.Close()

	client, err := mcpclient.NewClientWithHTTP("test-server", mcpServer.URL, nil, 30*time.Second)
	require.NoError(t, err)
	defer client.Close()

	toolManager.AddClient(client)
	serverManager.AddServer("test-server", client, "http")

	handler := NewMCPHandler(toolManager, serverManager)
	router := setupTestRouter(handler)

	req := httptest.NewRequest("GET", "/v1/mcp/servers", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "servers")
	assert.Contains(t, response, "count")
}

func TestMCPHandler_GetServer_NotFound(t *testing.T) {
	managerConfig := mcpclient.ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}
	serverManager := mcpclient.NewServerManager(managerConfig)
	defer serverManager.Close()

	toolManager := tools.NewManager()
	handler := NewMCPHandler(toolManager, serverManager)
	router := setupTestRouter(handler)

	req := httptest.NewRequest("GET", "/v1/mcp/servers/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMCPHandler_GetServerHealth(t *testing.T) {
	managerConfig := mcpclient.ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}
	serverManager := mcpclient.NewServerManager(managerConfig)
	defer serverManager.Close()

	toolManager := tools.NewManager()

	mcpServer := createMockMCPServer()
	defer mcpServer.Close()

	client, err := mcpclient.NewClientWithHTTP("test-server", mcpServer.URL, nil, 30*time.Second)
	require.NoError(t, err)
	defer client.Close()

	toolManager.AddClient(client)
	serverManager.AddServer("test-server", client, "http")

	handler := NewMCPHandler(toolManager, serverManager)
	router := setupTestRouter(handler)

	req := httptest.NewRequest("GET", "/v1/mcp/servers/test-server/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "server")
	assert.Contains(t, response, "status")
}

func TestMCPHandler_ListServerTools(t *testing.T) {
	managerConfig := mcpclient.ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}
	serverManager := mcpclient.NewServerManager(managerConfig)
	defer serverManager.Close()

	toolManager := tools.NewManager()

	mcpServer := createMockMCPServer()
	defer mcpServer.Close()

	client, err := mcpclient.NewClientWithHTTP("test-server", mcpServer.URL, nil, 30*time.Second)
	require.NoError(t, err)
	defer client.Close()

	// Add client to tool manager (may fail silently if no tools, but that's OK for testing)
	err = toolManager.AddClient(client)
	// Don't fail if AddClient returns error - it's OK if server has no tools
	_ = err

	serverManager.AddServer("test-server", client, "http")

	handler := NewMCPHandler(toolManager, serverManager)
	router := setupTestRouter(handler)

	req := httptest.NewRequest("GET", "/v1/mcp/servers/test-server/tools", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "server")
	assert.Contains(t, response, "tools")
	assert.Contains(t, response, "count")
}

func TestMCPHandler_ListServerResources(t *testing.T) {
	managerConfig := mcpclient.ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}
	serverManager := mcpclient.NewServerManager(managerConfig)
	defer serverManager.Close()

	toolManager := tools.NewManager()

	mcpServer := createMockMCPServer()
	defer mcpServer.Close()

	client, err := mcpclient.NewClientWithHTTP("test-server", mcpServer.URL, nil, 30*time.Second)
	require.NoError(t, err)
	defer client.Close()

	toolManager.AddClient(client)
	serverManager.AddServer("test-server", client, "http")

	handler := NewMCPHandler(toolManager, serverManager)
	router := setupTestRouter(handler)

	req := httptest.NewRequest("GET", "/v1/mcp/servers/test-server/resources", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "server")
	assert.Contains(t, response, "resources")
	assert.Contains(t, response, "count")
}

func TestMCPHandler_ListServerPrompts(t *testing.T) {
	managerConfig := mcpclient.ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}
	serverManager := mcpclient.NewServerManager(managerConfig)
	defer serverManager.Close()

	toolManager := tools.NewManager()

	mcpServer := createMockMCPServer()
	defer mcpServer.Close()

	client, err := mcpclient.NewClientWithHTTP("test-server", mcpServer.URL, nil, 30*time.Second)
	require.NoError(t, err)
	defer client.Close()

	toolManager.AddClient(client)
	serverManager.AddServer("test-server", client, "http")

	handler := NewMCPHandler(toolManager, serverManager)
	router := setupTestRouter(handler)

	req := httptest.NewRequest("GET", "/v1/mcp/servers/test-server/prompts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "server")
	assert.Contains(t, response, "prompts")
	assert.Contains(t, response, "count")
}

func TestMCPHandler_ExecuteTool(t *testing.T) {
	managerConfig := mcpclient.ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}
	serverManager := mcpclient.NewServerManager(managerConfig)
	defer serverManager.Close()

	toolManager := tools.NewManager()

	mcpServer := createMockMCPServer()
	defer mcpServer.Close()

	client, err := mcpclient.NewClientWithHTTP("test-server", mcpServer.URL, nil, 30*time.Second)
	require.NoError(t, err)
	defer client.Close()

	toolManager.AddClient(client)
	serverManager.AddServer("test-server", client, "http")

	handler := NewMCPHandler(toolManager, serverManager)
	router := setupTestRouter(handler)

	// Create execute request
	executeReq := map[string]interface{}{
		"server":    "test-server",
		"tool":      "test_tool",
		"arguments": map[string]interface{}{"arg1": "value1"},
	}
	reqBody, _ := json.Marshal(executeReq)

	req := httptest.NewRequest("POST", "/v1/mcp/execute", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should succeed with proper mock response
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "server")
	assert.Contains(t, response, "tool")
	assert.Contains(t, response, "result")
	assert.Contains(t, response, "duration")
	assert.Equal(t, false, response["is_error"])
	assert.Equal(t, "test-server", response["server"])
	assert.Equal(t, "test_tool", response["tool"])
}

func TestMCPHandler_RefreshServerMetadata(t *testing.T) {
	managerConfig := mcpclient.ManagerConfig{
		PoolSize:       5,
		PoolIdleTime:   1 * time.Minute,
		HealthInterval: 30 * time.Second,
		HealthTimeout:  3 * time.Second,
		CacheTTL:       2 * time.Minute,
	}
	serverManager := mcpclient.NewServerManager(managerConfig)
	defer serverManager.Close()

	toolManager := tools.NewManager()

	mcpServer := createMockMCPServer()
	defer mcpServer.Close()

	client, err := mcpclient.NewClientWithHTTP("test-server", mcpServer.URL, nil, 30*time.Second)
	require.NoError(t, err)
	defer client.Close()

	toolManager.AddClient(client)
	serverManager.AddServer("test-server", client, "http")

	handler := NewMCPHandler(toolManager, serverManager)
	router := setupTestRouter(handler)

	req := httptest.NewRequest("POST", "/v1/mcp/servers/test-server/refresh", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "server")
	assert.Contains(t, response, "status")
}
