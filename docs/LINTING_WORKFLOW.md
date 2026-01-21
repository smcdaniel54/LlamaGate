# Linting Workflow: Fix Errors When They Occur

**Date:** 2026-01-21  
**Purpose:** Tools to fix linting errors when you encounter them (reactive approach)

---

## When to Use These Tools

**Use these scripts only when:**
- Pre-commit hook fails with linting errors
- CI fails with linting errors
- You want to check for linting issues before pushing

**Don't run these proactively** - the pre-commit hook will catch issues automatically.

---

## Quick Start (When You Have Linting Errors)

### Step 1: Pre-Commit Hook Failed

If your commit was blocked by the pre-commit hook:

**Windows:**
```powershell
# Auto-fix what can be fixed automatically
.\scripts\windows\lint-fix.ps1 -AutoFix
```

**Unix/Linux/macOS:**
```bash
# Auto-fix what can be fixed automatically
./scripts/unix/lint-fix.sh --autofix
```

### Step 2: Fix Remaining Issues

After auto-fix, manually fix remaining issues, then verify:

**Windows:**
```powershell
.\scripts\windows\lint-fix.ps1
```

**Unix/Linux/macOS:**
```bash
./scripts/unix/lint-fix.sh
```

### Step 3: Commit When Clean

Once linting passes, commit normally:
```bash
git commit -m "your message"
```

---

## Quick Check (Staged Files Only)

If you want to check staged files before committing (optional):

**Windows:**
```powershell
.\scripts\windows\lint-staged.ps1
```

**Unix/Linux/macOS:**
```bash
./scripts/unix/lint-staged.sh
```

---

## Workflow Options

### 1. **Standard Workflow** (When Pre-Commit Fails)

```bash
# 1. Try to commit
git commit -m "your message"
# → Pre-commit hook fails with linting errors

# 2. Auto-fix what can be fixed
./scripts/unix/lint-fix.sh --autofix

# 3. Fix remaining issues manually
# ... edit files ...

# 4. Verify again
./scripts/unix/lint-fix.sh

# 5. Commit when clean
git commit -m "your message"
```

### 2. **Watch Mode** (When Fixing Multiple Issues)

Use watch mode when you have many linting errors to fix:

**Windows:**
```powershell
.\scripts\windows\lint-fix.ps1 -Watch
```

**Unix/Linux/macOS:**
```bash
./scripts/unix/lint-fix.sh --watch
```

**Use Case:** When fixing many linting errors, watch mode provides continuous feedback as you fix each issue.

**Requirements:**
- macOS: `brew install fswatch`
- Linux: `apt-get install inotify-tools` or `yum install inotify-tools`

### 3. **Auto-Fix Workflow**

Let the linter fix what it can automatically:

```bash
# Run with auto-fix
./scripts/unix/lint-fix.sh --autofix

# Review changes
git diff

# If satisfied, commit
git add .
git commit -m "fix: auto-fix linting issues"
```

---

## Scripts Overview

### `lint-fix.sh` / `lint-fix.ps1`

**Purpose:** Comprehensive linting with helpful feedback

**Features:**
- Shows top issues first
- Identifies fixable issues
- Provides helpful tips
- Supports watch mode
- Supports auto-fix mode

**Options:**
- `--watch` / `-Watch`: Watch for file changes
- `--autofix` / `-AutoFix`: Auto-fix issues where possible
- `[path]`: Lint specific directory (default: current directory)

**Example:**
```bash
# Lint specific package
./scripts/unix/lint-fix.sh ./internal/extensions

# Watch mode with auto-fix
./scripts/unix/lint-fix.sh --watch --autofix
```

### `lint-staged.sh` / `lint-staged.ps1`

**Purpose:** Quick check on staged files only

**Features:**
- Fast (only checks staged files)
- Shows which files are being checked
- Provides next steps if issues found

**Use Case:** Run before every commit for fast feedback

**Example:**
```bash
git add file.go
./scripts/unix/lint-staged.sh
```

---

## Pre-Commit Hook (Primary Safety Net)

The pre-commit hook automatically runs linting on staged files. **This is your primary tool** - you don't need to run linting manually unless the hook fails.

### When Pre-Commit Hook Fails

1. **Don't skip it** - Fix the issues instead
2. **Use lint-fix script** - It provides better error messages and auto-fix
3. **Auto-fix first** - Run with `--autofix` to fix what can be fixed automatically
4. **Fix remaining manually** - Then verify with lint-fix script
5. **Commit again** - The hook will pass once issues are fixed

### Bypassing Pre-Commit (Not Recommended)

Only use if absolutely necessary:

```bash
git commit --no-verify
```

**Warning:** This bypasses all pre-commit checks, not just linting. You'll likely fail CI.

---

## Common Issues & Solutions

### Issue: "package-comments: should have a package comment"

**Solution:**
```go
// Package mypackage provides functionality for...
package mypackage
```

### Issue: "unused-parameter: parameter 'ctx' seems to be unused"

**Solution:**
```go
// Prefix with underscore if intentionally unused
func MyFunc(_ context.Context, data string) error {
    // ctx not used, but required by interface
}
```

### Issue: "Error return value of 'Close' is not checked"

**Solution:**
```go
// Option 1: Check the error
if err := file.Close(); err != nil {
    return err
}

// Option 2: Explicitly ignore (document why)
_ = file.Close()  // File is read-only, errors are non-critical
```

### Issue: Formatting errors

**Solution:**
```bash
# Auto-format all Go files
go fmt ./...

# Or format specific file
go fmt ./internal/extensions/workflow.go
```

---

## Integration with IDE

### VS Code

Add to `.vscode/settings.json`:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.formatTool": "goimports",
  "editor.formatOnSave": true
}
```

### GoLand / IntelliJ

1. Settings → Tools → File Watchers
2. Add golangci-lint watcher
3. Trigger: On file save

### Vim/Neovim

Use `vim-go` plugin with `golangci-lint` integration:

```vim
let g:go_metalinter_command = 'golangci-lint'
let g:go_metalinter_enabled = ['errcheck', 'govet', 'staticcheck']
```

---

## Best Practices

### 1. **Let Pre-Commit Hook Do Its Job**

The pre-commit hook runs automatically. **Only use lint-fix scripts when the hook fails.**

```bash
# Normal workflow - just commit
git commit -m "your message"
# → Pre-commit hook runs automatically

# If hook fails, then use lint-fix
./scripts/unix/lint-fix.sh --autofix
# Fix remaining issues, then commit again
```

### 2. **Auto-Fix First, Manual Fix Second**

When the hook fails:
1. Run with `--autofix` first
2. Review auto-fixed changes
3. Manually fix remaining issues
4. Commit again

### 3. **Use Watch Mode Only When Fixing Many Issues**

Watch mode is helpful when you have many linting errors to fix:

```bash
# When fixing many errors
./scripts/unix/lint-fix.sh --watch
# Fix issues as they're reported
```

### 4. **Fix Issues Incrementally**

Don't try to fix everything at once:
- Fix issues in files you're actively working on
- Use exclude rules for legacy code
- Gradually improve codebase quality

### 5. **Don't Skip Pre-Commit Hook**

The pre-commit hook is your primary safety net. Fix issues instead of bypassing.

---

## Troubleshooting

### Linting is Slow

**Solutions:**
- Use `lint-staged.sh` for quick checks (only staged files)
- Lint specific directories: `./scripts/unix/lint-fix.sh ./internal/extensions`
- Reduce timeout in `.golangci.yml` (not recommended)

### Too Many Issues

**Solutions:**
- Start with auto-fix: `--autofix`
- Focus on files you're changing
- Use exclude rules for legacy code
- Fix incrementally, not all at once

### Watch Mode Not Working

**Check:**
- Is `fswatch` (macOS) or `inotifywait` (Linux) installed?
- Are you watching a valid directory?
- Check file permissions

**Install:**
```bash
# macOS
brew install fswatch

# Linux
apt-get install inotify-tools
# or
yum install inotify-tools
```

### Pre-Commit Hook Not Running

**Check:**
- Is the hook installed? Run: `./scripts/unix/setup-pre-commit.sh`
- Is the hook executable? `chmod +x .git/hooks/pre-commit`
- Check hook content: `cat .git/hooks/pre-commit`

---

## Summary

**When to Use:**

```bash
# Only when pre-commit hook fails or CI fails
./scripts/unix/lint-fix.sh --autofix  # Auto-fix issues
./scripts/unix/lint-fix.sh             # Verify fixes

# Optional: Check before committing (not required)
./scripts/unix/lint-staged.sh

# Optional: Watch mode when fixing many issues
./scripts/unix/lint-fix.sh --watch
```

**Normal Workflow:**
1. Code
2. `git commit -m "message"` → Pre-commit hook runs automatically
3. If hook fails → Use `lint-fix.sh --autofix`
4. Fix remaining issues
5. Commit again

**Remember:** The pre-commit hook is your primary tool. Only use lint-fix scripts when you encounter linting errors.
