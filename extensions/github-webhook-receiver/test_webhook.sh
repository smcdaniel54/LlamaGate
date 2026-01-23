#!/bin/bash

# Test GitHub Webhook Receiver Extension
# Requires: LlamaGate running with dynamic endpoints support

LLAMAGATE_URL="${LLAMAGATE_URL:-http://localhost:11435}"
ENDPOINT="${LLAMAGATE_URL}/v1/extensions/github-webhook-receiver/webhooks/github"

echo "Testing GitHub Webhook Receiver Extension"
echo "Endpoint: ${ENDPOINT}"
echo ""

# Send test webhook
response=$(curl -s -X POST "${ENDPOINT}" \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: pull_request" \
  -d @test_input.json)

echo "Response:"
echo "${response}" | jq '.'

# Check if successful
if echo "${response}" | jq -e '.success == true' > /dev/null; then
    echo ""
    echo "✅ Test passed!"
    exit 0
else
    echo ""
    echo "❌ Test failed!"
    exit 1
fi
