# LlamaGate Extensions v0.9.1 – Specification

**Version:** 0.9.1  
**Date:** 2026-01-10  
**Status:** Design Lock-In (Pre-Implementation)

---

## 1. What Is an Extension?

An **Extension** in LlamaGate v0.9.1 is a declarative, optional, lifecycle-managed capability module that augments the LlamaGate gateway without modifying core binaries. Extensions extend the system's functionality by:

- **Extending routing** – Adding custom HTTP endpoints beyond the standard API
- **Extending request/response handling** – Processing, transforming, or enriching requests/responses
- **Extending agentic workflows** – Providing reusable workflow definitions that integrate with LLM models and MCP tools
- **Extending tool execution** – Wrapping or composing MCP tools with additional logic
- **Extending system capabilities** – Adding new features like custom authentication, rate limiting, data processing, etc.

### Extension Characteristics

- **Server-side only** – Extensions run entirely on the LlamaGate server
- **Loaded at startup** – Extensions are discovered and loaded during server initialization
- **Static loading** – Extensions are loaded once at startup; hot-reload is **explicitly NOT supported** in v0.9.1
- **User-installable** – Extensions can be added by placing manifest files in the extensions directory; no code compilation required
- **Declarative** – Extensions are primarily defined via YAML manifests with optional executable logic

---

## 2. Extension Lifecycle

### Directory Structure

Extensions are stored in the **`extensions/`** directory (replacing `plugins/`):

```
extensions/
├── manifest.yaml          # Extension manifest (required)
├── config.yaml            # Extension-specific configuration (optional)
└── [executable files]     # Optional executable logic (if needed)
```

### Discovery

Extensions are discovered at **server startup** by:

1. Scanning the `extensions/` directory for `manifest.yaml` files
2. Validating each manifest against the Extension Manifest Schema
3. Loading extension configuration from `config.yaml` (if present) or from main config file under `extensions.configs.<extension_name>`
4. Registering valid extensions in the Extension Registry

**Discovery Process:**
- Extensions are discovered **synchronously** during server initialization
- Invalid extensions are **logged as errors** and **skipped** (server continues to start)
- Discovery failures are **non-fatal** (server starts with available extensions)

### Loading

Extensions are loaded when:
- Server starts up
- Extension manifest is valid
- Extension is **enabled** (see Enable/Disable below)

**Loading Order:**
- Extensions are loaded in **alphabetical order** by directory name
- Dependencies between extensions are **not supported** in v0.9.1
- All extensions load independently

### Enable / Disable

Extensions support **enable/disable** functionality:

- **Enabled by default** – If no explicit enable/disable setting exists, extension is enabled
- **Disable via config** – Set `enabled: false` in extension config or main config file
- **Disable via environment** – Set `EXTENSION_<NAME>_ENABLED=false`
- **When disabled** – Extension is discovered and registered but **not executed**; API endpoints return 503 Service Unavailable
- **Zero side effects when disabled** – Disabled extensions consume minimal resources (registry entry only)

### Hot-Reload

**Hot-reload is explicitly NOT supported in v0.9.1.**

- Extensions cannot be added, removed, or reloaded without server restart
- This is a **design decision** for v0.9.1 to ensure stability and simplicity
- Future versions may add hot-reload capability

---

## 3. Extension Definition Format

### YES – Extensions Use YAML Manifests

Extensions **MUST** be defined using YAML manifest files named `manifest.yaml` in the extension directory.

### Minimum Required Schema

```yaml
name: string                    # Required: Unique extension identifier
version: string                 # Required: Extension version (semver)
description: string             # Required: Human-readable description
enabled: boolean                # Optional: Enable/disable (default: true)
```

### Complete Extension Manifest Schema

```yaml
# Extension Metadata
name: string                    # Required: Unique identifier (alphanumeric + underscore)
version: string                 # Required: Semantic version (e.g., "1.0.0")
description: string             # Required: What this extension does
author: string                  # Optional: Extension author
enabled: boolean                # Optional: Enable/disable (default: true)

# Input/Output Schemas
input_schema:                   # Optional: JSON Schema for inputs
  type: object
  properties: {...}
  required: [...]

output_schema:                  # Optional: JSON Schema for outputs
  type: object
  properties: {...}

# Input Definition
required_inputs:                 # Optional: List of required input parameter names
  - string

optional_inputs:               # Optional: Map of optional inputs with defaults
  key: default_value

# Workflow Definition
workflow:                       # Optional: Agentic workflow definition
  id: string
  name: string
  description: string
  steps:
    - id: string
      type: string              # "llm_call", "tool_call", "data_transform", "condition"
      config: {...}
      dependencies: [...]
      on_error: string          # "stop", "continue", "retry"
  max_retries: integer
  timeout: string               # Duration string (e.g., "30s")

# API Endpoints
endpoints:                      # Optional: Custom HTTP endpoints
  - path: string                # Relative path (e.g., "/custom/action")
    method: string              # HTTP method (GET, POST, PUT, DELETE, etc.)
    description: string
    request_schema: {...}       # Optional: JSON Schema
    response_schema: {...}      # Optional: JSON Schema
    requires_auth: boolean      # Default: true
    requires_rate_limit: boolean # Default: true

# Agent Definition
agent:                          # Optional: Agent definition
  name: string
  description: string
  capabilities: [...]
  config: {...}
```

### What the YAML Represents

The YAML manifest represents:

1. **Metadata** – Name, version, description, author
2. **UI Schema** – Input/output schemas for API documentation and validation
3. **Workflow Steps** – Declarative workflow definition (LLM calls, tool calls, transforms)
4. **Policies** – Enable/disable, authentication, rate limiting
5. **API Contract** – Custom endpoint definitions

### Extension Configuration

Extension-specific configuration can be provided via:

1. **`config.yaml`** in the extension directory
2. **Main config file** under `extensions.configs.<extension_name>`
3. **Environment variables** with format `EXTENSION_<NAME>_<KEY>=value`

Configuration is merged in that order (environment variables override config files).

---

## 4. Adding and Removing Extensions

### Adding an Extension

**User Workflow:**

1. Create extension directory: `extensions/my_extension/`
2. Create `manifest.yaml` with required fields (name, version, description)
3. Optionally add `config.yaml` for extension-specific configuration
4. Restart LlamaGate server
5. Extension is automatically discovered and registered

**What Happens:**
- Server scans `extensions/` directory at startup
- Manifest is validated
- Extension is registered in Extension Registry
- If enabled, extension is available via API endpoints
- Success is logged: `"Extension 'my_extension' registered successfully"`

### Removing or Disabling an Extension

**Remove Extension:**
1. Delete extension directory: `rm -rf extensions/my_extension/`
2. Restart LlamaGate server
3. Extension is no longer available

**Disable Extension (without removal):**
1. Set `enabled: false` in extension config or main config
2. Restart server (or use future hot-reload if added)
3. Extension remains registered but returns 503 when accessed

### Failure Handling

**If Extension Fails to Load:**
- Error is **logged** with extension name and error details
- Extension is **skipped** (not registered)
- Server **continues to start** with other extensions
- Error format: `"Failed to load extension 'my_extension': <error>"`

**What Is Logged:**
- Extension discovery start: `"Discovering extensions in 'extensions/'"`  
- Extension loaded: `"Extension 'my_extension' (v1.0.0) loaded successfully"`  
- Extension failed: `"Failed to load extension 'my_extension': <error>"`  
- Extension disabled: `"Extension 'my_extension' is disabled"`  
- Extension count: `"Loaded N extensions"`

---

## 5. How Extensions Are Executed

### Invocation Methods

Extensions can be invoked via:

1. **API Endpoint** – `POST /v1/extensions/:name/execute`
   - Standard execution endpoint
   - Accepts JSON input in request body
   - Returns structured result

2. **Custom API Endpoints** – `[METHOD] /v1/extensions/:name/<custom_path>`
   - Defined in manifest `endpoints` section
   - Custom handlers (if executable logic provided)
   - Standard REST endpoints

3. **Internal Workflow Trigger** – Via agentic workflows
   - Extensions can be invoked as workflow steps
   - Integration with LLM tool calling (if extension exposes tools)

4. **MCP Tool Call** – Extensions can expose MCP-compatible tools
   - Tools are registered with MCP tool manager
   - Invoked via standard MCP tool execution flow

### Execution Flow

```
1. Request received (API endpoint or workflow trigger)
   ↓
2. Extension Registry lookup by name
   ↓
3. Input validation (against input_schema)
   ↓
4. Extension execution:
   - If workflow defined: Execute workflow steps
   - If custom handler: Execute handler function
   - If tool: Execute via MCP tool execution
   ↓
5. Result formatting (against output_schema)
   ↓
6. Response returned
```

### Extension Chaining

**Extensions CAN be chained together** in workflows:

- Workflow step can invoke another extension via `tool_call` type with `tool_name: "extension.<extension_name>"`
- Extensions can call other extensions programmatically (if executable logic provided)
- No circular dependency detection in v0.9.1 (extensions must avoid cycles)

### Model Access

Extensions can call models:

- **Only through LlamaGate core abstractions** – Via `ExtensionContext.LLMHandler`
- Extensions **cannot** call models directly
- All LLM calls go through LlamaGate proxy layer
- Supports caching, rate limiting, and other core features

---

## 6. Extension Boundaries & Safety

### What Extensions Are NOT Allowed to Do

Extensions have **hard boundaries**:

1. **Cannot modify core binaries** – Extensions are isolated from core code
2. **Cannot register global HTTP endpoints** – Only endpoints under `/v1/extensions/<name>/` are allowed
3. **Cannot access internal LlamaGate state** – Only via provided ExtensionContext
4. **Cannot bypass authentication/rate limiting** – All endpoints respect middleware
5. **Cannot modify other extensions** – Extensions are isolated from each other

### Network Calls

**Extensions CAN make outbound network calls** via:
- `ExtensionContext.HTTPClient` (provided HTTP client with timeout)
- Network calls are **allowed but monitored** (logged with extension name)
- No network restrictions in v0.9.1 (future versions may add allow/deny lists)

### File System Access

**Extensions CAN read/write files** via:
- Standard Go file I/O (if executable logic provided)
- **No explicit restrictions** in v0.9.1
- **Best practice**: Extensions should only access files in their own directory or explicitly configured paths
- Future versions may add sandboxing

### HTTP Endpoint Registration

**Extensions CAN register HTTP endpoints** but:
- Only under `/v1/extensions/<extension_name>/`
- Endpoints are automatically prefixed
- Endpoints respect authentication and rate limiting middleware
- Custom endpoints are registered at server startup

### Policy and Guardrail Controls

**Current State (v0.9.1):**
- **No explicit policy engine** – Extensions are trusted
- **No guardrails** – Extensions have full access to ExtensionContext capabilities
- **Logging** – All extension operations are logged with extension name for audit

**Future Considerations:**
- Sandboxing for untrusted extensions
- Policy engine for fine-grained permissions
- Resource quotas (CPU, memory, network)
- Allow/deny lists for network calls

---

## 7. Backward Compatibility

### Legacy Plugin Support

**NO – Legacy "plugin" support is NOT provided in v0.9.1.**

This is a **breaking change**. The migration is **mandatory**.

### Migration Steps Required

**Exact migration steps:**

1. **Rename directory**: `plugins/` → `extensions/`
2. **Convert Go plugins to YAML manifests**:
   - Extract metadata from Go `Metadata()` method → `manifest.yaml`
   - Extract input/output schemas → `manifest.yaml`
   - Convert workflow definitions → `manifest.yaml` workflow section
   - If custom endpoints exist, convert to `endpoints` section
3. **Update configuration**:
   - `plugins.configs` → `extensions.configs`
   - `PLUGINS_ENABLED` → `EXTENSIONS_ENABLED`
   - `PLUGIN_<NAME>_<KEY>` → `EXTENSION_<NAME>_<KEY>`
4. **Update API endpoints**:
   - `/v1/plugins` → `/v1/extensions`
   - `/v1/plugins/:name/execute` → `/v1/extensions/:name/execute`
5. **Update code references**:
   - All Go code: `plugin` → `extension`
   - All package names: `plugins` → `extensions`
   - All type names: `Plugin` → `Extension`
6. **Update documentation**:
   - All docs: "plugin" → "extension"
   - All examples: Update API paths and terminology

### Deprecation Communication

Since this is v0.9.1 (pre-1.0), deprecation communication:

- **CHANGELOG.md** – Document breaking change clearly
- **Migration Guide** – Create `docs/MIGRATION_V0.9.1.md` with step-by-step instructions
- **Release Notes** – Emphasize breaking change in release announcement
- **No grace period** – v0.9.1 removes plugin support immediately

---

## 8. Versioning & Contract Stability

### Extension Schema Stability

The extension schema is **STABLE** for v0.9.1.

- Schema changes will require a **major version bump** (v1.0.0+)
- Minor version bumps (v0.9.2, v0.9.3) will **not** change the manifest schema
- Patch versions (v0.9.1.1) are for bug fixes only

### Guarantees to Extension Authors

**LlamaGate v0.9.1 guarantees:**

1. **Manifest schema stability** – YAML schema will not change within v0.9.1.x
2. **API endpoint stability** – `/v1/extensions/*` endpoints will not change
3. **ExtensionContext API stability** – Context methods will not change
4. **Workflow step types** – Core step types (`llm_call`, `tool_call`, `data_transform`, `condition`) are stable

**No guarantees for:**

- Internal implementation details
- Extension discovery/loading order
- Performance characteristics
- Resource limits

### Changes Allowed Without Major Version Bump

**Within v0.9.1.x (minor/patch versions):**

- Bug fixes
- Performance improvements
- Additional optional manifest fields (backward compatible)
- Additional ExtensionContext methods (additive only)
- Additional workflow step types (additive only)
- Additional configuration options

**Requires major version bump (v1.0.0+):**

- Breaking manifest schema changes
- Removing manifest fields
- Changing required fields
- Breaking API endpoint changes
- Removing ExtensionContext methods
- Removing workflow step types

---

## Design Rationale: Why Extensions (Not Plugins)

### Industry Alignment

The term **"extension"** aligns with modern industry standards:

- **VS Code Extensions** – Declarative, manifest-based, lifecycle-managed
- **Browser Extensions** – YAML/JSON manifests, isolated, composable
- **GitHub Apps** – Configuration-first, optional, additive
- **AI Tool Extensions** – Model-friendly, declarative, safe

### Semantic Clarity

**"Extension"** implies:
- ✅ **Additive** – Extends without modifying core
- ✅ **Optional** – System works without extensions
- ✅ **Declarative** – Configuration-first design
- ✅ **Lifecycle-aware** – Install, enable, disable, remove

**"Plugin"** historically implies:
- ❌ **Invasive** – Modifies core behavior
- ❌ **Tightly coupled** – Hard to remove
- ❌ **Code-first** – Requires compilation
- ❌ **Legacy** – Older architecture patterns

### Technical Benefits

1. **YAML-First Design** – Extensions are model-friendly, easy to generate programmatically
2. **No Compilation Required** – Users can add extensions without rebuilding LlamaGate
3. **Clear Lifecycle** – Enable/disable, install/remove are explicit operations
4. **Better Isolation** – Extensions are clearly separated from core
5. **Future-Proof** – Foundation for sandboxing, policies, and advanced features

---

## Assumptions That Must Be True Before Implementation

### 1. Directory Structure

**Assumption:** The `extensions/` directory exists and is writable by the LlamaGate process.

**Validation Required:**
- Check if `extensions/` directory exists at startup
- Create directory if it doesn't exist (with appropriate permissions)
- Verify directory is readable

### 2. YAML Parser Availability

**Assumption:** Go YAML parsing library is available and stable.

**Validation Required:**
- Confirm YAML library choice (e.g., `gopkg.in/yaml.v3`)
- Test YAML parsing with edge cases (invalid YAML, missing fields, etc.)
- Handle YAML parsing errors gracefully

### 3. Manifest Schema Validation

**Assumption:** Manifest validation can be performed reliably.

**Validation Required:**
- Define strict schema validation rules
- Test with valid and invalid manifests
- Ensure validation errors are clear and actionable

### 4. Extension Isolation

**Assumption:** Extensions can be isolated from each other and from core.

**Validation Required:**
- Verify ExtensionContext provides only intended capabilities
- Test that extensions cannot access internal state
- Ensure extension failures don't crash server

### 5. Configuration Merging

**Assumption:** Configuration from multiple sources can be merged correctly.

**Validation Required:**
- Test config file + environment variable merging
- Verify precedence order (env > config file > defaults)
- Handle conflicting configuration values

### 6. API Endpoint Registration

**Assumption:** Custom extension endpoints can be registered dynamically.

**Validation Required:**
- Test endpoint registration at startup
- Verify path prefixing (`/v1/extensions/<name>/`)
- Ensure middleware (auth, rate limit) applies correctly

### 7. Workflow Execution

**Assumption:** YAML-defined workflows can be executed reliably.

**Validation Required:**
- Test workflow step execution
- Verify LLM and tool integration
- Handle workflow errors gracefully

### 8. Backward Compatibility Removal

**Assumption:** All plugin references can be removed without breaking core functionality.

**Validation Required:**
- Audit all code for "plugin" references
- Verify no hidden dependencies on plugin system
- Test that core functionality works without extensions

### 9. Performance

**Assumption:** Extension discovery and loading don't significantly impact startup time.

**Validation Required:**
- Measure startup time with 0, 10, 50, 100 extensions
- Optimize discovery process if needed
- Consider parallel loading if startup time is unacceptable

### 10. Error Handling

**Assumption:** Extension failures are handled gracefully without affecting server operation.

**Validation Required:**
- Test invalid manifests
- Test extension execution failures
- Verify server continues operating with failed extensions
- Ensure errors are logged clearly

---

## Summary

This specification defines **Extensions v0.9.1** as a complete replacement for the plugin system, with:

- ✅ **YAML manifest-based** definitions
- ✅ **Startup discovery** and registration
- ✅ **Enable/disable** capability
- ✅ **No hot-reload** (explicitly not supported)
- ✅ **Clear boundaries** and safety model
- ✅ **No backward compatibility** (breaking change)
- ✅ **Stable schema** for extension authors

The specification is **locked** and ready for implementation.

---

**Next Steps:**
1. Review and approve this specification
2. Create implementation plan
3. Begin code migration from plugins to extensions
4. Update all documentation
5. Create migration guide for users
