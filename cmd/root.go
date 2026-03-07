// Package cmd contains all Cobra CLI commands for aiswitch.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/anmolnagpal/aiswitch/internal/config"
	"github.com/anmolnagpal/aiswitch/internal/providers/claude"
	"github.com/anmolnagpal/aiswitch/internal/providers/github"
	"github.com/anmolnagpal/aiswitch/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:   "aiswitch [profile]",
	Short: "Switch between Claude and GitHub Copilot accounts instantly",
	Long: `aiswitch lets you manage multiple Claude / GitHub Copilot identities
and switch between them with a single command — similar to tfswitch for Terraform.

Run without arguments to open the interactive profile selector.
Provide a profile name to switch directly without the interactive UI.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		var target string

		if len(args) == 1 {
			target = args[0]
		} else {
			selected, err := ui.RunSelector(cfg.Profiles, cfg.ActiveProfile)
			if err != nil {
				return err
			}
			if selected == "" {
				return nil // user quit
			}
			target = selected
		}

		return applyProfile(cfg, target)
	},
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(currentCmd)
	rootCmd.AddCommand(shellInitCmd)
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(localInitCmd)
	rootCmd.AddCommand(detectCmd)
}

// applyProfile applies all providers for the named profile.
func applyProfile(cfg *config.Config, name string) error {
	profile, ok := cfg.Profiles[name]
	if !ok {
		return fmt.Errorf("profile %q not found — run: aiswitch list", name)
	}

	paths, err := config.GetEnvPaths()
	if err != nil {
		return err
	}

	// Ensure the directory and both env files exist.
	dir, _ := config.Dir()
	_ = os.MkdirAll(dir, 0o700)
	initEnvFile(paths.Sh, "# aiswitch env — source this file or add it to your shell profile\n")
	initEnvFile(paths.PS1, "# aiswitch env — dot-source this file: . ~/.aiswitch/env.ps1\n")

	if profile.Claude != nil {
		if err := claude.Apply(*profile.Claude, paths); err != nil {
			return fmt.Errorf("applying Claude config: %w", err)
		}
	}
	if profile.GitHub != nil {
		if err := github.Apply(*profile.GitHub, paths); err != nil {
			return fmt.Errorf("applying GitHub config: %w", err)
		}
	}

	cfg.ActiveProfile = name
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	ui.PrintSwitchResult(name, profile, paths.Sh)
	return nil
}

// initEnvFile creates path with header content if it does not already exist.
func initEnvFile(path, header string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.WriteFile(path, []byte(header), 0o600)
	}
}
