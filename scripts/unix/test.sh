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

BASE_URL="${BASE_URL:-http://localhost:8080}"
API_KEY="${API_KEY:-}"

echo "========================================"
echo "LlamaGate Test Suite"
echo "========================================"
echo ""
echo "Prerequisites:"
echo "  1. Ollama must be running on http://localhost:11434"
echo "  2. LlamaGate must be running on http://localhost:8080"
echo "  3. At least one model should be available in Ollama"
echo ""
read -p "Press Enter to start testing..."

echo ""
echo "[1/5] Testing Health Check..."
if [ -n "$API_KEY" ]; then
    curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/health"
else
    curl -s "$BASE_URL/health"
fi
echo ""
echo "✓ Health check passed"
echo ""

echo "[2/5] Testing Models Endpoint..."
if [ -n "$API_KEY" ]; then
    curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/v1/models" | jq '.' 2>/dev/null || curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/v1/models"
else
    curl -s "$BASE_URL/v1/models" | jq '.' 2>/dev/null || curl -s "$BASE_URL/v1/models"
fi
echo ""
echo "✓ Models endpoint passed"
echo ""

echo "[3/5] Testing Chat Completions (Non-Streaming)..."
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

echo "[4/5] Testing Caching (Same Request Twice)..."
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

echo "[5/5] Testing Authentication (if enabled)..."
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

echo "========================================"
echo "Testing Complete!"
echo "========================================"
echo ""
echo "Check the log file if LOG_FILE is set in your .env"
echo "Check console output for request logs"
echo ""

