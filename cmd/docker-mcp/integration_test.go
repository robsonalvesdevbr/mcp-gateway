package main

import (
	"os/exec"
	"testing"

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
