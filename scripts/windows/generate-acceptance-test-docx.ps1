# Generate DOCX from acceptance test markdown using Microsoft Word
# Requires: Microsoft Word installed

param(
    [string]$OutputPath = "docs\ACCEPTANCE_TEST.docx"
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $ScriptDir)
$AcceptanceTestMd = Join-Path $ProjectRoot "docs\ACCEPTANCE_TEST.md"
$OutputDocx = Join-Path $ProjectRoot $OutputPath

Write-Host "Generating DOCX from acceptance test document..." -ForegroundColor Cyan
Write-Host "Source: $AcceptanceTestMd" -ForegroundColor Gray
Write-Host "Output: $OutputDocx" -ForegroundColor Gray

# Check if source file exists
if (-not (Test-Path $AcceptanceTestMd)) {
    Write-Host "❌ Source file not found: $AcceptanceTestMd" -ForegroundColor Red
    exit 1
}

# Check for Word
try {
    $word = New-Object -ComObject Word.Application
    $word.Visible = $false
    $word.DisplayAlerts = $false
    
    Write-Host "Opening Markdown file in Word..." -ForegroundColor Green
    
    # Word can open markdown files directly
    $openParams = @{
        FileName = $AcceptanceTestMd
        ConfirmConversions = $false
        ReadOnly = $true
        AddToRecentFiles = $false
    }
    $doc = $word.Documents.Open(
        $openParams.FileName,
        [ref]$openParams.ConfirmConversions,
        [ref]$openParams.ReadOnly,
        [ref]$openParams.AddToRecentFiles
    )
    
    Write-Host "Converting to DOCX format..." -ForegroundColor Green
    
    # Save as DOCX (format 16 = wdFormatDocumentDefault which is DOCX)
    $wdFormatDocumentDefault = 16
    $saveParams = @{
        FileName = $OutputDocx
        FileFormat = $wdFormatDocumentDefault
    }
    $doc.SaveAs(
        [ref]$saveParams.FileName,
        [ref]$saveParams.FileFormat
    )
    
    # Close document and Word
    $doc.Close($false)
    $word.Quit()
    
    # Release COM objects
    [System.Runtime.Interopservices.Marshal]::ReleaseComObject($doc) | Out-Null
    [System.Runtime.Interopservices.Marshal]::ReleaseComObject($word) | Out-Null
    [System.GC]::Collect()
    [System.GC]::WaitForPendingFinalizers()
    
    if (Test-Path $OutputDocx) {
        Write-Host "✅ DOCX generated successfully: $OutputDocx" -ForegroundColor Green
        Write-Host ""
        Write-Host "File size: $((Get-Item $OutputDocx).Length / 1KB) KB" -ForegroundColor Gray
        exit 0
    } else {
        Write-Host "❌ DOCX file was not created" -ForegroundColor Red
        exit 1
    }
}
catch {
    Write-Host "❌ Error generating DOCX: $_" -ForegroundColor Red
    Write-Host ""
    Write-Host "Alternative methods:" -ForegroundColor Yellow
    Write-Host "1. Install pandoc and use: pandoc $AcceptanceTestMd -o $OutputDocx" -ForegroundColor Yellow
    Write-Host "2. Use online converter: https://cloudconvert.com/md-to-docx" -ForegroundColor Yellow
    Write-Host "3. Manually open in Word: File > Open > $AcceptanceTestMd > Save As > DOCX" -ForegroundColor Yellow
    exit 1
}
