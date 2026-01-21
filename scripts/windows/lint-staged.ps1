# Lint-staged script for Windows - Quick lint check on staged files only
# Run: .\scripts\windows\lint-staged.ps1
# Usage: Run this before committing to check only staged files (fast feedback)

$ErrorActionPreference = "Stop"

Write-Host "`n=== Lint-Staged: Quick Check ===" -ForegroundColor Cyan
Write-Host "Checking only staged files for fast feedback`n" -ForegroundColor Yellow

# Get GOPATH
$gopath = go env GOPATH
if (-not $gopath) {
    Write-Host "Error: GOPATH not set" -ForegroundColor Red
    exit 1
}

# Check if golangci-lint is installed
$lintPath = "$gopath\bin\golangci-lint.exe"

if (-not (Test-Path $lintPath)) {
    Write-Host "golangci-lint not found. Please install first:" -ForegroundColor Yellow
    Write-Host "  .\scripts\windows\install-golangci-lint.ps1" -ForegroundColor White
    exit 1
}

# Get staged Go files
$stagedFiles = git diff --cached --name-only --diff-filter=ACM | Where-Object { $_ -match '\.go$' }

if (-not $stagedFiles) {
    Write-Host "✅ No Go files staged, nothing to lint" -ForegroundColor Green
    exit 0
}

Write-Host "📝 Staged files:" -ForegroundColor Cyan
$stagedFiles | ForEach-Object { Write-Host "   $_" -ForegroundColor Gray }
Write-Host ""

# Run linting on staged files only
& $lintPath run --timeout=5m --tests $stagedFiles

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ Staged files passed linting! Ready to commit." -ForegroundColor Green
    exit 0
} else {
    Write-Host "`n❌ Linting failed on staged files" -ForegroundColor Red
    Write-Host "`n💡 Run full lint check:" -ForegroundColor Yellow
    Write-Host "   .\scripts\windows\lint-fix.ps1" -ForegroundColor White
    Write-Host "`n💡 Or auto-fix what's possible:" -ForegroundColor Yellow
    Write-Host "   .\scripts\windows\lint-fix.ps1 -AutoFix" -ForegroundColor White
    exit 1
}
