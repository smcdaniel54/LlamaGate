# LlamaGate Architecture

This document describes the high-level architecture of LlamaGate, including component interactions, data flow, and design decisions.

## Table of Contents

- [Overview](#overview)
- [System Architecture](#system-architecture)
- [Component Overview](#component-overview)
- [Request Flow](#request-flow)
- [Data Flow](#data-flow)
- [Component Interactions](#component-interactions)
- [Design Patterns](#design-patterns)
- [Concurrency Model](#concurrency-model)
- [Error Handling](#error-handling)
- [Configuration Management](#configuration-management)

## Overview

LlamaGate is a **single-binary HTTP proxy/gateway** that sits between clients and Ollama, providing:

- OpenAI-compatible API endpoints
- Request caching
- Authentication and rate limiting
- MCP (Model Context Protocol) client integration
- Extension system for extensibility
- Structured logging

**Architecture Principles:**
- **Single Binary**: Everything compiled into one executable
- **Stateless**: No persistent state (except in-memory cache)
- **Layered**: Clear separation of concerns
- **Extensible**: Extension system for custom functionality

## System Architecture

### High-Level Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                        Client Application                    │
│              (Python, JavaScript, cURL, etc.)                │
└───────────────────────┬─────────────────────────────────────┘
                        │ HTTP Requests
                        │ (OpenAI-compatible)
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                      LlamaGate Server                        │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              HTTP Layer (Gin Router)                  │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌───────────┐  │  │
│  │  │ Middleware   │  │ API Handlers │  │  Routes   │  │  │
│  │  │ - Auth       │  │ - Health     │  │ - /v1/*   │  │  │
│  │  │ - Rate Limit │  │ - MCP        │  │ - /health │  │  │
│  │  │ - Request ID │  │ - Extensions │  │           │  │  │
│  │  └──────────────┘  └──────────────┘  └───────────┘  │  │
│  └──────────────────────────────────────────────────────┘  │
│                        │                                     │
│                        ▼                                     │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Proxy Layer                              │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌───────────┐  │  │
│  │  │   Cache      │  │   Proxy      │  │ Tool Loop │  │  │
│  │  │   Manager    │  │   Handler    │  │ Executor  │  │  │
│  │  └──────────────┘  └──────────────┘  └───────────┘  │  │
│  └──────────────────────────────────────────────────────┘  │
│                        │                                     │
│                        ▼                                     │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         MCP Client Layer (Optional)                   │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌───────────┐  │  │
│  │  │   Manager    │  │   Client     │  │   Pool    │  │  │
│  │  │   (Servers)  │  │   (Per       │  │ (HTTP     │  │  │
│  │  │              │  │   Server)    │  │  Only)    │  │  │
│  │  └──────────────┘  └──────────────┘  └───────────┘  │  │
│  └──────────────────────────────────────────────────────┘  │
│                        │                                     │
│                        ▼                                     │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         Extension System (Optional)                   │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌───────────┐  │  │
│  │  │  Registry   │  │  Workflow   │  │  Context  │  │  │
│  │  │             │  │  Executor   │  │           │  │  │
│  │  └──────────────┘  └──────────────┘  └───────────┘  │  │
│  └──────────────────────────────────────────────────────┘  │
└───────────────────────┬─────────────────────────────────────┘
                        │ HTTP Requests
                        │ (Ollama API)
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                      Ollama Server                            │
│              (Local LLM Inference Engine)                    │
└─────────────────────────────────────────────────────────────┘
```

## Component Overview

### 1. Entry Point (`cmd/llamagate/main.go`)

**Responsibilities:**
- Load configuration
- Initialize logger
- Initialize components (cache, proxy, MCP, extensions)
- Set up HTTP router and middleware
- Register API routes
- Start HTTP server
- Handle graceful shutdown

**Key Functions:**
- `main()` - Application entry point
- Component initialization
- Route registration
- Server lifecycle management

### 2. Configuration (`internal/config/`)

**Responsibilities:**
- Load configuration from multiple sources
- Environment variables
- `.env` file
- YAML/JSON config files
- Validate configuration
- Provide defaults

**Key Components:**
- `Config` struct - Main configuration container
- `Load()` - Configuration loader
- `Validate()` - Configuration validator

**Configuration Sources (Priority):**
1. Environment variables (highest)
2. Config files (YAML/JSON)
3. `.env` file
4. Default values (lowest)

### 3. HTTP Layer (`internal/api/`, `internal/middleware/`)

**Responsibilities:**
- Handle HTTP requests
- Route requests to appropriate handlers
- Apply middleware (auth, rate limiting, logging)
- Format responses
- Error handling

**Key Components:**

#### Middleware (`internal/middleware/`)

- **RequestIDMiddleware** - Generate unique request IDs
- **AuthMiddleware** - API key authentication
- **RateLimitMiddleware** - Rate limiting (leaky bucket)
- **Helpers** - Path normalization, health endpoint detection

#### API Handlers (`internal/api/`)
- **HealthHandler** - Health check endpoint
- **MCPHandler** - MCP server management endpoints
- **ExtensionHandler** - Extension management endpoints

### 4. Proxy Layer (`internal/proxy/`)

**Responsibilities:**
- Handle OpenAI-compatible requests
- Convert between OpenAI and Ollama formats
- Manage caching
- Execute tool loops (MCP tools)
- Handle streaming responses
- Inject MCP resource context

**Key Components:**
- **Proxy** - Main proxy handler
- **ToolLoop** - Multi-round tool execution
- **ResourceContext** - MCP resource injection
- **Validation** - Request validation
- **ExtensionLLMHandler** - LLM handler for extensions

**Request Flow:**
1. Receive OpenAI-format request
2. Validate request
3. Check cache (non-streaming)
4. Inject MCP resource context (if MCP enabled)
5. Execute tool loop (if tools requested)
6. Convert to Ollama format
7. Forward to Ollama
8. Convert response back to OpenAI format
9. Cache response (non-streaming)
10. Return to client

### 5. Cache Layer (`internal/cache/`)

**Responsibilities:**
- In-memory caching of requests/responses
- TTL-based expiration
- Cache key generation
- Cache size management

**Key Features:**
- TTL-based expiration
- Model-aware caching
- Message-based cache keys
- Thread-safe operations

**Cache Key Format:**
```
{model}:{hash(messages)}
```

### 6. MCP Client Layer (`internal/mcpclient/`)

**Responsibilities:**
- Connect to MCP servers
- Discover tools, resources, prompts
- Execute tools
- Manage connections
- Health monitoring
- Connection pooling (HTTP transport)

**Key Components:**
- **Client** - MCP client per server
- **ServerManager** - Manages multiple servers
- **Transport** - Communication layer (stdio, HTTP, SSE)
- **HealthMonitor** - Health checking
- **ConnectionPool** - Connection pooling (HTTP)
- **Cache** - Metadata caching

**Transport Types:**
- **Stdio** - Process-based (fully implemented)
- **HTTP** - HTTP-based (fully implemented)
- **SSE** - Server-Sent Events (stub, not implemented)

### 7. Tool Management (`internal/tools/`)

**Responsibilities:**
- Register MCP tools
- Convert MCP tools to OpenAI format
- Apply security guardrails
- Tool allow/deny lists
- Tool execution limits

**Key Components:**
- **Manager** - Tool registry
- **Mapper** - Format conversion
- **Guardrails** - Security and limits

**Tool Naming:**
- Tools are namespaced: `mcp.<serverName>.<toolName>`
- Prevents collisions between servers

### 8. Extension System (`internal/extensions/`)

**Responsibilities:**
- Extension discovery and registration
- Workflow execution
- Middleware hooks
- Observer hooks
- Extension manifest management

**Key Components:**
- **Registry** - Extension registration
- **WorkflowExecutor** - Execute agentic workflows
- **HookManager** - Middleware and observer hooks
- **Manifest** - YAML-based extension definitions
- **Types** - Core extension types (LLMHandlerFunc, etc.)

**Extension Types:**
- **Workflow Extension** - Agentic workflows with LLM calls
- **Middleware Extension** - Request/response middleware hooks
- **Observer Extension** - Response observation hooks

### 9. Logging (`internal/logger/`)

**Responsibilities:**
- Initialize structured logging
- Configure log levels
- File/console output
- Request/response logging

**Key Features:**
- JSON structured logging
- Request ID correlation
- Configurable log levels
- File and console output

## Request Flow

### Standard Chat Completion Request

```
1. Client Request
   └─> POST /v1/chat/completions
       └─> Headers: X-API-Key, Content-Type
       └─> Body: { model, messages, ... }

2. HTTP Layer
   └─> RequestIDMiddleware (generate request ID)
   └─> AuthMiddleware (validate API key)
   └─> RateLimitMiddleware (check rate limits)
   └─> LoggingMiddleware (log request)

3. Proxy Layer
   └─> Parse and validate request
   └─> Check cache (if not streaming)
   └─> Inject MCP resource context (if MCP enabled)
   └─> Execute tool loop (if tools requested)
   └─> Convert to Ollama format
   └─> Forward to Ollama

4. Ollama Processing
   └─> Load model (if not loaded)
   └─> Process request
   └─> Generate response

5. Response Handling
   └─> Receive response from Ollama
   └─> Convert to OpenAI format
   └─> Cache response (if not streaming)
   └─> Return to client

6. HTTP Layer
   └─> LoggingMiddleware (log response)
   └─> Return HTTP response
```

### Tool Execution Request Flow

```
1. Client Request (with tools)
   └─> POST /v1/chat/completions
       └─> Body: { model, messages, tools: [...] }

2. Proxy Layer
   └─> Detect tool request
   └─> Enter tool loop

3. Tool Loop (Multi-Round)
   └─> Round 1:
       ├─> Call Ollama with tools
       ├─> Model returns tool calls
       ├─> Execute tools via MCP
       └─> Inject tool results
   └─> Round 2:
       ├─> Call Ollama with tool results
       ├─> Model may call more tools
       └─> Repeat until done or limit reached

4. Final Response
   └─> Model returns final answer
   └─> Convert to OpenAI format
   └─> Return to client
```

### Extension Execution Flow

```
1. Client Request
   └─> POST /v1/extensions/:name/execute
       └─> Body: { input: {...} }

2. Extension Handler
   └─> Get extension from registry
   └─> Validate input
   └─> Execute extension

3. Extension Execution
   └─> Execute workflow (if workflow type)
       ├─> Step 1: LLM call (via LLMHandler)
       ├─> Step 2: Template render
       ├─> Step 3: File write
       └─> Step N: Final result
   └─> Return result

4. Response
   └─> Format extension result
   └─> Return to client
```

## Data Flow

### Request Transformation

```
OpenAI Format (Client)
    │
    ├─> { model, messages, temperature, stream }
    │
    ▼
LlamaGate Processing
    │
    ├─> Add MCP resource context (if enabled)
    ├─> Add tool descriptions (if tools available)
    ├─> Generate cache key
    │
    ▼
Ollama Format
    │
    ├─> { model, messages, options: { temperature }, stream }
    │
    ▼
Ollama Server
    │
    ├─> Process request
    ├─> Generate response
    │
    ▼
Ollama Response
    │
    ├─> { message: { role, content }, ... }
    │
    ▼
LlamaGate Processing
    │
    ├─> Convert to OpenAI format
    ├─> Cache response (if not streaming)
    │
    ▼
OpenAI Format (Client)
    │
    ├─> { id, object, created, model, choices: [...] }
```

### MCP Tool Flow

```
Client Request
    │
    ├─> { model, messages, tools: [...] }
    │
    ▼
Tool Loop
    │
    ├─> Round 1:
    │   ├─> Call Ollama → Model returns tool_calls
    │   ├─> Extract tool calls
    │   ├─> Execute tools via MCP
    │   └─> Inject tool results
    │
    ├─> Round 2:
    │   ├─> Call Ollama with results
    │   ├─> Model may call more tools
    │   └─> Repeat...
    │
    └─> Final round:
        └─> Model returns final answer
```

## Component Interactions

### Component Dependency Graph

```
main.go
  │
  ├─> config.Load()
  │   └─> Loads from .env, YAML, environment
  │
  ├─> logger.Init()
  │   └─> Sets up Zerolog
  │
  ├─> cache.New()
  │   └─> Creates in-memory cache
  │
  ├─> proxy.NewWithTimeout()
  │   ├─> Uses cache
  │   └─> Creates HTTP client
  │
  ├─> setup.InitializeMCP()
  │   ├─> Creates ServerManager
  │   ├─> Creates Clients (per server)
  │   ├─> Creates ConnectionPool (HTTP)
  │   ├─> Creates HealthMonitor
  │   └─> Discovers tools/resources/prompts
  │
  ├─> setup.ConfigureProxy()
  │   ├─> Sets tool manager
  │   └─> Sets guardrails
  │
  └─> extensions.NewRegistry()
      └─> Creates extension registry
      └─> extensions.DiscoverExtensions()
          └─> Discovers extensions from extensions/ directory
```

### Component Communication

**Proxy ↔ Cache:**
- Proxy checks cache before forwarding requests
- Proxy stores responses in cache
- Cache provides TTL-based expiration

**Proxy ↔ MCP:**
- Proxy uses ToolManager to get available tools
- Proxy uses ToolManager to execute tools
- ToolManager uses MCP clients to execute tools

**Proxy ↔ Extensions:**
- Proxy provides LLM handler to extensions
- Extensions can make LLM calls through proxy
- Extensions can access MCP tools via workflow steps

**MCP ↔ Tools:**
- MCP clients discover tools from servers
- ToolManager registers tools from MCP
- ToolManager converts MCP tools to OpenAI format

## Design Patterns

### 1. Layered Architecture

**Layers:**
1. **HTTP Layer** - Request/response handling
2. **Proxy Layer** - Business logic, format conversion
3. **Service Layer** - MCP, extensions, tools
4. **Transport Layer** - HTTP client, MCP transports

**Benefits:**
- Clear separation of concerns
- Easy to test
- Easy to modify

### 2. Dependency Injection

**Pattern:**
- Components receive dependencies via constructors
- No global state
- Easy to mock for testing

**Example:**
```go
proxy := proxy.NewWithTimeout(ollamaHost, cache, timeout)
```

### 3. Interface-Based Design

**Pattern:**
- Components communicate via interfaces
- Easy to swap implementations
- Better testability

**Examples:**
- `Transport` interface (stdio, HTTP, SSE)
- `LLMHandlerFunc` interface (for extensions)
- `ServerManagerInterface` (for proxy)

### 4. Registry Pattern

**Pattern:**
- Centralized registration and lookup
- Thread-safe access
- Plugin and tool registries

**Examples:**
- `ExtensionRegistry` - Extension registration
- `ToolManager` - Tool registration
- `ServerManager` - MCP server registration

### 5. Factory Pattern

**Pattern:**
- Factory functions for creating instances
- Configuration-based creation
- Default values

**Examples:**
- `NewProxy()`, `NewClient()`, `NewRegistry()`
- `DefaultPoolConfig()`, `DefaultPoolConfig()`

## Concurrency Model

### Goroutines

**Background Goroutines:**
- **HealthMonitor** - Periodic health checks
- **Cache Cleanup** - TTL-based cache cleanup
- **MCP Connection Pool** - Connection management

**Request Handling:**
- Each HTTP request handled in separate goroutine
- Gin router manages goroutine pool
- No shared mutable state (except caches with locks)

### Synchronization

**Mutexes:**
- Cache operations (read/write locks)
- Extension registry (read/write locks)
- MCP server manager (read/write locks)
- Connection pool (mutex for pool operations)

**Channels:**
- Health monitor stop signal
- Cache cleanup stop signal
- Graceful shutdown coordination

**sync.Once:**
- Health monitor start (prevents race conditions)
- Cache cleanup start (prevents race conditions)
- Health monitor stop (prevents double-close)
- Cache stop (prevents double-close)

## Error Handling

### Error Propagation

**Pattern:**
- Errors bubble up from lower layers
- Structured error responses
- Request IDs for tracing

**Error Types:**
- `ValidationError` - Invalid input
- `InternalError` - Server errors
- `ServiceUnavailable` - MCP/extension system unavailable
- `NotFound` - Resource not found
- `RateLimitError` - Rate limit exceeded

### Error Response Format

```json
{
  "error": {
    "message": "Error description",
    "type": "error_type",
    "request_id": "550e8400-..."
  }
}
```

## Configuration Management

### Configuration Sources

**Priority Order:**
1. Environment variables (highest)
2. YAML/JSON config files
3. `.env` file
4. Default values (lowest)

### Configuration Loading

**Process:**
1. Load `.env` file (if exists)
2. Load YAML/JSON config (if exists)
3. Override with environment variables
4. Validate configuration
5. Apply defaults for missing values

### Configuration Structure

```go
type Config struct {
    // Core
    OllamaHost         string
    APIKey             string
    RateLimitRPS       float64
    Debug              bool
    Port               string
    LogFile            string
    Timeout            time.Duration
    
    // MCP (optional)
    MCP                *MCPConfig
    
    // Plugins (optional)
    Plugins            *PluginsConfig
}
```

## Key Design Decisions

### 1. Single Binary

**Decision:** Everything compiled into one executable

**Rationale:**
- Easy deployment
- No dependency management
- Fast startup
- Simple distribution

### 2. In-Memory Cache

**Decision:** Cache is in-memory only, lost on restart

**Rationale:**
- Simple implementation
- Fast access
- No external dependencies
- Good for most use cases

**Trade-off:**
- Cache lost on restart
- Limited by memory
- No persistence

### 3. Optional MCP

**Decision:** MCP is optional, disabled by default

**Rationale:**
- Reduces complexity for basic use cases
- Only enable when needed
- Faster startup without MCP

### 4. Extension System

**Decision:** YAML-based extension system for custom functionality

**Rationale:**
- Allows customization without modifying core
- Enables agentic workflows
- Model-friendly (YAML manifest definitions)
- No compilation required

### 5. OpenAI Compatibility

**Decision:** Perfect OpenAI API compatibility

**Rationale:**
- Zero migration effort
- Same SDKs work
- Drop-in replacement
- This is the core value proposition

## Performance Considerations

### Caching Strategy

- **Cache Key:** Model + message hash
- **TTL:** Configurable (default: 5 minutes)
- **Size Limits:** Configurable
- **Thread-Safe:** Read/write locks

### Connection Pooling

- **HTTP Transport:** Connection pooling enabled
- **Pool Size:** Configurable (default: 10)
- **Idle Timeout:** Configurable (default: 5 minutes)
- **Reuse:** Connections reused across requests

### Rate Limiting

- **Algorithm:** Leaky bucket
- **Scope:** Global (all requests)
- **Configurable:** Requests per second
- **Response:** 429 Too Many Requests

## Security Architecture

### Authentication

- **Method:** API key via header
- **Header:** `X-API-Key` or `Authorization: Bearer`
- **Optional:** Can be disabled
- **Implementation:** Constant-time comparison

### Rate Limiting

- **Algorithm:** Leaky bucket
- **Scope:** Global
- **Configurable:** RPS limit
- **Response:** 429 with retry-after

### Tool Security

- **Allow Lists:** Glob patterns for allowed tools
- **Deny Lists:** Glob patterns for denied tools
- **Timeouts:** Per-tool execution timeouts
- **Size Limits:** Maximum result size
- **Round Limits:** Maximum tool execution rounds

## Extension Points

### 1. Extensions

**How to Extend:**
- Create `manifest.yaml` in `extensions/` directory
- Define workflow steps, middleware hooks, or observer hooks
- Extensions are auto-discovered at startup
- Access LLM via `LLMHandlerFunc` in workflow steps
- Access MCP tools via workflow steps

### 2. Custom Workflows

**How to Extend:**
- Create workflow extension with `manifest.yaml`
- Define steps: template.load, template.render, llm.chat, file.write
- Extensions execute via `POST /v1/extensions/:name/execute`

### 3. MCP Servers

**How to Extend:**
- Add MCP server to config
- Server automatically discovered
- Tools automatically exposed

## Future Architecture Considerations

### Potential Enhancements

1. **Persistent Cache**
   - Redis integration
   - File-based cache
   - Database-backed cache

2. **HTTPS/TLS Support**
   - Native TLS support
   - Let's Encrypt integration
   - Certificate management

3. **Monitoring Dashboard**
   - Health dashboard
   - Metrics visualization
   - Performance monitoring

4. **Clustering**
   - Multi-instance support
   - Load balancing
   - Shared cache

5. **Plugin Marketplace**
   - Plugin discovery
   - Plugin sharing
   - Plugin versioning

## Related Documentation

- [Project Structure](STRUCTURE.md) - Directory structure
- [MCP Integration](MCP.md) - MCP client details
- [Extension System](EXTENSIONS_SPEC_V0.9.1.md) - Extension system details
- [API Reference](API.md) - HTTP API details
- [Configuration Guide](../README.md#configuration) - Configuration options

---

**Last Updated:** 2026-01-09
