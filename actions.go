package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// ─── Normal mode ──────────────────────────────────────────────────────────

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

// ─── Adding mode ──────────────────────────────────────────────────────────

func (m Model) enterAddMode() Model {
	m.Mode = ModeAdding
	m.AddStage = 0
	m.TitleInput.SetValue("")
	m.BodyTextarea.SetValue("")
	m.TitleInput.Focus()
	return m
}

func (m Model) cancelAddMode() Model {
	m.Mode = ModeNormal
	m.AddStage = 0
	m.TitleInput.Blur()
	m.BodyTextarea.Blur()
	return m
}

func (m Model) handleAddingMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}

	if m.AddStage == 0 {
		switch {
		case key.Matches(msg, m.keys.Deny):
			m = m.cancelAddMode()
			return m, nil
		case key.Matches(msg, m.keys.Confirm):
			return m.handleAddConfirm()
		}
		var cmd tea.Cmd
		m.TitleInput, cmd = m.TitleInput.Update(msg)
		return m, cmd
	}

	// Body stage: esc = save, enter = newline (textarea handles it)
	if key.Matches(msg, m.keys.Deny) {
		return m.handleAddConfirm()
	}

	var cmd tea.Cmd
	m.BodyTextarea, cmd = m.BodyTextarea.Update(msg)
	return m, cmd
}

func (m Model) handleAddConfirm() (Model, tea.Cmd) {
	if m.AddStage == 0 {
		title := m.TitleInput.Value()
		if title == "" {
			return m, nil
		}
		m.AddStage = 1
		m.TitleInput.Blur()
		return m, m.BodyTextarea.Focus()
	}

	title := m.TitleInput.Value()
	if title == "" {
		return m, nil
	}

	task := Task{
		ID:     fmt.Sprintf("%x", time.Now().UnixNano()),
		Title:  title,
		Body:   m.BodyTextarea.Value(),
		Status: StatusTodo,
	}
	m.Board.Columns[0].Tasks = append(m.Board.Columns[0].Tasks, task)

	m = m.cancelAddMode()
	return m, saveBoardCmd(m)
}

// ─── Editing mode ─────────────────────────────────────────────────────────

func (m Model) enterEditMode() Model {
	m.Mode = ModeEditing
	m.AddStage = 0
	task := m.Board.Columns[m.Cursor.Col].Tasks[m.Cursor.Row]
	m.TitleInput.SetValue(task.Title)
	m.BodyTextarea.SetValue(task.Body)
	m.TitleInput.Focus()
	return m
}

func (m Model) cancelEditMode() Model {
	m.Mode = ModeNormal
	m.AddStage = 0
	m.TitleInput.Blur()
	m.BodyTextarea.Blur()
	return m
}

func (m Model) handleEditingMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}

	if m.AddStage == 0 {
		switch {
		case key.Matches(msg, m.keys.Deny):
			m = m.cancelEditMode()
			return m, nil
		case key.Matches(msg, m.keys.Confirm):
			return m.handleEditConfirm()
		}
		var cmd tea.Cmd
		m.TitleInput, cmd = m.TitleInput.Update(msg)
		return m, cmd
	}

	// Body stage: esc = save, enter = newline (textarea handles it)
	if key.Matches(msg, m.keys.Deny) {
		return m.handleEditConfirm()
	}

	var cmd tea.Cmd
	m.BodyTextarea, cmd = m.BodyTextarea.Update(msg)
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
		return m, m.BodyTextarea.Focus()
	}

	title := m.TitleInput.Value()
	if title == "" {
		return m, nil
	}

	m.Board.Columns[m.Cursor.Col].Tasks[m.Cursor.Row].Title = title
	m.Board.Columns[m.Cursor.Col].Tasks[m.Cursor.Row].Body = m.BodyTextarea.Value()

	m = m.cancelEditMode()
	return m, saveBoardCmd(m)
}

// ─── Confirm delete mode ──────────────────────────────────────────────────

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

func (m Model) deleteCurrentTask() Model {
	col := &m.Board.Columns[m.Cursor.Col]
	if m.Cursor.Row < len(col.Tasks) {
		col.Tasks = append(col.Tasks[:m.Cursor.Row], col.Tasks[m.Cursor.Row+1:]...)
	}
	m.clampRow()
	return m
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
