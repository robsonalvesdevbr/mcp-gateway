package secret

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/docker/mcp-gateway/cmd/docker-mcp/internal/docker"
)

func NewSecretsCmd(docker docker.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret",
		Short:   "Manage secrets",
		Example: strings.Trim(setExample, "\n"),
	}
	cmd.AddCommand(rmCommand())
	cmd.AddCommand(listCommand())
	cmd.AddCommand(setCommand())
	cmd.AddCommand(exportCommand(docker))
	return cmd
}
