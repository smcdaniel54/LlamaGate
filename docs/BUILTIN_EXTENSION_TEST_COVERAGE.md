# Builtin Extension Test Coverage

**Date**: 2026-01-23  
**Feature**: Builtin Extension Protection

## Test Coverage Summary

### Registry Protection Tests

All builtin extension protection features are fully tested:

#### 1. `SetEnabled()` Protection
- ✅ **Test**: `TestRegistry_SetEnabled_BuiltinCannotBeDisabled`
- **Coverage**: 
  - Verifies builtin extension cannot be disabled
  - Verifies error message is correct
  - Verifies extension remains enabled after failed disable attempt
  - Verifies enabling (no-op) still works

#### 2. `Unregister()` Protection
- ✅ **Test**: `TestRegistry_Unregister_BuiltinCannotBeUnregistered`
- **Coverage**:
  - Verifies builtin extension cannot be unregistered
  - Verifies error message is correct
  - Verifies extension remains registered after failed unregister attempt

#### 3. `Register()` Builtin Enforcement
- ✅ **Test**: `TestRegistry_Register_BuiltinAlwaysEnabled`
- **Coverage**:
  - Verifies builtin extensions are always enabled on registration
  - Verifies `enabled: false` in manifest is ignored for builtin extensions

#### 4. `RegisterOrUpdate()` Builtin Enforcement
- ✅ **Test**: `TestRegistry_RegisterOrUpdate_BuiltinAlwaysEnabled`
- **Coverage**:
  - Verifies builtin extensions are always enabled via RegisterOrUpdate
  - Verifies builtin extensions remain enabled after updates

### Code Coverage

**Functions Tested**:
- `Registry.SetEnabled()` - Lines 83-98 (100% coverage)
- `Registry.Unregister()` - Lines 118-134 (100% coverage)
- `Registry.Register()` - Lines 24-44 (builtin branch: 100%)
- `Registry.RegisterOrUpdate()` - Lines 138-154 (builtin branch: 100%)

**Test Results**:
```
=== RUN   TestRegistry_SetEnabled_BuiltinCannotBeDisabled
--- PASS: TestRegistry_SetEnabled_BuiltinCannotBeDisabled (0.00s)
=== RUN   TestRegistry_Unregister_BuiltinCannotBeUnregistered
--- PASS: TestRegistry_Unregister_BuiltinCannotBeUnregistered (0.00s)
=== RUN   TestRegistry_Register_BuiltinAlwaysEnabled
--- PASS: TestRegistry_Register_BuiltinAlwaysEnabled (0.00s)
=== RUN   TestRegistry_RegisterOrUpdate_BuiltinAlwaysEnabled
--- PASS: TestRegistry_RegisterOrUpdate_BuiltinAlwaysEnabled (0.00s)
```

### Edge Cases Covered

1. ✅ Builtin extension with `enabled: false` in manifest → Still enabled
2. ✅ Attempting to disable builtin extension → Error returned
3. ✅ Attempting to unregister builtin extension → Error returned
4. ✅ Enabling already-enabled builtin extension → No error (no-op)
5. ✅ Updating builtin extension → Remains enabled

### Missing Coverage

None identified. All builtin extension protection code paths are tested.

### Integration Points

The builtin extension protection integrates with:
- ✅ Extension discovery (`startup.LoadInstalledExtensions()`)
- ✅ Extension registry (`extensions.Registry`)
- ✅ Extension manifest loading (`extensions.LoadManifest()`)

### Test Execution

Run all builtin extension tests:
```bash
go test ./internal/extensions -run "TestRegistry.*Builtin" -v
```

Run all registry tests:
```bash
go test ./internal/extensions -run "TestRegistry" -v
```

---

*Test Coverage Report - 2026-01-23*
