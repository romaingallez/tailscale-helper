package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"tailscale_helper/internal/render"
	"tailscale_helper/internal/tailscale"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show enriched Tailscale status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd)
		},
	}

	cmd.Flags().Bool("json", false, "print normalized status as JSON")

	return cmd
}

func runStatus(cmd *cobra.Command) error {
	client := tailscale.NewClient("tailscale")
	status, err := client.Status(cmd.Context())
	if err != nil {
		return err
	}

	asJSON, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}

	if asJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(status)
	}

	_, err = fmt.Fprint(os.Stdout, render.StatusTable(status))
	return err
}
