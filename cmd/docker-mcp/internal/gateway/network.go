package gateway

import (
	"context"
	"errors"
	"os"
	"sort"
)

func (g *Gateway) guessNetworks(ctx context.Context) ([]string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	found, response, err := g.docker.ContainerExists(ctx, hostname)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.New("can't inspect the current container. It probably has a non default hostname")
	}

	var networks []string
	for network := range response.NetworkSettings.Networks {
		networks = append(networks, network)
	}
	sort.Strings(networks)

	return networks, nil
}
