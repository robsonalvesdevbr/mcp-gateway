package secret

import (
	"strings"

	"github.com/spf13/cobra"
)

func NewSecretsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret",
		Short:   "Manage secrets",
		Example: strings.Trim(setExample, "\n"),
	}
	cmd.AddCommand(RmCommand())
	cmd.AddCommand(ListCommand())
	cmd.AddCommand(SetCommand())
	return cmd
}
