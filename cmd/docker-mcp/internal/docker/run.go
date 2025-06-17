package docker

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/docker/mcp-cli/cmd/docker-mcp/internal/desktop"
	"github.com/docker/mcp-cli/cmd/docker-mcp/internal/user"
)

func RunOnDockerDesktop(ctx context.Context, args ...string) ([]byte, error) {
	desktop.AvoidResourceSaverMode(ctx)

	// TODO(dga): use pinata code and handle linux
	var host string
	if runtime.GOOS == "windows" {
		host = "npipe:////./pipe/docker_engine_linux"
	} else {
		home, err := user.HomeDir()
		if err != nil {
			return nil, err
		}

		host = "unix://" + filepath.Join(home, "Library/Containers/com.docker.docker/Data/docker.raw.sock")
	}

	args = append([]string{"-H", host, "run", "--rm"}, args...)
	return exec.CommandContext(ctx, "docker", args...).Output()
}
