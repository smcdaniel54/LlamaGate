# Install golangci-lint by downloading pre-built binary
# This bypasses the Go module dependency issue

$ErrorActionPreference = "Stop"

Write-Host "Installing golangci-lint..." -ForegroundColor Cyan

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

# Use v2.x to match CI (golangci-lint-action uses 'latest' which resolves to v2.8.0+)
# This ensures compatibility with .golangci.yml version: 2
$version = "2.8.0"
$url = "https://github.com/golangci/golangci-lint/releases/download/v$version/golangci-lint-$version-windows-amd64.zip"
$zipFile = "$env:TEMP\golangci-lint.zip"
$extractDir = "$env:TEMP\golangci-lint"

try {
    Write-Host "Downloading golangci-lint v$version from GitHub releases..." -ForegroundColor Yellow
    Invoke-WebRequest -Uri $url -OutFile $zipFile -UseBasicParsing
    
    Write-Host "Extracting..." -ForegroundColor Yellow
    if (Test-Path $extractDir) {
        Remove-Item $extractDir -Recurse -Force
    }
    Expand-Archive -Path $zipFile -DestinationPath $extractDir -Force
    
    $exePath = "$extractDir\golangci-lint-$version-windows-amd64\golangci-lint.exe"
    if (Test-Path $exePath) {
        Copy-Item $exePath -Destination "$binDir\golangci-lint.exe" -Force
        Write-Host "Installed to $binDir\golangci-lint.exe" -ForegroundColor Green
        
        # Verify installation
        & "$binDir\golangci-lint.exe" --version
    } else {
        Write-Host "Error: Executable not found in archive" -ForegroundColor Red
        exit 1
    }
} finally {
    # Cleanup
    if (Test-Path $zipFile) {
        Remove-Item $zipFile -Force -ErrorAction SilentlyContinue
    }
    if (Test-Path $extractDir) {
        Remove-Item $extractDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

Write-Host "Installation complete!" -ForegroundColor Green

