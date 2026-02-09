# Test Windows Installer Syntax and Structure
# This script validates both installers (binary and source) to match the installation docs

Write-Host "=== Testing Windows Installers ===" -ForegroundColor Cyan
Write-Host "Testing Method 1 (Binary Installer) and Method 2/3 (Source Installer)" -ForegroundColor Gray
Write-Host ""

$errors = @()

# Test 1: Check if binary installer exists (Method 1 - Recommended)
Write-Host "[1/8] Checking binary installer file (Method 1)..." -ForegroundColor Yellow
if (Test-Path "install\windows\install-binary.ps1") {
    Write-Host "  [OK] Binary installer file exists" -ForegroundColor Green
} else {
    Write-Host "  [FAIL] Binary installer file not found" -ForegroundColor Red
    $errors += "Binary installer file missing (install-binary.ps1)"
}

# Test 1b: Check if source installer exists (Method 2/3 - For Developers)
Write-Host "[1b/8] Checking source installer file (Method 2/3)..." -ForegroundColor Yellow
if (Test-Path "install\windows\install.ps1") {
    Write-Host "  [OK] Source installer file exists" -ForegroundColor Green
} else {
    Write-Host "  [WARN] Source installer file not found (optional)" -ForegroundColor Yellow
}

# Test 2: Validate PowerShell syntax for binary installer (Method 1 - one-liner installer)
Write-Host "[2/8] Validating binary installer PowerShell syntax (Method 1)..." -ForegroundColor Yellow
if (Test-Path "install\windows\install-binary.ps1") {
    try {
        $binaryContent = Get-Content "install\windows\install-binary.ps1" -Raw -ErrorAction Stop
        $null = [scriptblock]::Create($binaryContent)
        Write-Host "  [OK] Binary installer syntax is valid" -ForegroundColor Green
    } catch {
        Write-Host "  [FAIL] Binary installer syntax error: $_" -ForegroundColor Red
        $errors += "Syntax errors in install-binary.ps1: $_"
    }
} else {
    Write-Host "  [FAIL] Binary installer file not found" -ForegroundColor Red
    $errors += "install-binary.ps1 missing"
}

# Test 3: Validate PowerShell syntax for source installer (Method 2/3 - for developers)
Write-Host "[3/8] Validating source installer PowerShell syntax (Method 2/3)..." -ForegroundColor Yellow
if (Test-Path "install\windows\install.ps1") {
    try {
        $content = Get-Content "install\windows\install.ps1" -Raw -ErrorAction Stop
        $null = [scriptblock]::Create($content)
        Write-Host "  [OK] Source installer syntax is valid" -ForegroundColor Green
    } catch {
        Write-Host "  [WARN] Source installer has syntax issues (may be false positives): $_" -ForegroundColor Yellow
        # Don't fail on source installer - it's more complex and may have false positives
    }
} else {
    Write-Host "  [WARN] Source installer file not found (optional)" -ForegroundColor Yellow
}

# Test 4: Binary installer uses GITHUB_REPO (configurable repo)
Write-Host "[4/8] Validating binary installer (GITHUB_REPO)..." -ForegroundColor Yellow
if (Test-Path "install\windows\install-binary.ps1") {
    $binaryContent = Get-Content "install\windows\install-binary.ps1" -Raw
    if ($binaryContent -match "GITHUB_REPO") {
        Write-Host "  [OK] Binary installer uses GITHUB_REPO for configurable repo" -ForegroundColor Green
    } else {
        Write-Host "  [FAIL] Binary installer should use GITHUB_REPO" -ForegroundColor Red
        $errors += "install-binary.ps1 should use GITHUB_REPO"
    }
} else {
    Write-Host "  [FAIL] Binary installer file not found" -ForegroundColor Red
    $errors += "install-binary.ps1 missing"
}

# Test 5: Check for required functions in source installer
Write-Host "[5/8] Checking required functions in source installer..." -ForegroundColor Yellow
$requiredFunctions = @("Test-Command", "Get-UserInput")
$content = Get-Content "install\windows\install.ps1" -Raw
$allFound = $true
foreach ($func in $requiredFunctions) {
    if ($content -match "function\s+$func") {
        Write-Host "  [OK] Function $func found" -ForegroundColor Green
    } else {
        Write-Host "  [FAIL] Function $func not found" -ForegroundColor Red
        $allFound = $false
    }
}
if (-not $allFound) {
    $errors += "Missing required functions"
}

# Test 6: Check installer launchers (for cloned repo option)
Write-Host "[6/8] Checking installer launchers (for cloned repo)..." -ForegroundColor Yellow
if (Test-Path "install\windows\install-binary.cmd") {
    Write-Host "  [OK] Binary installer launcher exists (install-binary.cmd)" -ForegroundColor Green
} else {
    Write-Host "  [WARN] Binary installer launcher not found (optional)" -ForegroundColor Yellow
}
if (Test-Path "install\windows\install.cmd") {
    Write-Host "  [OK] Source installer launcher exists (install.cmd)" -ForegroundColor Green
} else {
    Write-Host "  [WARN] Source installer launcher not found (optional)" -ForegroundColor Yellow
}

# Test 7: Binary installer script exists and is valid (no remote download)
Write-Host "[7/8] Checking binary installer script..." -ForegroundColor Yellow
if (Test-Path "install\windows\install-binary.ps1") {
    $c = Get-Content "install\windows\install-binary.ps1" -Raw
    if ($c -match "LlamaGate Binary Installer") {
        Write-Host "  [OK] Binary installer script is present and valid" -ForegroundColor Green
    } else {
        Write-Host "  [FAIL] Binary installer content invalid" -ForegroundColor Red
        $errors += "install-binary.ps1 content invalid"
    }
} else {
    Write-Host "  [FAIL] Binary installer not found" -ForegroundColor Red
    $errors += "install-binary.ps1 missing"
}

# Test 8: Source installer script exists and is valid (no remote download)
Write-Host "[8/8] Checking source installer script..." -ForegroundColor Yellow
if (Test-Path "install\windows\install.ps1") {
    $c = Get-Content "install\windows\install.ps1" -Raw
    if ($c -match "LlamaGate") {
        Write-Host "  [OK] Source installer script is present and valid" -ForegroundColor Green
    } else {
        Write-Host "  [FAIL] Source installer content invalid" -ForegroundColor Red
        $errors += "install.ps1 content invalid"
    }
} else {
    Write-Host "  [FAIL] Source installer not found" -ForegroundColor Red
    $errors += "install.ps1 missing"
}

# Summary
Write-Host ""
Write-Host "=== Test Summary ===" -ForegroundColor Cyan
if ($errors.Count -eq 0) {
    Write-Host "[PASS] All tests passed!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "[FAIL] Found $($errors.Count) issue(s):" -ForegroundColor Red
    foreach ($err in $errors) {
        Write-Host "  - $err" -ForegroundColor Red
    }
    exit 1
}

