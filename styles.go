package main

import (
	"github.com/charmbracelet/lipgloss"
)

var columnColours = [numStatuses]lipgloss.Color{
	StatusTodo:  "#58A6FF",
	StatusDoing: "#D29922",
	StatusDone:  "#3FB950",
}

var columnSelectionBg = [numStatuses]lipgloss.Color{
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

	BodyPreview = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B949E")).
			Italic(true)

	StatusSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3FB950"))

	StatusError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F85149")).
			Bold(true)

	StatusInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B949E"))

	Divider = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#21262D"))
)

var (
	SelectedItemStyle [numStatuses]lipgloss.Style
	SelectedBodyStyle [numStatuses]lipgloss.Style
)

func init() {
	for i := 0; i < numStatuses; i++ {
		SelectedItemStyle[i] = lipgloss.NewStyle().
			Background(columnSelectionBg[i]).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)
		SelectedBodyStyle[i] = lipgloss.NewStyle().
			Background(columnSelectionBg[i]).
			Foreground(lipgloss.Color("#8B949E"))
	}
}

func columnBorderColor(colIdx int) lipgloss.Color {
	return columnColours[colIdx]
}
