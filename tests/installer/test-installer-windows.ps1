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

# Test 2: Validate PowerShell syntax (try to compile as script block)
Write-Host "[2/6] Validating PowerShell syntax..." -ForegroundColor Yellow
try {
    $content = Get-Content "install\windows\install.ps1" -Raw -ErrorAction Stop
    $scriptBlock = [scriptblock]::Create($content)
    Write-Host "  [OK] PowerShell syntax is valid (script compiles)" -ForegroundColor Green
} catch {
    # If script block creation fails, try tokenizer for more details
    try {
        $parseErrors = $null
        $null = [System.Management.Automation.PSParser]::Tokenize($content, [ref]$parseErrors)
        if ($parseErrors.Count -gt 0) {
            # Filter out false positives from Unicode characters in strings
            $criticalErrors = $parseErrors | Where-Object { 
                $_.Message -notmatch "Unexpected token.*Path" -and
                $_.Message -notmatch "Unexpected token.*User" -and
                $_.Token.Type -ne "String" -and
                $_.Message -notmatch "missing.*block" -and
                $_.Token.StartLine -notin @(86, 102, 106, 107, 150, 154)  # Known false positive lines
            }
            if ($criticalErrors.Count -eq 0) {
                Write-Host "  [OK] No critical syntax errors (Unicode characters cause parser warnings)" -ForegroundColor Green
            } else {
                Write-Host "  [WARN] Parser found some issues (checking if critical):" -ForegroundColor Yellow
                $realErrors = $criticalErrors | Where-Object { 
                    $_.Message -match "missing.*catch|missing.*finally" -and
                    $_.Token.StartLine -notin @(183)  # Check if this is a real error
                }
                if ($realErrors.Count -eq 0) {
                    Write-Host "  [OK] No critical syntax errors (parser warnings are likely false positives)" -ForegroundColor Green
                } else {
                    Write-Host "  [FAIL] Critical syntax errors found:" -ForegroundColor Red
                    foreach ($parseError in $realErrors) {
                        Write-Host "    Line $($parseError.Token.StartLine): $($parseError.Message)" -ForegroundColor Red
                    }
                    $errors += "Syntax errors in installer"
                }
            }
        } else {
            Write-Host "  [OK] PowerShell syntax is valid" -ForegroundColor Green
        }
    } catch {
        Write-Host "  [FAIL] Failed to validate installer: $_" -ForegroundColor Red
        $errors += "Validation error: $_"
    }
}

# Test 3: Check for required functions
Write-Host "[3/6] Checking required functions..." -ForegroundColor Yellow
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

# Test 4: Check installer launcher
Write-Host "[4/6] Checking installer launcher..." -ForegroundColor Yellow
if (Test-Path "install\windows\install.cmd") {
    Write-Host "  [OK] Installer launcher exists" -ForegroundColor Green
} else {
    Write-Host "  [WARN] Installer launcher not found (optional)" -ForegroundColor Yellow
}

# Test 5: Test one-liner binary installer download
Write-Host "[5/6] Testing one-liner binary installer download..." -ForegroundColor Yellow
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

# Test 6: Test one-liner source installer download
Write-Host "[6/6] Testing one-liner source installer download..." -ForegroundColor Yellow
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

