package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"

	"github.com/docker/mcp-gateway/pkg/catalog"
	"github.com/docker/mcp-gateway/pkg/oci"
	"github.com/docker/mcp-gateway/pkg/telemetry"
)

// mcpFindTool implements a tool for finding MCP servers in the catalog
func (g *Gateway) createMcpFindTool(configuration Configuration) *ToolRegistration {
	tool := &mcp.Tool{
		Name:        "mcp-find",
		Description: "Find MCP servers in the current catalog by name or description. Returns matching servers with their details.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"query": {
					Type:        "string",
					Description: "Search query to find servers by name or description (case-insensitive)",
				},
				"limit": {
					Type:        "integer",
					Description: "Maximum number of results to return (default: 10)",
				},
			},
			Required: []string{"query"},
		},
	}

	handler := func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse parameters
		var params struct {
			Query string `json:"query"`
			Limit int    `json:"limit"`
		}

		if req.Params.Arguments == nil {
			return nil, fmt.Errorf("missing arguments")
		}

		paramsBytes, err := json.Marshal(req.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal arguments: %w", err)
		}

		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return nil, fmt.Errorf("failed to parse arguments: %w", err)
		}

		if params.Query == "" {
			return nil, fmt.Errorf("query parameter is required")
		}

		if params.Limit <= 0 {
			params.Limit = 10
		}

		// Search through the catalog servers
		query := strings.ToLower(strings.TrimSpace(params.Query))
		var matches []ServerMatch

		for serverName, server := range configuration.servers {
			match := false
			score := 0

			// Check server name (exact match gets higher score)
			serverNameLower := strings.ToLower(serverName)
			if serverNameLower == query {
				match = true
				score = 100
			} else if strings.Contains(serverNameLower, query) {
				match = true
				score = 50
			}

			// Check server description
			if server.Description != "" {
				descriptionLower := strings.ToLower(server.Description)
				if descriptionLower == query {
					match = true
					score = maxInt(score, 95)
				} else if strings.Contains(descriptionLower, query) {
					match = true
					score = maxInt(score, 45)
				}
			}

			// Check if it has tools that might match
			for _, tool := range server.Tools {
				toolNameLower := strings.ToLower(tool.Name)
				toolDescLower := strings.ToLower(tool.Description)

				if toolNameLower == query {
					match = true
					score = maxInt(score, 90)
				} else if strings.Contains(toolNameLower, query) {
					match = true
					score = maxInt(score, 40)
				} else if strings.Contains(toolDescLower, query) {
					match = true
					score = maxInt(score, 30)
				}
			}

			// Check image name
			if server.Image != "" {
				imageLower := strings.ToLower(server.Image)
				if strings.Contains(imageLower, query) {
					match = true
					score = maxInt(score, 20)
				}
			}

			if match {
				matches = append(matches, ServerMatch{
					Name:   serverName,
					Server: server,
					Score:  score,
				})
			}
		}

		// Sort matches by score (higher scores first)
		for i := range len(matches) - 1 {
			for j := i + 1; j < len(matches); j++ {
				if matches[i].Score < matches[j].Score {
					matches[i], matches[j] = matches[j], matches[i]
				}
			}
		}

		// Limit results
		if len(matches) > params.Limit {
			matches = matches[:params.Limit]
		}

		// Format results
		var results []map[string]any
		for _, match := range matches {
			serverInfo := map[string]any{
				"name": match.Name,
			}

			if match.Server.Description != "" {
				serverInfo["description"] = match.Server.Description
			}

			if len(match.Server.Secrets) > 0 {
				var secrets []string
				for _, secret := range match.Server.Secrets {
					secrets = append(secrets, secret.Name)
				}
				serverInfo["required_secrets"] = secrets
			}

			if len(match.Server.Config) > 0 {
				serverInfo["config_schema"] = match.Server.Config
			}

			serverInfo["long_lived"] = match.Server.LongLived

			results = append(results, serverInfo)
		}

		response := map[string]any{
			"query":         params.Query,
			"total_matches": len(results),
			"servers":       results,
		}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(responseBytes)}},
		}, nil
	}

	return &ToolRegistration{
		Tool:    tool,
		Handler: withToolTelemetry("mcp-find", handler),
	}
}

// ServerMatch represents a search result
type ServerMatch struct {
	Name   string
	Server catalog.Server
	Score  int
}

// mcpAddTool implements a tool for adding new servers to the registry
func (g *Gateway) createMcpAddTool(configuration Configuration, clientConfig *clientConfig) *ToolRegistration {
	tool := &mcp.Tool{
		Name:        "mcp-add",
		Description: "Add a new MCP server to the session. The server must exist in the catalog.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"name": {
					Type:        "string",
					Description: "Name of the MCP server to add to the registry (must exist in catalog)",
				},
			},
			Required: []string{"name"},
		},
	}

	handler := func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse parameters
		var params struct {
			Name string `json:"name"`
		}

		if req.Params.Arguments == nil {
			return nil, fmt.Errorf("missing arguments")
		}

		paramsBytes, err := json.Marshal(req.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal arguments: %w", err)
		}

		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return nil, fmt.Errorf("failed to parse arguments: %w", err)
		}

		if params.Name == "" {
			return nil, fmt.Errorf("name parameter is required")
		}

		serverName := strings.TrimSpace(params.Name)

		// Check if server exists in catalog
		_, _, found := configuration.Find(serverName)
		if !found {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error: Server '%s' not found in catalog. Use mcp-find to search for available servers.", serverName),
				}},
			}, nil
		}

		// Append the new server to the current serverNames if not already present
		found = false
		for _, existing := range configuration.serverNames {
			if existing == serverName {
				found = true
				break
			}
		}
		if !found {
			configuration.serverNames = append(configuration.serverNames, serverName)
		}

		// Fetch updated secrets for the new server list
		if g.configurator != nil {
			if fbc, ok := g.configurator.(*FileBasedConfiguration); ok {
				updatedSecrets, err := fbc.readDockerDesktopSecrets(ctx, configuration.servers, configuration.serverNames)
				if err == nil {
					configuration.secrets = updatedSecrets
				} else {
					log("Warning: Failed to update secrets:", err)
				}
			}
		}

		// Update the current configuration state
		updatedServerNames := configuration.serverNames
		if err := g.reloadConfiguration(ctx, configuration, updatedServerNames, clientConfig); err != nil {
			return nil, fmt.Errorf("failed to reload configuration: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Successfully added server '%s'. Assume that it is fully configured and ready to use.", serverName),
			}},
		}, nil
	}

	return &ToolRegistration{
		Tool:    tool,
		Handler: withToolTelemetry("mcp-add", handler),
	}
}

// mcpRemoveTool implements a tool for removing servers from the registry
func (g *Gateway) createMcpRemoveTool(_ Configuration, clientConfig *clientConfig) *ToolRegistration {
	tool := &mcp.Tool{
		Name:        "mcp-remove",
		Description: "Remove an MCP server from the registry and reload the configuration. This will disable the server.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"name": {
					Type:        "string",
					Description: "Name of the MCP server to remove from the registry",
				},
			},
			Required: []string{"name"},
		},
	}

	handler := func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse parameters
		var params struct {
			Name string `json:"name"`
		}

		if req.Params.Arguments == nil {
			return nil, fmt.Errorf("missing arguments")
		}

		paramsBytes, err := json.Marshal(req.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal arguments: %w", err)
		}

		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return nil, fmt.Errorf("failed to parse arguments: %w", err)
		}

		if params.Name == "" {
			return nil, fmt.Errorf("name parameter is required")
		}

		serverName := strings.TrimSpace(params.Name)

		// Remove the server from the current serverNames
		updatedServerNames := slices.DeleteFunc(slices.Clone(g.configuration.serverNames), func(name string) bool {
			return name == serverName
		})

		// Update the current configuration state
		g.configuration.serverNames = updatedServerNames

		if err := g.reloadConfiguration(ctx, g.configuration, updatedServerNames, clientConfig); err != nil {
			return nil, fmt.Errorf("failed to reload configuration: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Successfully removed server '%s'.", serverName),
			}},
		}, nil
	}

	return &ToolRegistration{
		Tool:    tool,
		Handler: withToolTelemetry("mcp-remove", handler),
	}
}

func (g *Gateway) createMcpRegistryImportTool(configuration Configuration, _ *clientConfig) *ToolRegistration {
	tool := &mcp.Tool{
		Name:        "mcp-registry-import",
		Description: "Import MCP servers from an MCP registry URL. Fetches server definitions via HTTP GET and adds them to the local catalog.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"url": {
					Type:        "string",
					Description: "URL to fetch the server details JSON (must be a valid HTTP/HTTPS URL)",
				},
			},
			Required: []string{"url"},
		},
	}

	handler := func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse parameters
		var params struct {
			URL string `json:"url"`
		}

		if req.Params.Arguments == nil {
			return nil, fmt.Errorf("missing arguments")
		}

		paramsBytes, err := json.Marshal(req.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal arguments: %w", err)
		}

		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return nil, fmt.Errorf("failed to parse arguments: %w", err)
		}

		if params.URL == "" {
			return nil, fmt.Errorf("url parameter is required")
		}

		registryURL := strings.TrimSpace(params.URL)

		// Validate URL scheme
		if !strings.HasPrefix(registryURL, "http://") && !strings.HasPrefix(registryURL, "https://") {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error: URL must start with http:// or https://, got: %s", registryURL),
				}},
			}, nil
		}

		// Fetch servers from the URL
		servers, err := g.readServersFromURL(ctx, registryURL)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Error fetching servers from URL %s: %v", registryURL, err),
				}},
			}, nil
		}

		if len(servers) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("No servers found at URL: %s", registryURL),
				}},
			}, nil
		}

		// Add the imported servers to the current configuration and build detailed summary
		var importedServerNames []string
		var serverSummaries []string

		for serverName, server := range servers {
			if _, exists := configuration.servers[serverName]; exists {
				log(fmt.Sprintf("Warning: server '%s' from URL %s overwrites existing server", serverName, registryURL))
			}
			configuration.servers[serverName] = server
			importedServerNames = append(importedServerNames, serverName)

			// Build detailed summary for this server
			summary := fmt.Sprintf("• %s", serverName)

			if server.Description != "" {
				summary += fmt.Sprintf("\n  Description: %s", server.Description)
			}

			if server.Image != "" {
				summary += fmt.Sprintf("\n  Image: %s", server.Image)
			}

			// List required secrets
			if len(server.Secrets) > 0 {
				var secretNames []string
				for _, secret := range server.Secrets {
					secretNames = append(secretNames, secret.Name)
				}
				summary += fmt.Sprintf("\n  Required Secrets: %s", strings.Join(secretNames, ", "))
				summary += "\n  ⚠️  Configure these secrets before using this server"
			}

			// List configuration schemas available
			if len(server.Config) > 0 {
				summary += fmt.Sprintf("\n  Configuration Schemas: %d available", len(server.Config))
				summary += "\n  ℹ️  Use mcp-config-set to configure optional settings"
			}

			if server.LongLived {
				summary += "\n  🔄 Long-lived server (stays running)"
			}

			serverSummaries = append(serverSummaries, summary)
		}

		// Create comprehensive result message
		resultText := fmt.Sprintf("Successfully imported %d servers from %s\n\n", len(importedServerNames), registryURL)
		resultText += strings.Join(serverSummaries, "\n\n")

		if len(importedServerNames) > 0 {
			resultText += fmt.Sprintf("\n\n✅ Servers ready to use: %s", strings.Join(importedServerNames, ", "))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: resultText,
			}},
		}, nil
	}

	return &ToolRegistration{
		Tool:    tool,
		Handler: withToolTelemetry("mcp-registry-import", handler),
	}
}

// readServersFromURL fetches and parses server definitions from a URL
func (g *Gateway) readServersFromURL(ctx context.Context, url string) (map[string]catalog.Server, error) {
	servers := make(map[string]catalog.Server)

	log(fmt.Sprintf("  - Reading servers from URL: %s", url))

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set a reasonable user agent
	req.Header.Set("User-Agent", "docker-mcp-gateway/1.0.0")

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to parse as oci.ServerDetail (the new structure)
	var serverDetail oci.ServerDetail
	if err := json.Unmarshal(body, &serverDetail); err == nil && serverDetail.Name != "" {
		// Successfully parsed as ServerDetail - convert to catalog.Server
		server := serverDetail.ToCatalogServer()

		serverName := serverDetail.Name
		servers[serverName] = server
		log(fmt.Sprintf("  - Added server '%s' from URL %s", serverName, url))
		return servers, nil
	}

	return nil, fmt.Errorf("unable to parse response as OCI catalog or direct catalog format")
}

type configValue struct {
	Server string `json:"server"`
	Key    string `json:"key"`
	Value  any    `json:"value"`
}

// mcpConfigSetTool implements a tool for setting configuration values for MCP servers
func (g *Gateway) createMcpConfigSetTool(configuration Configuration, clientConfig *clientConfig) *ToolRegistration {
	tool := &mcp.Tool{
		Name:        "mcp-config-set",
		Description: "Set configuration values for MCP servers. Creates or updates server configuration with the specified key-value pairs.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"server": {
					Type:        "string",
					Description: "Name of the MCP server to configure",
				},
				"key": {
					Type:        "string",
					Description: "Configuration key to set",
				},
				"value": {
					Description: "Configuration value to set (can be string, number, boolean, or object)",
				},
			},
			Required: []string{"server", "key", "value"},
		},
	}

	handler := func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse parameters
		var params configValue

		if req.Params.Arguments == nil {
			return nil, fmt.Errorf("missing arguments")
		}

		paramsBytes, err := json.Marshal(req.Params.Arguments)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal arguments: %w", err)
		}

		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return nil, fmt.Errorf("failed to parse arguments: %w", err)
		}

		if params.Server == "" {
			return nil, fmt.Errorf("server parameter is required")
		}

		if params.Key == "" {
			return nil, fmt.Errorf("key parameter is required")
		}

		serverName := strings.TrimSpace(params.Server)
		configKey := strings.TrimSpace(params.Key)

		// Check if server exists in catalog (optional check - we can configure servers that don't exist yet)
		_, _, serverExists := configuration.Find(serverName)

		// Initialize the server's config map if it doesn't exist
		if configuration.config[serverName] == nil {
			configuration.config[serverName] = make(map[string]any)
		}

		// Set the configuration value
		oldValue := configuration.config[serverName][configKey]
		configuration.config[serverName][configKey] = params.Value

		// Log the configuration change
		log(fmt.Sprintf("  - Set config for server '%s': %s = %v", serverName, configKey, params.Value))

		// Reload configuration with current server list to apply changes
		if err := g.reloadConfiguration(ctx, configuration, configuration.serverNames, clientConfig); err != nil {
			return nil, fmt.Errorf("failed to reload configuration: %w", err)
		}

		var resultMessage string
		if oldValue != nil {
			resultMessage = fmt.Sprintf("Successfully updated config for server '%s': %s = %v (was: %v)", serverName, configKey, params.Value, oldValue)
		} else {
			resultMessage = fmt.Sprintf("Successfully set config for server '%s': %s = %v", serverName, configKey, params.Value)
		}

		if !serverExists {
			resultMessage += fmt.Sprintf(" (Note: server '%s' is not in the current catalog)", serverName)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: resultMessage,
			}},
		}, nil
	}

	return &ToolRegistration{
		Tool:    tool,
		Handler: withToolTelemetry("mcp-config-set", handler),
	}
}

// maxInt returns the maximum of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// withToolTelemetry wraps a tool handler with telemetry instrumentation
func withToolTelemetry(toolName string, handler mcp.ToolHandler) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		serverName := "dynamic-mcps"

		// Debug logging to stderr
		if os.Getenv("DOCKER_MCP_TELEMETRY_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[MCP-HANDLER] Tool call received: %s from server: %s\n", toolName, serverName)
		}

		// Start telemetry span for tool call
		startTime := time.Now()
		serverType := "dynamic"

		// Build span attributes
		spanAttrs := []attribute.KeyValue{
			attribute.String("mcp.server.name", serverName),
			attribute.String("mcp.server.type", serverType),
		}

		ctx, span := telemetry.StartToolCallSpan(ctx, toolName, spanAttrs...)
		defer span.End()

		// Record tool call counter
		telemetry.ToolCallCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("mcp.server.name", serverName),
				attribute.String("mcp.server.type", serverType),
				attribute.String("mcp.tool.name", toolName),
				attribute.String("mcp.client.name", req.Session.InitializeParams().ClientInfo.Name),
			),
		)

		// Execute the wrapped handler
		result, err := handler(ctx, req)

		// Record duration
		duration := time.Since(startTime).Milliseconds()
		telemetry.ToolCallDuration.Record(ctx, float64(duration),
			metric.WithAttributes(
				attribute.String("mcp.server.name", serverName),
				attribute.String("mcp.server.type", serverType),
				attribute.String("mcp.tool.name", toolName),
				attribute.String("mcp.client.name", req.Session.InitializeParams().ClientInfo.Name),
			),
		)

		if err != nil {
			// Record error in telemetry
			telemetry.RecordToolError(ctx, span, serverName, serverType, toolName)
			span.SetStatus(codes.Error, "Tool execution failed")
			return nil, err
		}

		span.SetStatus(codes.Ok, "")
		return result, nil
	}
}
