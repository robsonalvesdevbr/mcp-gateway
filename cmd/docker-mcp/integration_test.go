package main

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func thisIsAnIntegrationTest(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test.")
	}
}

func TestIntegrationVersion(t *testing.T) {
	thisIsAnIntegrationTest(t)

	cmd := exec.CommandContext(t.Context(), "docker", "mcp", "version")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.NotEmpty(t, out)
}

func TestIntegrationCatalogLs(t *testing.T) {
	thisIsAnIntegrationTest(t)

	cmd := exec.CommandContext(t.Context(), "docker", "mcp", "catalog", "ls")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Contains(t, string(out), "docker-mcp: Docker MCP Catalog")
}

func TestIntegrationCatalogShow(t *testing.T) {
	thisIsAnIntegrationTest(t)

	cmd := exec.CommandContext(t.Context(), "docker", "mcp", "catalog", "show")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Contains(t, string(out), "playwright:")
}
