package user

import (
	"os"
	"path/filepath"
)

func HomeDir() (string, error) {
	//nolint:forbidigo
	home, err := os.UserHomeDir()
	if err != nil {
		// Probably HOME/USERPROFILE environment variable is hidden by the MCP client to improve security.
		// In that case, let's assume the current binary (docker-mcp) is in the standard docker cli plugins location
		// and derive the home directory from there.
		currentBinary := os.Args[0]
		if currentBinary == "" {
			return "", err
		}

		plugins := filepath.Dir(currentBinary)
		dotDocker := filepath.Dir(plugins)
		if filepath.Base(dotDocker) != ".docker" {
			return "", err
		}

		home := filepath.Dir(dotDocker)
		return home, nil
	}

	return home, nil
}
