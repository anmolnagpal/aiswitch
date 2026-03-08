package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/anmolnagpal/aiswitch/internal/config"
	"github.com/anmolnagpal/aiswitch/internal/localfile"
	"github.com/anmolnagpal/aiswitch/internal/providers/claude"
	"github.com/anmolnagpal/aiswitch/internal/providers/gemini"
	"github.com/anmolnagpal/aiswitch/internal/providers/github"
	"github.com/anmolnagpal/aiswitch/internal/providers/ide"
	"github.com/anmolnagpal/aiswitch/internal/providers/openai"
	"github.com/anmolnagpal/aiswitch/internal/ui"
)

var detectQuiet bool

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Auto-detect and apply the nearest .aiswitch file",
	Long: `Walk up from the current directory looking for a .aiswitch file.
If found, switch to the profile it specifies (with any local overrides applied).

This command is called automatically by the shell cd hook installed by
'aiswitch shell-init'. You can also call it manually after cloning a repo.

Use --quiet to suppress output when no .aiswitch file is present — this keeps
your shell prompt clean when the hook fires in directories that don't use aiswitch.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		filePath := localfile.Find(cwd)
		if filePath == "" {
			if !detectQuiet {
				fmt.Println(ui.StyleMuted.Render("  No .aiswitch file found in this directory or any parent"))
			}
			return nil
		}

		local, err := localfile.Read(filePath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", filePath, err)
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		profile, ok := cfg.Profiles[local.Profile]
		if !ok {
			return fmt.Errorf("profile %q (from .aiswitch) not found — run: aiswitch list", local.Profile)
		}

		// Apply per-project overrides on top of the base profile.
		if local.Claude != nil && local.Claude.Model != "" && profile.Claude != nil {
			overridden := *profile.Claude
			overridden.DefaultModel = local.Claude.Model
			profile.Claude = &overridden
		}
		if local.OpenAI != nil && local.OpenAI.Model != "" && profile.OpenAI != nil {
			overridden := *profile.OpenAI
			overridden.DefaultModel = local.OpenAI.Model
			profile.OpenAI = &overridden
		}
		if local.Gemini != nil && local.Gemini.Model != "" && profile.Gemini != nil {
			overridden := *profile.Gemini
			overridden.DefaultModel = local.Gemini.Model
			profile.Gemini = &overridden
		}
		if local.GitHub != nil && local.GitHub.Email != "" && profile.GitHub != nil {
			overridden := *profile.GitHub
			overridden.Email = local.GitHub.Email
			profile.GitHub = &overridden
		}

		// Skip re-applying if already on this profile with no overrides.
		hasOverrides := (local.Claude != nil && local.Claude.Model != "") ||
			(local.GitHub != nil && local.GitHub.Email != "")
		if cfg.ActiveProfile == local.Profile && !hasOverrides {
			if !detectQuiet {
				fmt.Println(ui.StyleMuted.Render("  Already on profile \"" + local.Profile + "\""))
			}
			return nil
		}

		paths, err := config.GetEnvPaths()
		if err != nil {
			return err
		}

		dir, _ := config.Dir()
		_ = os.MkdirAll(dir, 0o700)
		initEnvFile(paths.Sh, "# aiswitch env — source this file or add it to your shell profile\n")
		initEnvFile(paths.PS1, "# aiswitch env — dot-source this file: . ~/.aiswitch/env.ps1\n")

		if profile.Claude != nil {
			if err := claude.Apply(*profile.Claude, paths); err != nil {
				return fmt.Errorf("applying Claude config: %w", err)
			}
		}
		if profile.OpenAI != nil {
			if err := openai.Apply(*profile.OpenAI, paths); err != nil {
				return fmt.Errorf("applying OpenAI config: %w", err)
			}
		}
		if profile.Gemini != nil {
			if err := gemini.Apply(*profile.Gemini, paths); err != nil {
				return fmt.Errorf("applying Gemini config: %w", err)
			}
		}
		if profile.GitHub != nil {
			if err := github.Apply(*profile.GitHub, paths); err != nil {
				return fmt.Errorf("applying GitHub config: %w", err)
			}
		}
		if profile.IDE != nil {
			if err := ide.Apply(*profile.IDE, profile); err != nil {
				return fmt.Errorf("applying IDE config: %w", err)
			}
		}

		cfg.ActiveProfile = local.Profile
		if err := config.Save(cfg); err != nil {
			return err
		}

		// Quiet mode: one-line indicator so users know a switch happened.
		if detectQuiet {
			indicator := ui.StyleSuccess.Render("⬡ aiswitch") +
				ui.StyleMuted.Render(" → ") +
				ui.StyleSelected.Render(local.Profile)
			if hasOverrides {
				indicator += ui.StyleMuted.Render(" (local overrides)")
			}
			fmt.Println(indicator)
			return nil
		}

		// Verbose mode.
		fmt.Println()
		fmt.Println(ui.StyleSuccess.Render("✓ Auto-switched to \""+local.Profile+"\"") +
			ui.StyleMuted.Render("  (from "+filePath+")"))
		if local.Claude != nil && local.Claude.Model != "" {
			fmt.Println(ui.StyleHint.Render("  Claude model override: " + local.Claude.Model))
		}
		if local.GitHub != nil && local.GitHub.Email != "" {
			fmt.Println(ui.StyleHint.Render("  Git email override: " + local.GitHub.Email))
		}
		fmt.Println()
		return nil
	},
}

func init() {
	detectCmd.Flags().BoolVar(&detectQuiet, "quiet", false,
		"Only output when a switch occurs; silent when no .aiswitch found")
}
