package client

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func ManualCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "manual-instructions",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			printAsJSON, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			command := []string{"docker", "mcp", "gateway", "run"}
			if printAsJSON {
				buf, err := json.Marshal(command)
				if err != nil {
					return err
				}
				_, _ = cmd.OutOrStdout().Write(buf)
			} else {
				fmt.Fprint(cmd.OutOrStdout(), strings.Join(command, " "))
			}

			return nil
		},
		Hidden: true,
	}
	cmd.Flags().Bool("json", false, "Print as JSON.")
	return cmd
}
