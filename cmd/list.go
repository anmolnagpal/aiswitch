package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"

	"github.com/anmolnagpal/aiswitch/internal/config"
	"github.com/anmolnagpal/aiswitch/internal/ui"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all configured profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Profiles) == 0 {
			fmt.Println(ui.StyleMuted.Render("No profiles yet. Run: aiswitch add"))
			return nil
		}

		names := make([]string, 0, len(cfg.Profiles))
		for name := range cfg.Profiles {
			names = append(names, name)
		}
		sort.Strings(names)

		rows := make([][]string, 0, len(names))
		for _, name := range names {
			p := cfg.Profiles[name]

			status := ""
			if name == cfg.ActiveProfile {
				status = ui.StyleActiveTag.Render("active")
			}

			claudeInfo := ui.StyleMuted.Render("—")
			if p.Claude != nil {
				key := p.Claude.APIKey
				masked := maskSecret(key)
				claudeInfo = masked
				if p.Claude.DefaultModel != "" {
					claudeInfo += ui.StyleMuted.Render("  " + p.Claude.DefaultModel)
				}
			}

			ghInfo := ui.StyleMuted.Render("—")
			if p.GitHub != nil {
				ghInfo = "@" + p.GitHub.Username
			}

			desc := ui.StyleMuted.Render(p.Description)

			rows = append(rows, []string{
				name,
				status,
				claudeInfo,
				ghInfo,
				desc,
			})
		}

		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(ui.StyleMuted.GetForeground())).
			Headers("PROFILE", "STATUS", "CLAUDE KEY", "GITHUB", "DESCRIPTION").
			Rows(rows...).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return lipgloss.NewStyle().
						Foreground(lipgloss.Color("#7986CB")).
						Bold(true).
						Padding(0, 1)
				}
				base := lipgloss.NewStyle().Padding(0, 1)
				if col == 0 {
					base = base.Bold(true)
				}
				return base
			})

		fmt.Println(t)
		fmt.Println(ui.StyleHint.Render("  " + fmt.Sprintf("%d profile(s)  •  aiswitch add  •  aiswitch use <profile>", len(names))))
		return nil
	},
}

// maskSecret returns a masked version: first 8 chars visible + "..." + last 4.
func maskSecret(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 12 {
		return strings.Repeat("*", len(s))
	}
	return s[:8] + "..." + s[len(s)-4:]
}
