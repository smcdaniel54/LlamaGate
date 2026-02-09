# Phase 2: Memory & Introspection (Design)

Lean "self-awareness + file-based memory + hardware/model introspection" layer. No agentic workflows, no database, no secrets in memory or config output.

---

## Step 0 — Repo Orientation (Summary)

| Area | Location | Responsibility |
|------|----------|----------------|
| **A) Server / router** | `cmd/llamagate/main.go` | Gin router creation, global middleware (recovery, request ID, logging), auth/rate-limit; `router.GET("/health", ...)`, `router.GET("/v1/hardware/recommendations", ...)`, `v1 := router.Group("/v1")` with chat/completions, models, and conditional MCP group. |
| **B) Config** | `internal/config/config.go` | `Config` struct (OllamaHost, APIKey, Port, Timeout, MCP, TLS, etc.). `Load()` uses godotenv + viper; YAML/JSON from `.` and `$HOME/.llamagate`. No dedicated "data dir" yet—default memory dir will use homedir. |
| **C) Models / provider** | `internal/proxy/proxy.go` | No separate model registry. `HandleModels` GETs `OllamaHost/api/tags` and maps Ollama's `models` array to OpenAI-style list. Backend is single Ollama host. |
| **D) Health / info** | `internal/api/health.go` | `HealthHandler.CheckHealth` GETs `OllamaHost/api/tags` to verify Ollama reachability. Returns JSON with status, ollama_host. No `/info` endpoint yet. |

**Data dir:** `internal/homedir/homedir.go` provides `GetHomeDir()` → `~/.llamagate`. Phase 2 memory default: `~/.llamagate/data/memory` (or configurable).

**Build info:** No ldflags for version/commit in codebase; introspection will use `runtime.Version()` and optional build vars (settable via ldflags later).

---

## Goals

- Accurate chat about: local hardware, current model/provider info, LlamaGate capabilities/version/config (sanitized).
- File-based memory (JSON, atomic writes, locking); no DB.
- Read-only introspection by default; no mutating system state via chat.
- No secrets (API keys, tokens) in config output or memory.

## Non-Goals

- No agentic workflows, planners, or extension systems.
- No database or new heavy dependencies.
- No storing full chat transcripts; optional small system-card injection only.

---

## Endpoints (when introspection enabled)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/system/info` | Runtime snapshot + capability flags (mcp, memory, introspection). |
| GET | `/v1/system/hardware` | Hardware snapshot (sanitized per detail_level). |
| GET | `/v1/system/models` | Models snapshot (from backend). |
| GET | `/v1/system/health` | Backend health snapshot. |
| GET | `/v1/system/config` | Sanitized config (no secrets). |
| GET | `/v1/system/memory` | Memory status/summary only (counts, last updated). |

All return JSON envelope: `{ "ok": true, "data": <payload> }` or `{ "ok": false, "error": { "code": "...", "message": "..." } }`.

---

## Redaction behavior

- **Config:** Keys whose names (case-insensitive) contain secret-like substrings (e.g. `api_key`, `token`, `password`, `secret`, `auth`, `credential`, `private`) are replaced with `"<redacted>"`. Applied recursively to nested maps and arrays. Implemented in `internal/introspection/sanitize.go`.
- **Hardware:** `minimal` (default): no hostnames, usernames, serial/MAC, device IDs. `standard`: may include short hostname and more disk details. `full`: all collectible fields; still no secrets.
- **URLs:** Backend endpoint URLs in models/health output are sanitized (no user/password in query or userinfo).

---

## Schemas (summary)

- **SystemMemory:** `updated_at`, `notes[]` (capped), `capabilities` (endpoints, mcp_enabled, memory_enabled, introspection_enabled), optional `last_hardware_snapshot`, `last_models_snapshot`. Stored at `data/memory/system.json`.
- **UserMemory:** `user_id`, `updated_at`, `pinned[]`, `recent[]`, `tags[]` (all capped). Stored at `data/memory/users/<user_id>.json`.
- **Memory status (GET /v1/system/memory):** `system_updated_at`, `user_count`, `session_count`, `total_size_bytes`—no full dumps by default.

---

## When endpoints are unavailable

- **Introspection disabled** (`INTROSPECTION_ENABLED=false`): `/v1/system/*` routes are not registered; any request to those paths returns **404**.
- **Memory disabled** (`MEMORY_ENABLED=false`): GET `/v1/system/memory` returns **404** with envelope `{ "ok": false, "error": { "code": "disabled", "message": "memory is disabled" } }`. Chat system card does not include user memory.

---

## Optional chat system card

When **INTROSPECTION_CHAT_INJECT_ENABLED** is true, a short system message (target ≤40 tokens, hard cap 320 chars) is prepended to chat completion requests. It may include:

- LlamaGate version and enabled features (if **INTROSPECTION_CHAT_INJECT_INCLUDE_LLAMAGATE_INFO**).
- Hardware summary (if **INTROSPECTION_CHAT_INJECT_INCLUDE_HARDWARE**).
- Models summary (if **INTROSPECTION_CHAT_INJECT_INCLUDE_MODELS**).
- User pinned memory summary (if memory is enabled and a user ID is known).

**User identification:** Optional header `X-LlamaGate-User: <id>` supplies a stable user ID for user memory. If not provided, user memory is not included in the card. Raw API keys are never stored; a hashed key could be used in future for derived user ID.

Details implemented in `internal/introspection/systemcard.go` and `internal/proxy/proxy.go` (SetSystemCardFunc, injection before forwarding to backend).
