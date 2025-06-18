package catalog

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newForkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fork <src-catalog> <new-name>",
		Short: "Fork a catalog",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			return runFork(args[0], args[1])
		},
		Hidden: true,
	}
	return cmd
}

func runFork(src, dst string) error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	if _, ok := cfg.Catalogs[src]; !ok {
		return fmt.Errorf("catalog %q not found", src)
	}
	if _, ok := cfg.Catalogs[dst]; ok {
		return fmt.Errorf("catalog %q already exists", dst)
	}
	dstDisplayName := fmt.Sprintf("%s (forked from %s)", dst, src)
	cfg.Catalogs[dst] = Catalog{DisplayName: dstDisplayName}
	if err := WriteConfig(cfg); err != nil {
		return err
	}
	content, err := ReadCatalogFile(src)
	if err != nil {
		return err
	}
	dstContent, err := setCatalogMetaData(content, MetaData{DisplayName: dstDisplayName, Name: dst})
	if err != nil {
		return err
	}
	return WriteCatalogFile(dst, dstContent)
}
