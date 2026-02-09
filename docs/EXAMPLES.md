# LlamaGate Examples Index

This document provides a quick reference to all available examples for using LlamaGate with various SDKs and frameworks.

## Quick Links

### Basic Examples
- Simple Chat - See [API](API.md) and [README usage examples](../README.md#usage-examples)
- Streaming Chat - See [README streaming](../README.md#streaming-with-python-sdk)
- With Authentication - See [README authentication](../README.md#5-authentication-example-if-enabled)

### Advanced Examples
- [Error Handling](../README.md#error-handling) - Proper error handling patterns
- [Tool/Function Calling](../README.md#toolfunction-calling-with-mcp) - MCP tool execution
- [Environment Variables](../README.md#environment-variable-configuration) - Configuration via env vars
- [Production Patterns](../README.md#production-patterns) - Retries, timeouts, connection pooling

### Integration Examples
- [LangChain Integration](../README.md#using-with-langchain) - Using LlamaGate with LangChain
- [API](API.md) and [README](../README.md) - Python and cURL examples

## Example Sources

In-repo examples: see [API](API.md), [README](../README.md) (usage examples, streaming, authentication), and [MCP Quick Start](MCP_QUICKSTART.md).

(External example repositories previously linked here have been removed.)

### MCP Examples
**Status:** Coming soon

Examples for Model Context Protocol integration:
- MCP server setup
- Tool calling patterns
- Multi-server configurations

## Example Categories

### By SDK/Language

#### Python
- [OpenAI Python SDK](../README.md#usage-examples) - Main examples in README
- [LangChain](../README.md#using-with-langchain) - LangChain integration
- [API](API.md) and [README](../README.md) - In-repo examples

#### Node.js
- [OpenAI Node.js SDK](../README.md#4-using-openai-nodejs-sdk) - Basic Node.js examples

#### cURL
- [cURL Examples](../README.md#usage-examples) - Command-line examples

### By Use Case

#### Getting Started
1. [Quick Start Guide](../QUICKSTART.md) - Get running in 2 minutes
2. [Basic Chat Example](../README.md#usage-examples) - First API call
3. [Streaming Example](../README.md#streaming-with-python-sdk) - Real-time responses

#### Production Use
1. [Error Handling](../README.md#error-handling) - Handle errors gracefully
2. [Production Patterns](../README.md#production-patterns) - Retries, timeouts, pooling
3. [Authentication](../README.md#5-authentication-example-if-enabled) - Secure API access

#### Advanced Features
1. [Tool/Function Calling](../README.md#toolfunction-calling-with-mcp) - MCP tool execution
2. [MCP Quick Start](MCP_QUICKSTART.md) - MCP setup and configuration
3. [Environment Variables](../README.md#environment-variable-configuration) - Configuration management

#### Integrations
1. [LangChain](../README.md#using-with-langchain) - LangChain integration
2. [API](API.md) and [README](../README.md) - In-repo examples

## Prerequisites

Before running examples, ensure you have:

1. **LlamaGate installed and running**
   - See [Installation Guide](INSTALL.md)
   - Or [Quick Start Guide](../QUICKSTART.md)

2. **Ollama installed with at least one model**
   - Default model: `mistral` (Mistral 7B)
   - Install: `ollama pull mistral`
   - See [Model Recommendations](MODEL_RECOMMENDATIONS.md) for other options

3. **Required SDKs/libraries** (for code examples)
   - Python: `pip install openai`
   - Node.js: `npm install openai`
   - LangChain: `pip install langchain-openai`

## Running Examples

### Python Examples

```bash
# Install dependencies
pip install openai

# Run example (replace with actual example file)
python example.py
```

### Node.js Examples

```bash
# Install dependencies
npm install openai

# Run example (replace with actual example file)
node example.js
```

### cURL Examples

```bash
# Copy and paste the curl command from documentation
curl http://localhost:11435/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "mistral", "messages": [{"role": "user", "content": "Hello!"}]}'
```

## Troubleshooting

### Common Issues

1. **Connection refused**
   - Ensure LlamaGate is running: `scripts/windows/run.cmd` or `./scripts/unix/run.sh`
   - Check the port: Default is `11435`

2. **Model not found**
   - Pull the model: `ollama pull mistral`
   - Check available models: `curl http://localhost:11435/v1/models`

3. **Authentication errors**
   - Check if `API_KEY` is set in LlamaGate config
   - Verify the API key in your request matches the config

4. **Tool calling not working**
   - Ensure MCP is enabled in LlamaGate configuration
   - Check MCP server is running and connected
   - See [MCP Quick Start](MCP_QUICKSTART.md) for setup

## Related Documentation

- [Main README](../README.md) - Complete feature reference with examples
- [Quick Start Guide](../QUICKSTART.md) - Get started in 2 minutes
- [MCP Quick Start](MCP_QUICKSTART.md) - MCP integration guide
- [Model Recommendations](MODEL_RECOMMENDATIONS.md) - Best models for your hardware
- [API Documentation](API.md) - Complete API reference

## Contributing Examples

Have a great example to share? Contributions are welcome!

1. **Examples:** See [API](API.md) and [README](../README.md); submit changes to your LlamaGate repository.
2. **Documentation Examples:** Submit PRs to this repository with examples in README.md or new docs

---

**Last Updated:** 2026-01-15  
**Examples:** In-repo; see [API](API.md) and [README](../README.md).
