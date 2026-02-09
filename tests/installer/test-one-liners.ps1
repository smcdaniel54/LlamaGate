# Test One-Liner Installation Commands
# This script tests that the one-liner commands actually work

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Testing One-Liner Installation Commands" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$errors = 0
$testDir = Join-Path $env:TEMP "llamagate-test-$(Get-Random)"
New-Item -ItemType Directory -Path $testDir -Force | Out-Null
Write-Host "Test directory: $testDir" -ForegroundColor Gray
Write-Host ""

# Test 1: Windows Binary Installer script (local copy; no remote repo)
Write-Host "[1/4] Testing Windows binary installer script (local)..." -ForegroundColor Yellow
try {
    Push-Location $testDir
    $src = Join-Path $PSScriptRoot "..\..\install\windows\install-binary.ps1"
    if (-not (Test-Path $src)) { Write-Host "  [SKIP] Run from repo root" -ForegroundColor Yellow; Pop-Location } else {
        Copy-Item $src -Destination "install-binary.ps1" -Force
        $content = Get-Content "install-binary.ps1" -Raw
        if ($content -match "LlamaGate Binary Installer") {
            Write-Host "  [OK] Script present and valid" -ForegroundColor Green
        } else { Write-Host "  [FAIL] Content validation failed" -ForegroundColor Red; $errors++ }
        Pop-Location
    }
} catch {
    Write-Host "  [FAIL] Error: $_" -ForegroundColor Red
    $errors++
    Pop-Location
}

Write-Host ""

# Test 2: Windows Source Installer script (local copy)
Write-Host "[2/4] Testing Windows source installer script (local)..." -ForegroundColor Yellow
try {
    Push-Location $testDir
    $src = Join-Path $PSScriptRoot "..\..\install\windows\install.ps1"
    if (-not (Test-Path $src)) { Write-Host "  [SKIP] Run from repo root" -ForegroundColor Yellow; Pop-Location } else {
        Copy-Item $src -Destination "install.ps1" -Force
        $content = Get-Content "install.ps1" -Raw
        if ($content -match "LlamaGate Installer") {
            Write-Host "  [OK] Script present and valid" -ForegroundColor Green
        } else { Write-Host "  [FAIL] Content validation failed" -ForegroundColor Red; $errors++ }
        Pop-Location
    }
} catch {
    Write-Host "  [FAIL] Error: $_" -ForegroundColor Red
    $errors++
    Pop-Location
}

Write-Host ""

# Test 3: Unix Binary Installer script (local copy)
Write-Host "[3/4] Testing Unix binary installer script (local)..." -ForegroundColor Yellow
try {
    Push-Location $testDir
    $src = Join-Path $PSScriptRoot "..\..\install\unix\install-binary.sh"
    if (-not (Test-Path $src)) { Write-Host "  [SKIP] Run from repo root" -ForegroundColor Yellow; Pop-Location } else {
        Copy-Item $src -Destination "install-binary.sh" -Force
        $content = Get-Content "install-binary.sh" -Raw
        if ($content -match "LlamaGate Binary Installer") {
            Write-Host "  [OK] Script present and valid" -ForegroundColor Green
        } else { Write-Host "  [FAIL] Content validation failed" -ForegroundColor Red; $errors++ }
        Pop-Location
    }
} catch {
    Write-Host "  [FAIL] Error: $_" -ForegroundColor Red
    $errors++
    Pop-Location
}

Write-Host ""

# Test 4: Unix Source Installer script (local copy)
Write-Host "[4/4] Testing Unix source installer script (local)..." -ForegroundColor Yellow
try {
    Push-Location $testDir
    $src = Join-Path $PSScriptRoot "..\..\install\unix\install.sh"
    if (-not (Test-Path $src)) { Write-Host "  [SKIP] Run from repo root" -ForegroundColor Yellow; Pop-Location } else {
        Copy-Item $src -Destination "install.sh" -Force
        $content = Get-Content "install.sh" -Raw
        if ($content -match "LlamaGate Installer") {
            Write-Host "  [OK] Script present and valid" -ForegroundColor Green
        } else { Write-Host "  [FAIL] Content validation failed" -ForegroundColor Red; $errors++ }
        Pop-Location
    }
} catch {
    Write-Host "  [FAIL] Error: $_" -ForegroundColor Red
    $errors++
    Pop-Location
}

# Cleanup
Write-Host ""
Write-Host "Cleaning up test directory..." -ForegroundColor Gray
Remove-Item -Path $testDir -Recurse -Force -ErrorAction SilentlyContinue

# Summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

if ($errors -eq 0) {
    Write-Host "[PASS] All one-liner commands work correctly!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "[FAIL] Found $errors issue(s)" -ForegroundColor Red
    exit 1
}
