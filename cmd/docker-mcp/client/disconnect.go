package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type disconnectOpts struct {
	Global bool
	Quiet  bool
}

func DisconnectCommand(cwd string, cfg Config) *cobra.Command {
	opts := &disconnectOpts{}
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("disconnect [OPTIONS] <mcp-client>\n\nSupported clients: %s", strings.Join(getSupportedMCPClients(cfg), " ")),
		Short: "Disconnect the Docker MCP Toolkit from a client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDisconnect(cmd.Context(), cwd, cfg, args[0], *opts)
		},
	}
	flags := cmd.Flags()
	addGlobalFlag(flags, &opts.Global)
	addQuietFlag(flags, &opts.Quiet)
	return cmd
}

func runDisconnect(ctx context.Context, cwd string, config Config, vendor string, opts disconnectOpts) error {
	if vendor == vendorGordon && opts.Global {
		if err := disconnectGordon(ctx); err != nil {
			return err
		}
	} else {
		updater, err := GetUpdater(vendor, opts.Global, cwd, config)
		if err != nil {
			return err
		}
		if err := updater(DockerMCPCatalog, nil); err != nil {
			return err
		}
	}
	if opts.Quiet {
		return nil
	}
	if err := runList(ctx, cwd, config, listOptions{Global: opts.Global}); err != nil {
		return err
	}
	fmt.Printf("You might have to restart '%s'.\n", vendor)
	return nil
}
