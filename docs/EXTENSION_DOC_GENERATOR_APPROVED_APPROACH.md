# Extension Doc Generator - Approved Implementation Approach

**Purpose**: Document the approved approach for implementing `extension-doc-generator`  
**Status**: Approved  
**Date**: 2026-01-24

---

## Approved Approach: YAML Builtin Extension

**Decision**: Use **Hybrid Approach (Option 3 + Option 2)** from `EXTENSION_DOC_GENERATOR_YAML_BUILTIN_ACCOMMODATION.md`

### Key Decisions

1. **Directory Structure**: `extensions/builtin/` for builtin YAML extensions
2. **Manifest Flag**: `builtin: true` flag in manifest
3. **Priority Loading**: Load builtin YAML extensions first
4. **Documentation**: Clear distinction between extension types

---

## Implementation Details

### Directory Structure

```
LlamaGate/
└── extensions/
    ├── builtin/                    # Builtin YAML extensions
    │   └── extension-doc-generator/
    │       ├── manifest.yaml       # builtin: true
    │       └── README.md (optional)
    └── agenticmodule_runner/        # Regular default extensions
        └── manifest.yaml
```

### Manifest Schema Changes

**File**: `internal/extensions/manifest.go`

```go
type Manifest struct {
    Name        string                 `yaml:"name"`
    Version     string                 `yaml:"version"`
    Description string                 `yaml:"description"`
    Type        string                 `yaml:"type"`
    Builtin     bool                   `yaml:"builtin,omitempty"`  // NEW
    Enabled     bool                   `yaml:"enabled,omitempty"`
    // ... rest of fields
}
```

### Startup Code Changes

**File**: `internal/startup/startup.go`

```go
func LoadInstalledExtensions(extRegistry *extensions.Registry, legacyBaseDir string) (int, []string) {
    loadedCount := 0
    failures := []string{}
    
    // 1. Load builtin YAML extensions first (priority)
    builtinDir := filepath.Join(legacyBaseDir, "builtin")
    if builtinManifests, err := extensions.DiscoverExtensions(builtinDir); err == nil {
        for _, manifest := range builtinManifests {
            manifest.Builtin = true  // Set flag
            // Load with special handling (can't disable, priority, etc.)
            if err := extRegistry.RegisterOrUpdate(manifest); err != nil {
                failures = append(failures, fmt.Sprintf("%s: %v", manifest.Name, err))
            } else {
                loadedCount++
            }
        }
    }
    
    // 2. Load regular extensions (existing logic)
    // ... existing code for installed and legacy extensions
    // Note: Filter out builtin/ subdirectory from regular discovery
    
    return loadedCount, failures
}
```

---

## Extension Manifest

**File**: `extensions/builtin/extension-doc-generator/manifest.yaml`

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

---

## Documentation Updates

### Extension Types Documentation

**File**: `extensions/README.md`

```markdown
## Extension Types

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
```

### Extension Documentation Generator Entry

**File**: `extensions/README.md`

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

---

## Implementation Checklist

### For LlamaGate Repo

- [ ] Add `Builtin` field to `Manifest` struct in `internal/extensions/manifest.go`
- [ ] Update `startup.LoadInstalledExtensions()` to load `extensions/builtin/` first
- [ ] Filter `builtin/` subdirectory from regular extension discovery
- [ ] Create `extensions/builtin/extension-doc-generator/` directory
- [ ] Create `manifest.yaml` with `builtin: true` flag
- [ ] Test extension discovery at startup
- [ ] Verify API endpoint works
- [ ] Update `extensions/README.md` with extension types documentation
- [ ] Update `extensions/README.md` with extension-doc-generator entry
- [ ] Update `docs/EXTENSIONS_QUICKSTART.md` if needed
- [ ] Generate docs for test extensions
- [ ] Create PR demonstrating value

---

## Benefits of Approved Approach

- ✅ **Clear Separation**: Builtin YAML vs default YAML extensions
- ✅ **Priority Loading**: Builtin extensions loaded first
- ✅ **Manifest Flag**: `builtin: true` for clarity
- ✅ **Can Enforce Special Behavior**: Can't disable, higher priority, etc.
- ✅ **Easy to Identify**: Clear directory structure
- ✅ **Minimal Code Changes**: Just update startup loading logic
- ✅ **Uses Existing System**: Leverages `DiscoverExtensions()` mechanism

---

## Summary

**Approved Approach**: YAML Builtin Extension

**Key Elements**:
1. **Directory**: `extensions/builtin/extension-doc-generator/`
2. **Manifest Flag**: `builtin: true`
3. **Code Changes**: Update `startup.LoadInstalledExtensions()` to load `extensions/builtin/` first
4. **Schema Changes**: Add `Builtin` field to `Manifest` struct
5. **Documentation**: Clear distinction between extension types

**Status**: ✅ Approved and ready for implementation

---

*Approved Approach - 2026-01-24*
