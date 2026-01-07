package proxy

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/llamagate/llamagate/internal/mcpclient"
	"github.com/rs/zerolog/log"
)

// injectMCPResourceContext processes messages and injects MCP resource content
// where mcp:// URIs are found
func (p *Proxy) injectMCPResourceContext(ctx context.Context, requestID string, messages []Message) ([]Message, error) {
	if p.serverManager == nil || p.toolManager == nil {
		// No MCP support, return messages as-is
		return messages, nil
	}

	var enhancedMessages []Message
	var resourceContexts []string

	// Process each message
	for _, msg := range messages {
		// Extract content as string
		contentStr := ""
		switch v := msg.Content.(type) {
		case string:
			contentStr = v
		case []interface{}:
			// Handle array content (OpenAI format)
			var parts []string
			for _, part := range v {
				if partMap, ok := part.(map[string]interface{}); ok {
					if text, ok := partMap["text"].(string); ok {
						parts = append(parts, text)
					}
				}
			}
			contentStr = strings.Join(parts, " ")
		}

		// Extract MCP URIs from content
		mcpURIs := mcpclient.ExtractMCPURIs(contentStr)

		if len(mcpURIs) > 0 {
			// Fetch resources for each URI
			for _, mcpURI := range mcpURIs {
				resourceContent, err := p.fetchMCPResource(ctx, requestID, mcpURI)
				if err != nil {
					log.Warn().
						Str("request_id", requestID).
						Str("uri", mcpURI.String()).
						Err(err).
						Msg("Failed to fetch MCP resource, continuing without it")
					// Continue processing other URIs even if one fails
					continue
				}

				if resourceContent != "" {
					resourceContexts = append(resourceContexts, fmt.Sprintf("Resource from %s:\n%s", mcpURI.String(), resourceContent))
				}
			}
		}

		enhancedMessages = append(enhancedMessages, msg)
	}

	// Prepend resource contexts as a system message if any were found
	if len(resourceContexts) > 0 {
		systemMsg := Message{
			Role:    "system",
			Content: strings.Join(resourceContexts, "\n\n"),
		}
		enhancedMessages = append([]Message{systemMsg}, enhancedMessages...)
	}

	return enhancedMessages, nil
}

// fetchMCPResource fetches a resource from an MCP server
func (p *Proxy) fetchMCPResource(ctx context.Context, requestID string, mcpURI *mcpclient.MCPURI) (string, error) {
	// Get server from server manager
	serverInfo, err := p.serverManager.GetServer(mcpURI.Server)
	if err != nil {
		return "", fmt.Errorf("server not found: %w", err)
	}

	client := serverInfo.Client
	if client == nil {
		return "", fmt.Errorf("client not available for server %s", mcpURI.Server)
	}

	// Read resource with timeout
	timeout := p.resourceFetchTimeout
	if timeout == 0 {
		timeout = 30 * time.Second // Default fallback
	}
	resourceCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := client.ReadResource(resourceCtx, mcpURI.Resource)
	if err != nil {
		return "", fmt.Errorf("failed to read resource: %w", err)
	}

	// Extract text content from resource
	var content strings.Builder
	for _, contentItem := range result.Contents {
		if contentItem.MimeType == "text/plain" || contentItem.Text != "" {
			if contentItem.Text != "" {
				content.WriteString(contentItem.Text)
				content.WriteString("\n")
			}
		}
	}

	log.Info().
		Str("request_id", requestID).
		Str("uri", mcpURI.String()).
		Int("content_length", content.Len()).
		Msg("Fetched MCP resource for context injection")

	return content.String(), nil
}
