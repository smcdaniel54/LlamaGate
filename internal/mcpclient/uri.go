package mcpclient

import (
	"fmt"
	"net/url"
	"strings"
)

// MCPURI represents a parsed MCP URI
// Format: mcp://server/resource/uri
// Example: mcp://filesystem/file:///path/to/file.txt
type MCPURI struct {
	Server   string // MCP server name
	Resource string // Resource URI (may contain slashes)
}

// ParseMCPURI parses an MCP URI string
// Format: mcp://server/resource/uri
// Returns nil if the URI is not a valid MCP URI
func ParseMCPURI(uriString string) (*MCPURI, error) {
	if !strings.HasPrefix(uriString, "mcp://") {
		return nil, fmt.Errorf("not an MCP URI: %s", uriString)
	}

	// Parse as URL
	parsed, err := url.Parse(uriString)
	if err != nil {
		return nil, fmt.Errorf("invalid URI format: %w", err)
	}

	if parsed.Scheme != "mcp" {
		return nil, fmt.Errorf("invalid scheme: expected 'mcp', got '%s'", parsed.Scheme)
	}

	// Extract server name (host)
	server := parsed.Host
	if server == "" {
		return nil, fmt.Errorf("missing server name in URI")
	}

	// Extract resource URI (path, removing leading slash)
	resource := strings.TrimPrefix(parsed.Path, "/")
	if resource == "" {
		return nil, fmt.Errorf("missing resource URI in MCP URI")
	}

	// Include query and fragment if present
	if parsed.RawQuery != "" {
		resource += "?" + parsed.RawQuery
	}
	if parsed.Fragment != "" {
		resource += "#" + parsed.Fragment
	}

	return &MCPURI{
		Server:   server,
		Resource: resource,
	}, nil
}

// String returns the string representation of the MCP URI
func (u *MCPURI) String() string {
	return fmt.Sprintf("mcp://%s/%s", u.Server, u.Resource)
}

// ExtractMCPURIs extracts all MCP URIs from a text string
// Returns a list of parsed MCP URIs found in the text
func ExtractMCPURIs(text string) []*MCPURI {
	var uris []*MCPURI

	// Find all occurrences of mcp://
	start := 0
	for {
		idx := strings.Index(text[start:], "mcp://")
		if idx == -1 {
			break
		}

		actualIdx := start + idx

		// Find the end of the URI (whitespace, newline, or end of string)
		end := actualIdx + 6 // "mcp://"
		for end < len(text) {
			char := text[end]
			if char == ' ' || char == '\n' || char == '\r' || char == '\t' {
				break
			}
			end++
		}

		// Parse the URI
		uriString := text[actualIdx:end]
		if parsed, err := ParseMCPURI(uriString); err == nil {
			uris = append(uris, parsed)
		}

		start = end
	}

	return uris
}
