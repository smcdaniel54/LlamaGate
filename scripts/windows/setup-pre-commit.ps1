# Setup pre-commit hook for Windows
# Run: .\scripts\windows\setup-pre-commit.ps1

$ErrorActionPreference = "Stop"

Write-Host "Setting up pre-commit hook..." -ForegroundColor Cyan

$hooksDir = ".git\hooks"
if (-not (Test-Path $hooksDir)) {
    New-Item -ItemType Directory -Path $hooksDir -Force | Out-Null
}

$preCommitHook = Join-Path $hooksDir "pre-commit"

# Create pre-commit hook
$hookContent = @'
#!/bin/sh
# Pre-commit hook to auto-format and run golangci-lint before allowing commits
# This ensures code quality before pushing

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "${YELLOW}Running pre-commit checks (formatting + linting)...${NC}"

# Get GOPATH
GOPATH=$(go env GOPATH 2>/dev/null)
if [ -z "$GOPATH" ]; then
    echo "${RED}Error: GOPATH not set${NC}"
    exit 1
fi

# Check if golangci-lint is installed
LINT_PATH="$GOPATH/bin/golangci-lint"
if [ ! -f "$LINT_PATH" ]; then
    echo "${YELLOW}golangci-lint not found. Skipping pre-commit check.${NC}"
    echo "${YELLOW}Install with: ./scripts/unix/lint.sh (Unix) or .\\scripts\\windows\\install-golangci-lint.ps1 (Windows)${NC}"
    exit 0
fi

# Run linter on staged Go files only (faster than full project)
# Use --tests to lint test files locally (stricter than CI)
STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [ -z "$STAGED_GO_FILES" ]; then
    echo "${GREEN}No Go files staged, skipping lint check${NC}"
    exit 0
fi

# Auto-format staged files
echo "Auto-formatting staged Go files..."
go fmt $STAGED_GO_FILES 2>/dev/null || true

# Re-stage formatted files (in case formatting changed them)
FORMATTED_FILES=$(git diff --name-only | grep '\.go$' || true)
if [ -n "$FORMATTED_FILES" ]; then
    echo "${YELLOW}Re-staging formatted files...${NC}"
    git add $FORMATTED_FILES
fi

echo "Formatting complete ✓"
echo "Linting staged Go files..."
"$LINT_PATH" run --timeout=5m --tests $STAGED_GO_FILES

if [ $? -ne 0 ]; then
    echo "${RED}Linting failed! Please fix errors before committing.${NC}"
    echo ""
    echo "${YELLOW}💡 Quick fixes:${NC}"
    echo "   - Run: ./scripts/unix/lint-fix.sh (or .\\scripts\\windows\\lint-fix.ps1)"
    echo "   - Auto-fix: ./scripts/unix/lint-fix.sh --autofix"
    echo "   - Format: go fmt ./..."
    echo ""
    echo "${YELLOW}To skip this check: git commit --no-verify${NC}"
    exit 1
fi

echo "${GREEN}Pre-commit linting passed!${NC}"
exit 0
'@

# Write hook (use UTF-8 without BOM for Unix compatibility)
[System.IO.File]::WriteAllText($preCommitHook, $hookContent, [System.Text.UTF8Encoding]::new($false))

# Make executable (for Git Bash/WSL)
if (Get-Command bash -ErrorAction SilentlyContinue) {
    bash -c "chmod +x '$preCommitHook'" 2>$null
}

Write-Host "Pre-commit hook installed at: $preCommitHook" -ForegroundColor Green
Write-Host "The hook will auto-format and lint staged files before each commit." -ForegroundColor Yellow
Write-Host "To skip: git commit --no-verify" -ForegroundColor Yellow
