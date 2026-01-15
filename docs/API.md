# LlamaGate HTTP API Documentation

## Overview

LlamaGate provides a RESTful HTTP API for managing MCP (Model Context Protocol) servers and accessing their capabilities. All endpoints are available under the `/v1/mcp` prefix.

## Authentication

All MCP API endpoints require authentication when `API_KEY` is configured.

### Supported Authentication Headers

LlamaGate supports two authentication header formats (both are case-insensitive):

#### 1. X-API-Key Header (Recommended)
```bash
curl -H "X-API-Key: sk-llamagate" http://localhost:11435/v1/mcp/servers
```

The header name is case-insensitive. All of the following are accepted:
- `X-API-Key`
- `x-api-key`
- `X-Api-Key`
- Any other case variation

#### 2. Authorization Bearer Header (Alternative)
```bash
curl -H "Authorization: Bearer sk-llamagate" http://localhost:11435/v1/mcp/servers
```

The "Bearer" scheme is case-insensitive. All of the following are accepted:
- `Authorization: Bearer sk-llamagate`
- `Authorization: bearer sk-llamagate`
- `Authorization: BEARER sk-llamagate`

**Header Priority:** The `X-API-Key` header is checked first. If not present, `Authorization: Bearer` is checked.

### Authentication Behavior

- **When Authentication is Required:** All endpoints require authentication when `API_KEY` is configured.
- **When Authentication is Missing:** Requests without a valid authentication header return `401 Unauthorized` with an OpenAI-compatible error response.
- **When Authentication is Invalid:** Requests with an invalid or incorrect API key return `401 Unauthorized` with an OpenAI-compatible error response.

### Security

- **API keys are never logged:** Authentication failures are logged with a generic message ("Authentication failed") but the provided API key or bearer token is never included in logs.
- **Constant-time comparison:** API key validation uses constant-time comparison to prevent timing attacks.

## Request ID and Correlation

LlamaGate implements consistent request correlation across all components for easier troubleshooting and debugging.

### Request ID Header

LlamaGate supports the `X-Request-ID` header for request correlation:

- **If provided**: LlamaGate uses your provided request ID and includes it in the response
- **If not provided**: LlamaGate generates a UUID v4 request ID

**Example:**
```bash
curl -H "X-API-Key: sk-llamagate" \
     -H "X-Request-ID: my-custom-request-id-123" \
     http://localhost:11435/v1/chat/completions
```

The same request ID will appear in:
- Response header: `X-Request-ID: my-custom-request-id-123`
- All log entries for this request
- Error responses (if any)
- Downstream calls to Ollama (via `X-Request-ID` header)
- Tool/function calls (via context)
- MCP tool calls (via context and HTTP headers)

### Sensitive Data Redaction

LlamaGate automatically redacts sensitive values from all logs:

**Redacted Values:**
- API keys (`X-API-Key` header)
- Bearer tokens (`Authorization: Bearer` header)
- Secrets in headers, environment variables, or configuration

**Logging Behavior:**
- ✅ Request method, path, status, latency, IP address
- ✅ Request ID for correlation
- ✅ Error messages (without sensitive data)
- ❌ API key values
- ❌ Bearer token values
- ❌ Authorization header contents

This ensures that secrets never appear in log files, even if accidentally included in request headers.

## Base URL

All endpoints are prefixed with `/v1/mcp`:

```
http://localhost:11435/v1/mcp
```

## Endpoints

### Server Management

#### List All Servers

Get information about all configured MCP servers.

**Endpoint:** `GET /v1/mcp/servers`

**Response:**
```json
{
  "servers": [
    {
      "name": "filesystem",
      "transport": "stdio",
      "status": "healthy",
      "health": {
        "status": "healthy",
        "last_check": "2026-01-07T12:00:00Z",
        "last_success": "2026-01-07T12:00:00Z",
        "latency": "50ms"
      },
      "tools": 5,
      "resources": 10,
      "prompts": 2
    }
  ],
  "count": 1
}
```

#### Get Server Information

Get detailed information about a specific server.

**Endpoint:** `GET /v1/mcp/servers/:name`

**Parameters:**
- `name` (path) - Server name

**Response:**
```json
{
  "name": "filesystem",
  "transport": "stdio",
  "status": "healthy",
  "health": { ... },
  "tools": 5,
  "resources": 10,
  "prompts": 2
}
```

#### Get Server Health

Get current health status for a specific server.

**Endpoint:** `GET /v1/mcp/servers/:name/health`

**Parameters:**
- `name` (path) - Server name

**Response:**
```json
{
  "server": "filesystem",
  "status": "healthy",
  "health": {
    "status": "healthy",
    "last_check": "2026-01-07T12:00:00Z",
    "last_success": "2026-01-07T12:00:00Z",
    "latency": "50ms"
  }
}
```

#### Get All Server Health

Get health status for all servers.

**Endpoint:** `GET /v1/mcp/servers/health`

**Response:**
```json
{
  "servers": {
    "filesystem": {
      "status": "healthy",
      "health": { ... }
    }
  },
  "count": 1
}
```

#### Get Server Statistics

Get connection pool statistics for a server (HTTP transport only).

**Endpoint:** `GET /v1/mcp/servers/:name/stats`

**Parameters:**
- `name` (path) - Server name

**Response:**
```json
{
  "server": "remote-server",
  "stats": {
    "pool": {
      "total": 3,
      "in_use": 1,
      "idle": 2,
      "max_allowed": 10
    }
  }
}
```

#### Refresh Server Metadata

Refresh tools, resources, and prompts for a server.

**Endpoint:** `POST /v1/mcp/servers/:name/refresh`

**Parameters:**
- `name` (path) - Server name

**Response (Success):**
```json
{
  "server": "filesystem",
  "status": "refreshed"
}
```

**Response (Partial Failure):**
```json
{
  "server": "filesystem",
  "status": "partial",
  "errors": [
    "tools: connection timeout",
    "resources: server error"
  ]
}
```

**Status Codes:**
- `200 OK` - All metadata refreshed successfully
- `206 Partial Content` - Some refresh operations failed (check `errors` array)
- `404 Not Found` - Server not found
- `503 Service Unavailable` - MCP not enabled

### Tools

#### List Server Tools

Get all tools available from a server.

**Endpoint:** `GET /v1/mcp/servers/:name/tools`

**Parameters:**
- `name` (path) - Server name

**Response:**
```json
{
  "server": "filesystem",
  "tools": [
    {
      "name": "read_file",
      "description": "Read a file",
      "inputSchema": {
        "type": "object",
        "properties": {
          "path": {
            "type": "string",
            "description": "File path"
          }
        }
      }
    }
  ],
  "count": 1
}
```

### Resources

#### List Server Resources

Get all resources available from a server.

**Endpoint:** `GET /v1/mcp/servers/:name/resources`

**Parameters:**
- `name` (path) - Server name

**Response:**
```json
{
  "server": "filesystem",
  "resources": [
    {
      "uri": "file:///path/to/file.txt",
      "name": "file.txt",
      "description": "A text file",
      "mimeType": "text/plain"
    }
  ],
  "count": 1
}
```

#### Read Resource

Read the content of a specific resource.

**Endpoint:** `GET /v1/mcp/servers/:name/resources/*uri`

**Parameters:**
- `name` (path) - Server name
- `uri` (path) - Resource URI (can also be provided as query parameter `?uri=...`)

**Response:**
```json
{
  "server": "filesystem",
  "uri": "file:///path/to/file.txt",
  "contents": [
    {
      "uri": "file:///path/to/file.txt",
      "mimeType": "text/plain",
      "text": "File contents here"
    }
  ]
}
```

### Prompts

#### List Server Prompts

Get all prompts available from a server.

**Endpoint:** `GET /v1/mcp/servers/:name/prompts`

**Parameters:**
- `name` (path) - Server name

**Response:**
```json
{
  "server": "filesystem",
  "prompts": [
    {
      "name": "summarize",
      "description": "Summarize text",
      "arguments": [
        {
          "name": "text",
          "description": "Text to summarize",
          "required": true,
          "type": "string"
        }
      ]
    }
  ],
  "count": 1
}
```

#### Get Prompt Template

Get a rendered prompt template with provided arguments.

**Endpoint:** `POST /v1/mcp/servers/:name/prompts/:promptName`

**Parameters:**
- `name` (path) - Server name
- `promptName` (path) - Prompt name

**Request Body:**
```json
{
  "arguments": {
    "text": "Text to summarize"
  }
}
```

**Response:**
```json
{
  "server": "filesystem",
  "prompt": "summarize",
  "messages": [
    {
      "role": "user",
      "content": {
        "type": "text",
        "text": "Summarize this text: Text to summarize"
      }
    }
  ]
}
```

### Tool Execution

#### Execute Tool

Execute a tool on a specific server.

**Endpoint:** `POST /v1/mcp/execute`

**Request Body:**
```json
{
  "server": "filesystem",
  "tool": "read_file",
  "arguments": {
    "path": "/path/to/file.txt"
  }
}
```

**Response (Success):**
```json
{
  "server": "filesystem",
  "tool": "read_file",
  "result": "File contents here",
  "is_error": false,
  "duration": "50ms"
}
```

**Response (Error):**
```json
{
  "server": "filesystem",
  "tool": "read_file",
  "error": "File not found",
  "is_error": true,
  "duration": "10ms"
}
```

## Error Responses

All endpoints return standard error responses when an error occurs:

```json
{
  "error": {
    "message": "Error description",
    "type": "error_type",
    "details": "Additional error details (optional)"
  }
}
```

### Error Types

- `service_unavailable` - MCP is not enabled or service is unavailable
- `not_found` - Server, resource, tool, or prompt not found
- `invalid_request_error` - Invalid request parameters or body
- `internal_error` - Internal server error
- `rate_limit_error` - Rate limit exceeded

### HTTP Status Codes

- `200 OK` - Success
- `400 Bad Request` - Invalid request parameters or body
- `401 Unauthorized` - Authentication required or invalid API key
- `404 Not Found` - Resource not found
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - MCP not enabled

### Error Response Examples

#### MCP Not Enabled

**Status:** `503 Service Unavailable`

```json
{
  "error": {
    "message": "MCP is not enabled",
    "type": "service_unavailable"
  }
}
```

#### Server Not Found

**Status:** `404 Not Found`

```json
{
  "error": {
    "message": "Server not found",
    "type": "not_found"
  }
}
```

#### Invalid Request Body

**Status:** `400 Bad Request`

```json
{
  "error": {
    "message": "Invalid request body for prompt arguments",
    "type": "invalid_request_error",
    "details": "json: cannot unmarshal string into Go value of type map[string]interface {}"
  }
}
```

#### Resource Not Found

**Status:** `500 Internal Server Error`

```json
{
  "error": {
    "message": "Failed to read resource",
    "type": "internal_error",
    "details": "Resource not found"
  }
}
```

#### Authentication Required

**Status:** `401 Unauthorized`

```json
{
  "error": {
    "message": "Invalid API key",
    "type": "invalid_request_error",
    "request_id": "req-123456"
  }
}
```

**Note:** This error is returned when:
- No authentication header is provided
- An invalid or incorrect API key is provided
- The authentication header format is invalid (e.g., `Authorization: Bearer` without a token)

#### Rate Limit Exceeded

**Status:** `429 Too Many Requests`

**Headers:**
```
Retry-After: 1
```

The `Retry-After` header indicates the number of seconds to wait before retrying the request. The value is calculated based on the rate limiter's current state and is always at least 1 second.

**Response Body:**
```json
{
  "error": {
    "message": "Rate limit exceeded",
    "type": "rate_limit_error",
    "request_id": "req-123456"
  }
}
```

**Note:** Rate-limited requests are logged with structured fields including:
- `request_id` - Unique request identifier
- `ip` - Client IP address
- `path` - Request path
- `retry_after` - Duration until next request is allowed
- `retry_after_seconds` - Retry-After header value in seconds
- `limiter_decision` - Always "rate_limited" for rate-limited requests

## Examples

### List All Servers

```bash
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/mcp/servers
```

### Get Server Health

```bash
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/mcp/servers/filesystem/health
```

### List Server Tools

```bash
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/mcp/servers/filesystem/tools
```

### Execute Tool

```bash
curl -X POST \
  -H "X-API-Key: sk-llamagate" \
  -H "Content-Type: application/json" \
  -d '{"server":"filesystem","tool":"read_file","arguments":{"path":"/tmp/test.txt"}}' \
  http://localhost:11435/v1/mcp/execute
```

### Read Resource

```bash
curl -H "X-API-Key: sk-llamagate" \
  "http://localhost:11435/v1/mcp/servers/filesystem/resources/file%3A%2F%2F%2Ftmp%2Ftest.txt"
```

### Refresh Server Metadata

```bash
curl -X POST \
  -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/mcp/servers/filesystem/refresh
```

## Rate Limiting

All endpoints are subject to rate limiting as configured in the server settings. The default rate limit is 50 requests per second, configurable via `RATE_LIMIT_RPS`.

When the rate limit is exceeded:
- Requests receive HTTP `429 Too Many Requests` status
- A `Retry-After` header is included indicating seconds to wait before retrying
- Response body follows OpenAI-compatible error format with `rate_limit_error` type
- Structured logs capture the rate limit decision with request ID and retry information

## Extensions API

LlamaGate provides a RESTful API for managing and executing extensions. All extension endpoints are available under the `/v1/extensions` prefix.

### List All Extensions

Get information about all registered extensions.

**Endpoint:** `GET /v1/extensions`

**Response:**
```json
{
  "extensions": [
    {
      "name": "prompt-template-executor",
      "version": "1.0.0",
      "description": "Execute approved prompt templates with structured inputs.",
      "type": "workflow",
      "enabled": true
    },
    {
      "name": "request-inspector",
      "version": "1.0.0",
      "description": "Redacted audit logging for inbound and outbound requests.",
      "type": "middleware",
      "enabled": true
    },
    {
      "name": "cost-usage-reporter",
      "version": "1.0.0",
      "description": "Track token usage and estimated cost per request.",
      "type": "observer",
      "enabled": true
    }
  ],
  "count": 3
}
```

**Example:**
```bash
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/extensions
```

### Get Extension Details

Get detailed information about a specific extension, including input/output schemas.

**Endpoint:** `GET /v1/extensions/:name`

**Parameters:**
- `name` (path) - Extension name

**Response:**
```json
{
  "name": "prompt-template-executor",
  "version": "1.0.0",
  "description": "Execute approved prompt templates with structured inputs.",
  "type": "workflow",
  "enabled": true,
  "inputs": [
    {
      "id": "template_id",
      "type": "string",
      "required": true
    },
    {
      "id": "variables",
      "type": "object",
      "required": true
    }
  ],
  "outputs": [
    {
      "id": "result",
      "type": "file",
      "path": "./output/result.md"
    }
  ]
}
```

**Example:**
```bash
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/extensions/prompt-template-executor
```

**Error Responses:**
- `404 Not Found` - Extension not found
- `503 Service Unavailable` - Extension system not available

### Execute Workflow Extension

Execute a workflow extension with provided inputs.

**Endpoint:** `POST /v1/extensions/:name/execute`

**Parameters:**
- `name` (path) - Extension name (must be a workflow extension)

**Request Body:**
```json
{
  "template_id": "example",
  "variables": {
    "document_type": "executive summary",
    "format": "markdown"
  },
  "model": "llama3.2"
}
```

**Response (Success):**
```json
{
  "success": true,
  "data": {
    "output_file": "/path/to/extensions/prompt-template-executor/output/result.md"
  }
}
```

**Response (Error):**
```json
{
  "error": {
    "message": "Extension execution failed: template_id is required",
    "type": "invalid_request_error",
    "request_id": "req-123456"
  }
}
```

**Example:**
```bash
curl -X POST \
  -H "X-API-Key: sk-llamagate" \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "example",
    "variables": {
      "document_type": "executive summary",
      "format": "markdown"
    }
  }' \
  http://localhost:11435/v1/extensions/prompt-template-executor/execute
```

**Status Codes:**
- `200 OK` - Execution successful
- `400 Bad Request` - Invalid input or extension type
- `404 Not Found` - Extension not found
- `503 Service Unavailable` - Extension is disabled or system unavailable

**Notes:**
- Only workflow extensions can be executed via this endpoint
- Middleware and observer extensions run automatically (no API call needed)
- All required inputs must be provided
- Output files are written to the extension's output directory

## Health Endpoint

The main health endpoint (`/health`) does not require authentication and can be used for monitoring:

```bash
curl http://localhost:11435/health
```

