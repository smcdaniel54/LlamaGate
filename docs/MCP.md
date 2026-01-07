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

# Connection pooling (for HTTP transport)
MCP_CONNECTION_POOL_SIZE=10
MCP_CONNECTION_IDLE_TIME=5m

# Health monitoring
MCP_HEALTH_CHECK_INTERVAL=60s
MCP_HEALTH_CHECK_TIMEOUT=5s

# Caching
MCP_CACHE_TTL=5m

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
  # Connection pooling (for HTTP transport)
  connection_pool_size: 10
  connection_idle_time: 5m
  # Health monitoring
  health_check_interval: 60s
  health_check_timeout: 5s
  # Caching
  cache_ttl: 5m
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
- **transport**: `stdio`, `http`, or `sse` (SSE not yet implemented)
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
  timeout: 30s
```

### HTTP Transport

For HTTP transport, specify:

- **url**: HTTP endpoint URL for the MCP server
- **headers**: HTTP headers (optional, e.g., for authentication)
- **timeout**: Request timeout

Example:
```yaml
- name: remote-server
  transport: http
  url: http://remote-server:3000/mcp
  headers:
    Authorization: Bearer token123
    X-API-Key: secret-key
  timeout: 30s
```

### SSE Transport

SSE transport is planned for a future release. For now, use stdio or http transport.

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

## Connection Pooling

For HTTP transport, LlamaGate uses connection pooling to improve performance and handle concurrent requests efficiently.

**Configuration:**
- `connection_pool_size`: Maximum number of connections per server pool (default: 10)
- `connection_idle_time`: Maximum time a connection can be idle before being closed (default: 5m)

Connection pooling is automatically enabled for HTTP transport. stdio transport doesn't use pooling as each connection is a separate process.

## Health Monitoring

LlamaGate automatically monitors the health of all MCP servers, performing periodic health checks to detect connection issues.

**Configuration:**
- `health_check_interval`: Interval between health checks (default: 60s)
- `health_check_timeout`: Timeout for individual health checks (default: 5s)

Health checks verify that servers are responsive by attempting to list tools. Unhealthy servers are logged and can be monitored via logs.

## Caching

LlamaGate caches tool definitions, resources, and prompts to reduce server load and improve response times.

**Configuration:**
- `cache_ttl`: Time-to-live for cached metadata (default: 5m)

Cached data is automatically refreshed when:
- Cache TTL expires
- Server connection is lost and re-established
- Manual refresh is triggered (via future API endpoints)

## Resources and Prompts

LlamaGate discovers and caches resources and prompts from MCP servers:

- **Resources**: Readable data exposed by MCP servers (e.g., files, database results)
- **Prompts**: Reusable, parameterized prompt templates

These are currently available internally and will be exposed via HTTP API in Phase 2.

## Limitations

- Streaming tool calls are not yet supported (planned for 1.1.x)
- SSE transport is not yet implemented (use stdio or http)
- Tool execution is limited to non-streaming requests
- Ollama model must support function/tool calling for best results
- Resources and prompts are discovered but not yet exposed via HTTP API (planned for Phase 2)
- HTTP API endpoints for MCP management are planned for Phase 2

