# Extension Documentation Generator - Built-in Extension Proposal

**Purpose**: Generate comprehensive markdown documentation for individual extensions or modules  
**Status**: **Obsolete** — The extension system was removed in Phase 1. This proposal is retained for historical reference only.  
**Target**: Was intended as a builtin LlamaGate extension (YAML-based, included in repo)

---

## Problem Statement

Extensions and modules currently lack visible, discoverable documentation:
- ❌ No easy way to see what an extension does
- ❌ Inputs/outputs not clearly documented
- ❌ Usage examples scattered or missing
- ❌ API endpoints not obvious
- ❌ Hard to discover available capabilities

**Solution**: Builtin extension that generates markdown docs from manifest data

---

## Proposed Design

### Extension Name
`extension-doc-generator` (builtin extension)

### Status: Builtin Extension (YAML-based)

**Important Distinction**: LlamaGate has three types of extensions:

1. **Builtin Extensions** (Go code):
   - Located in `internal/extensions/builtin/`
   - Compiled into the binary
   - Core functionality (validation, tools, state, etc.)
   - Examples: `validation`, `tools`, `state`, `human`, `events`

2. **Builtin Extensions** (YAML-based):
   - Located in `extensions/builtin/` directory
   - Discovered at startup via `DiscoverExtensions()`
   - Core workflow capabilities
   - Examples: `extension-doc-generator`
   - Manifest flag: `builtin: true`

3. **Default Extensions** (YAML-based):
   - Located in `extensions/` directory
   - Discovered at startup via `DiscoverExtensions()`
   - YAML-based workflow extensions
   - Examples: `agenticmodule_runner`, `prompt-template-executor`

**This extension would be a "Builtin Extension"** (YAML-based), using the approved pattern with `extensions/builtin/` directory.

**Note**: While Go builtin extensions can also use LLMs (by calling LlamaGate's API internally), a YAML builtin extension is recommended for this use case.

**Why YAML instead of Go:**
- ✅ Uses existing extension system (minimal Go code changes needed)
- ✅ Can leverage LLM via `llm.chat` step for rich documentation generation
- ✅ Flexible and extensible (easy to refine prompts)
- ✅ Easy to update without recompiling
- ✅ Follows approved pattern for YAML builtin extensions
- ✅ Better fit: This is a workflow capability, not core system functionality

**If Go Builtin Preferred**: A Go builtin extension could also work (calling LlamaGate's chat completion API internally), but would require Go code changes and recompilation for updates. See `docs/EXTENSION_DOC_GENERATOR_GO_VS_YAML.md` for comparison.

**How it becomes "builtin":**
- Included in LlamaGate repo: `extensions/builtin/extension-doc-generator/`
- Discovered automatically at startup (via `DiscoverExtensions("extensions/builtin/")`)
- Always available (no installation needed)
- Loaded with priority
- Documented alongside other builtin extensions in `extensions/README.md`
- Listed in extension discovery output

### API Endpoint
```
POST /v1/extensions/extension-doc-generator/execute
```

### Inputs

```yaml
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
```

### Outputs

```yaml
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
```

### Implementation Approach

**Step 1: Fetch Extension/Module Info**
- Use `GET /v1/extensions/{name}` to get extension details
- Or load `agenticmodule.yaml` for modules
- Parse manifest.yaml for full metadata

**Step 2: Extract Information**
- Name, version, description
- Type (workflow, middleware, observer)
- Inputs (with types, descriptions, required flags)
- Outputs (with types, descriptions)
- Steps (workflow steps, hooks, etc.)
- Configuration options
- Dependencies
- API endpoints (for workflow extensions)

**Step 3: Generate Markdown**
- Use LLM to generate comprehensive documentation
- Include all extracted information
- Add usage examples
- Add API endpoint examples
- Format as readable markdown

**Step 4: Save Documentation**
- Write to specified path (or default location)
- Create directory if needed
- Return path and content

---

## Generated Documentation Structure

```markdown
# {Extension Name}

**Version**: {version}  
**Type**: {workflow|middleware|observer}  
**Status**: {enabled|disabled}

## Description

{description from manifest}

## API Endpoint

**Workflow Extensions:**
```
POST /v1/extensions/{name}/execute
```

**Custom Endpoints:**
{List any custom endpoints defined in manifest}

## Inputs

| Name | Type | Required | Description |
|------|------|----------|-------------|
| {input.id} | {input.type} | {yes/no} | {input.description} |

## Outputs

| Name | Type | Description |
|------|------|-------------|
| {output.id} | {output.type} | {output.description} |

## Configuration

{Configuration options if any}

## Dependencies

{List of dependencies}

## Usage Examples

### Basic Usage

```bash
curl -X POST http://localhost:11435/v1/extensions/{name}/execute \
  -H "Content-Type: application/json" \
  -H "X-API-Key: {api-key}" \
  -d '{
    "input": {
      "key": "value"
    }
  }'
```

### Python SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:11435/v1",
    api_key="sk-llamagate"
)

response = client.chat.completions.create(
    model="mistral",
    messages=[...]
)
```

## Workflow Steps

{For workflow extensions, document each step}

## Hooks

{For middleware/observer extensions, document hooks}

## Related Extensions

{List related or dependent extensions}

---
*Generated: {timestamp}*
```

---

## Benefits

### For Developers
- ✅ **Discoverability**: Easy to see what extensions do
- ✅ **API Documentation**: Clear endpoint examples
- ✅ **Usage Examples**: Copy-paste ready code
- ✅ **Input/Output Reference**: Quick reference guide

### For Users
- ✅ **Self-Documenting**: Extensions document themselves
- ✅ **Always Up-to-Date**: Generated from manifest (source of truth)
- ✅ **Consistent Format**: All extensions documented the same way
- ✅ **Easy Navigation**: Can browse all extension docs

### For LlamaGate
- ✅ **Better DX**: Improved developer experience
- ✅ **Reduced Support**: Less "how do I use this?" questions
- ✅ **Professional**: Shows maturity and completeness
- ✅ **Builtin Extension**: Always available (included in repo), no installation needed
- ✅ **Minimal Code Changes**: Uses existing extension system, just add YAML file

---

## Implementation Options

### Implementation: YAML Builtin Extension (Approved)

**Approach**: YAML-based builtin extension included in LlamaGate repo

**Pros:**
- ✅ Uses existing LlamaGate extension system
- ✅ Can use LLM to generate rich documentation
- ✅ Flexible and extensible
- ✅ Works with current architecture
- ✅ Minimal Go code changes needed
- ✅ Easy to update and maintain
- ✅ Follows approved pattern for YAML builtin extensions

**How it works:**
- Extension manifest in `extensions/builtin/extension-doc-generator/manifest.yaml`
- Discovered automatically at startup (via `DiscoverExtensions("extensions/builtin/")`)
- Uses `llm.chat` to generate documentation from manifest data
- Reads extension info via file system or `GET /v1/extensions/{name}`
- Saves generated docs to file system
- Returns generated content

**Location in LlamaGate repo:**
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

**Note**: This is a "builtin extension" (YAML-based, always included), not a "default extension" (YAML). The distinction:
- **Builtin extensions** (Go): Core functionality in `internal/extensions/builtin/`
- **Builtin extensions** (YAML): Core workflow capabilities in `extensions/builtin/`
- **Default extensions** (YAML): Included in repo, always available, YAML-based in `extensions/`

---

## Example Usage

### Generate Extension Documentation

```bash
curl -X POST http://localhost:11435/v1/extensions/extension-doc-generator/execute \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "input": {
      "target": "orchestrator",
      "output_path": "docs/extensions/orchestrator.md",
      "include_examples": true,
      "include_api_details": true
    }
  }'
```

### Generate Module Documentation

```bash
curl -X POST http://localhost:11435/v1/extensions/extension-doc-generator/execute \
  -H "Content-Type: application/json" \
  -d '{
    "input": {
      "target": "trio-core",
      "output_path": "docs/modules/trio-core.md"
    }
  }'
```

### Generate All Extensions

```bash
# List all extensions
EXTENSIONS=$(curl -s http://localhost:11435/v1/extensions | jq -r '.[].name')

# Generate docs for each
for ext in $EXTENSIONS; do
  curl -X POST http://localhost:11435/v1/extensions/extension-doc-generator/execute \
    -H "Content-Type: application/json" \
    -d "{\"input\": {\"target\": \"$ext\"}}"
done
```

---

## Integration Points

### With LlamaGate API
- Uses `GET /v1/extensions/{name}` to fetch extension info
- Can access manifest.yaml directly from file system
- Can use extension registry for discovery

### With AgenticModules
- Can generate docs for modules (reads agenticmodule.yaml)
- Can generate docs for extensions in modules
- Can be called from any client (Python, curl, etc.)

### With CI/CD
- Auto-generate docs on extension changes
- Keep documentation in sync with code
- Generate docs as part of build process

---

## Implementation in LlamaGate Repo

### Architecture Fit

**Approved Approach**: YAML Builtin Extension (Hybrid)

- **Builtin Extensions** (Go): Core functionality in `internal/extensions/builtin/`
- **Builtin Extensions** (YAML): Core workflow capabilities in `extensions/builtin/`
- **Default Extensions** (YAML): Workflow extensions in `extensions/` directory
- **This Extension**: YAML-based builtin extension (approved pattern)

**Approved Pattern**: Use `extensions/builtin/` directory with `builtin: true` flag for YAML extensions that are core functionality.

### Step 1: Create Extension Directory

In LlamaGate repository:
```
extensions/builtin/extension-doc-generator/
├── manifest.yaml
└── README.md (optional)
```

**Location**: `extensions/builtin/` (for builtin YAML extensions, not regular `extensions/`)

### Step 2: Extension Discovery

**Code Changes Required** (approved approach):

The extension will be discovered from `extensions/builtin/` directory:
- LlamaGate will call `DiscoverExtensions("extensions/builtin/")` at startup (priority loading)
- `DiscoverExtensions()` walks the directory looking for `manifest.yaml` files
- Extension is registered with `builtin: true` flag set
- **Code changes needed** - update `startup.LoadInstalledExtensions()` to load `extensions/builtin/` first
- **Approved pattern** - YAML builtin extensions in `extensions/builtin/` subdirectory

### Step 3: Documentation in LlamaGate

**Where to document:**
1. **`extensions/README.md`** - Add to list of builtin extensions
2. **`docs/EXTENSIONS_QUICKSTART.md`** - Mention as always-available extension
3. **`docs/API.md`** - Document the API endpoint

**Documentation approach:**
- Document as "builtin extension" (YAML-based, included in repo)
- Distinguish from "builtin extensions" (Go code) and "default extensions" (YAML)
- List alongside other builtin extensions
- Mention in extension discovery output

**Example documentation entry:**
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
```

**Benefits of this approach:**
- ✅ Clear separation (builtin YAML vs default YAML)
- ✅ Priority loading (builtin loaded first)
- ✅ Can enforce special behavior (can't disable, etc.)
- ✅ Easy to identify builtin YAML extensions
- ✅ Uses existing extension system
- ✅ Easy to update (just update YAML)
- ✅ Approved pattern for YAML builtin extensions

## Next Steps

1. **Create Extension Manifest**
   - Design manifest.yaml for extension-doc-generator
   - Define inputs/outputs
   - Design workflow steps using `llm.chat`

2. **Implement Documentation Generation**
   - Use LLM to generate comprehensive docs
   - Extract info from manifest or API
   - Format as markdown

3. **Test with Real Extensions**
   - Test with existing LlamaGate extensions
   - Test with AgenticModules
   - Verify output quality

4. **Add to LlamaGate Repo**
   - Create `extensions/builtin/extension-doc-generator/` directory
   - Add manifest.yaml with `builtin: true` flag
   - Update `startup.LoadInstalledExtensions()` to load `extensions/builtin/` first
   - Add `Builtin` field to `Manifest` struct in `internal/extensions/manifest.go`
   - Update `extensions/README.md` to document extension types
   - Update `docs/EXTENSIONS_QUICKSTART.md` if needed
   - Create PR demonstrating value
   - **Code changes needed** - see `docs/EXTENSION_DOC_GENERATOR_YAML_BUILTIN_ACCOMMODATION.md` for details

---

## Alternative: Module Documentation Generator

Could also create a separate `module-doc-generator` that:
- Reads `agenticmodule.yaml`
- Documents all extensions in module
- Documents workflow steps
- Documents inputs/outputs at module level
- Generates comprehensive module README

**Both could be builtin extensions in LlamaGate!**

---

*Proposal - 2026-01-24*
