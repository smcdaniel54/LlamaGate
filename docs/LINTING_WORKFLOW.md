# Linting Workflow: Tight Feedback Loop

**Date:** 2026-01-21  
**Purpose:** Catch and fix linting errors before committing

---

## Quick Start

### Before Committing

Run the lint-fix script to catch errors early:

**Windows:**
```powershell
.\scripts\windows\lint-fix.ps1
```

**Unix/Linux/macOS:**
```bash
./scripts/unix/lint-fix.sh
```

### Auto-Fix Issues

Many linting issues can be auto-fixed:

**Windows:**
```powershell
.\scripts\windows\lint-fix.ps1 -AutoFix
```

**Unix/Linux/macOS:**
```bash
./scripts/unix/lint-fix.sh --autofix
```

### Quick Check (Staged Files Only)

For faster feedback on just your changes:

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

### 1. **Standard Workflow** (Recommended)

```bash
# 1. Make your changes
git add .

# 2. Quick check on staged files (fast)
./scripts/unix/lint-staged.sh

# 3. If issues found, run full lint with auto-fix
./scripts/unix/lint-fix.sh --autofix

# 4. Fix remaining issues manually
# ... edit files ...

# 5. Verify again
./scripts/unix/lint-fix.sh

# 6. Commit when clean
git commit -m "your message"
```

### 2. **Watch Mode** (Continuous Feedback)

Automatically re-lint when files change:

**Windows:**
```powershell
.\scripts\windows\lint-fix.ps1 -Watch
```

**Unix/Linux/macOS:**
```bash
./scripts/unix/lint-fix.sh --watch
```

**Use Case:** Keep this running in a terminal while you code. It will automatically check your code as you save files.

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

## Pre-Commit Hook

The pre-commit hook automatically runs linting on staged files. If it fails:

1. **Don't skip it** - Fix the issues instead
2. **Use lint-fix script** - It provides better error messages
3. **Auto-fix when possible** - Many issues can be fixed automatically

### Bypassing Pre-Commit (Not Recommended)

Only use if absolutely necessary:

```bash
git commit --no-verify
```

**Warning:** This bypasses all pre-commit checks, not just linting.

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

### 1. **Run Before Committing**

Always run lint-fix before committing:

```bash
./scripts/unix/lint-fix.sh
git commit -m "your message"
```

### 2. **Use Watch Mode During Development**

Keep watch mode running while coding:

```bash
# Terminal 1: Watch mode
./scripts/unix/lint-fix.sh --watch

# Terminal 2: Your editor
code .
```

### 3. **Auto-Fix First, Manual Fix Second**

1. Run with `--autofix` first
2. Review auto-fixed changes
3. Manually fix remaining issues

### 4. **Fix Issues Incrementally**

Don't try to fix everything at once:
- Fix issues in files you're actively working on
- Use exclude rules for legacy code
- Gradually improve codebase quality

### 5. **Don't Skip Pre-Commit Hook**

The pre-commit hook is your safety net. Fix issues instead of bypassing.

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

**Quick Commands:**

```bash
# Before committing (recommended)
./scripts/unix/lint-fix.sh

# Auto-fix issues
./scripts/unix/lint-fix.sh --autofix

# Quick check (staged files only)
./scripts/unix/lint-staged.sh

# Watch mode (continuous feedback)
./scripts/unix/lint-fix.sh --watch
```

**Workflow:**
1. Code
2. `lint-fix.sh` (or `lint-staged.sh` for speed)
3. Fix issues
4. Commit

**Remember:** The goal is to catch issues early, not to fix everything at once. Use these tools to maintain code quality incrementally.
