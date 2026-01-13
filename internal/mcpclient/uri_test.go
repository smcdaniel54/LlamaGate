package mcpclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMCPURI(t *testing.T) {
	tests := []struct {
		name      string
		uriString string
		want      *MCPURI
		wantErr   bool
	}{
		{
			name:      "valid filesystem URI",
			uriString: "mcp://filesystem/file:///path/to/file.txt",
			want: &MCPURI{
				Server:   "filesystem",
				Resource: "file:///path/to/file.txt",
			},
			wantErr: false,
		},
		{
			name:      "valid simple resource",
			uriString: "mcp://server/resource/name",
			want: &MCPURI{
				Server:   "server",
				Resource: "resource/name",
			},
			wantErr: false,
		},
		{
			name:      "valid URI with query",
			uriString: "mcp://server/resource?param=value",
			want: &MCPURI{
				Server:   "server",
				Resource: "resource?param=value",
			},
			wantErr: false,
		},
		{
			name:      "valid URI with fragment",
			uriString: "mcp://server/resource#section",
			want: &MCPURI{
				Server:   "server",
				Resource: "resource#section",
			},
			wantErr: false,
		},
		{
			name:      "invalid scheme",
			uriString: "http://server/resource",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "missing server",
			uriString: "mcp:///resource",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "missing resource",
			uriString: "mcp://server/",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "not an MCP URI",
			uriString: "file:///path/to/file.txt",
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMCPURI(tt.uriString)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
				// Test String() method
				assert.Equal(t, tt.uriString, got.String())
			}
		})
	}
}

func TestExtractMCPURIs(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    []*MCPURI
		wantLen int
	}{
		{
			name: "single URI",
			text: "Please read mcp://filesystem/file:///test.txt",
			want: []*MCPURI{
				{Server: "filesystem", Resource: "file:///test.txt"},
			},
			wantLen: 1,
		},
		{
			name: "multiple URIs",
			text: "Read mcp://filesystem/file1.txt and mcp://filesystem/file2.txt",
			want: []*MCPURI{
				{Server: "filesystem", Resource: "file1.txt"},
				{Server: "filesystem", Resource: "file2.txt"},
			},
			wantLen: 2,
		},
		{
			name: "URI with newline",
			text: "Read mcp://filesystem/file.txt\nand process it",
			want: []*MCPURI{
				{Server: "filesystem", Resource: "file.txt"},
			},
			wantLen: 1,
		},
		{
			name:    "no URIs",
			text:    "This is just regular text",
			want:    []*MCPURI{},
			wantLen: 0,
		},
		{
			name: "invalid URI ignored",
			text: "Read mcp://filesystem/file.txt and http://example.com",
			want: []*MCPURI{
				{Server: "filesystem", Resource: "file.txt"},
			},
			wantLen: 1,
		},
		{
			name: "URI in sentence",
			text: "Please summarize the content from mcp://server/resource/uri and provide insights.",
			want: []*MCPURI{
				{Server: "server", Resource: "resource/uri"},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractMCPURIs(tt.text)
			assert.Len(t, got, tt.wantLen)
			if tt.wantLen > 0 {
				for i, uri := range got {
					assert.Equal(t, tt.want[i].Server, uri.Server)
					assert.Equal(t, tt.want[i].Resource, uri.Resource)
				}
			}
		})
	}
}

func TestMCPURI_String(t *testing.T) {
	uri := &MCPURI{
		Server:   "filesystem",
		Resource: "file:///test.txt",
	}
	assert.Equal(t, "mcp://filesystem/file:///test.txt", uri.String())
}
