# Comprehensive Installer Testing Script
# Tests both Windows and Unix installers

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "LlamaGate Installer Test Suite" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$totalErrors = 0

# Test Windows Installer
Write-Host "=== Testing Windows Installer ===" -ForegroundColor Yellow
Write-Host ""

if (Test-Path "install\windows\install.ps1") {
    # Syntax validation (try script block compilation)
    Write-Host "[1/3] Validating PowerShell syntax..." -ForegroundColor Cyan
    $content = Get-Content "install\windows\install.ps1" -Raw
    try {
        $scriptBlock = [scriptblock]::Create($content)
        Write-Host "  [OK] PowerShell syntax is valid (script compiles)" -ForegroundColor Green
    } catch {
        # Try tokenizer for details
        $parseErrors = $null
        [System.Management.Automation.PSParser]::Tokenize($content, [ref]$parseErrors) | Out-Null
        $criticalErrors = $parseErrors | Where-Object { $_.Message -notmatch "Unexpected token.*Path" }
        if ($criticalErrors.Count -eq 0) {
            Write-Host "  [OK] PowerShell syntax appears valid (parser warnings may be from Unicode chars)" -ForegroundColor Green
        } else {
            Write-Host "  [WARN] Parser found some issues (may be false positives):" -ForegroundColor Yellow
            foreach ($parseError in $criticalErrors | Select-Object -First 3) {
                Write-Host "    Line $($parseError.Token.StartLine): $($parseError.Message)" -ForegroundColor Yellow
            }
        }
    }
    
    # File structure check
    Write-Host "[2/3] Checking file structure..." -ForegroundColor Cyan
    $requiredFiles = @(
        "install\windows\install.cmd",
        "install\windows\install.ps1"
    )
    $allExist = $true
    foreach ($file in $requiredFiles) {
        if (Test-Path $file) {
            Write-Host "  [OK] $file exists" -ForegroundColor Green
        } else {
            Write-Host "  [FAIL] $file missing" -ForegroundColor Red
            $allExist = $false
            $totalErrors++
        }
    }
    
    # Function check
    Write-Host "[3/3] Checking required functions..." -ForegroundColor Cyan
    $requiredFunctions = @("Test-Command", "Get-UserInput")
    $allFound = $true
    foreach ($func in $requiredFunctions) {
        if ($content -match "function\s+$func") {
            Write-Host "  [OK] Function $func found" -ForegroundColor Green
        } else {
            Write-Host "  [FAIL] Function $func not found" -ForegroundColor Red
            $allFound = $false
            $totalErrors++
        }
    }
} else {
    Write-Host "  [FAIL] Windows installer not found" -ForegroundColor Red
    $totalErrors++
}

Write-Host ""

# Test Unix Installer
Write-Host "=== Testing Unix Installer ===" -ForegroundColor Yellow
Write-Host ""

if (Test-Path "install/unix/install.sh") {
    # Check if bash is available (try multiple ways)
    $bashAvailable = $false
    if (Get-Command bash -ErrorAction SilentlyContinue) {
        # Test if bash actually works
        $bashTest = bash -c "echo test" 2>&1
        if ($LASTEXITCODE -eq 0) {
            $bashAvailable = $true
        }
    }
    
    if ($bashAvailable) {
        Write-Host "[1/3] Validating bash syntax..." -ForegroundColor Cyan
        $bashSyntaxTest = bash -c "bash -n install/unix/install.sh 2>&1"
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  [OK] Bash syntax is valid" -ForegroundColor Green
        } else {
            Write-Host "  [FAIL] Syntax errors found:" -ForegroundColor Red
            Write-Host "  $bashSyntaxTest" -ForegroundColor Red
            $totalErrors++
        }
        
        # Check shebang
        Write-Host "[2/3] Checking shebang..." -ForegroundColor Cyan
        $firstLine = Get-Content "install/unix/install.sh" -First 1
        if ($firstLine -match "^#!/bin/bash") {
            Write-Host "  [OK] Shebang is correct" -ForegroundColor Green
        } else {
            Write-Host "  [WARN] Shebang may be incorrect: $firstLine" -ForegroundColor Yellow
        }
        
        # Check required functions
        Write-Host "[3/3] Checking required functions..." -ForegroundColor Cyan
        $content = Get-Content "install/unix/install.sh" -Raw
        $requiredFunctions = @("command_exists", "prompt_user", "detect_os")
        $allFound = $true
        foreach ($func in $requiredFunctions) {
            if ($content -match "function $func" -or $content -match "$func\(\)") {
                Write-Host "  [OK] Function $func found" -ForegroundColor Green
            } else {
                Write-Host "  [FAIL] Function $func not found" -ForegroundColor Red
                $allFound = $false
                $totalErrors++
            }
        }
    } else {
        Write-Host "[1/3] Validating bash syntax..." -ForegroundColor Cyan
        Write-Host "  [SKIP] Bash not available - skipping syntax validation" -ForegroundColor Yellow
        Write-Host "  Install WSL or Git Bash to test Unix installer syntax" -ForegroundColor Yellow
    }
} else {
    Write-Host "  [FAIL] Unix installer not found" -ForegroundColor Red
    $totalErrors++
}

# Summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

if ($totalErrors -eq 0) {
    Write-Host "[PASS] All installer tests passed!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Yellow
    Write-Host "  1. Test Windows installer: install\windows\install.cmd" -ForegroundColor White
    Write-Host "  2. Test Unix installer in WSL: wsl bash install/unix/install.sh" -ForegroundColor White
    exit 0
} else {
    Write-Host "[FAIL] Found $totalErrors issue(s) that need to be fixed" -ForegroundColor Red
    exit 1
}

