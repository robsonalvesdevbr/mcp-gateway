package catalog

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/internal/config"
)

func newResetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "reset",
		Aliases: []string{"empty"},
		Short:   "Empty the catalog",
		Args:    cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			catalogsDir, err := config.FilePath("catalogs")
			if err != nil {
				return err
			}
			if err := os.RemoveAll(catalogsDir); err != nil {
				return err
			}

			return WriteConfig(&Config{})
		},
	}
}
