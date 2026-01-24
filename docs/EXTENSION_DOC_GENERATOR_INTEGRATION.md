# Extension Doc Generator - LlamaGate Integration Guide

**Purpose**: Guide for integrating `extension-doc-generator` as a builtin extension in LlamaGate  
**Status**: Integration Plan (Approved Approach)  
**Type**: Builtin Extension (YAML-based)

---

## Architecture Fit

### Extension Types in LlamaGate

LlamaGate has three types of extensions:

1. **Builtin Extensions** (Go code):
   - Location: `internal/extensions/builtin/`
   - Compiled into binary
   - Core system functionality
   - Examples: `validation`, `tools`, `state`, `human`, `events`

2. **Builtin Extensions** (YAML-based):
   - Location: `extensions/builtin/` directory
   - Discovered at startup via `DiscoverExtensions()`
   - Core workflow capabilities
   - Manifest flag: `builtin: true`
   - Examples: `extension-doc-generator`
   - Loaded with priority

3. **Default Extensions** (YAML-based):
   - Location: `extensions/` directory
   - Discovered at startup via `DiscoverExtensions()`
   - Workflow extensions
   - Examples: `agenticmodule_runner`, `prompt-template-executor`

### Where This Fits

**`extension-doc-generator` is a Builtin Extension** (YAML-based):
- ✅ Approved approach: `extensions/builtin/` directory
- ✅ Uses `builtin: true` flag in manifest
- ✅ Priority loading in startup code
- ✅ Clear separation from default extensions
- ✅ See `docs/EXTENSION_DOC_GENERATOR_APPROVED_APPROACH.md` for details

---

## Implementation Steps

### Step 1: Create Extension Directory

In LlamaGate repository:

```bash
mkdir -p extensions/builtin/extension-doc-generator
```

**Note**: Using `extensions/builtin/` directory (approved approach for YAML builtin extensions)

### Step 2: Create Manifest

Create `extensions/builtin/extension-doc-generator/manifest.yaml`:

```yaml
name: extension-doc-generator
version: 1.0.0
description: Generates comprehensive markdown documentation for extensions and modules
type: workflow
builtin: true          # Mark as builtin extension
enabled: true

inputs:
  - id: target
    type: string
    description: Extension name or module name to document
    required: true
  - id: output_path
    type: string
    description: Path to save generated markdown (default: docs/extensions/{name}.md)
    required: false
  - id: format
    type: string
    enum: [markdown, html, json]
    default: markdown
    description: Output format
  - id: include_examples
    type: boolean
    default: true
    description: Include usage examples in documentation
  - id: include_api_details
    type: boolean
    default: true
    description: Include API endpoint details

outputs:
  - id: documentation
    type: string
    description: Generated markdown documentation
  - id: file_path
    type: string
    description: Path where documentation was saved
  - id: extension_info
    type: object
    description: Extension metadata used for generation

steps:
  - name: fetch_extension_info
    uses: http.get
    config:
      url: "http://localhost:11435/v1/extensions/{{inputs.target}}"
      headers:
        X-API-Key: "${API_KEY}"
    outputs:
      extension_data: ${steps.fetch_extension_info.response.body}
  
  - name: generate_documentation
    uses: llm.chat
    config:
      model: mistral
      system_prompt: |
        You are a technical documentation generator. Generate comprehensive, 
        well-formatted markdown documentation for LlamaGate extensions.
        
        Include:
        - Extension overview (name, version, type, description)
        - API endpoints (execute endpoint, custom endpoints if any)
        - Inputs table (name, type, required, description)
        - Outputs table (name, type, description)
        - Configuration options
        - Dependencies
        - Usage examples (curl, Python SDK)
        - Workflow steps (for workflow extensions)
        - Hooks (for middleware/observer extensions)
        
        Format as professional, readable markdown.
      user_prompt: |
        Generate documentation for this extension:
        
        Extension Data:
        {{extension_data}}
        
        Include examples: {{inputs.include_examples}}
        Include API details: {{inputs.include_api_details}}
        Output format: {{inputs.format}}
    outputs:
      documentation: ${steps.generate_documentation.output.content}
  
  - name: save_documentation
    uses: file.write
    config:
      path: "{{inputs.output_path | default: 'docs/extensions/{{inputs.target}}.md'}}"
      content: "{{documentation}}"
      create_directory: true
    outputs:
      file_path: ${steps.save_documentation.output.path}
```

### Step 3: Update LlamaGate Documentation

**Update `extensions/README.md`:**

Add to the list of builtin extensions:

```markdown
### Extension Documentation Generator

**Purpose:** Generate comprehensive markdown documentation for extensions and modules

**Type:** Builtin Extension (YAML-based)

**Location:** `builtin/extension-doc-generator/`

**Usage:**
```bash
curl -X POST http://localhost:11435/v1/extensions/extension-doc-generator/execute \
  -H "Content-Type: application/json" \
  -d '{
    "input": {
      "target": "orchestrator",
      "output_path": "docs/extensions/orchestrator.md"
    }
  }'
```

**Inputs:**
- `target` (required): Extension or module name
- `output_path` (optional): Where to save documentation
- `format` (optional): markdown, html, or json
- `include_examples` (optional): Include usage examples
- `include_api_details` (optional): Include API endpoint details
```

**Update `docs/EXTENSIONS_QUICKSTART.md`:**

Add section about builtin extensions:

```markdown
## Builtin Extensions

LlamaGate includes builtin extensions (YAML-based) that are always available:

- **extension-doc-generator**: Generates markdown documentation for extensions/modules

## Default Extensions

LlamaGate includes several default extensions (YAML-based) that are always available:

- **agenticmodule_runner**: Executes AgenticModules
- **prompt-template-executor**: Executes prompt templates with variables
```

### Step 4: Test Integration

1. **Start LlamaGate** - Extension should be discovered automatically
2. **Verify Discovery** - Check `GET /v1/extensions` includes `extension-doc-generator`
3. **Test Generation** - Generate docs for a test extension
4. **Verify Output** - Check generated markdown is complete and accurate

---

## Discovery Mechanism

### How It Works (Approved Approach)

1. **Startup**: LlamaGate calls `DiscoverExtensions("extensions/builtin/")` first (priority loading)
2. **Discovery**: Walks `extensions/builtin/` directory looking for `manifest.yaml` files
3. **Registration**: Each extension with valid manifest is registered with `builtin: true` flag
4. **Availability**: Extension becomes available at `/v1/extensions/{name}/execute`
5. **Regular Extensions**: Then loads regular extensions from `extensions/` (excluding `builtin/`)

**Code changes required** - update `startup.LoadInstalledExtensions()` to load `extensions/builtin/` first.

### Verification

After adding the extension:

```bash
# Check extension is discovered
curl http://localhost:11435/v1/extensions | jq '.[] | select(.name == "extension-doc-generator")'

# Should return extension metadata
```

---

## Documentation Strategy

### In LlamaGate Repo

**Clarify Extension Types:**

```markdown
## Extension Types

### Builtin Extensions (Go Code)
- Core functionality compiled into binary
- Location: `internal/extensions/builtin/`
- Examples: `validation`, `tools`, `state`

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
```

**Benefits:**
- Clear distinction between types
- No confusion about architecture
- Follows established patterns
- Easy to understand

---

## Benefits of This Approach

### For LlamaGate

- ✅ **Clear Separation**: Builtin YAML vs default YAML extensions
- ✅ **Priority Loading**: Builtin extensions loaded first
- ✅ **Standard Pattern**: Approved approach for YAML builtin extensions
- ✅ **Easy Maintenance**: Update YAML, no recompilation needed
- ✅ **Special Behavior**: Can enforce "can't disable" if needed
- ✅ **Discoverable**: Automatically discovered at startup

### For Users

- ✅ **Always Available**: Included in repo, no installation needed
- ✅ **Self-Documenting**: Extensions can document themselves
- ✅ **Consistent**: All extensions documented the same way
- ✅ **Discoverable**: Easy to see what extensions do

---

## Next Steps for LlamaGate Integration

1. **Update Code** (Required)
   - Add `Builtin` field to `Manifest` struct in `internal/extensions/manifest.go`
   - Update `startup.LoadInstalledExtensions()` to load `extensions/builtin/` first
   - Filter `builtin/` subdirectory from regular extension discovery

2. **Create Extension Manifest**
   - Create `extensions/builtin/extension-doc-generator/` directory
   - Design manifest.yaml with `builtin: true` flag
   - Define inputs/outputs
   - Design workflow steps

3. **Test Locally**
   - Test with LlamaGate extensions
   - Test with AgenticModules
   - Verify output quality
   - Verify priority loading

4. **Update Documentation**
   - Update `extensions/README.md` with extension types
   - Add extension-doc-generator entry
   - Update `docs/EXTENSIONS_QUICKSTART.md` if needed

5. **Create PR**
   - Add extension to `extensions/builtin/extension-doc-generator/`
   - Include code changes for builtin support
   - Demonstrate value

6. **Community Feedback**
   - Get feedback on generated docs
   - Refine based on usage
   - Iterate on template

---

*Integration Guide - 2026-01-24*
