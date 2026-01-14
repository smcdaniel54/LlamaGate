#!/bin/bash
# Generate PDF from acceptance test markdown
# Requires: pandoc and a LaTeX distribution (or use wkhtmltopdf)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ACCEPTANCE_TEST_MD="$PROJECT_ROOT/docs/ACCEPTANCE_TEST.md"
OUTPUT_PDF="$PROJECT_ROOT/docs/ACCEPTANCE_TEST.pdf"

echo "Generating PDF from acceptance test document..."

# Check for pandoc
if ! command -v pandoc &> /dev/null; then
    echo "Error: pandoc is not installed."
    echo "Install with:"
    echo "  - macOS: brew install pandoc"
    echo "  - Ubuntu/Debian: sudo apt-get install pandoc"
    echo "  - Or download from: https://pandoc.org/installing.html"
    exit 1
fi

# Generate PDF using pandoc
pandoc "$ACCEPTANCE_TEST_MD" \
    -o "$OUTPUT_PDF" \
    --pdf-engine=xelatex \
    -V geometry:margin=1in \
    -V fontsize=11pt \
    -V documentclass=article \
    --toc \
    --toc-depth=2 \
    -V colorlinks=true \
    -V linkcolor=blue \
    -V urlcolor=blue \
    -V toccolor=gray

if [ $? -eq 0 ]; then
    echo "✅ PDF generated successfully: $OUTPUT_PDF"
    echo ""
    echo "Alternative methods if pandoc fails:"
    echo "1. Use online converter: https://www.markdowntopdf.com/"
    echo "2. Use VS Code extension: 'Markdown PDF'"
    echo "3. Use wkhtmltopdf: wkhtmltopdf $ACCEPTANCE_TEST_MD $OUTPUT_PDF"
else
    echo "❌ PDF generation failed. Try alternative methods:"
    echo "1. Use online converter: https://www.markdowntopdf.com/"
    echo "2. Use VS Code extension: 'Markdown PDF'"
    echo "3. Use wkhtmltopdf: wkhtmltopdf $ACCEPTANCE_TEST_MD $OUTPUT_PDF"
    exit 1
fi
