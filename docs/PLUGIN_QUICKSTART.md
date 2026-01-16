# LlamaGate Plugin Quick Start - Migration Notice

> ⚠️ **IMPORTANT**: The Plugin System has been **removed** in LlamaGate v0.9.1 and replaced with the **Extension System**.

**Quick Summary:** If you're looking for the LlamaGate plugin quick start guide, the plugin system has been replaced with extensions. This page redirects you to the new extension quick start guide.

## Quick Migration Guide

The Go-based plugin system has been completely replaced with a YAML-based extension system that's easier to use and doesn't require compilation. To get started with extensions:

1. **Read the Extension Quick Start**: [Extension Quick Start Guide](./EXTENSION_QUICKSTART.md) - Get started with example extensions in 5 minutes
2. **See Example Extensions**: Check the `extensions/` directory in your LlamaGate installation
3. **Read the Full Guide**: [Extension Specification v0.9.1](./EXTENSIONS_SPEC_V0.9.1.md) - Complete extension system documentation
4. **Migration Help**: [Plugin Migration Guide](./PLUGINS.md) - Detailed migration instructions

### Key Differences: Plugins vs Extensions

| Feature | Old Plugin System | New Extension System |
|---------|------------------|---------------------|
| **Location** | `plugins/` directory | `extensions/` directory |
| **Format** | Go code (compiled) | YAML manifest files |
| **API Endpoints** | `/v1/plugins` | `/v1/extensions` |
| **Configuration** | `PLUGINS_ENABLED` env var | Auto-discovery (no config needed) |
| **Development** | Write Go code, compile | Write YAML, no compilation |
| **Discovery** | Manual registration | Automatic at startup |

### Quick Example

Instead of writing Go plugin code, you now create a YAML manifest:

```yaml
name: my-extension
version: 1.0.0
type: workflow
enabled: true
# ... rest of manifest
```

See [EXTENSION_QUICKSTART.md](./EXTENSION_QUICKSTART.md) for complete examples.

## Why Extensions Instead of Plugins?

The extension system offers several advantages:

- ✅ **No Compilation**: Create extensions with YAML files, no Go code needed
- ✅ **Auto-Discovery**: Extensions are automatically found and loaded
- ✅ **Easier Development**: Simple YAML syntax vs complex Go code
- ✅ **Better Documentation**: Self-documenting manifest files
- ✅ **Faster Deployment**: No build process required

## Next Steps

1. **Start with Examples**: Read [Extension Quick Start](./EXTENSION_QUICKSTART.md) to see working examples
2. **Understand the Format**: Review [Extension Specification](./EXTENSIONS_SPEC_V0.9.1.md) for manifest format
3. **Migrate Your Plugins**: Follow the [Plugin Migration Guide](./PLUGINS.md) for detailed steps
4. **Explore Examples**: Check the `extensions/` directory for complete extension examples

## Related Documentation

- **[Extension Quick Start](./EXTENSION_QUICKSTART.md)** - Get started with extensions in 5 minutes
- **[Plugin Migration Guide](./PLUGINS.md)** - Complete migration instructions
- **[Extension Specification](./EXTENSIONS_SPEC_V0.9.1.md)** - Full technical reference
- **[Extension API Reference](./API.md)** - REST API endpoints
- **[Extension Examples](../extensions/README.md)** - Example extensions with code

---

**For more information, see the [Extension System Documentation](./EXTENSIONS_SPEC_V0.9.1.md).**
