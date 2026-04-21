package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "tailscale-helper",
		Short:         "Tailscale status with better display",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd)
		},
	}

	rootCmd.PersistentFlags().String("config", "", "config file path")
	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	rootCmd.Flags().Bool("json", false, "print normalized status as JSON")

	viper.SetEnvPrefix("TAILSCALE_HELPER")
	viper.AutomaticEnv()

	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(newTUICmd())

	return rootCmd
}
