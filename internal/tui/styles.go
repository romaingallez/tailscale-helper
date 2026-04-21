package tui

import "github.com/charmbracelet/lipgloss"

type styles struct {
	Title        lipgloss.Style
	Status       lipgloss.Style
	Help         lipgloss.Style
	ListPanel    lipgloss.Style
	DetailsPanel lipgloss.Style
	PanelTitle   lipgloss.Style
	Section      lipgloss.Style
	ListItem     lipgloss.Style
	ListSelected lipgloss.Style
	Muted        lipgloss.Style
	Group        lipgloss.Style
}

func defaultStyles() styles {
	border := lipgloss.NormalBorder()
	return styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("62")).
			Padding(0, 1),
		Status: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("236")).
			Padding(0, 1),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Padding(0, 1),
		ListPanel: lipgloss.NewStyle().
			Border(border).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1),
		DetailsPanel: lipgloss.NewStyle().
			Border(border).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
		PanelTitle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")),
		Section:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("111")),
		ListItem:   lipgloss.NewStyle(),
		ListSelected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("57")),
		Muted: lipgloss.NewStyle().Foreground(lipgloss.Color("242")),
		Group: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")),
	}
}
