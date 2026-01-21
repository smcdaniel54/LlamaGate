package proxy

import (
	"context"

	"github.com/llamagate/llamagate/internal/mcpclient"
)

// ServerManagerInterface defines the interface for MCP server management
// This allows the proxy to use ServerManager without knowing the concrete type
type ServerManagerInterface interface {
	// GetServer gets server information by name
	GetServer(name string) (*mcpclient.ManagedServer, error)

	// GetClient gets a client for a server (with pooling for HTTP)
	GetClient(ctx context.Context, name string) (*mcpclient.Client, error)

	// ReleaseClient releases a pooled connection (no-op for non-pooled transports)
	ReleaseClient(name string, client *mcpclient.Client)
}
