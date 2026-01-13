# Installer Testing Guide

This guide explains how to test LlamaGate installers on different operating systems.

## Quick Test (Windows)

Run the comprehensive test script:
```powershell
.\tests\installer\test-all-installers.ps1
```

## Individual Tests

### Test Windows Installer Only
```powershell
.\tests\installer\test-installer-windows.ps1
```

### Test Unix Installer Only
```bash
# In WSL or Git Bash
chmod +x tests/installer/test-installer-unix.sh
./tests/installer/test-installer-unix.sh
```

## Manual Testing

### Windows Installer

**Full test:**
```cmd
install\windows\install.cmd
```

**Quick test (skip dependency checks):**
```powershell
powershell -ExecutionPolicy Bypass -File install\windows\install.ps1 -SkipGoCheck -SkipOllamaCheck
```

**Silent mode test:**
```powershell
powershell -ExecutionPolicy Bypass -File install\windows\install.ps1 -Silent
```

### Unix Installer (from Windows)

**Using WSL:**
```bash
# In WSL terminal
cd /mnt/c/path/to/LlamaGate
chmod +x install/unix/install.sh
./install/unix/install.sh --skip-go-check --skip-ollama-check
```

**Using Git Bash:**
```bash
# In Git Bash
cd /c/path/to/LlamaGate
chmod +x install/unix/install.sh
./install/unix/install.sh --skip-go-check --skip-ollama-check
```

## What Gets Tested

### Syntax Validation
- **Binary installer syntax** (one-liner installers):
  - `install/windows/install-binary.ps1` - PowerShell syntax validation
  - `install/unix/install-binary.sh` - Bash syntax validation
- **Source installer syntax**:
  - `install/windows/install.ps1` - PowerShell syntax validation
  - `install/unix/install.sh` - Bash syntax validation
- Function definitions
- File structure

### Functional Testing
- Dependency detection
- Configuration file creation
- Binary building
- Error handling

## Test Checklist

- [ ] Windows binary installer (`install-binary.ps1`) syntax is valid
- [ ] Unix binary installer (`install-binary.sh`) syntax is valid
- [ ] Windows source installer (`install.ps1`) syntax is valid
- [ ] Unix source installer (`install.sh`) syntax is valid
- [ ] Binary installers download and execute correctly
- [ ] Source installers handle missing dependencies
- [ ] Source installers create .env file correctly
- [ ] Source installers build the binary
- [ ] Silent mode works for source installers
- [ ] Skip flags work correctly for source installers

## Troubleshooting

### "Bash not found" error
- Install WSL: `wsl --install`
- Or install Git Bash
- Or use Docker to test Unix installer

### PowerShell execution policy error
- Run: `Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser`
- Or use: `powershell -ExecutionPolicy Bypass -File ...`

