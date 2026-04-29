package cmd

import (
	"github.com/spf13/cobra"

	"github.com/romaingallez/tailscale-helper/internal/tailscale"
	app "github.com/romaingallez/tailscale-helper/internal/tui"
)

func newTUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Browse Tailscale nodes in a TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := tailscale.NewClient("tailscale")
			return app.Run(cmd.Context(), client)
		},
	}
}
