package oauth

import (
	"github.com/spf13/cobra"
)

func NewOAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "oauth",
		Hidden: true,
	}
	cmd.AddCommand(NewLsCmd())
	cmd.AddCommand(NewAuthorizeCmd())
	cmd.AddCommand(NewRevokeCmd())
	return cmd
}
