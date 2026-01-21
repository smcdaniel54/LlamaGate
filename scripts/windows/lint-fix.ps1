# Lint-fix script for Windows - Tight feedback loop for fixing linting errors
# Run: .\scripts\windows\lint-fix.ps1 [path]
# Usage: Run this before committing to catch and fix linting errors quickly

param(
    [string]$Path = ".",
    [switch]$Watch = $false,
    [switch]$AutoFix = $false
)

$ErrorActionPreference = "Stop"

Write-Host "`n=== Lint-Fix: Tight Feedback Loop ===" -ForegroundColor Cyan
Write-Host "Catch and fix linting errors before committing`n" -ForegroundColor Yellow

# Get GOPATH
$gopath = go env GOPATH
if (-not $gopath) {
    Write-Host "Error: GOPATH not set" -ForegroundColor Red
    exit 1
}

# Check if golangci-lint is installed
$lintPath = "$gopath\bin\golangci-lint.exe"

if (-not (Test-Path $lintPath)) {
    Write-Host "golangci-lint not found. Installing..." -ForegroundColor Yellow
    $installScript = Join-Path $PSScriptRoot "install-golangci-lint.ps1"
    if (Test-Path $installScript) {
        & $installScript
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Failed to install golangci-lint." -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host "Install script not found. Please run:" -ForegroundColor Red
        Write-Host "  .\scripts\windows\install-golangci-lint.ps1" -ForegroundColor Yellow
        exit 1
    }
}

# Function to run linting with helpful output
function Run-Lint {
    param([string]$LintPath, [string]$TargetPath, [bool]$AutoFixMode)
    
    Write-Host "`n[$(Get-Date -Format 'HH:mm:ss')] Running lint check..." -ForegroundColor Cyan
    
    if ($AutoFixMode) {
        Write-Host "Attempting auto-fix for fixable issues..." -ForegroundColor Yellow
        & $LintPath run --timeout=10m --tests --fix $TargetPath 2>&1 | Tee-Object -Variable lintOutput
    } else {
        & $LintPath run --timeout=10m --tests $TargetPath 2>&1 | Tee-Object -Variable lintOutput
    }
    
    $exitCode = $LASTEXITCODE
    
    if ($exitCode -eq 0) {
        Write-Host "`n✅ Linting passed! Ready to commit." -ForegroundColor Green
        return $true
    } else {
        Write-Host "`n❌ Linting failed with $exitCode issues" -ForegroundColor Red
        Write-Host "`n📋 Summary:" -ForegroundColor Yellow
        
        # Extract issue summary
        $issueLines = $lintOutput | Select-String -Pattern "^\s+\*" | Select-Object -First 10
        if ($issueLines) {
            Write-Host "Top issues:" -ForegroundColor Yellow
            $issueLines | ForEach-Object { Write-Host "  $_" -ForegroundColor Gray }
        }
        
        # Show fixable issues
        $fixable = $lintOutput | Select-String -Pattern "fixable|can be auto-fixed"
        if ($fixable) {
            Write-Host "`n💡 Some issues can be auto-fixed. Run with -AutoFix flag:" -ForegroundColor Cyan
            Write-Host "   .\scripts\windows\lint-fix.ps1 -AutoFix" -ForegroundColor White
        }
        
        Write-Host "`n💡 Tips:" -ForegroundColor Cyan
        Write-Host "   - Run 'go fmt ./...' to fix formatting" -ForegroundColor White
        Write-Host "   - Run with -AutoFix to auto-fix what's possible" -ForegroundColor White
        Write-Host "   - Check full output above for details" -ForegroundColor White
        
        return $false
    }
}

# Watch mode
if ($Watch) {
    Write-Host "🔍 Watch mode: Monitoring for file changes..." -ForegroundColor Cyan
    Write-Host "Press Ctrl+C to stop`n" -ForegroundColor Yellow
    
    $watcher = New-Object System.IO.FileSystemWatcher
    $watcher.Path = Resolve-Path $Path
    $watcher.Filter = "*.go"
    $watcher.IncludeSubdirectories = $true
    $watcher.EnableRaisingEvents = $true
    
    $action = {
        $file = $Event.SourceEventArgs.FullPath
        Write-Host "`n📝 File changed: $file" -ForegroundColor Cyan
        Start-Sleep -Milliseconds 500  # Debounce
        Run-Lint -LintPath $lintPath -TargetPath $Path -AutoFixMode $AutoFix
    }
    
    Register-ObjectEvent -InputObject $watcher -EventName "Changed" -Action $action | Out-Null
    
    try {
        # Run initial check
        Run-Lint -LintPath $lintPath -TargetPath $Path -AutoFixMode $AutoFix
        
        # Keep running
        while ($true) {
            Start-Sleep -Seconds 1
        }
    } finally {
        $watcher.Dispose()
    }
} else {
    # Single run
    $success = Run-Lint -LintPath $lintPath -TargetPath $Path -AutoFixMode $AutoFix
    
    if (-not $success) {
        Write-Host "`n⚠️  Fix errors before committing, or use 'git commit --no-verify' to skip" -ForegroundColor Yellow
        exit 1
    }
}
