# Alexa Skill Plugin for LlamaGate

This plugin provides Alexa Skill integration for LlamaGate, allowing Alexa devices to interact with local LLM models through LlamaGate.

## Features

- ✅ **Alexa Request Handling**: Parses and processes Alexa JSON requests
- ✅ **Wake Word Detection**: Detects "Smart Voice" (and variations) in queries
- ✅ **LLM Integration**: Processes queries through LlamaGate's LLM
- ✅ **Response Formatting**: Formats responses in Alexa-compatible JSON format
- ✅ **Custom Endpoint**: Exposes `/v1/plugins/alexa_skill/alexa` endpoint

## Installation

The plugin is automatically registered when LlamaGate starts. No additional configuration needed!

## Usage

### Endpoint

The plugin exposes a custom endpoint:

```
POST /v1/plugins/alexa_skill/alexa
```

### Request Format

The endpoint accepts standard Alexa Skill request format:

```json
{
  "version": "1.0",
  "session": {
    "new": true,
    "sessionId": "session-id",
    "application": {
      "applicationId": "app-id"
    },
    "user": {
      "userId": "user-id"
    }
  },
  "request": {
    "type": "IntentRequest",
    "requestId": "request-id",
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
}
```

### Response Format

The endpoint returns standard Alexa Skill response format:

```json
{
  "version": "1.0",
  "response": {
    "outputSpeech": {
      "type": "PlainText",
      "text": "The weather today is sunny with a high of 75°F."
    },
    "shouldEndSession": true
  }
}
```

## Wake Word Detection

The plugin detects the wake word "Smart Voice" (and variations) in the query text:

- ✅ "Smart Voice" (two words, standard)
- ✅ "Smartvoice" (one word, no space)
- ✅ "smart voice" (lowercase)
- ✅ "smartvoice" (lowercase, one word)
- ✅ "Alexa Smartvoice" (with Alexa prefix)

When the wake word is detected:
1. The wake word is removed from the query
2. The processed query is sent to LlamaGate's LLM
3. The LLM response is formatted and returned to Alexa

When the wake word is NOT detected:
- Returns a default response: "I heard your request, but I'm not configured to handle it yet."

## Configuration

The plugin can be configured by modifying the `NewAlexaSkillPlugin()` function:

```go
func NewAlexaSkillPlugin() *AlexaSkillPlugin {
	return &AlexaSkillPlugin{
		wakeWord:        "Smart Voice",  // Change wake word
		caseSensitive:   false,          // Case sensitivity
		removeFromQuery: true,           // Remove wake word from query
		defaultModel:    "llama3.2",     // Default LLM model
	}
}
```

## Testing

### Using curl (Linux/macOS)

```bash
curl -X POST http://localhost:11435/v1/plugins/alexa_skill/alexa \
  -H "Content-Type: application/json" \
  -d '{
    "version": "1.0",
    "session": {
      "new": true,
      "sessionId": "test-session",
      "application": {"applicationId": "test-app"},
      "user": {"userId": "test-user"}
    },
    "request": {
      "type": "IntentRequest",
      "requestId": "test-request",
      "timestamp": "2024-01-01T00:00:00Z",
      "locale": "en-US",
      "intent": {
        "name": "QueryIntent",
        "slots": {
          "query": {
            "name": "query",
            "value": "Smart Voice what is the weather"
          }
        }
      }
    }
  }'
```

### Using PowerShell (Windows)

```powershell
$body = @{
    version = "1.0"
    session = @{
        new = $true
        sessionId = "test-session"
        application = @{ applicationId = "test-app" }
        user = @{ userId = "test-user" }
    }
    request = @{
        type = "IntentRequest"
        requestId = "test-request"
        timestamp = "2024-01-01T00:00:00Z"
        locale = "en-US"
        intent = @{
            name = "QueryIntent"
            slots = @{
                query = @{
                    name = "query"
                    value = "Smart Voice what is the weather"
                }
            }
        }
    }
} | ConvertTo-Json -Depth 10

Invoke-RestMethod -Uri "http://localhost:11435/v1/plugins/alexa_skill/alexa" -Method Post -Body $body -ContentType "application/json"
```

### Using Test Scripts

**Linux/macOS:**
```bash
chmod +x plugins/test_alexa.sh
./plugins/test_alexa.sh
```

**Windows:**
```powershell
.\plugins\test_alexa.ps1
```

## Alexa Skill Configuration

To use this plugin with an Alexa Skill:

1. **Create an Alexa Skill** in the Alexa Developer Console
2. **Set the endpoint URL** to: `http://your-server:11435/v1/plugins/alexa_skill/alexa`
3. **Configure the skill** to use the "QueryIntent" intent with a "query" slot
4. **Test the skill** using the Alexa Developer Console or an Alexa device

**Note:** For production use, you'll need:
- HTTPS endpoint (use a reverse proxy like nginx or Caddy)
- Public IP address or domain name
- SSL certificate (Let's Encrypt recommended)

## LLM Integration

Currently, the plugin returns a placeholder response. To integrate with LlamaGate's LLM:

1. The plugin needs access to LlamaGate's proxy instance or
2. Make HTTP calls to LlamaGate's `/v1/chat/completions` endpoint

The LLM integration is marked as "pending" in the code and needs to be completed.

## Status

- ✅ Plugin structure created
- ✅ Wake word detection implemented
- ✅ Alexa request/response handling implemented
- ✅ Custom endpoint registered
- ⏳ LLM integration pending (placeholder response)

## Next Steps

1. Complete LLM integration (call LlamaGate's LLM API)
2. Add configuration file support
3. Add session management
4. Add error handling improvements
5. Add logging and monitoring

## License

Same as LlamaGate project (MIT License)
