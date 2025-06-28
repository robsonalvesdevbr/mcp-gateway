package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

const jcatImage = "docker/jcat@sha256:76719466e8b99a65dd1d37d9ab94108851f009f0f687dce7ff8a6fc90575c4d4"

func (c *dockerClient) ReadSecrets(ctx context.Context, names []string, lenient bool) (map[string]string, error) {
	if err := c.PullImage(ctx, jcatImage); err != nil {
		return nil, err
	}

	flags := []string{"--network=none", "--pull=never"}
	var command []string

	for i, name := range names {
		file := fmt.Sprintf("/.s%d", i)
		flags = append(flags, "-l", "x-secret:"+name+"="+file)
		command = append(command, file)
	}

	args := []string{"run", "--rm"}
	args = append(args, flags...)
	args = append(args, jcatImage)
	args = append(args, command...)
	buf, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("fetching secrets %w: %s", err, string(buf))
	}

	var list []string
	if err := json.Unmarshal(buf, &list); err != nil {
		return nil, err
	}

	values := map[string]string{}
	for i := range names {
		values[names[i]] = list[i]
	}

	return values, nil
}
