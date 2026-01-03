# LlamaGate Installer for Windows
# This script installs dependencies and sets up LlamaGate

param(
    [switch]$SkipGoCheck,
    [switch]$SkipOllamaCheck,
    [switch]$Silent
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "LlamaGate Installer" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if running as Administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "Warning: Not running as Administrator. Some features may require elevation." -ForegroundColor Yellow
    Write-Host ""
}

# Function to check if a command exists
function Test-Command {
    param([string]$Command)
    $null = Get-Command $Command -ErrorAction SilentlyContinue
    return $?
}

# Function to prompt user
function Get-UserInput {
    param(
        [string]$Prompt,
        [string]$Default = "",
        [switch]$Required
    )
    
    if ($Silent -and $Default) {
        return $Default
    }
    
    while ($true) {
        if ($Default) {
            $input = Read-Host "$Prompt [$Default]"
            if ([string]::IsNullOrWhiteSpace($input)) {
                $input = $Default
            }
        } else {
            $input = Read-Host $Prompt
        }
        
        if (-not [string]::IsNullOrWhiteSpace($input) -or -not $Required) {
            return $input
        }
        Write-Host "This field is required. Please enter a value." -ForegroundColor Red
    }
}

# Step 1: Check Go installation
Write-Host "[1/6] Checking Go installation..." -ForegroundColor Yellow
if (-not $SkipGoCheck) {
    if (Test-Command "go") {
        $goVersion = go version
        Write-Host "  ✓ Go is installed: $goVersion" -ForegroundColor Green
    } else {
        Write-Host "  ✗ Go is not installed" -ForegroundColor Red
        $installGo = Get-UserInput "Would you like to install Go? (Y/n)" "Y"
        
        if ($installGo -eq "Y" -or $installGo -eq "y" -or [string]::IsNullOrWhiteSpace($installGo)) {
            Write-Host "  Installing Go..." -ForegroundColor Yellow
            
            # Download Go installer
            $goVersion = "1.23.4"
            $goInstaller = "$env:TEMP\go$goVersion.windows-amd64.msi"
            $goUrl = "https://go.dev/dl/go$goVersion.windows-amd64.msi"
            
            try {
                Write-Host "  Downloading Go installer..." -ForegroundColor Yellow
                Invoke-WebRequest -Uri $goUrl -OutFile $goInstaller -UseBasicParsing
                
                Write-Host "  Running Go installer (this may require elevation)..." -ForegroundColor Yellow
                Start-Process msiexec.exe -ArgumentList "/i `"$goInstaller`" /quiet" -Wait -Verb RunAs
                
                # Refresh PATH
                $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
                
                # Verify installation
                Start-Sleep -Seconds 2
                if (Test-Command "go") {
                    Write-Host "  ✓ Go installed successfully" -ForegroundColor Green
                } else {
                    Write-Host "  ✗ Go installation may have failed. Please install manually from https://go.dev/dl/" -ForegroundColor Red
                    Write-Host "  After installing, restart this script." -ForegroundColor Yellow
                    exit 1
                }
            } catch {
                Write-Host "  ✗ Failed to install Go automatically: $_" -ForegroundColor Red
                Write-Host "  Please install Go manually from https://go.dev/dl/" -ForegroundColor Yellow
                exit 1
            }
        } else {
            Write-Host "  Skipping Go installation. Please install Go manually and restart this script." -ForegroundColor Yellow
            exit 1
        }
    }
} else {
    Write-Host "  ⚠ Skipping Go check (--SkipGoCheck)" -ForegroundColor Yellow
}

# Step 2: Check Ollama installation
Write-Host ""
Write-Host "[2/6] Checking Ollama installation..." -ForegroundColor Yellow
if (-not $SkipOllamaCheck) {
    if (Test-Command "ollama") {
        Write-Host "  ✓ Ollama is installed" -ForegroundColor Green
        $ollamaRunning = $false
        try {
            $null = Invoke-WebRequest -Uri "http://localhost:11434/api/tags" -UseBasicParsing -TimeoutSec 2 -ErrorAction Stop
            $ollamaRunning = $true
            Write-Host "  ✓ Ollama is running" -ForegroundColor Green
        } catch {
            Write-Host "  ⚠ Ollama is installed but not running" -ForegroundColor Yellow
            $startOllama = Get-UserInput "Would you like to start Ollama now? (Y/n)" "Y"
            if ($startOllama -eq "Y" -or $startOllama -eq "y") {
                Write-Host "  Starting Ollama..." -ForegroundColor Yellow
                Start-Process "ollama" -ArgumentList "serve" -WindowStyle Hidden
                Start-Sleep -Seconds 3
                Write-Host "  ✓ Ollama started" -ForegroundColor Green
            }
        }
    } else {
        Write-Host "  ✗ Ollama is not installed" -ForegroundColor Red
        $installOllama = Get-UserInput "Would you like to install Ollama? (Y/n)" "Y"
        
        if ($installOllama -eq "Y" -or $installOllama -eq "y") {
            Write-Host "  Opening Ollama download page..." -ForegroundColor Yellow
            Write-Host "  Please download and install Ollama from: https://ollama.com/download" -ForegroundColor Cyan
            Start-Process "https://ollama.com/download"
            
            $continue = Get-UserInput "Press Enter after you have installed Ollama to continue..."
            
            # Verify installation
            if (Test-Command "ollama") {
                Write-Host "  ✓ Ollama installed successfully" -ForegroundColor Green
            } else {
                Write-Host "  ✗ Ollama installation not detected. Please restart this script after installing." -ForegroundColor Red
                exit 1
            }
        } else {
            Write-Host "  ⚠ Skipping Ollama installation. LlamaGate requires Ollama to function." -ForegroundColor Yellow
        }
    }
} else {
    Write-Host "  ⚠ Skipping Ollama check (--SkipOllamaCheck)" -ForegroundColor Yellow
}

# Get project root (two levels up from install/windows/)
$ProjectRoot = (Get-Item $PSScriptRoot).Parent.Parent.FullName

# Step 3: Install Go dependencies
Write-Host ""
Write-Host "[3/6] Installing Go dependencies..." -ForegroundColor Yellow
try {
    Push-Location $ProjectRoot
    go mod download
    Write-Host "  ✓ Dependencies installed" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Failed to install dependencies: $_" -ForegroundColor Red
    exit 1
} finally {
    Pop-Location
}

# Step 4: Build LlamaGate
Write-Host ""
Write-Host "[4/6] Building LlamaGate..." -ForegroundColor Yellow
try {
    Push-Location $ProjectRoot
    go build -o llamagate.exe ./cmd/llamagate
    if (Test-Path "llamagate.exe") {
        Write-Host "  ✓ Build successful" -ForegroundColor Green
    } else {
        Write-Host "  ✗ Build failed - llamagate.exe not found" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "  ✗ Build failed: $_" -ForegroundColor Red
    exit 1
} finally {
    Pop-Location
}

# Step 5: Create configuration file
Write-Host ""
Write-Host "[5/6] Setting up configuration..." -ForegroundColor Yellow
$envFile = Join-Path $ProjectRoot ".env"
if (Test-Path $envFile) {
    Write-Host "  ⚠ .env file already exists" -ForegroundColor Yellow
    $overwrite = Get-UserInput "Would you like to overwrite it? (y/N)" "N"
    if ($overwrite -ne "y" -and $overwrite -ne "Y") {
        Write-Host "  ⚠ Keeping existing .env file" -ForegroundColor Yellow
    } else {
        $createEnv = $true
    }
} else {
    $createEnv = $true
}

if ($createEnv) {
    Write-Host "  Creating .env file..." -ForegroundColor Yellow
    
    $ollamaHost = Get-UserInput "Ollama host" "http://localhost:11434"
    $apiKey = Get-UserInput "API key (leave empty to disable authentication)" ""
    $rateLimit = Get-UserInput "Rate limit (requests per second)" "10"
    $debug = Get-UserInput "Enable debug logging? (true/false)" "false"
    $port = Get-UserInput "Server port" "8080"
    $logFile = Get-UserInput "Log file path (leave empty for console only)" ""
    
    $envContent = @"
# LlamaGate Configuration
# Generated by installer on $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

# Ollama server URL
OLLAMA_HOST=$ollamaHost

# API key for authentication (leave empty to disable authentication)
API_KEY=$apiKey

# Rate limit (requests per second)
RATE_LIMIT_RPS=$rateLimit

# Enable debug logging (true/false)
DEBUG=$debug

# Server port
PORT=$port

# Log file path (leave empty to log only to console)
LOG_FILE=$logFile
"@
    
    $envContent | Out-File -FilePath $envFile -Encoding utf8 -NoNewline
    Write-Host "  ✓ Configuration file created" -ForegroundColor Green
}

# Step 6: Create shortcuts (optional)
Write-Host ""
Write-Host "[6/6] Creating shortcuts..." -ForegroundColor Yellow
$createShortcuts = Get-UserInput "Would you like to create desktop shortcuts? (Y/n)" "Y"

if ($createShortcuts -eq "Y" -or $createShortcuts -eq "y" -or [string]::IsNullOrWhiteSpace($createShortcuts)) {
    $desktop = [Environment]::GetFolderPath("Desktop")
    
    # Shortcut to run LlamaGate
    $shortcutPath = Join-Path $desktop "LlamaGate.lnk"
    $shell = New-Object -ComObject WScript.Shell
    $shortcut = $shell.CreateShortcut($shortcutPath)
    $shortcut.TargetPath = "powershell.exe"
    $shortcut.Arguments = "-NoExit -Command `"cd '$ProjectRoot'; .\scripts\windows\run.cmd`""
    $shortcut.WorkingDirectory = $ProjectRoot
    $shortcut.Description = "Run LlamaGate"
    $shortcut.Save()
    Write-Host "  ✓ Desktop shortcut created" -ForegroundColor Green
}

# Summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Installation Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "LlamaGate has been installed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Quick Start:" -ForegroundColor Yellow
Write-Host "  1. Double-click 'LlamaGate.lnk' on your desktop, or" -ForegroundColor White
Write-Host "  2. Run: scripts\windows\run.cmd" -ForegroundColor White
Write-Host "  3. Or run: .\llamagate.exe" -ForegroundColor White
Write-Host ""
Write-Host "Configuration:" -ForegroundColor Yellow
Write-Host "  Edit .env file to change settings" -ForegroundColor White
Write-Host ""
Write-Host "Documentation:" -ForegroundColor Yellow
Write-Host "  See README.md for full documentation" -ForegroundColor White
Write-Host "  See docs\TESTING.md for testing instructions" -ForegroundColor White
Write-Host ""
Write-Host "Test the installation:" -ForegroundColor Yellow
Write-Host "  Run: scripts\windows\test.cmd" -ForegroundColor White
Write-Host ""

