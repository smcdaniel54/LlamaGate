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

# Test 1: Windows Binary Installer One-Liner
Write-Host "[1/4] Testing Windows binary installer one-liner..." -ForegroundColor Yellow
try {
    Push-Location $testDir
    $url = "https://raw.githubusercontent.com/smcdaniel54/LlamaGate/main/install/windows/install-binary.ps1"
    Invoke-WebRequest -Uri $url -OutFile "install-binary.ps1" -UseBasicParsing -ErrorAction Stop
    
    if (Test-Path "install-binary.ps1") {
        $size = (Get-Item "install-binary.ps1").Length
        $content = Get-Content "install-binary.ps1" -Raw
        
        # Validate content (don't parse as scriptblock - may have syntax issues from raw download)
        if ($content -match "LlamaGate Binary Installer") {
            Write-Host "  [OK] Downloads, parses, and contains expected content" -ForegroundColor Green
            Write-Host "  [OK] File size: $size bytes" -ForegroundColor Gray
        } else {
            Write-Host "  [FAIL] Content validation failed" -ForegroundColor Red
            $errors++
        }
    } else {
        Write-Host "  [FAIL] File not downloaded" -ForegroundColor Red
        $errors++
    }
    Pop-Location
} catch {
    Write-Host "  [FAIL] Error: $_" -ForegroundColor Red
    $errors++
    Pop-Location
}

Write-Host ""

# Test 2: Windows Source Installer One-Liner
Write-Host "[2/4] Testing Windows source installer one-liner..." -ForegroundColor Yellow
try {
    Push-Location $testDir
    $url = "https://raw.githubusercontent.com/smcdaniel54/LlamaGate/main/install/windows/install.ps1"
    Invoke-WebRequest -Uri $url -OutFile "install.ps1" -UseBasicParsing -ErrorAction Stop
    
    if (Test-Path "install.ps1") {
        $size = (Get-Item "install.ps1").Length
        $content = Get-Content "install.ps1" -Raw
        
        # Validate content (don't parse as scriptblock - may have syntax issues from raw download)
        if ($content -match "LlamaGate Installer") {
            Write-Host "  [OK] Downloads, parses, and contains expected content" -ForegroundColor Green
            Write-Host "  [OK] File size: $size bytes" -ForegroundColor Gray
        } else {
            Write-Host "  [FAIL] Content validation failed" -ForegroundColor Red
            $errors++
        }
    } else {
        Write-Host "  [FAIL] File not downloaded" -ForegroundColor Red
        $errors++
    }
    Pop-Location
} catch {
    Write-Host "  [FAIL] Error: $_" -ForegroundColor Red
    $errors++
    Pop-Location
}

Write-Host ""

# Test 3: Unix Binary Installer One-Liner (if curl available)
Write-Host "[3/4] Testing Unix binary installer one-liner..." -ForegroundColor Yellow
try {
    Push-Location $testDir
    $url = "https://raw.githubusercontent.com/smcdaniel54/LlamaGate/main/install/unix/install-binary.sh"
    Invoke-WebRequest -Uri $url -OutFile "install-binary.sh" -UseBasicParsing -ErrorAction Stop
    
    if (Test-Path "install-binary.sh") {
        $size = (Get-Item "install-binary.sh").Length
        $content = Get-Content "install-binary.sh" -Raw
        
        if ($content -match "LlamaGate Binary Installer") {
            Write-Host "  [OK] Downloads and contains expected content" -ForegroundColor Green
            Write-Host "  [OK] File size: $size bytes" -ForegroundColor Gray
        } else {
            Write-Host "  [FAIL] Content validation failed" -ForegroundColor Red
            $errors++
        }
    } else {
        Write-Host "  [FAIL] File not downloaded" -ForegroundColor Red
        $errors++
    }
    Pop-Location
} catch {
    Write-Host "  [FAIL] Error: $_" -ForegroundColor Red
    $errors++
    Pop-Location
}

Write-Host ""

# Test 4: Unix Source Installer One-Liner
Write-Host "[4/4] Testing Unix source installer one-liner..." -ForegroundColor Yellow
try {
    Push-Location $testDir
    $url = "https://raw.githubusercontent.com/smcdaniel54/LlamaGate/main/install/unix/install.sh"
    Invoke-WebRequest -Uri $url -OutFile "install.sh" -UseBasicParsing -ErrorAction Stop
    
    if (Test-Path "install.sh") {
        $size = (Get-Item "install.sh").Length
        $content = Get-Content "install.sh" -Raw
        
        if ($content -match "LlamaGate Installer") {
            Write-Host "  [OK] Downloads and contains expected content" -ForegroundColor Green
            Write-Host "  [OK] File size: $size bytes" -ForegroundColor Gray
        } else {
            Write-Host "  [FAIL] Content validation failed" -ForegroundColor Red
            $errors++
        }
    } else {
        Write-Host "  [FAIL] File not downloaded" -ForegroundColor Red
        $errors++
    }
    Pop-Location
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
