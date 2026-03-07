package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/anmolnagpal/aiswitch/internal/config"
	"github.com/anmolnagpal/aiswitch/internal/ui"
)

var removeCmd = &cobra.Command{
	Use:     "remove <profile>",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a profile",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if _, ok := cfg.Profiles[name]; !ok {
			return fmt.Errorf("profile %q not found", name)
		}

		var confirmed bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Remove profile %q?", name)).
					Description("This cannot be undone.").
					Value(&confirmed),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}
		if !confirmed {
			fmt.Println(ui.StyleMuted.Render("Cancelled."))
			return nil
		}

		delete(cfg.Profiles, name)
		if cfg.ActiveProfile == name {
			cfg.ActiveProfile = ""
		}

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Println(ui.StyleSuccess.Render("✓ Removed profile \"" + name + "\""))
		return nil
	},
}
