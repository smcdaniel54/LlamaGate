#!/bin/bash
# Test Unix/Linux/macOS Installer Syntax and Structure
# This script validates both installers (binary and source) to match the installation docs

set -e

echo "=== Testing Unix Installers ==="
echo "Testing Method 1 (Binary Installer) and Method 2/3 (Source Installer)"
echo ""

ERRORS=0

# Test 1: Check if binary installer exists (Method 1 - Recommended)
echo "[1/9] Checking binary installer file (Method 1)..."
if [ -f "install/unix/install-binary.sh" ]; then
    echo "  ✓ Binary installer file exists"
else
    echo "  ✗ Binary installer file not found"
    ERRORS=$((ERRORS + 1))
fi

# Test 1b: Check if source installer exists (Method 2/3 - For Developers)
echo "[1b/9] Checking source installer file (Method 2/3)..."
if [ -f "install/unix/install.sh" ]; then
    echo "  ✓ Source installer file exists"
else
    echo "  ⚠ Source installer file not found (optional)"
fi

# Test 2: Validate bash syntax for binary installer (Method 1 - one-liner installer)
echo "[2/9] Validating binary installer bash syntax (Method 1)..."
if [ -f "install/unix/install-binary.sh" ]; then
    if bash -n install/unix/install-binary.sh 2>&1; then
        echo "  ✓ Binary installer bash syntax is valid"
    else
        echo "  ✗ Binary installer syntax errors found"
        ERRORS=$((ERRORS + 1))
    fi
else
    echo "  ✗ Binary installer file not found"
    ERRORS=$((ERRORS + 1))
fi

# Test 3: Validate bash syntax for source installer (Method 2/3 - for developers)
echo "[3/9] Validating source installer bash syntax (Method 2/3)..."
if [ -f "install/unix/install.sh" ]; then
    if bash -n install/unix/install.sh 2>&1; then
        echo "  ✓ Source installer bash syntax is valid"
    else
        echo "  ✗ Source installer syntax errors found"
        ERRORS=$((ERRORS + 1))
    fi
else
    echo "  ⚠ Source installer file not found (optional)"
fi

# Test 4: Validate binary installer uses configurable repo (GITHUB_REPO)
echo "[4/9] Validating binary installer (GITHUB_REPO)..."
if [ -f "install/unix/install-binary.sh" ]; then
    if grep -q "GITHUB_REPO" install/unix/install-binary.sh; then
        echo "  ✓ Binary installer uses GITHUB_REPO for configurable repo"
    else
        echo "  ✗ Binary installer should use GITHUB_REPO"
        ERRORS=$((ERRORS + 1))
    fi
else
    echo "  ✗ Binary installer file not found"
    ERRORS=$((ERRORS + 1))
fi

# Test 5: Check if files are executable
echo "[5/9] Checking file permissions..."
if [ -x "install/unix/install-binary.sh" ]; then
    echo "  ✓ Binary installer is executable"
else
    echo "  ⚠ Binary installer is not executable (will be set during install)"
fi
if [ -x "install/unix/install.sh" ]; then
    echo "  ✓ Source installer is executable"
else
    echo "  ⚠ Source installer is not executable (will be set during install)"
fi

# Test 6: Check for required functions in source installer
echo "[6/9] Checking required functions in source installer..."
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

# Test 7: Check for shebang
echo "[7/9] Checking shebang..."
if head -n 1 install/unix/install.sh | grep -q "^#!/bin/bash"; then
    echo "  ✓ Shebang is correct"
else
    echo "  ⚠ Shebang may be missing or incorrect"
fi

# Test 8: Binary installer script exists and is valid (no remote download; repo URL is configurable)
echo "[8/9] Checking binary installer script..."
if [ -f "install/unix/install-binary.sh" ] && grep -q "LlamaGate Binary Installer" install/unix/install-binary.sh; then
    echo "  ✓ Binary installer script is present and valid"
else
    echo "  ✗ Binary installer script missing or invalid"
    ERRORS=$((ERRORS + 1))
fi

# Test 9: Source installer script exists and is valid (no remote download)
echo "[9/9] Checking source installer script..."
if [ -f "install/unix/install.sh" ] && grep -q "LlamaGate" install/unix/install.sh; then
    echo "  ✓ Source installer script is present and valid"
else
    echo "  ✗ Source installer script missing or invalid"
    ERRORS=$((ERRORS + 1))
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

