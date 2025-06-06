package gateway

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/sync/errgroup"
)

//nolint:gocyclo
func (g *Gateway) listEverything(ctx context.Context, configuration Configuration, serverNames []string) (*Everything, error) {
	var (
		lock                    sync.Mutex
		serverTools             []server.ServerTool
		serverPrompts           []server.ServerPrompt
		serverResources         []server.ServerResource
		serverResourceTemplates []ServerResourceTemplate
	)

	errs, ctx := errgroup.WithContext(ctx)
	errs.SetLimit(runtime.NumCPU())
	for _, serverName := range serverNames {
		serverConfig, toolGroup, found := configuration.Find(serverName)

		switch {
		case !found:
			log("  - MCP server not found:", serverName)

		// It's an MCP Server
		case serverConfig != nil:
			errs.Go(func() error {
				client, err := g.startMCPClient(ctx, *serverConfig, &readOnly)
				if err != nil {
					logf("  > Can't start %s: %s", serverConfig.Name, err)
					return nil
				}
				defer client.Close()

				var log string

				tools, err := client.ListTools(ctx, mcp.ListToolsRequest{})
				if err != nil {
					logf("  > Can't list tools %s: %s", serverConfig.Name, err)
				} else if len(tools.Tools) > 0 {
					log += fmt.Sprintf(" (%d tools)", len(tools.Tools))
					for _, tool := range tools.Tools {
						if !isToolEnabled(serverConfig.Name, serverConfig.Spec.Image, tool.Name, g.ToolNames) {
							continue
						}

						lock.Lock()
						serverTools = append(serverTools, server.ServerTool{
							Tool:    tool,
							Handler: g.mcpServerToolHandler(*serverConfig, tool.Annotations),
						})
						lock.Unlock()
					}
				}

				prompts, err := client.ListPrompts(ctx, mcp.ListPromptsRequest{})
				if err == nil && len(prompts.Prompts) > 0 {
					log += fmt.Sprintf(" (%d prompts)", len(prompts.Prompts))
					lock.Lock()
					for _, prompt := range prompts.Prompts {
						serverPrompts = append(serverPrompts, server.ServerPrompt{
							Prompt:  prompt,
							Handler: g.mcpServerPromptHandler(*serverConfig),
						})
					}
					lock.Unlock()
				}

				resources, err := client.ListResources(ctx, mcp.ListResourcesRequest{})
				if err == nil && len(resources.Resources) > 0 {
					log += fmt.Sprintf(" (%d resources)", len(resources.Resources))
					lock.Lock()
					for _, resource := range resources.Resources {
						serverResources = append(serverResources, server.ServerResource{
							Resource: resource,
							Handler:  g.mcpServerResourceHandler(*serverConfig),
						})
					}
					lock.Unlock()
				}

				resourceTemplates, err := client.ListResourceTemplates(ctx, mcp.ListResourceTemplatesRequest{})
				if err == nil && len(resourceTemplates.ResourceTemplates) > 0 {
					log += fmt.Sprintf(" (%d resourceTemplates)", len(resourceTemplates.ResourceTemplates))
					lock.Lock()
					for _, resourceTemplate := range resourceTemplates.ResourceTemplates {
						serverResourceTemplates = append(serverResourceTemplates, ServerResourceTemplate{
							ResourceTemplate: resourceTemplate,
							Handler:          g.mcpServerResourceTemplateHandler(*serverConfig),
						})
					}
					lock.Unlock()
				}

				if log != "" {
					logf("  > %s:%s", serverConfig.Name, log)
				}

				return nil
			})

		// It's a POCI
		case toolGroup != nil:
			for _, tool := range *toolGroup {
				if !isToolEnabled(serverName, "", tool.Name, g.ToolNames) {
					continue
				}

				mcpTool := mcp.Tool{
					Name:        tool.Name,
					Description: tool.Description,
				}
				if len(tool.Parameters.Properties) == 0 {
					mcpTool.InputSchema.Type = "object"
				} else {
					mcpTool.InputSchema.Type = tool.Parameters.Type
					mcpTool.InputSchema.Properties = tool.Parameters.Properties.ToMap()
					mcpTool.InputSchema.Required = tool.Parameters.Required
				}

				serverTool := server.ServerTool{
					Tool:    mcpTool,
					Handler: g.mcpToolHandler(tool),
				}

				lock.Lock()
				serverTools = append(serverTools, serverTool)
				lock.Unlock()
			}
		}
	}

	if err := errs.Wait(); err != nil {
		return nil, err
	}

	return &Everything{
		Tools:             serverTools,
		Prompts:           serverPrompts,
		Resources:         serverResources,
		ResourceTemplates: serverResourceTemplates,
	}, nil
}

func isToolEnabled(serverName, serverImage, toolName string, enabledTools []string) bool {
	if len(enabledTools) == 0 {
		return true
	}

	for _, enabled := range enabledTools {
		if enabled == "*" ||
			strings.EqualFold(enabled, toolName) ||
			strings.EqualFold(enabled, serverName+":"+toolName) ||
			strings.EqualFold(enabled, serverName+":*") ||
			strings.EqualFold(enabled, "*") {
			return true
		}
	}

	if serverImage != "" {
		for _, enabled := range enabledTools {
			if strings.EqualFold(enabled, serverImage+":"+toolName) ||
				strings.EqualFold(enabled, serverImage+":*") {
				return true
			}
		}
	}

	return false
}
