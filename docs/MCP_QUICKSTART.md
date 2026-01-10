# MCP Quick Start Guide

Get LlamaGate working with MCP in 5 minutes.

## Prerequisites

- LlamaGate installed and running
- Node.js installed (for MCP servers)
- An MCP server to test with (we'll use the filesystem server)

## Step 1: Install an MCP Server

Install the filesystem MCP server (or use any other MCP server):

```bash
npm install -g @modelcontextprotocol/server-filesystem
```

Or use npx (no installation needed):
```bash
npx -y @modelcontextprotocol/server-filesystem /path/to/allowed/directory
```

## Step 2: Configure LlamaGate

Create a `llamagate.yaml` file in your LlamaGate directory:

```yaml
mcp:
  enabled: true
  max_tool_rounds: 10
  max_tool_calls_per_round: 10
  default_tool_timeout: 30s
  max_tool_result_size: 1048576
  servers:
    - name: filesystem
      enabled: true
      transport: stdio
      command: npx
      args:
        - -y
        - @modelcontextprotocol/server-filesystem
        - /tmp  # Directory the server can access
      timeout: 30s
```

Or use environment variables:

```bash
export MCP_ENABLED=true
export MCP_MAX_TOOL_ROUNDS=10
export MCP_MAX_TOOL_CALLS_PER_ROUND=10
export MCP_DEFAULT_TOOL_TIMEOUT=30s
```

## Step 3: Start LlamaGate

Start LlamaGate with MCP enabled:

```bash
# Windows
scripts\windows\run.cmd

# Unix/Linux/macOS
./scripts/unix/run.sh
```

You should see logs indicating MCP clients are connecting:

```
INFO MCP client initialized server=filesystem
INFO Discovered tools from MCP server server=filesystem tool_count=5
INFO MCP initialization complete total_tools=5
```

## Step 4: Test Tool Execution

Create a test file:

```bash
echo "Hello, MCP!" > /tmp/test.txt
```

Make a chat completion request with tool calling:

```bash
curl http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "X-API-Key: sk-llamagate" \
  -d '{
    "model": "llama2",
    "messages": [
      {
        "role": "user",
        "content": "Read the file /tmp/test.txt and tell me what it says"
      }
    ]
  }'
```

Or use Python:

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:11435/v1",
    api_key="sk-llamagate"
)

response = client.chat.completions.create(
    model="llama2",
    messages=[
        {
            "role": "user",
            "content": "Read the file /tmp/test.txt and tell me what it says"
        }
    ]
)

print(response.choices[0].message.content)
```

## Step 5: Verify Tool Execution

Check the LlamaGate logs for tool execution:

```
INFO Tool executed successfully tool=mcp.filesystem.read_file duration=150ms
INFO Tool execution round completed round=1 tool_calls=1
```

The model should respond with the file contents.

## Example: Multi-Step Tool Flow

Test a multi-step flow (read → summarize → write):

```python
response = client.chat.completions.create(
    model="llama2",
    messages=[
        {
            "role": "user",
            "content": "Read /tmp/test.txt, create a summary, and write it to /tmp/summary.txt"
        }
    ]
)
```

This should trigger multiple tool calls:
1. `mcp.filesystem.read_file` - Read the file
2. Model processes and creates summary
3. `mcp.filesystem.write_file` - Write the summary

## Troubleshooting

### "MCP client not initialized"

- Check that `MCP_ENABLED=true` or MCP is enabled in config file
- Verify the MCP server command is correct
- Check that Node.js/npx is available

### "Tool not found"

- Verify the tool name uses the namespaced format: `mcp.<server>.<tool>`
- Check that tools were discovered on startup (check logs)
- Ensure the MCP server is running and connected

### "Tool call denied"

- Check allow/deny lists in configuration
- Verify the tool name matches an allowed pattern

### Tool execution timeout

- Increase `MCP_DEFAULT_TOOL_TIMEOUT` or per-server timeout
- Check if the MCP server is responding

## Next Steps

- Add more MCP servers (fetch, database, etc.)
- Configure allow/deny lists for security
- Adjust timeouts and limits based on your needs
- See [MCP.md](MCP.md) for full documentation

