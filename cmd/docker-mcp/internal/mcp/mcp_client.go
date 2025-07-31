package mcp

import (
	"context"
	"fmt"
	"slices"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Client interface wraps the official MCP SDK client with our legacy interface
type Client interface {
	Initialize(ctx context.Context, params *mcp.InitializeParams, debug bool, serverSession *mcp.ServerSession) (*mcp.InitializeResult, error)
	ListTools(ctx context.Context, params *mcp.ListToolsParams) (*mcp.ListToolsResult, error)
	ListPrompts(ctx context.Context, params *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error)
	ListResources(ctx context.Context, params *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error)
	ListResourceTemplates(ctx context.Context, params *mcp.ListResourceTemplatesParams) (*mcp.ListResourceTemplatesResult, error)
	CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error)
	GetPrompt(ctx context.Context, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error)
	ReadResource(ctx context.Context, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error)
	Close() error
}

func allSessions(server *mcp.Server, callback func(session *mcp.ServerSession)) {
	for session := range server.Sessions() {
		callback(session)
	}
}

func serverNotifications(server *mcp.Server) *mcp.ClientOptions {
	return &mcp.ClientOptions{
		CreateMessageHandler: func(ctx context.Context, session *mcp.ClientSession, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
			// Handle create messages if needed
			return nil, fmt.Errorf("create messages not supported")
		},
		ToolListChangedHandler: func(ctx context.Context, session *mcp.ClientSession, params *mcp.ToolListChangedParams) {
			// Handle tool list changes if needed
			if server != nil {
				sessions := slices.Collect(server.Sessions())
				mcp.NotifySessions(sessions, "notifications/tools/list_changed", params)
			}
		},
		ResourceListChangedHandler: func(ctx context.Context, session *mcp.ClientSession, params *mcp.ResourceListChangedParams) {
			if server != nil {
				sessions := slices.Collect(server.Sessions())
				mcp.NotifySessions(sessions, "notifications/resources/list_changed", params)
			}
		},
		PromptListChangedHandler: func(ctx context.Context, session *mcp.ClientSession, params *mcp.PromptListChangedParams) {
			if server != nil {
				sessions := slices.Collect(server.Sessions())
				mcp.NotifySessions(sessions, "notifications/prompts/list_changed", params)
			}
		},
		ProgressNotificationHandler: func(ctx context.Context, session *mcp.ClientSession, params *mcp.ProgressNotificationParams) {
			allSessions(server, func (session *mcp.ServerSession) {session.NotifyProgress(ctx, params)})
		},
		LoggingMessageHandler: func(ctx context.Context, session *mcp.ClientSession, params *mcp.LoggingMessageParams) {
			allSessions(server, func (session *mcp.ServerSession) {session.Log(ctx, params)})
		},
	}
}

func stdioNotifications(serverSession *mcp.ServerSession) *mcp.ClientOptions {
	return &mcp.ClientOptions{
		CreateMessageHandler: func(ctx context.Context, session *mcp.ClientSession, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
			// Handle create messages if needed
			return nil, fmt.Errorf("create messages not supported")
		},
		ToolListChangedHandler: func(ctx context.Context, session *mcp.ClientSession, params *mcp.ToolListChangedParams) {
			mcp.HandleNotify(ctx, serverSession, "notifications/tools/list_changed", params)
		},
		ResourceListChangedHandler: func(ctx context.Context, session *mcp.ClientSession, params *mcp.ResourceListChangedParams) {
			mcp.HandleNotify(ctx, serverSession, "notifications/resources/list_changed", params)
		},
		PromptListChangedHandler: func(ctx context.Context, session *mcp.ClientSession, params *mcp.PromptListChangedParams) {
			mcp.HandleNotify(ctx, serverSession, "notifications/prompts/list_changed", params)
		},
		ProgressNotificationHandler: func(ctx context.Context, session *mcp.ClientSession, params *mcp.ProgressNotificationParams) {
			serverSession.NotifyProgress(ctx, params)
		},
		LoggingMessageHandler: func(ctx context.Context, session *mcp.ClientSession, params *mcp.LoggingMessageParams) {
			serverSession.Log(ctx, params)
		},
	}
}

