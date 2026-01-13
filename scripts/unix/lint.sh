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
    echo "golangci-lint not found. Installing using official script (matches CI)..."
    # Use v2.8.0 to match CI exactly (golangci-lint-action@v3 with version: v2.8.0)
    # This ensures we get the same version CI uses
    curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b "$GOPATH/bin" v2.8.0
    if [ $? -ne 0 ]; then
        echo "Failed to install golangci-lint. Please install manually:"
        echo "  curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b \$(go env GOPATH)/bin v2.8.0"
        exit 1
    fi
fi

# Run the linter with test files enabled (stricter than CI)
# CI uses tests: false for speed, but developers should lint everything locally
"$LINT_PATH" run --timeout=15m --tests

if [ $? -ne 0 ]; then
    echo "Linting failed!"
    exit 1
fi

echo "Linting passed!"

