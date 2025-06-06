package catalog

import (
	"github.com/spf13/cobra"
)

func newResetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "reset",
		Aliases: []string{"empty"},
		Short:   "Empty the catalog",
		Args:    cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			return writeConfig(&Config{})
		},
	}
}
