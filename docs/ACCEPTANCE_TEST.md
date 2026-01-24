# LlamaGate Manual Acceptance Test

**Version:** 1.0  
**Date:** _______________  
**Tester Name:** _______________  
**Test Environment:** _______________  
**LlamaGate Version:** _______________  

---

## Document Purpose

This document provides a comprehensive manual acceptance test checklist for LlamaGate. Use this document to verify that all features are working correctly before deploying to production or accepting delivery from a development team.

**Instructions:**
- Complete each test item in order
- Check the box (☐) when a test passes
- Mark with ✗ if a test fails
- Add notes in the "Notes" column for any observations
- Document any issues or deviations in the "Issues" section at the end

**Generating PDF:**
- **Unix/Linux/macOS:** Run `./scripts/unix/generate-acceptance-test-pdf.sh`
- **Windows:** Run `.\scripts\windows\generate-acceptance-test-pdf.ps1`
- **Alternative:** Use an online Markdown to PDF converter or VS Code extension "Markdown PDF"

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation Verification](#installation-verification)
3. [Basic Functionality](#basic-functionality)
4. [Authentication & Security](#authentication--security)
5. [Streaming](#streaming)
6. [Caching](#caching)
7. [Rate Limiting](#rate-limiting)
8. [MCP Integration](#mcp-integration)
9. [Extension System](#extension-system)
10. [Error Handling](#error-handling)
11. [Logging & Monitoring](#logging--monitoring)
12. [Performance](#performance)
13. [OpenAI SDK Compatibility](#openai-sdk-compatibility)
14. [Issues & Observations](#issues--observations)

---

## Prerequisites

Before starting the acceptance tests, ensure the following prerequisites are met:

| Prerequisite | Status | Notes |
|-------------|--------|-------|
| ☐ Ollama is installed and running | ☐ | Version: _______ |
| ☐ At least one model is available (e.g., `llama2`) | ☐ | Model: _______ |
| ☐ LlamaGate is installed | ☐ | Method: _______ |
| ☐ Network connectivity verified | ☐ | |
| ☐ Test environment configured | ☐ | |
| ☐ API key configured (if authentication enabled) | ☐ | Key: _______ |
| ☐ Test tools available (curl, Python, etc.) | ☐ | |

**Prerequisites Notes:**
```
_________________________________________________________________
_________________________________________________________________
_________________________________________________________________
```

---

## Installation Verification

### Test 1.1: Verify LlamaGate Installation

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Check LlamaGate binary exists | Binary file exists and is executable | ☐ | |
| Verify binary version | Version information displays correctly | ☐ | |
| Check default port availability | Port 11435 is available or configurable | ☐ | |

**Test Command:**
```bash
# Windows
llamagate.exe --version

# Unix/Linux/macOS
./llamagate --version
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 1.2: Verify Configuration File

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Check `.env` file exists | Configuration file is present | ☐ | |
| Verify configuration values | All required settings are present | ☐ | |
| Test configuration loading | No errors when loading config | ☐ | |

**Configuration Checklist:**
- ☐ `OLLAMA_HOST` configured (default: `http://localhost:11434`)
- ☐ `PORT` configured (default: `11435`)
- ☐ `API_KEY` configured (optional, leave empty to disable authentication)
- ☐ `RATE_LIMIT_RPS` configured (default: `50`)
- ☐ `DEBUG` configured (default: `false`)
- ☐ `LOG_FILE` configured (optional, leave empty for console only)
- ☐ `TIMEOUT` configured (default: `5m`)

**Notes:**
```
_________________________________________________________________
_________________________________________________________________
```

---

## Basic Functionality

### Test 2.1: Server Startup

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Start LlamaGate server | Server starts without errors | ☐ | |
| Verify server is listening | Server responds on configured port | ☐ | |
| Check startup logs | No error messages in logs | ☐ | |

**Test Command:**
```bash
# Start server
llamagate

# In another terminal, verify it's running
curl http://localhost:11435/health
```

**Expected Response:**
```json
{"status":"healthy"}
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 2.2: Health Check Endpoint

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| GET `/health` without auth | Returns `200 OK` with `{"status":"healthy"}` | ☐ | |
| Verify response format | Valid JSON response | ☐ | |
| Check response time | Response time < 100ms | ☐ | |

**Test Command:**
```bash
curl http://localhost:11435/health
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 2.3: List Models Endpoint

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| GET `/v1/models` | Returns list of available models | ☐ | |
| Verify response format | Matches OpenAI API format | ☐ | |
| Check model data | Model IDs match Ollama models | ☐ | |

**Test Command:**
```bash
curl http://localhost:11435/v1/models \
  -H "X-API-Key: sk-llamagate"
```

**Expected Response Format:**
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

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 2.4: Chat Completions (Non-Streaming)

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| POST `/v1/chat/completions` | Returns chat completion response | ☐ | |
| Verify response format | Matches OpenAI API format | ☐ | |
| Check response content | Contains valid assistant message | ☐ | |
| Verify model field | Model matches request | ☐ | |
| Check finish_reason | Finish reason is present | ☐ | |

**Test Command:**
```bash
curl -X POST http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "model": "llama2",
    "messages": [
      {"role": "user", "content": "Say hello in one sentence."}
    ]
  }'
```

**Expected Response Format:**
```json
{
  "id": "chatcmpl-...",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "llama2",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Hello! ..."
    },
    "finish_reason": "stop"
  }]
}
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 2.5: Multiple Message Conversation

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| POST with multiple messages | Handles conversation context | ☐ | |
| Verify context maintained | Response references previous messages | ☐ | |

**Test Command:**
```bash
curl -X POST http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "model": "llama2",
    "messages": [
      {"role": "user", "content": "My name is Alice."},
      {"role": "assistant", "content": "Nice to meet you, Alice!"},
      {"role": "user", "content": "What is my name?"}
    ]
  }'
```

**Expected:** Response should reference "Alice"

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

---

## Authentication & Security

### Test 3.1: Authentication Required (When Enabled)

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Request without API key | Returns `401 Unauthorized` | ☐ | |
| Request with invalid key | Returns `401 Unauthorized` | ☐ | |
| Request with valid key | Returns `200 OK` | ☐ | |

**Test Commands:**
```bash
# Test 1: No API key
curl -w "\nHTTP Status: %{http_code}\n" \
  http://localhost:11435/v1/models

# Test 2: Invalid API key
curl -w "\nHTTP Status: %{http_code}\n" \
  -H "X-API-Key: wrong-key" \
  http://localhost:11435/v1/models

# Test 3: Valid API key
curl -w "\nHTTP Status: %{http_code}\n" \
  -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/models
```

**Expected Results:**
- Test 1: `401 Unauthorized`
- Test 2: `401 Unauthorized`
- Test 3: `200 OK`

**Actual Results:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 3.2: Bearer Token Authentication

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Request with Bearer token | Returns `200 OK` | ☐ | |
| Verify case-insensitive | Works with "bearer", "Bearer", "BEARER" | ☐ | |

**Test Command:**
```bash
curl -H "Authorization: Bearer sk-llamagate" \
  http://localhost:11435/v1/models
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 3.3: API Key Not Logged

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Make request with API key | API key does not appear in logs | ☐ | |
| Check log files | Sensitive data is redacted | ☐ | |

**Test Command:**
```bash
# Make a request
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/models

# Check logs (adjust path as needed)
cat llamagate.log | grep -i "api-key"
```

**Expected:** No API key values in logs

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 3.4: Health Endpoint Bypass

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| GET `/health` without auth | Always returns `200 OK` | ☐ | |
| Verify bypass works | Health check doesn't require auth | ☐ | |

**Test Command:**
```bash
curl http://localhost:11435/health
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

---

## Streaming

### Test 4.1: Streaming Chat Completions

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| POST with `"stream": true` | Returns Server-Sent Events stream | ☐ | |
| Verify SSE format | Each chunk is `data: {...}` | ☐ | |
| Check final marker | Stream ends with `data: [DONE]` | ☐ | |
| Verify incremental content | Content appears progressively | ☐ | |

**Test Command:**
```bash
curl -X POST http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "model": "llama2",
    "messages": [
      {"role": "user", "content": "Count from 1 to 5"}
    ],
    "stream": true
  }'
```

**Expected Response Format:**
```
data: {"id":"chatcmpl-...","object":"chat.completion.chunk","created":1234567890,"model":"llama2","choices":[{"index":0,"delta":{"content":"1"},"finish_reason":null}]}

data: {"id":"chatcmpl-...","object":"chat.completion.chunk","created":1234567890,"model":"llama2","choices":[{"index":0,"delta":{"content":" 2"},"finish_reason":null}]}

...

data: [DONE]
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 4.2: Streaming with Python SDK

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Use OpenAI Python SDK | Streaming works correctly | ☐ | |
| Verify chunk reception | Chunks received incrementally | ☐ | |

**Test Script:**
```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:11435/v1",
    api_key="sk-llamagate"
)

stream = client.chat.completions.create(
    model="llama2",
    messages=[{"role": "user", "content": "Say hello"}],
    stream=True
)

for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="", flush=True)
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

---

## Caching

### Test 5.1: Cache Hit Verification

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| First request (cache miss) | Response from Ollama (slower) | ☐ | |
| Second identical request | Response from cache (faster) | ☐ | |
| Verify response identical | Cached response matches original | ☐ | |
| Check response time | Cached response < 100ms | ☐ | |

**Test Commands:**
```bash
# First request (cache miss)
curl -w "\nTime: %{time_total}s\n" \
  -X POST http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "model": "llama2",
    "messages": [
      {"role": "user", "content": "What is 2+2?"}
    ]
  }'

# Second request (cache hit)
curl -w "\nTime: %{time_total}s\n" \
  -X POST http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "model": "llama2",
    "messages": [
      {"role": "user", "content": "What is 2+2?"}
    ]
  }'
```

**Expected:** Second request should be significantly faster (< 100ms vs seconds)

**Actual Results:**
- First request time: _______ seconds
- Second request time: _______ seconds

**Notes:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 5.2: Cache Key Generation

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Different messages | Different cache keys | ☐ | |
| Different models | Different cache keys | ☐ | |
| Different parameters | Different cache keys | ☐ | |

**Test Commands:**
```bash
# Request 1
curl -X POST http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{"model":"llama2","messages":[{"role":"user","content":"Hello"}]}'

# Request 2 (different message)
curl -X POST http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{"model":"llama2","messages":[{"role":"user","content":"Hi"}]}'
```

**Expected:** Both should be cache misses (slow)

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

---

## Rate Limiting

### Test 6.1: Rate Limit Enforcement

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Make requests within limit | All requests succeed | ☐ | |
| Exceed rate limit | Returns `429 Too Many Requests` | ☐ | |
| Check Retry-After header | Header present in 429 response | ☐ | |
| Verify limit resets | Requests succeed after wait period | ☐ | |

**Test Command:**
```bash
# Make 60 rapid requests (assuming default limit is 50 RPS)
for i in {1..60}; do
  curl -w "\nRequest $i: HTTP %{http_code}\n" \
    -X POST http://localhost:11435/v1/chat/completions \
    -H "Content-Type: application/json" \
    -H "X-API-Key: sk-llamagate" \
    -d '{"model":"llama2","messages":[{"role":"user","content":"Test"}]}' \
    -o /dev/null -s
  sleep 0.1
done
```

**Expected:** First 50 requests succeed, then 429 errors

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 6.2: Health Endpoint Bypass

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Rapid health checks | Never rate limited | ☐ | |

**Test Command:**
```bash
for i in {1..100}; do
  curl -s http://localhost:11435/health
done
```

**Expected:** All requests return 200 OK

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

---

## MCP Integration

**Note:** These tests require MCP servers to be configured. Skip if MCP is not enabled.

### Test 7.1: List MCP Servers

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| GET `/v1/mcp/servers` | Returns list of configured servers | ☐ | |
| Verify server metadata | Server names and status present | ☐ | |

**Test Command:**
```bash
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/mcp/servers
```

**Expected Response Format:**
```json
{
  "servers": [
    {
      "name": "filesystem",
      "status": "healthy",
      "transport": "http"
    }
  ]
}
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 7.2: Server Health Check

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| GET `/v1/mcp/servers/:name/health` | Returns server health status | ☐ | |
| Verify health states | Status is "healthy", "unhealthy", or "unknown" | ☐ | |

**Test Command:**
```bash
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/mcp/servers/filesystem/health
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 7.3: List Server Tools

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| GET `/v1/mcp/servers/:name/tools` | Returns list of available tools | ☐ | |
| Verify tool metadata | Tool names, descriptions present | ☐ | |

**Test Command:**
```bash
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/mcp/servers/filesystem/tools
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 7.4: Execute MCP Tool

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| POST `/v1/mcp/execute` | Executes tool and returns result | ☐ | |
| Verify tool execution | Tool performs expected action | ☐ | |
| Check error handling | Invalid tools return appropriate errors | ☐ | |

**Test Command:**
```bash
curl -X POST -H "X-API-Key: sk-llamagate" \
  -H "Content-Type: application/json" \
  -d '{
    "server": "filesystem",
    "tool": "read_file",
    "arguments": {
      "path": "/tmp/test.txt"
    }
  }' \
  http://localhost:11435/v1/mcp/execute
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 7.5: MCP URI Scheme in Chat

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Chat with MCP URI | Resource content injected into context | ☐ | |
| Verify context enhancement | Response references resource content | ☐ | |

**Test Command:**
```bash
curl -X POST http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "model": "llama2",
    "messages": [
      {"role": "user", "content": "Summarize mcp://filesystem/file:///docs/readme.txt"}
    ]
  }'
```

**Expected:** Response should include content from the file

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

---

## Extension System

**Note:** These tests require extensions to be configured. Skip if extensions are not configured.

### Test 8.1: List Extensions

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| GET `/v1/extensions` | Returns list of registered extensions | ☐ | |
| Verify extension metadata | Extension names, descriptions present | ☐ | |

**Test Command:**
```bash
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/extensions
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 8.2: Get Extension Details

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| GET `/v1/extensions/:name` | Returns extension metadata | ☐ | |
| Verify input schema | Input parameters documented | ☐ | |

**Test Command:**
```bash
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/extensions/prompt-template-executor
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 8.3: Execute Extension

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| POST `/v1/extensions/:name/execute` | Executes extension and returns result | ☐ | |
| Verify input validation | Invalid inputs return 400 Bad Request | ☐ | |
| Check execution result | Valid inputs return 200 OK with results | ☐ | |

**Test Command:**
```bash
curl -X POST -H "X-API-Key: sk-llamagate" \
  -H "Content-Type: application/json" \
  -d '{
    "input": {
      "template": "Hello {{name}}",
      "variables": {"name": "World"}
    }
  }' \
  http://localhost:11435/v1/extensions/prompt-template-executor/execute
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

---

## Error Handling

### Test 9.1: Invalid Model

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Request with non-existent model | Returns appropriate error | ☐ | |
| Verify error format | Matches OpenAI error format | ☐ | |

**Test Command:**
```bash
curl -X POST http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "model": "non-existent-model",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

**Expected:** Error response with appropriate message

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 9.2: Invalid Request Format

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Malformed JSON | Returns 400 Bad Request | ☐ | |
| Missing required fields | Returns 400 Bad Request | ☐ | |

**Test Commands:**
```bash
# Malformed JSON
curl -X POST http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{invalid json}'

# Missing required fields
curl -X POST http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{}'
```

**Expected:** Both return 400 Bad Request

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 9.3: Ollama Connection Failure

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Stop Ollama service | LlamaGate handles gracefully | ☐ | |
| Make request | Returns appropriate error | ☐ | |
| Restart Ollama | LlamaGate recovers automatically | ☐ | |

**Test Steps:**
1. Stop Ollama: `pkill ollama` (Unix) or stop service (Windows)
2. Make a request
3. Restart Ollama: `ollama serve`
4. Make another request

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

---

## Logging & Monitoring

### Test 10.1: Structured Logging

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Check log file exists | Log file is created (if configured) | ☐ | |
| Verify JSON format | Logs are in JSON format | ☐ | |
| Check request correlation | Request IDs present in logs | ☐ | |

**Test Command:**
```bash
# Make a request
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/models

# Check logs
tail -n 20 llamagate.log
```

**Expected:** JSON logs with request IDs

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 10.2: Request ID Correlation

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Send custom request ID | Same ID appears in response and logs | ☐ | |
| Verify correlation | All related log entries share request ID | ☐ | |

**Test Command:**
```bash
curl -H "X-API-Key: sk-llamagate" \
  -H "X-Request-ID: test-request-123" \
  http://localhost:11435/v1/models
```

**Expected:** Response header contains `X-Request-ID: test-request-123`

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 10.3: Log Levels

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Set DEBUG mode | Debug logs appear | ☐ | |
| Set INFO mode | Only info and above appear | ☐ | |

**Test:** Change `LOG_LEVEL` in `.env` and restart server

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

---

## Performance

### Test 11.1: Response Time

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Measure health endpoint | Response time < 100ms | ☐ | |
| Measure models endpoint | Response time < 500ms | ☐ | |
| Measure chat completion | Response time reasonable for model | ☐ | |

**Test Commands:**
```bash
# Health check
curl -w "\nTime: %{time_total}s\n" \
  http://localhost:11435/health

# Models
curl -w "\nTime: %{time_total}s\n" \
  -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/models

# Chat completion
curl -w "\nTime: %{time_total}s\n" \
  -X POST http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{"model":"llama2","messages":[{"role":"user","content":"Hi"}]}'
```

**Actual Results:**
- Health: _______ seconds
- Models: _______ seconds
- Chat: _______ seconds

### Test 11.2: Concurrent Requests

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Send 10 concurrent requests | All requests handled correctly | ☐ | |
| Verify no errors | No 500 errors or timeouts | ☐ | |

**Test Command:**
```bash
# Send 10 concurrent requests
for i in {1..10}; do
  curl -H "X-API-Key: sk-llamagate" \
    http://localhost:11435/v1/models &
done
wait
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

---

## OpenAI SDK Compatibility

### Test 12.1: Python SDK

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Install OpenAI SDK | SDK installs successfully | ☐ | |
| Configure client | Client connects to LlamaGate | ☐ | |
| List models | Models endpoint works | ☐ | |
| Chat completion | Chat completions work | ☐ | |
| Streaming | Streaming works | ☐ | |

**Test Script:**
```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:11435/v1",
    api_key="sk-llamagate"
)

# Test models
models = client.models.list()
print("Models:", models)

# Test chat
response = client.chat.completions.create(
    model="llama2",
    messages=[{"role": "user", "content": "Hello!"}]
)
print("Response:", response.choices[0].message.content)

# Test streaming
stream = client.chat.completions.create(
    model="llama2",
    messages=[{"role": "user", "content": "Count to 3"}],
    stream=True
)
for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

### Test 12.2: Node.js SDK

| Test Step | Expected Result | Status | Notes |
|-----------|----------------|--------|-------|
| Install OpenAI SDK | SDK installs successfully | ☐ | |
| Configure client | Client connects to LlamaGate | ☐ | |
| List models | Models endpoint works | ☐ | |
| Chat completion | Chat completions work | ☐ | |

**Test Script:**
```javascript
import OpenAI from 'openai';

const client = new OpenAI({
  baseURL: 'http://localhost:11435/v1',
  apiKey: 'sk-llamagate',
});

// Test models
const models = await client.models.list();
console.log('Models:', models);

// Test chat
const response = await client.chat.completions.create({
  model: 'llama2',
  messages: [{ role: 'user', content: 'Hello!' }],
});
console.log('Response:', response.choices[0].message.content);
```

**Actual Result:**
```
_________________________________________________________________
_________________________________________________________________
```

---

## Issues & Observations

### Critical Issues

| Issue # | Description | Severity | Status | Notes |
|---------|-------------|----------|--------|-------|
| 1 | | | ☐ | |
| 2 | | | ☐ | |
| 3 | | | ☐ | |

### Minor Issues

| Issue # | Description | Severity | Status | Notes |
|---------|-------------|----------|--------|-------|
| 1 | | | ☐ | |
| 2 | | | ☐ | |
| 3 | | | ☐ | |

### Observations

```
_________________________________________________________________
_________________________________________________________________
_________________________________________________________________
_________________________________________________________________
_________________________________________________________________
_________________________________________________________________
```

---

## Test Summary

### Overall Test Results

| Category | Tests Passed | Tests Failed | Tests Skipped | Notes |
|----------|--------------|--------------|---------------|-------|
| Installation | ___ / ___ | ___ / ___ | ___ / ___ | |
| Basic Functionality | ___ / ___ | ___ / ___ | ___ / ___ | |
| Authentication | ___ / ___ | ___ / ___ | ___ / ___ | |
| Streaming | ___ / ___ | ___ / ___ | ___ / ___ | |
| Caching | ___ / ___ | ___ / ___ | ___ / ___ | |
| Rate Limiting | ___ / ___ | ___ / ___ | ___ / ___ | |
| MCP Integration | ___ / ___ | ___ / ___ | ___ / ___ | |
| Extension System | ___ / ___ | ___ / ___ | ___ / ___ | |
| Error Handling | ___ / ___ | ___ / ___ | ___ / ___ | |
| Logging | ___ / ___ | ___ / ___ | ___ / ___ | |
| Performance | ___ / ___ | ___ / ___ | ___ / ___ | |
| SDK Compatibility | ___ / ___ | ___ / ___ | ___ / ___ | |

**Total:** ___ Passed / ___ Failed / ___ Skipped

### Acceptance Decision

☐ **ACCEPTED** - All critical tests passed, system is ready for production  
☐ **CONDITIONALLY ACCEPTED** - Minor issues present, acceptable for production  
☐ **REJECTED** - Critical issues present, system not ready for production  

### Signatures

**Tester:**  
Name: _______________  
Signature: _______________  
Date: _______________  

**Reviewer (if applicable):**  
Name: _______________  
Signature: _______________  
Date: _______________  

---

## Appendix A: Test Environment Details

**Operating System:** _______________  
**LlamaGate Version:** _______________  
**Ollama Version:** _______________  
**Go Version:** _______________  
**Network Configuration:** _______________  
**Test Tools Used:** _______________  

---

## Appendix B: Configuration Used

```
_________________________________________________________________
_________________________________________________________________
_________________________________________________________________
_________________________________________________________________
_________________________________________________________________
```

---

**End of Acceptance Test Document**
