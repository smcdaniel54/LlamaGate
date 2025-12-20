#!/bin/bash
# LlamaGate Runner for Unix/Linux/macOS

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

# Load .env file if it exists
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Set defaults if not set
export OLLAMA_HOST="${OLLAMA_HOST:-http://localhost:11434}"
export API_KEY="${API_KEY:-}"
export RATE_LIMIT_RPS="${RATE_LIMIT_RPS:-10}"
export DEBUG="${DEBUG:-false}"
export PORT="${PORT:-8080}"
export LOG_FILE="${LOG_FILE:-}"

echo "========================================"
echo "LlamaGate - OpenAI-Compatible Proxy"
echo "========================================"
echo ""
if [ -f .env ]; then
    echo "Configuration loaded from .env file"
    echo "(Environment variables override .env values)"
    echo ""
else
    echo "Tip: Create a .env file for easier configuration"
    echo ""
fi
echo "Configuration:"
echo "  OLLAMA_HOST: $OLLAMA_HOST"
echo "  API_KEY: ${API_KEY:-<not set>}"
if [ -z "$API_KEY" ]; then
    echo "    (Authentication disabled)"
else
    echo "    (Authentication enabled)"
fi
echo "  RATE_LIMIT_RPS: $RATE_LIMIT_RPS"
echo "  DEBUG: $DEBUG"
echo "  PORT: $PORT"
echo ""
echo "Starting server..."
echo "Press Ctrl+C to stop"
echo "========================================"
echo ""

./llamagate

