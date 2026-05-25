package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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

	TitleInput   textinput.Model
	BodyTextarea textarea.Model
	AddStage     int

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

	ta := textarea.New()
	ta.Placeholder = "Description (optional)"
	ta.CharLimit = 2000
	ta.SetWidth(50)
	ta.SetHeight(4)
	ta.ShowLineNumbers = false

	savePath, _ := boardPath()

	return Model{
		Board: Board{
			Columns: [3]Column{
				{Status: StatusTodo, Tasks: []Task{}},
				{Status: StatusDoing, Tasks: []Task{}},
				{Status: StatusDone, Tasks: []Task{}},
			},
		},
		Cursor:       Cursor{Col: 0, Row: 0},
		Mode:         ModeNormal,
		TitleInput:   ti,
		BodyTextarea: ta,
		AddStage:     0,
		SavePath:     savePath,
		keys:         defaultKeyMap(),
		help:         help.New(),
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
