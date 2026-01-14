# Generate PDF from acceptance test markdown
# Requires: pandoc or alternative PDF generation tool

param(
    [string]$OutputPath = "docs\ACCEPTANCE_TEST.pdf"
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $ScriptDir)
$AcceptanceTestMd = Join-Path $ProjectRoot "docs\ACCEPTANCE_TEST.md"
$OutputPdf = Join-Path $ProjectRoot $OutputPath

Write-Host "Generating PDF from acceptance test document..." -ForegroundColor Cyan

# Check for pandoc
$pandocPath = Get-Command pandoc -ErrorAction SilentlyContinue

if ($pandocPath) {
    Write-Host "Using pandoc to generate PDF..." -ForegroundColor Green
    
    # Generate PDF using pandoc
    pandoc $AcceptanceTestMd `
        -o $OutputPdf `
        --pdf-engine=xelatex `
        -V geometry:margin=1in `
        -V fontsize=11pt `
        -V documentclass=article `
        --toc `
        --toc-depth=2 `
        -V colorlinks=true `
        -V linkcolor=blue `
        -V urlcolor=blue `
        -V toccolor=gray
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ PDF generated successfully: $OutputPdf" -ForegroundColor Green
        exit 0
    } else {
        Write-Host "❌ pandoc failed. Trying alternative methods..." -ForegroundColor Yellow
    }
} else {
    Write-Host "⚠️  pandoc not found. Trying alternative methods..." -ForegroundColor Yellow
}

# Alternative: Check for wkhtmltopdf
$wkhtmltopdfPath = Get-Command wkhtmltopdf -ErrorAction SilentlyContinue

if ($wkhtmltopdfPath) {
    Write-Host "Using wkhtmltopdf to generate PDF..." -ForegroundColor Green
    wkhtmltopdf $AcceptanceTestMd $OutputPdf
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ PDF generated successfully: $OutputPdf" -ForegroundColor Green
        exit 0
    }
}

# If all methods fail, provide instructions
Write-Host ""
Write-Host "❌ Could not generate PDF automatically." -ForegroundColor Red
Write-Host ""
Write-Host "Alternative methods:" -ForegroundColor Yellow
Write-Host "1. Install pandoc:"
Write-Host "   - Download from: https://pandoc.org/installing.html"
Write-Host "   - Or use Chocolatey: choco install pandoc"
Write-Host ""
Write-Host "2. Use online converter:"
Write-Host "   - https://www.markdowntopdf.com/"
Write-Host "   - Upload: $AcceptanceTestMd"
Write-Host ""
Write-Host "3. Use VS Code extension:"
Write-Host "   - Install 'Markdown PDF' extension"
Write-Host "   - Open $AcceptanceTestMd"
Write-Host "   - Right-click > 'Markdown PDF: Export (pdf)'"
Write-Host ""
Write-Host "4. Use PowerShell with Word (if Word is installed):"
Write-Host "   - Open Word"
Write-Host "   - File > Open > $AcceptanceTestMd"
Write-Host "   - File > Save As > PDF"
Write-Host ""

exit 1
