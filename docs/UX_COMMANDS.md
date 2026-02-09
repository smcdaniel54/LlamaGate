# UX Commands Reference

**Version:** 0.9.1+  
**Status:** **Obsolete** — Extensions and AgenticModules were removed in Phase 1. This document is retained for historical reference only. See [Core Contract](core_contract.md) for current API.

---

## Overview

This document described the recommended UX commands and endpoints for managing extensions and AgenticModules in LlamaGate. **The extension system and `llamagate-cli` have been removed.** The content below is historical only.

---

## CLI Commands

### Extension Commands

#### List Extensions

```bash
go run ./cmd/llamagate-cli ext list
```

**Output:**
```
Found 3 extension(s):

  prompt-template-executor (v1.0.0) [workflow] - Execute approved prompt templates
    Status: enabled
  intake_structured_summary (v1.0.0) [workflow] - Generate structured summary
    Status: enabled
  urgency_router (v1.0.0) [workflow] - Route items based on urgency
    Status: enabled
```

#### Show Extension Details

```bash
go run ./cmd/llamagate-cli ext show intake_structured_summary
```

**Output (JSON):**
```json
{
  "name": "intake_structured_summary",
  "version": "1.0.0",
  "description": "Generate structured summary from input text using LLM",
  "type": "workflow",
  "enabled": true,
  "inputs": [
    {
      "id": "input_text",
      "type": "string",
      "required": true
    }
  ],
  "outputs": [
    {
      "id": "summary",
      "type": "object"
    }
  ],
  "steps": [...]
}
```

#### Validate Extension

```bash
go run ./cmd/llamagate-cli ext validate extensions/intake_structured_summary/manifest.yaml
```

**Output:**
```
✓ Extension 'intake_structured_summary' (v1.0.0) is valid
  Type: workflow
  Description: Generate structured summary from input text using LLM
  Inputs: 2
  Outputs: 1
  Steps: 3
```

#### Run Extension (Helper)

```bash
go run ./cmd/llamagate-cli ext run intake_structured_summary
```

**Output:**
```
Note: 'ext run' requires LlamaGate server to be running.
Use the HTTP API instead:
  curl -X POST http://localhost:11435/v1/extensions/intake_structured_summary/execute \
    -H "Content-Type: application/json" \
    -H "X-API-Key: sk-llamagate" \
    -d '{"input": {...}}'
```

---

## HTTP API Endpoints

### Extension Endpoints

#### List Extensions

```bash
GET /v1/extensions
```

**Request:**
```bash
curl -H "X-API-Key: sk-llamagate" http://localhost:11435/v1/extensions
```

**Response:**
```json
{
  "extensions": [
    {
      "name": "intake_structured_summary",
      "version": "1.0.0",
      "description": "Generate structured summary",
      "type": "workflow",
      "enabled": true
    }
  ],
  "count": 1
}
```

#### Get Extension Details

```bash
GET /v1/extensions/:name
```

**Request:**
```bash
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/extensions/intake_structured_summary
```

**Response:**
```json
{
  "name": "intake_structured_summary",
  "version": "1.0.0",
  "description": "Generate structured summary",
  "type": "workflow",
  "enabled": true,
  "inputs": [...],
  "outputs": [...]
}
```

#### Execute Extension

```bash
POST /v1/extensions/:name/execute
```

**Request:**
```bash
curl -X POST http://localhost:11435/v1/extensions/intake_structured_summary/execute \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "input_text": "Customer reported critical issue...",
    "model": "mistral"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "summary": {
      "title": "Critical Issue",
      "urgency_level": "high"
    }
  }
}
```

---

## Module Commands

### Run Module (via Module Runner)

```bash
POST /v1/extensions/agenticmodule_runner/execute
```

**Request:**
```bash
curl -X POST http://localhost:11435/v1/extensions/agenticmodule_runner/execute \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "module_name": "intake_and_routing",
    "module_input": {
      "input_text": "Customer reported critical issue...",
      "model": "mistral"
    },
    "max_runtime_seconds": 300,
    "max_steps": 10
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "run_record": {
      "module_name": "intake_and_routing",
      "module_version": "1.0.0",
      "trace_id": "...",
      "started_at": "2026-01-15T10:00:00Z",
      "completed_at": "2026-01-15T10:00:15Z",
      "total_duration_ms": 15000,
      "steps": [...],
      "final_output": {...}
    }
  }
}
```

---

## Command-Line Options

### CLI Global Flags

```bash
-extensions-dir string    Directory containing extensions (default: "extensions")
```

**Example:**
```bash
go run ./cmd/llamagate-cli -extensions-dir custom/extensions ext list
```

---

## Integration with External Tools

### OrchestratorPlus Integration

External tools (like OrchestratorPlus) can use these endpoints to:

1. **Discover installed extensions:**
   ```bash
   GET /v1/extensions
   ```

2. **Inspect extension capabilities:**
   ```bash
   GET /v1/extensions/:name
   ```

3. **Execute extensions:**
   ```bash
   POST /v1/extensions/:name/execute
   ```

4. **Run modules:**
   ```bash
   POST /v1/extensions/agenticmodule_runner/execute
   ```

5. **Refresh extensions (after installation):**
   ```bash
   POST /v1/extensions/refresh
   ```

### Example: Python Integration

```python
import requests

base_url = "http://localhost:11435/v1"
api_key = "sk-llamagate"
headers = {"X-API-Key": api_key}

# List extensions
response = requests.get(f"{base_url}/extensions", headers=headers)
extensions = response.json()["extensions"]

# Get extension details
response = requests.get(f"{base_url}/extensions/intake_structured_summary", headers=headers)
extension_info = response.json()

# Execute extension
response = requests.post(
    f"{base_url}/extensions/intake_structured_summary/execute",
    headers=headers,
    json={"input_text": "Test input", "model": "mistral"}
)
result = response.json()
```

---

## Error Responses

### Extension Not Found

```json
{
  "error": {
    "message": "Extension not found",
    "type": "not_found_error",
    "request_id": "..."
  }
}
```

### Validation Error

```json
{
  "error": {
    "message": "required input 'input_text' is missing",
    "type": "invalid_request_error",
    "request_id": "..."
  }
}
```

### Execution Error

```json
{
  "error": {
    "message": "Extension execution failed: step 0 (llm.chat) failed: LLM call failed",
    "type": "internal_error",
    "request_id": "..."
  }
}
```

---

## Quick Reference

| Command | Purpose | Endpoint/CLI |
|---------|---------|--------------|
| List extensions | Show all registered extensions | `GET /v1/extensions` or `ext list` |
| Show extension | Get extension details | `GET /v1/extensions/:name` or `ext show <name>` |
| Validate extension | Validate manifest YAML | `ext validate <path>` |
| Upsert extension | Create/update manifest (enabled by default; set `EXTENSIONS_UPSERT_ENABLED=false` to disable) | `PUT /v1/extensions/:name` |
| Execute extension | Run a workflow extension | `POST /v1/extensions/:name/execute` |
| Run module | Execute an AgenticModule | `POST /v1/extensions/agenticmodule_runner/execute` |
| Refresh extensions | Re-discover extensions after installation | `POST /v1/extensions/refresh` |

---

## Next Steps

- **Extension Quick Start:** [EXTENSIONS_QUICKSTART.md](EXTENSIONS_QUICKSTART.md)
- **AgenticModules Guide:** [AGENTICMODULES.md](AGENTICMODULES.md)
- **Extension Specification:** [EXTENSIONS_SPEC_V0.9.1.md](EXTENSIONS_SPEC_V0.9.1.md)
- **Example Extensions:** [LlamaGate Extension Examples Repository](https://github.com/smcdaniel54/LlamaGate-extension-examples) - Copy/paste-ready examples and templates

---

**For more details, see the [Extension Specification](EXTENSIONS_SPEC_V0.9.1.md).**
