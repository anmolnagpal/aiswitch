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
			services    []string // "claude" | "openai" | "gemini" | "github"

			claudeAPIKey string
			claudeModel  string

			openAIAPIKey string
			openAIOrgID  string
			openAIModel  string

			geminiAPIKey  string
			geminiProject string
			geminiModel   string

			ghToken    string
			ghUsername string
			ghEmail    string
		)

		if len(args) == 1 {
			profileName = args[0]
		}

		// Pre-fill from existing profile when updating.
		if profileName != "" {
			if existing, ok := cfg.Profiles[profileName]; ok {
				description = existing.Description
				if existing.Claude != nil {
					claudeAPIKey = existing.Claude.APIKey
					claudeModel = existing.Claude.DefaultModel
				}
				if existing.OpenAI != nil {
					openAIAPIKey = existing.OpenAI.APIKey
					openAIOrgID = existing.OpenAI.OrgID
					openAIModel = existing.OpenAI.DefaultModel
				}
				if existing.Gemini != nil {
					geminiAPIKey = existing.Gemini.APIKey
					geminiProject = existing.Gemini.ProjectID
					geminiModel = existing.Gemini.DefaultModel
				}
				if existing.GitHub != nil {
					ghToken = existing.GitHub.Token
					ghUsername = existing.GitHub.Username
					ghEmail = existing.GitHub.Email
				}
			}
		}

		// ── Step 1: name, description, service selection ──────────────────────
		nameForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Profile name").
					Description("e.g. work, personal, client-x").
					Placeholder("work").
					Value(&profileName).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
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
					Title("Which AI providers to configure?").
					Options(
						huh.NewOption("Claude  (Anthropic)", "claude"),
						huh.NewOption("OpenAI  (GPT-4o, o1, o3…)", "openai"),
						huh.NewOption("Gemini  (Google AI / Vertex AI)", "gemini"),
						huh.NewOption("GitHub  (Copilot + gh CLI)", "github"),
					).
					Value(&services).
					Validate(func(s []string) error {
						if len(s) == 0 {
							return fmt.Errorf("select at least one provider")
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
		wantOpenAI := contains(services, "openai")
		wantGemini := contains(services, "gemini")
		wantGitHub := contains(services, "github")

		// ── Step 2: Claude ────────────────────────────────────────────────────
		if wantClaude {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Anthropic API Key").
						Description("console.anthropic.com → API Keys").
						Placeholder("sk-ant-api03-...").
						EchoMode(huh.EchoModePassword).
						Value(&claudeAPIKey).
						Validate(func(s string) error {
							if strings.TrimSpace(s) == "" {
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
			if err := form.Run(); err != nil {
				return err
			}
		}

		// ── Step 3: OpenAI ────────────────────────────────────────────────────
		if wantOpenAI {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("OpenAI API Key").
						Description("platform.openai.com → API keys").
						Placeholder("sk-proj-...").
						EchoMode(huh.EchoModePassword).
						Value(&openAIAPIKey).
						Validate(func(s string) error {
							if strings.TrimSpace(s) == "" {
								return fmt.Errorf("API key cannot be empty")
							}
							return nil
						}),

					huh.NewInput().
						Title("Organisation ID").
						Description("Optional — platform.openai.com → Settings → Organisation ID").
						Placeholder("org-...").
						Value(&openAIOrgID),

					huh.NewInput().
						Title("Default model").
						Description("Optional — e.g. gpt-4o, gpt-4o-mini, o3").
						Placeholder("gpt-4o").
						Value(&openAIModel),
				),
			)
			if err := form.Run(); err != nil {
				return err
			}
		}

		// ── Step 4: Gemini ────────────────────────────────────────────────────
		if wantGemini {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Gemini API Key").
						Description("aistudio.google.com → Get API key  (leave blank for Vertex AI ADC)").
						Placeholder("AIza...").
						EchoMode(huh.EchoModePassword).
						Value(&geminiAPIKey).
						Validate(func(s string) error {
							if strings.TrimSpace(s) == "" {
								return fmt.Errorf("API key cannot be empty")
							}
							return nil
						}),

					huh.NewInput().
						Title("Default model").
						Description("Optional — e.g. gemini-2.0-flash, gemini-1.5-pro").
						Placeholder("gemini-2.0-flash").
						Value(&geminiModel),

					huh.NewInput().
						Title("Google Cloud Project ID").
						Description("Optional — only needed for Vertex AI").
						Placeholder("my-project-123").
						Value(&geminiProject),
				),
			)
			if err := form.Run(); err != nil {
				return err
			}
		}

		// ── Step 5: GitHub ────────────────────────────────────────────────────
		if wantGitHub {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("GitHub Personal Access Token").
						Description("github.com/settings/tokens — needs repo, read:user, copilot scopes").
						Placeholder("ghp_...").
						EchoMode(huh.EchoModePassword).
						Value(&ghToken).
						Validate(func(s string) error {
							if strings.TrimSpace(s) == "" {
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
			if err := form.Run(); err != nil {
				return err
			}
		}

		// ── Build & save ──────────────────────────────────────────────────────
		profile := config.Profile{
			Description: strings.TrimSpace(description),
		}
		if wantClaude {
			profile.Claude = &config.ClaudeConfig{
				APIKey:       strings.TrimSpace(claudeAPIKey),
				DefaultModel: strings.TrimSpace(claudeModel),
			}
		}
		if wantOpenAI {
			profile.OpenAI = &config.OpenAIConfig{
				APIKey:       strings.TrimSpace(openAIAPIKey),
				OrgID:        strings.TrimSpace(openAIOrgID),
				DefaultModel: strings.TrimSpace(openAIModel),
			}
		}
		if wantGemini {
			profile.Gemini = &config.GeminiConfig{
				APIKey:       strings.TrimSpace(geminiAPIKey),
				ProjectID:    strings.TrimSpace(geminiProject),
				DefaultModel: strings.TrimSpace(geminiModel),
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
		fmt.Println(ui.StyleSuccess.Render("✓ Profile \"" + profileName + "\" saved"))
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
