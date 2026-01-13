# Test script for race condition detection (matches CI)
# Run: .\scripts\windows\test-race.ps1
# This matches the CI test-race job configuration

$ErrorActionPreference = "Stop"

Write-Host "Running race condition tests (matches CI)..." -ForegroundColor Cyan

# Check Go version matches CI (1.24)
$goVersion = go version
Write-Host "Go version: $goVersion" -ForegroundColor Yellow

# Set CGO_ENABLED=1 (required for race detector, matches CI)
$env:CGO_ENABLED = "1"

# Check if C compiler is available (required for CGO on Windows)
$gccAvailable = Get-Command gcc -ErrorAction SilentlyContinue
if (-not $gccAvailable) {
    Write-Host "Warning: C compiler (gcc) not found in PATH." -ForegroundColor Yellow
    Write-Host "Race detector requires CGO, which needs a C compiler on Windows." -ForegroundColor Yellow
    Write-Host "Install MinGW-w64 or TDM-GCC, or run race tests in CI/Linux environment." -ForegroundColor Yellow
    Write-Host "Attempting to continue anyway..." -ForegroundColor Yellow
}

# Get packages, filtering out .out directories (matches CI logic)
Write-Host "Discovering packages..." -ForegroundColor Yellow
$packages = go list ./... 2>$null | Where-Object { $_ -notmatch "^\.out$" }
if (-not $packages) {
    $packages = go list ./...
}

Write-Host "Found $($packages.Count) packages to test" -ForegroundColor Yellow

# Run tests with race detector (matches CI: go test -race -timeout=10m)
Write-Host "Running tests with race detector (timeout: 10m)..." -ForegroundColor Cyan
$testOutput = go test -race -timeout=10m $packages 2>&1
$exitCode = $LASTEXITCODE

if ($exitCode -eq 0) {
    Write-Host "`nRace condition tests passed!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`nRace condition tests failed!" -ForegroundColor Red
    Write-Host $testOutput
    exit $exitCode
}
