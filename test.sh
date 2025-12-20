#!/bin/bash
# LlamaGate Universal Test Script
# Detects OS and runs the appropriate test script

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ "$OSTYPE" == "linux-gnu"* ]] || [[ "$OSTYPE" == "darwin"* ]]; then
    # Unix/Linux/macOS
    if [ -f "$SCRIPT_DIR/scripts/unix/test.sh" ]; then
        chmod +x "$SCRIPT_DIR/scripts/unix/test.sh"
        "$SCRIPT_DIR/scripts/unix/test.sh"
    else
        echo "Error: Unix test script not found"
        exit 1
    fi
else
    echo "Error: This script is for Unix/Linux/macOS only"
    echo "Windows users should use: scripts\windows\test.cmd"
    exit 1
fi

