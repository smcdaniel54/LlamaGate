# Installation Guide

LlamaGate can be installed in several ways, from easiest to most flexible:

## 🚀 Method 1: Pre-built Binaries (Easiest - Recommended)

**No Go installation required!** Just download and run.

### Windows

1. **Download the binary:**
   - Go to [Releases](https://github.com/llamagate/llamagate/releases/latest)
   - Download `llamagate-windows-amd64.exe`

2. **Run it:**
   ```cmd
   llamagate-windows-amd64.exe
   ```

3. **Optional:** Rename to `llamagate.exe` and add to PATH for easier access

### Linux

```bash
# Download
curl -LO https://github.com/llamagate/llamagate/releases/latest/download/llamagate-linux-amd64

# Make executable
chmod +x llamagate-linux-amd64

# Run
./llamagate-linux-amd64
```

**For ARM64 (Raspberry Pi, etc.):**
```bash
curl -LO https://github.com/llamagate/llamagate/releases/latest/download/llamagate-linux-arm64
chmod +x llamagate-linux-arm64
./llamagate-linux-arm64
```

### macOS

**Apple Silicon (M1/M2/M3):**
```bash
curl -LO https://github.com/llamagate/llamagate/releases/latest/download/llamagate-darwin-arm64
chmod +x llamagate-darwin-arm64
./llamagate-darwin-arm64
```

**Intel Mac:**
```bash
curl -LO https://github.com/llamagate/llamagate/releases/latest/download/llamagate-darwin-amd64
chmod +x llamagate-darwin-amd64
./llamagate-darwin-amd64
```

### Verify Installation

After downloading, verify the binary works:

```bash
# Linux/macOS
./llamagate-* --help

# Windows
llamagate-windows-amd64.exe --help
```

You should see usage information. If you get a "command not found" error, make sure the file is executable (Linux/macOS) and in your current directory.

## 🔧 Method 2: Automated Installer

The installer script will:
- Check for Go and install it if needed
- Check for Ollama and guide you to install it
- Build LlamaGate from source
- Create a `.env` configuration file

### Windows

```cmd
install\windows\install.cmd
```

### Unix/Linux/macOS

```bash
chmod +x install/unix/install.sh
./install/unix/install.sh
```

**Silent mode** (uses defaults, no prompts):
```bash
./install/unix/install.sh --silent
```

## 💻 Method 3: Build from Source

If you have Go installed and want to build yourself:

```bash
# Clone the repository
git clone https://github.com/llamagate/llamagate.git
cd llamagate

# Build
go build -o llamagate ./cmd/llamagate

# Run
./llamagate
```

## 📦 Method 4: Using Go Install

If you have Go installed:

```bash
go install github.com/llamagate/llamagate/cmd/llamagate@latest
```

This installs to `$GOPATH/bin` (or `$HOME/go/bin` by default).

## 🐳 Method 5: Docker

```bash
# Build
docker build -t llamagate .

# Run
docker run -p 8080:8080 llamagate
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
   # Using pre-built binary
   ./llamagate-linux-amd64
   
   # Or if built from source
   ./llamagate
   ```

2. **Verify it's running:**
   ```bash
   curl http://localhost:8080/health
   ```

3. **See [Quick Start Guide](../QUICKSTART.md)** for usage examples

## Troubleshooting

### "Permission denied" (Linux/macOS)

Make the binary executable:
```bash
chmod +x llamagate-*
```

### "Command not found"

- Make sure you're in the directory where you downloaded the binary
- Or add the directory to your PATH
- Or use the full path: `/path/to/llamagate-*`

### Binary won't run

- Check it's the correct architecture for your system
- Verify the download completed (check file size)
- Try re-downloading from [Releases](https://github.com/llamagate/llamagate/releases)

### Need a different architecture?

Check [Releases](https://github.com/llamagate/llamagate/releases) for:
- Linux: amd64, arm64
- macOS: amd64 (Intel), arm64 (Apple Silicon)
- Windows: amd64

If you need a different architecture, build from source (Method 3).
