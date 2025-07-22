package mcp

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type sseMCPClient struct {
	name     string
	endpoint string

	initialized atomic.Bool
	*client.Client
}

func NewSSEClient(name string, endpoint string) Client {
	return &sseMCPClient{
		name:     name,
		endpoint: endpoint,
	}
}

func (c *sseMCPClient) Initialize(ctx context.Context, request mcp.InitializeRequest, _ bool) (*mcp.InitializeResult, error) {
	if c.initialized.Load() {
		return nil, fmt.Errorf("client already initialized")
	}

	sseClient, err := client.NewSSEMCPClient(c.endpoint)
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
