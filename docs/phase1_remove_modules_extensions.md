# Phase 1: Remove Modules & Extensions — Impact Map

This document inventories all references to **agent**, **agentic**, **workflow**, **module**, **modules**, **extension**, **extensions**, **tool** (extension-layer), **planner**, **orchestr**, and **MCP** (MCP is core; retained). It identifies what to remove for a lean OpenAI-compatible gateway and what remains.

---

## Search term summary

| Term | Where it appears | Action |
|------|------------------|--------|
| **agent / agentic** | `internal/extensions/` (AgenticModule, agent caller, events) | Remove with extensions |
| **workflow** | `internal/extensions/` (WorkflowExecutor, workflow steps, workflow type) | Remove with extensions |
| **module / modules** | `internal/extensions/`, `internal/startup/`, `internal/discovery/`, `internal/packaging/`, `internal/registry/`, `internal/homedir/` | Remove module loading; keep homedir/registry only if used by core (registry is extension/module-only → remove usage from main) |
| **extension / extensions** | `cmd/llamagate/main.go`, `internal/extensions/`, `internal/startup/`, `internal/config/`, discovery, packaging, handler, route_manager | Remove all extension loading, routes, handlers |
| **tool / tools** | **Core (keep):** `internal/tools/` (MCP tool manager, guardrails), `internal/proxy/tool_loop.go`, `internal/api/mcp.go`. **Extension (remove):** `internal/extensions/builtin/tools/` (extension tool framework), builtin types | Remove only extension-layer tools |
| **planner / orchestr** | No matches in Go code | N/A |
| **MCP** | `internal/setup/mcp.go`, `internal/api/mcp.go`, `internal/mcpclient/`, `internal/tools/`, `internal/proxy/` (tool_loop, resource context), config | **KEEP** — MCP is core, not an extension |

---

## A) Directories related to modules/extensions

| Path | Purpose | Safe to delete? |
|------|---------|------------------|
| **internal/extensions/** | Entire extension system: handler, workflow executor, hooks, route manager, manifest/agenticmodule loading, registry (in-memory), execution context, builtin (agent, decision, debug, state, tools, validation, transform, human, events, core) | **Yes — delete entire directory** |
| **internal/startup/** | `LoadInstalledExtensions()`, `LoadInstalledModules()`; calls discovery, packaging, extensions.DiscoverExtensions | **Remove or gut:** Remove extension/module loading; package can stay if we need other startup logic (currently only extensions/modules) → **Delete package** or leave empty stub |
| **internal/discovery/** | `DiscoverInstalledItems()`, `DiscoverEnabledItems()`, `DiscoverLegacyExtensions()`, scans extensions + modules dirs, uses registry & packaging | **Yes — delete** (only used for extensions/modules) |
| **internal/packaging/** | Import/Export/Remove for extensions and modules; LoadExtensionPackageManifest, LoadModulePackageManifest | **Yes — delete** (only used for extensions/modules) |
| **internal/registry/** | Installed items registry (extensions + agentic modules); used by discovery, packaging, startup | **Yes — delete** (only used for extensions/modules) |
| **internal/migration/** | `MigrateLegacyExtensions()` — migrates legacy extensions to new layout | **Yes — delete** (extension-only) |
| **internal/examples/** | Example extensions (visual, logging, extension.go) | **Yes — delete** (extension examples only) |
| **extensions/** (repo root) | YAML manifests and extension dirs (request-inspector, cost-usage-reporter, etc.) | **Yes — delete** (or keep as reference; user said remove functionality, not necessarily delete folder — **delete for clean core**) |
| **examples/agenticmodules/** | Example agentic module (intake_and_routing) | **Yes — delete** |

**Core directories to keep (unchanged for Phase 1):**

- `internal/api/` — keep health, hardware, **MCP** (mcp.go)
- `internal/config/` — keep; remove/deprecate only `ExtensionsUpsertEnabled`
- `internal/homedir/` — keep; used by migration/registry/discovery (all removed). If no other code uses GetExtensionsDir/GetAgenticModulesDir/GetRegistryDir, we can leave them for backward compat or remove later.
- `internal/proxy/` — keep; remove `plugin_handler.go` (CreateExtensionLLMHandler) and `plugin_handler_test.go`; do not call SetResponseHook/CreateExtensionLLMHandler from main
- `internal/tools/` — **keep** (MCP tool manager, guardrails)
- `internal/mcpclient/`, `internal/setup/` — **keep** (MCP is core)

---

## B) HTTP routes for modules/extensions

| Location | Route(s) | Action |
|---------|----------|--------|
| **cmd/llamagate/main.go** | `v1.Group("/extensions")`: GET `""`, GET `"/:name"`, PUT `"/:name"`, POST `"/:name/execute"`, POST `"/refresh"` | **Remove** entire extensions group |
| **internal/extensions/route_manager.go** | Dynamic routes per extension: `ExtensionRoutePrefix = "/v1/extensions"`, `RegisterExtensionRoutes(manifest)` | **Remove** (delete route_manager and all registration in main) |

No other routes are extension/module-specific. Core routes to **keep**:

- `GET /health`
- `GET /v1/hardware/recommendations`
- `POST /v1/chat/completions`, `GET /v1/models`
- `v1/mcp/*` (when MCP enabled)

---

## C) Middleware / request–response hooks

| Hook | Where registered | Purpose | Action |
|------|------------------|---------|--------|
| **extensionHookManager.CreateMiddlewareHook()** | main.go | Request inspection (e.g. request-inspector) | **Remove** — do not use middleware from extensions |
| **proxyInstance.SetResponseHook(extensionHookManager.ExecuteResponseHooks)** | main.go | Response observers (e.g. cost-usage-reporter) | **Remove** — stop calling SetResponseHook in main (proxy can keep the setter; hook will be nil) |

No stubbing required: main simply stops creating the hook manager and stops calling SetResponseHook.

---

## D) Config fields

| Field / env | File | Action |
|-------------|------|--------|
| **ExtensionsUpsertEnabled** / **EXTENSIONS_UPSERT_ENABLED** | internal/config/config.go | **Removed** — field and env var no longer exist. |

No other config is exclusively for modules/extensions. MCP config remains.

---

## E) Tests that depend on modules/extensions

| Path | Dependency | Action |
|------|------------|--------|
| **internal/extensions/** | All *_test.go | **Delete** with package |
| **internal/startup/** | No _test.go found | N/A |
| **internal/discovery/discovery_test.go** | Extensions/modules, registry | **Delete** with package |
| **internal/packaging/packaging_test.go**, **manifests_test.go** | Extensions/modules | **Delete** with package |
| **internal/registry/registry_test.go** | ItemTypeExtension, ItemTypeAgenticModule | **Delete** with package |
| **internal/homedir/homedir_test.go** | GetAgenticModulesDir, GetExtensionsDir | **Update or trim:** keep tests for GetHomeDir; optionally keep GetExtensionsDir/GetAgenticModulesDir tests if we keep the funcs for compat, or remove tests for removed funcs |
| **internal/proxy/plugin_handler_test.go** | CreateExtensionLLMHandler | **Delete** with plugin_handler.go |
| **cmd/llamagate/** | shutdown_test.go (if any) | Check; likely no extension dependency |

---

## F) Proxy–extensions coupling

| File | Coupling | Action |
|------|----------|--------|
| **internal/proxy/proxy.go** | `responseHookFunc` field, `SetResponseHook()`; call site only when non-nil | **Keep** field and setter; main stops calling SetResponseHook. No nil call. |
| **internal/proxy/plugin_handler.go** | `CreateExtensionLLMHandler()` uses `extensions.LLMHandlerFunc` | **Delete** file (and plugin_handler_test.go); remove `extensions` import from proxy. |

---

## G) Stubbing / no-op requirements

- **main.go:** No extension/module registry, no workflow executor, no route manager, no extension handler, no hook manager. No need for a “no-op” extension system; just remove the code.
- **proxy:** No CreateExtensionLLMHandler call; no SetResponseHook call. Optional: remove `responseHookFunc` and `SetResponseHook` from proxy for cleanliness (or leave for future use).

---

## H) Summary: safe to delete vs needs stubbing

| Category | Safe to delete | Needs stubbing |
|----------|----------------|----------------|
| **Packages** | internal/extensions, internal/startup, internal/discovery, internal/packaging, internal/registry, internal/migration, internal/examples | None |
| **Files in kept packages** | internal/proxy/plugin_handler.go, internal/proxy/plugin_handler_test.go | None |
| **Routes** | All /v1/extensions and dynamic extension routes | None |
| **Config** | ExtensionsUpsertEnabled — deprecate/ignore with warning | None |
| **Root dirs** | extensions/, examples/agenticmodules/ | None |

---

## I) Remaining core after Phase 1

- **Endpoints:** `/health`, `/v1/hardware/recommendations`, `POST /v1/chat/completions`, `GET /v1/models`, and when MCP enabled: `GET/POST /v1/mcp/...`.
- **Config:** OllamaHost, APIKey, RateLimitRPS, Debug, Port, LogFile, Timeout, HealthCheckTimeout, ShutdownTimeout, TLS*, MCP (full block). ExtensionsUpsertEnabled deprecated/ignored.
- **MCP:** Full MCP client, tool manager, guardrails, /v1/mcp/* API — **retained** as core.

This impact map is the source for the subsequent steps (core contract, route removal, loading/registry removal, config cleanup, dead code deletion, docs, tests).
