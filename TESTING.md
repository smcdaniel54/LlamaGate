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

## Manual Testing

### 1. Health Check

Test that the server is running:

```cmd
curl http://localhost:8080/health
```

**Expected response:**
```json
{"status":"healthy"}
```

### 2. List Models

Test the models endpoint (if API key is set, include it):

```cmd
curl http://localhost:8080/v1/models -H "X-API-Key: sk-llamagate"
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
curl -X POST http://localhost:8080/v1/chat/completions ^
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
curl -w "\nTime: %{time_total}s\n" -X POST http://localhost:8080/v1/chat/completions ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: sk-llamagate" ^
  -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"What is 2+2?\"}]}"

REM Second request (fast - cached)
curl -w "\nTime: %{time_total}s\n" -X POST http://localhost:8080/v1/chat/completions ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: sk-llamagate" ^
  -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"What is 2+2?\"}]}"
```

The second request should complete in milliseconds (cached) vs seconds (from Ollama).

### 5. Test Authentication

If `API_KEY` is set in your `.env`:

**Test with invalid key (should fail):**
```cmd
curl -w "\nHTTP Status: %{http_code}\n" http://localhost:8080/v1/models -H "X-API-Key: wrong-key"
```

**Expected:** `401 Unauthorized`

**Test with valid key (should succeed):**
```cmd
curl -w "\nHTTP Status: %{http_code}\n" http://localhost:8080/v1/models -H "X-API-Key: sk-llamagate"
```

**Expected:** `200 OK`

### 6. Test File Logging

1. Set `LOG_FILE=llamagate.log` in your `.env` file
2. Start LlamaGate
3. Make a few requests
4. Check that `llamagate.log` file exists and contains JSON log entries

```cmd
type llamagate.log
```

You should see structured JSON logs with request information.

### 7. Test Rate Limiting

Make rapid requests to test rate limiting:

```cmd
for /L %i in (1,1,15) do @curl -s -X POST http://localhost:8080/v1/chat/completions -H "Content-Type: application/json" -H "X-API-Key: sk-llamagate" -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"Test\"}]}" >nul && echo Request %i
```

After 10 requests (default `RATE_LIMIT_RPS=10`), you should start getting `429 Too Many Requests` responses.

### 8. Test Streaming

Test streaming chat completions:

```cmd
curl -X POST http://localhost:8080/v1/chat/completions ^
  -H "Content-Type: application/json" ^
  -H "X-API-Key: sk-llamagate" ^
  -d "{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"Count to 5\"}],\"stream\":true}"
```

**Expected:** Stream of data chunks (Server-Sent Events format)

## Testing with Python

You can also test using the OpenAI Python SDK:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
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

Run the unit tests:

```cmd
go test ./...
```

This will test:
- Cache functionality
- Proxy handlers
- Configuration loading

## Performance Testing

For load testing, you can use tools like:
- **Apache Bench (ab)**: `ab -n 100 -c 10 http://localhost:8080/health`
- **wrk**: `wrk -t4 -c100 -d30s http://localhost:8080/health`
- **k6**: Write a k6 script for more complex scenarios

