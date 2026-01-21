#!/bin/bash
# Setup pre-commit hook for Unix/Linux/macOS
# Run: ./scripts/unix/setup-pre-commit.sh

set -e

echo "Setting up pre-commit hook..."

HOOKS_DIR=".git/hooks"
PRE_COMMIT_HOOK="$HOOKS_DIR/pre-commit"

# Create hooks directory if it doesn't exist
mkdir -p "$HOOKS_DIR"

# Create pre-commit hook
cat > "$PRE_COMMIT_HOOK" << 'EOF'
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
    echo "${YELLOW}Install with: ./scripts/unix/lint.sh${NC}"
    exit 0
fi

# Run linter on staged Go files only (faster than full project)
# Use --tests to lint test files locally (stricter than CI)
STAGED_GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [ -z "$STAGED_GO_FILES" ]; then
    echo "${GREEN}No Go files staged, skipping checks${NC}"
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
    echo "   - Run: ./scripts/unix/lint-fix.sh"
    echo "   - Auto-fix: ./scripts/unix/lint-fix.sh --autofix"
    echo "   - Format: go fmt ./..."
    echo ""
    echo "${YELLOW}To skip this check: git commit --no-verify${NC}"
    exit 1
fi

echo "${GREEN}Pre-commit checks passed!${NC}"
exit 0
EOF

# Make executable
chmod +x "$PRE_COMMIT_HOOK"

echo "✅ Pre-commit hook installed at: $PRE_COMMIT_HOOK"
echo "The hook will auto-format and lint staged files before each commit."
echo "To skip: git commit --no-verify"
