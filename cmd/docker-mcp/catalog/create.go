package catalog

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new catalog",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runCreate(args[0])
		},
		Hidden: true,
	}
	return cmd
}

func runCreate(name string) error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	if _, ok := cfg.Catalogs[name]; ok {
		return fmt.Errorf("catalog %q already exists", name)
	}
	cfg.Catalogs[name] = Catalog{DisplayName: name}
	if err := WriteConfig(cfg); err != nil {
		return err
	}
	content, err := setCatalogMetaData([]byte{}, MetaData{Name: name, DisplayName: name})
	if err != nil {
		return err
	}
	if err := WriteCatalogFile(name, content); err != nil {
		return err
	}
	fmt.Printf("created empty catalog %s\n", name)
	return nil
}
