# Builtin Extensions Documentation Summary

**Date**: 2026-01-23  
**Status**: Complete

## Documentation Updated

All documentation has been updated to include information about builtin extensions:

### ✅ Updated Files

1. **`extensions/README.md`**
   - ✅ Added "Extension Types" section explaining three types of extensions
   - ✅ Added "Builtin Extensions" section with extension-doc-generator documentation
   - ✅ Includes usage examples and API endpoint information

2. **`docs/EXTENSIONS_QUICKSTART.md`**
   - ✅ Added "Extension Types" section at the top
   - ✅ Added "Builtin Extensions" section with extension-doc-generator details
   - ✅ Explains builtin vs default extensions

3. **`docs/API.md`**
   - ✅ Added "Extension Types" overview in Extensions API section
   - ✅ Added "Extension Documentation Generator (Builtin Extension)" endpoint documentation
   - ✅ Includes request/response examples and parameters

4. **`docs/EXTENSIONS_SPEC_V0.9.1.md`**
   - ✅ Updated directory structure to show `extensions/builtin/` subdirectory
   - ✅ Updated discovery process to mention builtin extensions loaded first
   - ✅ Added `builtin: boolean` field to manifest schema
   - ✅ Updated enable/disable section to note builtin extensions cannot be disabled

### Documentation Coverage

**Extension Types Explained:**
- ✅ Builtin Extensions (Go Code) - `internal/extensions/builtin/`
- ✅ Builtin Extensions (YAML-based) - `extensions/builtin/` with `builtin: true`
- ✅ Default Extensions (YAML-based) - `extensions/` directory

**Extension-Doc-Generator Documented:**
- ✅ Purpose and description
- ✅ API endpoint: `POST /v1/extensions/extension-doc-generator/execute`
- ✅ Input parameters (target, output_path, format, include_examples, include_api_details)
- ✅ Output parameters (documentation, file_path)
- ✅ Usage examples (curl commands)
- ✅ Location: `extensions/builtin/extension-doc-generator/`

**Builtin Extension Behavior Documented:**
- ✅ Loaded with priority (first in startup sequence)
- ✅ Cannot be disabled
- ✅ Cannot be unregistered
- ✅ Always enabled (even if manifest says `enabled: false`)
- ✅ Filtered from regular extension discovery

### Documentation Locations

Users can find builtin extension information in:

1. **Quick Reference**: `extensions/README.md`
   - Extension types overview
   - Extension-doc-generator usage

2. **Getting Started**: `docs/EXTENSIONS_QUICKSTART.md`
   - Extension types explanation
   - Builtin extensions section

3. **API Reference**: `docs/API.md`
   - Extension types overview
   - Extension-doc-generator API endpoint documentation

4. **Specification**: `docs/EXTENSIONS_SPEC_V0.9.1.md`
   - Complete manifest schema (includes `builtin` field)
   - Discovery process (builtin extensions loaded first)
   - Enable/disable behavior (builtin cannot be disabled)

### Example Documentation Entry

All docs now include this structure:

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

---

*Documentation Summary - 2026-01-23*
