package catalog

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Update a specific catalog or all catalogs if no name is provided",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd.Context(), args)
		},
	}
	return cmd
}

func runUpdate(ctx context.Context, args []string) error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	var names []string
	if len(args) == 0 {
		names = getAllCatalogNames(*cfg)
	}
	for _, arg := range args {
		if _, ok := cfg.Catalogs[arg]; ok {
			names = append(names, arg)
		} else {
			return fmt.Errorf("unknown catalog %q", arg)
		}
	}
	var errs []error
	for _, name := range names {
		catalog, ok := cfg.Catalogs[name]
		if !ok {
			continue
		}
		if err := updateCatalog(ctx, name, catalog); err != nil {
			errs = append(errs, err)
		}
		fmt.Println("updated:", name)

	}
	return errors.Join(errs...)
}

func getAllCatalogNames(cfg Config) []string {
	var names []string
	for name := range cfg.Catalogs {
		names = append(names, name)
	}
	return names
}

func updateCatalog(ctx context.Context, name string, catalog Catalog) error {
	url := catalog.URL

	var (
		catalogContent []byte
		err            error
	)
	if isValidURL(url) {
		// For the docker catalog, always take the most recent URL
		if name == DockerCatalogName {
			url = DockerCatalogURL
		}
		catalogContent, err = DownloadFile(ctx, url)
	} else {
		catalogContent, err = os.ReadFile(url)
	}
	if err != nil {
		return err
	}

	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	cfg.Catalogs[name] = Catalog{
		DisplayName: catalog.DisplayName,
		URL:         catalog.URL,
		LastUpdate:  time.Now().Format(time.RFC3339),
	}
	if err := WriteConfig(cfg); err != nil {
		return err
	}

	if err := WriteCatalogFile(name, catalogContent); err != nil {
		return fmt.Errorf("failed to write catalog %q: %w", name, err)
	}
	return nil
}
