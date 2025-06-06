package catalog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

type lsOpts struct {
	JSON bool
}

func newLsCommand() *cobra.Command {
	opts := &lsOpts{}
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List configured catalogs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runLs(cmd.Context(), *opts)
		},
	}
	flags := cmd.Flags()
	flags.BoolVar(&opts.JSON, "json", false, "Print as JSON.")
	return cmd
}

func runLs(ctx context.Context, opts lsOpts) error {
	cfg, err := ReadConfigWithDefaultCatalog(ctx)
	if err != nil {
		return err
	}

	if opts.JSON {
		data, err := json.Marshal(cfg)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	} else {
		humanPrintCatalog(*cfg)
	}
	return nil
}

func humanPrintCatalog(cfg Config) {
	if len(cfg.Catalogs) == 0 {
		fmt.Println("No catalogs configured.")
		return
	}
	for name, catalog := range cfg.Catalogs {
		fmt.Printf("%s: %s\n", name, catalog.DisplayName)
	}
}
