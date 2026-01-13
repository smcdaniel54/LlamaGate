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

func TestNewHTTPTransport(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		headers map[string]string
		timeout time.Duration
		wantErr bool
	}{
		{
			name:    "valid URL",
			url:     "http://localhost:3000/mcp",
			headers: nil,
			timeout: 30 * time.Second,
			wantErr: false,
		},
		{
			name:    "valid URL with headers",
			url:     "http://localhost:3000/mcp",
			headers: map[string]string{"Authorization": "Bearer token"},
			timeout: 30 * time.Second,
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			headers: nil,
			timeout: 30 * time.Second,
			wantErr: true,
		},
		{
			name:    "zero timeout uses default",
			url:     "http://localhost:3000/mcp",
			headers: nil,
			timeout: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, err := NewHTTPTransport(tt.url, tt.headers, tt.timeout)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, transport)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, transport)
				assert.False(t, transport.IsClosed())
			}
		})
	}
}

func TestHTTPTransport_SendRequest(t *testing.T) {
	// Create a mock MCP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Content-Type
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Read request
		var req JSONRPCRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Verify JSON-RPC version
		assert.Equal(t, JSONRPCVersion, req.JSONRPC)

		// Create response based on method
		var resp JSONRPCResponse
		resp.JSONRPC = JSONRPCVersion
		resp.ID = req.ID

		switch req.Method {
		case "initialize":
			result := InitializeResult{
				ProtocolVersion: "2024-11-05",
				Capabilities:    ServerCapabilities{},
				ServerInfo: ServerInfo{
					Name:    "test-server",
					Version: "1.0.0",
				},
			}
			resultJSON, _ := json.Marshal(result)
			resp.Result = resultJSON

		case "tools/list":
			result := ToolsListResult{
				Tools: []Tool{
					{
						Name:        "test_tool",
						Description: "A test tool",
						InputSchema: map[string]interface{}{
							"type": "object",
						},
					},
				},
			}
			resultJSON, _ := json.Marshal(result)
			resp.Result = resultJSON

		case "resources/list":
			result := ResourcesListResult{
				Resources: []Resource{
					{
						URI:      "file:///test.txt",
						Name:     "test.txt",
						MimeType: "text/plain",
					},
				},
			}
			resultJSON, _ := json.Marshal(result)
			resp.Result = resultJSON

		case "prompts/list":
			result := PromptsListResult{
				Prompts: []Prompt{
					{
						Name:        "test_prompt",
						Description: "A test prompt",
					},
				},
			}
			resultJSON, _ := json.Marshal(result)
			resp.Result = resultJSON

		default:
			resp.Error = &JSONRPCError{
				Code:    ErrCodeMethodNotFound,
				Message: "Method not found",
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	transport, err := NewHTTPTransport(server.URL, nil, 30*time.Second)
	require.NoError(t, err)
	defer transport.Close()

	tests := []struct {
		name    string
		method  string
		params  interface{}
		wantErr bool
	}{
		{
			name:    "initialize",
			method:  "initialize",
			params:  InitializeParams{ProtocolVersion: "2024-11-05"},
			wantErr: false,
		},
		{
			name:    "tools/list",
			method:  "tools/list",
			params:  nil,
			wantErr: false,
		},
		{
			name:    "resources/list",
			method:  "resources/list",
			params:  nil,
			wantErr: false,
		},
		{
			name:    "prompts/list",
			method:  "prompts/list",
			params:  nil,
			wantErr: false,
		},
		{
			name:    "unknown method",
			method:  "unknown/method",
			params:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := transport.SendRequest(ctx, tt.method, tt.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, JSONRPCVersion, resp.JSONRPC)
			}
		})
	}
}

func TestHTTPTransport_Headers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify custom headers
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
		assert.Equal(t, "secret-key", r.Header.Get("X-API-Key"))

		resp := JSONRPCResponse{
			JSONRPC: JSONRPCVersion,
			ID:      1,
			Result:  json.RawMessage(`{}`),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	headers := map[string]string{
		"Authorization": "Bearer token123",
		"X-API-Key":     "secret-key",
	}

	transport, err := NewHTTPTransport(server.URL, headers, 30*time.Second)
	require.NoError(t, err)
	defer transport.Close()

	ctx := context.Background()
	resp, err := transport.SendRequest(ctx, "test", nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestHTTPTransport_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		serverHandler  http.HandlerFunc
		wantErr        bool
		checkErrorType bool
	}{
		{
			name: "HTTP 500 error",
			serverHandler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			},
			wantErr: true,
		},
		{
			name: "JSON-RPC error",
			serverHandler: func(w http.ResponseWriter, _ *http.Request) {
				resp := JSONRPCResponse{
					JSONRPC: JSONRPCVersion,
					ID:      1,
					Error: &JSONRPCError{
						Code:    ErrCodeInternalError,
						Message: "Internal error",
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			},
			wantErr: true,
		},
		{
			name: "invalid JSON response",
			serverHandler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("invalid json"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			transport, err := NewHTTPTransport(server.URL, nil, 30*time.Second)
			require.NoError(t, err)
			defer transport.Close()

			ctx := context.Background()
			resp, err := transport.SendRequest(ctx, "test", nil)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestHTTPTransport_Close(t *testing.T) {
	transport, err := NewHTTPTransport("http://localhost:3000", nil, 30*time.Second)
	require.NoError(t, err)

	assert.False(t, transport.IsClosed())
	err = transport.Close()
	assert.NoError(t, err)
	assert.True(t, transport.IsClosed())
}

func TestHTTPTransport_ContextTimeout(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		resp := JSONRPCResponse{
			JSONRPC: JSONRPCVersion,
			ID:      1,
			Result:  json.RawMessage(`{}`),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	transport, err := NewHTTPTransport(server.URL, nil, 30*time.Second)
	require.NoError(t, err)
	defer transport.Close()

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = transport.SendRequest(ctx, "test", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}
