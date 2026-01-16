#!/usr/bin/env python3
"""
LlamaGate MCP Demo Workflow Script

This script demonstrates LlamaGate's MCP client capabilities by running
a complete document processing workflow using multiple MCP servers.

Prerequisites:
- LlamaGate running with MCP enabled (see docs/MCP_DEMO_QUICKSTART.md)
- OpenAI Python client: pip install openai
- Sample documents in the workspace directory
"""

import os
import sys
import json
from pathlib import Path
from openai import OpenAI

# Configuration
LLAMAGATE_URL = os.getenv("LLAMAGATE_URL", "http://localhost:11435/v1")
LLAMAGATE_API_KEY = os.getenv("LLAMAGATE_API_KEY", "sk-llamagate")
MODEL = os.getenv("MODEL", "mistral")
WORKSPACE_DIR = os.getenv("WORKSPACE_DIR", os.path.expanduser("~/llamagate-workspace"))

# Initialize OpenAI client (pointing to LlamaGate)
client = OpenAI(
    base_url=LLAMAGATE_URL,
    api_key=LLAMAGATE_API_KEY
)


def print_section(title):
    """Print a formatted section header."""
    print("\n" + "=" * 70)
    print(f"  {title}")
    print("=" * 70)


def print_step(step_num, description):
    """Print a formatted step."""
    print(f"\n[Step {step_num}] {description}")
    print("-" * 70)


def check_llamagate_connection():
    """Verify LlamaGate is running and accessible."""
    print_section("Checking LlamaGate Connection")
    try:
        response = client.chat.completions.create(
            model=MODEL,
            messages=[{"role": "user", "content": "Hello"}],
            max_tokens=10
        )
        print("✅ LlamaGate is running and accessible")
        return True
    except Exception as e:
        print(f"❌ Failed to connect to LlamaGate: {e}")
        print(f"   Make sure LlamaGate is running on {LLAMAGATE_URL}")
        return False


def list_available_tools():
    """Query the model about available tools."""
    print_section("Discovering Available Tools")
    try:
        response = client.chat.completions.create(
            model=MODEL,
            messages=[
                {
                    "role": "user",
                    "content": "What MCP tools are available? List all tools with their namespaces."
                }
            ],
            temperature=0.3
        )
        print(response.choices[0].message.content)
        return True
    except Exception as e:
        print(f"❌ Failed to discover tools: {e}")
        return False


def workflow_1_read_pdf():
    """Workflow 1: Read and summarize a PDF file."""
    print_section("Workflow 1: Read and Summarize PDF")
    
    # Check if workspace exists
    workspace = Path(WORKSPACE_DIR)
    if not workspace.exists():
        print(f"⚠️  Workspace directory not found: {WORKSPACE_DIR}")
        print("   Creating workspace directory...")
        workspace.mkdir(parents=True, exist_ok=True)
    
    # Look for PDF files
    pdf_files = list(workspace.glob("*.pdf"))
    if not pdf_files:
        print(f"⚠️  No PDF files found in {WORKSPACE_DIR}")
        print("   Please add a PDF file to test this workflow")
        return False
    
    pdf_path = pdf_files[0]
    print(f"📄 Found PDF: {pdf_path.name}")
    
    print_step(1, f"Reading PDF file: {pdf_path.name}")
    try:
        response = client.chat.completions.create(
            model=MODEL,
            messages=[
                {
                    "role": "user",
                    "content": f"Read the PDF file at {pdf_path} and provide a brief summary of its contents. Include the title, main topics, and key points."
                }
            ],
            temperature=0.7,
            max_tokens=1000
        )
        
        print("\n📝 Summary:")
        print(response.choices[0].message.content)
        
        # Check if tools were used
        if hasattr(response.choices[0].message, 'tool_calls') and response.choices[0].message.tool_calls:
            print("\n🔧 Tools used:")
            for tool_call in response.choices[0].message.tool_calls:
                print(f"   - {tool_call.function.name}")
        
        return True
    except Exception as e:
        print(f"❌ Failed to process PDF: {e}")
        return False


def workflow_2_multi_step_processing():
    """Workflow 2: Multi-step document processing."""
    print_section("Workflow 2: Multi-Step Document Processing")
    
    workspace = Path(WORKSPACE_DIR)
    if not workspace.exists():
        workspace.mkdir(parents=True, exist_ok=True)
    
    # Create a sample text file if it doesn't exist
    sample_file = workspace / "sample.txt"
    if not sample_file.exists():
        print(f"📝 Creating sample file: {sample_file}")
        sample_file.write_text("""
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
""")
    
    print_step(1, "Processing document through multiple steps")
    try:
        response = client.chat.completions.create(
            model=MODEL,
            messages=[
                {
                    "role": "system",
                    "content": "You are a document processing assistant. Use available tools to process documents."
                },
                {
                    "role": "user",
                    "content": f"""Process the file {sample_file}:
1. Read the file content
2. Extract the main sections (Overview, Key Features, Conclusion)
3. Create a structured summary
4. Save the summary to {workspace / 'summary.txt'}
5. List all files in the workspace to confirm the file was created"""
                }
            ],
            temperature=0.7,
            max_tokens=2000
        )
        
        print("\n📝 Processing Result:")
        print(response.choices[0].message.content)
        
        # Verify the summary file was created
        summary_file = workspace / "summary.txt"
        if summary_file.exists():
            print(f"\n✅ Summary file created: {summary_file}")
            print(f"   Size: {summary_file.stat().st_size} bytes")
        else:
            print(f"\n⚠️  Summary file not found: {summary_file}")
        
        return True
    except Exception as e:
        print(f"❌ Failed to process document: {e}")
        import traceback
        traceback.print_exc()
        return False


def workflow_3_list_and_process():
    """Workflow 3: List and process multiple documents."""
    print_section("Workflow 3: List and Process Multiple Documents")
    
    workspace = Path(WORKSPACE_DIR)
    if not workspace.exists():
        workspace.mkdir(parents=True, exist_ok=True)
    
    print_step(1, "Listing and processing all documents in workspace")
    try:
        response = client.chat.completions.create(
            model=MODEL,
            messages=[
                {
                    "role": "user",
                    "content": f"""List all files in the directory {workspace}, 
then for each text or markdown file, read it and create a brief description.
Present the results as a list of files with their descriptions."""
                }
            ],
            temperature=0.7,
            max_tokens=2000
        )
        
        print("\n📋 File Listing and Descriptions:")
        print(response.choices[0].message.content)
        
        return True
    except Exception as e:
        print(f"❌ Failed to list and process files: {e}")
        return False


def workflow_4_document_conversion():
    """Workflow 4: Document conversion (if supported)."""
    print_section("Workflow 4: Document Conversion")
    
    workspace = Path(WORKSPACE_DIR)
    if not workspace.exists():
        workspace.mkdir(parents=True, exist_ok=True)
    
    # Look for text files to "convert"
    text_files = list(workspace.glob("*.txt"))
    if not text_files:
        print("⚠️  No text files found for conversion")
        print("   This workflow requires a source document")
        return False
    
    source_file = text_files[0]
    target_file = workspace / f"{source_file.stem}_converted.md"
    
    print_step(1, f"Converting {source_file.name} to Markdown format")
    try:
        response = client.chat.completions.create(
            model=MODEL,
            messages=[
                {
                    "role": "user",
                    "content": f"""Read the file {source_file} and convert it to Markdown format.
Save the converted content to {target_file}.
Use proper Markdown formatting with headers, lists, and emphasis."""
                }
            ],
            temperature=0.7,
            max_tokens=2000
        )
        
        print("\n📝 Conversion Result:")
        print(response.choices[0].message.content)
        
        if target_file.exists():
            print(f"\n✅ Converted file created: {target_file}")
        else:
            print(f"\n⚠️  Converted file not found: {target_file}")
        
        return True
    except Exception as e:
        print(f"❌ Failed to convert document: {e}")
        return False


def main():
    """Run all demo workflows."""
    print_section("LlamaGate MCP Demo Workflow")
    print(f"LlamaGate URL: {LLAMAGATE_URL}")
    print(f"Model: {MODEL}")
    print(f"Workspace: {WORKSPACE_DIR}")
    
    # Check connection
    if not check_llamagate_connection():
        sys.exit(1)
    
    # Discover tools
    if not list_available_tools():
        print("⚠️  Continuing anyway...")
    
    # Run workflows
    results = []
    
    print("\n" + "=" * 70)
    print("  Running Demo Workflows")
    print("=" * 70)
    
    results.append(("Workflow 1: Read PDF", workflow_1_read_pdf()))
    results.append(("Workflow 2: Multi-Step Processing", workflow_2_multi_step_processing()))
    results.append(("Workflow 3: List and Process", workflow_3_list_and_process()))
    results.append(("Workflow 4: Document Conversion", workflow_4_document_conversion()))
    
    # Summary
    print_section("Workflow Summary")
    for name, success in results:
        status = "✅ PASSED" if success else "❌ FAILED"
        print(f"{status}: {name}")
    
    passed = sum(1 for _, success in results if success)
    total = len(results)
    print(f"\nTotal: {passed}/{total} workflows passed")
    
    if passed == total:
        print("\n🎉 All workflows completed successfully!")
        return 0
    else:
        print(f"\n⚠️  {total - passed} workflow(s) failed. Check the output above for details.")
        return 1


if __name__ == "__main__":
    try:
        sys.exit(main())
    except KeyboardInterrupt:
        print("\n\n⚠️  Interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n❌ Unexpected error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)

