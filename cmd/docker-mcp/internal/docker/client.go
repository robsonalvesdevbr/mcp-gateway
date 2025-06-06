package docker

import (
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/client"
)

type Client struct {
	client client.APIClient
}

func NewClient(cli command.Cli) *Client {
	return &Client{
		client: cli.Client(),
	}
}
