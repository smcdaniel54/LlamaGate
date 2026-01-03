#!/bin/bash

# LlamaGate Demo Script
# Showcases the power of LlamaGate in under 60 seconds

set -e

BASE_URL="http://localhost:8080"
API_KEY="${API_KEY:-sk-llamagate}"

echo "╔════════════════════════════════════════════════════════════╗"
echo "║          🚀 LlamaGate Demo - See the Magic! 🚀            ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Check if LlamaGate is running
echo -e "${BLUE}[1/5]${NC} Checking LlamaGate..."
if curl -s -f "${BASE_URL}/health" > /dev/null; then
    echo -e "${GREEN}✓${NC} LlamaGate is running!"
else
    echo -e "${YELLOW}✗${NC} LlamaGate is not running. Start it with: ./run.sh"
    exit 1
fi
echo ""

# Check Ollama connectivity
echo -e "${BLUE}[2/5]${NC} Checking Ollama connection..."
HEALTH=$(curl -s "${BASE_URL}/health")
if echo "$HEALTH" | grep -q "healthy"; then
    echo -e "${GREEN}✓${NC} Ollama is connected!"
else
    echo -e "${YELLOW}✗${NC} Ollama is not reachable. Make sure Ollama is running."
    exit 1
fi
echo ""

# List available models
echo -e "${BLUE}[3/5]${NC} Listing available models..."
MODELS=$(curl -s -H "X-API-Key: ${API_KEY}" "${BASE_URL}/v1/models" 2>/dev/null || curl -s "${BASE_URL}/v1/models")
MODEL_COUNT=$(echo "$MODELS" | grep -o '"id"' | wc -l | tr -d ' ')
echo -e "${GREEN}✓${NC} Found ${MODEL_COUNT} model(s) available"
echo "$MODELS" | grep -o '"id":"[^"]*"' | head -3 | sed 's/"id":"/  - /' | sed 's/"$//'
echo ""

# Model loading warning
echo -e "${CYAN}ℹ️  Note:${NC} First request to a model may take 5-30+ seconds"
echo "   (Ollama needs to load model weights into memory)"
echo "   Subsequent requests are fast once the model is loaded."
echo ""

# First request (slow - from Ollama, may also load model)
echo -e "${BLUE}[4/5]${NC} Making first request (this may be slow - loading model + hitting Ollama)..."
echo -e "   ${YELLOW}⏳ Please wait, this may take 10-30 seconds on first run...${NC}"
START_TIME=$(date +%s%N)
RESPONSE1=$(curl -s -X POST "${BASE_URL}/v1/chat/completions" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: ${API_KEY}" \
    -d '{
        "model": "llama2",
        "messages": [{"role": "user", "content": "Say hello in exactly one word"}]
    }' 2>/dev/null || curl -s -X POST "${BASE_URL}/v1/chat/completions" \
    -H "Content-Type: application/json" \
    -d '{
        "model": "llama2",
        "messages": [{"role": "user", "content": "Say hello in exactly one word"}]
    }')
END_TIME=$(date +%s%N)
TIME1=$((($END_TIME - $START_TIME) / 1000000))
CONTENT1=$(echo "$RESPONSE1" | grep -o '"content":"[^"]*"' | head -1 | sed 's/"content":"//' | sed 's/"$//' || echo "Response received")
echo -e "  Response: ${GREEN}${CONTENT1}${NC}"
if [ $TIME1 -lt 1000 ]; then
    echo -e "  Time: ${GREEN}${TIME1}ms${NC} (model already loaded)"
else
    echo -e "  Time: ${YELLOW}$(echo "scale=1; $TIME1/1000" | bc)s${NC} (includes model loading)"
fi
echo ""

# Second request (fast - from cache!)
echo -e "${BLUE}[5/5]${NC} Making identical request (this will be INSTANT - from cache!)..."
START_TIME=$(date +%s%N)
RESPONSE2=$(curl -s -X POST "${BASE_URL}/v1/chat/completions" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: ${API_KEY}" \
    -d '{
        "model": "llama2",
        "messages": [{"role": "user", "content": "Say hello in exactly one word"}]
    }' 2>/dev/null || curl -s -X POST "${BASE_URL}/v1/chat/completions" \
    -H "Content-Type: application/json" \
    -d '{
        "model": "llama2",
        "messages": [{"role": "user", "content": "Say hello in exactly one word"}]
    }')
END_TIME=$(date +%s%N)
TIME2=$((($END_TIME - $START_TIME) / 1000000))
CONTENT2=$(echo "$RESPONSE2" | grep -o '"content":"[^"]*"' | head -1 | sed 's/"content":"//' | sed 's/"$//' || echo "Response received")
echo -e "  Response: ${GREEN}${CONTENT2}${NC}"
echo -e "  Time: ${GREEN}${TIME2}ms${NC} (from cache!)"
echo ""

# Calculate speedup
if [ $TIME1 -gt 0 ] && [ $TIME2 -gt 0 ] && [ $TIME1 -gt $TIME2 ]; then
    SPEEDUP=$(echo "scale=1; $TIME1 / $TIME2" | bc)
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║                    🎉 Demo Complete! 🎉                    ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""
    echo -e "${GREEN}Cache Speedup: ${SPEEDUP}x faster!${NC}"
    echo ""
    echo "This is the power of LlamaGate:"
    if [ $TIME1 -lt 1000 ]; then
        echo "  • First request: ${TIME1}ms (model loaded, from Ollama)"
    else
        echo "  • First request: $(echo "scale=1; $TIME1/1000" | bc)s (includes model loading)"
    fi
    echo "  • Cached request: ${TIME2}ms (instant!)"
    echo ""
    echo "Your OpenAI code works immediately - just change the base_url!"
    echo ""
    echo "Next steps:"
    echo "  • See README.md for full documentation"
    echo "  • See QUICKSTART.md for quick setup guide"
    echo "  • See DEMO_QUICKSTART.md for migration examples"
    echo "  • Edit .env file to customize settings"
    echo ""
else
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║                    🎉 Demo Complete! 🎉                    ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""
    echo "Demo completed successfully!"
    echo ""
    echo "Note: If the first request was very fast, the model was already loaded."
    echo "      Try switching to a different model to see the loading time."
    echo ""
    echo "Next steps:"
    echo "  • See README.md for full documentation"
    echo "  • See QUICKSTART.md for quick setup guide"
    echo "  • See DEMO_QUICKSTART.md for migration examples"
    echo ""
fi

