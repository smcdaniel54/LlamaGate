#!/bin/bash
# Test Unix/Linux/macOS Installer Syntax and Structure
# This script validates the Unix installer without running it fully

set -e

echo "=== Testing Unix Installer ==="
echo ""

ERRORS=0

# Test 1: Check if installer file exists
echo "[1/5] Checking installer file..."
if [ -f "install/unix/install.sh" ]; then
    echo "  ✓ Installer file exists"
else
    echo "  ✗ Installer file not found"
    ERRORS=$((ERRORS + 1))
fi

# Test 2: Validate bash syntax
echo "[2/5] Validating bash syntax..."
if bash -n install/unix/install.sh 2>&1; then
    echo "  ✓ Bash syntax is valid"
else
    echo "  ✗ Syntax errors found"
    ERRORS=$((ERRORS + 1))
fi

# Test 3: Check if file is executable
echo "[3/5] Checking file permissions..."
if [ -x "install/unix/install.sh" ]; then
    echo "  ✓ File is executable"
else
    echo "  ⚠ File is not executable (will be set during install)"
fi

# Test 4: Check for required functions
echo "[4/5] Checking required functions..."
REQUIRED_FUNCS=("command_exists" "prompt_user" "detect_os" "print_info" "print_success" "print_error")
CONTENT=$(cat install/unix/install.sh)
ALL_FOUND=true

for func in "${REQUIRED_FUNCS[@]}"; do
    if echo "$CONTENT" | grep -q "function $func\|$func()"; then
        echo "  ✓ Function $func found"
    else
        echo "  ✗ Function $func not found"
        ALL_FOUND=false
    fi
done

if [ "$ALL_FOUND" = false ]; then
    ERRORS=$((ERRORS + 1))
fi

# Test 5: Check for shebang
echo "[5/5] Checking shebang..."
if head -n 1 install/unix/install.sh | grep -q "^#!/bin/bash"; then
    echo "  ✓ Shebang is correct"
else
    echo "  ⚠ Shebang may be missing or incorrect"
fi

# Summary
echo ""
echo "=== Test Summary ==="
if [ $ERRORS -eq 0 ]; then
    echo "✓ All tests passed!"
    exit 0
else
    echo "✗ Found $ERRORS issue(s)"
    exit 1
fi

