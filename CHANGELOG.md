# Changelog

All notable changes to LlamaGate will be documented in this file.

## [0.9.1] - 2026-01-15

### Breaking Changes

- **Plugin System Removed**: The Go-based plugin system has been completely removed and replaced with the YAML-based extension system. This is a **breaking change** requiring migration.

  **Migration Required:**
  - `/v1/plugins` endpoints → `/v1/extensions` endpoints
  - `plugins/` directory → `extensions/` directory
  - Go plugin code → YAML manifest files
  - `PLUGINS_ENABLED` env var → Extensions auto-discovered (no config needed)
  - `PluginsConfig` → Removed (extensions use YAML manifests)

  **What Changed:**
  - All plugin code in `internal/plugins/` has been removed
  - All plugin code in `plugins/` directory has been removed
  - Plugin API endpoints (`/v1/plugins/*`) have been removed
  - Plugin configuration (`PluginsConfig`) has been removed
  - Extension system is now the only extensibility mechanism

  **Migration Guide:** See `docs/PLANS_REVIEW.md` for migration completion details.

### Added

- **Extension System v0.9.1**: YAML-based extension system for workflows, middleware, and observability
  - Auto-discovery from `extensions/` directory
  - YAML manifest-based definitions (no compilation required)
  - Support for workflow, middleware, and observer extension types
  - Enable/disable functionality via manifest or environment variables
  - Comprehensive extension API endpoints (`/v1/extensions/*`)
  - Example extensions included: prompt-template-executor, request-inspector, cost-usage-reporter

### Changed

- **Default port changed from 8080 to 11435**: The default server port has been changed from `8080` to `11435` to avoid conflicts with common services like Jenkins, Tomcat, and development servers. Port 11435 follows Ollama's port pattern (11434 + 1) and is easy to remember. Existing installations with `.env` files will continue using their configured port. New installations will use port 11435 by default. You can override the port using the `PORT` environment variable.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- **Build from source**: Resolved invalid string literals in `internal/extensions/workflow.go` (`resolveTemplateString`) that caused Go compile errors on some environments (e.g. Windows) and broke downstream tooling (CI, E2E tests, forked automation) that build LlamaGate from source. Runtime behavior was unchanged; the fix ensures `go build ./...` succeeds everywhere.

### Changed
- **Workflow upsert default**: Upsert (`PUT /v1/extensions/:name`) is now **enabled by default**. Set `EXTENSIONS_UPSERT_ENABLED=false` to lock down.
- **CI / Build**: CI and build-binaries workflows now run an explicit `go build ./...` step so that build-from-source is validated before tests and release builds; failures are caught early for integrators.
- **Docs**: Installation and contributing docs updated to recommend `go build ./...` before tests and to document valid Go string literal usage so downstream build-from-source remains reliable.

### Added
- **Workflow upsert**: `PUT /v1/extensions/:name` to create or update an extension manifest in `~/.llamagate/extensions/installed/`. **Enabled by default**; set `EXTENSIONS_UPSERT_ENABLED=false` to lock down. Clients (e.g. LlamaGate Control) can save workflows to LlamaGate; after upsert, call `POST /v1/extensions/refresh` to load. When disabled, the endpoint returns 501 with `code: UPSERT_NOT_CONFIGURED`. See [API.md](docs/API.md#upsert-extension-optional) and `.env.example`.

## [0.9.0] - 2026-01-05

**Note:** This is a pre-1.0.0 release to gather community feedback. The API may evolve based on user feedback before reaching 1.0.0.

### Added
- **OpenAI-Compatible API**: Drop-in replacement for OpenAI API endpoints
  - `/v1/chat/completions` endpoint with full streaming support
  - `/v1/models` endpoint for model listing
  - `/health` endpoint for health checks
- **MCP (Model Context Protocol) Client Support**: Connect to MCP servers and expose their tools to models
  - Tool discovery and automatic namespacing (`mcp.<server>.<tool>`)
  - Multi-round tool execution loop
  - Support for stdio transport (SSE transport interface prepared)
  - Tool execution guardrails (allow/deny lists, timeouts, result size limits)
  - MCP configuration via YAML/JSON files or environment variables
  - Comprehensive MCP documentation and demo examples
- **Caching**: In-memory caching for identical prompts to reduce Ollama load
  - TTL-based expiration
  - Configurable cache size limits
  - Cache key based on full request context
- **Authentication**: Optional API key authentication via headers
  - Constant-time comparison to prevent timing attacks
  - Configurable via environment variables or config files
- **Rate Limiting**: Configurable rate limiting using leaky bucket algorithm
  - Global rate limiting
  - Configurable requests per second
- **Structured Logging**: JSON logging with request IDs using Zerolog
  - Request/response logging
  - Error tracking with context
  - Configurable log levels
  - Secure log file permissions (0600)
- **Streaming Support**: Full support for streaming chat completions
  - Server-Sent Events (SSE) streaming
  - Proper streaming error handling
- **Tool/Function Calling**: Execute MCP tools in multi-round loops
  - Automatic tool call detection
  - Tool result injection
  - Configurable max rounds and calls per round
- **Graceful Shutdown**: Clean shutdown on SIGINT/SIGTERM
  - Proper resource cleanup
  - In-flight request handling
- **Configuration Management**: Flexible configuration system
  - Environment variables
  - `.env` file support
  - YAML/JSON config files (for MCP)
  - Configuration validation
  - Sensible defaults
- **Cross-Platform Support**: Windows, Linux, and macOS
  - Platform-specific installers
  - Platform-specific scripts
  - Single binary deployment
- **Docker Support**: Multi-stage Dockerfile for minimal image size
- **Comprehensive Documentation**:
  - Quick Start Guide
  - MCP Quick Start Guide
  - MCP Demo Guide with multiple server examples
  - API documentation
  - Security best practices
  - Contributing guidelines
  - Installation guides

### Security
- Constant-time API key comparison to prevent timing attacks
- Secure log file permissions (0600)
- Request body size limits (recommended)
- Tool execution guardrails (allow/deny lists, timeouts)
- Input validation and sanitization
- Security policy and responsible disclosure process

### Fixed
- Cache memory leak (added TTL and size limits)
- Timing attack vulnerability in authentication
- Insecure file permissions for log files
- Missing configuration validation
- Missing request IDs in error responses
- Cache lookup using incorrect message context (tool injection)

[Unreleased]: https://github.com/llamagate/llamagate/compare/v0.9.1...HEAD
[0.9.1]: https://github.com/llamagate/llamagate/compare/v0.9.0...v0.9.1
[0.9.0]: https://github.com/llamagate/llamagate/releases/tag/v0.9.0
