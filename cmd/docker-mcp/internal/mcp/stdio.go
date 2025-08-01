package mcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync/atomic"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/internal/logs"
)

type stdioMCPClient struct {
	name        string
	command     string
	env         []string
	args        []string
	client      *mcp.Client
	session     *mcp.ClientSession
	initialized atomic.Bool
}

func NewStdioCmdClient(name string, command string, env []string, args ...string) Client {
	return &stdioMCPClient{
		name:    name,
		command: command,
		env:     env,
		args:    args,
	}
}

func (c *stdioMCPClient) Initialize(ctx context.Context, params *mcp.InitializeParams, debug bool, s *mcp.ServerSession) (*mcp.InitializeResult, error) {
	if c.initialized.Load() {
		return nil, fmt.Errorf("client already initialized")
	}

	cmd := exec.CommandContext(ctx, c.command, c.args...)
	cmd.Env = c.env

	if debug {
		cmd.Stderr = logs.NewPrefixer(os.Stderr, "- "+c.name+": ")
	}

	transport := mcp.NewCommandTransport(cmd)
	c.client = mcp.NewClient(&mcp.Implementation{
		Name:    "docker-mcp-gateway",
		Version: "1.0.0",
	}, stdioNotifications(s))

	session, err := c.client.Connect(ctx, transport)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	c.session = session
	c.initialized.Store(true)

	// The Connect method handles initialization automatically in the new SDK
	// We just return a basic result structure
	return &mcp.InitializeResult{
		ProtocolVersion: "2024-11-05",
		ServerInfo: &mcp.Implementation{
			Name:    "docker-mcp-gateway",
			Version: "1.0.0",
		},
		// Capabilities field is private and will be set by the SDK
	}, nil
}

func (c *stdioMCPClient) ListTools(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.session.ListTools(ctx, params)
}

func (c *stdioMCPClient) ListPrompts(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.session.ListPrompts(ctx, params)
}

func (c *stdioMCPClient) ListResources(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.session.ListResources(ctx, params)
}

func (c *stdioMCPClient) ListResourceTemplates(ctx context.Context, params *mcp.ListResourceTemplatesParams) (*mcp.ListResourceTemplatesResult, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.session.ListResourceTemplates(ctx, params)
}

func (c *stdioMCPClient) CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.session.CallTool(ctx, params)
}

func (c *stdioMCPClient) GetPrompt(ctx context.Context, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.session.GetPrompt(ctx, params)
}

func (c *stdioMCPClient) ReadResource(ctx context.Context, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	if !c.initialized.Load() {
		return nil, fmt.Errorf("client not initialized")
	}
	return c.session.ReadResource(ctx, params)
}

func (c *stdioMCPClient) Close() error {
	if c.session != nil {
		return c.session.Close()
	}
	return nil
}
