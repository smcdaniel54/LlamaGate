#!/bin/bash
# Lint script for Unix/Linux/macOS
# Run: ./scripts/unix/lint.sh

set -e

echo "Running golangci-lint..."

# Get GOPATH
GOPATH=$(go env GOPATH)
if [ -z "$GOPATH" ]; then
    echo "Error: GOPATH not set"
    exit 1
fi

# Check if golangci-lint is installed
LINT_PATH="$GOPATH/bin/golangci-lint"
if [ ! -f "$LINT_PATH" ]; then
    echo "golangci-lint not found. Installing..."
    # Use v2.x to match CI (golangci-lint-action uses 'latest' which resolves to v2.8.0+)
    # This ensures compatibility with .golangci.yml version: 2
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.8.0
    if [ $? -ne 0 ]; then
        echo "Failed to install golangci-lint. Please install manually:"
        echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8"
        exit 1
    fi
fi

# Run the linter
"$LINT_PATH" run --timeout=15m

if [ $? -ne 0 ]; then
    echo "Linting failed!"
    exit 1
fi

echo "Linting passed!"

