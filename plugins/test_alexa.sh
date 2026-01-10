#!/bin/bash
# Test script for Alexa Skill plugin

echo "Testing Alexa Skill Plugin"
echo "=========================="
echo ""

# Test 1: List plugins
echo "1. Listing plugins..."
curl -s http://localhost:11435/v1/plugins | jq '.'
echo ""

# Test 2: Get Alexa plugin metadata
echo "2. Getting Alexa plugin metadata..."
curl -s http://localhost:11435/v1/plugins/alexa_skill | jq '.'
echo ""

# Test 3: Test Alexa endpoint with wake word
echo "3. Testing Alexa endpoint with wake word..."
curl -X POST http://localhost:11435/v1/plugins/alexa_skill/alexa \
  -H "Content-Type: application/json" \
  -d '{
    "version": "1.0",
    "session": {
      "new": true,
      "sessionId": "test-session-1",
      "application": {
        "applicationId": "test-app"
      },
      "user": {
        "userId": "test-user"
      }
    },
    "request": {
      "type": "IntentRequest",
      "requestId": "test-request-1",
      "timestamp": "2024-01-01T00:00:00Z",
      "locale": "en-US",
      "intent": {
        "name": "QueryIntent",
        "slots": {
          "query": {
            "name": "query",
            "value": "Smart Voice what is the weather today"
          }
        }
      }
    }
  }' | jq '.'
echo ""

# Test 4: Test Alexa endpoint without wake word
echo "4. Testing Alexa endpoint without wake word..."
curl -X POST http://localhost:11435/v1/plugins/alexa_skill/alexa \
  -H "Content-Type: application/json" \
  -d '{
    "version": "1.0",
    "session": {
      "new": true,
      "sessionId": "test-session-2",
      "application": {
        "applicationId": "test-app"
      },
      "user": {
        "userId": "test-user"
      }
    },
    "request": {
      "type": "IntentRequest",
      "requestId": "test-request-2",
      "timestamp": "2024-01-01T00:00:00Z",
      "locale": "en-US",
      "intent": {
        "name": "QueryIntent",
        "slots": {
          "query": {
            "name": "query",
            "value": "what is the weather today"
          }
        }
      }
    }
  }' | jq '.'
echo ""

echo "Tests complete!"
