# GitHub Webhook Receiver Extension

This extension receives GitHub webhooks via a custom HTTP endpoint and processes them with LLM analysis. It demonstrates the dynamic endpoints feature in LlamaGate.

## What It Does

Creates a custom HTTP endpoint (`POST /v1/extensions/github-webhook-receiver/webhooks/github`) that:
- Receives GitHub webhook events (pull requests, issues, pushes, etc.)
- Analyzes events using LLM (Mistral model)
- Stores events for audit/history
- Generates friendly acknowledgment responses

Useful for:
- Automated GitHub event processing
- Event logging and audit trails
- Integration with GitHub workflows
- Real-time notifications and analysis

## Inputs

The extension receives inputs from the webhook endpoint:

- **webhook_payload** (object, required): Full GitHub webhook payload
- **github_signature** (string, optional): X-Hub-Signature-256 header for verification
- **event_type** (string, optional): X-GitHub-Event header (pull_request, issues, push, etc.)

## Outputs

- **processed_event** (object): Processed event with LLM analysis summary
- **response_message** (string): Human-readable acknowledgment message
- **event_stored** (boolean): Whether event was stored successfully
- **event_id** (string): Unique identifier for stored event

## How to Run

1. **Ensure LlamaGate is running** with dynamic endpoints support (v0.9.2+)
2. **Ensure Ollama is running** with the `mistral` model available
3. **Install the extension**:
   ```bash
   cp -r examples/extensions/github-webhook-receiver /path/to/LlamaGate/extensions/
   ```
4. **Restart LlamaGate** to register the endpoint
5. **Configure GitHub webhook**:
   - Go to your repository Settings → Webhooks → Add webhook
   - **Payload URL**: `https://your-server.com/v1/extensions/github-webhook-receiver/webhooks/github`
   - **Content type**: `application/json`
   - **Events**: Select events to receive (or "Send me everything")
6. **Test the webhook**:
   ```bash
   curl -X POST http://localhost:11435/v1/extensions/github-webhook-receiver/webhooks/github \
     -H "Content-Type: application/json" \
     -H "X-GitHub-Event: pull_request" \
     -d @test_input.json
   ```

## Configuration

### Model Settings

The extension uses:
- **Provider**: Ollama (via `llm.chat`)
- **Model**: mistral (LlamaGate standard model for testing)
- **Temperature**: 0.3 (for analysis), 0.5 (for response generation)

### Endpoint Configuration

- **Path**: `/webhooks/github`
- **Method**: POST
- **Full URL**: `/v1/extensions/github-webhook-receiver/webhooks/github`
- **Auth**: Disabled (GitHub uses signature verification)
- **Rate Limiting**: Enabled

### Customization

To adjust behavior:
- Modify the LLM prompts in the workflow steps
- Change the event storage path in Step 2
- Adjust temperature settings for different response styles
- Update guardrails if needed

## What "Good Output" Looks Like

A successful webhook processing should:
- ✅ Return HTTP 200 status
- ✅ Include `success: true` in response
- ✅ Contain `processed_event` with LLM analysis summary
- ✅ Include `response_message` with acknowledgment
- ✅ Show `event_stored: true` if storage succeeded
- ✅ Provide `event_id` for tracking

### Example Good Output

```json
{
  "success": true,
  "data": {
    "processed_event": {
      "summary": "Pull request #123 opened by testuser in owner/repo",
      "action": "opened",
      "repository": "owner/repo"
    },
    "response_message": "Received and processed pull request event. Event stored and analyzed successfully.",
    "event_stored": true,
    "event_id": "owner-repo-2026-01-22T10-30-45-opened"
  }
}
```

## Test Files

- `test_input.json`: Sample GitHub webhook payload for testing
- `expected_output.json`: Expected response structure (values may vary)
- `test_webhook.sh`: Test script for manual testing (requires `jq`)

## Limitations

- Requires LlamaGate with dynamic endpoints feature (v0.9.2+)
- Event storage uses local filesystem (`./var/github-events/`)
- Network access guardrail is disabled (required for webhook receiving)
- Signature verification not implemented (can be added)

## Demonstrates

- **Dynamic endpoints**: Custom HTTP endpoint definition
- **Webhook integration**: Real-world external service integration
- **LLM processing workflow**: Multi-step LLM analysis
- **File storage pattern**: Event persistence for audit trails
- **Workflow extensions**: Complex multi-step processing
