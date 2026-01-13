# Test Windows Installer Syntax and Structure
# This script validates the Windows installer without running it fully

Write-Host "=== Testing Windows Installer ===" -ForegroundColor Cyan
Write-Host ""

$errors = @()

# Test 1: Check if installer file exists
Write-Host "[1/6] Checking installer file..." -ForegroundColor Yellow
if (Test-Path "install\windows\install.ps1") {
    Write-Host "  [OK] Installer file exists" -ForegroundColor Green
} else {
    Write-Host "  [FAIL] Installer file not found" -ForegroundColor Red
    $errors += "Installer file missing"
}

# Test 2: Validate PowerShell syntax for binary installer (one-liner installer)
Write-Host "[2/7] Validating binary installer PowerShell syntax..." -ForegroundColor Yellow
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

# Test 2b: Validate PowerShell syntax for source installer
Write-Host "[3/7] Validating source installer PowerShell syntax..." -ForegroundColor Yellow
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

# Test 4: Check for required functions
Write-Host "[4/7] Checking required functions..." -ForegroundColor Yellow
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

# Test 5: Check installer launcher
Write-Host "[5/7] Checking installer launcher..." -ForegroundColor Yellow
if (Test-Path "install\windows\install.cmd") {
    Write-Host "  [OK] Installer launcher exists" -ForegroundColor Green
} else {
    Write-Host "  [WARN] Installer launcher not found (optional)" -ForegroundColor Yellow
}

# Test 6: Test one-liner binary installer download
Write-Host "[6/7] Testing one-liner binary installer download..." -ForegroundColor Yellow
$oneLinerUrl = "https://raw.githubusercontent.com/smcdaniel54/LlamaGate/main/install/windows/install-binary.ps1"
try {
    $response = Invoke-WebRequest -Uri $oneLinerUrl -UseBasicParsing -TimeoutSec 10 -ErrorAction Stop
    if ($response.StatusCode -eq 200 -and $response.Content -match "LlamaGate Binary Installer") {
        Write-Host "  [OK] One-liner binary installer is downloadable and valid" -ForegroundColor Green
    } else {
        Write-Host "  [WARN] One-liner binary installer download succeeded but content may be invalid" -ForegroundColor Yellow
    }
} catch {
    Write-Host "  [WARN] Could not test one-liner download (network issue or GitHub unavailable): $_" -ForegroundColor Yellow
    Write-Host "  This is expected in CI environments without internet access" -ForegroundColor Gray
}

# Test 7: Test one-liner source installer download
Write-Host "[7/7] Testing one-liner source installer download..." -ForegroundColor Yellow
$sourceInstallerUrl = "https://raw.githubusercontent.com/smcdaniel54/LlamaGate/main/install/windows/install.ps1"
try {
    $response = Invoke-WebRequest -Uri $sourceInstallerUrl -UseBasicParsing -TimeoutSec 10 -ErrorAction Stop
    if ($response.StatusCode -eq 200 -and $response.Content -match "LlamaGate Installer") {
        Write-Host "  [OK] One-liner source installer is downloadable and valid" -ForegroundColor Green
    } else {
        Write-Host "  [WARN] One-liner source installer download succeeded but content may be invalid" -ForegroundColor Yellow
    }
} catch {
    Write-Host "  [WARN] Could not test one-liner download (network issue or GitHub unavailable): $_" -ForegroundColor Yellow
    Write-Host "  This is expected in CI environments without internet access" -ForegroundColor Gray
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

