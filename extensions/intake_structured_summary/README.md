# Intake Structured Summary Extension

This extension processes raw request text and extracts structured information including summary, intent, urgency level, and actionable items.

## What It Does

Takes unstructured text input (e.g., support tickets, feature requests, bug reports) and converts it into a structured JSON format that can be used for:
- Automated routing and prioritization
- Ticket categorization
- Action item extraction
- Urgency assessment

## Inputs

- **request_text** (string, required): Raw text of the incoming request or ticket

## Outputs

- **summary** (string): Brief 2-3 sentence summary of the request
- **intent** (string): Primary intent category (e.g., "support", "feature_request", "bug_report")
- **urgency** (string): One of "low", "medium", "high", or "critical"
- **action_items** (array of strings): List of 3-5 specific, actionable steps

## How to Run

1. Ensure LlamaGate is running with access to a local LLM (e.g., Ollama with llama3.2:1b)
2. Place this extension in your LlamaGate extensions directory
3. Provide input in the format shown in `test_input.json`
4. Execute the extension via LlamaGate API, CLI, or interface
5. Review the structured output

## Configuration

### Model Settings

The extension uses:
- **Provider:** Ollama (change if using a different provider)
- **Model:** llama3.2:1b (low-resource 1B parameter model - update to match your available model)
- **Temperature:** 0.3 (lower for more consistent structured output)

### Customization

To adjust behavior:
- Modify the prompt template to change extraction criteria
- Update urgency enum values if your system uses different levels
- Adjust `max_tokens` if you need longer summaries or more action items

## What "Good Output" Looks Like

A successful output should:
- ✅ Contain all required fields (summary, intent, urgency, action_items)
- ✅ Have a concise summary that captures the request essence
- ✅ Use one of the valid urgency values
- ✅ Include 3-5 specific, actionable items
- ✅ Be valid JSON that can be parsed

### Example Good Output

```json
{
  "summary": "User reports that the login page is not loading on mobile devices. They've tried multiple browsers and cleared cache. This is blocking access to the application.",
  "intent": "bug_report",
  "urgency": "high",
  "action_items": [
    "Verify mobile browser compatibility",
    "Check server logs for mobile user agent errors",
    "Test login flow on multiple mobile devices",
    "Review recent mobile-related code changes"
  ]
}
```

## Test Files

- `test_input.json`: Sample input request text
- `expected_output.json`: Expected output structure (values may vary based on model)

## Limitations

- Output quality depends on the LLM model used
- Urgency assessment is based on text analysis only (no external context)
- Action items are suggestions and should be reviewed by humans
- JSON parsing may fail if the model doesn't strictly follow the format (consider adding output validation)
