# Import/Export System for Extensions and Modules

**LlamaGate Version:** 0.9.1+  
**Feature Version:** 1.0  
**Status:** Ready for Use

---

## Overview

LlamaGate provides a file-based import/export system for Extensions and Agentic Modules. This allows you to:

- Package extensions/modules as zip files
- Install packages that persist across LlamaGate rebuilds
- Share extensions/modules between installations
- Manage installed items via CLI commands

All installed items are stored in `~/.llamagate/` (Linux/mac) or `%USERPROFILE%\.llamagate` (Windows), separate from the LlamaGate repository.

---

## Quick Start

### Import an Extension

```bash
llamagate import extension my-extension.zip
```

### Export an Extension

```bash
llamagate export extension my-ext --out my-extension.zip
```

### List Installed Items

```bash
llamagate list extensions
llamagate list agentic-modules
```

### Remove an Item

```bash
llamagate remove extension my-ext
```

---

## Filesystem Layout

Installed items are stored in your home directory:

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
    ├── import/          # Staging for imports
    └── install/          # Staging for atomic installs
```

---

## CLI Commands

### Import

Import an extension or module from a zip file:

```bash
# Explicit type
llamagate import extension <zip-file>
llamagate import agentic-module <zip-file>

# Auto-detect type (convenience)
llamagate import <zip-file>
```

**Examples:**
```bash
llamagate import extension my-extension.zip
llamagate import agentic-module my-module.zip
llamagate import package.zip  # Auto-detects type
```

**What happens:**
1. Zip is extracted to staging directory
2. Package type is detected (extension vs module)
3. Manifest is validated
4. Installation is performed atomically (with backup if updating)
5. Registry is updated
6. Item is ready to use (enabled by default)

### Export

Export an installed extension or module to a zip file:

```bash
llamagate export extension <id> --out <zip-file>
llamagate export agentic-module <id> --out <zip-file>
```

**Examples:**
```bash
llamagate export extension my-ext --out my-extension.zip
llamagate export agentic-module my-module --out my-module.zip
```

**What's included:**
- All files from the installed directory
- `checksums.txt` with SHA256 hashes for integrity verification

### List

List all installed extensions or modules:

```bash
llamagate list extensions
llamagate list agentic-modules
```

**Output:**
```
Found 3 extensions:

  my-extension (v1.0.0) [enabled]
    ID: my-ext
    Path: ~/.llamagate/extensions/installed/my-ext

  another-ext (v2.1.0) [disabled]
    ID: another-ext
    Path: ~/.llamagate/extensions/installed/another-ext
```

### Remove

Remove an installed extension or module:

```bash
llamagate remove extension <id>
llamagate remove agentic-module <id>
```

**What happens:**
1. Item is removed from filesystem
2. Item is removed from registry
3. Backups are cleaned up

### Enable/Disable

Enable or disable an extension/module:

```bash
llamagate enable extension <id>
llamagate disable extension <id>
llamagate enable agentic-module <id>
llamagate disable agentic-module <id>
```

**Note:** Changes take effect on next server restart.

### Migrate

Migrate legacy extensions from repository to installed directory:

```bash
llamagate migrate [legacy-directory]
```

**Default:** Scans `extensions/` directory in current working directory.

**What happens:**
1. Scans legacy directory for extensions
2. Copies to `~/.llamagate/extensions/installed/`
3. Registers in registry
4. Creates migration marker (won't run again automatically)

**Note:** Migration runs automatically on first LlamaGate startup if not already done.

### Sync

Synchronize registry with filesystem:

```bash
llamagate sync
```

**Use cases:**
- Registry is out of sync with filesystem
- Manual cleanup of filesystem
- Recovery from corruption

---

## Manifest Schemas

### extension.yaml (Packaging Manifest)

New packaging manifest for extensions. Can coexist with `manifest.yaml`:

```yaml
id: my-extension              # Required: unique identifier (filesystem-safe)
name: My Extension           # Required: display name
version: 1.0.0               # Required: semantic version
description: Description     # Optional
enabled_default: true        # Optional: default enabled state (default: true)
hot_reload: true            # Optional: allow hot reload (default: true)
load_mode: eager            # Optional: eager|lazy (default: eager)
autostart: false            # Optional: auto-start on load (default: false)
```

**Note:** If `extension.yaml` exists, it's used for packaging metadata. The actual extension definition should be in `manifest.yaml` (existing format).

### module.yaml (Packaging Manifest)

New packaging manifest for modules:

```yaml
id: my-module                # Required: unique identifier
name: My Module             # Required: display name
version: 1.0.0              # Required: semantic version
entry_workflow: main        # Optional: entry workflow name
description: Description    # Optional
enabled_default: true       # Optional: default enabled state (default: true)
hot_reload: true           # Optional: allow hot reload (default: true)
```

**Note:** If `module.yaml` exists, it's used for packaging metadata. The actual module definition should be in `agenticmodule.yaml` (existing format).

### Backward Compatibility

The system supports both new and legacy manifest formats:

- **Extensions:** `extension.yaml` (new) or `manifest.yaml` (legacy)
- **Modules:** `module.yaml` (new) or `agenticmodule.yaml` (legacy)

If both exist, the new format takes precedence for packaging metadata.

---

## Package Structure

### Extension Package

```
my-extension.zip
├── manifest.yaml          # Extension manifest (required)
├── extension.yaml         # Packaging manifest (optional)
└── ...                    # Other files (scripts, templates, etc.)
```

### Module Package

```
my-module.zip
├── agenticmodule.yaml    # Module manifest (required)
├── module.yaml           # Packaging manifest (optional)
├── extensions/            # Module-specific extensions (optional)
│   └── ...
└── ...                    # Other files
```

---

## Import Process

1. **Extract to staging:** Zip is extracted to `~/.llamagate/tmp/import/<random>/`
2. **Detect type:** System checks for `extension.yaml`, `manifest.yaml`, `module.yaml`, or `agenticmodule.yaml`
3. **Validate:** Both packaging manifest and extension/module manifest are validated
4. **Atomic install:**
   - If updating: current installation is backed up to `backups/<id>/<timestamp>/`
   - New installation is copied to `~/.llamagate/tmp/install/<id>/`
   - Atomic swap: old removed, new renamed to final location
5. **Update registry:** Entry added/updated in `installed.json`
6. **Cleanup:** Backup removed after successful registry update

**Safety features:**
- Atomic operations prevent partial installs
- Automatic backups (keeps last 2)
- Rollback on failure
- Concurrent import protection

---

## Export Process

1. **Lookup:** Item is found in registry or scanned from disk
2. **Create zip:** All files from installed directory are added
3. **Add checksums:** `checksums.txt` with SHA256 hashes is included
4. **Write:** Zip is written to specified output path

---

## Startup Behavior

On LlamaGate startup:

1. **Migration (first run only):** Legacy extensions are automatically migrated
2. **Registry sync:** Registry is synchronized with filesystem (self-healing)
3. **Load installed:** Enabled installed extensions are loaded
4. **Load legacy:** Legacy extensions (not yet migrated) are loaded as fallback
5. **Log summary:** Counts and failures are logged

**Priority:** Installed extensions take precedence over legacy extensions.

---

## Best Practices

### Package Creation

1. **Include all dependencies:** Ensure all required files are in the zip
2. **Use semantic versioning:** Follow `MAJOR.MINOR.PATCH` format
3. **Set appropriate defaults:** Use `enabled_default: false` for experimental extensions
4. **Document dependencies:** Note any required extensions or modules

### ID Naming

- Use filesystem-safe characters: `[a-zA-Z0-9_-]`
- Keep IDs short and descriptive
- Match ID to extension name when possible
- Avoid special characters or spaces

### Version Management

- Increment version on changes
- Use semantic versioning
- Document breaking changes in version notes

### Backup Strategy

- Backups are created automatically on update
- Last 2 backups are kept
- Backups are in `backups/<id>/<timestamp>/`
- Manual backup: copy installed directory before update

---

## Troubleshooting

### Import Fails

**Error:** "manifest validation failed"
- **Solution:** Check manifest format and required fields

**Error:** "item not found"
- **Solution:** Verify zip contains valid manifest file

**Error:** "failed to install"
- **Solution:** Check disk space and permissions

### Export Fails

**Error:** "item not found"
- **Solution:** Verify ID exists: `llamagate list extensions`

**Error:** "source path does not exist"
- **Solution:** Run `llamagate sync` to repair registry

### Registry Issues

**Registry out of sync:**
```bash
llamagate sync
```

**Corrupted registry:**
- Delete `~/.llamagate/registry/installed.json`
- Run `llamagate sync` to rebuild from filesystem

---

## Examples

### Creating an Extension Package

1. **Create extension directory:**
```bash
mkdir my-extension
cd my-extension
```

2. **Create manifest.yaml:**
```yaml
name: my-extension
version: 1.0.0
description: My awesome extension
type: workflow
enabled: true
steps:
  - uses: llm.chat
```

3. **Create extension.yaml (optional):**
```yaml
id: my-extension
name: My Extension
version: 1.0.0
description: My awesome extension
enabled_default: true
hot_reload: true
```

4. **Package:**
```bash
zip -r ../my-extension.zip .
```

5. **Import:**
```bash
llamagate import extension ../my-extension.zip
```

### Sharing Extensions

1. **Export:**
```bash
llamagate export extension my-ext --out my-extension-v1.0.0.zip
```

2. **Share:** Send zip file to other users

3. **Import (on other system):**
```bash
llamagate import extension my-extension-v1.0.0.zip
```

---

## Migration from Legacy Layout

Legacy extensions in the repository `extensions/` directory are automatically discovered and can be used alongside installed extensions. To migrate them to the new layout:

```bash
llamagate migrate
```

This copies extensions to `~/.llamagate/extensions/installed/` and registers them. Migration runs automatically on first startup.

---

## API Integration

The import/export system integrates with the existing extension system:

- **Installed extensions** are loaded on startup
- **Legacy extensions** are loaded as fallback
- **Automatic discovery**: Discovery is automatically triggered after import and migration
- **Refresh endpoint** (`POST /v1/extensions/refresh`) is available for manual refresh if needed

---

## Limitations

- **Route removal:** Gin doesn't support route removal at runtime. Routes remain until server restart.
- **Concurrent imports:** Protected by mutex (process-level, not file-level locking).
- **Discovery:** Automatic discovery is best-effort. If the server is not running during import, extensions will be discovered on next server startup.

---

## See Also

- [Extension System Documentation](EXTENSIONS_SPEC_V0.9.1.md)
- [AgenticModules Guide](AGENTICMODULES.md)
- [Testing Guide](TESTING.md)
