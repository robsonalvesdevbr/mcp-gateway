package mcp

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/internal/catalog"
)

type remoteMCPClient struct {
	config      catalog.ServerConfig
	client      *mcp.Client
	session     *mcp.ClientSession
	initialized atomic.Bool
}

func NewRemoteMCPClient(config catalog.ServerConfig) Client {
	return &remoteMCPClient{
		config: config,
	}
}

func (c *remoteMCPClient) Initialize(ctx context.Context, params *mcp.InitializeParams, _ bool, _ *mcp.ServerSession) (*mcp.InitializeResult, error) {
	if c.initialized.Load() {
		return nil, fmt.Errorf("client already initialized")
	}

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

	var mcpTransport mcp.Transport
	var err error

	switch strings.ToLower(transport) {
	case "sse":
		// TODO: Need to implement custom HTTP client with headers for SSE
		mcpTransport = mcp.NewSSEClientTransport(url, &mcp.SSEClientTransportOptions{})
	case "http", "streamable", "streaming", "streamable-http":
		// TODO: Need to implement custom HTTP client with headers for streaming
		mcpTransport = mcp.NewStreamableClientTransport(url, &mcp.StreamableClientTransportOptions{})
	default:
		return nil, fmt.Errorf("unsupported remote transport: %s", transport)
	}

	c.client = mcp.NewClient(&mcp.Implementation{
		Name:    "docker-mcp-gateway",
		Version: "1.0.0",
	}, nil)

	session, err := c.client.Connect(ctx, mcpTransport)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	c.session = session
	c.initialized.Store(true)

	// The Connect method handles initialization automatically in the new SDK
	return &mcp.InitializeResult{
		ProtocolVersion: "2024-11-05",
		ServerInfo: &mcp.Implementation{
			Name:    "docker-mcp-gateway",
			Version: "1.0.0",
		},
		// Capabilities field is private and will be set by the SDK
	}, nil
}

func (c *remoteMCPClient) ListTools(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.session.ListTools(ctx, params)
}

func (c *remoteMCPClient) ListPrompts(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.session.ListPrompts(ctx, params)
}

func (c *remoteMCPClient) ListResources(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.session.ListResources(ctx, params)
}

func (c *remoteMCPClient) ListResourceTemplates(ctx context.Context, params *mcp.ListResourceTemplatesParams) (*mcp.ListResourceTemplatesResult, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.session.ListResourceTemplates(ctx, params)
}

func (c *remoteMCPClient) CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.session.CallTool(ctx, params)
}

func (c *remoteMCPClient) GetPrompt(ctx context.Context, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.session.GetPrompt(ctx, params)
}

func (c *remoteMCPClient) ReadResource(ctx context.Context, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.session.ReadResource(ctx, params)
}

func (c *remoteMCPClient) Close() error {
	if c.session != nil {
		return c.session.Close()
	}
	return nil
}

func expandEnv(value string, secrets map[string]string) string {
	return os.Expand(value, func(name string) string {
		return secrets[name]
	})
}
