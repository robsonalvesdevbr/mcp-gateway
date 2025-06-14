package docker

import (
	"context"

	"github.com/docker/docker/api/types/network"
)

func (c *Client) CreateNetwork(ctx context.Context, name string, internal bool, labels map[string]string) error {
	_, err := c.client.NetworkCreate(ctx, name, network.CreateOptions{
		Internal: internal,
		Labels:   labels,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) RemoveNetwork(ctx context.Context, name string) error {
	return c.client.NetworkRemove(ctx, name)
}

func (c *Client) ConnectNetwork(ctx context.Context, networkName, containerID, hostname string) error {
	return c.client.NetworkConnect(ctx, networkName, containerID, &network.EndpointSettings{
		Aliases: []string{hostname},
	})
}
