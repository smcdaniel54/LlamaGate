# Changelog

All notable changes to LlamaGate will be documented in this file.

## [Unreleased]

### Breaking Changes (Phase 1 – Core-only gateway)

- **Extensions and agentic modules removed**: The extension system and agentic modules have been removed. LlamaGate is now a lean, OpenAI-compatible gateway only.

  **Removed:**
  - All `/v1/extensions` endpoints (list, get, upsert, execute, refresh) and dynamic extension routes
  - Extension and agentic-module loading, discovery, packaging, and registry
  - The `llamagate-cli` tool (import/export/list/remove/enable/disable extensions and modules, migrate, sync)
  - Config option `EXTENSIONS_UPSERT_ENABLED`
  - Root directories `extensions/` and `examples/agenticmodules/`

  **Still supported:**
  - Core endpoints: `/health`, `/v1/hardware/recommendations`, `POST /v1/chat/completions`, `GET /v1/models`
  - Full MCP support: `/v1/mcp/*` when MCP is enabled
  - All existing config except `EXTENSIONS_UPSERT_ENABLED`

  **Migration:** See [Core Contract](docs/core_contract.md) and [Migration Notes](README.md#migration-notes-phase-1-extensionsmodules-removed) in the README.

### Fixed
- **Build from source**: Resolved invalid string literals that caused Go compile errors on some environments; `go build ./...` succeeds everywhere.

### Changed
- **CI / Build**: CI and build-binaries workflows run an explicit `go build ./...` step before tests and release builds.

## [0.11.0] - 2026-02-09

### Added

- **Phase 2: Memory & Introspection** – Self-awareness and file-based memory without agentic workflows.
  - **File-based memory store** (`internal/memory`): atomic JSON writes, per-file locking, configurable caps (pinned/recent/notes).
  - **System endpoints** (when `INTROSPECTION_ENABLED=true`): `GET /v1/system/info`, `/v1/system/hardware`, `/v1/system/models`, `/v1/system/health`, `/v1/system/config`, `/v1/system/memory`. Response envelope: `{ "ok": true, "data": ... }`.
  - **Config**: `INTROSPECTION_*`, `MEMORY_*` (dir, limits). Sanitized config and hardware output (no secrets).
  - **Optional chat system card**: short system message injected into chat when `INTROSPECTION_CHAT_INJECT_ENABLED=true` (hardware/models/LlamaGate summary; optional user memory via `X-LlamaGate-User` header).
  - **Docs**: `docs/core_contract.md` (Phase 2), `docs/phase2_memory_introspection.md`.

### Fixed

- **Lint**: errcheck and staticcheck fixes across introspection, memory, and proxy tests.

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

  **Migration:** See `docs/phase1_remove_modules_extensions.md` and `docs/core_contract.md`.

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

(Release links omitted; use your repository's tags and compare URLs.)
