package mcp

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type streamableHTTPMCPClient struct {
	endpoint string

	initialized atomic.Bool
	*client.Client
}

func NewStreamableHTTPMCPClient(endpoint string) Client {
	return &streamableHTTPMCPClient{
		endpoint: endpoint,
	}
}

func (c *streamableHTTPMCPClient) Initialize(ctx context.Context, request mcp.InitializeRequest, _ bool) (*mcp.InitializeResult, error) {
	if c.initialized.Load() {
		return nil, fmt.Errorf("client already initialized")
	}

	sseClient, err := client.NewStreamableHttpClient(c.endpoint)
	if err != nil {
		return nil, err
	}

	if err := sseClient.Start(context.WithoutCancel(ctx)); err != nil {
		return nil, err
	}

	result, err := sseClient.Initialize(ctx, request)
	if err != nil {
		return nil, err
	}

	c.Client = sseClient
	c.initialized.Store(true)
	return result, nil
}
