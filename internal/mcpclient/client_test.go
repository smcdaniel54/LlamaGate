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

// mockTransport is a test transport implementation
type mockTransport struct {
	requests     []requestCall
	closed       bool
	responseFunc func(method string, params interface{}) (*JSONRPCResponse, error)
}

type requestCall struct {
	method string
	params interface{}
}

func newMockTransport() *mockTransport {
	return &mockTransport{
		requests: make([]requestCall, 0),
		closed:   false,
	}
}

func (m *mockTransport) SendRequest(_ context.Context, method string, params interface{}) (*JSONRPCResponse, error) {
	if m.closed {
		return nil, ErrConnectionClosed
	}

	m.requests = append(m.requests, requestCall{method: method, params: params})

	if m.responseFunc != nil {
		return m.responseFunc(method, params)
	}

	// Default responses
	var resp JSONRPCResponse
	resp.JSONRPC = JSONRPCVersion
	resp.ID = len(m.requests)

	switch method {
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

	case "notifications/initialized":
		resp.Result = json.RawMessage(`{}`)

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
					Arguments: []PromptArgument{
						{
							Name:     "text",
							Required: true,
						},
					},
				},
			},
		}
		resultJSON, _ := json.Marshal(result)
		resp.Result = resultJSON

	default:
		resp.Result = json.RawMessage(`{}`)
	}

	return &resp, nil
}

func (m *mockTransport) Close() error {
	m.closed = true
	return nil
}

func (m *mockTransport) IsClosed() bool {
	return m.closed
}

func TestClient_Resources(t *testing.T) {
	transport := newMockTransport()
	client := &Client{
		name:         "test",
		transport:    transport,
		toolsMap:     make(map[string]*Tool),
		resourcesMap: make(map[string]*Resource),
		promptsMap:   make(map[string]*Prompt),
	}

	ctx := context.Background()

	// Initialize
	err := client.initialize(ctx)
	require.NoError(t, err)

	// Discover resources
	err = client.discoverResources(ctx)
	require.NoError(t, err)

	// Test GetResources
	resources := client.GetResources()
	assert.Len(t, resources, 1)
	assert.Equal(t, "file:///test.txt", resources[0].URI)

	// Test GetResource
	resource, err := client.GetResource("file:///test.txt")
	require.NoError(t, err)
	assert.Equal(t, "test.txt", resource.Name)
	assert.Equal(t, "text/plain", resource.MimeType)

	// Test GetResource not found
	_, err = client.GetResource("file:///notfound.txt")
	assert.Error(t, err)

	// Test ReadResource
	transport.responseFunc = func(method string, _ interface{}) (*JSONRPCResponse, error) {
		if method == "resources/read" {
			result := ResourceReadResult{
				Contents: []ResourceContent{
					{
						URI:      "file:///test.txt",
						MimeType: "text/plain",
						Text:     "Hello, World!",
					},
				},
			}
			resultJSON, _ := json.Marshal(result)
			return &JSONRPCResponse{
				JSONRPC: JSONRPCVersion,
				ID:      len(transport.requests) + 1,
				Result:  resultJSON,
			}, nil
		}
		// Default response for other methods
		return &JSONRPCResponse{
			JSONRPC: JSONRPCVersion,
			ID:      len(transport.requests) + 1,
			Result:  json.RawMessage(`{}`),
		}, nil
	}

	result, err := client.ReadResource(ctx, "file:///test.txt")
	require.NoError(t, err)
	assert.Len(t, result.Contents, 1)
	assert.Equal(t, "Hello, World!", result.Contents[0].Text)

	// Test RefreshResources
	err = client.RefreshResources(ctx)
	require.NoError(t, err)
}

func TestClient_Prompts(t *testing.T) {
	transport := newMockTransport()
	client := &Client{
		name:         "test",
		transport:    transport,
		toolsMap:     make(map[string]*Tool),
		resourcesMap: make(map[string]*Resource),
		promptsMap:   make(map[string]*Prompt),
	}

	ctx := context.Background()

	// Initialize
	err := client.initialize(ctx)
	require.NoError(t, err)

	// Discover prompts
	err = client.discoverPrompts(ctx)
	require.NoError(t, err)

	// Test GetPrompts
	prompts := client.GetPrompts()
	assert.Len(t, prompts, 1)
	assert.Equal(t, "test_prompt", prompts[0].Name)

	// Test GetPrompt
	prompt, err := client.GetPrompt("test_prompt")
	require.NoError(t, err)
	assert.Equal(t, "A test prompt", prompt.Description)
	assert.Len(t, prompt.Arguments, 1)

	// Test GetPrompt not found
	_, err = client.GetPrompt("notfound")
	assert.Error(t, err)

	// Test GetPromptTemplate
	transport.responseFunc = func(method string, _ interface{}) (*JSONRPCResponse, error) {
		if method == "prompts/get" {
			result := PromptGetResult{
				Messages: []PromptMessage{
					{
						Role:    "user",
						Content: "Summarize this text: test text",
					},
				},
			}
			resultJSON, _ := json.Marshal(result)
			return &JSONRPCResponse{
				JSONRPC: JSONRPCVersion,
				ID:      len(transport.requests) + 1,
				Result:  resultJSON,
			}, nil
		}
		// Default response for other methods
		return &JSONRPCResponse{
			JSONRPC: JSONRPCVersion,
			ID:      len(transport.requests) + 1,
			Result:  json.RawMessage(`{}`),
		}, nil
	}

	result, err := client.GetPromptTemplate(ctx, "test_prompt", map[string]interface{}{
		"text": "test text",
	})
	require.NoError(t, err)
	assert.Len(t, result.Messages, 1)
	assert.Equal(t, "user", result.Messages[0].Role)

	// Test RefreshPrompts
	err = client.RefreshPrompts(ctx)
	require.NoError(t, err)
}

func TestClient_ReadResource_NotInitialized(t *testing.T) {
	transport := newMockTransport()
	client := &Client{
		name:         "test",
		transport:    transport,
		toolsMap:     make(map[string]*Tool),
		resourcesMap: make(map[string]*Resource),
		promptsMap:   make(map[string]*Prompt),
	}

	ctx := context.Background()
	_, err := client.ReadResource(ctx, "file:///test.txt")
	assert.Error(t, err)
	assert.Equal(t, ErrNotInitialized, err)
}

func TestClient_GetPromptTemplate_NotInitialized(t *testing.T) {
	transport := newMockTransport()
	client := &Client{
		name:         "test",
		transport:    transport,
		toolsMap:     make(map[string]*Tool),
		resourcesMap: make(map[string]*Resource),
		promptsMap:   make(map[string]*Prompt),
	}

	ctx := context.Background()
	_, err := client.GetPromptTemplate(ctx, "test_prompt", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrNotInitialized, err)
}

func TestNewClientWithHTTP(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req JSONRPCRequest
		json.NewDecoder(r.Body).Decode(&req)

		var resp JSONRPCResponse
		resp.JSONRPC = JSONRPCVersion
		resp.ID = req.ID

		switch req.Method {
		case "initialize":
			result := InitializeResult{
				ProtocolVersion: "2024-11-05",
				Capabilities:    ServerCapabilities{},
				ServerInfo: ServerInfo{
					Name:    "http-server",
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
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	headers := map[string]string{
		"Authorization": "Bearer token",
	}

	client, err := NewClientWithHTTP("http-test", server.URL, headers, 30*time.Second)
	require.NoError(t, err)
	defer client.Close()

	assert.True(t, client.IsInitialized())
	assert.Equal(t, "http-test", client.GetName())
	assert.NotNil(t, client.GetServerInfo())
	assert.Equal(t, "http-server", client.GetServerInfo().Name)
}

func TestClient_Resources_EmptyList(t *testing.T) {
	transport := newMockTransport()
	transport.responseFunc = func(method string, _ interface{}) (*JSONRPCResponse, error) {
		if method == "resources/list" {
			result := ResourcesListResult{Resources: []Resource{}}
			resultJSON, _ := json.Marshal(result)
			return &JSONRPCResponse{
				JSONRPC: JSONRPCVersion,
				ID:      len(transport.requests) + 1,
				Result:  resultJSON,
			}, nil
		}
		// Default response for other methods
		return &JSONRPCResponse{
			JSONRPC: JSONRPCVersion,
			ID:      len(transport.requests) + 1,
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

	ctx := context.Background()
	err := client.initialize(ctx)
	require.NoError(t, err)

	err = client.discoverResources(ctx)
	require.NoError(t, err)

	resources := client.GetResources()
	assert.Len(t, resources, 0)
}

func TestClient_Prompts_EmptyList(t *testing.T) {
	transport := newMockTransport()
	transport.responseFunc = func(method string, _ interface{}) (*JSONRPCResponse, error) {
		if method == "prompts/list" {
			result := PromptsListResult{Prompts: []Prompt{}}
			resultJSON, _ := json.Marshal(result)
			return &JSONRPCResponse{
				JSONRPC: JSONRPCVersion,
				ID:      len(transport.requests) + 1,
				Result:  resultJSON,
			}, nil
		}
		// Default response for other methods
		return &JSONRPCResponse{
			JSONRPC: JSONRPCVersion,
			ID:      len(transport.requests) + 1,
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

	ctx := context.Background()
	err := client.initialize(ctx)
	require.NoError(t, err)

	err = client.discoverPrompts(ctx)
	require.NoError(t, err)

	prompts := client.GetPrompts()
	assert.Len(t, prompts, 0)
}
