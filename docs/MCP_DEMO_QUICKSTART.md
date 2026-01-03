# MCP Demo QuickStart Guide

This guide demonstrates how to run LlamaGate as an MCP client connected to multiple MCP servers. You'll learn how to configure and use:

- **Official Filesystem Server** - File system operations (read, write, list, search)
- **Official Fetch Server** - HTTP request capabilities
- **Official Puppeteer Server** - Web browsing and screenshots (optional)
- **Official Database Servers** - PostgreSQL and SQLite support (optional)
- **Official GitHub Server** - GitHub API access (optional)
- **Official Slack Server** - Slack integration (optional)

## Prerequisites

Before starting, ensure you have:

1. **LlamaGate installed** - See [Quick Start Guide](../QUICKSTART.md)
2. **Node.js installed** (v18+) - Required for MCP servers
   ```bash
   # Check Node.js version
   node --version
   npm --version
   ```
3. **Ollama running** - With a model that supports function/tool calling
   ```bash
   # Recommended models: llama3.2, llama3.1, mistral, qwen2.5
   ollama pull llama3.2
   ```
4. **Working directory** - Create a directory for document processing
   ```bash
   # Unix/Linux/macOS
   mkdir -p ~/llamagate-workspace
   
   # Windows
   mkdir C:\llamagate-workspace
   ```

## Step 1: Find and Install MCP Servers

The MCP ecosystem is rapidly evolving. Here's how to find the exact packages:

### Finding MCP Servers

1. **Search npm registry:**
   ```bash
   npm search mcp-server
   npm search mcp pdf
   npm search mcp document
   ```

2. **Check MCP Server directories:**
   - [MCP Servers Directory](https://github.com/modelcontextprotocol/servers)
   - [Awesome MCP](https://github.com/modelcontextprotocol/awesome-mcp)
   - [MCP Stack](https://www.mcpstack.org/)

3. **Search GitHub:**
   - Search for "mcp-server" repositories
   - Look for packages with "mcp" in the name

### Installing Official MCP Servers

The demo configuration uses official `@modelcontextprotocol` servers that are verified to exist:

```bash
# Option 1: Install globally (recommended for development)
npm install -g @modelcontextprotocol/server-filesystem
npm install -g @modelcontextprotocol/server-fetch
npm install -g @modelcontextprotocol/server-puppeteer  # Optional
npm install -g @modelcontextprotocol/server-postgres   # Optional
npm install -g @modelcontextprotocol/server-sqlite    # Optional
npm install -g @modelcontextprotocol/server-github    # Optional
npm install -g @modelcontextprotocol/server-slack     # Optional

# Option 2: Use npx (no installation needed)
# The configuration file uses npx, so no global install required
```

**Note:** The official servers are well-maintained and documented. For community servers, verify they exist before using.

## Step 2: Configure LlamaGate

1. **Copy the demo configuration:**
   ```bash
   # Copy the demo config to your config location
   cp mcp-demo-config.yaml llamagate.yaml
   
   # Or place it in ~/.llamagate/ (Unix/Linux/macOS)
   mkdir -p ~/.llamagate
   cp mcp-demo-config.yaml ~/.llamagate/llamagate.yaml
   
   # Windows
   mkdir %USERPROFILE%\.llamagate
   copy mcp-demo-config.yaml %USERPROFILE%\.llamagate\llamagate.yaml
   ```

2. **Update the configuration:**
   - Edit `llamagate.yaml` and update package names if they differ
   - Set the allowed directory for Mikado filesystem server
   - Adjust timeouts if needed
   - Enable/disable servers as needed

3. **Verify configuration:**
   ```bash
   # Check that the config file is valid YAML
   # You can use a YAML validator or just start LlamaGate
   ```

## Step 3: Start LlamaGate

Start LlamaGate with MCP enabled:

```bash
# Windows
scripts\windows\run.cmd

# Unix/Linux/macOS
./scripts/unix/run.sh
```

**Expected output:**
```
INFO Starting LlamaGate ollama_host=http://localhost:11434 port=8080
INFO Initializing MCP clients...
INFO MCP client initialized server=filesystem transport=stdio
INFO Discovered tools from MCP server server=filesystem tool_count=5
INFO MCP client initialized server=fetch transport=stdio
INFO Discovered tools from MCP server server=fetch tool_count=2
INFO MCP initialization complete total_tools=7
INFO Starting HTTP server on :8080
```

## Step 4: Verify Tools Are Available

Check that tools are discovered and namespaced correctly:

```bash
# Make a simple request to see available tools
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "model": "llama3.2",
    "messages": [
      {
        "role": "user",
        "content": "What tools are available?"
      }
    ]
  }' | jq '.choices[0].message.content'
```

The model should mention tools like:
- `mcp.filesystem.read_file`
- `mcp.filesystem.write_file`
- `mcp.filesystem.list_directory`
- `mcp.fetch.fetch_url`

## Step 5: Run Sample Workflows

### Workflow 1: Read and Summarize a PDF

**Prerequisites:** Place a PDF file in your workspace directory.

```bash
# Copy a PDF to your workspace
cp sample.pdf ~/llamagate-workspace/  # Unix/Linux/macOS
copy sample.pdf C:\llamagate-workspace\  # Windows
```

**Python Example:**
```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="sk-llamagate"
)

# Read and summarize a text file (filesystem server)
response = client.chat.completions.create(
    model="llama3.2",
    messages=[
        {
            "role": "user",
            "content": "Read the file at ~/llamagate-workspace/sample.txt and provide a summary of its contents"
        }
    ]
)

print(response.choices[0].message.content)
```

**cURL Example:**
```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "model": "llama3.2",
    "messages": [
      {
        "role": "user",
        "content": "Read the file at ~/llamagate-workspace/sample.txt and provide a summary"
      }
    ]
  }'
```

### Workflow 2: Multi-Step Document Processing

Process a document through multiple steps:

```python
response = client.chat.completions.create(
    model="llama3.2",
    messages=[
        {
            "role": "user",
            "content": """Process the document workflow.txt:
            1. Read the file workflow.txt
            2. Extract key information
            3. Create a summary
            4. Save the summary to summary.txt"""
        }
    ]
)

print(response.choices[0].message.content)
```

### Workflow 3: Fetch and Process Web Content

Fetch content from a URL and process it:

```python
response = client.chat.completions.create(
    model="llama3.2",
    messages=[
        {
            "role": "user",
            "content": "Fetch the content from https://example.com/article and create a summary, then save it to article_summary.txt"
        }
    ]
)
```

### Workflow 4: List and Process Multiple Documents

```python
response = client.chat.completions.create(
    model="llama3.2",
    messages=[
        {
            "role": "user",
            "content": """List all text files in the workspace directory, 
            then read each one and create a combined summary"""
        }
    ]
)
```

## Step 6: Monitor Tool Execution

Watch the LlamaGate logs to see tool execution:

```bash
# The logs will show:
INFO Tool executed successfully tool=mcp.filesystem.read_file duration=150ms
INFO Tool execution round completed round=1 tool_calls=1
INFO Tool executed successfully tool=mcp.filesystem.write_file duration=120ms
INFO Tool execution round completed round=2 tool_calls=1
```

## Troubleshooting

### MCP Server Not Found

**Error:** `Failed to initialize MCP client: exec: "npx": executable file not found`

**Solution:**
- Ensure Node.js is installed: `node --version`
- Add Node.js to your PATH
- Or install the MCP server globally: `npm install -g <package-name>`

### Package Not Found

**Error:** `npm ERR! code E404 - Package not found`

**Solution:**
- Verify the package name is correct
- Search npm: `npm search <package-name>`
- Check the MCP server's GitHub repository
- Update the configuration with the correct package name

### Tools Not Discovered

**Error:** No tools appear in logs

**Solution:**
- Check MCP server logs (stderr output)
- Verify the server is compatible with MCP protocol
- Check that the server initializes correctly
- Review LlamaGate logs for connection errors

### Tool Execution Timeout

**Error:** `Tool execution timeout`

**Solution:**
- Increase `default_tool_timeout` in configuration
- Increase per-server `timeout` for slow operations
- Check if the MCP server is responding

### Permission Denied

**Error:** `Permission denied` when accessing files

**Solution:**
- Check file permissions
- Verify the allowed directory in Mikado filesystem config
- Ensure the working directory exists and is accessible

## Advanced Configuration

### Custom Environment Variables

Add environment variables for MCP servers:

```yaml
- name: aws-document
  env:
    NODE_ENV: production
    AWS_ACCESS_KEY_ID: your-key-id
    AWS_SECRET_ACCESS_KEY: your-secret-key
    AWS_REGION: us-east-1
```

### Security: Restrict Tool Access

Use allow/deny lists to restrict tool access:

```yaml
allow_tools:
  - "mcp.sylphxai.read_*"  # Only read operations
  - "mcp.mikado.read_*"
  
deny_tools:
  - "*.delete_*"  # Deny all delete operations
  - "*.write_*"   # Deny all write operations (if needed)
```

### Performance Tuning

Adjust limits for document processing:

```yaml
max_tool_rounds: 20  # More rounds for complex workflows
max_tool_calls_per_round: 5  # Fewer calls per round for large documents
max_tool_result_size: 10485760  # 10MB for large documents
```

## Available Official MCP Servers

The demo configuration includes these official servers (some disabled by default):

1. **Filesystem** - File operations (enabled)
2. **Fetch** - HTTP requests (enabled)
3. **Puppeteer** - Web browsing (disabled, requires Chrome)
4. **PostgreSQL** - Database queries (disabled, requires DB)
5. **SQLite** - Database queries (disabled)
6. **GitHub** - GitHub API (disabled, requires token)
7. **Slack** - Slack integration (disabled, requires token)

Enable additional servers in `llamagate.yaml` as needed.

## Next Steps

- Explore more MCP servers from the [MCP Servers Directory](https://github.com/modelcontextprotocol/servers)
- Enable optional servers (Puppeteer, databases, GitHub, Slack) for extended capabilities
- Create custom workflows combining multiple tools
- Set up production configurations with proper security
- See [MCP.md](MCP.md) for full documentation
- See [MCP_QUICKSTART.md](MCP_QUICKSTART.md) for basic MCP setup

## Example: Complete Document Processing Pipeline

Here's a complete example that demonstrates the full workflow:

```python
from openai import OpenAI
import json

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="sk-llamagate"
)

# Complete document processing workflow
response = client.chat.completions.create(
    model="llama3.2",
    messages=[
        {
            "role": "system",
            "content": "You are a document processing assistant. Use available tools to process documents."
        },
        {
            "role": "user",
            "content": """Process the document 'report.txt':
            1. Read the text file
            2. Extract the main topics
            3. Create a structured summary with:
               - Title
               - Key points (bullet list)
               - Conclusion
            4. Save the summary as 'report_summary.txt'
            5. List all files in the workspace to confirm"""
        }
    ],
    temperature=0.7,
    max_tokens=2000
)

# Print the response
print("Response:")
print(response.choices[0].message.content)

# Check if tools were used
if response.choices[0].message.tool_calls:
    print("\nTools used:")
    for tool_call in response.choices[0].message.tool_calls:
        print(f"  - {tool_call.function.name}")
```

This workflow will:
1. Use `mcp.filesystem.read_file` to read the file
2. Process the content with the model
3. Use `mcp.filesystem.write_file` to save the summary
4. Use `mcp.filesystem.list_directory` to list files

All of this happens automatically through the tool execution loop!

