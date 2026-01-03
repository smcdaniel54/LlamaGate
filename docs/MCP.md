# MCP Client Integration

LlamaGate supports the Model Context Protocol (MCP) as a client, allowing you to connect to MCP servers and expose their tools to your local LLM models through OpenAI-compatible chat completions.

## Overview

When MCP is enabled, LlamaGate:

1. Connects to one or more MCP servers on startup
2. Discovers available tools from each server
3. Namespaces tools as `mcp.<serverName>.<toolName>` to avoid collisions
4. Exposes tools to chat completion requests
5. Executes tool calls when models request them
6. Injects tool results back into the conversation

## Configuration

MCP can be configured via environment variables or a YAML/JSON configuration file.

### Environment Variables

```bash
# Enable MCP
MCP_ENABLED=true

# Tool execution limits
MCP_MAX_TOOL_ROUNDS=10
MCP_MAX_TOOL_CALLS_PER_ROUND=10
MCP_DEFAULT_TOOL_TIMEOUT=30s
MCP_MAX_TOOL_RESULT_SIZE=1048576  # 1MB in bytes

# Allow/deny lists (comma-separated glob patterns)
MCP_ALLOW_TOOLS="mcp.filesystem.*,mcp.fetch.*"
MCP_DENY_TOOLS="mcp.dangerous.*"
```

### YAML Configuration File

Create a `llamagate.yaml` file in the project root or `~/.llamagate/`:

```yaml
mcp:
  enabled: true
  max_tool_rounds: 10
  max_tool_calls_per_round: 10
  default_tool_timeout: 30s
  max_tool_result_size: 1048576
  allow_tools:
    - "mcp.filesystem.*"
    - "mcp.fetch.*"
  deny_tools:
    - "mcp.dangerous.*"
  servers:
    - name: filesystem
      enabled: true
      transport: stdio
      command: npx
      args:
        - -y
        - @modelcontextprotocol/server-filesystem
        - /path/to/allowed/directory
      env:
        NODE_ENV: production
      timeout: 30s
    - name: fetch
      enabled: true
      transport: stdio
      command: npx
      args:
        - -y
        - @modelcontextprotocol/server-fetch
      timeout: 60s
```

## MCP Server Configuration

Each MCP server requires:

- **name**: Unique identifier for the server (used in tool namespacing)
- **enabled**: Whether to connect to this server
- **transport**: `stdio` (required) or `sse` (optional, not yet implemented)
- **timeout**: Per-server timeout for operations

### stdio Transport

For stdio transport, specify:

- **command**: Command to execute (e.g., `npx`, `python`, `node`)
- **args**: Command arguments
- **env**: Environment variables (optional)

Example:
```yaml
- name: filesystem
  transport: stdio
  command: npx
  args:
    - -y
    - @modelcontextprotocol/server-filesystem
    - /path/to/directory
  env:
    NODE_ENV: production
```

### SSE Transport

SSE transport is planned for a future release. For now, use stdio transport.

## Tool Namespacing

Tools are automatically namespaced to avoid collisions:

- Format: `mcp.<serverName>.<toolName>`
- Example: `mcp.filesystem.read_file`

This ensures tools from different servers don't conflict, even if they have the same name.

## Tool Execution Flow

1. Client sends chat completion request with tools (or tools are auto-injected)
2. Model processes request and may return tool calls
3. LlamaGate validates tool calls against allow/deny lists
4. Tools are executed via MCP with timeout enforcement
5. Tool results are injected back into the conversation
6. Process repeats until no more tool calls or max rounds reached

## Guardrails

### Allow/Deny Lists

Use glob patterns to control which tools can be executed:

- `mcp.filesystem.*` - Allow all filesystem tools
- `mcp.*.read_*` - Allow all read operations from any server
- `mcp.dangerous.*` - Deny all dangerous tools

Deny patterns take precedence over allow patterns.

### Limits

- **max_tool_rounds**: Maximum number of tool execution rounds (default: 10)
- **max_tool_calls_per_round**: Maximum tool calls per round (default: 10)
- **default_tool_timeout**: Timeout for tool execution (default: 30s)
- **max_tool_result_size**: Maximum result size before truncation (default: 1MB)

## Usage Example

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="sk-llamagate"
)

# Tools are automatically available if MCP is enabled
response = client.chat.completions.create(
    model="llama2",
    messages=[
        {"role": "user", "content": "Read the file /path/to/file.txt and summarize it"}
    ],
    tools=[  # Optional: explicitly request tools
        {
            "type": "function",
            "function": {
                "name": "mcp.filesystem.read_file",
                "description": "Read a file"
            }
        }
    ]
)

print(response.choices[0].message.content)
```

## Troubleshooting

### MCP Server Not Connecting

- Check that the command and arguments are correct
- Verify the MCP server is installed and accessible
- Check logs for connection errors

### Tools Not Available

- Ensure MCP is enabled in configuration
- Verify servers are enabled and connected
- Check that tools were discovered on startup (check logs)

### Tool Execution Fails

- Check allow/deny lists - tool might be blocked
- Verify timeout is sufficient for the operation
- Check MCP server logs for errors

### Tool Results Truncated

- Increase `max_tool_result_size` if needed
- Consider splitting large operations into smaller ones

## Security Considerations

- Use allow/deny lists to restrict tool access
- Set appropriate timeouts to prevent hanging operations
- Limit tool result sizes to prevent memory issues
- Review tool permissions before enabling servers
- Use environment variables for sensitive configuration

## Limitations

- Streaming tool calls are not yet supported (planned for 1.1.x)
- SSE transport is not yet implemented
- Tool execution is limited to non-streaming requests
- Ollama model must support function/tool calling for best results

