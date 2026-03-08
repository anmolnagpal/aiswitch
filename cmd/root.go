// Package cmd contains all Cobra CLI commands for aiswitch.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/anmolnagpal/aiswitch/internal/config"
	"github.com/anmolnagpal/aiswitch/internal/providers/claude"
	"github.com/anmolnagpal/aiswitch/internal/providers/gemini"
	"github.com/anmolnagpal/aiswitch/internal/providers/github"
	"github.com/anmolnagpal/aiswitch/internal/providers/ide"
	"github.com/anmolnagpal/aiswitch/internal/providers/openai"
	"github.com/anmolnagpal/aiswitch/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:   "aiswitch [profile]",
	Short: "Switch between Claude, OpenAI, Gemini, and GitHub Copilot accounts",
	Long: `aiswitch — manage multiple AI provider accounts and switch between them
with a single command. Works with Claude, OpenAI, Gemini, GitHub Copilot,
Cursor, and Windsurf. Like tfswitch for Terraform, but for AI.

Run without arguments to open the interactive profile picker.
Provide a profile name to switch directly without the TUI.`,
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

// SetVersion injects build-time version information into the root command.
func SetVersion(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
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
	rootCmd.AddCommand(setupCmd)
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

	// Ensure the directory and env file(s) exist.
	dir, _ := config.Dir()
	_ = os.MkdirAll(dir, 0o700)
	initEnvFile(paths.Sh, "# aiswitch env — source this file or add it to your shell profile\n")
	if paths.PS1 != "" {
		initEnvFile(paths.PS1, "# aiswitch env — dot-source this file: . ~/.aiswitch/env.ps1\n")
	}

	if profile.Claude != nil {
		if err := claude.Apply(*profile.Claude, paths); err != nil {
			return fmt.Errorf("applying Claude config: %w", err)
		}
	} else {
		claude.Clear(paths)
	}
	if profile.OpenAI != nil {
		if err := openai.Apply(*profile.OpenAI, paths); err != nil {
			return fmt.Errorf("applying OpenAI config: %w", err)
		}
	} else {
		openai.Clear(paths)
	}
	if profile.Gemini != nil {
		if err := gemini.Apply(*profile.Gemini, paths); err != nil {
			return fmt.Errorf("applying Gemini config: %w", err)
		}
	} else {
		gemini.Clear(paths)
	}
	if profile.GitHub != nil {
		if err := github.Apply(*profile.GitHub, paths); err != nil {
			return fmt.Errorf("applying GitHub config: %w", err)
		}
	} else {
		github.Clear(paths)
	}
	if profile.IDE != nil {
		if err := ide.Apply(*profile.IDE, profile); err != nil {
			return fmt.Errorf("applying IDE config: %w", err)
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
