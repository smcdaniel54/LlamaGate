# Linting Philosophy: Is It Necessary?

**Date:** 2026-01-21  
**Purpose:** Address concerns about linting overhead and provide practical guidance

---

## The Question: Why So Much Trouble?

You're right to question this. Linting can feel like unnecessary friction, especially when:
- You're trying to merge a PR
- CI keeps failing on what seem like minor issues
- You have to fix things that "work fine"
- It slows down development

Let's be honest about when linting helps vs. when it's just busywork.

---

## When Linting Is Actually Valuable

### 1. **Catching Real Bugs** ✅

**Examples:**
- **Unchecked errors** (`errcheck`) - Missing error handling can cause silent failures
- **Unused variables** (`unused`) - Dead code that should be removed
- **Ineffectual assignments** (`ineffassign`) - Code that does nothing
- **Static analysis** (`staticcheck`) - Finds logic errors, race conditions, etc.

**Real Impact:** These catch bugs before they reach production.

### 2. **Maintaining Code Quality** ✅

**Examples:**
- **Consistent formatting** (`go fmt`) - Makes code easier to read and review
- **Package comments** - Helps new developers understand code structure
- **Unused parameters** - Clean interfaces, easier refactoring

**Real Impact:** Easier code reviews, faster onboarding, better maintainability.

### 3. **Team Standards** ✅

**Examples:**
- Everyone follows the same style
- Code reviews focus on logic, not style
- Automated enforcement reduces arguments

**Real Impact:** Less time debating style, more time building features.

---

## When Linting Is Just Busywork ❌

### 1. **Overly Strict Style Rules**

**Examples:**
- Package comments for every tiny package
- Unused parameters in interface implementations (required by interface)
- Formatting nitpicks that don't affect functionality

**Impact:** Wastes time on cosmetic issues.

### 2. **False Positives**

**Examples:**
- Linter flags something that's intentional
- Test helpers that need unused parameters for interface compliance
- Placeholder code that will be implemented later

**Impact:** Developers learn to ignore linting, defeating the purpose.

### 3. **Legacy Code**

**Examples:**
- Existing codebase with hundreds of issues
- Can't fix everything at once
- New code gets blocked by old code issues

**Impact:** Prevents progress, creates frustration.

---

## Current Situation Analysis

### What's Happening Now

1. **We just made linting stricter** - Added issue limits (50 per linter, 5 duplicates)
2. **This is catching existing issues** - Not just new code
3. **Some are legitimate** - Unchecked errors, unused code
4. **Some are style-only** - Package comments, unused parameters in interfaces

### The Real Problem

The trouble isn't linting itself - it's that we're trying to fix everything at once on a PR that's adding new functionality. This creates friction.

---

## Practical Solutions

### Option 1: **Relax for This PR** (Quick Fix)

Make linting less strict temporarily:

```yaml
issues:
  max-issues-per-linter: 100  # More lenient
  max-same-issues: 10          # More lenient
```

**Pros:** Unblocks merging immediately  
**Cons:** Allows more issues to accumulate

### Option 2: **Focus on Critical Issues Only** (Recommended)

Only enforce linting rules that catch bugs:

```yaml
linters:
  enable:
    - errcheck      # ✅ Catches bugs
    - govet         # ✅ Catches bugs
    - staticcheck   # ✅ Catches bugs
    - ineffassign   # ✅ Catches bugs
  disable:
    - unused        # ⚠️ Style only
    - unparam       # ⚠️ Style only (interface compliance issues)
    - revive        # ⚠️ Style only (package comments, etc.)
    - misspell      # ⚠️ Style only
    - unconvert     # ⚠️ Style only
```

**Pros:** Catches real bugs, ignores style  
**Cons:** Less consistent code style

### Option 3: **Gradual Improvement** (Best Long-term)

1. **Fix critical issues now** (unchecked errors, bugs)
2. **Fix style issues gradually** (one PR at a time)
3. **Exclude examples/legacy code** from strict linting
4. **Enforce strict linting only on new code**

**Pros:** Maintains quality, reduces friction  
**Cons:** Requires discipline

### Option 4: **Disable Linting for Examples** (Pragmatic)

Examples don't need production-level linting:

```yaml
# In CI workflow, exclude examples:
args: --timeout=10m --skip-dirs=internal/examples
```

**Pros:** Examples stay simple, focus on production code  
**Cons:** Examples might have issues

---

## Recommendation

### For This PR (Immediate)

**Fix only critical issues:**
1. Unchecked errors (real bugs)
2. Compilation errors
3. Test failures

**Ignore style issues:**
- Package comments in examples
- Unused parameters in interface implementations
- Formatting nitpicks

### For Future (Long-term)

**Adopt Option 3 (Gradual Improvement):**
1. Keep critical linters (errcheck, govet, staticcheck)
2. Make style linters warnings, not errors
3. Fix style issues in dedicated cleanup PRs
4. Exclude examples from strict linting

---

## Is Linting Necessary?

### Short Answer: **Yes, but selectively**

**Critical (Keep):**
- ✅ Error checking (`errcheck`)
- ✅ Static analysis (`staticcheck`, `govet`)
- ✅ Dead code detection (`unused`, `ineffassign`)

**Nice to Have (Can Relax):**
- ⚠️ Style rules (`revive` package-comments, `unparam` for interfaces)
- ⚠️ Formatting (`go fmt` handles this automatically)
- ⚠️ Spelling (`misspell`)

**The Balance:**
- Too strict = Frustration, ignored rules, bypassed checks
- Too lenient = Bugs slip through, inconsistent code
- **Just right** = Catches bugs, allows style flexibility

---

## Action Plan

### Immediate (Unblock PR)

1. Fix only critical linting errors (unchecked errors, bugs)
2. Temporarily increase issue limits if needed
3. Document style issues as "technical debt" to fix later

### Short-term (Next Week)

1. Review which linters are actually catching bugs vs. style
2. Disable or make warnings-only for style linters
3. Update CONTRIBUTING.md with practical guidelines

### Long-term (Next Month)

1. Gradually fix style issues in dedicated PRs
2. Exclude examples/legacy code from strict linting
3. Focus linting on new code and critical paths

---

## Bottom Line

**Linting is necessary for catching bugs, but style enforcement can be relaxed.**

The current trouble is because we:
1. Just made linting stricter
2. Are trying to fix everything at once
3. Haven't distinguished between bugs and style

**Solution:** Focus on bugs, relax on style, fix gradually.

---

## Quick Fix for This PR

If you want to unblock immediately, we can:

1. **Increase issue limits** (allow more issues temporarily)
2. **Disable style linters** (keep only bug-catching linters)
3. **Exclude examples** (they don't need production-level linting)

Which approach would you prefer?
