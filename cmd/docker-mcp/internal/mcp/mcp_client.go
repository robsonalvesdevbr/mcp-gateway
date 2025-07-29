package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Client interface wraps the official MCP SDK client with our legacy interface
type Client interface {
	Initialize(ctx context.Context, params *mcp.InitializeParams, debug bool) (*mcp.InitializeResult, error)
	ListTools(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error)
	ListPrompts(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error)
	ListResources(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error)
	ListResourceTemplates(ctx context.Context, params *mcp.ListResourceTemplatesParams) (*mcp.ListResourceTemplatesResult, error)
	CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error)
	GetPrompt(ctx context.Context, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error)
	ReadResource(ctx context.Context, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error)
	Close() error
}
