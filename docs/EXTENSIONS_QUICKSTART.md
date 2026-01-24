# Extensions Quick Start Guide

**Version:** 0.9.1+  
**Status:** Ready for Use

---

## Overview

This guide shows you how to write, validate, load, and run extensions in LlamaGate. Extensions are YAML-based modules that extend LlamaGate functionality.

---

## Prerequisites

- LlamaGate server installed and running
- `extensions/` directory exists (created automatically)
- Basic understanding of YAML

---

## Quick Start: Write Your First Extension

### Step 1: Create Extension Directory

```bash
mkdir -p extensions/my-first-extension
```

### Step 2: Create Manifest

Create `extensions/my-first-extension/manifest.yaml`:

```yaml
name: my-first-extension
version: 1.0.0
description: My first LlamaGate extension
type: workflow
enabled: true

inputs:
  - id: message
    type: string
    required: true

outputs:
  - id: response
    type: object

steps:
  - uses: llm.chat
    with:
      model: "mistral"
      prompt: "Echo back: {{message}}"
  - uses: summary.parse
```

### Step 3: Validate Extension

Use the CLI to validate:

```bash
go run ./cmd/llamagate-cli ext validate extensions/my-first-extension/manifest.yaml
```

Or use HTTP API:

```bash
# Option 1: Refresh extensions (no restart needed)
curl -X POST \
  -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/extensions/refresh

# Option 2: Restart LlamaGate to load the extension
# Then check if it's loaded:
curl -H "X-API-Key: sk-llamagate" http://localhost:11435/v1/extensions
```

### Step 4: Run Extension

```bash
curl -X POST http://localhost:11435/v1/extensions/my-first-extension/execute \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "message": "Hello, world!"
  }'
```

---

## Extension Types

LlamaGate has three types of extensions:

### Builtin Extensions (Go Code)
- Core functionality compiled into binary
- Location: `internal/extensions/builtin/`
- Examples: `validation`, `tools`, `state`, `human`, `events`

### Builtin Extensions (YAML-based)
- Core workflow capabilities included in repo
- Location: `extensions/builtin/`
- Manifest flag: `builtin: true`
- Examples: `extension-doc-generator`
- Loaded with priority, can't be disabled

### Default Extensions (YAML-based)
- Workflow extensions included in repo
- Location: `extensions/` (not in `builtin/` subdirectory)
- Examples: `agenticmodule_runner`, `prompt-template-executor`
- Discovered at startup

---

## Builtin Extensions

### Extension Documentation Generator

**Purpose:** Generate comprehensive markdown documentation for extensions and modules

**Type:** Builtin Extension (YAML-based)

**Location:** `extensions/builtin/extension-doc-generator/`

**Usage:**
```bash
curl -X POST http://localhost:11435/v1/extensions/extension-doc-generator/execute \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "target": "prompt-template-executor",
    "output_path": "docs/extensions/prompt-template-executor.md"
  }'
```

**Inputs:**
- `target` (required): Extension or module name
- `output_path` (optional): Where to save documentation (default: `docs/extensions/{target}.md`)
- `format` (optional): markdown, html, or json
- `include_examples` (optional): Include usage examples
- `include_api_details` (optional): Include API endpoint details

---

## Extension Types (by Functionality)

### 1. Workflow Extensions

Execute sequences of steps (LLM calls, file operations, etc.)

**Example:**
```yaml
name: text-processor
version: 1.0.0
description: Process text through LLM
type: workflow
enabled: true

steps:
  - uses: llm.chat
    with:
      model: "mistral"
      prompt: "Summarize: {{input_text}}"
  - uses: summary.parse
```

### 2. Middleware Extensions

Intercept HTTP requests before processing

**Example:**
```yaml
name: request-logger
version: 1.0.0
description: Log all requests
type: middleware
enabled: true

hooks:
  - on: http.request
    action: log_request
```

### 3. Observer Extensions

Monitor LLM responses after processing

**Example:**
```yaml
name: cost-tracker
version: 1.0.0
description: Track token usage
type: observer
enabled: true

hooks:
  - on: llm.response
    action: track_usage
```

---

## Available Step Types

### Built-in Steps

- `template.load` - Load template from file
- `template.render` - Render template with variables
- `llm.chat` - Call LLM with prompt
- `file.write` - Write output to file
- `extension.call` - Call another extension
- `summary.parse` - Parse LLM response as JSON
- `rules.evaluate` - Evaluate if-then rules

### Module Runner Steps

- `module.load` - Load AgenticModule manifest
- `module.validate` - Validate module and extensions
- `module.execute` - Execute module workflow
- `module.record` - Create execution record

---

## CLI Commands

### List Extensions

```bash
go run ./cmd/llamagate-cli ext list
```

### Show Extension Details

```bash
go run ./cmd/llamagate-cli ext show my-first-extension
```

### Validate Extension

```bash
go run ./cmd/llamagate-cli ext validate extensions/my-first-extension/manifest.yaml
```

### Run Extension (via API)

```bash
# CLI shows how to use HTTP API
go run ./cmd/llamagate-cli ext run my-first-extension
```

---

## Where Logs Are

### Extension Execution Logs

Logs appear in:
- **Console output** (if `LOG_FILE` not set)
- **Log file** (if `LOG_FILE` is set in `.env`)

### Log Format

```json
{
  "level": "info",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "extension": "my-first-extension",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "call_depth": 0,
  "message": "Executing extension-to-extension call"
}
```

### Finding Extension Logs

```bash
# If using log file
grep "extension" llamagate.log

# Or filter by extension name
grep "my-first-extension" llamagate.log
```

---

## Common Failure Modes

### 1. Extension Not Found

**Error:** `extension 'my-extension' not found`

**Causes:**
- Extension not in `extensions/` directory
- `manifest.yaml` file missing
- Extension name mismatch

**Fix:**
- Verify extension directory exists: `ls extensions/my-extension/`
- Check manifest file: `cat extensions/my-extension/manifest.yaml`
- Refresh extensions via API (no restart needed):
  ```bash
  curl -X POST \
    -H "X-API-Key: sk-llamagate" \
    http://localhost:11435/v1/extensions/refresh
  ```
- Or restart LlamaGate to reload extensions

### 2. Validation Errors

**Error:** `validation error: 'name' field is required`

**Causes:**
- Missing required fields in manifest
- Invalid field values
- YAML syntax errors

**Fix:**
- Run validation: `llamagate-cli ext validate extensions/my-extension/manifest.yaml`
- Check manifest against [Extension Specification](EXTENSIONS_SPEC_V0.9.1.md)

### 3. Extension Disabled

**Error:** `Extension is disabled`

**Causes:**
- `enabled: false` in manifest
- Extension disabled via environment variable

**Fix:**
- Set `enabled: true` in manifest
- Check environment variables

### 4. Missing Required Inputs

**Error:** `required input 'message' is missing`

**Causes:**
- Input not provided in API request
- Input ID mismatch

**Fix:**
- Check extension inputs: `llamagate-cli ext show my-extension`
- Provide all required inputs in request body

### 5. Extension-to-Extension Call Failures

**Error:** `maximum call depth exceeded`

**Causes:**
- Recursive extension calls
- Too many nested calls

**Fix:**
- Check call depth in logs
- Reduce nesting in extension workflows
- Default max depth is 10

### 6. Module Execution Failures

**Error:** `module step 0 (intake_structured_summary) failed`

**Causes:**
- Referenced extension not found
- Extension execution error
- Module manifest validation failed

**Fix:**
- Verify all referenced extensions exist: `llamagate-cli ext list`
- Check extension logs for details
- Validate module manifest

---

## Next Steps

- **Learn more:** [Extension Specification](EXTENSIONS_SPEC_V0.9.1.md)
- **Create modules:** [AgenticModules Guide](AGENTICMODULES.md)
- **CLI reference:** [UX Commands](UX_COMMANDS.md)
- **Example Extensions:** [LlamaGate Extension Examples Repository](https://github.com/smcdaniel54/LlamaGate-extension-examples) - High-value, copy/paste-ready examples
- **In-Repo Examples:** See `extensions/` directory and `examples/agenticmodules/`

---

## Troubleshooting

### Extension Not Loading

1. Check extension directory structure:
   ```bash
   ls -la extensions/my-extension/
   # Should show: manifest.yaml
   ```

2. Validate manifest:
   ```bash
   llamagate-cli ext validate extensions/my-extension/manifest.yaml
   ```

3. Check LlamaGate logs for errors:
   ```bash
   tail -f llamagate.log | grep -i extension
   ```

4. Refresh extensions via API (no restart needed):
   ```bash
   curl -X POST \
     -H "X-API-Key: sk-llamagate" \
     http://localhost:11435/v1/extensions/refresh
   ```
   Or restart LlamaGate to reload extensions

### Extension Execution Failing

1. Check extension is enabled:
   ```bash
   llamagate-cli ext show my-extension
   # Check "enabled" field
   ```

2. Verify inputs match manifest:
   ```bash
   llamagate-cli ext show my-extension
   # Check "inputs" field
   ```

3. Check execution logs:
   ```bash
   grep "my-extension" llamagate.log
   ```

4. Test with minimal input:
   ```bash
   curl -X POST http://localhost:11435/v1/extensions/my-extension/execute \
     -H "Content-Type: application/json" \
     -H "X-API-Key: sk-llamagate" \
     -d '{"required_input": "test"}'
   ```

---

**For more details, see the [Extension Specification](EXTENSIONS_SPEC_V0.9.1.md).**
