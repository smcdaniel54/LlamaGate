# Changelog

All notable changes to LlamaGate will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Future features and improvements

## [1.0.0] - 2026-01-05

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

[Unreleased]: https://github.com/llamagate/llamagate/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/llamagate/llamagate/releases/tag/v1.0.0
