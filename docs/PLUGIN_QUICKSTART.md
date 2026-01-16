# Plugin Quick Start - Migration Notice

> ⚠️ **IMPORTANT**: The Plugin System has been **removed** in LlamaGate v0.9.1 and replaced with the **Extension System**.

## Quick Migration

The plugin system has been replaced with a YAML-based extension system. To get started:

1. **Read the Extension Quick Start**: [EXTENSION_QUICKSTART.md](./EXTENSION_QUICKSTART.md)
2. **See Example Extensions**: Check the `extensions/` directory
3. **Read the Full Guide**: [EXTENSIONS_SPEC_V0.9.1.md](./EXTENSIONS_SPEC_V0.9.1.md)

### What You Need to Know

- **Old**: Go-based plugins in `plugins/` directory
- **New**: YAML-based extensions in `extensions/` directory
- **Old**: `/v1/plugins` API endpoints
- **New**: `/v1/extensions` API endpoints

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

---

**For more information, see the [Extension System Documentation](./EXTENSIONS_SPEC_V0.9.1.md).**
