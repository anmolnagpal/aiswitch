package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary  = lipgloss.AdaptiveColor{Light: "#5C6BC0", Dark: "#7986CB"}
	colorAccent   = lipgloss.AdaptiveColor{Light: "#00897B", Dark: "#26A69A"}
	colorMuted    = lipgloss.AdaptiveColor{Light: "#9E9E9E", Dark: "#757575"}
	colorSuccess  = lipgloss.AdaptiveColor{Light: "#388E3C", Dark: "#66BB6A"}
	colorWarning  = lipgloss.AdaptiveColor{Light: "#F57C00", Dark: "#FFA726"}
	colorDanger   = lipgloss.AdaptiveColor{Light: "#D32F2F", Dark: "#EF5350"}
	colorSelected = lipgloss.AdaptiveColor{Light: "#1565C0", Dark: "#90CAF9"}

	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 2).
			MarginBottom(1)

	StyleActiveTag = lipgloss.NewStyle().
			Background(colorAccent).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)

	StyleMuted = lipgloss.NewStyle().Foreground(colorMuted)

	StyleSelected = lipgloss.NewStyle().
			Foreground(colorSelected).
			Bold(true)

	StyleSuccess = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	StyleWarning = lipgloss.NewStyle().Foreground(colorWarning)

	StyleDanger = lipgloss.NewStyle().Foreground(colorDanger)

	StyleHint = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	StyleBox = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(0, 1)

	StyleServiceBadge = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Padding(0, 1)
)
