package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/catalog"
)

func catalogCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "catalog",
		Aliases: []string{"catalogs"},
		Short:   "Manage the catalog",
	}
	cmd.AddCommand(importCatalogCommand())
	cmd.AddCommand(lsCatalogCommand())
	cmd.AddCommand(rmCatalogCommand())
	cmd.AddCommand(updateCatalogCommand())
	cmd.AddCommand(showCatalogCommand())
	cmd.AddCommand(forkCatalogCommand())
	cmd.AddCommand(createCatalogCommand())
	cmd.AddCommand(initCatalogCommand())
	cmd.AddCommand(addCatalogCommand())
	cmd.AddCommand(resetCatalogCommand())
	return cmd
}

func importCatalogCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "import <alias|url|file>",
		Short: "Import a catalog",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return catalog.Import(cmd.Context(), args[0])
		},
		Hidden: true,
	}
}

func lsCatalogCommand() *cobra.Command {
	var opts struct {
		JSON bool
	}
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List configured catalogs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return catalog.Ls(cmd.Context(), opts.JSON)
		},
	}
	flags := cmd.Flags()
	flags.BoolVar(&opts.JSON, "json", false, "Print as JSON.")
	return cmd
}

func rmCatalogCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove a catalog",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return catalog.Rm(args[0])
		},
		Hidden: true,
	}
}

func updateCatalogCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update [name]",
		Short: "Update a specific catalog or all catalogs if no name is provided",
		RunE: func(cmd *cobra.Command, args []string) error {
			return catalog.Update(cmd.Context(), args)
		},
	}
}

func showCatalogCommand() *cobra.Command {
	var opts struct {
		Format catalog.Format
	}
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show a catalog",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := catalog.DockerCatalogName
			if len(args) > 0 {
				name = args[0]
			}

			return catalog.Show(cmd.Context(), name, opts.Format)
		},
	}
	flags := cmd.Flags()
	flags.Var(&opts.Format, "format", fmt.Sprintf("Supported: %s.", catalog.SupportedFormats()))
	return cmd
}

func forkCatalogCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "fork <src-catalog> <new-name>",
		Short: "Fork a catalog",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			return catalog.Fork(args[0], args[1])
		},
		Hidden: true,
	}
}

func createCatalogCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new catalog",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return catalog.Create(args[0])
		},
		Hidden: true,
	}
}

func initCatalogCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize the catalog",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return catalog.Init(cmd.Context())
		},
	}
}

func addCatalogCommand() *cobra.Command {
	var opts struct {
		Force bool
	}
	cmd := &cobra.Command{
		Use:   "add <catalog> <server-name> <catalog-file>",
		Short: "Add a server to your catalog",
		Args:  cobra.ExactArgs(3),
		RunE: func(_ *cobra.Command, args []string) error {
			parsedArgs := catalog.ParseAddArgs(args[0], args[1], args[2])
			if err := catalog.ValidateArgs(*parsedArgs); err != nil {
				return err
			}
			return catalog.Add(*parsedArgs, opts.Force)
		},
		Hidden: true,
	}
	flags := cmd.Flags()
	flags.BoolVar(&opts.Force, "force", false, "Overwrite existing server in the catalog")
	return cmd
}

func resetCatalogCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "reset",
		Aliases: []string{"empty"},
		Short:   "Empty the catalog",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return catalog.Reset(cmd.Context())
		},
	}
}
