package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/anmolnagpal/aiswitch/internal/config"
	"github.com/anmolnagpal/aiswitch/internal/localfile"
	"github.com/anmolnagpal/aiswitch/internal/ui"
)

var localInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a .aiswitch file in the current directory",
	Long: `Create a .aiswitch file that pins this project to a specific profile.

Commit this file to your repo — it contains no secrets, only the profile
name and optional per-project overrides.

When shell integration is active (eval "$(aiswitch shell-init)"), the profile
is switched automatically every time you cd into this directory.

This mirrors how .terraform-version works with tfswitch, or .nvmrc with nvm.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if len(cfg.Profiles) == 0 {
			return fmt.Errorf("no profiles configured — run: aiswitch add")
		}

		names := make([]string, 0, len(cfg.Profiles))
		for name := range cfg.Profiles {
			names = append(names, name)
		}
		sort.Strings(names)

		// Pre-fill from existing .aiswitch if present.
		var (
			profileName string
			claudeModel string
			githubEmail string
		)
		cwd, _ := os.Getwd()
		existing, _ := localfile.Read(filepath.Join(cwd, localfile.FileName))
		if existing != nil {
			profileName = existing.Profile
			if existing.Claude != nil {
				claudeModel = existing.Claude.Model
			}
			if existing.GitHub != nil {
				githubEmail = existing.GitHub.Email
			}
		}

		// Build profile select options.
		options := make([]huh.Option[string], len(names))
		for i, n := range names {
			p := cfg.Profiles[n]
			label := n + "  " + ui.StyleMuted.Render("("+p.Services()+")")
			options[i] = huh.NewOption(label, n)
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Profile for this project").
					Description("Written to .aiswitch in the current directory — no secrets inside").
					Options(options...).
					Value(&profileName),

				huh.NewInput().
					Title("Claude model override").
					Description("Optional — overrides the profile's default model for this project only").
					Placeholder("claude-opus-4-5").
					Value(&claudeModel),

				huh.NewInput().
					Title("Git commit email override").
					Description("Optional — overrides global git user.email for this project only").
					Placeholder("me@project.com").
					Value(&githubEmail),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		localCfg := &localfile.LocalConfig{Profile: profileName}
		if claudeModel != "" {
			localCfg.Claude = &localfile.ClaudeOverride{Model: claudeModel}
		}
		if githubEmail != "" {
			localCfg.GitHub = &localfile.GitHubOverride{Email: githubEmail}
		}

		path := filepath.Join(cwd, localfile.FileName)
		if err := localfile.Write(path, localCfg); err != nil {
			return err
		}

		fmt.Println()
		fmt.Println(ui.StyleSuccess.Render("✓ Created .aiswitch"))
		fmt.Println(ui.StyleMuted.Render("  Profile : " + profileName))
		if claudeModel != "" {
			fmt.Println(ui.StyleMuted.Render("  Model   : " + claudeModel))
		}
		if githubEmail != "" {
			fmt.Println(ui.StyleMuted.Render("  Email   : " + githubEmail))
		}
		fmt.Println()
		fmt.Println(ui.StyleHint.Render("  Commit .aiswitch to your repo (no secrets inside)"))
		fmt.Println(ui.StyleHint.Render("  Apply now: aiswitch detect"))
		return nil
	},
}
