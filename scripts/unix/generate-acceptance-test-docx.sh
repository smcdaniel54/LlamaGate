#!/bin/bash
# Generate DOCX from acceptance test markdown
# Requires: pandoc

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ACCEPTANCE_TEST_MD="$PROJECT_ROOT/docs/ACCEPTANCE_TEST.md"
OUTPUT_DOCX="$PROJECT_ROOT/docs/ACCEPTANCE_TEST.docx"

echo "Generating DOCX from acceptance test document..."

# Check for pandoc
if ! command -v pandoc &> /dev/null; then
    echo "Error: pandoc is not installed."
    echo "Install with:"
    echo "  - macOS: brew install pandoc"
    echo "  - Ubuntu/Debian: sudo apt-get install pandoc"
    echo "  - Or download from: https://pandoc.org/installing.html"
    exit 1
fi

# Generate DOCX using pandoc
pandoc "$ACCEPTANCE_TEST_MD" \
    -o "$OUTPUT_DOCX" \
    --toc \
    --toc-depth=2 \
    -V geometry:margin=1in \
    -V fontsize=11pt

if [ $? -eq 0 ]; then
    echo "✅ DOCX generated successfully: $OUTPUT_DOCX"
    echo ""
    echo "File size: $(du -h "$OUTPUT_DOCX" | cut -f1)"
else
    echo "❌ DOCX generation failed."
    echo ""
    echo "Alternative methods:"
    echo "1. Use online converter: https://cloudconvert.com/md-to-docx"
    echo "2. Use VS Code extension: 'Markdown PDF' (supports DOCX export)"
    exit 1
fi
