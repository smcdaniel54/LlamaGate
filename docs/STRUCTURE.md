# LlamaGate Project Structure

This document describes the organized structure of LlamaGate, with OS-specific installers and scripts.

## Directory Structure

```text
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
├── docs/                      # Documentation
│   ├── README.md              # Documentation index
│   ├── INSTALL.md             # Installation guide
│   ├── TESTING.md             # Testing guide
│   ├── STRUCTURE.md           # This file
│   └── INSTALLER_TESTING.md   # Installer testing guide
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

## Script Organization

All scripts are organized by OS in their respective directories. Use the scripts directly from their OS-specific directories:

### Installers

- **`install/windows/install.cmd`** (Windows) - Windows installer
- **`install/unix/install.sh`** (Unix/Linux/macOS) - Unix installer

### Runners

- **`scripts/windows/run.cmd`** (Windows) - Main Windows runner
- **`scripts/windows/run-with-auth.cmd`** (Windows) - Runner with authentication
- **`scripts/windows/run-debug.cmd`** (Windows) - Runner with debug mode
- **`scripts/unix/run.sh`** (Unix/Linux/macOS) - Main Unix runner

### Test Scripts

- **`scripts/windows/test.cmd`** (Windows) - Windows test script
- **`scripts/unix/test.sh`** (Unix/Linux/macOS) - Unix test script

## Usage

### Installation

**Windows:**

```cmd
install\windows\install.cmd
```

**Unix/Linux/macOS:**

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
./scripts/unix/run.sh
```

### Testing

**Windows:**

```cmd
scripts\windows\test.cmd
```

**Unix/Linux/macOS:**

```bash
./scripts/unix/test.sh
```

## Benefits of This Structure

1. **Organization**: Clear separation of OS-specific files
2. **Maintainability**: Easy to find and update OS-specific code
3. **Clarity**: Users know which files are for their OS
4. **Flexibility**: Can add OS-specific features without cluttering root
5. **Consistency**: All scripts follow the same directory structure pattern

## Adding New OS Support

To add support for a new OS:

1. Create `install/<new-os>/` directory
2. Create `scripts/<new-os>/` directory
3. Add OS-specific installers and scripts
4. Update this documentation

## Notes

- Unix scripts are shared between Linux and macOS (both use bash)
- Windows scripts use both `.cmd` (batch) and `.ps1` (PowerShell)
- All scripts in `scripts/` and `install/` directories are OS-specific
- Scripts are accessed directly from their OS-specific directories
