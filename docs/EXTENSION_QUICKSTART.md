# Extension Quick Start Guide

**Version:** 0.9.1  
**Status:** Ready for Use

---

## Overview

This guide shows you how to use the three example extensions included with LlamaGate v0.9.1.

---

## Prerequisites

- LlamaGate server running
- Extensions directory exists: `extensions/`
- Three example extensions installed (included by default)

---

## Example Extension 1: Prompt Template Executor

### What It Does

Executes approved prompt templates with structured inputs and produces deterministic output files.

### Quick Test

1. **Check extension is loaded:**
   ```bash
   curl -H "X-API-Key: sk-llamagate" \
     http://localhost:11435/v1/extensions/prompt-template-executor
   ```

2. **Execute the extension:**
   ```bash
   curl -X POST \
     -H "X-API-Key: sk-llamagate" \
     -H "Content-Type: application/json" \
     -d '{
       "template_id": "example",
       "variables": {
         "document_type": "executive summary",
         "format": "markdown"
       },
       "model": "llama3.2"
     }' \
     http://localhost:11435/v1/extensions/prompt-template-executor/execute
   ```

3. **Check output file:**
   ```bash
   cat extensions/prompt-template-executor/output/result.md
   ```

### Expected Output

The extension will:
1. Load template from `extensions/prompt-template-executor/templates/example.txt`
2. Render template with your variables
3. Call LLM with rendered prompt
4. Write result to `extensions/prompt-template-executor/output/result.md`

---

## Example Extension 2: Request Inspector

### What It Does

Automatically intercepts HTTP requests and creates redacted audit logs.

### Quick Test

1. **Check extension is loaded:**
   ```bash
   curl -H "X-API-Key: sk-llamagate" \
     http://localhost:11435/v1/extensions/request-inspector
   ```

2. **Make any API request** (the extension intercepts automatically):
   ```bash
   curl -H "X-API-Key: sk-llamagate" \
     http://localhost:11435/v1/chat/completions \
     -d '{"model": "llama3.2", "messages": [{"role": "user", "content": "Hello"}]}'
   ```

3. **Check audit log:**
   ```bash
   cat extensions/request-inspector/var/audit/audit-$(date +%Y-%m-%d).jsonl
   ```

### Expected Output

The audit log will contain entries like:
```json
{"timestamp":"2026-01-10T12:00:00Z","method":"POST","path":"/v1/chat/completions","request_id":"abc-123","ip":"127.0.0.1","redacted":true,"max_length":120}
```

**Note:** This extension runs automatically - no API call needed to execute it.

---

## Example Extension 3: Cost Usage Reporter

### What It Does

Tracks token usage and estimated cost from LLM responses.

### Quick Test

1. **Check extension is loaded:**
   ```bash
   curl -H "X-API-Key: sk-llamagate" \
     http://localhost:11435/v1/extensions/cost-usage-reporter
   ```

2. **Make LLM requests** (the extension tracks automatically):
   ```bash
   curl -H "X-API-Key: sk-llamagate" \
     http://localhost:11435/v1/chat/completions \
     -d '{"model": "llama3.2", "messages": [{"role": "user", "content": "Hello"}]}'
   ```

3. **Check usage report:**
   ```bash
   cat extensions/cost-usage-reporter/output/usage_report.json
   ```

### Expected Output

The usage report will contain entries like:
```json
[
  {
    "timestamp": "2026-01-10T12:00:00Z",
    "request_id": "abc-123",
    "model": "llama3.2",
    "prompt_tokens": 10,
    "completion_tokens": 20,
    "total_tokens": 30,
    "estimated_cost": 0.0
  }
]
```

**Note:** This extension runs automatically - no API call needed to execute it.

---

## Listing All Extensions

To see all available extensions:

```bash
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/extensions
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
    },
    {
      "name": "request-inspector",
      "version": "1.0.0",
      "description": "Redacted audit logging...",
      "type": "middleware",
      "enabled": true
    },
    {
      "name": "cost-usage-reporter",
      "version": "1.0.0",
      "description": "Track token usage...",
      "type": "observer",
      "enabled": true
    }
  ],
  "count": 3
}
```

---

## Enabling/Disabling Extensions

### Disable an Extension

**Method 1: Via Environment Variable**
```bash
export EXTENSION_request-inspector_ENABLED=false
# Restart LlamaGate
```

**Method 2: Edit Manifest**
Edit `extensions/request-inspector/manifest.yaml`:
```yaml
enabled: false
```
Restart LlamaGate.

### Verify Extension is Disabled

```bash
curl -H "X-API-Key: sk-llamagate" \
  http://localhost:11435/v1/extensions/request-inspector
```

Response will show `"enabled": false`.

---

## Troubleshooting

### Extension Not Found

**Problem:** `404 Not Found` when accessing extension

**Solutions:**
1. Check extension directory exists: `extensions/<name>/`
2. Verify `manifest.yaml` exists in extension directory
3. Check server logs for discovery errors
4. Restart LlamaGate to re-discover extensions

### Extension Execution Fails

**Problem:** `400 Bad Request` or `500 Internal Server Error`

**Solutions:**
1. Check all required inputs are provided
2. Verify extension is enabled
3. Check server logs for detailed error messages
4. Verify template files exist (for prompt-template-executor)

### No Output Files

**Problem:** Output files not created

**Solutions:**
1. Check output directory exists and is writable
2. Verify file paths in manifest are correct
3. Check server logs for file I/O errors
4. Ensure extension has write permissions

### Middleware/Observer Extensions Not Working

**Problem:** request-inspector or cost-usage-reporter not creating logs

**Solutions:**
1. Verify extension is enabled
2. Check extension type is correct (middleware/observer)
3. Verify hooks are defined in manifest
4. Check server logs for hook execution errors
5. Ensure output directories exist and are writable

---

## Next Steps

- Read [Extension Specification](./EXTENSIONS_SPEC_V0.9.1.md) for complete details
- See [Extensions README](../extensions/README.md) for extension structure
- Review [API Documentation](./API.md) for all endpoints
- Check [Testing Documentation](./EXTENSIONS_TESTING.md) for test examples

---

**Status:** Ready for use ✅

*Last Updated: 2026-01-10*
