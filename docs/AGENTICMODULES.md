# AgenticModules Guide

**Version:** 0.9.1+  
**Status:** Ready for Use

---

## What is an AgenticModule?

An **AgenticModule** is a versioned, exportable bundle of extensions that provides a cohesive capability. It's defined by:

- `agenticmodule.yaml` manifest file
- `extensions/` folder containing extension YAML files
- Optional `tests/` and `docs/` folders

AgenticModules enable you to package related extensions together and execute them as a unified workflow.

---

## Key Concepts

- **Extension**: Atomic unit of capability, defined by YAML and executed by LlamaGate
- **AgenticModule**: Versioned bundle/group of extensions providing a cohesive capability
- **Module Runner**: Built-in extension (`agenticmodule_runner`) that executes modules

---

## Recommended Folder Layout

```
agenticmodules/
  <module-name>/
    agenticmodule.yaml      # Module manifest
    extensions/              # Module-specific extensions (optional)
      extension1.yaml
      extension2.yaml
    tests/                   # Test fixtures (optional)
      test_input.json
      expected_output.json
    docs/                    # Module documentation (optional)
      README.md
```

**Note:** Extensions referenced in a module must be discoverable by LlamaGate. They can be:
- In the main `extensions/` directory
- In the module's `extensions/` subdirectory (if discovered)
- In nested directories (supported by directory-based loading)

---

## Module Manifest Format

### Basic Structure

```yaml
name: my-module
version: 1.0.0
description: Description of what this module does

steps:
  - extension: extension1
    on_error: stop    # or "continue"
  - extension: extension2
    input:            # Optional: specific input for this step
      key: value
    on_error: continue
```

### Step Configuration

Each step in a module references an extension:

```yaml
steps:
  - extension: intake_structured_summary    # Extension name (required)
    input:                                  # Optional: step-specific input
      input_text: "{{module_input.text}}"
    on_error: stop                          # Optional: "stop" (default) or "continue"
```

**Fields:**
- `extension` (required): Name of the extension to execute
- `input` (optional): Step-specific input. If omitted, module input is passed through
- `on_error` (optional): Behavior on error - `"stop"` (default) or `"continue"`

---

## Running Modules

### Using the Module Runner Extension

The `agenticmodule_runner` extension executes AgenticModules:

```bash
curl -X POST http://localhost:11435/v1/extensions/agenticmodule_runner/execute \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "module_name": "intake_and_routing",
    "module_input": {
      "input_text": "Customer reported critical system outage...",
      "model": "mistral"
    },
    "max_runtime_seconds": 300,
    "max_steps": 10
  }'
```

### Response Format

```json
{
  "success": true,
  "data": {
    "run_record": {
      "module_name": "intake_and_routing",
      "module_version": "1.0.0",
      "trace_id": "550e8400-e29b-41d4-a716-446655440000",
      "started_at": "2026-01-15T10:00:00Z",
      "completed_at": "2026-01-15T10:00:15Z",
      "total_duration_ms": 15000,
      "steps": [
        {
          "step_index": 0,
          "extension": "intake_structured_summary",
          "status": "success",
          "duration_ms": 12000,
          "output": {
            "summary": {
              "title": "Critical System Outage",
              "urgency_level": "high"
            }
          }
        },
        {
          "step_index": 1,
          "extension": "urgency_router",
          "status": "success",
          "duration_ms": 5,
          "output": {
            "route": "immediate",
            "priority": 1
          }
        }
      ],
      "final_output": {
        "summary": {...},
        "route": "immediate",
        "priority": 1
      }
    }
  }
}
```

---

## Module Guardrails

The module runner enforces guardrails to prevent runaway execution:

- **Max Runtime**: Default 5 minutes (configurable via `max_runtime_seconds`)
- **Max Steps**: Default 100 steps (configurable via `max_steps`)
- **Call Depth**: Maximum 10 levels of nested extension calls
- **Call Budget**: Tracks remaining extension calls

If limits are exceeded, execution stops with an error.

---

## Example: Intake and Routing Module

See `examples/agenticmodules/intake_and_routing/` for a complete example:

### Module Structure

```
intake_and_routing/
  agenticmodule.yaml
  extensions/
    intake_structured_summary.yaml
    urgency_router.yaml
  tests/
    test_input.json
    expected_output.json
```

### Module Manifest

```yaml
name: intake_and_routing
version: 1.0.0
description: Intake structured data and route based on urgency

steps:
  - extension: intake_structured_summary
    on_error: stop
  - extension: urgency_router
    on_error: stop
```

### Running the Example

```bash
curl -X POST http://localhost:11435/v1/extensions/agenticmodule_runner/execute \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d @examples/agenticmodules/intake_and_routing/tests/test_input.json
```

---

## Packaging and Exporting Modules

### Exporting a Module

1. **Create module directory:**
   ```bash
   mkdir -p my-module
   ```

2. **Include manifest and extensions:**
   ```bash
   cp agenticmodule.yaml my-module/
   cp -r extensions/ my-module/
   ```

3. **Package as archive:**
   ```bash
   tar -czf my-module-v1.0.0.tar.gz my-module/
   ```

### Importing a Module

1. **Extract to agenticmodules directory:**
   ```bash
   tar -xzf my-module-v1.0.0.tar.gz -C agenticmodules/
   ```

2. **Ensure extensions are discoverable:**
   - Copy extensions to main `extensions/` directory, OR
   - Ensure module's `extensions/` directory is in discovery path

3. **Restart LlamaGate** to load the module

---

## Module Validation

### Validate Module Manifest

```bash
# Using CLI (if implemented)
llamagate-cli module validate agenticmodules/my-module/agenticmodule.yaml
```

### Validate Referenced Extensions

The module runner automatically validates that:
- All referenced extensions exist
- All extensions are enabled
- All extensions are workflow type (required for execution)

---

## Best Practices

### 1. Version Your Modules

Always include version in `agenticmodule.yaml`:
```yaml
version: 1.0.0
```

### 2. Document Your Module

Include a `README.md` in the module directory:
```markdown
# My Module

## Description
What this module does

## Usage
How to run it

## Dependencies
Required extensions
```

### 3. Test Your Module

Include test fixtures in `tests/`:
- `test_input.json` - Sample input
- `expected_output.json` - Expected result

### 4. Handle Errors Gracefully

Use `on_error: continue` for non-critical steps:
```yaml
steps:
  - extension: critical_step
    on_error: stop
  - extension: optional_step
    on_error: continue
```

### 5. Keep Modules Focused

Each module should provide one cohesive capability. Split complex workflows into multiple modules.

---

## Module vs. Extension

| Feature | Extension | AgenticModule |
|---------|-----------|---------------|
| **Definition** | Single YAML manifest | Manifest + extensions folder |
| **Execution** | Direct via API | Via module runner extension |
| **Scope** | Single capability | Cohesive workflow |
| **Versioning** | Per extension | Per module |
| **Packaging** | Single file | Directory structure |

---

## Troubleshooting

### Module Not Found

**Error:** `failed to load module manifest`

**Fix:**
- Verify `agenticmodule.yaml` exists in `agenticmodules/<name>/`
- Check file path and permissions
- Validate YAML syntax

### Extension Not Found

**Error:** `module step 0 references extension 'x' which is not found`

**Fix:**
- Ensure extension is in `extensions/` directory
- Verify extension name matches exactly
- Restart LlamaGate to reload extensions

### Module Execution Timeout

**Error:** `maximum runtime exceeded`

**Fix:**
- Increase `max_runtime_seconds` in module runner input
- Optimize extension execution time
- Check for infinite loops in extensions

---

## Next Steps

- **Extension Quick Start:** [EXTENSIONS_QUICKSTART.md](EXTENSIONS_QUICKSTART.md)
- **Extension Specification:** [EXTENSIONS_SPEC_V0.9.1.md](EXTENSIONS_SPEC_V0.9.1.md)
- **UX Commands:** [UX_COMMANDS.md](UX_COMMANDS.md)
- **Example Modules:** [LlamaGate Extension Examples Repository](https://github.com/smcdaniel54/LlamaGate-extension-examples) - Real-world AgenticModule patterns and templates
- **In-Repo Examples:** `examples/agenticmodules/intake_and_routing/`

---

**For more details, see the [Extension Specification](EXTENSIONS_SPEC_V0.9.1.md).**
