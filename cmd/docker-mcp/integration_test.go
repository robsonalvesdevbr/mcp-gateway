package main

import (
	"os/exec"
	"testing"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/catalog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runDockerMCP(t *testing.T, args ...string) string {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test.")
	}

	args = append([]string{"mcp"}, args...)
	cmd := exec.CommandContext(t.Context(), "docker", args...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)

	return string(out)
}

func TestIntegrationVersion(t *testing.T) {
	out := runDockerMCP(t, "version")
	assert.NotEmpty(t, out)
}

func TestIntegrationCatalogLs(t *testing.T) {
	out := runDockerMCP(t, "catalog", "ls")
	assert.Contains(t, out, "docker-mcp: Docker MCP Catalog")
}

func TestIntegrationCatalogShow(t *testing.T) {
	out := runDockerMCP(t, "catalog", "show")
	assert.Contains(t, out, "playwright:")
}

func TestIntegrationCatalogDryRunEmpty(t *testing.T) {
	out := runDockerMCP(t, "gateway", "run", "--dry-run", "--servers=")
	assert.Contains(t, out, "Initialized in")
}

func TestIntegrationCatalogDryRunFetch(t *testing.T) {
	out := runDockerMCP(t, "gateway", "run", "--dry-run", "--servers=fetch", "--catalog="+catalog.DockerCatalogURL)
	assert.Contains(t, out, "fetch: (1 tools)")
	assert.Contains(t, out, "Initialized in")
}
