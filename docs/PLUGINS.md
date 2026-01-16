# Plugin System - Migration Notice

> ⚠️ **IMPORTANT**: The Plugin System has been **removed** in LlamaGate v0.9.1 and replaced with the **Extension System**.

## Migration Required

The Go-based plugin system has been completely removed. If you were using plugins, you need to migrate to the new YAML-based extension system.

### What Changed

- **Plugin System** → **Extension System**
- `/v1/plugins` endpoints → `/v1/extensions` endpoints
- `plugins/` directory → `extensions/` directory
- Go plugin code → YAML manifest files
- `PLUGINS_ENABLED` env var → Extensions auto-discover (no config needed)

### Migration Guide

1. **Read the Extension Specification**: [EXTENSIONS_SPEC_V0.9.1.md](./EXTENSIONS_SPEC_V0.9.1.md)
2. **Quick Start**: [EXTENSION_QUICKSTART.md](./EXTENSION_QUICKSTART.md)
3. **Full Documentation**: [EXTENSIONS_DOCUMENTATION_INDEX.md](./EXTENSIONS_DOCUMENTATION_INDEX.md)

### Benefits of Extensions

- ✅ **No Compilation**: YAML-based, no Go code needed
- ✅ **Auto-Discovery**: Automatically found in `extensions/` directory
- ✅ **Multiple Types**: Workflow, middleware, and observer extensions
- ✅ **Enable/Disable**: Control via manifest or environment variables
- ✅ **Comprehensive API**: Full REST API for managing extensions

### Example Extensions

Three example extensions are included:
- `prompt-template-executor` - Execute approved prompt templates
- `request-inspector` - Redacted audit logging
- `cost-usage-reporter` - Track token usage and costs

See `extensions/` directory for examples.

---

**For more information, see the [Extension System Documentation](./EXTENSIONS_SPEC_V0.9.1.md).**
