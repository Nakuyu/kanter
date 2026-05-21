package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
)

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
	Status Status
	Tasks  []Task
}

type Board struct {
	Columns [3]Column
}

type Cursor struct {
	Col int
	Row int
}

type Mode int

const (
	ModeNormal Mode = iota
	ModeAdding
	ModeConfirmDelete
)

type Model struct {
	Board  Board
	Cursor Cursor
	Mode   Mode

	Width  int
	Height int
	Err    error

	keys keyMap
	help help.Model
}

func NewModel() Model {
	return Model{
		Board: Board{
			Columns: [3]Column{
				{Status: StatusTodo},
				{Status: StatusDoing},
				{Status: StatusDone},
			},
		},
		Cursor: Cursor{Col: 0, Row: 0},
		Mode:   ModeNormal,
		keys:   defaultKeyMap(),
		help:   help.New(),
	}
}

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Enter  key.Binding
	Add    key.Binding
	Delete key.Binding
	Quit   key.Binding
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
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch m.Mode {
	case ModeNormal:
		return m.handleNormalMode(msg)
	case ModeAdding:
		return m.handleAddingMode(msg)
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

	case key.Matches(msg, m.keys.Add):
		m.Mode = ModeAdding

	case key.Matches(msg, m.keys.Delete):
		if len(m.Board.Columns[m.Cursor.Col].Tasks) > 0 {
			m.Mode = ModeConfirmDelete
		}

	case key.Matches(msg, m.keys.Enter):
		m.moveCurrentTask()
	}

	return m, nil
}

func (m Model) handleAddingMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	// TODO: implement task adding with textinput
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleConfirmDeleteMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	// TODO: implement confirmation dialog
	return m, nil
}

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

func (m *Model) moveCurrentTask() {
	col := &m.Board.Columns[m.Cursor.Col]
	if len(col.Tasks) == 0 || m.Cursor.Row >= len(col.Tasks) {
		return
	}
	if col.Status >= StatusDone {
		return
	}

	task := col.Tasks[m.Cursor.Row]
	task.Status++

	col.Tasks = append(col.Tasks[:m.Cursor.Row], col.Tasks[m.Cursor.Row+1:]...)
	m.Board.Columns[task.Status].Tasks = append(
		m.Board.Columns[task.Status].Tasks, task,
	)

	m.clampRow()
}

func (m Model) View() string {
	// TODO: render three columns side-by-side with Lip Gloss
	return "kanter — press q to quit\n"
}
