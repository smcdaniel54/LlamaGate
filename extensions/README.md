# LlamaGate Extensions

This directory contains example extensions for LlamaGate v0.9.1. Extensions are declarative, YAML-based modules that extend LlamaGate's capabilities.

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

## Quick Start

Extensions are automatically discovered at server startup. Simply place extension directories in this folder, each containing a `manifest.yaml` file.

## Builtin Extensions

### Extension Documentation Generator

**Purpose:** Generate comprehensive markdown documentation for extensions and modules

**Type:** Builtin Extension (YAML-based)

**Location:** `builtin/extension-doc-generator/`

**Usage:**
```bash
curl -X POST http://localhost:11435/v1/extensions/extension-doc-generator/execute \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "input": {
      "target": "prompt-template-executor",
      "output_path": "docs/extensions/prompt-template-executor.md"
    }
  }'
```

**Inputs:**
- `target` (required): Extension or module name
- `output_path` (optional): Where to save documentation (default: `docs/extensions/{target}.md`)
- `format` (optional): markdown, html, or json
- `include_examples` (optional): Include usage examples
- `include_api_details` (optional): Include API endpoint details

**Outputs:**
- `documentation`: Generated markdown documentation
- `file_path`: Path where documentation was saved

---

## Example Extensions

### 1. Prompt Template Executor

**Purpose:** Execute approved prompt templates with structured inputs and produce deterministic output artifacts.

**Type:** Workflow Extension

**Location:** `prompt-template-executor/`

**Features:**
- Load prompt templates from files
- Render templates with structured variables
- Execute LLM calls via LlamaGate
- Write results to output files

**Usage:**

```bash
# Execute the extension
curl -X POST http://localhost:11435/v1/extensions/prompt-template-executor/execute \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "template_id": "example",
    "variables": {
      "document_type": "executive summary",
      "format": "markdown"
    },
    "model": "llama3.2"
  }'
```

**Inputs:**
- `template_id` (required): Name of template file (without .txt extension)
- `variables` (required): Object containing template variables
- `model` (optional): LLM model to use (default: "llama3.2")

**Outputs:**
- `result.md`: Generated content written to `output/result.md`

**Template Location:**
- Templates are stored in `templates/` subdirectory
- Template files use Go template syntax
- Example: `templates/example.txt`

---

### 2. Request Inspector

**Purpose:** Intercept requests and responses flowing through LlamaGate and generate redacted audit records.

**Type:** Middleware Extension

**Location:** `request-inspector/`

**Features:**
- Automatic request interception
- Configurable redaction rules
- JSONL audit log generation
- Path-based filtering

**Configuration:**

The extension is configured via `manifest.yaml`:

```yaml
config:
  enabled: true
  sample_rate: 1.0
  audit_dir: ./var/audit
  redact:
    - path: $.messages[*].content
      mode: truncate
      max_len: 120
```

**Audit Log Format:**

Audit logs are written to `var/audit/audit-YYYY-MM-DD.jsonl`:

```json
{
  "timestamp": "2026-01-10T12:00:00Z",
  "method": "POST",
  "path": "/v1/chat/completions",
  "request_id": "abc-123",
  "ip": "127.0.0.1",
  "redacted": true,
  "max_length": 120
}
```

**Automatic Operation:**
- No API calls needed
- Automatically intercepts all `/v1/*` requests
- Runs as middleware on every request

---

### 3. Cost Usage Reporter

**Purpose:** Track token usage and estimated cost per request, producing machine-readable usage reports.

**Type:** Observer Extension

**Location:** `cost-usage-reporter/`

**Features:**
- Automatic usage tracking from LLM responses
- Token count extraction
- JSON report generation
- Report accumulation over time

**Usage Report Format:**

Reports are written to `output/usage_report.json`:

```json
[
  {
    "timestamp": "2026-01-10T12:00:00Z",
    "request_id": "abc-123",
    "model": "llama3.2",
    "prompt_tokens": 100,
    "completion_tokens": 200,
    "total_tokens": 300,
    "estimated_cost": 0.0
  }
]
```

**Automatic Operation:**
- No API calls needed
- Automatically tracks all LLM responses
- Accumulates usage data in JSON array

---

## Extension API Endpoints

### List All Extensions

```bash
GET /v1/extensions
```

**Response:**
```json
{
  "extensions": [
    {
      "name": "prompt-template-executor",
      "version": "1.0.0",
      "description": "Execute approved prompt templates...",
      "type": "workflow",
      "enabled": true
    }
  ],
  "count": 3
}
```

### Get Extension Details

```bash
GET /v1/extensions/:name
```

**Response:**
```json
{
  "name": "prompt-template-executor",
  "version": "1.0.0",
  "description": "Execute approved prompt templates...",
  "type": "workflow",
  "enabled": true,
  "inputs": [...],
  "outputs": [...]
}
```

### Upsert Extension (Optional)

Create or update an extension manifest. **Enabled by default**; set `EXTENSIONS_UPSERT_ENABLED=false` to lock down. Writes to `~/.llamagate/extensions/installed/:name/manifest.yaml`. After upsert, call `POST /v1/extensions/refresh` to load the extension.

```bash
PUT /v1/extensions/:name
```

**Request Body:** YAML or JSON manifest (same schema as `manifest.yaml`). The path `:name` overrides `name` in the body.

**When disabled** (`EXTENSIONS_UPSERT_ENABLED=false`): Returns `501 Not Implemented` with `{"code": "UPSERT_NOT_CONFIGURED", "error": "Workflow upsert is not enabled"}`.

See [API.md](docs/API.md) for full request/response and examples.

### Execute Workflow Extension

```bash
POST /v1/extensions/:name/execute
```

**Request Body:**
```json
{
  "template_id": "example",
  "variables": {
    "key": "value"
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "output_file": "/path/to/output.md"
  }
}
```

---

## Extension Structure

Each extension directory should contain:

```
extension-name/
├── manifest.yaml          # Required: Extension definition
├── config.yaml            # Optional: Extension configuration
├── templates/            # Optional: Template files (for workflow extensions)
├── output/                # Optional: Output directory
└── [other files]          # Optional: Additional resources
```

## Enabling/Disabling Extensions

### Via Manifest

```yaml
enabled: false  # Disable extension
```

### Via Environment Variable

```bash
EXTENSION_request-inspector_ENABLED=false
```

### Via Config File

```yaml
extensions:
  configs:
    request-inspector:
      enabled: false
```

---

## Creating Your Own Extension

1. **Create extension directory:**
   ```bash
   mkdir extensions/my-extension
   ```

2. **Create manifest.yaml:**
   ```yaml
   name: my-extension
   version: 1.0.0
   description: My custom extension
   type: workflow
   enabled: true
   
   inputs:
     - id: input1
       type: string
       required: true
   
   steps:
     - uses: llm.chat
   ```

3. **Restart LlamaGate** - Extension will be automatically discovered

See [Extension Specification](../../docs/EXTENSIONS_SPEC_V0.9.1.md) for complete manifest schema.

---

## Troubleshooting

### Extension Not Discovered

- Check that `manifest.yaml` exists in extension directory
- Verify manifest YAML is valid
- Check server logs for discovery errors

### Extension Execution Fails

- Verify all required inputs are provided
- Check that extension is enabled
- Review server logs for detailed error messages

### Output Files Not Created

- Verify output directory exists and is writable
- Check file paths in manifest are correct
- Ensure extension has write permissions

---

## More Information

- [Extension Specification](../../docs/EXTENSIONS_SPEC_V0.9.1.md) - Complete specification
- [Extension Testing](../../docs/EXTENSIONS_TESTING.md) - Test documentation
- [API Documentation](../../docs/API.md) - API endpoint details

---

*Last Updated: 2026-01-10*
