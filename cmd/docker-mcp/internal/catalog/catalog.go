package catalog

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/internal/user"
)

func Get(ctx context.Context) (Catalog, error) {
	return ReadFrom(ctx, []string{"docker-mcp.yaml"})
}

func ReadFrom(ctx context.Context, fileOrURLs []string) (Catalog, error) {
	mergedServers := map[string]Server{}

	for _, fileOrURL := range fileOrURLs {
		servers, err := readMCPServers(ctx, fileOrURL)
		if err != nil {
			return Catalog{}, err
		}

		// Merge servers into the combined map, checking for overlaps
		for key, server := range servers {
			if _, exists := mergedServers[key]; exists {
				log.Printf("Warning: overlapping key '%s' found in catalog '%s', overwriting previous value", key, fileOrURL)
			}
			mergedServers[key] = server
		}
	}

	return Catalog{
		Servers: mergedServers,
	}, nil
}

func readMCPServers(ctx context.Context, fileOrURL string) (map[string]Server, error) {
	buf, err := readFileOrURL(ctx, fileOrURL)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]Server{}, nil
		}
		return nil, err
	}

	var topLevel topLevel
	if err := yaml.Unmarshal(buf, &topLevel); err != nil {
		return nil, err
	}

	return topLevel.Registry, nil
}

func readFileOrURL(ctx context.Context, fileOrURL string) ([]byte, error) {
	switch {
	case isURL(fileOrURL):
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileOrURL, nil)
		if err != nil {
			return nil, err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch URL: %s, status: %s", fileOrURL, resp.Status)
		}

		return io.ReadAll(resp.Body)

	case filepath.IsAbs(fileOrURL) || strings.HasPrefix(fileOrURL, "./"):
		buf, err := os.ReadFile(fileOrURL)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			return nil, err
		}
		return buf, nil

	default:
		homeDir, err := user.HomeDir()
		if err != nil {
			return nil, err
		}

		path := filepath.Join(homeDir, ".docker", "mcp", "catalogs", fileOrURL)

		buf, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			return nil, err
		}
		return buf, nil
	}
}

func isURL(fileOrURL string) bool {
	return strings.HasPrefix(fileOrURL, "http://") || strings.HasPrefix(fileOrURL, "https://")
}
