# Installation Guide

LlamaGate can be installed in two ways:

## ⚡ Option 1: Use Installers (Recommended)

**Download and run directly from GitHub - no cloning required!**

**Windows (PowerShell):**
```powershell
# Binary installer (downloads pre-built binary)
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/smcdaniel54/LlamaGate/main/install/windows/install-binary.ps1" -OutFile install-binary.ps1; .\install-binary.ps1
```

**Unix/Linux/macOS:**
```bash
# Binary installer (downloads pre-built binary)
curl -fsSL https://raw.githubusercontent.com/smcdaniel54/LlamaGate/main/install/unix/install-binary.sh | bash
```

The installer will:
- ✅ Automatically detect your platform
- ✅ Download the correct binary from GitHub releases
- ✅ Set up the executable
- ✅ Create a default `.env` configuration file

**That's it!** You're ready to run LlamaGate.

**If you've already cloned the repository:**

**Windows:**
```cmd
install\windows\install-binary.cmd
```

**Unix/Linux/macOS:**
```bash
chmod +x install/unix/install-binary.sh
./install/unix/install-binary.sh
```

## 🔨 Option 2: Build from Source (For Developers)

If you need to build from source or want to customize the build:

**One-liner (download and run installer):**

**Windows (PowerShell):**
```powershell
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/smcdaniel54/LlamaGate/main/install/windows/install.ps1" -OutFile install.ps1; .\install.ps1
```

**Unix/Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/smcdaniel54/LlamaGate/main/install/unix/install.sh | bash
```

**From cloned repository:**

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

**Manual build (if you already have Go installed):**

```bash
# Clone the repository
git clone https://github.com/smcdaniel54/LlamaGate.git
cd LlamaGate

# Build
go build -o llamagate ./cmd/llamagate

# Or install to $GOPATH/bin
go install ./cmd/llamagate
```

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
- Use the source installer instead (Option 2)
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
- Build from source (Option 2)
- The installers automatically detect your platform
