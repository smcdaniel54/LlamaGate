# LlamaGate MCP Demo Workflow Script (PowerShell version)
#
# This script demonstrates LlamaGate's MCP client capabilities using PowerShell.
# It's a Windows-friendly alternative to the bash script.

param(
    [string]$LlamaGateUrl = $env:LLAMAGATE_URL,
    [string]$ApiKey = $env:LLAMAGATE_API_KEY,
    [string]$Model = $env:MODEL,
    [string]$WorkspaceDir = $env:WORKSPACE_DIR
)

# Set defaults
if ([string]::IsNullOrEmpty($LlamaGateUrl)) { $LlamaGateUrl = "http://localhost:11435/v1" }
if ([string]::IsNullOrEmpty($ApiKey)) { $ApiKey = "sk-llamagate" }
if ([string]::IsNullOrEmpty($Model)) { $Model = "llama3.2" }
if ([string]::IsNullOrEmpty($WorkspaceDir)) { $WorkspaceDir = "$env:USERPROFILE\llamagate-workspace" }

function Print-Section {
    param([string]$Title)
    Write-Host ""
    Write-Host "======================================================================" -ForegroundColor Cyan
    Write-Host "  $Title" -ForegroundColor Cyan
    Write-Host "======================================================================" -ForegroundColor Cyan
}

function Print-Step {
    param([int]$StepNum, [string]$Description)
    Write-Host ""
    Write-Host "[Step $StepNum] $Description" -ForegroundColor Yellow
    Write-Host "----------------------------------------------------------------------" -ForegroundColor Gray
}

function Test-LlamaGateConnection {
    Print-Section "Checking LlamaGate Connection"
    
    try {
        $response = Invoke-RestMethod -Uri "$LlamaGateUrl/v1/models" `
            -Method Get `
            -Headers @{ "X-API-Key" = $ApiKey } `
            -ErrorAction Stop
        
        Write-Host "✅ LlamaGate is running and accessible" -ForegroundColor Green
        return $true
    }
    catch {
        Write-Host "❌ Failed to connect to LlamaGate: $_" -ForegroundColor Red
        Write-Host "   Make sure LlamaGate is running on $LlamaGateUrl" -ForegroundColor Yellow
        return $false
    }
}

function Get-AvailableTools {
    Print-Section "Discovering Available Tools"
    
    Print-Step 1 "Querying available MCP tools"
    
    $body = @{
        model = $Model
        messages = @(
            @{
                role = "user"
                content = "What MCP tools are available? List all tools with their namespaces."
            }
        )
        temperature = 0.3
    } | ConvertTo-Json -Depth 10
    
    try {
        $response = Invoke-RestMethod -Uri "$LlamaGateUrl/v1/chat/completions" `
            -Method Post `
            -Headers @{
                "Content-Type" = "application/json"
                "X-API-Key" = $ApiKey
            } `
            -Body $body `
            -ErrorAction Stop
        
        Write-Host $response.choices[0].message.content
    }
    catch {
        Write-Host "Failed to query tools: $_" -ForegroundColor Red
    }
}

function Start-WorkflowReadPdf {
    Print-Section "Workflow 1: Read and Summarize PDF"
    
    # Check workspace
    if (-not (Test-Path $WorkspaceDir)) {
        Write-Host "⚠️  Workspace directory not found: $WorkspaceDir" -ForegroundColor Yellow
        Write-Host "   Creating workspace directory..." -ForegroundColor Yellow
        New-Item -ItemType Directory -Path $WorkspaceDir -Force | Out-Null
    }
    
    # Find PDF files
    $pdfFiles = Get-ChildItem -Path $WorkspaceDir -Filter "*.pdf" -ErrorAction SilentlyContinue | Select-Object -First 1
    
    if ($null -eq $pdfFiles) {
        Write-Host "⚠️  No PDF files found in $WorkspaceDir" -ForegroundColor Yellow
        Write-Host "   Please add a PDF file to test this workflow" -ForegroundColor Yellow
        return $false
    }
    
    Write-Host "📄 Found PDF: $($pdfFiles.Name)" -ForegroundColor Cyan
    
    Print-Step 1 "Reading and summarizing PDF: $($pdfFiles.Name)"
    
    $body = @{
        model = $Model
        messages = @(
            @{
                role = "user"
                content = "Read the PDF file at $($pdfFiles.FullName) and provide a brief summary of its contents. Include the title, main topics, and key points."
            }
        )
        temperature = 0.7
        max_tokens = 1000
    } | ConvertTo-Json -Depth 10
    
    try {
        $response = Invoke-RestMethod -Uri "$LlamaGateUrl/v1/chat/completions" `
            -Method Post `
            -Headers @{
                "Content-Type" = "application/json"
                "X-API-Key" = $ApiKey
            } `
            -Body $body `
            -ErrorAction Stop
        
        Write-Host ""
        Write-Host "📝 Summary:" -ForegroundColor Cyan
        Write-Host $response.choices[0].message.content
        
        Write-Host ""
        Write-Host "✅ PDF processing completed" -ForegroundColor Green
        return $true
    }
    catch {
        Write-Host "❌ Failed to process PDF: $_" -ForegroundColor Red
        return $false
    }
}

function Start-WorkflowMultiStep {
    Print-Section "Workflow 2: Multi-Step Document Processing"
    
    # Create workspace if needed
    if (-not (Test-Path $WorkspaceDir)) {
        New-Item -ItemType Directory -Path $WorkspaceDir -Force | Out-Null
    }
    
    # Create sample file
    $sampleFile = Join-Path $WorkspaceDir "sample.txt"
    if (-not (Test-Path $sampleFile)) {
        Write-Host "📝 Creating sample file: $sampleFile" -ForegroundColor Cyan
        @"
Project Report: LlamaGate MCP Integration

Overview:
This project demonstrates the integration of Model Context Protocol (MCP) servers
with LlamaGate, enabling AI models to interact with external tools and data sources.

Key Features:
- PDF document processing
- Multi-format document handling
- File system operations
- Document conversion

Conclusion:
The integration successfully enables complex document processing workflows through
a unified interface.
"@ | Out-File -FilePath $sampleFile -Encoding UTF8
    }
    
    Print-Step 1 "Processing document through multiple steps"
    
    $body = @{
        model = $Model
        messages = @(
            @{
                role = "system"
                content = "You are a document processing assistant. Use available tools to process documents."
            }
            @{
                role = "user"
                content = "Process the file $sampleFile`: 1. Read the file content 2. Extract the main sections (Overview, Key Features, Conclusion) 3. Create a structured summary 4. Save the summary to $(Join-Path $WorkspaceDir 'summary.txt') 5. List all files in the workspace to confirm the file was created"
            }
        )
        temperature = 0.7
        max_tokens = 2000
    } | ConvertTo-Json -Depth 10
    
    try {
        $response = Invoke-RestMethod -Uri "$LlamaGateUrl/v1/chat/completions" `
            -Method Post `
            -Headers @{
                "Content-Type" = "application/json"
                "X-API-Key" = $ApiKey
            } `
            -Body $body `
            -ErrorAction Stop
        
        Write-Host ""
        Write-Host "📝 Processing Result:" -ForegroundColor Cyan
        Write-Host $response.choices[0].message.content
        
        # Check if summary was created
        $summaryFile = Join-Path $WorkspaceDir "summary.txt"
        if (Test-Path $summaryFile) {
            $fileSize = (Get-Item $summaryFile).Length
            Write-Host ""
            Write-Host "✅ Summary file created: $summaryFile" -ForegroundColor Green
            Write-Host "   Size: $fileSize bytes" -ForegroundColor Gray
        }
        else {
            Write-Host ""
            Write-Host "⚠️  Summary file not found" -ForegroundColor Yellow
        }
        
        return $true
    }
    catch {
        Write-Host "❌ Failed to process document: $_" -ForegroundColor Red
        return $false
    }
}

function Start-WorkflowListFiles {
    Print-Section "Workflow 3: List and Process Multiple Documents"
    
    if (-not (Test-Path $WorkspaceDir)) {
        New-Item -ItemType Directory -Path $WorkspaceDir -Force | Out-Null
    }
    
    Print-Step 1 "Listing and processing all documents in workspace"
    
    $body = @{
        model = $Model
        messages = @(
            @{
                role = "user"
                content = "List all files in the directory $WorkspaceDir, then for each text or markdown file, read it and create a brief description. Present the results as a list of files with their descriptions."
            }
        )
        temperature = 0.7
        max_tokens = 2000
    } | ConvertTo-Json -Depth 10
    
    try {
        $response = Invoke-RestMethod -Uri "$LlamaGateUrl/v1/chat/completions" `
            -Method Post `
            -Headers @{
                "Content-Type" = "application/json"
                "X-API-Key" = $ApiKey
            } `
            -Body $body `
            -ErrorAction Stop
        
        Write-Host ""
        Write-Host "📋 File Listing and Descriptions:" -ForegroundColor Cyan
        Write-Host $response.choices[0].message.content
        
        Write-Host ""
        Write-Host "✅ File listing completed" -ForegroundColor Green
        return $true
    }
    catch {
        Write-Host "❌ Failed to list files: $_" -ForegroundColor Red
        return $false
    }
}

# Main execution
Print-Section "LlamaGate MCP Demo Workflow"
Write-Host "LlamaGate URL: $LlamaGateUrl"
Write-Host "Model: $Model"
Write-Host "Workspace: $WorkspaceDir"

# Check connection
if (-not (Test-LlamaGateConnection)) {
    exit 1
}

# Discover tools
Get-AvailableTools

# Run workflows
Write-Host ""
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host "  Running Demo Workflows" -ForegroundColor Cyan
Write-Host "======================================================================" -ForegroundColor Cyan

$results = @()
$results += @{ Name = "Workflow 1: Read PDF"; Success = (Start-WorkflowReadPdf) }
$results += @{ Name = "Workflow 2: Multi-Step Processing"; Success = (Start-WorkflowMultiStep) }
$results += @{ Name = "Workflow 3: List and Process"; Success = (Start-WorkflowListFiles) }

# Summary
Print-Section "Workflow Summary"
foreach ($result in $results) {
    if ($result.Success) {
        Write-Host "✅ PASSED: $($result.Name)" -ForegroundColor Green
    }
    else {
        Write-Host "❌ FAILED: $($result.Name)" -ForegroundColor Red
    }
}

$passed = ($results | Where-Object { $_.Success }).Count
$total = $results.Count

Write-Host ""
Write-Host "Total: $passed/$total workflows passed" -ForegroundColor Cyan

if ($passed -eq $total) {
    Write-Host ""
    Write-Host "🎉 All workflows completed successfully!" -ForegroundColor Green
}
else {
    Write-Host ""
    Write-Host "⚠️  $($total - $passed) workflow(s) failed. Check the output above for details." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Check the workspace directory for generated files:" -ForegroundColor Cyan
Write-Host "  $WorkspaceDir" -ForegroundColor Gray

