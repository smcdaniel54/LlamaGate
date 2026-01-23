# Import/Export System Implementation Plan

**LlamaGate Version:** 0.9.1+  
**Implementation Date:** 2026-01-22

## Overview
This document outlines the plan for implementing a minimal, file-based Import/Export system for Extensions and Agentic Modules in LlamaGate, plus Startup Reload functionality.

## Status: ✅ COMPLETE

### ✅ Completed Components

1. **internal/homedir** - Home directory resolution
   - ✅ `GetHomeDir()` - Returns `~/.llamagate` (Linux/mac) or `%USERPROFILE%\.llamagate` (Windows)
   - ✅ `GetExtensionsDir()` - Returns `~/.llamagate/extensions/installed/`
   - ✅ `GetAgenticModulesDir()` - Returns `~/.llamagate/agentic-modules/installed/`
   - ✅ `GetRegistryDir()` - Returns `~/.llamagate/registry/`
   - ✅ `GetTempDir()` - Returns `~/.llamagate/tmp/`
   - ✅ `GetImportStagingDir()` - Returns `~/.llamagate/tmp/import/`
   - ✅ `GetInstallStagingDir()` - Returns `~/.llamagate/tmp/install/`
   - ✅ Tests implemented

2. **internal/registry** - Registry store/load
   - ✅ `NewRegistry()` - Creates/loads registry from `installed.json`
   - ✅ `Register()` - Adds/updates item in registry
   - ✅ `Unregister()` - Removes item from registry
   - ✅ `Get()` - Retrieves item by ID
   - ✅ `List()` - Lists all items (optionally filtered by type)
   - ✅ `SetEnabled()` - Sets enabled status
   - ✅ `Exists()` - Checks if item exists
   - ✅ Atomic save with temp file + rename
   - ✅ Tests implemented

3. **internal/packaging** - Zip import/export
   - ✅ `Import()` - Imports zip file (extracts, validates, installs atomically)
   - ✅ `Export()` - Exports installed item to zip (with checksums.txt)
   - ✅ `DetectPackageType()` - Detects extension vs module by manifest
   - ✅ `LoadExtensionPackageManifest()` - Loads extension.yaml or manifest.yaml
   - ✅ `LoadModulePackageManifest()` - Loads module.yaml or agenticmodule.yaml
   - ✅ `ValidatePackageManifest()` - Validates packaging manifest fields
  - ✅ Atomic install with staging directory
  - ✅ Zip slip protection
  - ✅ Automatic discovery support (always calls /v1/extensions/refresh after import)
  - ✅ Tests implemented

4. **internal/discovery** - Discovery scan
   - ✅ `DiscoverInstalledItems()` - Discovers all installed items (registry + disk scan fallback)
   - ✅ `DiscoverEnabledItems()` - Discovers only enabled items
   - ✅ `scanExtensions()` - Scans extensions directory
  - ✅ `scanModules()` - Scans modules directory
  - ✅ `DiscoverLegacyExtensions()` - Discovers legacy repo-based extensions
  - ✅ Tests implemented

5. **internal/startup** - Startup integration
  - ✅ `LoadInstalledExtensions()` - Loads installed and legacy extensions at startup
  - ✅ `LoadInstalledModules()` - Discovers installed modules at startup
  - ✅ Automatic migration support
  - ✅ Conflict resolution (installed takes precedence)

6. **internal/migration** - Legacy migration
  - ✅ `HasMigrated()` - Checks if migration has been performed
  - ✅ `MarkMigrated()` - Marks migration as complete
  - ✅ `MigrateLegacyExtensions()` - Migrates extensions from repo to home directory

### ✅ Completed Components (Continued)

7. **CLI Tool (cmd/llamagate-cli)**
  - ✅ `llamagate import extension <zip>` / `import agentic-module <zip>` / `import <zip>` (auto-detect)
  - ✅ `llamagate export extension <id> --out <zip>` / `export agentic-module <id> --out <zip>`
  - ✅ `llamagate list extensions` / `list agentic-modules`
  - ✅ `llamagate remove extension <id>` / `remove agentic-module <id>`
  - ✅ `llamagate enable extension <id>` / `enable agentic-module <id>`
  - ✅ `llamagate disable extension <id>` / `disable agentic-module <id>`
  - ✅ `llamagate migrate` - Migrate legacy extensions
  - ✅ `llamagate sync` - Sync registry with filesystem

8. **Startup Reload Integration**
  - ✅ Modified `cmd/llamagate/main.go` to:
    - Load installed extensions from `~/.llamagate/extensions/installed/`
    - Load installed modules from `~/.llamagate/agentic-modules/installed/`
    - Register enabled items
    - Log summary of loaded counts and failures
    - Support legacy extensions directory (repo-based) as fallback

9. **Automatic Discovery Support**
  - ✅ Automatic discovery always triggers after import (extensions and modules)
  - ✅ Automatic discovery triggers after legacy extension migration
  - ✅ Calls `/v1/extensions/refresh` endpoint (best-effort, fails silently if server not running)
  - ✅ Respects `LLAMAGATE_API_KEY` environment variable for authentication
  - ✅ Manual refresh endpoint calls are now obsolete for normal operations

10. **Manifest Schema Updates**
  - ✅ Support for new manifest fields:
    - `extension.yaml`: `id`, `enabled_default`, `hot_reload`, `load_mode`, `autostart`
    - `module.yaml`: `id`, `entry_workflow`, `enabled_default`, `hot_reload`
  - ✅ Backward compatibility with existing `manifest.yaml` and `agenticmodule.yaml`
  - ✅ Always loads `manifest.yaml`/`agenticmodule.yaml` first, then merges packaging metadata

11. **Backward Compatibility**
  - ✅ Support existing `extensions/` directory (repo-based) as fallback
  - ✅ Migration path: automatic migration on first startup
  - ✅ Keep existing extension discovery working
  - ✅ Installed extensions take precedence over legacy extensions

12. **Tests**
  - ✅ Packaging tests (import, export, validation, remove, detect)
  - ✅ Discovery tests (scan, enabled items, empty directory)
  - ✅ Registry tests (register, unregister, enable/disable, persistence)
  - ✅ Homedir tests (platform-specific paths)

13. **Documentation**
  - ✅ `docs/packaging.md` - User guide for import/export
  - ✅ `docs/PACKAGING_PLAN.md` - Implementation plan
  - ✅ `docs/PACKAGING_RECOMMENDATIONS.md` - Recommendations and decisions

## Filesystem Layout

```
~/.llamagate/
├── extensions/
│   └── installed/
│       ├── <id1>/
│       │   ├── manifest.yaml (or extension.yaml)
│       │   └── ...
│       └── <id2>/
│           └── ...
├── agentic-modules/
│   └── installed/
│       ├── <id1>/
│       │   ├── module.yaml (or agenticmodule.yaml)
│       │   └── ...
│       └── <id2>/
│           └── ...
├── registry/
│   └── installed.json
└── tmp/
    ├── import/
    │   └── <random>/
    └── install/
        └── <id>/
```

## Manifest Schemas

### extension.yaml (new packaging manifest)
```yaml
id: my-extension              # Required: unique identifier
name: My Extension           # Required: display name
version: 1.0.0               # Required: semantic version
description: Description     # Optional
enabled_default: true        # Optional: default enabled state (default: true)
hot_reload: true            # Optional: allow hot reload (default: true)
load_mode: eager            # Optional: eager|lazy (default: eager)
autostart: false            # Optional: auto-start on load (default: false)
```

**Note:** If `extension.yaml` exists, it's used for packaging metadata. The actual extension definition can be in `manifest.yaml` (existing format) or embedded in `extension.yaml`.

### module.yaml (new packaging manifest)
```yaml
id: my-module                # Required: unique identifier
name: My Module             # Required: display name
version: 1.0.0              # Required: semantic version
entry_workflow: main        # Optional: entry workflow name
description: Description    # Optional
enabled_default: true       # Optional: default enabled state (default: true)
hot_reload: true           # Optional: allow hot reload (default: true)
```

**Note:** If `module.yaml` exists, it's used for packaging metadata. The actual module definition can be in `agenticmodule.yaml` (existing format) or embedded in `module.yaml`.

## Import Semantics

1. **Unzip to staging**: Extract zip to `~/.llamagate/tmp/import/<random>/`
2. **Detect type**: Check for `extension.yaml`, `manifest.yaml`, `module.yaml`, or `agenticmodule.yaml`
3. **Validate manifest**: Validate packaging manifest + extension/module manifest
4. **Atomic install**:
   - Copy to `~/.llamagate/tmp/install/<id>/`
   - Remove old installation if exists
   - Rename staging to `~/.llamagate/extensions/installed/<id>/` or `~/.llamagate/agentic-modules/installed/<id>/`
5. **Update registry**: Add/update entry in `installed.json`
6. **Automatic discovery**: Always trigger discovery by calling `/v1/extensions/refresh` (best-effort, fails silently if server not running)

## Export Semantics

1. **Lookup by ID**: Find item in registry or scan disk
2. **Zip directory**: Create zip with all files from installed directory
3. **Include checksums**: Add `checksums.txt` with SHA256 hashes
4. **Write to output**: Save zip to specified path

## CLI Command Structure

```bash
# Import
llamagate import extension <zip-file>
llamagate import agentic-module <zip-file>

# Export
llamagate export extension <id> --out <zip-file>
llamagate export agentic-module <id> --out <zip-file>

# List
llamagate list extensions
llamagate list agentic-modules

# Remove
llamagate remove extension <id>
llamagate remove agentic-module <id>

# Enable/Disable
llamagate enable extension <id>
llamagate disable extension <id>
llamagate enable agentic-module <id>
llamagate disable agentic-module <id>
```

## Startup Reload Flow

1. **Load registry**: Read `~/.llamagate/registry/installed.json`
2. **Discover installed items**: Use registry + disk scan fallback
3. **Filter enabled**: Only load items with `enabled: true`
4. **Register extensions**: Load extension manifests, register with extension registry
5. **Register modules**: Load module manifests (for future use)
6. **Log summary**: 
   ```
   Loaded 5 extensions, 2 modules
   Failed to load: extension-xyz (validation error: ...)
   ```
7. **Fallback to legacy**: If no installed items found, scan `extensions/` directory (repo-based)

## Backward Compatibility Strategy

1. **Dual discovery**: Check both `~/.llamagate/extensions/installed/` and `extensions/` (repo)
2. **No breaking changes**: Existing extension discovery continues to work
3. **Migration**: On first startup, optionally offer to migrate legacy extensions
4. **Priority**: Installed extensions take precedence over repo extensions

## Testing Strategy

1. **Unit tests**: Each package (homedir, registry, packaging, discovery)
2. **Integration tests**: 
   - Import → Export → Verify
   - Startup reload with various states
   - Legacy extension fallback
3. **CLI tests**: Test each command with various inputs
4. **Edge cases**: 
   - Invalid zip files
   - Missing manifests
   - Corrupted registry
   - Concurrent imports

## Open Questions / Decisions Needed

1. **CLI library**: Use standard library `flag` or add dependency (cobra/spf13)?
   - **Decision**: Use standard library to keep lean

2. **Hot reload**: How to trigger refresh if server is running?
   - **Option A**: HTTP call to `/v1/extensions/refresh` (if server running)
   - **Option B**: Signal/event system
   - **Decision**: Try HTTP call, log if fails (non-critical)

3. **Legacy migration**: Automatic or manual?
   - **Decision**: Automatic discovery, optional explicit migration command

4. **Manifest format**: Separate `extension.yaml` vs embed in `manifest.yaml`?
   - **Decision**: Support both - prefer `extension.yaml` for packaging, fallback to `manifest.yaml`

5. **Module support**: How to use installed modules?
   - **Decision**: Modules are discovered but execution still uses `agenticmodule_runner` extension
   - Modules are just packaged bundles of extensions

## Implementation Status

1. ✅ Create homedir package
2. ✅ Create registry package
3. ✅ Create packaging package (import/export)
4. ✅ Create discovery package
5. ✅ Create CLI tool
6. ✅ Integrate startup reload
7. ✅ Add tests
8. ✅ Write documentation
9. ✅ Add hot reload support
10. ✅ Fix critical bugs (manifest loading, deadlock)

**All planned features are now complete!**

## Notes

- All packages use standard library only (no external dependencies for core functionality)
- Atomic operations ensure consistency (temp files + rename)
- Registry is JSON for simplicity and human-readability
- Zip format is standard and widely supported
- Checksums provide integrity verification
