package catalog

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func Get() (Catalog, error) {
	return ReadFrom("docker-mcp.yaml")
}

func ReadFrom(fileOrURL string) (Catalog, error) {
	servers, err := readMCPServers(fileOrURL)
	if err != nil {
		return Catalog{}, err
	}
	if servers == nil {
		servers = map[string]Server{}
	}

	return Catalog{
		Servers: servers,
	}, nil
}

func readMCPServers(fileOrURL string) (map[string]Server, error) {
	buf, err := readFileOrURL(fileOrURL)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var topLevel topLevel
	if err := yaml.Unmarshal(buf, &topLevel); err != nil {
		return nil, err
	}

	return topLevel.Registry, nil
}

func readFileOrURL(fileOrURL string) ([]byte, error) {
	switch {
	case strings.HasPrefix(fileOrURL, "http://") || strings.HasPrefix(fileOrURL, "https://"):
		resp, err := http.Get(fileOrURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("failed to fetch URL: %s, status: %s", fileOrURL, resp.Status)
		}

		return io.ReadAll(resp.Body)

	case filepath.IsAbs(fileOrURL):
		buf, err := os.ReadFile(fileOrURL)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			return nil, err
		}
		return buf, nil

	default:
		homeDir, err := os.UserHomeDir()
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
