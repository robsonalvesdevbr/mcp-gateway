package docker

import (
	"context"
	"fmt"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
)

func (c *Client) ContainerExists(ctx context.Context, container string) (bool, container.InspectResponse, error) {
	response, err := c.client.ContainerInspect(ctx, container)
	if cerrdefs.IsNotFound(err) {
		return false, response, nil
	}

	return err == nil, response, err
}

func (c *Client) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	return c.client.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: force,
	})
}

func (c *Client) StartContainer(ctx context.Context, containerID string, containerConfig container.Config, hostConfig container.HostConfig) error {
	resp, err := c.client.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, nil, containerID)
	if err != nil {
		return fmt.Errorf("creating container: %w", err)
	}

	if err := c.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("starting container: %w", err)
	}

	return nil
}
