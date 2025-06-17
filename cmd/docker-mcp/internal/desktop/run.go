package desktop

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
)

func RunWithRawDockerSocket(ctx context.Context, args ...string) ([]byte, error) {
	AvoidResourceSaverMode(ctx)

	var path string
	if runtime.GOOS == "windows" {
		path = "npipe://" + strings.ReplaceAll(Paths().RawDockerSocket, `\`, `/`)
	} else {
		path = "unix://" + Paths().RawDockerSocket
	}

	args = append([]string{"-H", path, "run", "--rm"}, args...)
	return exec.CommandContext(ctx, "docker", args...).Output()
}
