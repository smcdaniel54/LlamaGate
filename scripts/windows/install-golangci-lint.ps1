# Install golangci-lint using official method to match CI
# CI uses golangci-lint-action@v3 with version: latest
# This script uses the same official install method

$ErrorActionPreference = "Stop"

Write-Host "Installing golangci-lint using official method (matches CI)..." -ForegroundColor Cyan

# Get GOPATH
$gopath = go env GOPATH
if (-not $gopath) {
    Write-Host "Error: GOPATH not set" -ForegroundColor Red
    exit 1
}

$binDir = "$gopath\bin"
if (-not (Test-Path $binDir)) {
    New-Item -ItemType Directory -Path $binDir -Force | Out-Null
}

# Use official install script (same method CI uses internally)
# This ensures we get the same 'latest' version CI uses
try {
    Write-Host "Downloading and installing latest version..." -ForegroundColor Yellow
    $installScript = Invoke-WebRequest -Uri "https://golangci-lint.run/install.sh" -UseBasicParsing
    
    # Execute install script via bash (available on Windows via Git Bash or WSL)
    # If bash is not available, fall back to manual download
    if (Get-Command bash -ErrorAction SilentlyContinue) {
        $installScript.Content | bash -s -- -b $binDir latest
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Installed to $binDir\golangci-lint.exe" -ForegroundColor Green
            & "$binDir\golangci-lint.exe" --version
            Write-Host "Installation complete!" -ForegroundColor Green
            exit 0
        }
    }
    
    # Fallback: Download latest release manually
    Write-Host "Bash not available, downloading latest release manually..." -ForegroundColor Yellow
    $latestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/golangci/golangci-lint/releases/latest" -UseBasicParsing
    $version = $latestRelease.tag_name -replace '^v', ''
    $url = "https://github.com/golangci/golangci-lint/releases/download/v$version/golangci-lint-$version-windows-amd64.zip"
    $zipFile = "$env:TEMP\golangci-lint.zip"
    $extractDir = "$env:TEMP\golangci-lint"
    
    Invoke-WebRequest -Uri $url -OutFile $zipFile -UseBasicParsing
    
    if (Test-Path $extractDir) {
        Remove-Item $extractDir -Recurse -Force
    }
    Expand-Archive -Path $zipFile -DestinationPath $extractDir -Force
    
    $exePath = "$extractDir\golangci-lint-$version-windows-amd64\golangci-lint.exe"
    if (Test-Path $exePath) {
        Copy-Item $exePath -Destination "$binDir\golangci-lint.exe" -Force
        Write-Host "Installed golangci-lint v$version to $binDir\golangci-lint.exe" -ForegroundColor Green
        & "$binDir\golangci-lint.exe" --version
    } else {
        Write-Host "Error: Executable not found in archive" -ForegroundColor Red
        exit 1
    }
    
    # Cleanup
    Remove-Item $zipFile -Force -ErrorAction SilentlyContinue
    Remove-Item $extractDir -Recurse -Force -ErrorAction SilentlyContinue
    
} catch {
    Write-Host "Error installing golangci-lint: $_" -ForegroundColor Red
    exit 1
}

Write-Host "Installation complete!" -ForegroundColor Green

