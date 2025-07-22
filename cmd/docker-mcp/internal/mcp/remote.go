package mcp

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/internal/catalog"
)

type remoteMCPClient struct {
	name   string
	config catalog.Remote

	initialized atomic.Bool
	*client.Client
}

func NewRemoteMCPClient(name string, config catalog.Remote) Client {
	return &remoteMCPClient{
		name:   name,
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

	switch c.config.Transport {
	case "sse":
		remoteClient, err = client.NewSSEMCPClient(c.config.URL)
		if err != nil {
			return nil, err
		}
	case "http":
		remoteClient, err = client.NewStreamableHttpClient(c.config.URL)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported remote transport: %s", c.config.Transport)
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
