package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/version"
)

func versionCommand() *cobra.Command {
	return &cobra.Command{
		Short: "Show the version information",
		Use:   "version",
		Args:  cobra.ExactArgs(0),
		// Deactivate PersistentPreRun for this command only
		// We don't want to check if Docker Desktop is running.
		PersistentPreRun: func(*cobra.Command, []string) {},
		Run: func(*cobra.Command, []string) {
			fmt.Println(version.Version)
		},
	}
}
