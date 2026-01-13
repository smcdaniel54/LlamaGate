#!/bin/bash
# Test script for race condition detection (matches CI)
# Run: ./scripts/unix/test-race.sh
# This matches the CI test-race job configuration

set -e

echo "Running race condition tests (matches CI)..."

# Check Go version matches CI (1.24)
echo "Go version: $(go version)"

# Set CGO_ENABLED=1 (required for race detector, matches CI)
export CGO_ENABLED=1

# Get packages, filtering out .out directories (matches CI logic)
echo "Discovering packages..."
PACKAGES=$(go list ./... 2>/dev/null | grep -v "^\.out$" || go list ./...)
if [ -z "$PACKAGES" ]; then
    PACKAGES=$(go list ./...)
fi

PACKAGE_COUNT=$(echo "$PACKAGES" | wc -l)
echo "Found $PACKAGE_COUNT packages to test"

# Run tests with race detector (matches CI: go test -race -timeout=10m)
echo "Running tests with race detector (timeout: 10m)..."
go test -race -timeout=10m $PACKAGES

if [ $? -eq 0 ]; then
    echo ""
    echo "Race condition tests passed!"
    exit 0
else
    echo ""
    echo "Race condition tests failed!"
    exit 1
fi
