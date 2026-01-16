# LlamaGate Plugin System Migration Guide

> ⚠️ **IMPORTANT**: The Plugin System has been **removed** in LlamaGate v0.9.1 and replaced with the **Extension System**.

**Quick Summary:** If you're looking for LlamaGate plugin documentation or need to migrate from plugins to extensions, this guide explains what changed and how to migrate. The Go-based plugin system has been completely replaced with a YAML-based extension system that's easier to use and doesn't require compilation.

## Migration Required

The Go-based plugin system has been completely removed in LlamaGate v0.9.1. If you were using plugins, you need to migrate to the new YAML-based extension system. This migration guide explains what changed and provides step-by-step instructions.

### What Changed in LlamaGate v0.9.1

The following changes were made when migrating from plugins to extensions:

- **Plugin System** → **Extension System** (complete replacement)
- `/v1/plugins` API endpoints → `/v1/extensions` API endpoints
- `plugins/` directory → `extensions/` directory
- Go plugin code → YAML manifest files (no compilation needed)
- `PLUGINS_ENABLED` environment variable → Extensions auto-discover (no config needed)
- Plugin registration code → Automatic discovery from `extensions/` directory

### How to Migrate from Plugins to Extensions

Follow these steps to migrate your LlamaGate plugins to the new extension system:

1. **Read the Extension Specification**: [Extension Specification v0.9.1](./EXTENSIONS_SPEC_V0.9.1.md) - Complete guide to the extension system architecture and manifest format
2. **Quick Start Guide**: [Extension Quick Start](./EXTENSION_QUICKSTART.md) - Get started with example extensions in 5 minutes
3. **Full Documentation Index**: [Extensions Documentation Index](./EXTENSIONS_SPEC_V0.9.1.md) - Browse all extension-related documentation
4. **Migration Checklist**: See [Implementation Plan](./EXTENSIONS_IMPLEMENTATION_PLAN.md) for detailed migration steps

**Key Migration Steps:**
- Convert your Go plugin code to a YAML manifest file
- Move files from `plugins/` to `extensions/` directory
- Update API calls from `/v1/plugins` to `/v1/extensions`
- Remove `PLUGINS_ENABLED` from your configuration (extensions auto-discover)

### Why Extensions Are Better Than Plugins

The new extension system offers significant advantages over the old plugin system:

- ✅ **No Compilation Required**: YAML-based manifests, no Go code compilation needed
- ✅ **Auto-Discovery**: Extensions are automatically discovered from the `extensions/` directory at startup
- ✅ **Multiple Extension Types**: Support for workflow, middleware, and observer extensions
- ✅ **Easy Enable/Disable**: Control extensions via manifest `enabled` field or environment variables
- ✅ **Comprehensive REST API**: Full REST API for managing extensions (`GET /v1/extensions`, `POST /v1/extensions/:name/execute`)
- ✅ **Simpler Development**: No need to write Go code - just create YAML manifest files
- ✅ **Better Documentation**: Extensions are self-documenting through their manifest files

### Example Extensions

Three example extensions are included with LlamaGate v0.9.1 to help you get started:

- **`prompt-template-executor`** - Execute approved prompt templates with structured inputs and produce deterministic output files
- **`request-inspector`** - Automatically intercept HTTP requests and create redacted audit logs for security compliance
- **`cost-usage-reporter`** - Track token usage and costs across all API requests for budget monitoring

See the `extensions/` directory in your LlamaGate installation for complete examples. Each extension includes a `manifest.yaml` file that demonstrates the extension format.

## Frequently Asked Questions (FAQ)

### Why was the plugin system removed?

The plugin system was replaced with extensions to provide a simpler, more maintainable architecture. Extensions use YAML manifests instead of Go code, making them easier to create, modify, and deploy without compilation.

### Can I still use my old plugins?

No, plugins are not compatible with LlamaGate v0.9.1+. You must migrate to the extension system. The migration process involves converting your plugin code to a YAML manifest file.

### How long does migration take?

Migration time depends on the complexity of your plugins. Simple plugins can be migrated in minutes, while complex plugins with custom logic may take longer. See the [Extension Specification](./EXTENSIONS_SPEC_V0.9.1.md) for detailed migration guidance.

### Are extensions as powerful as plugins?

Yes, extensions support the same functionality as plugins (workflows, middleware, observers) but with a simpler YAML-based approach. The extension system is designed to be more maintainable and easier to use.

### Where can I find help with migration?

- [Extension Specification](./EXTENSIONS_SPEC_V0.9.1.md) - Complete technical reference
- [Extension Quick Start](./EXTENSION_QUICKSTART.md) - Step-by-step examples
- [Extensions Documentation Index](./EXTENSIONS_DOCUMENTATION_INDEX.md) - All extension docs
- [Implementation Plan](./EXTENSIONS_IMPLEMENTATION_PLAN.md) - Detailed migration steps

## Related Documentation

- **[Extension System Overview](./EXTENSIONS_SPEC_V0.9.1.md)** - Complete extension system specification
- **[Extension Quick Start Guide](./EXTENSION_QUICKSTART.md)** - Get started with extensions in 5 minutes
- **[Extension API Reference](./API.md)** - REST API endpoints for extensions
- **[Extension Examples](../extensions/README.md)** - Example extensions with source code
- **[LlamaGate Main README](../README.md)** - Complete LlamaGate documentation

## Search Keywords

This page helps users searching for:
- LlamaGate plugin migration
- LlamaGate plugin to extension migration
- LlamaGate plugin system removed
- LlamaGate v0.9.1 breaking changes
- LlamaGate extension system
- How to migrate LlamaGate plugins
- LlamaGate plugin documentation
- LlamaGate extension guide

---

**For more information, see the [Extension System Documentation](./EXTENSIONS_SPEC_V0.9.1.md).**
