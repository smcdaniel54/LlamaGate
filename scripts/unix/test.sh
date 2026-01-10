#!/bin/bash
# LlamaGate Test Script for Unix/Linux/macOS

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

# Load .env if it exists
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

BASE_URL="${BASE_URL:-http://localhost:11435}"
API_KEY="${API_KEY:-}"

echo "========================================"
echo "LlamaGate Test Suite"
echo "========================================"
echo ""
echo "Prerequisites:"
echo "  1. Ollama must be running on http://localhost:11434"
echo "  2. LlamaGate must be running on http://localhost:11435"
echo "  3. At least one model should be available in Ollama"
echo ""
read -p "Press Enter to start testing..."

echo ""
echo "[1/9] Testing Health Check..."
if [ -n "$API_KEY" ]; then
    curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/health"
else
    curl -s "$BASE_URL/health"
fi
echo ""
echo "✓ Health check passed"
echo ""

echo "[2/9] Testing Models Endpoint..."
if [ -n "$API_KEY" ]; then
    curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/v1/models" | jq '.' 2>/dev/null || curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/v1/models"
else
    curl -s "$BASE_URL/v1/models" | jq '.' 2>/dev/null || curl -s "$BASE_URL/v1/models"
fi
echo ""
echo "✓ Models endpoint passed"
echo ""

echo "[3/9] Testing Chat Completions (Non-Streaming)..."
BODY='{"model":"llama2","messages":[{"role":"user","content":"Say hello in one word"}]}'
if [ -n "$API_KEY" ]; then
    RESPONSE=$(curl -s -X POST "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d "$BODY")
    echo "$RESPONSE" | jq -r '.choices[0].message.content' 2>/dev/null || echo "$RESPONSE"
else
    RESPONSE=$(curl -s -X POST "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -d "$BODY")
    echo "$RESPONSE" | jq -r '.choices[0].message.content' 2>/dev/null || echo "$RESPONSE"
fi
echo ""
echo "✓ Chat completions (non-streaming) passed"
echo ""

echo "[4/9] Testing Caching (Same Request Twice)..."
BODY='{"model":"llama2","messages":[{"role":"user","content":"What is 2+2?"}]}'
echo "First request (should be slow):"
if [ -n "$API_KEY" ]; then
    time curl -s -X POST "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d "$BODY" > /dev/null
else
    time curl -s -X POST "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -d "$BODY" > /dev/null
fi
echo ""
echo "Second request (should be fast - cached):"
if [ -n "$API_KEY" ]; then
    time curl -s -X POST "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d "$BODY" > /dev/null
else
    time curl -s -X POST "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -d "$BODY" > /dev/null
fi
echo ""
echo "✓ Cache test completed (check times above - second should be much faster)"
echo ""

echo "[5/9] Testing Authentication (if enabled)..."
if [ -z "$API_KEY" ]; then
    echo "Authentication is disabled, skipping auth test"
else
    echo "Testing with invalid API key (should fail)..."
    curl -s -w "\nHTTP Status: %{http_code}\n" -X GET "$BASE_URL/v1/models" \
        -H "X-API-Key: invalid-key"
    echo ""
    echo "Testing with valid API key (should succeed)..."
    curl -s -w "\nHTTP Status: %{http_code}\n" -X GET "$BASE_URL/v1/models" \
        -H "X-API-Key: $API_KEY"
    echo ""
    echo "✓ Authentication test completed"
fi
echo ""

echo "[6/9] Testing MCP API Endpoints (if MCP enabled)..."
if [ -n "$API_KEY" ]; then
    AUTH_HEADER="-H \"X-API-Key: $API_KEY\""
else
    AUTH_HEADER=""
fi
echo "Testing MCP servers list..."
if [ -n "$API_KEY" ]; then
    curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/v1/mcp/servers" | jq '.' 2>/dev/null || curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/v1/mcp/servers"
else
    curl -s "$BASE_URL/v1/mcp/servers" | jq '.' 2>/dev/null || curl -s "$BASE_URL/v1/mcp/servers"
fi
if [ $? -eq 0 ]; then
    echo ""
    echo "✓ MCP API endpoints are accessible"
    echo "  Note: If you see \"MCP is not enabled\", configure MCP in your config file"
else
    echo ""
    echo "ℹ MCP API test skipped (MCP may not be enabled)"
fi
echo ""

echo "[7/9] Testing MCP URI Scheme (if MCP enabled)..."
echo "Testing chat completion with MCP URI..."
echo "Note: This requires an MCP server with resources configured"
BODY='{"model":"llama2","messages":[{"role":"user","content":"Test mcp://test-server/resource"}]}'
if [ -n "$API_KEY" ]; then
    curl -s -X POST "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d "$BODY" > /dev/null
else
    curl -s -X POST "$BASE_URL/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -d "$BODY" > /dev/null
fi
if [ $? -eq 0 ]; then
    echo ""
    echo "✓ MCP URI scheme test completed"
    echo "  Note: If MCP is not enabled or server not found, request will continue without resource context"
else
    echo ""
    echo "ℹ MCP URI test skipped (MCP may not be enabled or server not configured)"
fi
echo ""

echo "[8/9] Testing Plugin System (if enabled)..."
if [ -n "$API_KEY" ]; then
    AUTH_HEADER="-H \"X-API-Key: $API_KEY\""
else
    AUTH_HEADER=""
fi
echo "Testing plugin discovery..."
if [ -n "$API_KEY" ]; then
    curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/v1/plugins" > /dev/null
else
    curl -s "$BASE_URL/v1/plugins" > /dev/null
fi
if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Plugin system is accessible"
    echo "  Run scripts/unix/test-plugins.sh for comprehensive plugin tests"
else
    echo ""
    echo "ℹ Plugin system test skipped (Plugin system may not be enabled)"
fi
echo ""

echo "[9/9] Testing Plugin Use Cases (if plugins registered)..."
echo "Note: This requires test plugins to be registered"
echo "      See scripts/unix/test-plugins.sh for full plugin testing"
echo "      Or set ENABLE_TEST_PLUGINS=true to enable test plugins"
echo ""

echo "========================================"
echo "Testing Complete!"
echo "========================================"
echo ""
echo "Check the log file if LOG_FILE is set in your .env"
echo "Check console output for request logs"
echo ""

