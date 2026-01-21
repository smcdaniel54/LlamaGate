#!/bin/bash
# Lint-staged script for Unix/Linux/macOS - Quick lint check on staged files only
# Run: ./scripts/unix/lint-staged.sh
# Usage: Run this before committing to check only staged files (fast feedback)

set -e

echo ""
echo "=== Lint-Staged: Quick Check ==="
echo "Checking only staged files for fast feedback"
echo ""

# Get GOPATH
GOPATH=$(go env GOPATH)
if [ -z "$GOPATH" ]; then
    echo "Error: GOPATH not set"
    exit 1
fi

# Check if golangci-lint is installed
LINT_PATH="$GOPATH/bin/golangci-lint"

if [ ! -f "$LINT_PATH" ]; then
    echo "golangci-lint not found. Please install first:"
    echo "  ./scripts/unix/lint.sh"
    exit 1
fi

# Get staged Go files
STAGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [ -z "$STAGED_FILES" ]; then
    echo "✅ No Go files staged, nothing to lint"
    exit 0
fi

echo "📝 Staged files:"
echo "$STAGED_FILES" | sed 's/^/   /'
echo ""

# Run linting on staged files only
if $LINT_PATH run --timeout=5m --tests $STAGED_FILES; then
    echo ""
    echo "✅ Staged files passed linting! Ready to commit."
    exit 0
else
    echo ""
    echo "❌ Linting failed on staged files"
    echo ""
    echo "💡 Run full lint check:"
    echo "   ./scripts/unix/lint-fix.sh"
    echo ""
    echo "💡 Or auto-fix what's possible:"
    echo "   ./scripts/unix/lint-fix.sh --autofix"
    exit 1
fi
