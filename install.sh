#!/bin/bash
# LlamaGate Universal Installer Launcher
# Detects OS and launches the appropriate installer

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Detect OS
if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
    # Windows
    echo "Detected Windows. Launching Windows installer..."
    if [ -f "$SCRIPT_DIR/install/windows/install.cmd" ]; then
        "$SCRIPT_DIR/install/windows/install.cmd"
    else
        echo "Error: Windows installer not found at install/windows/install.cmd"
        exit 1
    fi
elif [[ "$OSTYPE" == "linux-gnu"* ]] || [[ "$OSTYPE" == "darwin"* ]]; then
    # Unix/Linux/macOS
    echo "Detected Unix/Linux/macOS. Launching Unix installer..."
    if [ -f "$SCRIPT_DIR/install/unix/install.sh" ]; then
        chmod +x "$SCRIPT_DIR/install/unix/install.sh"
        "$SCRIPT_DIR/install/unix/install.sh" "$@"
    else
        echo "Error: Unix installer not found at install/unix/install.sh"
        exit 1
    fi
else
    echo "Error: Unsupported operating system: $OSTYPE"
    echo "Please use the OS-specific installer directly:"
    echo "  Windows: install/windows/install.cmd"
    echo "  Unix/Linux/macOS: install/unix/install.sh"
    exit 1
fi

