# Changelog

All notable changes to LlamaGate will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release
- OpenAI-compatible API endpoints (`/v1/chat/completions`, `/v1/models`)
- In-memory caching with TTL and size limits
- Optional API key authentication
- Rate limiting with leaky bucket algorithm
- Structured JSON logging with request IDs
- Streaming support for chat completions
- Graceful shutdown handling
- Health check endpoint with Ollama connectivity verification
- Configuration validation
- Configurable HTTP client timeout
- Cross-platform support (Windows, Linux, macOS)
- Docker support
- Comprehensive documentation

### Security
- Constant-time API key comparison to prevent timing attacks
- Secure log file permissions (0600)
- Request body size limits (recommended)

### Fixed
- Cache memory leak (added TTL and size limits)
- Timing attack vulnerability in authentication
- Insecure file permissions for log files
- Missing configuration validation
- Missing request IDs in error responses

## [0.1.0] - YYYY-MM-DD

### Added
- Initial public release

[Unreleased]: https://github.com/llamagate/llamagate/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/llamagate/llamagate/releases/tag/v0.1.0

