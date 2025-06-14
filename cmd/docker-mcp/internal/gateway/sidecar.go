package gateway

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os/exec"
	"strings"

	"github.com/docker/mcp-cli/cmd/docker-mcp/internal/mcp"
)

const image = "davidgageot135/http-proxy@sha256:372365af904f6dc148f51286784a193d2a48db8ae6d2b8abb98be6f77f9b3efb"

func (g *Gateway) runProxySideCar(ctx context.Context, allowedHosts []string) (func(context.Context) error, string, error) {
	log("  - Running proxy sidecar for hosts", allowedHosts)

	if err := g.dockerClient.PullImage(ctx, image); err != nil {
		return nil, "", fmt.Errorf("pulling image %s: %w", image, err)
	}

	// Start the proxy.
	// TODO: use the docker api
	cmd := exec.CommandContext(ctx, "docker", "run", "-d", "--rm", "--label", "docker-mcp=true", "-e", "ALLOWED_HOSTS", image)
	cmd.Env = []string{"ALLOWED_HOSTS=" + strings.Join(allowedHosts, ",")}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, "", fmt.Errorf("starting http proxy: %w", err)
	}
	id := strings.TrimSpace(string(out))

	// Create internal network, for the MCP Servers.
	network := "docker-mcp-internal" + randString(6)
	if err := g.dockerClient.CreateNetwork(ctx, network, true, map[string]string{"docker-mcp": "true"}); err != nil {
		_ = g.dockerClient.RemoveContainer(ctx, id, true)
		return nil, "", fmt.Errorf("creating internal network %s: %w", network, err)
	}

	cleanup := func(ctx context.Context) error {
		return errors.Join(
			g.dockerClient.RemoveContainer(ctx, id, true),
			g.dockerClient.RemoveNetwork(ctx, network),
		)
	}

	// Connect the proxy to the internal network, in addition to the default bridge network.
	if err := g.dockerClient.ConnectNetwork(ctx, network, id, "proxy"); err != nil {
		_ = cleanup(ctx)
		return nil, "", fmt.Errorf("attaching proxy to internal network %s: %w", id, err)
	}

	return cleanup, network, nil
}

func randString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}

func newClientWithCleanup(client mcp.Client, cleanup func(context.Context) error) mcp.Client {
	return &clientWithCleanup{
		Client:  client,
		cleanup: cleanup,
	}
}

func (c *clientWithCleanup) Close() error {
	return errors.Join(c.Client.Close(), c.cleanup(context.TODO()))
}

type clientWithCleanup struct {
	mcp.Client
	cleanup func(context.Context) error
}
