# LlamaGate Installation Guide

This guide explains how to install LlamaGate on Windows using the automated installer.

## Quick Install

1. **Download or clone** the LlamaGate repository
2. **Run the installer:**
   ```cmd
   install.cmd
   ```
   Or simply double-click `install.cmd`

3. **Follow the prompts** - the installer will guide you through the process

## What the Installer Does

The installer performs the following steps:

### 1. Checks Go Installation
- Verifies if Go is installed
- If not found, offers to download and install Go automatically
- Requires administrator privileges for installation

### 2. Checks Ollama Installation
- Verifies if Ollama is installed
- Checks if Ollama is running
- Offers to start Ollama if installed but not running
- Guides you to install Ollama if not found

### 3. Installs Go Dependencies
- Downloads all required Go packages
- Sets up the Go module

### 4. Builds LlamaGate
- Compiles the `llamagate.exe` binary
- Creates a ready-to-run executable

### 5. Creates Configuration
- Prompts for configuration values:
  - Ollama host (default: http://localhost:11434)
  - API key (optional)
  - Rate limit (default: 10 RPS)
  - Debug mode (default: false)
  - Server port (default: 8080)
  - Log file path (optional)
- Creates `.env` file with your settings

### 6. Creates Shortcuts (Optional)
- Creates a desktop shortcut to run LlamaGate
- Makes it easy to start the application

## Manual Installation

If you prefer to install manually:

### Prerequisites

1. **Install Go** (1.23+)
   - Download from: https://go.dev/dl/
   - Or use: `winget install GoLang.Go`

2. **Install Ollama**
   - Download from: https://ollama.com/download
   - Or use: `winget install Ollama.Ollama`

### Steps

1. **Clone or download** the repository
2. **Open terminal** in the project directory
3. **Install dependencies:**
   ```cmd
   go mod download
   ```
4. **Build the binary:**
   ```cmd
   go build -o llamagate.exe ./cmd/llamagate
   ```
5. **Create `.env` file:**
   ```cmd
   copy .env.example .env
   ```
6. **Edit `.env`** with your settings

## Installation Options

### Silent Installation

Run the installer with silent mode (uses defaults):
```powershell
.\install.ps1 -Silent
```

### Skip Checks

Skip Go check:
```powershell
.\install.ps1 -SkipGoCheck
```

Skip Ollama check:
```powershell
.\install.ps1 -SkipOllamaCheck
```

## Post-Installation

After installation:

1. **Verify installation:**
   ```cmd
   llamagate.exe --version
   ```
   (Note: version flag may not be implemented yet)

2. **Test the installation:**
   ```cmd
   test.cmd
   ```

3. **Start LlamaGate:**
   ```cmd
   run.cmd
   ```
   Or use the desktop shortcut

## Troubleshooting

### Go Installation Fails

- **Issue:** Installer can't install Go automatically
- **Solution:** 
  1. Download Go manually from https://go.dev/dl/
  2. Install it
  3. Restart the installer

### Ollama Not Found

- **Issue:** Ollama installation not detected
- **Solution:**
  1. Install Ollama from https://ollama.com/download
  2. Restart your terminal
  3. Run `ollama serve` to start Ollama
  4. Re-run the installer

### Build Fails

- **Issue:** `go build` fails
- **Solution:**
  1. Ensure Go is in your PATH: `go version`
  2. Check internet connection (needed for dependencies)
  3. Try: `go mod tidy` then rebuild

### Permission Errors

- **Issue:** Access denied errors
- **Solution:**
  1. Run PowerShell as Administrator
  2. Or run `install.cmd` as Administrator

### Port Already in Use

- **Issue:** Port 8080 is already in use
- **Solution:**
  1. Change `PORT` in `.env` file
  2. Or stop the service using port 8080

## Uninstallation

To uninstall LlamaGate:

1. **Stop the application** if running
2. **Delete the LlamaGate folder**
3. **Remove desktop shortcuts** (if created)
4. **Optional:** Remove Go and Ollama if not used elsewhere

## Next Steps

After installation:

1. Read the [README.md](README.md) for usage instructions
2. Check [TESTING.md](TESTING.md) to test your installation
3. Configure your `.env` file as needed
4. Start using LlamaGate!

## Support

If you encounter issues:

1. Check the troubleshooting section above
2. Review the logs in `llamagate.log` (if configured)
3. Ensure Ollama is running: `ollama serve`
4. Verify Go is installed: `go version`

