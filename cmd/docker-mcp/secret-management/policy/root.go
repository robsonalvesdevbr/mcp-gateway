package policy

import (
	"github.com/spf13/cobra"
)

func NewPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "policy",
		Aliases: []string{"policies"},
		Short:   "Manage secret policies",
	}
	cmd.AddCommand(SetCommand())
	cmd.AddCommand(DumpCommand())
	return cmd
}
