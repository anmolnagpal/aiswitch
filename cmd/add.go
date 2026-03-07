package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/anmolnagpal/aiswitch/internal/config"
	"github.com/anmolnagpal/aiswitch/internal/ui"
)

var addCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add or update a profile",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		var (
			profileName string
			description string
			services    []string // "claude" | "github"

			claudeAPIKey string
			claudeModel  string

			ghToken    string
			ghUsername string
			ghEmail    string
		)

		if len(args) == 1 {
			profileName = args[0]
		}

		// ── Step 1: name & description ────────────────────────────────────────
		nameForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Profile name").
					Description("e.g. work, personal, client-x").
					Placeholder("work").
					Value(&profileName).
					Validate(func(s string) error {
						s = strings.TrimSpace(s)
						if s == "" {
							return fmt.Errorf("name cannot be empty")
						}
						return nil
					}),

				huh.NewInput().
					Title("Description").
					Description("Optional — shown in the profile list").
					Placeholder("Day-job accounts").
					Value(&description),

				huh.NewMultiSelect[string]().
					Title("Which services to configure?").
					Options(
						huh.NewOption("Claude / Anthropic", "claude"),
						huh.NewOption("GitHub / Copilot", "github"),
					).
					Value(&services).
					Validate(func(s []string) error {
						if len(s) == 0 {
							return fmt.Errorf("select at least one service")
						}
						return nil
					}),
			),
		)

		if err := nameForm.Run(); err != nil {
			return err
		}

		profileName = strings.TrimSpace(profileName)
		wantClaude := contains(services, "claude")
		wantGitHub := contains(services, "github")

		// ── Step 2: Claude credentials ────────────────────────────────────────
		if wantClaude {
			// Pre-fill from existing profile if updating.
			if existing, ok := cfg.Profiles[profileName]; ok && existing.Claude != nil {
				claudeAPIKey = existing.Claude.APIKey
				claudeModel = existing.Claude.DefaultModel
			}

			claudeForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Anthropic API Key").
						Description("Starts with sk-ant-  —  get it from console.anthropic.com").
						Placeholder("sk-ant-api03-...").
						EchoMode(huh.EchoModePassword).
						Value(&claudeAPIKey).
						Validate(func(s string) error {
							s = strings.TrimSpace(s)
							if s == "" {
								return fmt.Errorf("API key cannot be empty")
							}
							return nil
						}),

					huh.NewInput().
						Title("Default model").
						Description("Optional — e.g. claude-opus-4-5, claude-sonnet-4-5").
						Placeholder("claude-opus-4-5").
						Value(&claudeModel),
				),
			)
			if err := claudeForm.Run(); err != nil {
				return err
			}
		}

		// ── Step 3: GitHub credentials ────────────────────────────────────────
		if wantGitHub {
			if existing, ok := cfg.Profiles[profileName]; ok && existing.GitHub != nil {
				ghToken = existing.GitHub.Token
				ghUsername = existing.GitHub.Username
				ghEmail = existing.GitHub.Email
			}

			ghForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("GitHub Personal Access Token").
						Description("Create at github.com/settings/tokens — needs repo, user, copilot scopes").
						Placeholder("ghp_...").
						EchoMode(huh.EchoModePassword).
						Value(&ghToken).
						Validate(func(s string) error {
							s = strings.TrimSpace(s)
							if s == "" {
								return fmt.Errorf("token cannot be empty")
							}
							return nil
						}),

					huh.NewInput().
						Title("GitHub username").
						Placeholder("octocat").
						Value(&ghUsername).
						Validate(func(s string) error {
							if strings.TrimSpace(s) == "" {
								return fmt.Errorf("username cannot be empty")
							}
							return nil
						}),

					huh.NewInput().
						Title("Git commit email").
						Description("Optional — updates global git config user.email when you switch").
						Placeholder("you@example.com").
						Value(&ghEmail),
				),
			)
			if err := ghForm.Run(); err != nil {
				return err
			}
		}

		// ── Build & save profile ───────────────────────────────────────────────
		profile := config.Profile{
			Description: strings.TrimSpace(description),
		}
		if wantClaude {
			profile.Claude = &config.ClaudeConfig{
				APIKey:       strings.TrimSpace(claudeAPIKey),
				DefaultModel: strings.TrimSpace(claudeModel),
			}
		}
		if wantGitHub {
			profile.GitHub = &config.GitHubConfig{
				Token:    strings.TrimSpace(ghToken),
				Username: strings.TrimSpace(ghUsername),
				Email:    strings.TrimSpace(ghEmail),
			}
		}

		if cfg.Profiles == nil {
			cfg.Profiles = map[string]config.Profile{}
		}
		cfg.Profiles[profileName] = profile

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Println()
		fmt.Println(ui.StyleSuccess.Render("✓ Profile \""+profileName+"\" saved"))
		fmt.Println(ui.StyleHint.Render("  Run: aiswitch use " + profileName))
		return nil
	},
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
