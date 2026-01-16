# LlamaGate v0.9.1 Release Notes

**Release Date:** 2026-01-15  
**Version:** 0.9.1

---

## 🚨 Breaking Changes

### Plugin System Removed

The Go-based plugin system has been **completely removed** and replaced with the YAML-based extension system. This is a **breaking change** requiring migration.

**What You Need to Do:**

1. **Update API Endpoints:**
   - `/v1/plugins` → `/v1/extensions`
   - `/v1/plugins/:name` → `/v1/extensions/:name`
   - `/v1/plugins/:name/execute` → `/v1/extensions/:name/execute`

2. **Migrate Your Plugins:**
   - Move from `plugins/` directory → `extensions/` directory
   - Convert Go plugin code → YAML manifest files
   - See `docs/EXTENSIONS_SPEC_V0.9.1.md` for manifest format

3. **Update Configuration:**
   - Remove `PLUGINS_ENABLED` environment variable (extensions auto-discover)
   - Remove `PluginsConfig` from config files
   - Extensions use YAML manifests - no config needed

4. **Update Code References:**
   - All plugin-related Go code has been removed
   - Use the new extension system instead

**Migration Guide:** See `docs/MIGRATION_STATUS.md` for detailed migration information.

---

## ✨ New Features

### Extension System v0.9.1

A new YAML-based extension system that replaces the old plugin system:

- **Auto-Discovery**: Extensions are automatically discovered from the `extensions/` directory
- **YAML Manifests**: Define extensions using YAML - no compilation required
- **Multiple Types**: Support for workflow, middleware, and observer extensions
- **Enable/Disable**: Control extensions via manifest or environment variables
- **Comprehensive API**: Full REST API for managing extensions (`/v1/extensions/*`)
- **Example Extensions**: Three example extensions included:
  - `prompt-template-executor` - Execute approved prompt templates
  - `request-inspector` - Redacted audit logging
  - `cost-usage-reporter` - Track token usage and costs

### Benefits Over Plugins

- **No Compilation**: Extensions are defined in YAML - no Go code needed
- **Hot-Reload Ready**: Architecture supports future hot-reload (not in v0.9.1)
- **Easier Distribution**: Share extensions as YAML files
- **Better Lifecycle Management**: Built-in enable/disable and discovery

---

## 🔄 Changes

### Default Port Changed

The default server port has been changed from `8080` to `11435` to avoid conflicts with common services. Existing installations with `.env` files will continue using their configured port.

---

## 📦 Installation

### Upgrade from v0.9.0

1. **Backup your configuration** (if you have custom configs)
2. **Migrate plugins to extensions** (see migration guide above)
3. **Update to v0.9.1**:
   ```bash
   # Using installer
   ./install/unix/install.sh  # or install.ps1 on Windows
   
   # Or build from source
   git checkout v0.9.1
   go build ./cmd/llamagate
   ```
4. **Verify extensions are discovered**:
   ```bash
   curl http://localhost:11435/v1/extensions
   ```

### Fresh Installation

Follow the installation guide in `docs/INSTALL.md` or `README.md`.

---

## 📚 Documentation

- **Extension Specification**: `docs/EXTENSIONS_SPEC_V0.9.1.md`
- **Migration Guide**: `docs/MIGRATION_STATUS.md`
- **Extension Quick Start**: `docs/EXTENSIONS_QUICKSTART.md`
- **API Documentation**: `docs/API.md`

---

## 🐛 Bug Fixes

- Fixed plugin system removal (breaking change, not a bug fix)
- All tests passing (10/10 packages)
- Build verified and working

---

## ⚠️ Known Issues

None at this time.

---

## 🙏 Thank You

Thank you for using LlamaGate! If you encounter any issues or have questions about the migration, please:

- Check the migration guide: `docs/MIGRATION_STATUS.md`
- Review the extension specification: `docs/EXTENSIONS_SPEC_V0.9.1.md`
- Open an issue on GitHub if you need help

---

## 📝 Full Changelog

See `CHANGELOG.md` for the complete list of changes.

---

**Previous Version:** [v0.9.0](https://github.com/llamagate/llamagate/releases/tag/v0.9.0)  
**Next Version:** [Unreleased](https://github.com/llamagate/llamagate/compare/v0.9.1...HEAD)
