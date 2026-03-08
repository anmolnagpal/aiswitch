package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/anmolnagpal/aiswitch/internal/config"
	"github.com/anmolnagpal/aiswitch/internal/providers/claude"
	"github.com/anmolnagpal/aiswitch/internal/providers/gemini"
	"github.com/anmolnagpal/aiswitch/internal/providers/github"
	"github.com/anmolnagpal/aiswitch/internal/providers/ide"
	"github.com/anmolnagpal/aiswitch/internal/providers/openai"
	"github.com/anmolnagpal/aiswitch/internal/ui"
)

var currentCmd = &cobra.Command{
	Use:     "current",
	Aliases: []string{"status"},
	Short:   "Show the currently active profile and live system state",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		fmt.Println()

		// ── Active profile ────────────────────────────────────────────────────
		if cfg.ActiveProfile == "" {
			fmt.Println(ui.StyleWarning.Render("  No active profile"))
			fmt.Println(ui.StyleHint.Render("  Run: aiswitch  to select a profile"))
		} else {
			p, ok := cfg.Profiles[cfg.ActiveProfile]
			if !ok {
				fmt.Println(ui.StyleDanger.Render("  Active profile \"" + cfg.ActiveProfile + "\" no longer exists"))
			} else {
				fmt.Println(ui.StyleSuccess.Render("  Active profile: ") + ui.StyleSelected.Render(cfg.ActiveProfile))
				if p.Description != "" {
					fmt.Println(ui.StyleMuted.Render("  " + p.Description))
				}

				fmt.Println()
				printProfileDetail(p)
			}
		}

		// ── Live system state ─────────────────────────────────────────────────
		fmt.Println()
		fmt.Println(ui.StyleMuted.Render("  ─── live system state ───────────────────────────"))
		fmt.Println()

		claudeKey := claude.Detect()
		if claudeKey != "" {
			fmt.Println("  " + ui.StyleServiceBadge.Render("Claude") + "  " + maskSecret(claudeKey))
		} else {
			fmt.Println("  " + ui.StyleServiceBadge.Render("Claude") + "  " + ui.StyleMuted.Render("not detected"))
		}

		openAIKey := openai.Detect()
		if openAIKey != "" {
			fmt.Println("  " + ui.StyleServiceBadge.Render("OpenAI") + "  " + maskSecret(openAIKey))
		} else {
			fmt.Println("  " + ui.StyleServiceBadge.Render("OpenAI") + "  " + ui.StyleMuted.Render("not detected"))
		}

		geminiKey := gemini.Detect()
		if geminiKey != "" {
			fmt.Println("  " + ui.StyleServiceBadge.Render("Gemini") + "  " + maskSecret(geminiKey))
		} else {
			fmt.Println("  " + ui.StyleServiceBadge.Render("Gemini") + "  " + ui.StyleMuted.Render("not detected"))
		}

		ghUser := github.Detect()
		if ghUser != "" {
			fmt.Println("  " + ui.StyleServiceBadge.Render("GitHub") + "  @" + ghUser)
		} else {
			fmt.Println("  " + ui.StyleServiceBadge.Render("GitHub") + "  " + ui.StyleMuted.Render("not detected"))
		}

		for _, ideName := range []string{"Cursor", "Windsurf"} {
			status := ide.DetectIDE(ideName)
			badge := ui.StyleServiceBadge.Render(ideName)
			if status != "" {
				fmt.Println("  " + badge + "  " + ui.StyleMuted.Render("settings.json patched · ") + status)
			} else {
				fmt.Println("  " + badge + "  " + ui.StyleMuted.Render("not installed or not yet patched"))
			}
		}

		fmt.Println()
		return nil
	},
}

func printProfileDetail(p config.Profile) {
	if p.Claude != nil {
		fmt.Println("  " + ui.StyleServiceBadge.Render("Claude"))
		fmt.Println(ui.StyleMuted.Render("    API Key  ") + maskSecret(p.Claude.APIKey))
		if p.Claude.DefaultModel != "" {
			fmt.Println(ui.StyleMuted.Render("    Model    ") + p.Claude.DefaultModel)
		}
		fmt.Println()
	}

	if p.OpenAI != nil {
		fmt.Println("  " + ui.StyleServiceBadge.Render("OpenAI"))
		fmt.Println(ui.StyleMuted.Render("    API Key  ") + maskSecret(p.OpenAI.APIKey))
		if p.OpenAI.OrgID != "" {
			fmt.Println(ui.StyleMuted.Render("    Org ID   ") + p.OpenAI.OrgID)
		}
		if p.OpenAI.DefaultModel != "" {
			fmt.Println(ui.StyleMuted.Render("    Model    ") + p.OpenAI.DefaultModel)
		}
		fmt.Println()
	}

	if p.Gemini != nil {
		fmt.Println("  " + ui.StyleServiceBadge.Render("Gemini"))
		fmt.Println(ui.StyleMuted.Render("    API Key  ") + maskSecret(p.Gemini.APIKey))
		if p.Gemini.ProjectID != "" {
			fmt.Println(ui.StyleMuted.Render("    Project  ") + p.Gemini.ProjectID)
		}
		if p.Gemini.DefaultModel != "" {
			fmt.Println(ui.StyleMuted.Render("    Model    ") + p.Gemini.DefaultModel)
		}
		fmt.Println()
	}

	if p.GitHub != nil {
		fmt.Println("  " + ui.StyleServiceBadge.Render("GitHub"))
		fmt.Println(ui.StyleMuted.Render("    Username  ") + "@" + p.GitHub.Username)
		if p.GitHub.Email != "" {
			fmt.Println(ui.StyleMuted.Render("    Email     ") + p.GitHub.Email)
		}
		fmt.Println(ui.StyleMuted.Render("    Token     ") + maskSecret(p.GitHub.Token))
		fmt.Println()
	}

	if p.IDE != nil {
		fmt.Println("  " + ui.StyleServiceBadge.Render("IDE patching"))
		if p.IDE.Cursor {
			fmt.Println(ui.StyleMuted.Render("    Cursor    ") + "enabled")
		}
		if p.IDE.Windsurf {
			fmt.Println(ui.StyleMuted.Render("    Windsurf  ") + "enabled")
		}
		fmt.Println()
	}

}
