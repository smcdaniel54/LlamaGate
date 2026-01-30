# Installation Guide

LlamaGate can be installed using three methods:

## ⚡ Method 1: One-Line Command (Recommended)

**Copy and paste one command - it downloads the installer and runs it automatically!**

This method downloads a pre-built binary (no Go required):

**Windows (PowerShell):**
```powershell
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/smcdaniel54/LlamaGate/main/install/windows/install-binary.ps1" -OutFile install-binary.ps1; .\install-binary.ps1
```

**Unix/Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/smcdaniel54/LlamaGate/main/install/unix/install-binary.sh | bash
```

**What happens:**
1. Downloads the installer script from GitHub
2. Runs the installer automatically
3. Installer downloads the pre-built binary for your platform
4. Sets up the executable and creates `.env` configuration file

**That's it!** You're ready to run LlamaGate.

## 🔧 Method 2: Run Installer Directly (If You've Cloned the Repo)

If you've already cloned the repository, you can run the installer directly:

**Binary installer (downloads pre-built binary):**

**Windows:**
```cmd
install\windows\install-binary.cmd
```

**Unix/Linux/macOS:**
```bash
chmod +x install/unix/install-binary.sh
./install/unix/install-binary.sh
```

**Source installer (builds from source):**

**Windows:**
```cmd
install\windows\install.cmd
```

**Unix/Linux/macOS:**
```bash
chmod +x install/unix/install.sh
./install/unix/install.sh
```

The source installer will:
- ✅ Check for Go and install it if needed
- ✅ Check for Ollama and guide you to install it
- ✅ Install all Go dependencies
- ✅ Build the LlamaGate binary from source
- ✅ Create a `.env` configuration file

## 🔨 Method 3: Build from Source (For Developers)

If you need to build from source, you have two options:

### Option A: One-Line Command (Downloads Source Installer)

**Windows (PowerShell):**
```powershell
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/smcdaniel54/LlamaGate/main/install/windows/install.ps1" -OutFile install.ps1; .\install.ps1
```

**Unix/Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/smcdaniel54/LlamaGate/main/install/unix/install.sh | bash
```

This downloads and runs the source installer, which handles Go installation and builds from source.

### Option B: Manual Build (If You Have Go Installed)

If you already have Go installed and want to build manually: run `go build ./...` first so that all packages compile (required for downstream CI, E2E, and build-from-source tooling), then build the main binary.

**Unix/Linux/macOS:**
```bash
# Clone the repository
git clone https://github.com/smcdaniel54/LlamaGate.git
cd LlamaGate

# Build all packages (required for CI/E2E/build-from-source integrators)
go build ./...

# Build main binary
go build -o llamagate ./cmd/llamagate

# Or install to $GOPATH/bin
go install ./cmd/llamagate
```

**Windows (PowerShell):**
```powershell
# Clone the repository (handle stderr output)
$ErrorActionPreference = "Continue"  # Git writes progress to stderr
git clone https://github.com/smcdaniel54/LlamaGate.git
$ErrorActionPreference = "Stop"  # Restore if needed

cd LlamaGate

# Build all packages (required for CI/E2E/build-from-source integrators)
go build ./...

# Build main binary
go build -o llamagate.exe ./cmd/llamagate

# Or install to $GOPATH/bin
go install ./cmd/llamagate
```

**Note:** Git writes progress messages to stderr even on success. For integrators (CI, E2E, forked automation), ensuring `go build ./...` succeeds is required; see [Testing Guide](TESTING.md) for test commands. In PowerShell with `$ErrorActionPreference = "Stop"`, this can cause failures. See [Troubleshooting](#troubleshooting) section below for details.

## Configuration

After installation, create a `.env` file (or use environment variables):

```bash
# Copy example
cp .env.example .env

# Edit as needed
# Windows: notepad .env
# Linux/macOS: nano .env
```

See [Configuration](#configuration) section in README.md for all options.

## Next Steps

1. **Start LlamaGate:**
   ```bash
   # Using installer (binary will be in project root)
   ./llamagate
   
   # Or if built from source
   ./llamagate
   ```

2. **Verify it's running:**
   ```bash
   curl http://localhost:11435/health
   ```

3. **See [Quick Start Guide](../QUICKSTART.md)** for usage examples

## Troubleshooting

### Installer fails with 404 error

If the binary installer fails because binaries aren't available yet:
- Use the source installer instead (Method 2 or Method 3)
- Or wait for binaries to be published to releases

### "Permission denied" (Linux/macOS)

Make the binary executable:
```bash
chmod +x llamagate
```

### "Command not found"

- Make sure you're in the directory where the binary was installed
- Or add the directory to your PATH
- Or use the full path: `/path/to/llamagate`

### Need a different architecture?

If you need a different architecture than what's available:
- Build from source (Method 3)
- The installers automatically detect your platform

### Git clone fails in PowerShell with "Stop" error action

**Issue:** When using PowerShell with `$ErrorActionPreference = "Stop"`, `git clone` may fail even though the clone succeeds. This happens because Git writes progress messages to stderr, which PowerShell treats as errors.

**Symptoms:**
- Script fails with error even though `git clone` completes successfully
- Error message appears but repository is actually cloned

**Solution (PowerShell):**

**Option 1: Temporarily change error action (Recommended)**
```powershell
# Save current setting
$oldErrorAction = $ErrorActionPreference

# Temporarily allow errors during git clone
$ErrorActionPreference = "Continue"

# Clone the repository
git clone https://github.com/smcdaniel54/LlamaGate.git

# Restore original setting
$ErrorActionPreference = $oldErrorAction
```

**Option 2: Check exit code instead**
```powershell
# Clone and check exit code
git clone https://github.com/smcdaniel54/LlamaGate.git 2>&1 | Out-Null
if ($LASTEXITCODE -ne 0) {
    Write-Error "Git clone failed with exit code $LASTEXITCODE"
    exit 1
}
```

**Option 3: Redirect stderr**
```powershell
# Redirect stderr to null (suppress progress messages)
git clone https://github.com/smcdaniel54/LlamaGate.git 2>$null
```

**Note:** This is a known Git behavior - progress messages go to stderr even on success. The installers handle this automatically, but manual `git clone` commands in PowerShell scripts need this workaround.
