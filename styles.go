package main

import (
	"github.com/charmbracelet/lipgloss"
)

var columnColours = [3]lipgloss.Color{
	StatusTodo:  "#58A6FF",
	StatusDoing: "#D29922",
	StatusDone:  "#3FB950",
}

var (
	SelectedItem = lipgloss.NewStyle().
			Background(lipgloss.Color("#1F6FEB")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	NormalTask = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C9D1D9"))

	EmptyColumn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#484F58")).
			Italic(true)

	HelpBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B949E"))

	HelpBarKey = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#58A6FF")).
			Bold(true)

	PromptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#58A6FF")).
			Bold(true)

	ConfirmStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D29922")).
			Bold(true)

	ErrorBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F85149")).
			Bold(true)

	BodyPreview = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#484F58"))
)

func ColumnBorder(colIdx int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(columnColours[colIdx]).
		Padding(0, 1).
		Width(30)
}

func ColumnHeader(colIdx int) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(columnColours[colIdx])
}
