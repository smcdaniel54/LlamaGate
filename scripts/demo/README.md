# MCP Demo Scripts

This directory contains demonstration scripts for running LlamaGate with multiple MCP servers.

## Quick Start

1. **Configure LlamaGate** - Copy `mcp-demo-config.yaml` to `llamagate.yaml`:
   ```bash
   cp ../../mcp-demo-config.yaml llamagate.yaml
   # Or place in ~/.llamagate/ or %USERPROFILE%\.llamagate\
   ```

2. **Start LlamaGate** with MCP enabled (see [MCP Demo QuickStart](../../docs/MCP_DEMO_QUICKSTART.md))

3. **Run a demo script**:
   ```bash
   # Python (recommended)
   python mcp-demo-workflow.py
   
   # Bash (Unix/Linux/macOS)
   ./mcp-demo-workflow.sh
   
   # PowerShell (Windows)
   .\mcp-demo-workflow.ps1
   ```

## Scripts

### `mcp-demo-workflow.py`

Python script with comprehensive workflow demonstrations:
- ✅ Connection verification
- ✅ Tool discovery
- ✅ PDF reading and summarization
- ✅ Multi-step document processing
- ✅ File listing and processing
- ✅ Document conversion

**Requirements:**
```bash
pip install openai
```

**Usage:**
```bash
python mcp-demo-workflow.py
```

**Environment Variables:**
- `LLAMAGATE_URL` - LlamaGate API URL (default: `http://localhost:11435/v1`)
- `LLAMAGATE_API_KEY` - API key (default: `sk-llamagate`)
- `MODEL` - Model name (default: `mistral`)
- `WORKSPACE_DIR` - Working directory (default: `~/llamagate-workspace`)

### `mcp-demo-workflow.sh`

Bash script for Unix/Linux/macOS systems. Simpler alternative using `curl` and `jq`.

**Requirements:**
- `curl`
- `jq` (recommended for better output)

**Usage:**
```bash
chmod +x mcp-demo-workflow.sh
./mcp-demo-workflow.sh
```

### `mcp-demo-workflow.ps1`

PowerShell script for Windows systems. Windows-friendly alternative.

**Usage:**
```powershell
.\mcp-demo-workflow.ps1
```

**Parameters:**
- `-LlamaGateUrl` - LlamaGate API URL
- `-ApiKey` - API key
- `-Model` - Model name
- `-WorkspaceDir` - Working directory

## Workflows Demonstrated

1. **Read and Summarize PDF** - Uses SylphxAI PDF Reader to read and summarize PDF documents
2. **Multi-Step Processing** - Processes documents through multiple steps (read → extract → summarize → save)
3. **List and Process** - Lists all files in workspace and processes each one
4. **Document Conversion** - Converts documents between formats (if supported)

## Configuration

The demo scripts use the MCP servers configured in `mcp-demo-config.yaml`:

- **SylphxAI PDF Reader** - PDF document processing
- **AWS Document Loader** - Multi-format document handling
- **PulseMCP** - Document editing and conversions
- **Mikado Filesystem** - File system operations
- **MCPZoo** (optional) - Collection of MCP servers

See [MCP Demo QuickStart](../../docs/MCP_DEMO_QUICKSTART.md) for detailed setup instructions.

## Troubleshooting

### Script fails to connect

- Ensure LlamaGate is running: `curl http://localhost:11435/health`
- Check the `LLAMAGATE_URL` environment variable
- Verify the API key matches your configuration

### Tools not available

- Check LlamaGate logs for MCP server initialization
- Verify MCP servers are installed and accessible
- Check that servers are enabled in `llamagate.yaml`

### Workflow fails

- Ensure the workspace directory exists
- Check file permissions
- Verify required files (PDFs, text files) are in the workspace
- Review LlamaGate logs for tool execution errors

## Next Steps

- Customize workflows for your use case
- Add more MCP servers to the configuration
- Create your own workflow scripts
- See [MCP.md](../../docs/MCP.md) for full MCP documentation

