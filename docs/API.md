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

Most endpoints are prefixed with `/v1/mcp`. Hardware endpoints are under `/v1/hardware`:

```
http://localhost:11435/v1/mcp
http://localhost:11435/v1/hardware
```

## Hardware Detection

### Get Hardware Recommendations

Automatically detect system hardware (CPU, RAM, GPU, VRAM) and get recommended local LLM models based on your hardware capabilities. All models are verified to be available in Ollama and sourced from [Artificial Analysis Open Source Models](https://artificialanalysis.ai/models/open-source).

**Note:** The model recommendations data is embedded directly in the LlamaGate binary - no external files or configuration required. The data is automatically loaded when the server starts.

**Endpoint:** `GET /v1/hardware/recommendations`

**Authentication:** Not required (public endpoint, similar to `/health`)

**Response:**
```json
{
  "success": true,
  "data": {
    "hardware": {
      "cpu_cores": 8,
      "cpu_model": "Intel Core i7-9700K",
      "total_ram_gb": 32,
      "gpu_detected": true,
      "gpu_name": "NVIDIA GeForce RTX 3060",
      "gpu_vram_gb": 12,
      "detection_method": "nvidia-smi"
    },
    "hardware_group": "gpu_8_16gb_vram",
    "recommendations": [
      {
        "name": "Mistral 7B",
        "ollama_name": "mistral",
        "priority": 1,
        "description": "Best balance - quantized for optimal performance",
        "intelligence_score": 7.0,
        "parameters_b": 7.0,
        "min_ram_gb": 8,
        "min_vram_gb": 8,
        "quantized": true,
        "ollama_command": "ollama pull mistral",
        "use_cases": ["general chat", "fast responses", "production workloads"],
        "artificial_analysis_url": "https://artificialanalysis.ai/models/mistral-7b-instruct"
      },
      {
        "name": "Llama 3.2 11B",
        "ollama_name": "llama3.2:11b",
        "priority": 2,
        "description": "Better quality - quantized (requires 12GB+ VRAM)",
        "intelligence_score": 11.0,
        "parameters_b": 11.0,
        "min_ram_gb": 12,
        "min_vram_gb": 12,
        "quantized": true,
        "ollama_command": "ollama pull llama3.2:11b",
        "use_cases": ["general chat", "balanced performance", "quality tasks"],
        "artificial_analysis_url": "https://artificialanalysis.ai/models/llama-3-2-instruct-11b"
      },
      {
        "name": "Qwen 2.5 7B",
        "ollama_name": "qwen2.5:7b",
        "priority": 3,
        "description": "Multilingual option - quantized",
        "intelligence_score": 10.0,
        "parameters_b": 7.0,
        "min_ram_gb": 8,
        "min_vram_gb": 8,
        "quantized": true,
        "ollama_command": "ollama pull qwen2.5:7b",
        "use_cases": ["multilingual", "structured output", "code generation"],
        "artificial_analysis_url": "https://artificialanalysis.ai/models/qwen2-5-7b-instruct"
      }
    ]
  }
}
```

**Note:** The `recommendations` array contains **multiple models (typically 3-4) sorted by priority**. Priority 1 is the best overall match, Priority 2 is the second choice, etc. This gives users multiple options to choose from based on their specific needs (speed, quality, multilingual support, etc.).

**Response Fields:**
- `hardware`: Detected system specifications
  - `cpu_cores`: Number of logical CPU cores
  - `cpu_model`: CPU model name
  - `total_ram_gb`: Total system RAM in GB
  - `gpu_detected`: Whether a GPU was detected
  - `gpu_name`: GPU name (if detected)
  - `gpu_vram_gb`: GPU VRAM in GB (if detected)
  - `detection_method`: Method used for GPU detection (nvidia-smi, wmi, lspci, system_profiler)
- `hardware_group`: Classified hardware group ID
- `recommendations`: **Array of 3-4 recommended models, sorted by priority** (Priority 1 = best match, Priority 2 = second choice, etc.). Each model includes:
  - `name`: Human-readable model name
  - `ollama_name`: Ollama model identifier
  - `priority`: Recommendation priority (1 = highest, 2 = second choice, etc.). Models are sorted by this field.
  - `description`: Model description
  - `intelligence_score`: Artificial Analysis Intelligence Index (optional)
  - `parameters_b`: Model size in billions of parameters (optional)
  - `min_ram_gb`: Minimum RAM required
  - `min_vram_gb`: Minimum VRAM required (0 for CPU-only)
  - `quantized`: Whether the model is quantized
  - `ollama_command`: Ready-to-use Ollama pull command
  - `use_cases`: Recommended use cases
  - `artificial_analysis_url`: Link to detailed benchmarks (optional)

**Example:**
```bash
curl http://localhost:11435/v1/hardware/recommendations
```

**Hardware Groups:**
- `cpu_only_16_32gb`: CPU-only systems with 16-32GB RAM
- `cpu_only_32_64gb`: CPU-only systems with 32-64GB RAM (most common)
- `cpu_only_64gb_plus`: CPU-only systems with 64GB+ RAM
- `gpu_4_8gb_vram`: GPUs with 4-8GB VRAM
- `gpu_8_16gb_vram`: GPUs with 8-16GB VRAM (most common GPU config)
- `gpu_16_24gb_vram`: GPUs with 16-24GB VRAM
- `gpu_24_32gb_vram`: GPUs with 24-32GB VRAM
- `gpu_32gb_plus_vram`: GPUs with 32GB+ VRAM (enterprise)

**Status Codes:**
- `200 OK` - Hardware detected and recommendations provided
- `500 Internal Server Error` - Hardware detection failed

---

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

## Extensions API (Removed in Phase 1)

**The extension system has been removed.** LlamaGate is now a core-only, OpenAI-compatible gateway. All `/v1/extensions` endpoints (list, get, upsert, execute, refresh) and dynamic extension routes have been removed; requests to these paths return **404**.

For current API surface, see [Core Contract](core_contract.md). For migration notes, see the [README Migration Notes](../README.md#migration-notes-phase-1-extensionsmodules-removed).

---

*The following section is retained for historical reference only.*

### Extension Types (obsolete)

LlamaGate previously had three types of extensions:

1. **Builtin Extensions (Go Code)**: Core functionality compiled into binary (`internal/extensions/builtin/`) — removed
2. **Builtin Extensions (YAML-based)**: Core workflow capabilities in `extensions/builtin/` — removed
3. **Default Extensions (YAML-based)**: Workflow extensions in `extensions/` directory — removed

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

### Extension Documentation Generator (Builtin Extension)

Generate comprehensive markdown documentation for extensions and modules using the builtin `extension-doc-generator` extension.

**Endpoint:** `POST /v1/extensions/extension-doc-generator/execute`

**Request Body:**
```json
{
  "target": "prompt-template-executor",
  "output_path": "docs/extensions/prompt-template-executor.md",
  "format": "markdown",
  "include_examples": true,
  "include_api_details": true
}
```

**Parameters:**
- `target` (required, string) - Extension or module name to document
- `output_path` (optional, string) - Path to save generated markdown (default: `docs/extensions/{target}.md`)
- `format` (optional, string) - Output format: `markdown`, `html`, or `json` (default: `markdown`)
- `include_examples` (optional, boolean) - Include usage examples in documentation (default: `true`)
- `include_api_details` (optional, boolean) - Include API endpoint details (default: `true`)

**Response (Success):**
```json
{
  "success": true,
  "data": {
    "documentation": "# Prompt Template Executor\n\n...",
    "file_path": "docs/extensions/prompt-template-executor.md"
  }
}
```

**Example:**
```bash
curl -X POST \
  -H "X-API-Key: sk-llamagate" \
  -H "Content-Type: application/json" \
  -d '{
    "target": "prompt-template-executor",
    "output_path": "docs/extensions/prompt-template-executor.md"
  }' \
  http://localhost:11435/v1/extensions/extension-doc-generator/execute
```

**Status Codes:**
- `200 OK` - Documentation generated successfully
- `400 Bad Request` - Invalid input (e.g., missing `target`)
- `404 Not Found` - Target extension not found
- `500 Internal Server Error` - Generation failed

**Notes:**
- This is a builtin extension (YAML-based) located in `extensions/builtin/extension-doc-generator/`
- Always available - no installation needed
- Cannot be disabled (builtin extension protection)
- Generated documentation includes: overview, API endpoints, inputs/outputs, configuration, usage examples, workflow steps/hooks
- Documentation is saved to the specified path (or default location)

### Upsert Extension (Optional)

Create or update an extension manifest by writing to `~/.llamagate/extensions/installed/:name/`. **Enabled by default**; set `EXTENSIONS_UPSERT_ENABLED=false` to lock down. Used by clients (e.g. LlamaGate Control) to save workflows to LlamaGate. After upsert, call `POST /v1/extensions/refresh` to load the extension.

**Endpoint:** `PUT /v1/extensions/:name`

**Request Body:** YAML or JSON manifest (same schema as `manifest.yaml`). The path `:name` overrides `name` in the body.

**Response (Success, upsert enabled):**
```json
{
  "status": "ok",
  "name": "my-workflow",
  "path": "/path/to/.llamagate/extensions/installed/my-workflow/manifest.yaml"
}
```

**Response (Upsert disabled, EXTENSIONS_UPSERT_ENABLED=false):**
```json
{
  "error": "Workflow upsert is not enabled",
  "code": "UPSERT_NOT_CONFIGURED"
}
```

**Example (when enabled):**
```bash
curl -X PUT \
  -H "X-API-Key: sk-llamagate" \
  -H "Content-Type: application/yaml" \
  -d 'name: my-workflow
version: 1.0.0
description: My workflow
type: workflow
enabled: true
steps:
  - uses: llm.chat' \
  http://localhost:11435/v1/extensions/my-workflow
```

**Status Codes:**
- `200 OK` - Manifest written successfully
- `400 Bad Request` - Invalid name or manifest body
- `501 Not Implemented` - Upsert disabled (`EXTENSIONS_UPSERT_ENABLED=false`)
- `500 Internal Server Error` - Failed to write manifest

**Notes:**
- Extension name in the path must be alphanumeric, underscore, or hyphen only
- Writes to `~/.llamagate/extensions/installed/:name/manifest.yaml` (or configured installed dir)
- Call `POST /v1/extensions/refresh` after upsert to load the extension without restart
- Set `EXTENSIONS_UPSERT_ENABLED=false` to disable upsert (returns 501)

### Refresh Extensions

Re-discover extensions from the `extensions/` directory and update the registry. This is useful after installing new extensions or updating existing ones, as it allows extensions to be discovered without restarting the server.

**Endpoint:** `POST /v1/extensions/refresh`

**Response (Success):**
```json
{
  "status": "refreshed",
  "added": ["new-extension"],
  "updated": ["existing-extension"],
  "removed": ["deleted-extension"],
  "total": 5
}
```

**Response (Partial Failure):**
```json
{
  "status": "partial",
  "added": ["new-extension"],
  "updated": [],
  "removed": [],
  "total": 4,
  "errors": [
    "invalid-extension: manifest validation failed"
  ]
}
```

**Example:**
```bash
curl -X POST \
  -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/extensions/refresh
```

**Status Codes:**
- `200 OK` - All extensions refreshed successfully
- `206 Partial Content` - Some extensions failed to load (check `errors` array)
- `500 Internal Server Error` - Discovery failed

**Notes:**
- Re-scans the `extensions/` directory for `manifest.yaml` files
- Adds newly discovered extensions
- Updates existing extensions if their manifests changed
- Removes extensions that no longer exist in the directory
- Invalid extensions are logged but don't prevent other extensions from loading
- This endpoint is typically called by install tools after installing extensions

**Use Case:**
After installing a new extension (e.g., via `llamagate-cli`), discovery happens automatically:

```bash
# Install extension using the new CLI tool
llamagate import extension my-extension.zip
# Discovery is automatically triggered - no manual refresh needed!

# The refresh endpoint is still available for manual refresh if needed:
curl -X POST \
  -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/extensions/refresh
```

**Note:** The new `llamagate-cli` tool replaces the previous `agentic-module-tool`. Use `llamagate import` for installing extensions and modules.

## Health Endpoint

The main health endpoint (`/health`) does not require authentication and can be used for monitoring:

```bash
curl http://localhost:11435/health
```

