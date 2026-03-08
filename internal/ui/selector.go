// Package ui provides the interactive profile selector TUI built with
// Bubble Tea and Lip Gloss.
package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/anmolnagpal/aiswitch/internal/config"
)

// ─── list.Item implementation ────────────────────────────────────────────────

type profileItem struct {
	name    string
	profile config.Profile
	active  bool
}

func (p profileItem) FilterValue() string { return p.name }
func (p profileItem) Title() string       { return p.name }
func (p profileItem) Description() string { return p.profile.Services() }

// ─── custom item delegate ─────────────────────────────────────────────────────

type profileDelegate struct{}

func (d profileDelegate) Height() int                             { return 2 }
func (d profileDelegate) Spacing() int                            { return 1 }
func (d profileDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d profileDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(profileItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	cursor := "  "
	nameStyle := StyleMuted
	descStyle := StyleMuted

	if isSelected {
		cursor = StyleSelected.Render("▶ ")
		nameStyle = StyleSelected
		descStyle = lipgloss.NewStyle().Foreground(colorMuted)
	}

	activeMark := "○ "
	if item.active {
		activeMark = StyleSuccess.Render("● ")
	}

	name := nameStyle.Render(item.name)
	desc := descStyle.Render(item.profile.Services())

	activeTag := ""
	if item.active {
		activeTag = " " + StyleActiveTag.Render("active")
	}

	descTag := ""
	if item.profile.Description != "" {
		descTag = StyleHint.Render("  " + item.profile.Description)
	}

	_, _ = fmt.Fprintf(w, "%s%s%s%s\n  %s%s",
		cursor,
		activeMark,
		name,
		activeTag,
		desc,
		descTag,
	)
}

// ─── Bubble Tea model ─────────────────────────────────────────────────────────

type selectorModel struct {
	list     list.Model
	chosen   string
	quitting bool
}

func newSelectorModel(profiles map[string]config.Profile, active string) selectorModel {
	items := make([]list.Item, 0, len(profiles))
	for name, p := range profiles {
		items = append(items, profileItem{
			name:    name,
			profile: p,
			active:  name == active,
		})
	}

	l := list.New(items, profileDelegate{}, 60, min(len(items)*3+6, 20))
	l.Title = "Switch AI Profile"
	l.Styles.Title = StyleHeader.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorPrimary).
		Foreground(colorPrimary).
		Bold(true).
		Padding(0, 1)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	// Select current active item.
	for i, it := range items {
		if it.(profileItem).name == active {
			l.Select(i)
			break
		}
	}

	return selectorModel{list: l}
}

func (m selectorModel) Init() tea.Cmd { return nil }

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(profileItem); ok {
				m.chosen = item.name
			}
			return m, tea.Quit
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectorModel) View() string {
	if m.chosen != "" || m.quitting {
		return ""
	}
	hint := StyleHint.Render("\n  ↑/↓ navigate  •  enter select  •  / filter  •  q quit\n")
	return "\n" + m.list.View() + hint
}

// RunSelector shows the interactive profile list and returns the selected
// profile name, or an empty string if the user quit without selecting.
func RunSelector(profiles map[string]config.Profile, active string) (string, error) {
	if len(profiles) == 0 {
		return "", fmt.Errorf("no profiles configured — run: aiswitch add")
	}

	m, err := tea.NewProgram(newSelectorModel(profiles, active)).Run()
	if err != nil {
		return "", err
	}

	result := m.(selectorModel)
	if result.quitting {
		return "", nil
	}
	return result.chosen, nil
}

// PrintSwitchResult prints a styled summary after switching profiles.
func PrintSwitchResult(name string, profile config.Profile, envPath string) {
	lines := []string{
		StyleSuccess.Render("✓ Switched to \"" + name + "\""),
		"",
	}

	if profile.Claude != nil {
		lines = append(lines,
			"  "+StyleServiceBadge.Render("Claude")+"  API key set",
		)
	}
	if profile.GitHub != nil {
		lines = append(lines,
			"  "+StyleServiceBadge.Render("GitHub")+"  @"+profile.GitHub.Username,
		)
	}

	lines = append(lines,
		"",
		StyleHint.Render("  ℹ  Apply env vars in this shell:"),
		StyleMuted.Render("       source "+envPath),
		"",
		StyleHint.Render("  ℹ  Auto-apply in every new shell — add to ~/.zshrc or ~/.bashrc:"),
		StyleMuted.Render(`       [ -f `+envPath+` ] && source `+envPath),
	)

	fmt.Println(strings.Join(lines, "\n"))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
