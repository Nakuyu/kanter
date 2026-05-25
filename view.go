package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

const maxColWidth = 40

func (m Model) View() string {
	cols := make([]string, 3)
	for i := range cols {
		cols[i] = m.renderColumn(i)
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

	detail := m.taskDetailView()

	var modeOverlay string
	switch m.Mode {
	case ModeAdding, ModeEditing:
		var label, input string
		if m.AddStage == 0 {
			label = PromptStyle.Render("Title:")
			input = m.TitleInput.View()
		} else {
			label = PromptStyle.Render("Body:")
			input = m.BodyTextarea.View()
		}
		modeOverlay = label + "\n" + input
	case ModeConfirmDelete:
		col := m.Board.Columns[m.Cursor.Col]
		if m.Cursor.Row < len(col.Tasks) {
			task := col.Tasks[m.Cursor.Row]
			modeOverlay = ConfirmStyle.Render(
				fmt.Sprintf("Delete %q? (y/enter = yes, n/esc = no)", task.Title),
			)
		}
	}

	var errBar string
	if m.Err != nil {
		errBar = ErrorBar.Render(fmt.Sprintf("Error: %v", m.Err))
	}

	saveLine := HelpBar.Render(fmt.Sprintf("data: %s", m.SavePath))

	parts := []string{board}
	if detail != "" {
		parts = append(parts, detail)
	}
	if modeOverlay != "" {
		parts = append(parts, "", modeOverlay)
	}
	parts = append(parts, "", m.helpView(), saveLine)
	if errBar != "" {
		parts = append(parts, errBar)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) columnWidth(colIdx int) int {
	base := 30
	for _, task := range m.Board.Columns[colIdx].Tasks {
		w := len([]rune(task.Title)) + 4
		if w > base {
			base = w
		}
	}
	if base > maxColWidth {
		base = maxColWidth
	}
	return base
}

func (m Model) renderColumn(colIdx int) string {
	col := m.Board.Columns[colIdx]
	label := fmt.Sprintf("(%d) %s", colIdx, col.Status.String())

	var header string
	if colIdx == m.Cursor.Col {
		header = SelectedItem.Render(label)
	} else {
		header = ColumnHeader(colIdx).Render(label)
	}

	innerW := m.columnWidth(colIdx) - 4
	maxTitle := min(innerW, 36)

	rows := []string{header}
	for i, task := range col.Tasks {
		prefix := "  "
		style := NormalTask
		if colIdx == m.Cursor.Col && i == m.Cursor.Row {
			prefix = "> "
			style = SelectedItem
		}
		title := truncate(task.Title, maxTitle)
		row := style.Render(prefix + title)
		if task.Body != "" {
			bodyLine := truncate(task.Body, innerW)
			row += "\n" + BodyPreview.Render("  " + bodyLine)
		}
		rows = append(rows, row)
	}

	if len(col.Tasks) == 0 {
		rows = append(rows, EmptyColumn.Render("(empty)"))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return ColumnBorder(colIdx).Width(m.columnWidth(colIdx)).Render(content)
}

func (m Model) taskDetailView() string {
	if m.Mode != ModeNormal {
		return ""
	}
	col := m.Board.Columns[m.Cursor.Col]
	if len(col.Tasks) == 0 || m.Cursor.Row >= len(col.Tasks) {
		return ""
	}
	task := col.Tasks[m.Cursor.Row]

	title := SelectedItem.Render(" " + task.Title + " ")
	body := task.Body
	if body == "" {
		body = "(no description)"
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		BodyPreview.Render(body),
	)
}

func (m Model) helpView() string {
	switch m.Mode {
	case ModeNormal:
		return HelpBar.Render(
			"j/k navigate  h/l column  a add  e edit  d delete  enter move  0/1/2 jump  q quit",
		)
	case ModeAdding:
		return HelpBar.Render(
			"enter confirm  esc cancel",
		)
	case ModeEditing:
		return HelpBar.Render(
			"enter confirm  esc cancel",
		)
	case ModeConfirmDelete:
		return HelpBar.Render(
			"y/enter confirm  n/esc cancel",
		)
	}
	return ""
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
