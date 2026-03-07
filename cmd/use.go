package cmd

import (
	"github.com/spf13/cobra"

	"github.com/anmolnagpal/aiswitch/internal/config"
)

var useCmd = &cobra.Command{
	Use:   "use <profile>",
	Short: "Switch to a named profile directly (non-interactive)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		return applyProfile(cfg, args[0])
	},
}
