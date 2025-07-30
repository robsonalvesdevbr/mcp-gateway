package mcp

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	"github.com/mark3labs/mcp-go/client"
	mcptransport "github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/internal/catalog"
)

type remoteMCPClient struct {
	config catalog.ServerConfig

	initialized atomic.Bool
	*client.Client
}

func NewRemoteMCPClient(config catalog.ServerConfig) Client {
	return &remoteMCPClient{
		config: config,
	}
}

func (c *remoteMCPClient) Initialize(ctx context.Context, request mcp.InitializeRequest, _ bool) (*mcp.InitializeResult, error) {
	if c.initialized.Load() {
		return nil, fmt.Errorf("client already initialized")
	}

	var (
		remoteClient *client.Client
		err          error
	)

	// Read configuration.
	var (
		url       string
		transport string
	)
	if c.config.Spec.SSEEndpoint != "" {
		// Deprecated
		url = c.config.Spec.SSEEndpoint
		transport = "sse"
	} else {
		url = c.config.Spec.Remote.URL
		transport = c.config.Spec.Remote.Transport
	}

	// Secrets to env
	env := map[string]string{}
	for _, secret := range c.config.Spec.Secrets {
		env[secret.Env] = c.config.Secrets[secret.Name]
	}

	// Headers
	headers := map[string]string{}
	for k, v := range c.config.Spec.Remote.Headers {
		headers[k] = expandEnv(v, env)
	}

	switch strings.ToLower(transport) {
	case "sse":
		remoteClient, err = client.NewSSEMCPClient(url, client.WithHeaders(headers))
		if err != nil {
			return nil, err
		}
	case "http", "streamable", "streaming", "streamable-http":
		remoteClient, err = client.NewStreamableHttpClient(url, mcptransport.WithHTTPHeaders(headers))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported remote transport: %s", transport)
	}

	if err := remoteClient.Start(context.WithoutCancel(ctx)); err != nil {
		return nil, err
	}

	result, err := remoteClient.Initialize(ctx, request)
	if err != nil {
		return nil, err
	}

	c.Client = remoteClient
	c.initialized.Store(true)
	return result, nil
}

func expandEnv(value string, secrets map[string]string) string {
	return os.Expand(value, func(name string) string {
		return secrets[name]
	})
}
