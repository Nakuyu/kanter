package main

import (
	"github.com/charmbracelet/lipgloss"
)

var columnColours = [3]lipgloss.Color{
	StatusTodo:  "#58A6FF",
	StatusDoing: "#D29922",
	StatusDone:  "#3FB950",
}

var columnSelectionBg = [3]lipgloss.Color{
	StatusTodo:  "#174E9B",
	StatusDoing: "#7B5E00",
	StatusDone:  "#1A6B2F",
}

var (
	NormalTask = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C9D1D9"))

	EmptyColumn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#484F58")).
			Italic(true)

	HelpBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B949E"))

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

	Divider = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#21262D"))
)

func SelectedItemStyle(colIdx int) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(columnSelectionBg[colIdx]).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
}

func SelectedBodyStyle(colIdx int) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(columnSelectionBg[colIdx]).
		Foreground(lipgloss.Color("#8B949E"))
}

func ColumnBorder(colIdx int, active bool) lipgloss.Style {
	color := columnColours[colIdx]
	if !active {
		color = "#21262D"
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(0, 1).
		Width(30)
}

func ColumnHeader(colIdx int) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(columnColours[colIdx])
}
