package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const maxColWidth = 40

func (m Model) View() string {
	footer := m.footerView()

	if !m.ViewportReady {
		return m.renderContent() + "\n" + footer
	}

	vp := m.Viewport
	vp.SetContent(m.renderContent())
	vp.SetYOffset(m.scrollOffset())
	return vp.View() + "\n" + footer
}

func (m Model) footerView() string {
	var parts []string
	if m.Width > 0 {
		parts = append(parts, Divider.Render(strings.Repeat("─", m.Width)))
	}
	parts = append(parts, m.statusView())
	parts = append(parts, m.helpView())
	parts = append(parts, HelpBar.Render(fmt.Sprintf("data: %s", m.SavePath)))
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) statusView() string {
	if m.StatusMsg == "" {
		return ""
	}
	var style lipgloss.Style
	switch m.StatusMsgType {
	case MsgSuccess:
		style = StatusSuccess
	case MsgError:
		style = StatusError
	default:
		style = StatusInfo
	}
	w := m.Width
	if w <= 0 {
		w = 80
	}
	return style.Render(padRight("  "+m.StatusMsg, w))
}

func (m Model) renderContent() string {
	cols := make([]string, 3)
	for i := range cols {
		cols[i] = m.renderColumn(i)
	}
	board := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

	var parts []string
	parts = append(parts, board)

	detail := m.taskDetailView()
	if detail != "" {
		if m.Width > 0 {
			parts = append(parts, Divider.Render(strings.Repeat("─", m.Width)))
		}
		parts = append(parts, detail)
	}

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
	if modeOverlay != "" {
		parts = append(parts, "", modeOverlay)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// ─── Border rendering ─────────────────────────────────────────────────────

func renderTopBorder(w int, header, scroll string, color lipgloss.Color, active bool) string {
	hw := lipgloss.Width(header)
	style := lipgloss.NewStyle().Foreground(color)

	tl, tr, hz := "┌", "┐", "─"
	if active {
		tl, tr, hz = "╔", "╗", "═"
	}

	if scroll == "" {
		fill := max(w-hw-4, 0)
		return style.Render(tl + hz + header + strings.Repeat(hz, fill) + hz + tr)
	}

	sw := lipgloss.Width(scroll)
	fill := max(w-hw-sw-8, 0)
	return style.Render(tl + hz + header + hz + hz + strings.Repeat(hz, fill) + hz + scroll + hz + hz + tr)
}

func renderBottomBorder(w int, color lipgloss.Color, active bool) string {
	bl, br, hz := "└", "┘", "─"
	if active {
		bl, br, hz = "╚", "╝", "═"
	}
	return lipgloss.NewStyle().Foreground(color).Render(
		bl + strings.Repeat(hz, w-2) + br,
	)
}

func renderContentRow(content string, borderColor lipgloss.Color, taskStyle lipgloss.Style, active bool) string {
	edge := "│"
	if active {
		edge = "║"
	}
	border := lipgloss.NewStyle().Foreground(borderColor).Render(edge)
	inner := taskStyle.Render(" " + content + " ")
	return border + inner + border
}

func renderEmptyRow(w int, borderColor lipgloss.Color, active bool) string {
	edge := "│"
	if active {
		edge = "║"
	}
	border := lipgloss.NewStyle().Foreground(borderColor).Render(edge)
	return border + strings.Repeat(" ", w-2) + border
}

// ─── Column rendering ─────────────────────────────────────────────────────

func (m Model) columnWidth(colIdx int) int {
	return m.ColWidths[colIdx]
}

func (m Model) renderColumn(colIdx int) string {
	col := m.Board.Columns[colIdx]
	w := m.columnWidth(colIdx)
	innerW := w - 4

	active := colIdx == m.Cursor.Col
	bColor := columnBorderColor(colIdx)

	header := fmt.Sprintf("(%d) %s (%d)", colIdx, col.Status.String(), len(col.Tasks))

	scroll := ""
	if active && len(col.Tasks) > 0 {
		scroll = fmt.Sprintf("%d/%d", m.Cursor.Row+1, len(col.Tasks))
	}

	top := renderTopBorder(w, header, scroll, bColor, active)

	var rows []string
	for i, task := range col.Tasks {
		title := truncate(strings.SplitN(task.Title, "\n", 2)[0], innerW-2)
		selected := active && i == m.Cursor.Row

		if selected {
			rows = append(rows, renderContentRow(
				padRight("  "+title, innerW), bColor, SelectedItemStyle(colIdx), active))
		} else {
			rows = append(rows, renderContentRow(
				padRight("  "+title, innerW), bColor, NormalTask, active))
		}

		if task.Body != "" {
			bodyLine := truncate(strings.SplitN(task.Body, "\n", 2)[0], innerW-2)
			if selected {
				rows = append(rows, renderContentRow(
					padRight("  "+bodyLine, innerW), bColor, SelectedBodyStyle(colIdx), active))
			} else {
				rows = append(rows, renderContentRow(
					padRight("  "+bodyLine, innerW), bColor, BodyPreview, active))
			}
		}

		if i < len(col.Tasks)-1 {
			rows = append(rows, renderEmptyRow(w, bColor, active))
		}
	}

	if len(col.Tasks) == 0 {
		if active {
			rows = append(rows, renderContentRow(
				padRight("  (no tasks)", innerW), bColor, SelectedItemStyle(colIdx), active))
		} else {
			rows = append(rows, renderContentRow(
				padRight("  (no tasks)", innerW), bColor, EmptyColumn, active))
		}
	}

	bottom := renderBottomBorder(w, bColor, active)

	allRows := make([]string, 0, len(rows)+2)
	allRows = append(allRows, top)
	allRows = append(allRows, rows...)
	allRows = append(allRows, bottom)

	return lipgloss.JoinVertical(lipgloss.Left, allRows...)
}

// ─── Task detail ──────────────────────────────────────────────────────────

func (m Model) taskDetailView() string {
	if m.Mode != ModeNormal {
		return ""
	}

	w := m.Width
	if w <= 0 {
		w = 80
	}

	bColor := columnBorderColor(m.Cursor.Col)

	col := m.Board.Columns[m.Cursor.Col]
	if len(col.Tasks) == 0 || m.Cursor.Row >= len(col.Tasks) {
		return renderDetailPanel("Task Detail", "", "", "", w, bColor)
	}

	task := col.Tasks[m.Cursor.Row]
	return renderDetailPanel(task.Title, task.Body,
		task.Status.String(), relativeTime(task.UpdatedAt),
		w, bColor)
}

func renderDetailPanel(title, body, status, updated string, w int, color lipgloss.Color) string {
	innerW := w - 4

	displayTitle := truncate(title, max(w-8, 10))
	top := renderTopBorder(w, displayTitle, "", color, true)

	var rows []string

	if body != "" {
		lines := strings.Split(body, "\n")
		for _, line := range lines {
			truncated := truncate(line, innerW-2)
			rows = append(rows, renderContentRow(
				padRight("  "+truncated, innerW), color, NormalTask, true))
		}
	} else if title != "" {
		rows = append(rows, renderContentRow(
			padRight("  (no description)", innerW), color, BodyPreview, true))
	} else {
		rows = append(rows, renderContentRow(
			padRight("  (no task selected)", innerW), color, BodyPreview, true))
	}

	if status != "" {
		meta := fmt.Sprintf("  Status: %s    Updated: %s", status, updated)
		rows = append(rows, renderEmptyRow(w, color, true))
		rows = append(rows, renderContentRow(
			padRight(meta, innerW), color, BodyPreview, true))
	}

	bottom := renderBottomBorder(w, color, true)

	allRows := make([]string, 0, len(rows)+2)
	allRows = append(allRows, top)
	allRows = append(allRows, rows...)
	allRows = append(allRows, bottom)

	return lipgloss.JoinVertical(lipgloss.Left, allRows...)
}

func relativeTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}

// ─── Help bar ─────────────────────────────────────────────────────────────

func (m Model) helpView() string {
	switch m.Mode {
	case ModeNormal:
		return HelpBar.Render(
			"j/k navigate  tab/h/l column  a add  e edit  d delete  enter→  shift+enter← 0/1/2 jump  q quit",
		)
	case ModeAdding:
		return HelpBar.Render(
			"enter/tab next  shift+tab back  esc cancel  alt+enter save",
		)
	case ModeEditing:
		return HelpBar.Render(
			"enter/tab next  shift+tab back  esc cancel  alt+enter save",
		)
	case ModeConfirmDelete:
		return HelpBar.Render(
			"y/enter confirm  n/esc cancel",
		)
	}
	return ""
}

// ─── Scrolling helpers ────────────────────────────────────────────────────

func (m Model) cursorLine() int {
	line := 0
	for i := 0; i < m.Cursor.Col; i++ {
		col := m.Board.Columns[i]
		line++ // header
		if len(col.Tasks) == 0 {
			line++ // (empty)
		} else {
			for _, t := range col.Tasks {
				line++ // task row
				if t.Body != "" {
					line++ // body preview
				}
				line++ // spacer between tasks
			}
			line-- // no trailing spacer
		}
	}
	col := m.Board.Columns[m.Cursor.Col]
	line++ // header
	for i := 0; i < m.Cursor.Row && i < len(col.Tasks); i++ {
		line++ // task row
		if col.Tasks[i].Body != "" {
			line++ // body preview
		}
		line++ // spacer between tasks
	}
	return line
}

func (m Model) scrollOffset() int {
	target := m.cursorLine()
	half := m.Viewport.Height / 2
	offset := target - half
	if offset < 0 {
		return 0
	}
	return offset
}

// ─── Utilities ────────────────────────────────────────────────────────────

func padRight(s string, w int) string {
	n := lipgloss.Width(s)
	if n >= w {
		return s
	}
	return s + strings.Repeat(" ", w-n)
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}
