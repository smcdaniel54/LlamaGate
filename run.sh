#!/bin/bash
# LlamaGate Universal Runner
# Detects OS and runs the appropriate script

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ "$OSTYPE" == "linux-gnu"* ]] || [[ "$OSTYPE" == "darwin"* ]]; then
    # Unix/Linux/macOS
    if [ -f "$SCRIPT_DIR/scripts/unix/run.sh" ]; then
        chmod +x "$SCRIPT_DIR/scripts/unix/run.sh"
        "$SCRIPT_DIR/scripts/unix/run.sh"
    else
        echo "Error: Unix runner not found"
        exit 1
    fi
else
    echo "Error: This script is for Unix/Linux/macOS only"
    echo "Windows users should use: scripts\windows\run.cmd"
    exit 1
fi

