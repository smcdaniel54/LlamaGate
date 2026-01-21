# CI Linting Strictness Review & Recommendations

**Date:** 2026-01-21  
**Reviewer:** AI Assistant  
**Status:** Recommendations for Improvement

---

## Executive Summary

This document reviews the current CI linting configuration and provides recommendations for optimizing strictness, performance, and developer experience.

---

## Current Configuration Analysis

### CI Workflow (`.github/workflows/ci.yml`)

**Current Setup:**
- **Tool:** golangci-lint v2.8.0
- **Timeout:** 15 minutes
- **Scope:** Production code only (`tests: false`)
- **Arguments:** `--timeout=10m --verbose`
- **No skip-files specified** (relies on `.golangci.yml` config)

### Linter Configuration (`.golangci.yml`)

**Enabled Linters:**
- `errcheck` - Check for unchecked errors
- `govet` - Go vet checks
- `ineffassign` - Detect ineffectual assignments
- `staticcheck` - Advanced static analysis
- `unused` - Detect unused code
- `misspell` - Spelling checker
- `unconvert` - Detect unnecessary conversions
- `unparam` - Detect unused function parameters
- `revive` - Fast, configurable linter (replacement for golint)

**Disabled Linters:**
- `gocritic` - Disabled for performance reasons

**Issue Limits:**
- `max-issues-per-linter: 0` (unlimited)
- `max-same-issues: 0` (unlimited)

**Test Files:**
- `tests: false` - Test files excluded from CI linting

### Local vs CI Differences

**Local Linting (scripts):**
- Includes test files (`--tests` flag)
- Stricter enforcement
- Pre-commit hook runs on staged files only

**CI Linting:**
- Excludes test files (faster)
- Production code focus
- Full project scan

---

## Current Issues Identified

### 1. **Unlimited Issue Counts**
- `max-issues-per-linter: 0` means unlimited issues per linter
- `max-same-issues: 0` means unlimited duplicate issues
- **Impact:** CI can pass with hundreds of issues, defeating the purpose of linting

### 2. **Missing Issue Severity Configuration**
- No distinction between errors and warnings
- All issues treated equally
- **Impact:** Can't prioritize critical issues

### 3. **No Exclude Rules**
- Comment mentions v2.8.0 doesn't support `exclude-rules`
- However, golangci-lint v2.x does support exclude patterns
- **Impact:** Can't exclude known false positives or legacy code

### 4. **Test Files Not Linted in CI**
- Test files excluded for performance
- **Impact:** Test code quality not enforced in CI
- **Mitigation:** Pre-commit hook enforces test linting locally

### 5. **No Linter-Specific Settings**
- No custom rules for individual linters
- **Impact:** Can't fine-tune strictness per linter

### 6. **Performance Concerns**
- `gocritic` disabled for performance
- 15-minute timeout suggests potential slowness
- **Impact:** May need optimization

---

## Recommendations

### 🔴 Critical (Implement Immediately)

#### 1. Set Reasonable Issue Limits

**Current:** Unlimited issues allowed  
**Recommended:** Set limits to enforce quality

```yaml
issues:
  # Maximum issues per linter (prevents one linter from overwhelming output)
  max-issues-per-linter: 50
  
  # Maximum duplicate issues (prevents same error repeated 100+ times)
  max-same-issues: 5
  
  # New files only (gradually improve existing code)
  new: false  # Set to true to only check new code
```

**Rationale:**
- Prevents CI from passing with hundreds of issues
- Focuses attention on most common problems
- Allows gradual improvement of existing code

#### 2. Add Exclude Patterns for Known Issues

**Recommended:** Add exclude patterns for legacy code and false positives

```yaml
issues:
  exclude-rules:
    # Exclude generated files
    - path: _gen\.go$
      linters:
        - errcheck
        - revive
    
    # Exclude test helper files from certain checks
    - path: _test\.go$
      linters:
        - unparam  # Test helpers often have unused params for interface compliance
    
    # Exclude specific known issues in legacy code
    - path: internal/mcpclient/
      text: "Error return value.*is not checked"
      linters:
        - errcheck
```

**Rationale:**
- Allows gradual migration of legacy code
- Reduces noise from known issues
- Focuses on new code quality

### 🟡 High Priority (Implement Soon)

#### 3. Enable Severity-Based Filtering

**Recommended:** Configure severity levels

```yaml
issues:
  # Only show errors, not warnings (for CI)
  severity: error
  
  # Or use exclude-by-msg for specific warnings
  exclude-use-default: false
```

**Rationale:**
- CI should fail on errors, not warnings
- Warnings can be addressed gradually
- Keeps CI focused on critical issues

#### 4. Add Linter-Specific Settings

**Recommended:** Fine-tune individual linters

```yaml
linters-settings:
  revive:
    rules:
      - name: exported
        severity: warning  # Make exported function checks warnings, not errors
      - name: package-comments
        severity: warning
  
  errcheck:
    check-type-assertions: true
    check-blank: true
    exclude-functions:
      - (io.ReadCloser).Close
      - (*os.File).Close
  
  unparam:
    check-exported: false  # Don't check exported functions (interface compliance)
```

**Rationale:**
- Allows fine-tuning per linter
- Reduces false positives
- Better developer experience

#### 5. Add Performance Optimizations

**Recommended:** Optimize for CI speed

```yaml
run:
  timeout: 10m  # Reduce from 15m if possible
  
  # Build cache for faster runs
  build-cache: true
  
  # Skip vendor and generated files
  skip-dirs:
    - vendor
    - .git
    - _gen
  
  # Skip specific files
  skip-files:
    - ".*_gen\\.go$"
    - ".*\\.pb\\.go$"
```

**Rationale:**
- Faster CI feedback
- Better resource utilization
- Still maintains quality

### 🟢 Medium Priority (Consider)

#### 6. Re-enable gocritic with Selective Rules

**Recommended:** Enable gocritic with specific rules only

```yaml
linters:
  enable:
    - gocritic

linters-settings:
  gocritic:
    enabled-tags:
      - performance
      - diagnostic
    disabled-tags:
      - experimental
      - opinionated
    disabled-checks:
      - dupImport  # Can be noisy
      - ifElseChain  # Sometimes necessary
```

**Rationale:**
- Adds valuable checks without performance hit
- Selective rules reduce noise
- Can disable if still too slow

#### 7. Add Test File Linting (Optional)

**Recommended:** Add separate job for test linting

```yaml
# In CI workflow, add:
test-lint:
  name: Lint Tests
  runs-on: ubuntu-latest
  continue-on-error: true  # Non-blocking
  steps:
    - uses: actions/checkout@v6
    - name: Set up Go
      uses: actions/setup-go@v6
      with:
        go-version: '1.24'
    - name: Run golangci-lint on tests
      uses: golangci/golangci-lint-action@v9
      with:
        version: v2.8.0
        args: --timeout=5m --tests
```

**Rationale:**
- Enforces test code quality
- Non-blocking so doesn't slow main CI
- Can be made blocking later

#### 8. Add Issue Count Reporting

**Recommended:** Add summary reporting

```yaml
# In CI workflow, after lint step:
- name: Lint Summary
  if: always()
  run: |
    echo "## Linting Summary" >> $GITHUB_STEP_SUMMARY
    echo "- Total issues: $(grep -c 'issues:' lint-output.txt || echo 0)" >> $GITHUB_STEP_SUMMARY
```

**Rationale:**
- Better visibility into linting status
- Helps track improvement over time
- Useful for PR reviews

---

## Recommended Configuration

### Updated `.golangci.yml`

```yaml
# golangci-lint configuration file
# See https://golangci-lint.run/usage/configuration/

version: "2"

run:
  # Timeout for analysis
  timeout: 10m
  
  # Exclude test files from linting in CI (set to false)
  # This speeds up CI significantly - developers should lint tests locally
  tests: false
  
  # Build tags to use
  build-tags: []
  
  # Skip directories
  skip-dirs:
    - vendor
    - .git
  
  # Skip files
  skip-files:
    - ".*_gen\\.go$"
    - ".*\\.pb\\.go$"

linters:
  # Enable specific linters
  enable:
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - unused
    - misspell
    - unconvert
    - unparam
    - revive
  # Disabled gocritic for performance - can re-enable selectively
  disable:
    - gocritic

linters-settings:
  revive:
    rules:
      - name: package-comments
        severity: warning  # Make package comment checks warnings
      - name: exported
        severity: warning   # Exported function checks as warnings
  
  errcheck:
    check-type-assertions: true
    check-blank: true
    exclude-functions:
      - (io.ReadCloser).Close
      - (*os.File).Close
  
  unparam:
    check-exported: false  # Don't check exported functions (interface compliance)

issues:
  # Maximum issues per linter (prevents one linter from overwhelming output)
  max-issues-per-linter: 50
  
  # Maximum duplicate issues (prevents same error repeated many times)
  max-same-issues: 5
  
  # Exclude rules for known issues
  exclude-rules:
    # Exclude test helper files from unparam (interface compliance)
    - path: _test\.go$
      linters:
        - unparam
    
    # Exclude generated files
    - path: _gen\.go$
      linters:
        - errcheck
        - revive
  
  # Only show errors in CI (warnings can be addressed gradually)
  severity: error
  
  # Exclude use of default excludes (be explicit)
  exclude-use-default: false
```

### Updated CI Workflow (Optional Enhancements)

```yaml
lint:
  name: Lint
  runs-on: ubuntu-latest
  steps:
  - uses: actions/checkout@v6
  
  - name: Set up Go
    uses: actions/setup-go@v6
    with:
      go-version: '1.24'
  
  - name: Download dependencies
    run: go mod download
  
  - name: Run golangci-lint
    uses: golangci/golangci-lint-action@v9
    with:
      version: v2.8.0
      args: --timeout=10m --verbose
    timeout-minutes: 12
  
  - name: Lint Summary
    if: always()
    run: |
      echo "## ✅ Linting Complete" >> $GITHUB_STEP_SUMMARY
      echo "Production code linted successfully" >> $GITHUB_STEP_SUMMARY
```

---

## Implementation Plan

### Phase 1: Critical Fixes (Week 1)
1. ✅ Set `max-issues-per-linter: 50`
2. ✅ Set `max-same-issues: 5`
3. ✅ Add basic exclude rules for test files
4. ✅ Reduce timeout to 10m

### Phase 2: Enhancements (Week 2)
1. Add linter-specific settings
2. Configure severity levels
3. Add skip-dirs and skip-files
4. Test performance improvements

### Phase 3: Optional (Month 1)
1. Consider re-enabling gocritic selectively
2. Add test linting job (non-blocking)
3. Add linting summary reporting
4. Monitor and adjust based on feedback

---

## Migration Strategy

### For Existing Codebase

1. **Start with lenient limits:**
   - `max-issues-per-linter: 100` (temporary)
   - Gradually reduce to 50

2. **Add excludes incrementally:**
   - Start with test files
   - Add legacy code directories as needed
   - Document why each exclude exists

3. **Monitor CI performance:**
   - Track linting time
   - Adjust timeout if needed
   - Consider parallel linting for large repos

4. **Developer communication:**
   - Document changes in CONTRIBUTING.md
   - Provide migration guide
   - Set expectations for new code

---

## Metrics to Track

1. **CI Performance:**
   - Linting time (target: < 5 minutes)
   - CI job duration
   - Resource usage

2. **Code Quality:**
   - Total issues count (trending down)
   - Issues per PR
   - Time to fix linting errors

3. **Developer Experience:**
   - Pre-commit hook effectiveness
   - False positive rate
   - Developer feedback

---

## Conclusion

The current CI linting configuration is **too lenient** with unlimited issues allowed. The recommended changes will:

1. ✅ **Enforce quality** - Set reasonable issue limits
2. ✅ **Improve performance** - Optimize configuration
3. ✅ **Better DX** - Fine-tune linter settings
4. ✅ **Gradual migration** - Exclude legacy code while enforcing new code quality

**Priority:** Implement Phase 1 immediately, Phase 2 within 2 weeks, Phase 3 as needed.

---

## References

- [golangci-lint Documentation](https://golangci-lint.run/)
- [golangci-lint Configuration](https://golangci-lint.run/usage/configuration/)
- [Best Practices](https://golangci-lint.run/usage/configuration/#best-practices)
