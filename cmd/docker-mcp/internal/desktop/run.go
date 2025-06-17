package desktop

import (
	"context"
	"os/exec"
	"runtime"
)

func RunWithRawDockerSocket(ctx context.Context, args ...string) ([]byte, error) {
	AvoidResourceSaverMode(ctx)

	prefix := "unix://"
	if runtime.GOOS == "windows" {
		prefix = "npipe://"
	}
	path := prefix + Paths().RawDockerSocket
	args = append([]string{"-H", path, "run", "--rm"}, args...)
	return exec.CommandContext(ctx, "docker", args...).Output()
}
