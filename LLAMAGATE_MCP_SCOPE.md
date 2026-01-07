# LlamaGate MCP Scope Document

This document defines the scope, boundaries, and implementation status of Model Context Protocol (MCP) functionality in LlamaGate.

## Overview

LlamaGate implements MCP as a **client** that connects to MCP servers and exposes their tools to local LLM models through OpenAI-compatible chat completions. LlamaGate does **not** implement MCP server functionality.

## In Scope (Implemented)

### Core MCP Client Features

✅ **MCP Client Implementation**
- Connect to one or more MCP servers on startup
- Support for stdio transport (primary transport)
- Automatic tool discovery from connected servers
- Tool namespacing: `mcp.<serverName>.<toolName>`
- Multi-server support with independent configuration

✅ **Tool Execution**
- Execute tool calls from LLM models
- Multi-round tool execution loops
- Tool result injection back into conversation
- Per-tool timeout enforcement
- Result size limits with truncation

✅ **Security & Guardrails**
- Allow/deny lists using glob patterns
- Maximum tool execution rounds limit
- Maximum tool calls per round limit
- Per-tool timeout configuration
- Maximum tool result size limits
- Tool validation before execution

✅ **Configuration**
- Environment variable configuration
- YAML/JSON configuration file support
- Per-server configuration (command, args, env, timeout)
- Enable/disable individual servers
- Global MCP enable/disable

✅ **Integration**
- OpenAI-compatible tool/function calling
- Automatic tool injection into chat completions
- Support for non-streaming requests
- Request ID tracking for tool execution

## Planned (Future Releases)

🔄 **SSE Transport Support**
- Server-Sent Events (SSE) transport for remote MCP servers
- HTTP-based MCP server connections
- Authentication headers for remote servers
- Status: Interface prepared, implementation pending

🔄 **Streaming Tool Calls**
- Support for tool execution in streaming chat completions
- Real-time tool result streaming
- Status: Planned for version 1.1.x

## Out of Scope (Not Included)

❌ **MCP Server Implementation**
- LlamaGate does not implement MCP server functionality
- LlamaGate is a client-only implementation
- Use separate MCP server packages (e.g., `@modelcontextprotocol/server-*`)

❌ **MCP Protocol Extensions**
- Custom MCP protocol extensions
- Non-standard MCP features
- Protocol modifications or deviations

❌ **Built-in MCP Servers**
- LlamaGate does not bundle any MCP servers
- Users must install and configure MCP servers separately
- LlamaGate only provides the client connection layer

❌ **MCP Server Management**
- Automatic MCP server installation
- MCP server lifecycle management (start/stop/restart)
- MCP server health monitoring
- MCP server auto-recovery

❌ **Advanced MCP Features**
- MCP resource management (beyond tools)
- MCP prompt templates
- MCP sampling parameters
- MCP server discovery/registry

❌ **Enterprise MCP Features**
- Multi-tenant MCP isolation
- MCP access control/authorization
- MCP audit logging
- MCP usage analytics

## Implementation Details

### Supported MCP Protocol Features

| Feature | Status | Notes |
|---------|--------|-------|
| Tool Discovery | ✅ Implemented | Via `tools/list` |
| Tool Execution | ✅ Implemented | Via `tools/call` |
| stdio Transport | ✅ Implemented | Primary transport |
| SSE Transport | 🔄 Planned | Interface ready |
| Resources | ❌ Not Implemented | Out of scope |
| Prompts | ❌ Not Implemented | Out of scope |
| Sampling | ❌ Not Implemented | Out of scope |

### Tool Execution Flow

1. **Startup**: LlamaGate connects to configured MCP servers
2. **Discovery**: Tools are discovered and namespaced
3. **Request**: Client sends chat completion with tool support
4. **Tool Call**: Model requests tool execution
5. **Validation**: LlamaGate validates against allow/deny lists
6. **Execution**: Tool is executed via MCP with timeout
7. **Injection**: Result is injected back into conversation
8. **Loop**: Process repeats until max rounds or no more tool calls

### Configuration Scope

**Supported Configuration:**
- MCP enable/disable
- Server connection details (command, args, env)
- Transport type (stdio)
- Timeout settings
- Tool allow/deny lists
- Execution limits (rounds, calls per round, result size)

**Not Supported:**
- Dynamic server addition/removal (requires restart)
- Runtime configuration changes
- Server health check configuration
- Automatic failover configuration

## Security Boundaries

### What LlamaGate Controls

✅ **Tool Access Control**
- Allow/deny lists for tool execution
- Tool execution limits and timeouts
- Result size limits

✅ **Request Validation**
- Tool call validation before execution
- Timeout enforcement
- Size limit enforcement

### What MCP Servers Control

🔒 **Tool Implementation**
- Actual tool functionality
- Tool permissions and capabilities
- Server-side security

🔒 **Resource Access**
- File system access (for filesystem server)
- Network access (for fetch server)
- Database access (for database servers)

### Shared Responsibility

⚠️ **Security Best Practices**
- Users must configure appropriate allow/deny lists
- Users must set appropriate timeouts
- Users must review MCP server permissions
- Users must secure MCP server configurations

## Compatibility

### MCP Protocol Version

- **Target**: MCP protocol as defined by Model Context Protocol specification
- **Compatibility**: Works with standard MCP servers
- **Testing**: Tested with official `@modelcontextprotocol/server-*` packages

### LLM Model Requirements

- **Required**: Model must support function/tool calling
- **Recommended**: Models with strong tool calling capabilities (e.g., Llama 3.1+, Mistral, etc.)
- **Not Required**: Specific model versions (works with any tool-calling model)

## Limitations

### Current Limitations

1. **Transport**: Only stdio transport is implemented
2. **Streaming**: Tool calls not supported in streaming requests
3. **Resources**: MCP resources are not supported
4. **Prompts**: MCP prompt templates are not supported
5. **Dynamic**: Server configuration requires restart to change

### Design Limitations

1. **Client-Only**: LlamaGate is a client, not a server
2. **Local Focus**: Designed for local Ollama instances
3. **Single Process**: MCP servers run as child processes
4. **No Clustering**: No built-in multi-instance support

## Testing Scope

### What is Tested

✅ Unit tests for tool execution logic
✅ Unit tests for guardrails and validation
✅ Integration tests with mock MCP servers
✅ Configuration loading tests

### What is Not Tested

❌ Compatibility with all MCP servers (tested with official servers)
❌ Performance under high tool call load
❌ Multi-server stress testing
❌ SSE transport (not yet implemented)

## Documentation Scope

### Included Documentation

✅ MCP integration guide (`docs/MCP.md`)
✅ MCP quick start guide (`docs/MCP_QUICKSTART.md`)
✅ MCP demo guide (`docs/MCP_DEMO_QUICKSTART.md`)
✅ Configuration examples (`mcp-config.example.yaml`)
✅ Demo configurations (`mcp-demo-config.yaml`)

### Not Included

❌ MCP protocol specification (refer to official docs)
❌ MCP server development guide (refer to official docs)
❌ Detailed MCP server documentation (refer to server docs)

## Version History

### v1.0.0 (Current)
- ✅ stdio transport support
- ✅ Tool execution with multi-round loops
- ✅ Security guardrails (allow/deny lists, limits)
- ✅ YAML/JSON configuration
- ✅ Environment variable configuration

### Planned v1.1.x
- 🔄 SSE transport support
- 🔄 Streaming tool calls

## References

- [Model Context Protocol Specification](https://modelcontextprotocol.io/)
- [Official MCP Servers](https://github.com/modelcontextprotocol/servers)
- [LlamaGate MCP Documentation](docs/MCP.md)

---

**Last Updated**: 2026-01-XX  
**Maintained By**: LlamaGate Project  
**Scope Version**: 1.0

