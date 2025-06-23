package desktop

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const jcatImage = "docker/jcat@sha256:76719466e8b99a65dd1d37d9ab94108851f009f0f687dce7ff8a6fc90575c4d4"

type StoredSecret struct {
	Name     string `json:"name"`
	Provider string `json:"provider,omitempty"`
}

type Secret struct {
	Name     string `json:"name"`
	Provider string `json:"provider,omitempty"`
	Value    string `json:"value"`
}

func NewSecretsClient() *Secrets {
	return &Secrets{
		rawClient: newRawClient(dialSecrets),
	}
}

type Secrets struct {
	rawClient *RawClient
}

func (c *Secrets) DeleteJfsSecret(ctx context.Context, secret string) error {
	AvoidResourceSaverMode(ctx)

	return c.rawClient.Delete(ctx, fmt.Sprintf("/secrets/%v", secret))
}

func (c *Secrets) GetJfsPolicy(ctx context.Context) (string, error) {
	AvoidResourceSaverMode(ctx)

	var result string
	err := c.rawClient.Get(ctx, "/policy", &result)
	return result, err
}

func (c *Secrets) ListJfsSecrets(ctx context.Context) ([]StoredSecret, error) {
	AvoidResourceSaverMode(ctx)

	var result []StoredSecret
	err := c.rawClient.Get(ctx, "/secrets", &result)
	return result, err
}

func (c *Secrets) SetJfsPolicy(ctx context.Context, body string) error {
	AvoidResourceSaverMode(ctx)

	return c.rawClient.Post(ctx, "/policy", body, nil)
}

func (c *Secrets) SetJfsSecret(ctx context.Context, secret Secret) error {
	AvoidResourceSaverMode(ctx)

	return c.rawClient.Post(ctx, "/secrets", secret, nil)
}

func ReadSecretValues(ctx context.Context, names []string) (map[string]string, error) {
	flags := []string{"--network=none", "--pull=missing"}
	var command []string

	for i, name := range names {
		file := fmt.Sprintf("/.s%d", i)
		flags = append(flags, "-l", "x-secret:"+name+"="+file)
		command = append(command, file)
	}

	var args []string
	args = append(args, flags...)
	args = append(args, jcatImage)
	args = append(args, command...)

	buf, err := runWithRawDockerSocket(ctx, args...)
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

func runWithRawDockerSocket(ctx context.Context, args ...string) ([]byte, error) {
	AvoidResourceSaverMode(ctx)

	var path string
	if runtime.GOOS == "windows" {
		path = "npipe://" + strings.ReplaceAll(Paths().RawDockerSocket, `\`, `/`)
	} else {
		path = "unix://" + Paths().RawDockerSocket
	}

	args = append([]string{"-H", path, "run", "--rm"}, args...)
	return exec.CommandContext(ctx, "docker", args...).CombinedOutput()
}
