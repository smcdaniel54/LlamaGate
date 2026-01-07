# LlamaGate HTTP API Documentation

## Overview

LlamaGate provides a RESTful HTTP API for managing MCP (Model Context Protocol) servers and accessing their capabilities. All endpoints are available under the `/v1/mcp` prefix.

## Authentication

All MCP API endpoints require authentication when `API_KEY` is configured. Use either:

- **X-API-Key header**: `X-API-Key: sk-llamagate`
- **Authorization Bearer header**: `Authorization: Bearer sk-llamagate`

## Base URL

All endpoints are prefixed with `/v1/mcp`:

```
http://localhost:8080/v1/mcp
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

**Response:**
```json
{
  "server": "filesystem",
  "status": "refreshed"
}
```

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

All endpoints return standard error responses:

```json
{
  "error": {
    "message": "Error description",
    "type": "error_type"
  }
}
```

**Error Types:**
- `service_unavailable` - MCP is not enabled
- `not_found` - Server or resource not found
- `invalid_request_error` - Invalid request parameters
- `internal_error` - Server error

**HTTP Status Codes:**
- `200 OK` - Success
- `400 Bad Request` - Invalid request
- `401 Unauthorized` - Authentication required
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - MCP not enabled

## Examples

### List All Servers

```bash
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:8080/v1/mcp/servers
```

### Get Server Health

```bash
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:8080/v1/mcp/servers/filesystem/health
```

### List Server Tools

```bash
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:8080/v1/mcp/servers/filesystem/tools
```

### Execute Tool

```bash
curl -X POST \
  -H "X-API-Key: sk-llamagate" \
  -H "Content-Type: application/json" \
  -d '{"server":"filesystem","tool":"read_file","arguments":{"path":"/tmp/test.txt"}}' \
  http://localhost:8080/v1/mcp/execute
```

### Read Resource

```bash
curl -H "X-API-Key: sk-llamagate" \
  "http://localhost:8080/v1/mcp/servers/filesystem/resources/file%3A%2F%2F%2Ftmp%2Ftest.txt"
```

### Refresh Server Metadata

```bash
curl -X POST \
  -H "X-API-Key: sk-llamagate" \
  http://localhost:8080/v1/mcp/servers/filesystem/refresh
```

## Rate Limiting

All endpoints are subject to rate limiting as configured in the server settings. The default rate limit is 10 requests per second.

## Health Endpoint

The main health endpoint (`/health`) does not require authentication and can be used for monitoring:

```bash
curl http://localhost:8080/health
```

