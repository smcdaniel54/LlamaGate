# Essential Recommendations for Import/Export System

**LlamaGate Version:** 0.9.1+  
**Date:** 2026-01-22

## Critical Issues to Address

### 1. **Missing Remove Functionality** ⚠️ HIGH PRIORITY
**Issue**: Plan includes `llamagate remove` commands but `packaging.Remove()` function is missing.

**Recommendation**: 
- Add `Remove(id string, itemType registry.ItemType) error` to `internal/packaging/packaging.go`
- Should:
  - Remove from registry
  - Delete installed directory
  - Unregister from extension registry (if extension)
  - Unregister routes (if extension with routes)
  - Handle errors gracefully (log, don't fail if already removed)

**Implementation**:
```go
func Remove(id string, itemType registry.ItemType) error {
    reg, err := registry.NewRegistry()
    if err != nil {
        return err
    }
    
    item, exists := reg.Get(id)
    if !exists {
        return fmt.Errorf("item not found: %s", id)
    }
    
    // Remove from filesystem
    if err := os.RemoveAll(item.SourcePath); err != nil {
        return fmt.Errorf("failed to remove directory: %w", err)
    }
    
    // Remove from registry
    return reg.Unregister(id)
}
```

### 2. **ID vs Name Consistency** ⚠️ HIGH PRIORITY
**Issue**: Extensions use `name` as identifier, but packaging uses `id`. Need to ensure they're consistent.

**Recommendation**:
- Use `name` from extension manifest as the `id` in packaging registry (if `id` not explicitly set)
- Validate that `id` matches extension name validation rules (alphanumeric + underscore + hyphen)
- Add validation: `id` must match `^[a-zA-Z0-9_-]+$` pattern
- Document that `id` should match `name` for consistency

**Code Fix Needed**:
```go
// In LoadExtensionPackageManifest, ensure ID matches name validation
if pkgManifest.ID == "" {
    pkgManifest.ID = extManifest.Name
}
// Validate ID format matches name format
nameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
if !nameRegex.MatchString(pkgManifest.ID) {
    return nil, fmt.Errorf("invalid id format: %s", pkgManifest.ID)
}
```

### 3. **Extension Registry Synchronization** ⚠️ HIGH PRIORITY
**Issue**: Two separate registries:
- `internal/extensions/registry.go` - Runtime extension registry
- `internal/registry/registry.go` - Packaging registry (installed.json)

**Recommendation**:
- On startup, sync packaging registry with extension registry
- When importing, register in BOTH registries
- When removing, unregister from BOTH registries
- Add helper function: `SyncToExtensionRegistry(reg *extensions.Registry) error`

**Implementation Location**: Add to `internal/packaging/packaging.go` or create `internal/packaging/sync.go`

### 4. **Startup Integration Strategy** ⚠️ HIGH PRIORITY
**Issue**: Current `main.go` only loads from `extensions/` directory. Need to merge with installed extensions.

**Recommendation**:
- Load installed extensions FIRST (from `~/.llamagate/extensions/installed/`)
- Then load legacy extensions (from `extensions/` directory)
- Handle conflicts: installed extensions take precedence
- Log summary: "Loaded X installed extensions, Y legacy extensions"
- If same name exists in both, log warning and use installed version

**Implementation**:
```go
// In main.go, replace current extension loading:
// 1. Discover installed extensions
installedItems, _ := discovery.DiscoverEnabledItems()
for _, item := range installedItems {
    if item.Type == registry.ItemTypeExtension {
        // Load and register
    }
}

// 2. Discover legacy extensions (only if not already installed)
legacyManifests, _ := extensions.DiscoverExtensions("extensions")
for _, manifest := range legacyManifests {
    if !extensionRegistry.Exists(manifest.Name) {
        // Register legacy extension
    }
}
```

### 5. **Route Cleanup on Remove** ⚠️ MEDIUM PRIORITY
**Issue**: When removing an extension, routes need to be unregistered from RouteManager.

**Recommendation**:
- `Remove()` function should accept optional RouteManager parameter
- Call `routeManager.UnregisterExtensionRoutes(id)` before removing
- Note: Gin routes can't be removed at runtime, but tracking map should be cleaned

### 6. **Concurrent Operation Safety** ⚠️ MEDIUM PRIORITY
**Issue**: No locking for import/export operations. Concurrent imports could cause conflicts.

**Recommendation**:
- Add file-level locking using `github.com/gofrs/flock` or similar
- Lock file: `~/.llamagate/.import.lock` during import
- Lock file: `~/.llamagate/.export.lock` during export
- Or use mutex in packaging package (simpler, but only works within process)

**Simple Solution**:
```go
var importMutex sync.Mutex

func Import(zipPath string) (*ImportResult, error) {
    importMutex.Lock()
    defer importMutex.Unlock()
    // ... existing code
}
```

### 7. **Error Recovery & Atomicity** ⚠️ MEDIUM PRIORITY
**Issue**: If install fails mid-way, partial state could remain.

**Recommendation**:
- Current atomic install (temp + rename) is good, but add:
  - Rollback: If registry update fails, restore old installation
  - Cleanup: Ensure temp directories are cleaned up even on error
  - Validation: Verify installation succeeded before updating registry

**Enhancement**:
```go
// Before atomic swap, save old path
oldPath := installDir + ".backup"
if exists {
    os.Rename(installDir, oldPath)
}

// After successful install and registry update, remove backup
defer func() {
    if err == nil {
        os.RemoveAll(oldPath)
    } else {
        // Rollback: restore backup
        os.Rename(oldPath, installDir)
    }
}()
```

### 8. **Hot Reload Implementation** ⚠️ LOW PRIORITY
**Issue**: Plan mentions HTTP call to `/v1/extensions/refresh` but no implementation.

**Recommendation**:
- Add optional `HotReload()` function to `packaging` package
- Try HTTP POST to `http://localhost:11435/v1/extensions/refresh` (configurable)
- If fails (server not running), log and continue (non-critical)
- Only attempt if `hot_reload: true` in manifest

**Implementation**:
```go
func attemptHotReload(apiKey string) {
    client := &http.Client{Timeout: 5 * time.Second}
    req, _ := http.NewRequest("POST", "http://localhost:11435/v1/extensions/refresh", nil)
    if apiKey != "" {
        req.Header.Set("X-API-Key", apiKey)
    }
    resp, err := client.Do(req)
    if err != nil {
        log.Debug().Err(err).Msg("Hot reload skipped (server not running)")
        return
    }
    defer resp.Body.Close()
    // Log success/failure
}
```

### 9. **Manifest Schema Validation** ⚠️ MEDIUM PRIORITY
**Issue**: New manifest fields (`id`, `enabled_default`, etc.) need proper validation.

**Recommendation**:
- Validate `id` format (alphanumeric + underscore + hyphen)
- Validate `load_mode` is "eager" or "lazy" (if set)
- Validate `version` follows semantic versioning (optional but recommended)
- Add validation in `ValidatePackageManifest()`

### 10. **Legacy Extension Migration** ⚠️ LOW PRIORITY
**Issue**: Plan mentions migration but no explicit migration command.

**Recommendation**:
- Add `llamagate migrate` command (optional)
- Scans `extensions/` directory
- Offers to import each as a package
- Or: Automatic on first startup (silent migration)

**Decision Needed**: Automatic vs Manual migration?

### 11. **CLI Command Structure** ⚠️ DESIGN DECISION
**Issue**: Current plan has verbose command structure.

**Recommendation**: Consider shorter aliases:
```bash
# Current plan:
llamagate import extension <zip>
llamagate import agentic-module <zip>

# Alternative (shorter):
llamagate import ext <zip>
llamagate import module <zip>
# Or:
llamagate import <zip>  # Auto-detect type
```

**Decision Needed**: Which format is preferred?

### 12. **Export Path Resolution** ⚠️ LOW PRIORITY
**Issue**: Export needs to find item by ID, but should also support name lookup.

**Recommendation**:
- `Export()` should try ID first, then fall back to name
- Or: Add `ExportByName(name string)` function
- Document that ID is preferred, name is fallback

### 13. **Testing Strategy Gaps** ⚠️ HIGH PRIORITY
**Issue**: Plan mentions tests but specific test cases needed.

**Recommendation**: Add tests for:
- Import with invalid zip
- Import with missing manifest
- Import with invalid manifest
- Export non-existent item
- Remove non-existent item
- Concurrent imports (if locking added)
- Registry corruption recovery
- Partial install rollback
- ID/name conflict resolution
- Legacy extension discovery

### 14. **Documentation Updates** ⚠️ MEDIUM PRIORITY
**Issue**: Need to update existing docs with new manifest fields.

**Recommendation**:
- Update `docs/EXTENSIONS_SPEC_V0.9.1.md` with new fields
- Update `docs/AGENTICMODULES.md` with module.yaml fields
- Create migration guide for existing extensions
- Add examples showing both old and new manifest formats

### 15. **Backup Strategy** ⚠️ LOW PRIORITY
**Issue**: Plan mentions "keep a backup folder optional" but not implemented.

**Recommendation**:
- On update, create backup: `~/.llamagate/extensions/installed/<id>.backup`
- Keep last N backups (configurable, default 1)
- Add `llamagate restore <id>` command (optional, future enhancement)

## Priority Summary

**Must Fix Before Release**:
1. Remove functionality
2. ID vs Name consistency
3. Extension registry synchronization
4. Startup integration strategy

**Should Fix Soon**:
5. Route cleanup on remove
6. Concurrent operation safety
7. Error recovery & atomicity
8. Manifest schema validation
9. Testing strategy gaps

**Nice to Have**:
10. Hot reload implementation
11. Legacy extension migration
12. CLI command structure improvements
13. Export path resolution
14. Documentation updates
15. Backup strategy

## Implementation Order

1. Fix ID/Name consistency and validation
2. Add Remove functionality
3. Implement startup integration
4. Add extension registry sync
5. Add tests for critical paths
6. Implement remaining features

## Questions for Review

1. **Migration Strategy**: Automatic silent migration or explicit `migrate` command?
2. **CLI Format**: Verbose (`import extension`) or short (`import ext` or `import` with auto-detect)?
3. **Backup**: Implement now or defer to future?
4. **Hot Reload**: Required for v1 or can be deferred?
5. **Concurrent Operations**: File locking or process-level mutex sufficient?
