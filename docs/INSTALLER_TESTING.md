# Installer Testing Guide

This guide explains how to test LlamaGate installers on different operating systems.

## Quick Test (Windows)

Run the comprehensive test script:
```powershell
.\test-all-installers.ps1
```

## Individual Tests

### Test Windows Installer Only
```powershell
.\test-installer-windows.ps1
```

### Test Unix Installer Only
```bash
# In WSL or Git Bash
chmod +x test-installer-unix.sh
./test-installer-unix.sh
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
- PowerShell syntax (Windows)
- Bash syntax (Unix)
- Function definitions
- File structure

### Functional Testing
- Dependency detection
- Configuration file creation
- Binary building
- Error handling

## Test Checklist

- [ ] Windows installer syntax is valid
- [ ] Unix installer syntax is valid
- [ ] Both installers handle missing dependencies
- [ ] Both installers create .env file correctly
- [ ] Both installers build the binary
- [ ] Silent mode works for both
- [ ] Skip flags work correctly

## Troubleshooting

### "Bash not found" error
- Install WSL: `wsl --install`
- Or install Git Bash
- Or use Docker to test Unix installer

### PowerShell execution policy error
- Run: `Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser`
- Or use: `powershell -ExecutionPolicy Bypass -File ...`

