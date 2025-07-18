package catalog

import (
	"fmt"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/internal/config"
)

func Rm(name string) error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	if _, ok := cfg.Catalogs[name]; !ok {
		return fmt.Errorf("catalog %q not found", name)
	}
	delete(cfg.Catalogs, name)
	if err := WriteConfig(cfg); err != nil {
		return err
	}
	if err := config.RemoveCatalogFile(name); err != nil {
		return err
	}

	fmt.Printf("removed catalog %q\n", name)
	return nil
}
