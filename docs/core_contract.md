# LlamaGate Core Contract (Phase 1 + Phase 2)

After Phase 1 removal of agentic modules and extensions, LlamaGate is a **lean OpenAI-compatible gateway**. Phase 2 adds a read-only self-awareness layer (memory + introspection). This document defines the core contract: what remains, what is removed, and backward-compatibility notes.

---

## Core endpoints (must remain functional)

These are the HTTP endpoints implemented and supported after Phase 1 (and Phase 2 where noted):

| Method | Path | Handler | Notes |
|--------|------|---------|--------|
| GET | `/health` | api.HealthHandler | No auth required. |
| GET | `/v1/hardware/recommendations` | api.HardwareHandler | No auth required. |
| POST | `/v1/chat/completions` | proxy.HandleChatCompletions | OpenAI-compatible; routes to configured backend (Ollama). Supports streaming, tool/function calling when MCP enabled. Optional system card injection (Phase 2). |
| GET | `/v1/models` | proxy.HandleModels | OpenAI-compatible models list. |
| GET | `/v1/system/info` | api.SystemHandler | **Phase 2.** Only when introspection enabled. Runtime + capabilities. |
| GET | `/v1/system/hardware` | api.SystemHandler | **Phase 2.** Only when introspection enabled. Hardware snapshot (sanitized). |
| GET | `/v1/system/models` | api.SystemHandler | **Phase 2.** Only when introspection enabled. Backend models snapshot. |
| GET | `/v1/system/health` | api.SystemHandler | **Phase 2.** Only when introspection enabled. Backend health. |
| GET | `/v1/system/config` | api.SystemHandler | **Phase 2.** Only when introspection enabled. Sanitized config (no secrets). |
| GET | `/v1/system/memory` | api.SystemHandler | **Phase 2.** Only when introspection and memory enabled. Status/summary only. |
| GET | `/v1/mcp/servers` | api.MCPHandler | Only when MCP enabled. |
| GET | `/v1/mcp/servers/health` | api.MCPHandler | Only when MCP enabled. |
| GET | `/v1/mcp/servers/:name` | api.MCPHandler | Only when MCP enabled. |
| GET | `/v1/mcp/servers/:name/health` | api.MCPHandler | Only when MCP enabled. |
| GET | `/v1/mcp/servers/:name/stats` | api.MCPHandler | Only when MCP enabled. |
| GET | `/v1/mcp/servers/:name/tools` | api.MCPHandler | Only when MCP enabled. |
| GET | `/v1/mcp/servers/:name/resources` | api.MCPHandler | Only when MCP enabled. |
| GET | `/v1/mcp/servers/:name/resources/*uri` | api.MCPHandler | Only when MCP enabled. |
| GET | `/v1/mcp/servers/:name/prompts` | api.MCPHandler | Only when MCP enabled. |
| POST | `/v1/mcp/servers/:name/prompts/:promptName` | api.MCPHandler | Only when MCP enabled. |
| POST | `/v1/mcp/execute` | api.MCPHandler | Only when MCP enabled. |
| POST | `/v1/mcp/servers/:name/refresh` | api.MCPHandler | Only when MCP enabled. |

**Not implemented in current codebase:** `/v1/embeddings` — not part of the core contract until/unless added.

---

## Core config fields (remain supported)

- **OLLAMA_HOST** — Backend for chat/models.
- **API_KEY** — Optional; when set, auth middleware is applied to protected routes.
- **RATE_LIMIT_RPS** — Rate limiting.
- **DEBUG** — Log level / Gin mode.
- **PORT** — Server listen port.
- **LOG_FILE** — Optional log file path.
- **TIMEOUT** — HTTP client timeout.
- **HEALTH_CHECK_TIMEOUT** — Health check timeout.
- **SHUTDOWN_TIMEOUT** — Graceful shutdown timeout.
- **TLS_ENABLED**, **TLS_CERT_FILE**, **TLS_KEY_FILE** — HTTPS.
- **MCP_*** — All MCP-related env vars and config (MCP is core). See README / config for full list.

**Phase 2 config (optional):**

- **INTROSPECTION_ENABLED** — When true, `/v1/system/*` routes are registered. When false, requests to those paths return 404.
- **INTROSPECTION_HARDWARE_ENABLED**, **INTROSPECTION_HARDWARE_DETAIL_LEVEL** (`minimal` \| `standard` \| `full`) — Hardware snapshot behavior.
- **INTROSPECTION_CHAT_INJECT_ENABLED**, **INTROSPECTION_CHAT_INJECT_INCLUDE_HARDWARE**, **INTROSPECTION_CHAT_INJECT_INCLUDE_MODELS**, **INTROSPECTION_CHAT_INJECT_INCLUDE_LLAMAGATE_INFO** — Optional system card injected into chat requests.
- **MEMORY_ENABLED** — When true, file-based memory store is used; when false, `/v1/system/memory` returns 404 and chat injection does not use user memory.
- **MEMORY_DIR** — Directory for memory JSON files (default: `~/.llamagate/data/memory` when empty).
- **MEMORY_PINNED_MAX**, **MEMORY_RECENT_MAX**, **MEMORY_MAX_ENTRY_CHARS** — Caps for memory entries.

**Removed:** `EXTENSIONS_UPSERT_ENABLED` (extensions removed in Phase 1).

**Phase 2 guarantees:** Introspection endpoints are read-only. Config output is sanitized (no API keys, tokens, passwords, or secret-like fields). Memory stores no secrets. No agentic workflows or state mutation via chat.

---

## Explicitly removed (Phase 1)

- **Agentic modules** — No loading, discovery, or execution of agentic modules (e.g. `agenticmodule.yaml`, `agenticmodule_runner`).
- **Extensions** — No extension registry, workflow executor, or extension-specific endpoints.
- **Extension/module endpoints:**
  - `GET /v1/extensions`
  - `GET /v1/extensions/:name`
  - `PUT /v1/extensions/:name`
  - `POST /v1/extensions/:name/execute`
  - `POST /v1/extensions/refresh`
  - Any dynamic routes under `/v1/extensions/...` (e.g. custom extension endpoints).
- **Extension middleware and response hooks** — No request-inspector style middleware, no cost-usage-reporter style response hooks.
- **Workflow orchestration** — No YAML-defined workflow steps, tool dispatch for extensions, planners, multi-step loops, or replay systems.
- **Packaging / discovery / registry** — No import/export of extensions or modules, no installed-items registry for extensions/agentic modules.
- **Migration** — No legacy extension migration on first run.

---

## Backward-compatibility notes

- **EXTENSIONS_UPSERT_ENABLED** has been removed from config; remove it from `.env` or config files if present.
- **Removed endpoints:** Requests to `/v1/extensions` or `/v1/extensions/*` return **404** (routes no longer registered).

---

## MCP status

**MCP is core.** It is not implemented as an extension layer. The MCP client, tool manager, guardrails, `/v1/mcp/*` API, and tool execution in the chat completion flow remain fully supported after Phase 1.
