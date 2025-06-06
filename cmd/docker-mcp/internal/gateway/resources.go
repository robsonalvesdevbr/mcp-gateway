package gateway

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Everything struct {
	Tools             []server.ServerTool
	Prompts           []server.ServerPrompt
	Resources         []server.ServerResource
	ResourceTemplates []ServerResourceTemplate
}

type ServerResourceTemplate struct {
	ResourceTemplate mcp.ResourceTemplate
	Handler          server.ResourceTemplateHandlerFunc
}
