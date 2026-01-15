# LlamaGate Extensions v0.9.1 – Migration Checklist

Quick reference checklist for tracking migration progress.

---

## Phase 1: Preparation

- [ ] Audit all "plugin" references in codebase
- [ ] Document findings in `MIGRATION_AUDIT.md`
- [ ] Create backup branch: `backup/pre-extension-migration`
- [ ] Tag current state: `v0.9.0-final`
- [ ] Run full test suite and verify all tests pass

---

## Phase 2: Core Types

- [ ] Rename package: `internal/plugins` → `internal/extensions`
- [ ] Rename `Plugin` → `Extension`
- [ ] Rename `PluginMetadata` → `ExtensionMetadata`
- [ ] Rename `PluginResult` → `ExtensionResult`
- [ ] Rename `PluginContext` → `ExtensionContext`
- [ ] Rename `PluginDefinition` → `ExtensionDefinition`
- [ ] Rename `Registry` → `ExtensionRegistry` (or keep as `Registry`)
- [ ] Rename `ExtendedPlugin` → `ExtendedExtension`
- [ ] Update all imports
- [ ] Verify code compiles

---

## Phase 3: Configuration

- [ ] Rename `PluginsConfig` → `ExtensionsConfig`
- [ ] Rename `cfg.Plugins` → `cfg.Extensions`
- [ ] Rename `loadPluginsConfig()` → `loadExtensionsConfig()`
- [ ] Update `PLUGINS_ENABLED` → `EXTENSIONS_ENABLED`
- [ ] Update `PLUGIN_<NAME>_<KEY>` → `EXTENSION_<NAME>_<KEY>`
- [ ] Update config file keys: `plugins.configs` → `extensions.configs`
- [ ] Update example config files
- [ ] Verify config loading works

---

## Phase 4: API Layer

- [ ] Rename `internal/api/plugins.go` → `internal/api/extensions.go`
- [ ] Rename `internal/api/plugin_routes.go` → `internal/api/extension_routes.go`
- [ ] Rename `PluginHandler` → `ExtensionHandler`
- [ ] Rename `ListPlugins()` → `ListExtensions()`
- [ ] Rename `GetPlugin()` → `GetExtension()`
- [ ] Rename `ExecutePlugin()` → `ExecuteExtension()`
- [ ] Update routes: `/v1/plugins` → `/v1/extensions`
- [ ] Update `docs/API.md`
- [ ] Test API endpoints

---

## Phase 5: Setup & Registration

- [ ] Rename `internal/setup/plugins.go` → `internal/setup/extensions.go`
- [ ] Rename `RegisterTestPlugins()` → `RegisterTestExtensions()`
- [ ] Rename `RegisterAlexaPlugin()` → `RegisterAlexaExtension()`
- [ ] Update `main.go` registration code
- [ ] Update `ENABLE_TEST_PLUGINS` → `ENABLE_TEST_EXTENSIONS`
- [ ] Update proxy integration (if any)

---

## Phase 6: Extension Discovery

- [ ] Create `internal/extensions/discovery.go`
- [ ] Implement `DiscoverExtensions()` function
- [ ] Add YAML dependency: `gopkg.in/yaml.v3`
- [ ] Implement `ParseManifest()` function
- [ ] Implement `LoadManifestFromFile()` function
- [ ] Implement `ValidateManifest()` function
- [ ] Integrate discovery into `main.go` startup
- [ ] Implement enable/disable support
- [ ] Test discovery with valid/invalid manifests

---

## Phase 7: Directory Structure

- [ ] Create `extensions/` directory
- [ ] Convert `plugins/alexa_skill.go` to YAML manifest (if keeping)
- [ ] Create example extension structure
- [ ] Update templates (or remove if YAML-only)
- [ ] Create example YAML manifests

---

## Phase 8: Documentation

- [ ] Rename `docs/PLUGINS.md` → `docs/EXTENSIONS.md`
- [ ] Replace all "plugin" → "extension" in docs
- [ ] Update `docs/PLUGIN_QUICKSTART.md` → `docs/EXTENSION_QUICKSTART.md`
- [ ] Update `docs/ARCHITECTURE.md`
- [ ] Update `plugins/README.md` → `extensions/README.md`
- [ ] Update main `README.md`
- [ ] Create `docs/MIGRATION_V0.9.1.md`
- [ ] Update `CHANGELOG.md`

---

## Phase 9: Tests

- [ ] Update all test files in `internal/extensions/`
- [ ] Update API handler tests
- [ ] Update `tests/plugins/` → `tests/extensions/`
- [ ] Rename `CreateTestPlugins()` → `CreateTestExtensions()`
- [ ] Update test utilities
- [ ] Run full test suite: `go test ./...`
- [ ] Fix any test failures

---

## Phase 10: Scripts

- [ ] Rename `scripts/unix/test-plugins.sh` → `scripts/unix/test-extensions.sh`
- [ ] Rename `scripts/windows/test-plugins.cmd` → `scripts/windows/test-extensions.cmd`
- [ ] Update script content
- [ ] Update demo scripts
- [ ] Update CI/CD workflows

---

## Phase 11: Cleanup

- [ ] Delete `internal/plugins/` directory
- [ ] Delete `plugins/` directory (after conversion)
- [ ] Update all comments mentioning "plugin"
- [ ] Update all error messages
- [ ] Update all log messages
- [ ] Update variable names
- [ ] Final search for "plugin" references

---

## Phase 12: Final Validation

- [ ] Code compiles: `go build ./...`
- [ ] All tests pass: `go test ./...`
- [ ] Manual testing: Start server
- [ ] Test extension discovery
- [ ] Test `GET /v1/extensions`
- [ ] Test `GET /v1/extensions/:name`
- [ ] Test `POST /v1/extensions/:name/execute`
- [ ] Test enable/disable functionality
- [ ] Test with invalid manifests
- [ ] Verify all documentation
- [ ] Code review: No "plugin" references
- [ ] CHANGELOG updated
- [ ] Migration guide complete

---

## Final Sign-off

- [ ] All phases complete
- [ ] All tests passing
- [ ] Documentation complete
- [ ] Code review approved
- [ ] Ready for merge
- [ ] PR created and ready

---

**Status:** ⬜ Not Started | 🟡 In Progress | ✅ Complete
