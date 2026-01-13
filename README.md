# LlamaGate

[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)](README.md)

LlamaGate is a production-ready, OpenAI-compatible HTTP proxy/gateway for local Ollama instances. It's a lightweight, single-binary tool that forwards requests to Ollama with added features like caching, authentication, rate limiting, and structured logging.

> 🚀 **New to LlamaGate?**  
> **[Quick Start Guide](QUICKSTART.md)** - Get running in 2 minutes

## Features

- ✅ **OpenAI-Compatible API**: Drop-in replacement for OpenAI API endpoints
- ✅ **MCP Client Support**: Connect to MCP servers and expose their tools to models ([MCP Guide](docs/MCP.md) | [Quick Start](docs/MCP_QUICKSTART.md))
- ✅ **Plugin System**: Create reusable plugins and agentic workflows ([Plugin Guide](docs/PLUGINS.md) | [Quick Start](docs/PLUGIN_QUICKSTART.md))
- ✅ **Caching**: In-memory caching for identical prompts to reduce Ollama load
- ✅ **Authentication**: Optional API key authentication via headers
- ✅ **Rate Limiting**: Configurable rate limiting using leaky bucket algorithm
- ✅ **Structured Logging**: JSON logging with request IDs using Zerolog
- ✅ **Streaming Support**: Full support for streaming chat completions
- ✅ **Tool/Function Calling**: Execute MCP tools in multi-round loops
- ✅ **Graceful Shutdown**: Clean shutdown on SIGINT/SIGTERM
- ✅ **Single Binary**: Lightweight, easy to deploy
- ✅ **Docker Support**: Multi-stage Dockerfile for minimal image size

## Documentation

- 📖 **[Quick Start Guide](QUICKSTART.md)** - Get running in 2 minutes
- 📚 **[Full Documentation Index](docs/README.md)** - Browse all documentation
- 🔧 **[MCP Integration](docs/MCP.md)** - Model Context Protocol guide
- 🚀 **[MCP Quick Start](docs/MCP_QUICKSTART.md)** - Get started with MCP in 5 minutes
- 🎯 **[MCP Demo Guide](docs/MCP_DEMO_QUICKSTART.md)** - Full demo with multiple servers
- 🌐 **[MCP HTTP API](docs/API.md)** - Complete API reference for MCP management
- 🔌 **[Plugin System](docs/PLUGINS.md)** - Create reusable plugins and workflows
- ⚡ **[Plugin Quick Start](docs/PLUGIN_QUICKSTART.md)** - Get started with plugins in 5 minutes
- 🧪 **[Testing Guide](docs/TESTING.md)** - Testing your setup
- 📦 **[Installation Guide](docs/INSTALL.md)** - Detailed installation instructions

## Installation

### ⚡ One-Click Install (Easiest - Recommended!)

**Automatically downloads the correct binary for your platform - no Go required!**

**Windows:**
```cmd
install\windows\install-binary.cmd
```

**Unix/Linux/macOS:**
```bash
chmod +x install/unix/install-binary.sh
./install/unix/install-binary.sh
```

This will:
- ✅ Automatically detect your platform
- ✅ Download the correct binary
- ✅ Set up the executable
- ✅ Create a default `.env` configuration file

**That's it!** You're ready to run LlamaGate.

### 🚀 Manual Download (Also Easy - No Go Required!)

**Prefer to download manually?** Get pre-built binaries from [Releases](https://github.com/llamagate/llamagate/releases/latest):

**Windows:**
1. Download [llamagate-windows-amd64.exe](https://github.com/llamagate/llamagate/releases/latest/download/llamagate-windows-amd64.exe)
2. Run: `.\llamagate-windows-amd64.exe`

**Linux:**
```bash
curl -LO https://github.com/llamagate/llamagate/releases/latest/download/llamagate-linux-amd64
chmod +x llamagate-linux-amd64
./llamagate-linux-amd64
```

**macOS:**
```bash
# Apple Silicon (M1/M2/M3)
curl -LO https://github.com/llamagate/llamagate/releases/latest/download/llamagate-darwin-arm64
chmod +x llamagate-darwin-arm64
./llamagate-darwin-arm64

# Intel Mac
curl -LO https://github.com/llamagate/llamagate/releases/latest/download/llamagate-darwin-amd64
chmod +x llamagate-darwin-amd64
./llamagate-darwin-amd64
```

**All platforms:** See [Releases](https://github.com/llamagate/llamagate/releases) for all available binaries (amd64, arm64, etc.).

### Build from Source (For Developers)

If you prefer to build from source or need custom builds:

**Windows:**

```cmd
install\windows\install.cmd
```

**Unix/Linux/macOS:**

```bash
chmod +x install/unix/install.sh
./install/unix/install.sh
```

**The installer will:**

- Check for Go and install it if needed
- Check for Ollama and guide you to install it
- Install all Go dependencies
- Build the LlamaGate binary
- Create a `.env` configuration file

**Follow the prompts** to configure your installation

### From Source

```bash
go install github.com/llamagate/llamagate/cmd/llamagate@latest
```

### Windows Quick Start

For Windows users, convenient batch files are provided:

- **`scripts/windows/run.cmd`** - Run with default settings (no authentication)
- **`scripts/windows/run-with-auth.cmd`** - Run with API key authentication enabled
- **`scripts/windows/run-debug.cmd`** - Run with debug logging enabled
- **`scripts/windows/build.cmd`** - Build the binary (`llamagate.exe`)

Run from command prompt:

```cmd
scripts\windows\run.cmd
```

To customize settings, edit the batch file or set environment variables before running:

```cmd
set OLLAMA_HOST=http://localhost:11434
set API_KEY=sk-llamagate
set RATE_LIMIT_RPS=20
scripts\windows\run.cmd
```

### From Docker

```bash
docker build -t llamagate .
docker run -p 11435:11435 llamagate
```

## Configuration

LlamaGate can be configured via:

1. **`.env` file** (recommended for development) - Create a `.env` file in the project root
2. **Environment variables** - Takes precedence over `.env` file values
3. **Default values** - Used if neither `.env` nor environment variables are set

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `OLLAMA_HOST` | `http://localhost:11434` | Ollama server URL |
| `API_KEY` | (empty) | API key for authentication (optional) |
| `RATE_LIMIT_RPS` | `50` | Requests per second limit |
| `DEBUG` | `false` | Enable debug logging |
| `PORT` | `11435` | Server port |
| `LOG_FILE` | (empty) | Path to log file (optional, logs to console if empty) |
| `TLS_ENABLED` | `false` | Enable HTTPS/TLS |
| `TLS_CERT_FILE` | (empty) | Path to TLS certificate file (required if TLS_ENABLED=true) |
| `TLS_KEY_FILE` | (empty) | Path to TLS private key file (required if TLS_ENABLED=true) |
| `TIMEOUT` | `5m` | HTTP client timeout for Ollama requests (e.g., `5m`, `30s`, `30m` - max 30 minutes) |
| `MCP_ENABLED` | `false` | Enable MCP client functionality (see [MCP docs](docs/MCP.md)) |
| `MCP_MAX_TOOL_ROUNDS` | `10` | Maximum tool execution rounds |
| `MCP_MAX_TOOL_CALLS_PER_ROUND` | `10` | Maximum tool calls per round |
| `MCP_DEFAULT_TOOL_TIMEOUT` | `30s` | Default timeout for tool execution |
| `MCP_MAX_TOOL_RESULT_SIZE` | `1048576` | Maximum tool result size in bytes (1MB) |
| `MCP_ALLOW_TOOLS` | (empty) | Comma-separated glob patterns for allowed tools |
| `MCP_DENY_TOOLS` | (empty) | Comma-separated glob patterns for denied tools |

**Note:** MCP server configuration is best done via YAML/JSON config file. See [mcp-config.example.yaml](mcp-config.example.yaml) and [MCP Documentation](docs/MCP.md).

### Using .env File (Recommended)

Create a `.env` file in the project root (copy from `.env.example`):

```bash
# .env
OLLAMA_HOST=http://localhost:11434
API_KEY=sk-llamagate
RATE_LIMIT_RPS=50
DEBUG=false
PORT=11435
LOG_FILE=llamagate.log
TIMEOUT=5m
```

The `.env` file is automatically loaded when the application starts. Environment variables set directly will override `.env` file values, making it easy to override settings for specific runs.

## Authentication

When `API_KEY` is configured, all API endpoints (except `/health`) require authentication.

LlamaGate supports two authentication header formats:

### X-API-Key Header (Recommended)
```bash
curl -H "X-API-Key: sk-llamagate" http://localhost:11435/v1/models
```

### Authorization Bearer Header (Alternative)
```bash
curl -H "Authorization: Bearer sk-llamagate" http://localhost:11435/v1/models
```

The `X-API-Key` header is checked first. If not present, `Authorization: Bearer` is checked.

**Note:** The `/health` endpoint does not require authentication and can be used for monitoring and load balancer health checks.

### Example (Linux/Mac)

```bash
export OLLAMA_HOST="http://localhost:11434"
export API_KEY="sk-llamagate"
export RATE_LIMIT_RPS=20
export DEBUG=true
export PORT=11435

llamagate
```

### Example (Windows)

```cmd
set OLLAMA_HOST=http://localhost:11434
set API_KEY=sk-llamagate
set RATE_LIMIT_RPS=20
set DEBUG=true
set PORT=11435

llamagate.exe
```

Or use the provided batch files (see Windows Quick Start above).

**Note:** If you use a `.env` file, you don't need to set environment variables manually - just create `.env` and run the application!

### Supported Authentication Headers

LlamaGate supports two authentication header formats (both are case-insensitive):

#### 1. X-API-Key Header (Recommended)
```bash
curl -H "X-API-Key: sk-llamagate" http://localhost:11435/v1/models
```

The header name is case-insensitive. All of the following are accepted:
- `X-API-Key`
- `x-api-key`
- `X-Api-Key`
- Any other case variation

#### 2. Authorization Bearer Header (Alternative)
```bash
curl -H "Authorization: Bearer sk-llamagate" http://localhost:11435/v1/models
```

The "Bearer" scheme is case-insensitive. All of the following are accepted:
- `Authorization: Bearer sk-llamagate`
- `Authorization: bearer sk-llamagate`
- `Authorization: BEARER sk-llamagate`

**Header Priority:** The `X-API-Key` header is checked first. If not present, `Authorization: Bearer` is checked.

### Authentication Behavior

- **When Authentication is Required:** All endpoints except `/health` require authentication when `API_KEY` is configured.
- **When Authentication is Missing:** Requests without a valid authentication header return `401 Unauthorized` with an OpenAI-compatible error response.
- **When Authentication is Invalid:** Requests with an invalid or incorrect API key return `401 Unauthorized` with an OpenAI-compatible error response.

### Error Response Format

Authentication errors return HTTP `401 Unauthorized` with a JSON response in OpenAI-compatible format:

```json
{
  "error": {
    "message": "Invalid API key",
    "type": "invalid_request_error",
    "request_id": "req-123456"
  }
}
```

### Security

- **API keys are never logged:** Authentication failures are logged with a generic message ("Authentication failed") but the provided API key or bearer token is never included in logs.
- **Constant-time comparison:** API key validation uses constant-time comparison to prevent timing attacks.
- **Health endpoint bypass:** The `/health` endpoint does not require authentication and can be used for monitoring and load balancer health checks.

## Usage

> 💡 **Migrating from OpenAI?** See the [Quick Start Guide](QUICKSTART.md) for step-by-step migration examples.

> 🔧 **Using MCP Tools?** See the [MCP Quick Start Guide](docs/MCP_QUICKSTART.md) to get started with MCP integration. For complete details, see the [MCP Documentation](docs/MCP.md).

> 🎯 **Want to see MCP in action?** Check out the [MCP Demo QuickStart](docs/MCP_DEMO_QUICKSTART.md) for a complete example with multiple document processing servers.

### Usage Examples

All examples below assume:
- **LlamaGate** running locally on `http://localhost:11435` (default port)
- **Ollama** running locally on `http://localhost:11434` (default port)
- **Default configuration** (no authentication unless specified)

#### 1. Non-Streaming Request (curl)

```bash
curl http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama2",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ]
  }'
```

**Response:**
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
      "content": "Hello! I'm doing well, thank you for asking..."
    },
    "finish_reason": "stop"
  }]
}
```

#### 2. Streaming Request (curl)

```bash
curl http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama2",
    "messages": [
      {"role": "user", "content": "Tell me a short story"}
    ],
    "stream": true
  }'
```

**Response (Server-Sent Events):**
```
data: {"id":"chatcmpl-...","object":"chat.completion.chunk","created":1234567890,"model":"llama2","choices":[{"index":0,"delta":{"content":"Once"},"finish_reason":null}]}

data: {"id":"chatcmpl-...","object":"chat.completion.chunk","created":1234567890,"model":"llama2","choices":[{"index":0,"delta":{"content":" upon"},"finish_reason":null}]}

data: [DONE]
```

#### 3. Using OpenAI Python SDK

Point the OpenAI Python SDK to LlamaGate using a custom `base_url`:

```python
from openai import OpenAI

# Configure client to use LlamaGate instead of OpenAI
client = OpenAI(
    base_url="http://localhost:11435/v1",  # LlamaGate endpoint
    api_key="not-needed"  # Optional: only needed if API_KEY is set in LlamaGate
)

# Use it exactly like the OpenAI API
response = client.chat.completions.create(
    model="llama2",  # Use any model available in your local Ollama
    messages=[
        {"role": "user", "content": "Hello! How are you?"}
    ]
)

print(response.choices[0].message.content)
```

**Streaming with Python SDK:**
```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:11435/v1",
    api_key="not-needed"
)

stream = client.chat.completions.create(
    model="llama2",
    messages=[
        {"role": "user", "content": "Count to 5"}
    ],
    stream=True
)

for chunk in stream:
    if chunk.choices[0].delta.content is not None:
        print(chunk.choices[0].delta.content, end="", flush=True)
```

#### 4. Using OpenAI Node.js SDK

Point the OpenAI Node.js SDK to LlamaGate using a custom `baseURL`:

```javascript
import OpenAI from 'openai';

// Configure client to use LlamaGate instead of OpenAI
const client = new OpenAI({
  baseURL: 'http://localhost:11435/v1',  // LlamaGate endpoint
  apiKey: 'not-needed'  // Optional: only needed if API_KEY is set in LlamaGate
});

// Use it exactly like the OpenAI API
const response = await client.chat.completions.create({
  model: 'llama2',  // Use any model available in your local Ollama
  messages: [
    { role: 'user', content: 'Hello! How are you?' }
  ]
});

console.log(response.choices[0].message.content);
```

**Streaming with Node.js SDK:**
```javascript
import OpenAI from 'openai';

const client = new OpenAI({
  baseURL: 'http://localhost:11435/v1',
  apiKey: 'not-needed'
});

const stream = await client.chat.completions.create({
  model: 'llama2',
  messages: [
    { role: 'user', content: 'Count to 5' }
  ],
  stream: true
});

for await (const chunk of stream) {
  if (chunk.choices[0]?.delta?.content) {
    process.stdout.write(chunk.choices[0].delta.content);
  }
}
```

#### 5. Authentication Example (if enabled)

If you've set `API_KEY` in your LlamaGate configuration, include it in requests:

**curl with authentication:**
```bash
curl http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "model": "llama2",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ]
  }'
```

**Python SDK with authentication:**
```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:11435/v1",
    api_key="sk-llamagate"  # Your API_KEY from LlamaGate config
)

response = client.chat.completions.create(
    model="llama2",
    messages=[
        {"role": "user", "content": "Hello!"}
    ]
)

print(response.choices[0].message.content)
```

**Node.js SDK with authentication:**
```javascript
import OpenAI from 'openai';

const client = new OpenAI({
  baseURL: 'http://localhost:11435/v1',
  apiKey: 'sk-llamagate'  // Your API_KEY from LlamaGate config
});

const response = await client.chat.completions.create({
  model: 'llama2',
  messages: [
    { role: 'user', content: 'Hello!' }
  ]
});

console.log(response.choices[0].message.content);
```

**Note:** Authentication is optional. If `API_KEY` is not set in LlamaGate, you can omit the `api_key` parameter or use any value.

### Health Check

```bash
curl http://localhost:11435/health
```

### List Models

```bash
curl http://localhost:11435/v1/models
```

### Using with LangChain

```python
from langchain.llms import Ollama
from langchain.chat_models import ChatOpenAI

# Use ChatOpenAI with LlamaGate endpoint
llm = ChatOpenAI(
    model="llama2",
    openai_api_base="http://localhost:11435/v1",
    openai_api_key="not-needed"  # Optional: only if API_KEY is set
)

response = llm.invoke("Hello, how are you?")
print(response.content)
```

## API Endpoints

### `POST /v1/chat/completions`

OpenAI-compatible chat completions endpoint. Forwards requests to Ollama `/api/chat`.

**Request Body:**

```json
{
  "model": "llama2",
  "messages": [
    {"role": "user", "content": "Hello!"}
  ],
  "stream": false,
  "temperature": 0.7
}
```

### `GET /v1/models`

Lists available Ollama models. Forwards requests to Ollama `/api/tags` and converts to OpenAI format.

### `GET /health`

Health check endpoint that verifies both server and Ollama connectivity.

**Response (healthy):**

```json
{
  "status": "healthy",
  "ollama": "connected",
  "ollama_host": "http://localhost:11434"
}
```

**Response (unhealthy):**

```json
{
  "status": "unhealthy",
  "error": "Ollama unreachable",
  "ollama_host": "http://localhost:11434"
}
```

Returns `200 OK` when healthy, `503 Service Unavailable` when Ollama is unreachable.

## Authentication

When `API_KEY` is configured, all API endpoints (except `/health`) require authentication.

### Supported Authentication Headers

LlamaGate supports two authentication header formats (both are case-insensitive):

#### 1. X-API-Key Header (Recommended)
```bash
curl -H "X-API-Key: sk-llamagate" http://localhost:11435/v1/models
```

The header name is case-insensitive. All of the following are accepted:
- `X-API-Key`
- `x-api-key`
- `X-Api-Key`
- Any other case variation

#### 2. Authorization Bearer Header (Alternative)
```bash
curl -H "Authorization: Bearer sk-llamagate" http://localhost:11435/v1/models
```

The "Bearer" scheme is case-insensitive. All of the following are accepted:
- `Authorization: Bearer sk-llamagate`
- `Authorization: bearer sk-llamagate`
- `Authorization: BEARER sk-llamagate`

**Header Priority:** The `X-API-Key` header is checked first. If not present, `Authorization: Bearer` is checked.

### Authentication Behavior

- **When Authentication is Required:** All endpoints except `/health` require authentication when `API_KEY` is configured.
- **When Authentication is Missing:** Requests without a valid authentication header return `401 Unauthorized` with an OpenAI-compatible error response.
- **When Authentication is Invalid:** Requests with an invalid or incorrect API key return `401 Unauthorized` with an OpenAI-compatible error response.

### Error Response Format

Authentication errors return HTTP `401 Unauthorized` with a JSON response in OpenAI-compatible format:

```json
{
  "error": {
    "message": "Invalid API key",
    "type": "invalid_request_error",
    "request_id": "req-123456"
  }
}
```

### Security

- **API keys are never logged:** Authentication failures are logged with a generic message ("Authentication failed") but the provided API key or bearer token is never included in logs.
- **Constant-time comparison:** API key validation uses constant-time comparison to prevent timing attacks.
- **Health endpoint bypass:** The `/health` endpoint does not require authentication and can be used for monitoring and load balancer health checks.

If `API_KEY` is not set, authentication is disabled and all requests are allowed.

## Caching

LlamaGate caches responses for non-streaming requests. The cache key is based on:

- Model name
- Messages content

Identical requests (same model + same messages) will return cached responses, reducing load on Ollama.

## Rate Limiting

Rate limiting is implemented using a leaky bucket algorithm. The default limit is 50 requests per second, configurable via `RATE_LIMIT_RPS`.

When the limit is exceeded, requests receive a `429 Too Many Requests` response with:

- **HTTP Status**: `429 Too Many Requests`
- **Retry-After Header**: Number of seconds to wait before retrying (e.g., `Retry-After: 1`)
- **Response Body**: OpenAI-compatible JSON error format

### Rate Limit Response Format

**Status:** `429 Too Many Requests`

**Headers:**
```
Retry-After: 1
```

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

Rate-limited requests are logged with structured fields including request ID, IP address, path, retry time, and limiter decision.

## Request ID and Logging

LlamaGate implements consistent request correlation and secure logging across all components.

### Request ID Generation

Every inbound request receives a unique request ID:

- **If `X-Request-ID` header is provided**: LlamaGate uses the provided request ID
- **If no header is provided**: LlamaGate generates a UUID v4 request ID

The request ID is:
- Included in the `X-Request-ID` response header
- Propagated to all downstream components:
  - Ollama upstream calls (via `X-Request-ID` header)
  - Tool/function calling (via context)
  - MCP/plugin calls (via context and HTTP headers)
- Included in all structured log entries for the request

### Sensitive Data Redaction

LlamaGate automatically redacts sensitive values from logs to prevent secret leakage:

**Redacted Values:**
- API keys (`X-API-Key` header values)
- Bearer tokens (`Authorization: Bearer` header values)
- Any other secrets in headers, environment variables, or configuration

**What is Logged:**
- Request method, path, status code, latency
- Request ID for correlation
- Client IP address
- Error messages (without sensitive data)
- Authentication failures (generic message only)

**What is NOT Logged:**
- API key values
- Bearer token values
- Authorization header contents
- Any header values that contain secrets

**Example Log Entry:**
```json
{
  "level": "info",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "POST",
  "path": "/v1/chat/completions",
  "status": 200,
  "latency": "1.234s",
  "ip": "192.168.1.100",
  "time": "2026-01-12T10:00:00Z",
  "message": "HTTP request"
}
```

Notice that the API key is not present in the log, even though it was sent in the request headers.

## Graceful Shutdown

LlamaGate implements graceful shutdown to ensure clean termination without dropping in-flight requests.

### Shutdown Behavior

When LlamaGate receives `SIGINT` or `SIGTERM`:

1. **Stop accepting new requests**: The server immediately stops accepting new connections
2. **Allow in-flight requests to complete**: Active requests are allowed to finish up to a configurable timeout
3. **Close downstream connections cleanly**:
   - Ollama HTTP client connections are closed
   - MCP server connections are closed
   - Cache cleanup goroutines are stopped
4. **Handle streaming responses safely**: Streaming responses check for context cancellation and stop gracefully when the server shuts down

### Configuration

The shutdown timeout is configurable via the `SHUTDOWN_TIMEOUT` environment variable:

```bash
# Default: 30 seconds
SHUTDOWN_TIMEOUT=30s

# Examples:
SHUTDOWN_TIMEOUT=10s   # 10 seconds
SHUTDOWN_TIMEOUT=1m     # 1 minute
SHUTDOWN_TIMEOUT=2m30s  # 2 minutes 30 seconds
```

**Timeout Behavior:**
- If all in-flight requests complete before the timeout: Clean shutdown
- If the timeout is reached: Remaining requests are terminated, and the server exits

### Shutdown Process

1. Signal received (`SIGINT` or `SIGTERM`)
2. Server stops accepting new requests
3. Cache cleanup goroutines stopped
4. Downstream connections closed (Ollama, MCP)
5. In-flight requests allowed to complete (up to timeout)
6. Server exits gracefully

**Note:** Streaming responses automatically detect server shutdown via context cancellation and stop gracefully, preventing abrupt connection resets.

## HTTPS/SSL Support

LlamaGate supports native HTTPS/TLS encryption. To enable HTTPS:

1. **Set TLS configuration in `.env`**:
   ```bash
   TLS_ENABLED=true
   TLS_CERT_FILE=/path/to/certificate.crt
   TLS_KEY_FILE=/path/to/private.key
   ```

2. **Or use YAML config**:
   ```yaml
   tls_enabled: true
   tls_cert_file: /path/to/certificate.crt
   tls_key_file: /path/to/private.key
   ```

3. **Generate self-signed certificate (for testing)**:
   ```bash
   openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
   ```

4. **For production with Let's Encrypt**, use a reverse proxy (nginx, Caddy) for automatic certificate management and renewal.

**Note:** When `TLS_ENABLED=true`, the server will use HTTPS. Make sure to use `https://` in your client URLs.

## Logging

LlamaGate uses structured JSON logging with Zerolog. Each request is assigned a unique request ID.

**Log Levels:**

- `INFO`: Default level, logs all requests and important events
- `DEBUG`: Enabled with `DEBUG=true`, includes detailed debugging information

**Log Output:**

- By default, logs are written to stdout (console)
- To also write logs to a file, set the `LOG_FILE` environment variable:

  ```bash
  LOG_FILE=llamagate.log
  ```

- When `LOG_FILE` is set, logs are written to both console and file
- The log file is created automatically if it doesn't exist, and new logs are appended to it
- **Note:** The log file is not automatically rotated. For production use, consider using a log rotation tool or process manager

**Example log entry:**

```json
{
  "level": "info",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "POST",
  "path": "/v1/chat/completions",
  "status": 200,
  "latency": "1.234s",
  "ip": "127.0.0.1",
  "time": 1703123456
}
```

## Testing

### Application Testing

See [docs/TESTING.md](docs/TESTING.md) for a comprehensive testing guide, or use the provided test script:

**Windows:**

```cmd
scripts\windows\test.cmd
```

**Unix/Linux/macOS:**

```bash
./scripts/unix/test.sh
```

This will test all endpoints, caching, authentication, and more.

### Installer Testing

To validate installer scripts before deployment, see [docs/INSTALLER_TESTING.md](docs/INSTALLER_TESTING.md):

```powershell
# Test all installers
.\tests\installer\test-all-installers.ps1

# Test Windows installer only
.\tests\installer\test-installer-windows.ps1

# Test Unix installer (requires bash/WSL)
chmod +x tests/installer/test-installer-unix.sh
./tests/installer/test-installer-unix.sh
```

## Development

### Building

**Using the installer (recommended):**

```cmd
install\windows\install.cmd
```

**Manual build:**

```bash
go build -o llamagate ./cmd/llamagate
```

Or use the build script:

```cmd
scripts\windows\build.cmd
```

### Running Tests

```bash
go test ./...
```

### Running Locally

```bash
# Make sure Ollama is running
ollama serve

# In another terminal, run LlamaGate
go run ./cmd/llamagate
```

## MCP Client Support

LlamaGate includes support for the Model Context Protocol (MCP) as a client. This allows you to:

- Connect to MCP servers and discover their tools, resources, and prompts
- Expose tools to chat completion requests  
- Execute tool calls in multi-round loops
- Reference MCP resources directly in chat messages using `mcp://` URIs
- Enforce security with allow/deny lists
- Access MCP management via HTTP API

See [MCP Documentation](docs/MCP.md) for full details and [MCP Quick Start](docs/MCP_QUICKSTART.md) for a getting started guide.

### MCP URI Scheme

You can reference MCP resources directly in chat completion messages:

```bash
curl -X POST http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "model": "llama3.2",
    "messages": [{
      "role": "user",
      "content": "Summarize mcp://filesystem/file:///docs/readme.txt"
    }]
  }'
```

LlamaGate will automatically fetch the resource content and inject it as context.

## Project Scope & Paid Tier Boundary

**LlamaGate Core is Open Source**

This repository contains the core LlamaGate functionality:
- OpenAI-compatible API gateway
- MCP client support
- Caching, authentication, rate limiting
- Basic tool execution

**Advanced Features (Separate Modules)**

The following features are **not included** in this open-source core and are available as separate modules:
- Advanced workflow automation packs
- Enterprise connectors and integrations
- Cloud fallback capabilities
- Compatibility validation suites
- Premium support and consulting

These advanced features are maintained separately and are not part of this repository.

## Known Limitations

### Supported Platforms
- ✅ Windows (amd64)
- ✅ Linux (amd64, arm64)
- ✅ macOS (amd64, arm64)

### Model Backends
- ✅ **Ollama** - Fully supported (primary backend)
- ❌ Direct OpenAI API - Not included (use OpenAI SDK directly)
- ❌ Other LLM providers - Not included in core

### MCP Implementation Status
- ✅ **stdio transport** - Fully implemented
- ⚠️ **SSE transport** - Interface prepared, implementation pending
- ✅ **Tool execution** - Multi-round loops supported
- ✅ **Security guardrails** - Allow/deny lists, timeouts, size limits

### Other Limitations
- **HTTPS/TLS** - Native HTTPS support available via `TLS_ENABLED`, `TLS_CERT_FILE`, and `TLS_KEY_FILE` configuration. For production with Let's Encrypt, a reverse proxy (nginx, Caddy) is still recommended for automatic certificate management.
- **In-memory cache only** - Cache is lost on restart (persistent cache not included in core)
- **Global rate limiting** - Per-IP rate limiting not included in core
- **No cloud fallback** - Core is designed for local Ollama instances only
- **Single binary deployment** - No built-in clustering or load balancing
- **Single instance per machine** - Only one LlamaGate instance should run per machine. Multiple applications can connect to the same instance. If you try to start a second instance, you'll get a clear error message indicating the port is already in use.

### What's Not Included
- Database persistence (cache, logs, etc.)
- Multi-tenant isolation
- Advanced monitoring/observability dashboards
- Enterprise SSO/authentication providers
- High-availability/clustering features

## Project Structure

```text
.
├── cmd/
│   └── llamagate/
│       └── main.go          # Entry point
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── logger/
│   │   └── logger.go        # Logger initialization
│   ├── cache/
│   │   └── cache.go         # In-memory cache
│   ├── mcpclient/
│   │   ├── client.go        # MCP client implementation
│   │   ├── stdio.go         # stdio transport
│   │   ├── sse.go           # SSE transport (stub)
│   │   ├── types.go         # MCP protocol types
│   │   └── errors.go        # MCP errors
│   ├── tools/
│   │   ├── manager.go       # Tool registry and management
│   │   ├── mapper.go        # MCP to OpenAI format conversion
│   │   ├── guardrails.go    # Security and limits
│   │   └── types.go         # Tool types
│   ├── middleware/
│   │   ├── auth.go          # Authentication middleware
│   │   ├── rate_limit.go    # Rate limiting middleware
│   │   └── request_id.go    # Request ID middleware
│   └── proxy/
│       ├── proxy.go          # Proxy handlers
│       └── tool_loop.go     # Tool execution loop
├── docs/
│   ├── MCP.md               # MCP documentation
│   └── MCP_QUICKSTART.md    # MCP quick start guide
├── mcp-config.example.yaml  # MCP configuration example
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
