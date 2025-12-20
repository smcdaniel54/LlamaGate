# LlamaGate Project Structure

This document describes the organized structure of LlamaGate, with OS-specific installers and scripts.

## Directory Structure

```
LlamaGate/
├── install/                    # Installation scripts organized by OS
│   ├── windows/                # Windows installers
│   │   ├── install.cmd        # Windows installer launcher
│   │   └── install.ps1        # Windows PowerShell installer
│   └── unix/                  # Unix/Linux/macOS installers
│       └── install.sh         # Unix/Linux/macOS installer
│
├── scripts/                    # Runtime scripts organized by OS
│   ├── windows/               # Windows scripts
│   │   ├── run.cmd            # Main runner
│   │   ├── run-with-auth.cmd  # Runner with authentication
│   │   ├── run-debug.cmd      # Runner with debug mode
│   │   ├── test.cmd           # Test script
│   │   └── build.cmd          # Build script
│   └── unix/                  # Unix/Linux/macOS scripts
│       ├── run.sh             # Main runner
│       └── test.sh            # Test script
│
├── cmd/llamagate/             # Application entry point
│   └── main.go
│
├── internal/                  # Internal packages
│   ├── cache/                 # Caching implementation
│   ├── config/                # Configuration management
│   ├── logger/                # Logging setup
│   ├── middleware/             # HTTP middleware
│   └── proxy/                 # Proxy handlers
│
├── install.sh                 # Universal installer (auto-detects OS)
├── run.sh                     # Universal runner (Unix/macOS)
├── test.sh                    # Universal test script (Unix/macOS)
│
├── README.md                  # Main documentation
├── INSTALL.md                 # Installation guide
├── TESTING.md                 # Testing guide
├── STRUCTURE.md               # This file
│
├── Dockerfile                 # Docker build file
├── .env.example              # Configuration template
├── go.mod                    # Go module definition
└── go.sum                    # Go dependencies checksum
```

## Universal Launchers

For convenience, root-level launchers are provided that automatically detect the OS and call the appropriate script:

### Installers
- **`install.sh`** (Unix/macOS) - Detects OS and launches appropriate installer
- **`install/windows/install.cmd`** (Windows) - Windows installer

### Runners
- **`run.sh`** (Unix/macOS) - Launches Unix runner
- **`scripts/windows/run.cmd`** (Windows) - Windows runner

### Test Scripts
- **`test.sh`** (Unix/macOS) - Launches Unix test script
- **`scripts/windows/test.cmd`** (Windows) - Windows test script

## Usage

### Installation

**Windows:**
```cmd
install\windows\install.cmd
```

**Unix/Linux/macOS:**
```bash
chmod +x install.sh
./install.sh
```
or
```bash
chmod +x install/unix/install.sh
./install/unix/install.sh
```

### Running

**Windows:**
```cmd
scripts\windows\run.cmd
```

**Unix/Linux/macOS:**
```bash
./run.sh
```
or
```bash
./scripts/unix/run.sh
```

### Testing

**Windows:**
```cmd
scripts\windows\test.cmd
```

**Unix/Linux/macOS:**
```bash
./test.sh
```
or
```bash
./scripts/unix/test.sh
```

## Benefits of This Structure

1. **Organization**: Clear separation of OS-specific files
2. **Maintainability**: Easy to find and update OS-specific code
3. **Clarity**: Users know which files are for their OS
4. **Flexibility**: Can add OS-specific features without cluttering root
5. **Convenience**: Root-level launchers provide easy access

## Adding New OS Support

To add support for a new OS:

1. Create `install/<new-os>/` directory
2. Create `scripts/<new-os>/` directory
3. Add OS-specific installers and scripts
4. Update `install.sh` to detect and route to new OS
5. Update this documentation

## Notes

- Unix scripts are shared between Linux and macOS (both use bash)
- Windows scripts use both `.cmd` (batch) and `.ps1` (PowerShell)
- All scripts in `scripts/` and `install/` directories are OS-specific
- Root-level `install.sh`, `run.sh`, and `test.sh` are convenience launchers for Unix/macOS

