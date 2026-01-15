# LlamaGate v0.9.1 Release Checklist

**Release Date:** 2026-01-15  
**Version:** 0.9.1

---

## ✅ Pre-Release Verification

- [x] **Version Updated**: `internal/mcpclient/client.go` updated to 0.9.1
- [x] **All Tests Pass**: 10/10 packages passing
- [x] **Build Successful**: `go build ./cmd/llamagate` succeeds
- [x] **CHANGELOG Updated**: v0.9.1 entry complete with breaking changes
- [x] **Release Notes Created**: `RELEASE_NOTES_v0.9.1.md` created
- [x] **Migration Status**: `docs/MIGRATION_STATUS.md` documents completion
- [x] **Documentation Updated**: Main docs updated (README, ARCHITECTURE, TESTING)

---

## 📋 Release Steps

### 1. Final Verification
```bash
# Run full test suite
go test ./cmd/... ./internal/... -v

# Verify build
go build ./cmd/llamagate

# Check for any remaining plugin references (should be none in core code)
grep -r "internal/plugins" internal/ cmd/ || echo "No plugin references found"
```

### 2. Commit All Changes
```bash
# Review changes
git status

# Stage all changes
git add .

# Commit with release message
git commit -m "Release v0.9.1: Complete plugins to extensions migration

- Remove all plugin system code
- Implement YAML-based extension system
- Update all documentation
- Update version to 0.9.1

Breaking Changes:
- Plugin system completely removed
- /v1/plugins endpoints removed
- PluginsConfig removed

See CHANGELOG.md for full details."
```

### 3. Create Release Tag
```bash
# Create annotated tag
git tag -a v0.9.1 -m "Release v0.9.1

Breaking Changes:
- Plugin system removed, replaced with extension system
- /v1/plugins endpoints removed
- PluginsConfig removed

New Features:
- YAML-based extension system
- Auto-discovery from extensions/ directory
- Support for workflow, middleware, and observer extensions

See RELEASE_NOTES_v0.9.1.md for details."
```

### 4. Push to Remote
```bash
# Push commits
git push origin main  # or your default branch

# Push tags
git push origin v0.9.1
```

### 5. Create GitHub Release (if using GitHub)

1. Go to GitHub repository
2. Click "Releases" → "Draft a new release"
3. **Tag**: Select `v0.9.1`
4. **Title**: `v0.9.1 - Extension System Release`
5. **Description**: Copy from `RELEASE_NOTES_v0.9.1.md`
6. **Mark as**: Pre-release (since this is 0.9.1, not 1.0.0)
7. **Attach binaries** (if building release binaries):
   - `llamagate-windows-amd64.exe`
   - `llamagate-linux-amd64`
   - `llamagate-darwin-amd64`
8. Click "Publish release"

---

## 📦 Optional: Build Release Binaries

If you want to provide pre-built binaries:

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o llamagate-windows-amd64.exe ./cmd/llamagate

# Linux
GOOS=linux GOARCH=amd64 go build -o llamagate-linux-amd64 ./cmd/llamagate

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o llamagate-darwin-amd64 ./cmd/llamagate

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o llamagate-darwin-arm64 ./cmd/llamagate
```

---

## 🔍 Post-Release Verification

After release:

- [ ] Verify tag exists: `git tag -l v0.9.1`
- [ ] Verify GitHub release created (if applicable)
- [ ] Test installation from release
- [ ] Verify extension discovery works
- [ ] Test `/v1/extensions` endpoints
- [ ] Update any CI/CD workflows if needed

---

## 📝 Release Summary

**Version:** 0.9.1  
**Release Type:** Breaking Change Release  
**Migration Required:** Yes (plugins → extensions)

**Key Changes:**
- ✅ Plugin system completely removed
- ✅ Extension system fully implemented
- ✅ All tests passing
- ✅ Documentation updated
- ✅ Version updated in code

**Files Changed:**
- `internal/mcpclient/client.go` - Version updated
- `CHANGELOG.md` - v0.9.1 entry added
- `RELEASE_NOTES_v0.9.1.md` - Release notes created
- All migration-related files updated

---

## 🚀 Ready for Release

All pre-release checks are complete. The codebase is ready for v0.9.1 release.

**Next Steps:**
1. Review the checklist above
2. Commit and tag the release
3. Push to remote repository
4. Create GitHub release (if applicable)
5. Announce the release

---

**Status:** ✅ **READY FOR RELEASE**
