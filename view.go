package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if !m.ViewportReady {
		return m.renderBoardContent() + "\n" + m.renderFooter()
	}

	boardContent := m.renderBoardContent()

	// In normal mode, i m splitting header line 0 and bottom border
	// from viewport so the columns or double pillars (inactive single pillars)
	// stay framed and look together, in other modes overlay is appened below the board so
	// only splitting the header there.
	if m.Mode == ModeNormal {
		allLines := strings.Split(boardContent, "\n")
		if len(allLines) >= 3 {
			header := allLines[0]
			bottomBorder := allLines[len(allLines)-1]
			body := strings.Join(allLines[1:len(allLines)-1], "\n")

			m.Viewport.SetContent(body)
			m.Viewport.SetYOffset(m.scrollOffset())

			parts := []string{header}
			parts = append(parts, m.Viewport.View())
			parts = append(parts, bottomBorder)
			if m.Layout.DetailMode != DetailHidden {
				parts = append(parts, m.renderDetailOutside())
			}
			parts = append(parts, m.renderFooter())
			return lipgloss.JoinVertical(lipgloss.Left, parts...)
		}
	}

	lines := strings.SplitN(boardContent, "\n", 2)
	header := lines[0]
	body := ""
	if len(lines) > 1 {
		body = lines[1]
	}

	m.Viewport.SetContent(body)
	m.Viewport.SetYOffset(m.scrollOffset())

	parts := []string{header}
	parts = append(parts, m.Viewport.View())

	if m.Layout.DetailMode != DetailHidden {
		parts = append(parts, m.renderDetailOutside())
	}

	parts = append(parts, m.renderFooter())

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) renderFooter() string {
	w := m.Width
	divider := Divider.Render(strings.Repeat("─", w))
	status := m.statusLine()
	help := m.helpView()
	// data line replaced with open space so add/edit feels less cramped
	return lipgloss.JoinVertical(lipgloss.Left, divider, status, help, "")
}

func (m Model) statusLine() string {
	w := m.Width
	if w <= 0 {
		w = 80
	}
	if m.StatusMsg == "" {
		return strings.Repeat(" ", w)
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
	return style.Render(padRight("  "+m.StatusMsg, w))
}

func (m Model) renderBoardContent() string {
	targetInnerRows := 0
	for i := range m.Board.Columns {
		if m.Layout.ColWidths[i] == 0 {
			continue
		}
		if n := m.columnInnerRowCount(i); n > targetInnerRows {
			targetInnerRows = n
		}
	}
	if m.Mode == ModeNormal {
		// Fill the entire viewport height so there's no gap between the last
		// content row and the sticky bottom border. The board has targetInnerRows
		// content rows (all with column borders) sandwiched between the header
		// and bottom border; it's just not a fun sandwich, that's all.
		if avail := m.Layout.ViewportHeight; avail > targetInnerRows {
			targetInnerRows = avail
		}
	}

	var cols []string
	for i := range m.Board.Columns {
		if m.Layout.ColWidths[i] == 0 {
			continue
		}
		cols = append(cols, m.renderColumn(i, targetInnerRows))
	}
	board := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

	var parts []string
	parts = append(parts, board)

	var modeOverlay string
	switch m.Mode {
	case ModeAdding, ModeEditing:
		var label, input string
		if m.FormStage == formStageTitle {
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

func (m Model) columnWidth(colIdx int) int {
	return m.Layout.ColWidths[colIdx]
}

func (m Model) columnInnerRowCount(colIdx int) int {
	col := m.Board.Columns[colIdx]
	if len(col.Tasks) == 0 {
		return 1
	}
	n := 0
	for i, t := range col.Tasks {
		n++ // title
		if t.Body != "" {
			n++ // body preview
		}
		if i < len(col.Tasks)-1 {
			n++ // spacer
		}
	}
	return n
}

func (m Model) renderColumn(colIdx int, targetInnerRows int) string {
	col := m.Board.Columns[colIdx]
	w := m.columnWidth(colIdx)
	innerW := w - 4

	active := colIdx == m.Cursor.Col
	bColor := columnBorderColor(colIdx)

	header := fmt.Sprintf("[%d] %s (%d)", colIdx+1, col.Status.String(), len(col.Tasks))

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
				padRight("  "+title, innerW), bColor, SelectedItemStyle[colIdx], active))
		} else {
			rows = append(rows, renderContentRow(
				padRight("  "+title, innerW), bColor, NormalTask, active))
		}

		if task.Body != "" {
			bodyLine := truncate(strings.SplitN(task.Body, "\n", 2)[0], innerW-2)
			if selected {
				rows = append(rows, renderContentRow(
					padRight("  "+bodyLine, innerW), bColor, SelectedBodyStyle[colIdx], active))
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
				padRight("  (no tasks)", innerW), bColor, SelectedItemStyle[colIdx], active))
		} else {
			rows = append(rows, renderContentRow(
				padRight("  (no tasks)", innerW), bColor, EmptyColumn, active))
		}
	}

	padding := targetInnerRows - len(rows)

	bottom := renderBottomBorder(w, bColor, active)

	allRows := make([]string, 0, len(rows)+max(padding, 0)+2)
	allRows = append(allRows, top)
	allRows = append(allRows, rows...)
	for i := 0; i < padding; i++ {
		allRows = append(allRows, renderEmptyRow(w, bColor, active))
	}
	allRows = append(allRows, bottom)

	return lipgloss.JoinVertical(lipgloss.Left, allRows...)
}

func (m Model) renderDetailOutside() string {
	w := m.Width
	if w <= 0 {
		w = 80
	}
	bColor := columnBorderColor(m.Cursor.Col)
	compact := m.Layout.DetailMode == DetailCompact

	col := m.Board.Columns[m.Cursor.Col]
	if len(col.Tasks) == 0 || m.Cursor.Row >= len(col.Tasks) {
		return m.renderDetailPanel("Task Detail", "", "", "", w, bColor, compact)
	}

	task := col.Tasks[m.Cursor.Row]
	return m.renderDetailPanel(task.Title, task.Body,
		task.Status.String(), relativeTime(task.UpdatedAt),
		w, bColor, compact)
}

func (m Model) renderDetailPanel(title, body, status, updated string, w int, color lipgloss.Color, compact bool) string {
	innerW := w - 4

	var detailScroll string
	if m.DetailVPReady && m.DetailVP.TotalLineCount() > m.DetailVP.Height {
		pct := m.DetailVP.ScrollPercent()
		if pct < 0 {
			pct = 0
		} else if pct > 1 {
			pct = 1
		}
		detailScroll = fmt.Sprintf("%d%%", int(pct*100))
	}

	displayTitle := truncate(title, max(w-8-lipgloss.Width(detailScroll), 10))
	top := renderTopBorder(w, displayTitle, detailScroll, color, true)
	bottom := renderBottomBorder(w, color, true)

	if compact {
		meta := fmt.Sprintf("  Status: %s    Updated: %s", status, updated)
		content := renderContentRow(padRight(meta, innerW), color, BodyPreview, true)
		return lipgloss.JoinVertical(lipgloss.Left, top, content, bottom)
	}

	var vpContent string
	if body != "" {
		lines := strings.Split(body, "\n")
		var styledLines []string
		for _, line := range lines {
			styledLines = append(styledLines, padRight("  "+truncate(line, innerW-2), innerW))
		}
		vpContent = strings.Join(styledLines, "\n")
		if status != "" {
			vpContent += "\n" + strings.Repeat(" ", innerW) // spacer
			vpContent += "\n" + padRight(
				fmt.Sprintf("  Status: %s    Updated: %s", status, updated), innerW)
		}
	} else if title != "" {
		vpContent = padRight("  (no description)", innerW)
	} else {
		vpContent = padRight("  (no task selected)", innerW)
	}

	var rows []string
	if m.DetailVPReady {
		m.DetailVP.SetContent(vpContent)
		vpLines := strings.Split(m.DetailVP.View(), "\n")
		for _, line := range vpLines {
			rows = append(rows, renderContentRow(line, color, NormalTask, true))
		}
	} else {
		for i, line := range strings.Split(vpContent, "\n") {
			if i >= 4 {
				rows = append(rows, renderContentRow(
					padRight("  ...", innerW), color, BodyPreview, true))
				break
			}
			rows = append(rows, renderContentRow(line, color, NormalTask, true))
		}
	}

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

func (m Model) helpView() string {
	switch m.Mode {
	case ModeNormal:
		return HelpBar.Render(
			"j/k navigate  tab/shift+tab column  h/l column  a add  e edit  d delete  enter→  backspace← 1/2/3 jump  q quit",
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

func (m Model) cursorLine() int {
	col := m.Board.Columns[m.Cursor.Col]
	line := 0
	for i := 0; i < m.Cursor.Row && i < len(col.Tasks); i++ {
		line++
		if col.Tasks[i].Body != "" {
			line++
		}
		line++
	}
	return line
}

func (m Model) scrollOffset() int {
	target := m.cursorLine()
	// so use the current offset as base what this does is that, then this scrolls when cursor is close to leaving the
	// visible area that is margin off = 2 so, this helps avoid jump breaks of viewports which is a hassle to handle
	// ngl every cursor move recalculates offset = target - half
	currentOffset := m.Viewport.YOffset
	scrollOffMargin := 2

	switch {
	case target < currentOffset+scrollOffMargin:
		return max(0, target-scrollOffMargin)
	case target > currentOffset+m.Viewport.Height-scrollOffMargin:
		return target - m.Viewport.Height + scrollOffMargin
	default:
		return currentOffset
	}
}

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

// comments are left intentionally.
// code explains behavior; comments explain intent.
// part of the alive internet theory.
//⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
//⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣄⣀⣀⠀⠀⢹⣿⣿⣿⣿⣿⣿⣿⣿
//⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡟⠉⠛⠛⠻⣿⣿⣿⡿⠋⠀⣴⣿⣿⠟⠻⠿⠿⣿⣿⣿
//⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⠂⠀⣠⣿⣿⡏⠀⠠⠾⣿⣿⣿⣦⣤⠀⠀⣸⣿⣿
//⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠁⠀⠚⠻⣿⣿⣷⣤⣤⣀⣠⣿⣿⠟⠁⠠⠾⣿⣿⣿
//⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣶⣶⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⣤⣤⣤⣾⣿⣿
//⣿⣿⣿⣿⣿⣿⠿⠻⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⠿⣿⣿⣿⣿⣿⣿
//⣿⣿⣿⣿⠟⠁⠀⣠⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠿⠿⠿⠿⣿⣿⣿⣷⣄⠀⠈⠻⣿⣿⣿⣿
//⣿⣿⡿⠃⠀⣠⣾⣿⣿⣿⣿⣿⣿⣿⣿⡃⠀⠀⠀⠀⢈⣿⡟⠀⢠⣿⡿⠋⠉⣿⣿⣏⠀⢻⣧⣀⣀⣀⣀⣀⣼⣿⣿⣿⣿⣷⣄⠀⠙⢿⣿⣿
//⣿⡿⠁⠀⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⠻⠟⠁⠀⠀⠙⠟⠃⠀⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⠀⠈⢿⣿
//⣿⠃⠀⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⣤⣤⣶⣿⣷⣦⣤⣶⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⠀⠘⣿
//⡿⠀⢠⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡆⠀⢸
//⡇⠀⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⢸
//⣧⠀⠘⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⢸
//⣿⡀⠀⢻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠀⢀⣿
//⣿⣷⡀⠀⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠁⢀⣼⣿
//⣿⣿⣷⣄⠀⠙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋⠀⢀⣾⣿⣿
//⣿⣿⣿⣿⣷⣄⠀⠈⠙⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠛⠉⠀⢀⣴⣿⣿⣿⣿
//⣿⣿⣿⣿⣿⣿⣿⣶⣶⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⣴⣾⣿⣿⣿⣿⣿⣿
