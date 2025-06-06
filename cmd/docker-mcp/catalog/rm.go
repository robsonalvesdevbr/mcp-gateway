package catalog

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newRmCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove a catalog",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runRm(args[0])
		},
		Hidden: true,
	}
	return cmd
}

func runRm(name string) error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	if _, ok := cfg.Catalogs[name]; !ok {
		return fmt.Errorf("catalog %q not found", name)
	}
	delete(cfg.Catalogs, name)
	if err := writeConfig(cfg); err != nil {
		return err
	}
	file, err := toCatalogFilePath(name)
	if err != nil {
		return err
	}
	if err := os.Remove(file); err != nil {
		return err
	}
	fmt.Printf("removed catalog %q\n", name)
	return nil
}
