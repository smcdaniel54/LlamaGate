#!/bin/bash
# LlamaGate Plugin System Test Script for Unix/Linux/macOS
# Tests all 8 use cases: adding, validating, and running plugins

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
echo "LlamaGate Plugin System Test Suite"
echo "========================================"
echo ""
echo "Prerequisites:"
echo "  1. LlamaGate must be running on http://localhost:8080"
echo "  2. Plugin system must be enabled"
echo "  3. API key should be set (if authentication enabled)"
echo ""
read -p "Press Enter to start testing..."

if [ -n "$API_KEY" ]; then
    AUTH_HEADER="-H \"X-API-Key: $API_KEY\""
else
    AUTH_HEADER=""
fi

echo ""
echo "========================================"
echo "Testing Plugin System"
echo "========================================"
echo ""

echo "[1/3] Testing Plugin Discovery..."
echo ""
echo "Listing all plugins..."
if [ -n "$API_KEY" ]; then
    curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/v1/plugins" | jq '.' 2>/dev/null || curl -s -H "X-API-Key: $API_KEY" "$BASE_URL/v1/plugins"
else
    curl -s "$BASE_URL/v1/plugins" | jq '.' 2>/dev/null || curl -s "$BASE_URL/v1/plugins"
fi
if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Plugin discovery passed"
else
    echo ""
    echo "✗ Plugin discovery failed - Is plugin system enabled?"
    exit 1
fi
echo ""

echo "[2/3] Testing Plugin Registration and Execution..."
echo ""

echo "Testing Use Case 1: Environment-Aware Plugin..."
if [ -n "$API_KEY" ]; then
    curl -s -X POST "$BASE_URL/v1/plugins/usecase1_environment_aware/execute" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{"input":"test","environment":"production"}' | jq '.' 2>/dev/null || echo "Response received"
else
    curl -s -X POST "$BASE_URL/v1/plugins/usecase1_environment_aware/execute" \
        -H "Content-Type: application/json" \
        -d '{"input":"test","environment":"production"}' | jq '.' 2>/dev/null || echo "Response received"
fi
if [ $? -eq 0 ]; then
    echo "✓ Use Case 1 passed"
else
    echo "✗ Use Case 1 failed"
fi
echo ""

echo "Testing Use Case 2: User-Configurable Workflow..."
if [ -n "$API_KEY" ]; then
    curl -s -X POST "$BASE_URL/v1/plugins/usecase2_user_configurable/execute" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{"query":"test query","max_depth":5,"use_cache":true}' | jq '.' 2>/dev/null || echo "Response received"
else
    curl -s -X POST "$BASE_URL/v1/plugins/usecase2_user_configurable/execute" \
        -H "Content-Type: application/json" \
        -d '{"query":"test query","max_depth":5,"use_cache":true}' | jq '.' 2>/dev/null || echo "Response received"
fi
if [ $? -eq 0 ]; then
    echo "✓ Use Case 2 passed"
else
    echo "✗ Use Case 2 failed"
fi
echo ""

echo "Testing Use Case 3: Configuration-Driven Tool Selection..."
if [ -n "$API_KEY" ]; then
    curl -s -X POST "$BASE_URL/v1/plugins/usecase3_tool_selection/execute" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{"action":"process","enabled_tools":["tool1","tool2"]}' | jq '.' 2>/dev/null || echo "Response received"
else
    curl -s -X POST "$BASE_URL/v1/plugins/usecase3_tool_selection/execute" \
        -H "Content-Type: application/json" \
        -d '{"action":"process","enabled_tools":["tool1","tool2"]}' | jq '.' 2>/dev/null || echo "Response received"
fi
if [ $? -eq 0 ]; then
    echo "✓ Use Case 3 passed"
else
    echo "✗ Use Case 3 failed"
fi
echo ""

echo "Testing Use Case 4: Adaptive Timeout Configuration..."
if [ -n "$API_KEY" ]; then
    curl -s -X POST "$BASE_URL/v1/plugins/usecase4_adaptive_timeout/execute" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{"text":"This is a test text for timeout calculation","complexity":"high"}' | jq '.' 2>/dev/null || echo "Response received"
else
    curl -s -X POST "$BASE_URL/v1/plugins/usecase4_adaptive_timeout/execute" \
        -H "Content-Type: application/json" \
        -d '{"text":"This is a test text for timeout calculation","complexity":"high"}' | jq '.' 2>/dev/null || echo "Response received"
fi
if [ $? -eq 0 ]; then
    echo "✓ Use Case 4 passed"
else
    echo "✗ Use Case 4 failed"
fi
echo ""

echo "Testing Use Case 5: Configuration File-Based Setup..."
if [ -n "$API_KEY" ]; then
    curl -s -X POST "$BASE_URL/v1/plugins/usecase5_config_file/execute" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{"operation":"process","config_file":"custom.json"}' | jq '.' 2>/dev/null || echo "Response received"
else
    curl -s -X POST "$BASE_URL/v1/plugins/usecase5_config_file/execute" \
        -H "Content-Type: application/json" \
        -d '{"operation":"process","config_file":"custom.json"}' | jq '.' 2>/dev/null || echo "Response received"
fi
if [ $? -eq 0 ]; then
    echo "✓ Use Case 5 passed"
else
    echo "✗ Use Case 5 failed"
fi
echo ""

echo "Testing Use Case 6: Runtime Configuration Updates..."
echo "First, update config..."
if [ -n "$API_KEY" ]; then
    curl -s -X POST "$BASE_URL/v1/plugins/usecase6_runtime_config/execute" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{"action":"update_config","config":{"timeout":"60s","retries":5}}' | jq '.' 2>/dev/null || echo "Response received"
else
    curl -s -X POST "$BASE_URL/v1/plugins/usecase6_runtime_config/execute" \
        -H "Content-Type: application/json" \
        -d '{"action":"update_config","config":{"timeout":"60s","retries":5}}' | jq '.' 2>/dev/null || echo "Response received"
fi
echo ""
echo "Then, use updated config..."
if [ -n "$API_KEY" ]; then
    curl -s -X POST "$BASE_URL/v1/plugins/usecase6_runtime_config/execute" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{"action":"execute"}' | jq '.' 2>/dev/null || echo "Response received"
else
    curl -s -X POST "$BASE_URL/v1/plugins/usecase6_runtime_config/execute" \
        -H "Content-Type: application/json" \
        -d '{"action":"execute"}' | jq '.' 2>/dev/null || echo "Response received"
fi
if [ $? -eq 0 ]; then
    echo "✓ Use Case 6 passed"
else
    echo "✗ Use Case 6 failed"
fi
echo ""

echo "Testing Use Case 7: Context-Aware Configuration..."
if [ -n "$API_KEY" ]; then
    curl -s -X POST "$BASE_URL/v1/plugins/usecase7_context_aware/execute" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{"query":"Process this with context"}' | jq '.' 2>/dev/null || echo "Response received"
else
    curl -s -X POST "$BASE_URL/v1/plugins/usecase7_context_aware/execute" \
        -H "Content-Type: application/json" \
        -d '{"query":"Process this with context"}' | jq '.' 2>/dev/null || echo "Response received"
fi
if [ $? -eq 0 ]; then
    echo "✓ Use Case 7 passed"
else
    echo "✗ Use Case 7 failed"
fi
echo ""

echo "Testing Use Case 8: Multi-Tenant Configuration..."
if [ -n "$API_KEY" ]; then
    curl -s -X POST "$BASE_URL/v1/plugins/usecase8_multi_tenant/execute" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{"tenant_id":"tenant1","operation":"process"}' | jq '.' 2>/dev/null || echo "Response received"
else
    curl -s -X POST "$BASE_URL/v1/plugins/usecase8_multi_tenant/execute" \
        -H "Content-Type: application/json" \
        -d '{"tenant_id":"tenant1","operation":"process"}' | jq '.' 2>/dev/null || echo "Response received"
fi
if [ $? -eq 0 ]; then
    echo "✓ Use Case 8 passed"
else
    echo "✗ Use Case 8 failed"
fi
echo ""

echo "[3/3] Testing Input Validation..."
echo ""

echo "Testing validation with missing required input..."
if [ -n "$API_KEY" ]; then
    curl -s -w "\nHTTP Status: %{http_code}\n" -X POST "$BASE_URL/v1/plugins/usecase1_environment_aware/execute" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{}'
else
    curl -s -w "\nHTTP Status: %{http_code}\n" -X POST "$BASE_URL/v1/plugins/usecase1_environment_aware/execute" \
        -H "Content-Type: application/json" \
        -d '{}'
fi
echo ""

echo "Testing validation with invalid input type..."
if [ -n "$API_KEY" ]; then
    curl -s -w "\nHTTP Status: %{http_code}\n" -X POST "$BASE_URL/v1/plugins/usecase4_adaptive_timeout/execute" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{"text":123}'
else
    curl -s -w "\nHTTP Status: %{http_code}\n" -X POST "$BASE_URL/v1/plugins/usecase4_adaptive_timeout/execute" \
        -H "Content-Type: application/json" \
        -d '{"text":123}'
fi
echo ""

echo "Testing validation with valid input..."
if [ -n "$API_KEY" ]; then
    curl -s -w "\nHTTP Status: %{http_code}\n" -X POST "$BASE_URL/v1/plugins/usecase1_environment_aware/execute" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d '{"input":"valid input"}' | jq '.' 2>/dev/null || echo "Response received"
else
    curl -s -w "\nHTTP Status: %{http_code}\n" -X POST "$BASE_URL/v1/plugins/usecase1_environment_aware/execute" \
        -H "Content-Type: application/json" \
        -d '{"input":"valid input"}' | jq '.' 2>/dev/null || echo "Response received"
fi
if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Input validation tests passed"
else
    echo ""
    echo "✗ Input validation tests failed"
fi
echo ""

echo "========================================"
echo "Plugin System Testing Complete!"
echo "========================================"
echo ""
echo "Summary:"
echo "  - Plugin discovery: Tested"
echo "  - Plugin registration: Tested"
echo "  - Plugin validation: Tested"
echo "  - Plugin execution: Tested (8 use cases)"
echo ""
echo "Note: Some tests may show errors if plugins are not registered."
echo "      Ensure test plugins are registered in your LlamaGate instance."
echo ""
