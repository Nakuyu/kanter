package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Types ────────────────────────────────────────────────────────────────

type Status int

const (
	StatusTodo Status = iota
	StatusDoing
	StatusDone
)

func (s Status) String() string {
	switch s {
	case StatusTodo:
		return "TODO"
	case StatusDoing:
		return "DOING"
	case StatusDone:
		return "DONE"
	default:
		return "???"
	}
}

type Task struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body,omitempty"`
	Status Status `json:"status"`
}

type Column struct {
	Status Status  `json:"status"`
	Tasks  []Task  `json:"tasks"`
}

type Board struct {
	Columns [3]Column `json:"columns"`
}

type Cursor struct {
	Col int
	Row int
}

type Mode int

const (
	ModeNormal Mode = iota
	ModeAdding
	ModeEditing
	ModeConfirmDelete
)

type Model struct {
	Board  Board
	Cursor Cursor
	Mode   Mode

	TitleInput textinput.Model
	BodyInput  textinput.Model
	AddStage   int

	Width  int
	Height int
	Err    error

	SavePath string

	keys keyMap
	help help.Model
}

// ─── Key bindings ─────────────────────────────────────────────────────────

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	Enter   key.Binding
	Add     key.Binding
	Edit    key.Binding
	Delete  key.Binding
	Confirm key.Binding
	Deny    key.Binding
	Quit    key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("j/k", "navigate"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("", ""),
		),
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/l", "column"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("", ""),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "move task"),
		),
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add task"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit task"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Deny: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Left, k.Add, k.Edit, k.Enter, k.Delete, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Add, k.Edit, k.Enter, k.Delete},
		{k.Confirm, k.Deny, k.Quit},
	}
}

// ─── Model construction ───────────────────────────────────────────────────

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Task title"
	ti.CharLimit = 100
	ti.Width = 50

	bi := textinput.New()
	bi.Placeholder = "Description (optional)"
	bi.CharLimit = 500
	bi.Width = 50

	savePath, _ := boardPath()

	return Model{
		Board: Board{
			Columns: [3]Column{
				{Status: StatusTodo, Tasks: []Task{}},
				{Status: StatusDoing, Tasks: []Task{}},
				{Status: StatusDone, Tasks: []Task{}},
			},
		},
		Cursor:    Cursor{Col: 0, Row: 0},
		Mode:      ModeNormal,
		TitleInput: ti,
		BodyInput:  bi,
		AddStage:  0,
		SavePath:  savePath,
		keys:      defaultKeyMap(),
		help:      help.New(),
	}
}

// ─── Messages ─────────────────────────────────────────────────────────────

type loadResultMsg struct {
	board Board
}

type loadFirstRunMsg struct{}

type errMsg struct {
	err error
}

// ─── Commands ─────────────────────────────────────────────────────────────

func loadBoardCmd() tea.Cmd {
	return func() tea.Msg {
		board, err := loadBoard()
		if err != nil {
			return loadFirstRunMsg{}
		}
		// Ensure column slices are non-nil
		for i := range board.Columns {
			if board.Columns[i].Tasks == nil {
				board.Columns[i].Tasks = []Task{}
			}
		}
		return loadResultMsg{board: board}
	}
}

func saveBoardCmd(m Model) tea.Cmd {
	return func() tea.Msg {
		if err := saveBoard(m.Board); err != nil {
			return errMsg{err: err}
		}
		return nil
	}
}

// ─── Init ─────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	return loadBoardCmd()
}

// ─── Update ───────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		m.Err = nil
		return m.handleKeyMsg(msg)

	case loadResultMsg:
		m.Board = msg.board
		return m, nil

	case loadFirstRunMsg:
		return m, nil

	case errMsg:
		m.Err = msg.err
		return m, nil
	}

	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch m.Mode {
	case ModeNormal:
		return m.handleNormalMode(msg)
	case ModeAdding:
		return m.handleAddingMode(msg)
	case ModeEditing:
		return m.handleEditingMode(msg)
	case ModeConfirmDelete:
		return m.handleConfirmDeleteMode(msg)
	}
	return m, nil
}

func (m Model) handleNormalMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		m.moveCursorUp()
	case key.Matches(msg, m.keys.Down):
		m.moveCursorDown()
	case key.Matches(msg, m.keys.Left):
		m.moveCursorLeft()
	case key.Matches(msg, m.keys.Right):
		m.moveCursorRight()

	case msg.String() == "0":
		m.Cursor.Col = 0
		m.clampRow()
	case msg.String() == "1":
		m.Cursor.Col = 1
		m.clampRow()
	case msg.String() == "2":
		m.Cursor.Col = 2
		m.clampRow()

	case key.Matches(msg, m.keys.Add):
		m = m.enterAddMode()
		return m, nil

	case key.Matches(msg, m.keys.Edit):
		if len(m.Board.Columns[m.Cursor.Col].Tasks) > 0 {
			m = m.enterEditMode()
			return m, nil
		}

	case key.Matches(msg, m.keys.Delete):
		if len(m.Board.Columns[m.Cursor.Col].Tasks) > 0 {
			m.Mode = ModeConfirmDelete
		}

	case key.Matches(msg, m.keys.Enter):
		m.moveCurrentTask()
		return m, saveBoardCmd(m)
	}

	return m, nil
}

func (m Model) handleAddingMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}

	switch {
	case key.Matches(msg, m.keys.Deny):
		m = m.cancelAddMode()
		return m, nil

	case key.Matches(msg, m.keys.Confirm):
		return m.handleAddConfirm()
	}

	var cmd tea.Cmd
	if m.AddStage == 0 {
		m.TitleInput, cmd = m.TitleInput.Update(msg)
	} else {
		m.BodyInput, cmd = m.BodyInput.Update(msg)
	}
	return m, cmd
}

func (m Model) handleConfirmDeleteMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Confirm), msg.String() == "y":
		m = m.deleteCurrentTask()
		m.Mode = ModeNormal
		return m, saveBoardCmd(m)

	case key.Matches(msg, m.keys.Deny), msg.String() == "n":
		m.Mode = ModeNormal
		return m, nil
	}

	return m, nil
}

// ─── Add task helpers ─────────────────────────────────────────────────────

func (m Model) enterAddMode() Model {
	m.Mode = ModeAdding
	m.AddStage = 0
	m.TitleInput.SetValue("")
	m.BodyInput.SetValue("")
	m.TitleInput.Focus()
	return m
}

func (m Model) cancelAddMode() Model {
	m.Mode = ModeNormal
	m.AddStage = 0
	m.TitleInput.Blur()
	m.BodyInput.Blur()
	return m
}

func (m Model) handleAddConfirm() (Model, tea.Cmd) {
	if m.AddStage == 0 {
		title := m.TitleInput.Value()
		if title == "" {
			return m, nil
		}
		m.AddStage = 1
		m.TitleInput.Blur()
		m.BodyInput.Focus()
		return m, nil
	}

	title := m.TitleInput.Value()
	if title == "" {
		return m, nil
	}

	task := Task{
		ID:     fmt.Sprintf("%x", time.Now().UnixNano()),
		Title:  title,
		Body:   m.BodyInput.Value(),
		Status: StatusTodo,
	}
	m.Board.Columns[0].Tasks = append(m.Board.Columns[0].Tasks, task)

	m = m.cancelAddMode()
	return m, saveBoardCmd(m)
}

// ─── Delete task helper ───────────────────────────────────────────────────

func (m Model) deleteCurrentTask() Model {
	col := &m.Board.Columns[m.Cursor.Col]
	if m.Cursor.Row < len(col.Tasks) {
		col.Tasks = append(col.Tasks[:m.Cursor.Row], col.Tasks[m.Cursor.Row+1:]...)
	}
	m.clampRow()
	return m
}

// ─── Edit task helpers ────────────────────────────────────────────────────

func (m Model) enterEditMode() Model {
	m.Mode = ModeEditing
	m.AddStage = 0
	task := m.Board.Columns[m.Cursor.Col].Tasks[m.Cursor.Row]
	m.TitleInput.SetValue(task.Title)
	m.BodyInput.SetValue(task.Body)
	m.TitleInput.Focus()
	return m
}

func (m Model) cancelEditMode() Model {
	m.Mode = ModeNormal
	m.AddStage = 0
	m.TitleInput.Blur()
	m.BodyInput.Blur()
	return m
}

func (m Model) handleEditingMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}

	switch {
	case key.Matches(msg, m.keys.Deny):
		m = m.cancelEditMode()
		return m, nil

	case key.Matches(msg, m.keys.Confirm):
		return m.handleEditConfirm()
	}

	var cmd tea.Cmd
	if m.AddStage == 0 {
		m.TitleInput, cmd = m.TitleInput.Update(msg)
	} else {
		m.BodyInput, cmd = m.BodyInput.Update(msg)
	}
	return m, cmd
}

func (m Model) handleEditConfirm() (Model, tea.Cmd) {
	if m.AddStage == 0 {
		title := m.TitleInput.Value()
		if title == "" {
			return m, nil
		}
		m.AddStage = 1
		m.TitleInput.Blur()
		m.BodyInput.Focus()
		return m, nil
	}

	title := m.TitleInput.Value()
	if title == "" {
		return m, nil
	}

	m.Board.Columns[m.Cursor.Col].Tasks[m.Cursor.Row].Title = title
	m.Board.Columns[m.Cursor.Col].Tasks[m.Cursor.Row].Body = m.BodyInput.Value()

	m = m.cancelEditMode()
	return m, saveBoardCmd(m)
}

// ─── Cursor movement ──────────────────────────────────────────────────────

func (m *Model) moveCursorUp() {
	if m.Cursor.Row > 0 {
		m.Cursor.Row--
	}
}

func (m *Model) moveCursorDown() {
	maxRow := len(m.Board.Columns[m.Cursor.Col].Tasks) - 1
	if m.Cursor.Row < maxRow {
		m.Cursor.Row++
	}
}

func (m *Model) moveCursorLeft() {
	if m.Cursor.Col > 0 {
		m.Cursor.Col--
	}
	m.clampRow()
}

func (m *Model) moveCursorRight() {
	if m.Cursor.Col < len(m.Board.Columns)-1 {
		m.Cursor.Col++
	}
	m.clampRow()
}

func (m *Model) clampRow() {
	maxRow := len(m.Board.Columns[m.Cursor.Col].Tasks) - 1
	if maxRow < 0 {
		m.Cursor.Row = 0
		return
	}
	if m.Cursor.Row > maxRow {
		m.Cursor.Row = maxRow
	}
}

// ─── Move task ────────────────────────────────────────────────────────────

func (m *Model) moveCurrentTask() {
	if m.Cursor.Col >= len(m.Board.Columns)-1 {
		return
	}

	col := &m.Board.Columns[m.Cursor.Col]
	if len(col.Tasks) == 0 || m.Cursor.Row >= len(col.Tasks) {
		return
	}

	task := col.Tasks[m.Cursor.Row]
	newStatus := Status(m.Cursor.Col + 1)
	task.Status = newStatus

	col.Tasks = append(col.Tasks[:m.Cursor.Row], col.Tasks[m.Cursor.Row+1:]...)

	m.Board.Columns[newStatus].Tasks = append(
		m.Board.Columns[newStatus].Tasks, task,
	)

	m.clampRow()
}

// ─── View ─────────────────────────────────────────────────────────────────

func (m Model) View() string {
	cols := make([]string, 3)
	for i := range cols {
		cols[i] = m.renderColumn(i)
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

	var modeOverlay string
	switch m.Mode {
	case ModeAdding, ModeEditing:
		var label, input string
		if m.AddStage == 0 {
			label = PromptStyle.Render("Title:")
			input = m.TitleInput.View()
		} else {
			label = PromptStyle.Render("Body:")
			input = m.BodyInput.View()
		}
		modeOverlay = label + " " + input
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

	rows := []string{header}
	for i, task := range col.Tasks {
		prefix := "  "
		style := NormalTask
		if colIdx == m.Cursor.Col && i == m.Cursor.Row {
			prefix = "> "
			style = SelectedItem
		}
		row := style.Render(prefix + task.Title)
		if task.Body != "" {
			row += "\n" + BodyPreview.Render("  " + truncate(task.Body, 28))
		}
		rows = append(rows, row)
	}

	if len(col.Tasks) == 0 {
		rows = append(rows, EmptyColumn.Render("(empty)"))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return ColumnBorder(colIdx).Width(m.columnWidth(colIdx)).Render(content)
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

// ─── Utilities ────────────────────────────────────────────────────────────

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
