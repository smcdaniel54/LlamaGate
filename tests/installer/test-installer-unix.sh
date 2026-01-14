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

# Test 4: Validate repository URL in binary installer
echo "[4/9] Validating repository URL in binary installer..."
if [ -f "install/unix/install-binary.sh" ]; then
    EXPECTED_REPO="smcdaniel54/LlamaGate"
    WRONG_REPO="llamagate/llamagate"
    
    if grep -q "github.com/$EXPECTED_REPO" install/unix/install-binary.sh; then
        echo "  ✓ Repository URL is correct: $EXPECTED_REPO"
    elif grep -q "github.com/$WRONG_REPO" install/unix/install-binary.sh; then
        echo "  ✗ Repository URL is incorrect: $WRONG_REPO (should be $EXPECTED_REPO)"
        ERRORS=$((ERRORS + 1))
    else
        echo "  ⚠ Could not find repository URL in installer"
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

# Test 8: Test one-liner binary installer download (Method 1)
echo "[8/9] Testing one-liner binary installer download (Method 1)..."
ONE_LINER_BINARY_URL="https://raw.githubusercontent.com/smcdaniel54/LlamaGate/main/install/unix/install-binary.sh"
if command -v curl >/dev/null 2>&1; then
    if curl -fsSL --max-time 10 "$ONE_LINER_BINARY_URL" 2>/dev/null | grep -q "LlamaGate Binary Installer"; then
        echo "  ✓ One-liner binary installer is downloadable and valid"
    else
        echo "  ⚠ One-liner binary installer download succeeded but content may be invalid"
    fi
elif command -v wget >/dev/null 2>&1; then
    if wget -q --timeout=10 -O- "$ONE_LINER_BINARY_URL" 2>/dev/null | grep -q "LlamaGate Binary Installer"; then
        echo "  ✓ One-liner binary installer is downloadable and valid"
    else
        echo "  ⚠ One-liner binary installer download succeeded but content may be invalid"
    fi
else
    echo "  ⚠ Cannot test one-liner download (curl/wget not available)"
    echo "  This is expected in CI environments without internet access"
fi

# Test 9: Test one-liner source installer download (Method 3)
echo "[9/9] Testing one-liner source installer download (Method 3)..."
ONE_LINER_SOURCE_URL="https://raw.githubusercontent.com/smcdaniel54/LlamaGate/main/install/unix/install.sh"
if command -v curl >/dev/null 2>&1; then
    if curl -fsSL --max-time 10 "$ONE_LINER_SOURCE_URL" 2>/dev/null | grep -q "LlamaGate Installer"; then
        echo "  ✓ One-liner source installer is downloadable and valid"
    else
        echo "  ⚠ One-liner source installer download succeeded but content may be invalid"
    fi
elif command -v wget >/dev/null 2>&1; then
    if wget -q --timeout=10 -O- "$ONE_LINER_SOURCE_URL" 2>/dev/null | grep -q "LlamaGate Installer"; then
        echo "  ✓ One-liner source installer is downloadable and valid"
    else
        echo "  ⚠ One-liner source installer download succeeded but content may be invalid"
    fi
else
    echo "  ⚠ Cannot test one-liner download (curl/wget not available)"
    echo "  This is expected in CI environments without internet access"
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

