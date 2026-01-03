#!/bin/bash
# LlamaGate MCP Demo Workflow Script (Bash version)
#
# This script demonstrates LlamaGate's MCP client capabilities using curl.
# It's a simpler alternative to the Python script for quick testing.

set -euo pipefail

# Configuration
LLAMAGATE_URL="${LLAMAGATE_URL:-http://localhost:8080/v1}"
LLAMAGATE_API_KEY="${LLAMAGATE_API_KEY:-sk-llamagate}"
MODEL="${MODEL:-llama3.2}"
WORKSPACE_DIR="${WORKSPACE_DIR:-$HOME/llamagate-workspace}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_section() {
    echo ""
    echo "======================================================================"
    echo "  $1"
    echo "======================================================================"
}

print_step() {
    echo ""
    echo "[Step $1] $2"
    echo "----------------------------------------------------------------------"
}

check_llamagate() {
    print_section "Checking LlamaGate Connection"
    
    if curl -s -f "${LLAMAGATE_URL}/v1/models" \
        -H "X-API-Key: ${LLAMAGATE_API_KEY}" > /dev/null; then
        echo -e "${GREEN}✅ LlamaGate is running and accessible${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed to connect to LlamaGate${NC}"
        echo "   Make sure LlamaGate is running on ${LLAMAGATE_URL}"
        return 1
    fi
}

list_tools() {
    print_section "Discovering Available Tools"
    
    print_step 1 "Querying available MCP tools"
    
    curl -s "${LLAMAGATE_URL}/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: ${LLAMAGATE_API_KEY}" \
        -d "{
            \"model\": \"${MODEL}\",
            \"messages\": [
                {
                    \"role\": \"user\",
                    \"content\": \"What MCP tools are available? List all tools with their namespaces.\"
                }
            ],
            \"temperature\": 0.3
        }" | jq -r '.choices[0].message.content' || echo "Failed to query tools"
}

workflow_read_pdf() {
    print_section "Workflow 1: Read and Summarize PDF"
    
    # Check workspace
    if [ ! -d "${WORKSPACE_DIR}" ]; then
        echo -e "${YELLOW}⚠️  Workspace directory not found: ${WORKSPACE_DIR}${NC}"
        echo "   Creating workspace directory..."
        mkdir -p "${WORKSPACE_DIR}"
    fi
    
    # Find PDF files
    PDF_FILES=$(find "${WORKSPACE_DIR}" -maxdepth 1 -name "*.pdf" 2>/dev/null | head -1)
    
    if [ -z "${PDF_FILES}" ]; then
        echo -e "${YELLOW}⚠️  No PDF files found in ${WORKSPACE_DIR}${NC}"
        echo "   Please add a PDF file to test this workflow"
        return 1
    fi
    
    PDF_NAME=$(basename "${PDF_FILES}")
    echo "📄 Found PDF: ${PDF_NAME}"
    
    print_step 1 "Reading and summarizing PDF: ${PDF_NAME}"
    
    curl -s "${LLAMAGATE_URL}/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: ${LLAMAGATE_API_KEY}" \
        -d "{
            \"model\": \"${MODEL}\",
            \"messages\": [
                {
                    \"role\": \"user\",
                    \"content\": \"Read the PDF file at ${PDF_FILES} and provide a brief summary of its contents. Include the title, main topics, and key points.\"
                }
            ],
            \"temperature\": 0.7,
            \"max_tokens\": 1000
        }" | jq -r '.choices[0].message.content'
    
    echo ""
    echo -e "${GREEN}✅ PDF processing completed${NC}"
}

workflow_multi_step() {
    print_section "Workflow 2: Multi-Step Document Processing"
    
    # Create workspace if needed
    mkdir -p "${WORKSPACE_DIR}"
    
    # Create sample file
    SAMPLE_FILE="${WORKSPACE_DIR}/sample.txt"
    if [ ! -f "${SAMPLE_FILE}" ]; then
        echo "📝 Creating sample file: ${SAMPLE_FILE}"
        cat > "${SAMPLE_FILE}" << 'EOF'
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
EOF
    fi
    
    print_step 1 "Processing document through multiple steps"
    
    curl -s "${LLAMAGATE_URL}/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: ${LLAMAGATE_API_KEY}" \
        -d "{
            \"model\": \"${MODEL}\",
            \"messages\": [
                {
                    \"role\": \"system\",
                    \"content\": \"You are a document processing assistant. Use available tools to process documents.\"
                },
                {
                    \"role\": \"user\",
                    \"content\": \"Process the file ${SAMPLE_FILE}: 1. Read the file content 2. Extract the main sections (Overview, Key Features, Conclusion) 3. Create a structured summary 4. Save the summary to ${WORKSPACE_DIR}/summary.txt 5. List all files in the workspace to confirm the file was created\"
                }
            ],
            \"temperature\": 0.7,
            \"max_tokens\": 2000
        }" | jq -r '.choices[0].message.content'
    
    # Check if summary was created
    if [ -f "${WORKSPACE_DIR}/summary.txt" ]; then
        echo ""
        echo -e "${GREEN}✅ Summary file created: ${WORKSPACE_DIR}/summary.txt${NC}"
        echo "   Size: $(stat -f%z "${WORKSPACE_DIR}/summary.txt" 2>/dev/null || stat -c%s "${WORKSPACE_DIR}/summary.txt" 2>/dev/null) bytes"
    else
        echo ""
        echo -e "${YELLOW}⚠️  Summary file not found${NC}"
    fi
}

workflow_list_files() {
    print_section "Workflow 3: List and Process Multiple Documents"
    
    mkdir -p "${WORKSPACE_DIR}"
    
    print_step 1 "Listing and processing all documents in workspace"
    
    curl -s "${LLAMAGATE_URL}/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: ${LLAMAGATE_API_KEY}" \
        -d "{
            \"model\": \"${MODEL}\",
            \"messages\": [
                {
                    \"role\": \"user\",
                    \"content\": \"List all files in the directory ${WORKSPACE_DIR}, then for each text or markdown file, read it and create a brief description. Present the results as a list of files with their descriptions.\"
                }
            ],
            \"temperature\": 0.7,
            \"max_tokens\": 2000
        }" | jq -r '.choices[0].message.content'
    
    echo ""
    echo -e "${GREEN}✅ File listing completed${NC}"
}

main() {
    print_section "LlamaGate MCP Demo Workflow"
    echo "LlamaGate URL: ${LLAMAGATE_URL}"
    echo "Model: ${MODEL}"
    echo "Workspace: ${WORKSPACE_DIR}"
    
    # Check prerequisites
    if ! command -v curl &> /dev/null; then
        echo -e "${RED}❌ curl is required but not installed${NC}"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        echo -e "${YELLOW}⚠️  jq is recommended for better output formatting${NC}"
        echo "   Install with: brew install jq (macOS) or apt-get install jq (Linux)"
    fi
    
    # Check connection
    if ! check_llamagate; then
        exit 1
    fi
    
    # Discover tools
    list_tools
    
    # Run workflows
    echo ""
    echo "======================================================================"
    echo "  Running Demo Workflows"
    echo "======================================================================"
    
    workflow_read_pdf || echo -e "${YELLOW}⚠️  Workflow 1 skipped${NC}"
    workflow_multi_step || echo -e "${YELLOW}⚠️  Workflow 2 skipped${NC}"
    workflow_list_files || echo -e "${YELLOW}⚠️  Workflow 3 skipped${NC}"
    
    print_section "Demo Complete"
    echo -e "${GREEN}✅ All workflows completed!${NC}"
    echo ""
    echo "Check the workspace directory for generated files:"
    echo "  ${WORKSPACE_DIR}"
}

main "$@"

