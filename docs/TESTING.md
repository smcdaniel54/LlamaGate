# Testing LlamaGate

This guide explains how to test LlamaGate to ensure everything is working correctly.

## Prerequisites

1. **Ollama must be running**

   ```cmd
   ollama serve
   ```

   Or ensure Ollama is running as a service.

2. **At least one model available**

   ```cmd
   ollama pull llama2
   ```

   Or any other model you want to test with.

3. **LlamaGate built and ready**

   ```cmd
   scripts\windows\build.cmd
   ```

   Or use `go run ./cmd/llamagate`

## Quick Test

### Windows

1. **Start LlamaGate:**

   ```cmd
   scripts\windows\run.cmd
   ```

   Or if you have a `.env` file configured:

   ```cmd
   llamagate.exe
   ```

2. **Run the test script:**

   ```cmd
   scripts\windows\test.cmd
   ```

## Plugin System Testing

The plugin system includes comprehensive tests for plugin registration, validation, and execution.

### Quick Test

**Windows:**
```cmd
scripts\windows\test-plugins.cmd
```

**Unix/Linux/macOS:**
```bash
chmod +x scripts/unix/test-plugins.sh
./scripts/unix/test-plugins.sh
```

### What Gets Tested

1. **Plugin Discovery**
   - `GET /v1/plugins` - List all registered plugins
   - `GET /v1/plugins/:name` - Get plugin metadata

2. **Input Validation**
   - Required inputs validated
   - Type validation
   - Error messages clear

3. **Plugin Execution**
   - `POST /v1/plugins/:name/execute` - Execute plugin
   - Valid inputs return 200 OK
   - Invalid inputs return 400 Bad Request

### Test Plugins

Test plugins are defined in `tests/plugins/test_plugins.go`. To enable them:

1. Set `ENABLE_TEST_PLUGINS=true` in `.env`
2. Start LlamaGate
3. Run test scripts

### Expected Results

- ✅ Plugin discovery returns list of plugins
- ✅ Valid execution returns HTTP 200 with results
- ✅ Invalid inputs return HTTP 400 with error details
- ✅ Execution metadata included in responses

## Manual Testing

### 1. Health Check

Test that the server is running:

```cmd
curl http://localhost:11435/health
```

**Expected response:**

```json
{"status":"healthy"}
```

### 2. List Models

Test the models endpoint (if API key is set, include it):

```cmd
curl http://localhost:11435/v1/models -H "X-API-Key: sk-llamagate"
```

**Expected response:**

```json
{
  "object": "list",
  "data": [
    {
      "id": "llama2",
      "object": "model",
      "created": 0,
      "owned_by": "ollama"
    }
  ]
}
```

### 3. Chat Completions (Non-Streaming)

Test a simple chat completion:

```cmd
curl -X POST http://localhost:11435/v1/chat/completions ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: sk-llamagate" ^
  -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello!\"}]}"
```

**Expected response:**

```json
{
  "model": "llama2",
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "..."
      }
    }
  ]
}
```

### 4. Test Caching

Make the same request twice. The second should be much faster (cached):

```cmd
REM First request (slow)
curl -w "\nTime: %{time_total}s\n" -X POST http://localhost:11435/v1/chat/completions ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: sk-llamagate" ^
  -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"What is 2+2?\"}]}"

REM Second request (fast - cached)
curl -w "\nTime: %{time_total}s\n" -X POST http://localhost:11435/v1/chat/completions ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: sk-llamagate" ^
  -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"What is 2+2?\"}]}"
```

The second request should complete in milliseconds (cached) vs seconds (from Ollama).

### 5. Test Authentication

If `API_KEY` is set in your `.env`:

**Test with invalid key (should fail):**

```cmd
curl -w "\nHTTP Status: %{http_code}\n" http://localhost:11435/v1/models -H "X-API-Key: wrong-key"
```

**Expected:** `401 Unauthorized`

**Test with valid key (should succeed):**

```cmd
curl -w "\nHTTP Status: %{http_code}\n" http://localhost:11435/v1/models -H "X-API-Key: sk-llamagate"
```

**Expected:** `200 OK`

### 6. Test MCP API Endpoints (Optional)

If MCP is enabled in your configuration, you can test the MCP API endpoints:

**List MCP Servers:**

```cmd
curl -H "X-API-Key: sk-llamagate" http://localhost:11435/v1/mcp/servers
```

**Get Server Health:**

```cmd
curl -H "X-API-Key: sk-llamagate" http://localhost:11435/v1/mcp/servers/filesystem/health
```

**List Server Tools:**

```cmd
curl -H "X-API-Key: sk-llamagate" http://localhost:11435/v1/mcp/servers/filesystem/tools
```

**Execute a Tool:**

```cmd
curl -X POST -H "X-API-Key: sk-llamagate" -H "Content-Type: application/json" ^
  -d "{\"server\":\"filesystem\",\"tool\":\"read_file\",\"arguments\":{\"path\":\"/tmp/test.txt\"}}" ^
  http://localhost:11435/v1/mcp/execute
```

**Note:** If MCP is not enabled, these endpoints will return `503 Service Unavailable`. See [MCP Documentation](MCP.md) for configuration details.

### 7. Test MCP URI Scheme (Optional)

If MCP is enabled and you have servers with resources configured, you can test the MCP URI scheme:

**Test with MCP URI in message:**

```cmd
curl -X POST http://localhost:11435/v1/chat/completions ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: sk-llamagate" ^
  -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"Summarize mcp://filesystem/file:///docs/readme.txt\"}]}"
```

**Expected behavior:**
- LlamaGate will parse the `mcp://filesystem/file:///docs/readme.txt` URI
- Fetch the resource content from the filesystem MCP server
- Inject the resource content as context
- Send the enhanced conversation to Ollama

**Note:** If the MCP server is not available or the resource doesn't exist, LlamaGate will log a warning and continue processing the request without the resource context.

### 8. Test File Logging

1. Set `LOG_FILE=llamagate.log` in your `.env` file
2. Start LlamaGate
3. Make a few requests
4. Check that `llamagate.log` file exists and contains JSON log entries

```cmd
type llamagate.log
```

You should see structured JSON logs with request information.

### 9. Test Rate Limiting

Make rapid requests to test rate limiting:

```cmd
for /L %i in (1,1,15) do @curl -s -X POST http://localhost:11435/v1/chat/completions -H "Content-Type: application/json" -H "X-API-Key: sk-llamagate" -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"Test\"}]}" >nul && echo Request %i
```

After 50 requests (default `RATE_LIMIT_RPS=50`), you should start getting `429 Too Many Requests` responses.

### 10. Test Streaming

Test streaming chat completions:

```cmd
curl -X POST http://localhost:11435/v1/chat/completions ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: sk-llamagate" ^
  -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"Count to 5\"}],\"stream\":true}"
```

**Expected:** Stream of data chunks (Server-Sent Events format)

## Unit Test Coverage

LlamaGate includes comprehensive unit tests for all major components:

### Running Unit Tests

**Windows:**
```cmd
go test ./...
```

**Unix/Linux/macOS:**
```bash
go test ./...
```

### Test Coverage

Current test coverage:
- `internal/config`: 81.8% - Configuration loading and validation
- `internal/middleware`: 83.7% - Authentication and rate limiting
- `internal/logger`: 94.7% - Logger initialization and file handling
- `internal/proxy`: 34.5% - Chat completions, models, streaming, caching
- `internal/mcpclient`: 62.5% - MCP client, transports, pooling, health
- `internal/api`: 41.7% - MCP API endpoints
- `internal/tools`: 38.0% - Tool management and guardrails
- `internal/cache`: 31.8% - Caching implementation

### Running Tests with Coverage

**Windows:**
```cmd
go test ./... -cover
```

**Unix/Linux/macOS:**
```bash
go test ./... -cover
```

### Running Specific Test Packages

```bash
# Test only proxy package
go test ./internal/proxy -v

# Test only config package
go test ./internal/config -v

# Test with coverage report
go test ./internal/proxy -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Testing with Python

You can also test using the OpenAI Python SDK:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:11435/v1",
    api_key="sk-llamagate"  # Your API_KEY from .env
)

# Test models
models = client.models.list()
print(models)

# Test chat completion
response = client.chat.completions.create(
    model="llama2",
    messages=[
        {"role": "user", "content": "Hello!"}
    ]
)
print(response.choices[0].message.content)
```

## Testing Debug Mode

1. Set `DEBUG=true` in your `.env` file
2. Start LlamaGate
3. Make requests and check logs for additional debug information

You should see:

- Cache miss/hit messages
- More detailed request/response logging
- Gin framework debug output

## Common Issues

### "Connection refused" or "Failed to connect"

- **Issue:** Ollama is not running
- **Solution:** Start Ollama: `ollama serve`

### "Model not found"

- **Issue:** Model doesn't exist in Ollama
- **Solution:** Pull the model: `ollama pull llama2`

### "401 Unauthorized"

- **Issue:** API key mismatch or missing
- **Solution:** Check your `.env` file `API_KEY` matches the header value

### "429 Too Many Requests"

- **Issue:** Rate limit exceeded
- **Solution:** Wait a moment or increase `RATE_LIMIT_RPS` in `.env`

### Log file not created

- **Issue:** `LOG_FILE` path might be invalid or permissions issue
- **Solution:** Check file path is valid and you have write permissions

## Automated Testing

### Unit Tests

Run all unit tests:

**Windows:**
```cmd
go test ./...
```

**Unix/Linux/macOS:**
```bash
go test ./...
```

### Test Coverage

Current test coverage includes:

- **Configuration** (81.8%): Loading, validation, environment variables, MCP config
- **Middleware** (83.7%): Authentication (X-API-Key, Bearer token), rate limiting, health endpoint bypass
- **Logger** (94.7%): Initialization, file handling, debug/info modes
- **Proxy** (34.5%): Chat completions, models endpoint, streaming, caching, error handling
- **MCP Client** (62.5%): Client initialization, transports, connection pooling, health monitoring
- **API Handlers** (41.7%): MCP API endpoints, error handling
- **Tools** (38.0%): Tool management, guardrails, mapper
- **Cache** (31.8%): Cache operations, TTL, size limits

### Running Tests with Coverage

**Windows:**
```cmd
go test ./... -cover
```

**Unix/Linux/macOS:**
```bash
go test ./... -cover
```

### Running Specific Test Packages

```bash
# Test only proxy package
go test ./internal/proxy -v

# Test only config package
go test ./internal/config -v

# Test only middleware
go test ./internal/middleware -v

# Test with coverage report
go test ./internal/proxy -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### What's Tested

The unit tests cover:

- **Cache functionality**: Get, Set, TTL expiration, size limits
- **Proxy handlers**: Chat completions, models endpoint, streaming, error handling
- **Configuration loading**: Environment variables, YAML/JSON configs, validation
- **Authentication middleware**: API key validation, Bearer token support, health endpoint bypass
- **Rate limiting**: Leaky bucket algorithm, health endpoint bypass
- **Logger**: Initialization, file handling, debug/info modes
- **MCP client**: Initialization, tool/resource/prompt discovery, HTTP transport
- **MCP API**: All endpoints, error handling, authentication

## Performance Testing

For load testing, you can use tools like:

- **Apache Bench (ab)**: `ab -n 100 -c 10 http://localhost:11435/health`
- **wrk**: `wrk -t4 -c100 -d30s http://localhost:11435/health`
- **k6**: Write a k6 script for more complex scenarios
