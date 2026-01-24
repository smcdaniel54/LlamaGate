# Accommodating YAML "Builtin" Extensions

**Purpose**: Explain how to treat YAML extensions as "builtin" (core functionality)  
**Status**: Architecture Design  
**Question**: How do we accommodate a YAML builtin extension?

---

## Understanding the Current Architecture

### Two Extension Systems

LlamaGate has two separate extension systems:

1. **Builtin Extensions (Go Code)**:
   - Location: `internal/extensions/builtin/`
   - Registry: `core.Registry` (for framework extensions)
   - Registration: Manual Go code registration
   - Purpose: Core system functionality (validation, tools, state, etc.)

2. **YAML Extensions (Workflow)**:
   - Location: `extensions/` directory (or `~/.llamagate/extensions/installed/`)
   - Registry: `extensions.Registry` (for workflow extensions)
   - Discovery: Automatic via `DiscoverExtensions()`
   - Purpose: Workflow capabilities (agenticmodule_runner, etc.)

### Current Discovery Flow

```go
// In main.go
extensionRegistry := extensions.NewRegistry()
extensionBaseDir := "extensions"

// Load installed extensions (from ~/.llamagate/extensions/installed/) 
// and legacy extensions (from extensions/)
loadedCount, failures := startup.LoadInstalledExtensions(extensionRegistry, extensionBaseDir)
```

**YAML extensions are discovered from**:
1. `~/.llamagate/extensions/installed/` (installed extensions)
2. `extensions/` (legacy/repo extensions)

---

## Options for YAML "Builtin" Extensions

### Option 1: Special Directory for Builtin YAML Extensions

**Approach**: Create a dedicated directory for YAML extensions that should be treated as "builtin"

**Implementation**:
```
LlamaGate/
├── internal/
│   └── extensions/
│       └── builtin/
│           └── yaml/              # NEW: YAML builtin extensions
│               └── extension-doc-generator/
│                   └── manifest.yaml
└── extensions/                    # Regular default extensions
    └── agenticmodule_runner/
```

**Code Changes Needed**:
```go
// In startup.go or main.go
func LoadBuiltinYAMLExtensions(extRegistry *extensions.Registry) (int, []string) {
    builtinYAMLDir := "internal/extensions/builtin/yaml"
    manifests, err := extensions.DiscoverExtensions(builtinYAMLDir)
    // ... load manifests
}
```

**Pros**:
- ✅ Clear separation (builtin YAML vs default YAML)
- ✅ Loaded before regular extensions
- ✅ Can't be disabled (if desired)
- ✅ Clear intent: "this is builtin"

**Cons**:
- ❌ Requires code changes
- ❌ New directory structure
- ❌ May confuse "builtin" (Go) vs "builtin YAML"

---

### Option 2: Priority Loading with Special Flag

**Approach**: Add a `builtin: true` flag to YAML manifest, load with priority

**Implementation**:
```yaml
# extensions/extension-doc-generator/manifest.yaml
name: extension-doc-generator
version: 1.0.0
builtin: true          # NEW: Mark as builtin
type: workflow
enabled: true
```

**Code Changes Needed**:
```go
// In startup.go
func LoadInstalledExtensions(extRegistry *extensions.Registry, legacyBaseDir string) (int, []string) {
    // First, load builtin YAML extensions
    builtinManifests := []*Manifest{}
    regularManifests := []*Manifest{}
    
    for _, manifest := range allManifests {
        if manifest.Builtin {
            builtinManifests = append(builtinManifests, manifest)
        } else {
            regularManifests = append(regularManifests, manifest)
        }
    }
    
    // Load builtin first
    for _, manifest := range builtinManifests {
        // Load with priority, can't be disabled
    }
    
    // Then load regular extensions
    for _, manifest := range regularManifests {
        // Normal loading
    }
}
```

**Pros**:
- ✅ No new directory structure
- ✅ Clear flag in manifest
- ✅ Can control loading order
- ✅ Minimal code changes

**Cons**:
- ❌ Requires manifest schema change
- ❌ Requires code changes
- ❌ Still in `extensions/` directory

---

### Option 3: Separate Discovery Path for Builtin YAML

**Approach**: Discover builtin YAML extensions from a specific subdirectory

**Implementation**:
```
LlamaGate/
└── extensions/
    ├── builtin/                    # NEW: Builtin YAML extensions
    │   └── extension-doc-generator/
    │       └── manifest.yaml
    └── agenticmodule_runner/        # Regular default extensions
        └── manifest.yaml
```

**Code Changes Needed**:
```go
// In startup.go
func LoadInstalledExtensions(extRegistry *extensions.Registry, legacyBaseDir string) (int, []string) {
    // Load builtin YAML extensions first
    builtinDir := filepath.Join(legacyBaseDir, "builtin")
    builtinManifests, _ := extensions.DiscoverExtensions(builtinDir)
    
    // Load them with special handling
    for _, manifest := range builtinManifests {
        // Mark as builtin, load with priority
        manifest.Builtin = true
        // ... register
    }
    
    // Then load regular extensions
    regularManifests, _ := extensions.DiscoverExtensions(legacyBaseDir)
    // ... filter out builtin/ subdirectory
}
```

**Pros**:
- ✅ Clear directory structure
- ✅ Easy to identify builtin YAML extensions
- ✅ Can load with priority
- ✅ No manifest schema change needed

**Cons**:
- ❌ Requires code changes
- ❌ New subdirectory structure
- ❌ Need to filter `builtin/` from regular discovery

---

### Option 4: Documentation-Only "Builtin" Status

**Approach**: Keep YAML extensions in `extensions/`, but document some as "builtin"

**Implementation**:
- No code changes
- Extensions in `extensions/` directory
- Document certain extensions as "builtin" in README
- Loaded normally via `DiscoverExtensions()`

**Documentation**:
```markdown
## Extension Types

### Builtin Extensions (Go Code)
- Core functionality compiled into binary
- Location: `internal/extensions/builtin/`
- Examples: `validation`, `tools`, `state`

### Builtin Extensions (YAML-based)
- Core workflow capabilities included in repo
- Location: `extensions/builtin/` or `extensions/` with `builtin: true`
- Examples: `extension-doc-generator`, `agenticmodule_runner`

### Default Extensions (YAML-based)
- Workflow extensions included in repo
- Location: `extensions/` directory
- Examples: `prompt-template-executor`
```

**Pros**:
- ✅ No code changes needed
- ✅ Clear documentation
- ✅ Flexible

**Cons**:
- ❌ No runtime distinction
- ❌ Can't enforce "can't disable" behavior
- ❌ Purely documentation-based

---

## Recommended Approach

### Hybrid: Option 3 + Option 2

**Best of both worlds**:

1. **Directory Structure**: `extensions/builtin/` for builtin YAML extensions
2. **Manifest Flag**: `builtin: true` flag for clarity
3. **Priority Loading**: Load builtin YAML extensions first
4. **Documentation**: Clear distinction in docs

**Implementation**:

```
LlamaGate/
└── extensions/
    ├── builtin/                    # Builtin YAML extensions
    │   └── extension-doc-generator/
    │       └── manifest.yaml       # builtin: true
    └── agenticmodule_runner/        # Regular default extensions
        └── manifest.yaml
```

**Code Changes**:
```go
// In startup.go
func LoadInstalledExtensions(extRegistry *extensions.Registry, legacyBaseDir string) (int, []string) {
    // 1. Load builtin YAML extensions first
    builtinDir := filepath.Join(legacyBaseDir, "builtin")
    if builtinManifests, err := extensions.DiscoverExtensions(builtinDir); err == nil {
        for _, manifest := range builtinManifests {
            manifest.Builtin = true  // Set flag
            // Load with priority, can't be disabled
            // ... register extension
        }
    }
    
    // 2. Load regular extensions (excluding builtin/ subdirectory)
    regularManifests, _ := extensions.DiscoverExtensions(legacyBaseDir)
    // Filter out any in builtin/ subdirectory
    // ... register extensions
}
```

**Benefits**:
- ✅ Clear directory structure (`extensions/builtin/`)
- ✅ Manifest flag for clarity (`builtin: true`)
- ✅ Priority loading (builtin loaded first)
- ✅ Can enforce "can't disable" if needed
- ✅ Easy to identify builtin YAML extensions
- ✅ Minimal code changes

---

## Implementation Steps

### Step 1: Add Builtin Flag to Manifest Schema

```go
// In internal/extensions/manifest.go
type Manifest struct {
    Name        string                 `yaml:"name"`
    Version     string                 `yaml:"version"`
    Description string                 `yaml:"description"`
    Type        string                 `yaml:"type"`
    Builtin     bool                   `yaml:"builtin,omitempty"`  // NEW
    Enabled     bool                   `yaml:"enabled,omitempty"`
    // ... rest of fields
}
```

### Step 2: Update Discovery to Handle Builtin Directory

```go
// In internal/startup/startup.go
func LoadInstalledExtensions(extRegistry *extensions.Registry, legacyBaseDir string) (int, []string) {
    loadedCount := 0
    failures := []string{}
    
    // Load builtin YAML extensions first
    builtinDir := filepath.Join(legacyBaseDir, "builtin")
    if builtinManifests, err := extensions.DiscoverExtensions(builtinDir); err == nil {
        for _, manifest := range builtinManifests {
            manifest.Builtin = true
            // Load with special handling (can't disable, priority, etc.)
            if err := extRegistry.RegisterOrUpdate(manifest); err != nil {
                failures = append(failures, fmt.Sprintf("%s: %v", manifest.Name, err))
            } else {
                loadedCount++
            }
        }
    }
    
    // Load regular extensions (excluding builtin/)
    // ... existing logic
}
```

### Step 3: Update Documentation

```markdown
## Extension Types

### Builtin Extensions (Go Code)
- Core functionality compiled into binary
- Location: `internal/extensions/builtin/`
- Examples: `validation`, `tools`, `state`

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

---

## Summary

**Question**: How do we accommodate a YAML builtin?

**Answer**: Use **Option 3 (Hybrid)**:
1. **Directory**: `extensions/builtin/` for builtin YAML extensions
2. **Flag**: `builtin: true` in manifest for clarity
3. **Loading**: Priority loading in startup code
4. **Documentation**: Clear distinction in docs

**Benefits**:
- ✅ Clear separation (builtin YAML vs default YAML)
- ✅ Priority loading
- ✅ Can enforce special behavior (can't disable, etc.)
- ✅ Easy to identify
- ✅ Minimal code changes

**For `extension-doc-generator`**:
- Place in `extensions/builtin/extension-doc-generator/`
- Add `builtin: true` to manifest
- Document as "Builtin Extension (YAML-based)"

---

*Architecture Design - 2026-01-24*
