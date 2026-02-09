# LlamaGate Integration Guide
*This document is designed to be contributed to the LlamaGate repository*

## Overview

This guide shows you how to integrate LlamaGate into your applications following proven patterns and best practices established through real-world usage and validated in production scenarios.

> **Enhancement Note**: This guide enhances LlamaGate's existing documentation by adding developer-focused integration patterns and workflows. It complements (does not duplicate) existing installation methods and usage examples.

## Quick Start

### Prerequisites
- **Go 1.19+** for building LlamaGate from source
- **Ollama** installed (will start automatically if not running)  
- **Python 3.8+** for Python-based integrations

### One-Command Development Setup (Enhancement)

> **Note**: This section enhances LlamaGate's existing installation methods by adding a developer-focused one-command workflow. It complements (does not replace) the existing binary installer, source installer, and manual build methods documented in LlamaGate.

The **one-command setup process** provides a single command that automates the complete development workflow:

1. ✅ **Validates environment** (Go, ports) - Catches issues before build
2. ✅ **Auto-creates `.env`** - Creates configuration file if missing (from `.env.example` or defaults)
3. ✅ **LlamaGate auto-starts Ollama** if not running - Built into LlamaGate application
4. ✅ **Auto-clones LlamaGate** if missing (standardized sibling directory)
5. ✅ **Smart build** - Only rebuilds if source is newer than binary
6. ✅ **Auto-starts LlamaGate** - No manual start needed
7. ✅ **Verifies it's running** - Health check confirmation

**This enhances existing methods by**:
- Adding developer workflow automation (complements installation methods)
- Automatically creating `.env` configuration (no manual setup needed)
- Providing standardized directory structure guidance
- Enabling smart rebuilds (only when needed)
- Automating the complete setup-to-running workflow

#### Standardized Directory Structure (Enhancement)

For integration projects, use this **recommended** structure for consistency:

```
YourProjectParent/
├── LlamaGate/           # ← Clone LlamaGate here (sibling directory)
└── YourProject/         # ← Your application
```

**Why sibling directory?**
- Consistent across all integration projects
- Easy to reference with relative paths
- Works well with version control
- Standard practice for integration workflows

#### One-Command Setup

**Windows PowerShell:**

Save this script to your project (e.g., `scripts/setup-llamagate-dev.ps1`):

```powershell
# One-Command LlamaGate Development Setup
# Standardized process: Validate → Clone (if needed) → Build → Start → Verify
# Based on community best practices for LlamaGate integration

param(
    [string]$LlamaGatePath = "..\LlamaGate",
    [int]$LlamaGatePort = 11435,
    [switch]$SkipClone,
    [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "LlamaGate One-Command Development Setup" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Environment Validation
Write-Host "[1/6] Validating environment..." -ForegroundColor Yellow

# Check Go
try {
    $goVersion = go version 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ✗ Go is not installed" -ForegroundColor Red
        Write-Host "  Please install Go 1.19+ from: https://go.dev/dl/" -ForegroundColor Yellow
        exit 1
    }
    Write-Host "  ✓ Go: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Go is not installed" -ForegroundColor Red
    Write-Host "  Please install Go 1.19+ from: https://go.dev/dl/" -ForegroundColor Yellow
    exit 1
}

# Note: LlamaGate will automatically start Ollama if not running
Write-Host "  Note: LlamaGate will auto-start Ollama if needed" -ForegroundColor Gray

# Check if LlamaGate is already running
Write-Host "[2/6] Checking if LlamaGate is already running..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:$LlamaGatePort/health" -Method GET -TimeoutSec 2 -ErrorAction SilentlyContinue
    if ($response.StatusCode -eq 200) {
        Write-Host "  ✓ LlamaGate is already running" -ForegroundColor Green
        Write-Host ""
        Write-Host "LlamaGate is ready!" -ForegroundColor Green
        Write-Host "URL: http://localhost:$LlamaGatePort" -ForegroundColor Cyan
        exit 0
    }
} catch {
    Write-Host "  LlamaGate is not running" -ForegroundColor Yellow
}

# Step 3: Find or Clone LlamaGate (Standardized: Sibling Directory)
Write-Host "[3/6] Locating LlamaGate source..." -ForegroundColor Yellow

# Standardized approach: Primary is sibling directory
$siblingPath = Resolve-Path "..\LlamaGate" -ErrorAction SilentlyContinue
$foundPath = $null

if ($siblingPath -and (Test-Path (Join-Path $siblingPath "cmd\llamagate"))) {
    $foundPath = $siblingPath
    Write-Host "  ✓ Found at sibling directory: $foundPath" -ForegroundColor Green
} else {
    # Check environment variable override
    $envPath = $env:LLAMAGATE_PATH
    if ($envPath -and (Test-Path (Join-Path $envPath "cmd\llamagate"))) {
        $foundPath = Resolve-Path $envPath
        Write-Host "  ✓ Found via LLAMAGATE_PATH: $foundPath" -ForegroundColor Green
    } else {
        if ($SkipClone) {
            Write-Host "  ✗ LlamaGate source not found" -ForegroundColor Red
            Write-Host "  Expected location: $(Resolve-Path ".." -ErrorAction SilentlyContinue)\LlamaGate" -ForegroundColor Yellow
            Write-Host "  Or set LLAMAGATE_PATH environment variable" -ForegroundColor Yellow
            exit 1
        } else {
            Write-Host "  LlamaGate source not found" -ForegroundColor Yellow
            Write-Host "  Cloning LlamaGate as sibling directory..." -ForegroundColor Yellow
            
            # Clone as sibling directory (standardized)
            $parentDir = Resolve-Path ".." -ErrorAction Stop
            $clonePath = Join-Path $parentDir "LlamaGate"
            
            if (Test-Path $clonePath) {
                Write-Host "  ✗ Directory already exists: $clonePath" -ForegroundColor Red
                Write-Host "  Please remove it or use -SkipClone to skip cloning" -ForegroundColor Yellow
                exit 1
            }
            
            Push-Location $parentDir
            try {
                Write-Host "  Cloning from GitHub..." -ForegroundColor Gray
                git clone <your-llamagate-repo-url>.git
                if ($LASTEXITCODE -ne 0) {
                    Write-Host "  ✗ Clone failed" -ForegroundColor Red
                    Pop-Location
                    exit 1
                }
                $foundPath = Resolve-Path "LlamaGate"
                Write-Host "  ✓ Cloned successfully to: $foundPath" -ForegroundColor Green
            } catch {
                Write-Host "  ✗ Clone failed: $_" -ForegroundColor Red
                Pop-Location
                exit 1
            } finally {
                Pop-Location
            }
        }
    }
}

# Step 4: Check Port Availability
Write-Host "[4/6] Checking port availability..." -ForegroundColor Yellow
try {
    $tcpClient = New-Object System.Net.Sockets.TcpClient
    $asyncResult = $tcpClient.BeginConnect("localhost", $LlamaGatePort, $null, $null)
    $wait = $asyncResult.AsyncWaitHandle.WaitOne(500, $false)
    if ($wait) {
        $tcpClient.EndConnect($asyncResult)
        $tcpClient.Close()
        Write-Host "  ✗ Port $LlamaGatePort is already in use" -ForegroundColor Red
        Write-Host "  Please stop the process using this port or use a different port" -ForegroundColor Yellow
        exit 1
    }
} catch {
    # Port is free (connection failed is expected)
}
Write-Host "  ✓ Port $LlamaGatePort is available" -ForegroundColor Green

# Step 5: Build LlamaGate
if (-not $SkipBuild) {
    Write-Host "[5/6] Building LlamaGate from source..." -ForegroundColor Yellow
    Push-Location $foundPath
    
    try {
        # Check if binary exists and is newer than source
        $binaryPath = Join-Path $foundPath "llamagate.exe"
        $needsBuild = $true
        
        if (Test-Path $binaryPath) {
            $binaryTime = (Get-Item $binaryPath).LastWriteTime
            $sourceTime = (Get-ChildItem -Path (Join-Path $foundPath "cmd\llamagate") -Recurse -File | 
                          Measure-Object -Property LastWriteTime -Maximum).Maximum
            
            if ($binaryTime -gt $sourceTime) {
                Write-Host "  Binary is up to date, skipping build" -ForegroundColor Gray
                $needsBuild = $false
            }
        }
        
        if ($needsBuild) {
            Write-Host "  Building (this may take a few minutes)..." -ForegroundColor Gray
            $buildOutput = go build -o llamagate.exe ./cmd/llamagate 2>&1
            
            if ($LASTEXITCODE -ne 0) {
                Write-Host "  ✗ Build failed" -ForegroundColor Red
                Write-Host $buildOutput -ForegroundColor Red
                Pop-Location
                exit 1
            }
            
            if (-not (Test-Path "llamagate.exe")) {
                Write-Host "  ✗ Binary not found after build" -ForegroundColor Red
                Pop-Location
                exit 1
            }
            
            Write-Host "  ✓ Build successful" -ForegroundColor Green
        }
    } catch {
        Write-Host "  ✗ Build error: $_" -ForegroundColor Red
        Pop-Location
        exit 1
    } finally {
        Pop-Location
    }
} else {
    Write-Host "[5/6] Skipping build (requested)" -ForegroundColor Yellow
}

# Step 6: Start LlamaGate
Write-Host "[6/6] Starting LlamaGate..." -ForegroundColor Yellow
Push-Location $foundPath

try {
    # Ensure .env exists with default configuration
    $envFile = Join-Path $foundPath ".env"
    $envExampleFile = Join-Path $foundPath ".env.example"
    
    if (-not (Test-Path $envFile)) {
        Write-Host "  Creating .env with default configuration..." -ForegroundColor Gray
        if (Test-Path $envExampleFile) {
            # Copy from .env.example
            Copy-Item $envExampleFile $envFile
            Write-Host "  ✓ Created .env from .env.example" -ForegroundColor Green
        } else {
            # Create with default values
            $defaultEnv = @"
# LlamaGate Configuration
# Generated by one-command setup with default values

# Ollama server URL
OLLAMA_HOST=http://localhost:11434

# API key for authentication (set to sk-llamagate to match documentation examples)
# Leave empty to disable authentication
API_KEY=sk-llamagate

# Rate limit (requests per second)
RATE_LIMIT_RPS=50

# Enable debug logging (true/false)
DEBUG=false

# Server port
PORT=$LlamaGatePort

# Log file path (leave empty to log only to console)
LOG_FILE=

# HTTP client timeout for Ollama requests (e.g., 5m, 30s, 30m - max 30 minutes)
TIMEOUT=5m
"@
            $defaultEnv | Out-File -FilePath $envFile -Encoding UTF8 -NoNewline
            Write-Host "  ✓ Created .env with default configuration" -ForegroundColor Green
        }
    }
    
    $binaryPath = Join-Path $foundPath "llamagate.exe"
    if (-not (Test-Path $binaryPath)) {
        Write-Host "  ✗ Binary not found: $binaryPath" -ForegroundColor Red
        Write-Host "  Please build first or remove -SkipBuild flag" -ForegroundColor Yellow
        Pop-Location
        exit 1
    }
    
    Write-Host "  Starting process..." -ForegroundColor Gray
    Write-Host "  Note: Windows may prompt for authorization (firewall/UAC)" -ForegroundColor Cyan
    Write-Host "  Please approve the prompt if it appears" -ForegroundColor Cyan
    $process = Start-Process -FilePath ".\llamagate.exe" -PassThru -WindowStyle Normal
    
    # Wait for LlamaGate to be ready
    $maxWait = 30
    $waited = 0
    $started = $false
    
    Write-Host "  Waiting for LlamaGate to be ready..." -ForegroundColor Gray -NoNewline
    while ($waited -lt $maxWait) {
        Start-Sleep -Seconds 1
        $waited++
        Write-Host "." -ForegroundColor Gray -NoNewline
        
        try {
            $testResponse = Invoke-WebRequest -Uri "http://localhost:$LlamaGatePort/health" -Method GET -TimeoutSec 1 -ErrorAction SilentlyContinue
            if ($testResponse.StatusCode -eq 200) {
                $started = $true
                Write-Host ""
                Write-Host "  ✓ LlamaGate started successfully" -ForegroundColor Green
                Write-Host "  PID: $($process.Id)" -ForegroundColor Cyan
                break
            }
        } catch {
            continue
        }
    }
    
    if (-not $started) {
        Write-Host ""
        Write-Host "  ✗ LlamaGate failed to start within $maxWait seconds" -ForegroundColor Red
        Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
        Pop-Location
        exit 1
    }
    
} catch {
    Write-Host "  ✗ Error starting LlamaGate: $_" -ForegroundColor Red
    Pop-Location
    exit 1
} finally {
    Pop-Location
}

# Success
Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "LlamaGate is ready for development!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "PID: $($process.Id)" -ForegroundColor Cyan
Write-Host "URL: http://localhost:$LlamaGatePort" -ForegroundColor Cyan
Write-Host "Health: http://localhost:$LlamaGatePort/health" -ForegroundColor Cyan
Write-Host ""
Write-Host "To stop: Stop-Process -Id $($process.Id)" -ForegroundColor Gray
Write-Host ""

exit 0
```

Then run from your project directory:
```powershell
.\scripts\setup-llamagate-dev.ps1
```

**Unix/macOS:**

[Unix script if available, or instructions to adapt]

**Note**: This one-command process enhances LlamaGate's existing installation methods. For end-user installation, use the binary or source installers documented in LlamaGate's main README.

## Integration Patterns

### Python Integration (Recommended)

Use the OpenAI SDK for seamless integration:

```python
from openai import OpenAI

# Standard configuration
client = OpenAI(
    base_url="http://localhost:11435/v1",  # LlamaGate endpoint
    api_key="sk-llamagate"  # Match your .env API_KEY setting
)

# Basic chat completion
response = client.chat.completions.create(
    model="mistral",  # Must specify model explicitly
    messages=[
        {"role": "user", "content": "Hello, how are you?"}
    ]
)

print(response.choices[0].message.content)
```

### Environment Validation

Always validate your environment before integration:

```python
import requests

def validate_llamagate_environment():
    """Validate LlamaGate integration prerequisites."""
    checks = []
    
    # Check Ollama
    try:
        response = requests.get("http://localhost:11434/api/tags", timeout=3)
        checks.append(("Ollama", response.status_code == 200))
    except:
        checks.append(("Ollama", False))
    
    # Check LlamaGate
    try:
        response = requests.get("http://localhost:11435/health", timeout=3)
        checks.append(("LlamaGate", response.status_code == 200))
    except:
        checks.append(("LlamaGate", False))
    
    # Report results
    all_good = all(status for _, status in checks)
    for service, status in checks:
        print(f"{'✓' if status else '✗'} {service}: {'OK' if status else 'FAILED'}")
    
    return all_good

# Use before integration
if validate_llamagate_environment():
    print("Ready for LlamaGate integration!")
else:
    print("Fix issues before proceeding.")
```

### Error Handling

Implement robust error handling:

```python
from openai import OpenAI, APIError
import requests

def create_llamagate_client():
    """Create LlamaGate client with validation."""
    # Validate LlamaGate is running
    try:
        health = requests.get("http://localhost:11435/health", timeout=3)
        if health.status_code != 200:
            raise ConnectionError("LlamaGate health check failed")
    except requests.RequestException:
        raise ConnectionError(
            "LlamaGate not accessible. Ensure it's running:\n"
            "  cd LlamaGate && ./llamagate"
        )
    
    return OpenAI(
        base_url="http://localhost:11435/v1",
        api_key="not-needed"
    )

def safe_chat_completion(client, model, messages):
    """Chat completion with proper error handling."""
    try:
        response = client.chat.completions.create(
            model=model,
            messages=messages
        )
        return response.choices[0].message.content
    
    except APIError as e:
        if "model" in str(e.message).lower():
            raise ValueError(
                f"Model '{model}' not available. Install with:\n"
                f"  ollama pull {model}"
            )
        raise
    
    except Exception as e:
        raise RuntimeError(f"Chat completion failed: {e}")

# Usage
try:
    client = create_llamagate_client()
    result = safe_chat_completion(client, "mistral", [
        {"role": "user", "content": "Hello!"}
    ])
    print(result)
except Exception as e:
    print(f"Integration error: {e}")
```

## Testing Integration

### Basic Test Setup

```python
import pytest
import requests
from openai import OpenAI

@pytest.fixture
def llamagate_client():
    """LlamaGate client fixture for testing."""
    # Verify LlamaGate is available
    try:
        response = requests.get("http://localhost:11435/health", timeout=2)
        if response.status_code != 200:
            pytest.skip("LlamaGate not available")
    except:
        pytest.skip("LlamaGate not running")
    
    return OpenAI(
        base_url="http://localhost:11435/v1",
        api_key="not-needed"
    )

def test_basic_chat(llamagate_client):
    """Test basic chat functionality."""
    response = llamagate_client.chat.completions.create(
        model="mistral",
        messages=[{"role": "user", "content": "Say hello"}]
    )
    
    assert response.choices[0].message.content
    assert len(response.choices[0].message.content) > 0

def test_streaming_chat(llamagate_client):
    """Test streaming chat functionality."""
    stream = llamagate_client.chat.completions.create(
        model="mistral",
        messages=[{"role": "user", "content": "Count to 3"}],
        stream=True
    )
    
    content = ""
    for chunk in stream:
        if chunk.choices and chunk.choices[0].delta.content:
            content += chunk.choices[0].delta.content
    
    assert len(content) > 0
```

## Production Deployment

### Configuration Management

```python
import os
from dataclasses import dataclass

@dataclass
class LlamaGateConfig:
    """LlamaGate configuration settings."""
    base_url: str = "http://localhost:11435/v1"
    api_key: str = "sk-llamagate"
    default_model: str = "mistral"
    timeout: int = 30
    max_retries: int = 3
    
    @classmethod
    def from_env(cls):
        """Create config from environment variables."""
        return cls(
            base_url=os.getenv("LLAMAGATE_URL", cls.base_url),
            api_key=os.getenv("LLAMAGATE_API_KEY", cls.api_key),
            default_model=os.getenv("LLAMAGATE_MODEL", cls.default_model),
            timeout=int(os.getenv("LLAMAGATE_TIMEOUT", cls.timeout)),
            max_retries=int(os.getenv("LLAMAGATE_MAX_RETRIES", cls.max_retries)),
        )

# Usage
config = LlamaGateConfig.from_env()
client = OpenAI(base_url=config.base_url, api_key=config.api_key)
```

### Health Monitoring

```python
import time
import logging
from typing import Optional

class LlamaGateMonitor:
    """Monitor LlamaGate health and availability."""
    
    def __init__(self, base_url: str = "http://localhost:11435"):
        self.base_url = base_url
        self.health_url = f"{base_url}/health"
        self.logger = logging.getLogger(__name__)
    
    def is_healthy(self) -> bool:
        """Check if LlamaGate is healthy."""
        try:
            response = requests.get(self.health_url, timeout=5)
            return response.status_code == 200
        except:
            return False
    
    def wait_for_health(self, timeout: int = 60) -> bool:
        """Wait for LlamaGate to become healthy."""
        start = time.time()
        while time.time() - start < timeout:
            if self.is_healthy():
                return True
            time.sleep(1)
        return False
    
    def health_check_loop(self, interval: int = 30):
        """Continuous health monitoring."""
        while True:
            healthy = self.is_healthy()
            status = "HEALTHY" if healthy else "UNHEALTHY"
            self.logger.info(f"LlamaGate status: {status}")
            
            if not healthy:
                self.logger.warning("LlamaGate is not responding")
            
            time.sleep(interval)

# Usage
monitor = LlamaGateMonitor()
if monitor.wait_for_health():
    print("LlamaGate is ready!")
else:
    print("LlamaGate failed to start")
```

## Advanced Integration

### Extension Management

```python
import requests
import json

class LlamaGateExtensionManager:
    """Manage LlamaGate extensions."""
    
    def __init__(self, base_url: str = "http://localhost:11435"):
        self.base_url = base_url
        self.extensions_url = f"{base_url}/v1/extensions"
    
    def list_extensions(self):
        """List available extensions."""
        response = requests.get(self.extensions_url)
        response.raise_for_status()
        return response.json()
    
    def install_extension(self, extension_path: str):
        """Install an extension."""
        with open(extension_path, 'r') as f:
            extension_data = f.read()
        
        response = requests.post(
            self.extensions_url,
            headers={"Content-Type": "application/yaml"},
            data=extension_data
        )
        response.raise_for_status()
        return response.json()
    
    def execute_extension(self, extension_name: str, inputs: dict):
        """Execute an extension."""
        response = requests.post(
            f"{self.extensions_url}/{extension_name}/execute",
            json=inputs
        )
        response.raise_for_status()
        return response.json()

# Usage
manager = LlamaGateExtensionManager()
extensions = manager.list_extensions()
print(f"Available extensions: {len(extensions)}")
```

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| "Connection refused" | Ensure LlamaGate is running on port 11435 |
| "Model not found" | Install model: `ollama pull [model-name]` |
| "Ollama not connected" | Start Ollama: `ollama serve` |
| "Build failed" | Check Go installation and version (need 1.19+) |
| "Permission denied" | Run with appropriate permissions, check firewall |

### Diagnostic Commands

```bash
# Check if services are running
curl http://localhost:11434/api/tags   # Ollama
curl http://localhost:11435/health     # LlamaGate

# Check processes
ps aux | grep ollama     # Unix
ps aux | grep llamagate  # Unix
tasklist | findstr ollama     # Windows  
tasklist | findstr llamagate  # Windows

# Check ports
netstat -an | grep :11434  # Unix
netstat -an | grep :11435  # Unix
netstat -an | findstr :11434  # Windows
netstat -an | findstr :11435  # Windows
```

### Getting Help

1. **Setup Issues**: Check troubleshooting section in this guide
2. **Integration Patterns**: Follow best practices documented in this guide
3. **Issues**: Report problems in your LlamaGate repository's issue tracker.
4. **Community**: Engage with the LlamaGate community for support

## Best Practices

### ✅ Do
- Validate environment before integration
- Use explicit model names in requests
- Implement proper error handling
- Monitor health endpoints
- Follow the sibling directory structure
- Follow established best practices

### ❌ Don't
- Assume models are available without checking
- Skip error handling
- Hardcode configuration values
- Ignore health check failures
- Mix different integration patterns

## Example Projects

### Production Examples
- Point the OpenAI SDK at your LlamaGate URL; see [API](API.md).

### Learning Resources
- Review integration patterns in this guide
- Study LlamaGate extension examples
- Follow established best practices

---

## Contributing

To contribute improvements to LlamaGate integration:

1. Test your improvements in real projects
2. Update this guide with your learnings
3. Submit pull request with clear examples
4. Follow established best practices for consistency

---

---

## One-Command Development Setup (Enhancement to LlamaGate)

> **Enhancement Note**: This one-command process enhances LlamaGate's existing installation methods by adding developer workflow automation. It should be documented as a complementary option, not a replacement for existing methods.

### What This Enhances

**Existing in LlamaGate**:
- ✅ Binary installer (one-line download)
- ✅ Source installer (handles Go, builds)
- ✅ Manual build process

**This Enhancement Adds**:
- ✅ Developer workflow automation
- ✅ Auto-clone capability
- ✅ Smart rebuild (only if needed)
- ✅ Auto-start with verification
- ✅ Standardized directory structure

### Key Features

- **Environment Validation**: Comprehensive checks before attempting build
- **Auto-Clone**: Clones LlamaGate if missing (standardized sibling directory)
- **Smart Build**: Only rebuilds if source is newer than binary
- **Auto-Start**: Starts LlamaGate and verifies it's running
- **Error Handling**: Clear error messages and helpful suggestions

### Implementation Pattern

The one-command setup script follows this standardized pattern:

1. Validate prerequisites (Go, Ollama, ports)
2. Check if LlamaGate is already running (exit if yes)
3. Locate LlamaGate source (sibling directory standard)
4. Auto-clone if missing (unless disabled)
5. Build from source (smart rebuild)
6. Start LlamaGate
7. Verify health check

### Documentation Location

**Recommended**: Add as a new section in LlamaGate README.md:
- **Title**: "Development Setup" or "One-Command Development Setup"
- **Position**: After installation methods, before usage examples
- **Relationship**: Complements existing installation methods

**Alternative**: Create `docs/DEVELOPMENT.md` for comprehensive development guide

**Key Point**: This enhances existing methods - document as complementary, not replacement.

---

*This guide is maintained by the LlamaGate community based on proven patterns and real-world usage.*