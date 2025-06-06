package gateway

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/docker/cli/cli/command"
	"github.com/mark3labs/mcp-go/server"

	"github.com/docker/mcp-cli/cmd/docker-mcp/internal/docker"
)

type Config struct {
	Options
	ServerNames  []string
	CatalogPath  string
	ConfigPath   string
	RegistryPath string
	SecretsPath  string
}

type Options struct {
	Port             int
	Transport        string
	ToolNames        []string
	Verbose          bool
	KeepContainers   bool
	LogCalls         bool
	BlockSecrets     bool
	VerifySignatures bool
	DryRun           bool
	Watch            bool
}

type Gateway struct {
	Options
	dockerClient *docker.Client
	configurator Configurator
}

func NewGateway(config Config, dockerCli command.Cli) *Gateway {
	return &Gateway{
		Options:      config.Options,
		dockerClient: docker.NewClient(dockerCli),
		configurator: &FileBasedConfiguration{
			ServerNames:  config.ServerNames,
			CatalogPath:  config.CatalogPath,
			RegistryPath: config.RegistryPath,
			ConfigPath:   config.ConfigPath,
			SecretsPath:  config.SecretsPath,
			Watch:        config.Watch,
			DockerClient: dockerCli.Client(),
		},
	}
}

//nolint:gocyclo
func (g *Gateway) Run(ctx context.Context) error {
	start := time.Now()

	// Listen as early as possible to not lose client connections.
	var ln net.Listener
	if port := g.Port; port != 0 {
		var (
			lc  net.ListenConfig
			err error
		)
		ln, err = lc.Listen(ctx, "tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return err
		}
	}

	// Read the configuration.
	configuration, configurationUpdates, stopConfigWatcher, err := g.configurator.Read(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = stopConfigWatcher() }()

	// Which servers are enabled in the registry.yaml?
	serverNames := configuration.ServerNames()
	if len(serverNames) == 0 {
		log("- No server is enabled")
	} else {
		log("- Those servers are enabled:", strings.Join(serverNames, ", "))
	}

	// Which docker images are used?
	// Pull them and verify them if possible.
	if err := g.pullAndVerify(ctx, configuration); err != nil {
		return err
	}

	// List all the available tools.
	startList := time.Now()
	log("- Listing MCP tools...")
	capabilities, err := g.listCapabilities(ctx, configuration, serverNames)
	if err != nil {
		return fmt.Errorf("listing resources: %w", err)
	}
	log(">", len(capabilities.Tools), "tools listed in", time.Since(startList))

	toolCallbacks := callbacks(g.LogCalls, g.BlockSecrets)

	// TODO: cleanup stopped servers. That happens in stdio over TCP mode.
	var (
		lock            sync.Mutex
		changeListeners []func(*Capabilities)
	)

	newMCPServer := func() *server.MCPServer {
		mcpServer := server.NewMCPServer(
			"Docker AI MCP Gateway",
			"2.0.1",
			server.WithToolHandlerMiddleware(toolCallbacks),
		)

		current := capabilities
		mcpServer.AddTools(current.Tools...)
		mcpServer.AddPrompts(current.Prompts...)
		mcpServer.AddResources(current.Resources...)
		for _, v := range current.ResourceTemplates {
			mcpServer.AddResourceTemplate(v.ResourceTemplate, v.Handler)
		}

		lock.Lock()
		changeListeners = append(changeListeners, func(newConfig *Capabilities) {
			mcpServer.DeleteTools(toolNames(current.Tools)...)
			mcpServer.AddTools(newConfig.Tools...)

			// TODO: sync other things than tools

			current = newConfig
		})
		lock.Unlock()

		return mcpServer
	}

	// Optionally watch for configuration updates.
	if configurationUpdates != nil {
		log("- Watching for configuration updates...")
		go func() {
			for {
				select {
				case <-ctx.Done():
					log("> Stop watching for updates")
					return
				case configuration := <-configurationUpdates:
					log("> Configuration updated, reloading...")

					if err := g.pullAndVerify(ctx, configuration); err != nil {
						logf("> Unable to pull and verify images: %s", err)
						continue
					}

					capabilities, err := g.listCapabilities(ctx, configuration, configuration.ServerNames())
					if err != nil {
						logf("> Unable to list capabilities: %s", err)
						continue
					}

					lock.Lock()
					for _, listener := range changeListeners {
						listener(capabilities)
					}
					lock.Unlock()
				}
			}
		}()
	}

	log("> Initialized in", time.Since(start))
	if g.DryRun {
		log("Dry run mode enabled, not starting the server.")
		return nil
	}

	// Start the server
	switch {
	case strings.EqualFold(g.Transport, "sse") && g.Port > 0:
		log("> Starting SSE server on port", g.Port)

		return startSseServer(ctx, newMCPServer, ln)
	case strings.EqualFold(g.Transport, "stdio") && g.Port == 0:
		log("> Starting STDIO server")

		return startStdioServer(ctx, newMCPServer, os.Stdin, os.Stdout)
	case strings.EqualFold(g.Transport, "stdio") && g.Port > 0:
		log("> Starting STDIO over TCP server on port", g.Port)

		return startStdioOverTCPServer(ctx, newMCPServer, ln)
	default:
		return fmt.Errorf("unknown transport %q, expected 'stdio' or 'sse'", g.Transport)
	}
}

func toolNames(tools []server.ServerTool) []string {
	var names []string
	for _, tool := range tools {
		names = append(names, tool.Tool.Name)
	}
	return names
}
