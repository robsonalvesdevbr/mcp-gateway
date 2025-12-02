package oauth

import (
	"os"

	"github.com/docker/mcp-gateway/pkg/desktop"
)

// IsCEMode returns true if running in Docker CE mode (standalone OAuth flows).
// When false, uses Docker Desktop for OAuth orchestration.
//
// Detection logic:
// 1. Check DOCKER_MCP_USE_CE env var (explicit override)
// 2. Auto-detect by checking if Docker Desktop backend socket exists
// 3. If socket doesn't exist, assume CE mode
func IsCEMode() bool {
	// Allow explicit override via env var
	if ceMode := os.Getenv("DOCKER_MCP_USE_CE"); ceMode != "" {
		return ceMode == "true"
	}

	// Auto-detect based on Desktop socket availability
	_, err := os.Stat(desktop.Paths().BackendSocket)

	// If socket doesn't exist, we're in CE mode
	return err != nil
}
