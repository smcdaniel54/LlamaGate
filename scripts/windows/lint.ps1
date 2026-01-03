# Lint script for Windows
# Run: .\scripts\windows\lint.ps1

$ErrorActionPreference = "Stop"

Write-Host "Running golangci-lint..." -ForegroundColor Cyan

# Get GOPATH
$gopath = go env GOPATH
if (-not $gopath) {
    Write-Host "Error: GOPATH not set" -ForegroundColor Red
    exit 1
}

# Check if golangci-lint is installed
$lintPath = "$gopath\bin\golangci-lint.exe"

if (-not (Test-Path $lintPath)) {
    Write-Host "golangci-lint not found. Installing via binary download..." -ForegroundColor Yellow
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

# Run the linter
& $lintPath run --timeout=15m

if ($LASTEXITCODE -ne 0) {
    Write-Host "Linting failed!" -ForegroundColor Red
    exit $LASTEXITCODE
}

Write-Host "Linting passed!" -ForegroundColor Green

