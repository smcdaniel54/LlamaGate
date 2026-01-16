# LlamaGate Extensions v0.9.1 – Implementation Plan

**Version:** 0.9.1  
**Status:** ✅ **COMPLETE - Released v0.9.1 (2026-01-15)**  
**Reference:** [EXTENSIONS_SPEC_V0.9.1.md](./EXTENSIONS_SPEC_V0.9.1.md)

---

## Overview

This document provides a step-by-step implementation plan for migrating LlamaGate from the plugin system to the extension system. The migration is a **breaking change** and must be completed in a single Pull Request.

---

## Implementation Phases

### Phase 1: Preparation & Analysis

**Goal:** Understand current state and prepare for migration

- [ ] **1.1** Audit all code for "plugin" references
  - [ ] Search codebase for `plugin` (case-insensitive)
  - [ ] Search codebase for `Plugin` (case-sensitive)
  - [ ] Search codebase for `plugins` (case-insensitive)
  - [ ] Document all occurrences in `MIGRATION_AUDIT.md`

- [ ] **1.2** Identify all plugin-related files
  - [ ] List all files in `internal/plugins/`
  - [ ] List all files in `plugins/`
  - [ ] List all test files referencing plugins
  - [ ] List all documentation files mentioning plugins

- [ ] **1.3** Create backup branch
  - [ ] Create branch: `backup/pre-extension-migration`
  - [ ] Tag current state: `v0.9.0-final`

- [ ] **1.4** Verify test coverage
  - [ ] Run all tests: `go test ./...`
  - [ ] Document test failures (if any)
  - [ ] Ensure all tests pass before migration

---

### Phase 2: Core Type System Migration

**Goal:** Rename core types and interfaces from Plugin to Extension

- [ ] **2.1** Rename package `internal/plugins` → `internal/extensions`
  - [ ] Create new directory: `internal/extensions/`
  - [ ] Move all files from `internal/plugins/` to `internal/extensions/`
  - [ ] Update package declaration in all files

- [ ] **2.2** Rename core types
  - [ ] `Plugin` interface → `Extension` interface
  - [ ] `PluginMetadata` → `ExtensionMetadata`
  - [ ] `PluginResult` → `ExtensionResult`
  - [ ] `PluginContext` → `ExtensionContext`
  - [ ] `PluginDefinition` → `ExtensionDefinition`
  - [ ] `Registry` → `ExtensionRegistry` (or keep as `Registry` in extensions package)
  - [ ] `ExtendedPlugin` → `ExtendedExtension`
  - [ ] `DefinitionBasedPlugin` → `DefinitionBasedExtension`

- [ ] **2.3** Update method names
  - [ ] `Metadata()` → Keep as `Metadata()` (no change needed)
  - [ ] `ValidateInput()` → Keep as `ValidateInput()`
  - [ ] `Execute()` → Keep as `Execute()`
  - [ ] `GetAPIEndpoints()` → Keep as `GetAPIEndpoints()`
  - [ ] `GetAgentDefinition()` → Keep as `GetAgentDefinition()`

- [ ] **2.4** Update all imports
  - [ ] Find all `import "github.com/llamagate/llamagate/internal/plugins"`
  - [ ] Replace with `import "github.com/llamagate/llamagate/internal/extensions"`
  - [ ] Update all references to types

---

### Phase 3: Configuration System Migration

**Goal:** Update configuration from plugins to extensions

- [ ] **3.1** Update config types (`internal/config/config.go`)
  - [ ] `PluginsConfig` → `ExtensionsConfig`
  - [ ] `cfg.Plugins` → `cfg.Extensions`
  - [ ] `loadPluginsConfig()` → `loadExtensionsConfig()`

- [ ] **3.2** Update environment variable names
  - [ ] `PLUGINS_ENABLED` → `EXTENSIONS_ENABLED`
  - [ ] `PLUGIN_<NAME>_<KEY>` → `EXTENSION_<NAME>_<KEY>`
  - [ ] Update config file keys: `plugins.configs` → `extensions.configs`

- [ ] **3.3** Update config validation
  - [ ] Update validation logic for extensions
  - [ ] Update error messages

- [ ] **3.4** Update example config files
  - [ ] `mcp-config.example.yaml` (if it references plugins)
  - [ ] `mcp-demo-config.yaml` (if it references plugins)
  - [ ] Any other config examples

---

### Phase 4: API Layer Migration

**Goal:** Update HTTP API endpoints and handlers

- [ ] **4.1** Rename API package files
  - [ ] `internal/api/plugins.go` → `internal/api/extensions.go`
  - [ ] `internal/api/plugin_routes.go` → `internal/api/extension_routes.go`
  - [ ] `internal/api/plugin_handler_test.go` → `internal/api/extension_handler_test.go`

- [ ] **4.2** Update API handler types
  - [ ] `PluginHandler` → `ExtensionHandler`
  - [ ] `NewPluginHandler()` → `NewExtensionHandler()`
  - [ ] `ListPlugins()` → `ListExtensions()`
  - [ ] `GetPlugin()` → `GetExtension()`
  - [ ] `ExecutePlugin()` → `ExecuteExtension()`

- [ ] **4.3** Update API routes (`cmd/llamagate/main.go`)
  - [ ] `/v1/plugins` → `/v1/extensions`
  - [ ] `/v1/plugins/:name` → `/v1/extensions/:name`
  - [ ] `/v1/plugins/:name/execute` → `/v1/extensions/:name/execute`
  - [ ] Update route group: `pluginsGroup` → `extensionsGroup`
  - [ ] Update route registration: `RegisterPluginRoutes()` → `RegisterExtensionRoutes()`

- [ ] **4.4** Update API documentation
  - [ ] `docs/API.md` – Update all plugin endpoints to extension endpoints
  - [ ] Update request/response examples
  - [ ] Update OpenAPI/Swagger definitions (if any)

---

### Phase 5: Setup & Registration Migration

**Goal:** Update extension registration and setup

- [ ] **5.1** Update setup package (`internal/setup/`)
  - [ ] `plugins.go` → `extensions.go`
  - [ ] `alexa_plugin.go` → `alexa_extension.go` (or keep as setup helper)
  - [ ] `RegisterTestPlugins()` → `RegisterTestExtensions()`
  - [ ] `RegisterAlexaPlugin()` → `RegisterAlexaExtension()`

- [ ] **5.2** Update main.go registration
  - [ ] `pluginRegistry` → `extensionRegistry`
  - [ ] `plugins.NewRegistry()` → `extensions.NewRegistry()`
  - [ ] Update all registration calls
  - [ ] Update environment variable: `ENABLE_TEST_PLUGINS` → `ENABLE_TEST_EXTENSIONS`

- [ ] **5.3** Update proxy integration
  - [ ] `internal/proxy/plugin_handler.go` → `internal/proxy/extension_handler.go` (if exists)
  - [ ] Update any plugin-related proxy code

---

### Phase 6: Extension Discovery & YAML Support

**Goal:** Implement YAML manifest-based extension discovery

- [ ] **6.1** Create extension discovery system
  - [ ] Create `internal/extensions/discovery.go`
  - [ ] Implement `DiscoverExtensions(dir string) ([]*ExtensionDefinition, error)`
  - [ ] Scan `extensions/` directory for `manifest.yaml` files
  - [ ] Parse YAML manifests
  - [ ] Validate manifests against schema

- [ ] **6.2** Implement YAML manifest parser
  - [ ] Add YAML dependency: `gopkg.in/yaml.v3`
  - [ ] Create `ParseManifest(data []byte) (*ExtensionDefinition, error)`
  - [ ] Create `LoadManifestFromFile(path string) (*ExtensionDefinition, error)`
  - [ ] Handle YAML parsing errors gracefully

- [ ] **6.3** Implement manifest validation
  - [ ] Create `ValidateManifest(def *ExtensionDefinition) error`
  - [ ] Validate required fields: name, version, description
  - [ ] Validate name format (alphanumeric + underscore)
  - [ ] Validate version format (semver)
  - [ ] Validate workflow steps (if present)
  - [ ] Validate endpoints (if present)

- [ ] **6.4** Integrate discovery into startup
  - [ ] Call `DiscoverExtensions("extensions/")` in `main.go`
  - [ ] Register discovered extensions
  - [ ] Log discovery results
  - [ ] Handle discovery errors (non-fatal)

- [ ] **6.5** Implement enable/disable support
  - [ ] Check `enabled` field in manifest
  - [ ] Check config file: `extensions.configs.<name>.enabled`
  - [ ] Check environment: `EXTENSION_<NAME>_ENABLED`
  - [ ] Skip execution if disabled (return 503)
  - [ ] Log disabled extensions

---

### Phase 7: Directory Structure Migration

**Goal:** Move from plugins/ to extensions/ directory

- [ ] **7.1** Create extensions directory structure
  - [ ] Create `extensions/` directory (if doesn't exist)
  - [ ] Create example extension structure
  - [ ] Document directory structure in README

- [ ] **7.2** Migrate existing plugins to extensions
  - [ ] Convert `plugins/alexa_skill.go` to YAML manifest (if keeping)
  - [ ] Create `extensions/alexa_skill/manifest.yaml`
  - [ ] Extract metadata, schemas, workflows to YAML
  - [ ] Move any executable logic (if needed)

- [ ] **7.3** Update templates
  - [ ] `plugins/templates/` → `extensions/templates/` (or remove if YAML-only)
  - [ ] Create YAML manifest template
  - [ ] Update template documentation

- [ ] **7.4** Update examples
  - [ ] `plugins/examples/` → Convert to YAML examples or remove
  - [ ] Create example YAML manifests
  - [ ] Document example extensions

---

### Phase 8: Documentation Migration

**Goal:** Update all documentation from plugins to extensions

- [ ] **8.1** Update main documentation files
  - [x] `docs/PLUGINS.md` → Removed (plugin system deleted)
  - [ ] Replace all "plugin" with "extension"
  - [ ] Update API endpoint examples
  - [ ] Update code examples
  - [ ] Update directory references

- [ ] **8.2** Update quickstart guides
  - [x] `docs/PLUGIN_QUICKSTART.md` → Removed (plugin system deleted)
  - [ ] Update step-by-step instructions
  - [ ] Update YAML examples

- [ ] **8.3** Update architecture documentation
  - [ ] `docs/ARCHITECTURE.md` – Update plugin section to extension section
  - [ ] Update diagrams (if any)
  - [ ] Update component descriptions

- [ ] **8.4** Update README files
  - [ ] `plugins/README.md` → `extensions/README.md`
  - [ ] `README.md` (main) – Update plugin references
  - [ ] Update any other README files

- [ ] **8.5** Create migration guide
  - [ ] Create `docs/MIGRATION_V0.9.1.md`
  - [ ] Document breaking changes
  - [ ] Provide step-by-step migration instructions
  - [ ] Include code examples
  - [ ] Include YAML conversion examples

- [ ] **8.6** Update CHANGELOG
  - [ ] Add v0.9.1 entry
  - [ ] Document breaking changes
  - [ ] List migration steps
  - [ ] Document new features

---

### Phase 9: Test Migration

**Goal:** Update all tests to use extensions

- [ ] **9.1** Update test files
  - [ ] `internal/extensions/*_test.go` – Update all test files
  - [ ] `internal/api/*_test.go` – Update extension handler tests
  - [ ] `tests/plugins/` → `tests/extensions/` (or update)
  - [ ] Update all test plugin references

- [ ] **9.2** Update test helpers
  - [ ] `CreateTestPlugins()` → `CreateTestExtensions()`
  - [ ] Update test plugin creation to use YAML manifests
  - [ ] Update test utilities

- [ ] **9.3** Run all tests
  - [ ] `go test ./...` – Ensure all tests pass
  - [ ] Fix any test failures
  - [ ] Update test expectations if needed

- [ ] **9.4** Update integration tests
  - [ ] Update API endpoint tests
  - [ ] Update extension discovery tests
  - [ ] Update extension execution tests

---

### Phase 10: Scripts & Tools Migration

**Goal:** Update scripts and tooling

- [ ] **10.1** Update shell scripts
  - [ ] `scripts/unix/test-plugins.sh` → `scripts/unix/test-extensions.sh`
  - [ ] Update script content
  - [ ] Update Windows scripts: `scripts/windows/test-plugins.cmd` → `scripts/windows/test-extensions.cmd`

- [ ] **10.2** Update demo scripts
  - [ ] Update any demo scripts referencing plugins
  - [ ] Update API endpoint URLs in scripts

- [ ] **10.3** Update CI/CD
  - [ ] `.github/workflows/ci.yml` – Update test commands
  - [ ] Update any plugin-related CI steps

---

### Phase 11: Code Cleanup

**Goal:** Remove all plugin references and clean up code

- [ ] **11.1** Remove old plugin code
  - [ ] Delete `internal/plugins/` directory (after migration)
  - [ ] Delete `plugins/` directory (after conversion to extensions/)
  - [ ] Remove any unused plugin-related code

- [ ] **11.2** Update comments and docstrings
  - [ ] Search for "plugin" in comments
  - [ ] Replace with "extension"
  - [ ] Update function documentation

- [ ] **11.3** Update error messages
  - [ ] Update all error messages mentioning "plugin"
  - [ ] Update log messages
  - [ ] Update user-facing messages

- [ ] **11.4** Update variable names
  - [ ] Search for variable names containing "plugin"
  - [ ] Rename to "extension" equivalents
  - [ ] Ensure consistency

---

### Phase 12: Final Validation

**Goal:** Ensure everything works and migration is complete

- [ ] **12.1** Build verification
  - [ ] `go build ./...` – Ensure code compiles
  - [ ] Fix any compilation errors
  - [ ] Ensure no import errors

- [ ] **12.2** Test verification
  - [ ] Run full test suite: `go test ./... -v`
  - [ ] Run integration tests
  - [ ] Verify all tests pass

- [ ] **12.3** Manual testing
  - [ ] Start server: `./llamagate`
  - [ ] Test extension discovery
  - [ ] Test extension execution: `POST /v1/extensions/:name/execute`
  - [ ] Test extension listing: `GET /v1/extensions`
  - [ ] Test enable/disable functionality
  - [ ] Test with invalid manifests (error handling)

- [ ] **12.4** Documentation verification
  - [ ] Verify all docs updated
  - [ ] Check for broken links
  - [ ] Verify examples work
  - [ ] Verify migration guide is complete

- [ ] **12.5** Code review checklist
  - [ ] No "plugin" references in code (except in comments explaining migration)
  - [ ] No "Plugin" type names
  - [ ] All imports updated
  - [ ] All tests passing
  - [ ] Documentation updated
  - [ ] CHANGELOG updated

---

## Implementation Order

**Recommended sequence:**

1. **Phase 1** – Preparation (do first)
2. **Phase 2** – Core types (foundation)
3. **Phase 3** – Configuration (needed by other phases)
4. **Phase 4** – API layer (depends on Phase 2)
5. **Phase 5** – Setup (depends on Phase 2)
6. **Phase 6** – Discovery (new feature, can be done in parallel)
7. **Phase 7** – Directory structure (can be done in parallel)
8. **Phase 8** – Documentation (ongoing, but complete before Phase 12)
9. **Phase 9** – Tests (update as you go, complete before Phase 12)
10. **Phase 10** – Scripts (can be done in parallel)
11. **Phase 11** – Cleanup (do after all other phases)
12. **Phase 12** – Validation (final step)

---

## Risk Mitigation

### High-Risk Areas

1. **Extension Discovery** – New feature, needs thorough testing
   - Mitigation: Implement with extensive error handling
   - Test with invalid manifests, missing files, etc.

2. **YAML Parsing** – New dependency, potential parsing issues
   - Mitigation: Use well-tested library (`gopkg.in/yaml.v3`)
   - Add comprehensive error handling

3. **Breaking Changes** – Users must migrate
   - Mitigation: Clear migration guide, detailed CHANGELOG
   - Provide examples and step-by-step instructions

4. **Test Coverage** – Ensure no regressions
   - Mitigation: Run full test suite after each phase
   - Add new tests for extension discovery

### Rollback Plan

If critical issues are discovered:

1. Revert to backup branch: `backup/pre-extension-migration`
2. Tag as `v0.9.0-hotfix`
3. Document issues in GitHub issue
4. Plan fix and re-attempt migration

---

## Success Criteria

Migration is complete when:

- ✅ All code compiles without errors
- ✅ All tests pass
- ✅ No "plugin" references in code (except migration comments)
- ✅ Extension discovery works
- ✅ Extensions can be executed via API
- ✅ Enable/disable functionality works
- ✅ All documentation updated
- ✅ Migration guide created
- ✅ CHANGELOG updated
- ✅ Manual testing successful

---

## Estimated Effort

**Rough estimates:**

- Phase 1 (Preparation): 2-4 hours
- Phase 2 (Core Types): 4-6 hours
- Phase 3 (Configuration): 2-3 hours
- Phase 4 (API Layer): 3-4 hours
- Phase 5 (Setup): 2-3 hours
- Phase 6 (Discovery): 8-12 hours (new feature)
- Phase 7 (Directory): 2-4 hours
- Phase 8 (Documentation): 6-8 hours
- Phase 9 (Tests): 4-6 hours
- Phase 10 (Scripts): 1-2 hours
- Phase 11 (Cleanup): 2-3 hours
- Phase 12 (Validation): 4-6 hours

**Total: ~40-60 hours** (1-2 weeks for one developer)

---

## Next Steps

1. Review this implementation plan
2. Assign phases to developers (if team)
3. Create GitHub issue/PR for tracking
4. Begin Phase 1 (Preparation)
5. Execute phases in order
6. Complete Phase 12 (Validation)
7. Merge PR and release v0.9.1

---

**Status:** Ready for implementation  
**Last Updated:** 2026-01-10
