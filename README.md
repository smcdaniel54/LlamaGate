# LlamaGate

[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)](README.md)

LlamaGate is a production-ready, OpenAI-compatible HTTP proxy/gateway for local Ollama instances. It's a lightweight, single-binary tool that forwards requests to Ollama with added features like caching, authentication, rate limiting, and structured logging.

> 🚀 **New to LlamaGate?**  
> **[Quick Start Guide](QUICKSTART.md)** - Get running in 2 minutes

## Features

- ✅ **OpenAI-Compatible API**: Drop-in replacement for OpenAI API endpoints
- ✅ **Caching**: In-memory caching for identical prompts to reduce Ollama load
- ✅ **Authentication**: Optional API key authentication via headers
- ✅ **Rate Limiting**: Configurable rate limiting using leaky bucket algorithm
- ✅ **Structured Logging**: JSON logging with request IDs using Zerolog
- ✅ **Streaming Support**: Full support for streaming chat completions
- ✅ **Graceful Shutdown**: Clean shutdown on SIGINT/SIGTERM
- ✅ **Single Binary**: Lightweight, easy to deploy
- ✅ **Docker Support**: Multi-stage Dockerfile for minimal image size

## Installation

### Automated Installer (Recommended)

The easiest way to install LlamaGate:

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
- Optionally create shortcuts/services

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
docker run -p 8080:8080 llamagate
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
| `RATE_LIMIT_RPS` | `10` | Requests per second limit |
| `DEBUG` | `false` | Enable debug logging |
| `PORT` | `8080` | Server port |
| `LOG_FILE` | (empty) | Path to log file (optional, logs to console if empty) |
| `TIMEOUT` | `5m` | HTTP client timeout for Ollama requests (e.g., `5m`, `30s`, `30m` - max 30 minutes) |

### Using .env File (Recommended)

Create a `.env` file in the project root (copy from `.env.example`):

```bash
# .env
OLLAMA_HOST=http://localhost:11434
API_KEY=sk-llamagate
RATE_LIMIT_RPS=10
DEBUG=false
PORT=8080
LOG_FILE=llamagate.log
TIMEOUT=5m
```

The `.env` file is automatically loaded when the application starts. Environment variables set directly will override `.env` file values, making it easy to override settings for specific runs.

### Example (Linux/Mac)

```bash
export OLLAMA_HOST="http://localhost:11434"
export API_KEY="sk-llamagate"
export RATE_LIMIT_RPS=20
export DEBUG=true
export PORT=8080

llamagate
```

### Example (Windows)

```cmd
set OLLAMA_HOST=http://localhost:11434
set API_KEY=sk-llamagate
set RATE_LIMIT_RPS=20
set DEBUG=true
set PORT=8080

llamagate.exe
```

Or use the provided batch files (see Windows Quick Start above).

**Note:** If you use a `.env` file, you don't need to set environment variables manually - just create `.env` and run the application!

## Usage

> 💡 **Migrating from OpenAI?** See the [Quick Start Guide](QUICKSTART.md) for step-by-step migration examples.

### Health Check

```bash
curl http://localhost:8080/health
```

### List Models

```bash
curl http://localhost:8080/v1/models \
  -H "X-API-Key: sk-llamagate"
```

### Chat Completions (Non-Streaming)

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "model": "llama2",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ]
  }'
```

### Chat Completions (Streaming)

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "model": "llama2",
    "messages": [
      {"role": "user", "content": "Tell me a story"}
    ],
    "stream": true
  }'
```

### Using OpenAI Python SDK

LlamaGate is compatible with the OpenAI Python SDK:

```python
from openai import OpenAI

# Point to LlamaGate instead of OpenAI
client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="sk-llamagate"  # Your API_KEY from env
)

# Use it like OpenAI
response = client.chat.completions.create(
    model="llama2",
    messages=[
        {"role": "user", "content": "Hello!"}
    ]
)

print(response.choices[0].message.content)
```

### Using with LangChain

```python
from langchain.llms import Ollama
from langchain.chat_models import ChatOpenAI

# Use ChatOpenAI with LlamaGate endpoint
llm = ChatOpenAI(
    model="llama2",
    openai_api_base="http://localhost:8080/v1",
    openai_api_key="sk-llamagate"
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

If `API_KEY` is set, all requests must include the API key in one of these ways:

1. **X-API-Key header:**

   ```bash
   curl -H "X-API-Key: sk-llamagate" ...
   ```

2. **Authorization Bearer header:**

   ```bash
   curl -H "Authorization: Bearer sk-llamagate" ...
   ```

If `API_KEY` is not set, authentication is disabled and all requests are allowed.

## Caching

LlamaGate caches responses for non-streaming requests. The cache key is based on:

- Model name
- Messages content

Identical requests (same model + same messages) will return cached responses, reducing load on Ollama.

## Rate Limiting

Rate limiting is implemented using a leaky bucket algorithm. The default limit is 10 requests per second, configurable via `RATE_LIMIT_RPS`.

When the limit is exceeded, requests receive a `429 Too Many Requests` response.

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
.\test-all-installers.ps1

# Test Windows installer only
.\test-installer-windows.ps1

# Test Unix installer (requires bash/WSL)
chmod +x test-installer-unix.sh
./test-installer-unix.sh
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
│   ├── middleware/
│   │   ├── auth.go          # Authentication middleware
│   │   ├── rate_limit.go    # Rate limiting middleware
│   │   └── request_id.go    # Request ID middleware
│   └── proxy/
│       └── proxy.go          # Proxy handlers
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
