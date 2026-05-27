package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleNormalMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.Focused == FocusDetail {
		return m.handleDetailFocusMode(msg)
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		m = m.moveCursorUp()
	case key.Matches(msg, m.keys.Down):
		m = m.moveCursorDown()
	case key.Matches(msg, m.keys.Left):
		m = m.moveCursorLeft()
	case key.Matches(msg, m.keys.Right):
		m = m.moveCursorRight()

	case msg.String() == "1":
		m.Cursor.Col = 0
		m = m.clampRow()
	case msg.String() == "2":
		m.Cursor.Col = 1
		m = m.clampRow()
	case msg.String() == "3":
		m.Cursor.Col = 2
		m = m.clampRow()

	case key.Matches(msg, m.keys.Tab):
		m = m.focusNext()
	case key.Matches(msg, m.keys.ShiftTab):
		m = m.focusPrev()

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

	case key.Matches(msg, m.keys.ShiftEnter):
		m = m.moveTask(-1)
		return m, saveBoardCmd(m)

	case key.Matches(msg, m.keys.Enter):
		m = m.moveTask(+1)
		return m, saveBoardCmd(m)
	}

	return m, nil
}

// for tab and shift tab wrapping around.
func (m Model) focusNext() Model {
	if m.Focused == FocusDetail {
		m.Focused = FocusList
		m.Cursor.Col = 0
		return m.clampRow()
	}

	if m.Cursor.Col < len(m.Board.Columns)-1 {
		m.Cursor.Col++
		return m.clampRow()
	}

	col := m.Board.Columns[m.Cursor.Col]
	if m.Layout.DetailMode != DetailHidden &&
		m.Cursor.Row < len(col.Tasks) &&
		col.Tasks[m.Cursor.Row].Body != "" {
		m.Focused = FocusDetail
		if m.DetailVPReady {
			m.DetailVP.GotoTop()
		}
		return m
	}
	m.Cursor.Col = 0
	return m.clampRow()
}

func (m Model) focusPrev() Model {
	if m.Focused == FocusDetail {
		m.Focused = FocusList
		m.Cursor.Col = len(m.Board.Columns) - 1
		return m.clampRow()
	}

	if m.Cursor.Col > 0 {
		m.Cursor.Col--
		return m.clampRow()
	}

	col := m.Board.Columns[m.Cursor.Col]
	if m.Layout.DetailMode != DetailHidden &&
		m.Cursor.Row < len(col.Tasks) &&
		col.Tasks[m.Cursor.Row].Body != "" {
		m.Focused = FocusDetail
		if m.DetailVPReady {
			m.DetailVP.GotoTop()
		}
		return m
	}

	m.Cursor.Col = len(m.Board.Columns) - 1
	return m.clampRow()
}

func (m Model) handleDetailFocusMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Deny): // esc
		m.Focused = FocusList
		return m, nil

	case key.Matches(msg, m.keys.Tab):
		m = m.focusNext()
		return m, nil

	case key.Matches(msg, m.keys.ShiftTab):
		m = m.focusPrev()
		return m, nil

	default:
		var cmd tea.Cmd
		m.DetailVP, cmd = m.DetailVP.Update(msg)
		return m, cmd
	}
}
func (m Model) enterAddMode() Model {
	m.Mode = ModeAdding
	m.FormStage = formStageTitle
	m.TitleInput.SetValue("")
	m.BodyTextarea.SetValue("")
	m.TitleInput.Focus()
	return m
}

func (m Model) enterEditMode() Model {
	m.Mode = ModeEditing
	m.FormStage = formStageTitle
	task := m.Board.Columns[m.Cursor.Col].Tasks[m.Cursor.Row]
	m.TitleInput.SetValue(task.Title)
	m.BodyTextarea.SetValue(task.Body)
	m.TitleInput.Focus()
	return m
}

func (m Model) cancelForm() Model {
	m.Mode = ModeNormal
	m.FormStage = formStageTitle
	m.TitleInput.Blur()
	m.BodyTextarea.Blur()
	return m
}

func (m Model) handleFormMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}

	switch {
	case key.Matches(msg, m.keys.Deny):
		return m.cancelForm(), nil

	case m.FormStage == formStageTitle &&
		(key.Matches(msg, m.keys.Confirm) || key.Matches(msg, m.keys.Tab)):
		return m.advanceOrSubmitForm()

	case m.FormStage == formStageBody && msg.Alt && msg.Type == tea.KeyEnter:
		return m.advanceOrSubmitForm()

	case m.FormStage == formStageBody && key.Matches(msg, m.keys.ShiftTab):
		m.FormStage = formStageTitle
		m.BodyTextarea.Blur()
		return m, m.TitleInput.Focus()
	}

	return m, nil
}

func (m Model) advanceOrSubmitForm() (Model, tea.Cmd) {
	if m.FormStage == formStageTitle {
		if m.TitleInput.Value() == "" {
			return m, nil
		}
		m.FormStage = formStageBody
		m.TitleInput.Blur()
		return m, m.BodyTextarea.Focus()
	}

	title := strings.TrimSpace(m.TitleInput.Value())
	if title == "" {
		return m, nil
	}
	body := m.BodyTextarea.Value()

	switch m.Mode {
	case ModeAdding:
		return m.submitNewTask(title, body)
	case ModeEditing:
		return m.submitEditedTask(title, body)
	}
	return m, nil
}

func (m Model) submitNewTask(title, body string) (Model, tea.Cmd) {
	task := Task{
		ID:        newID(),
		Title:     title,
		Body:      body,
		Status:    StatusTodo,
		UpdatedAt: time.Now(),
	}
	idx := columnIndex(task.Status)
	m.Board.Columns[idx].Tasks = append(m.Board.Columns[idx].Tasks, task)

	m.StatusMsg = "Task created"
	m.StatusMsgType = MsgSuccess
	m = m.cancelForm()
	return m, saveBoardCmd(m)
}

func (m Model) submitEditedTask(title, body string) (Model, tea.Cmd) {
	t := &m.Board.Columns[m.Cursor.Col].Tasks[m.Cursor.Row]
	t.Title = title
	t.Body = body
	t.UpdatedAt = time.Now()

	m.StatusMsg = "Task saved"
	m.StatusMsgType = MsgSuccess
	m = m.cancelForm()
	return m, saveBoardCmd(m)
}

func (m Model) handleConfirmDeleteMode(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Confirm), msg.String() == "y":
		m.StatusMsg = "Task deleted"
		m.StatusMsgType = MsgSuccess
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
	return m.clampRow()
}

func (m Model) moveCursorUp() Model {
	if m.Cursor.Row > 0 {
		m.Cursor.Row--
	}
	return m
}

func (m Model) moveCursorDown() Model {
	maxRow := len(m.Board.Columns[m.Cursor.Col].Tasks) - 1
	if m.Cursor.Row < maxRow {
		m.Cursor.Row++
	}
	return m
}

func (m Model) moveCursorLeft() Model {
	if m.Cursor.Col > 0 {
		m.Cursor.Col--
	}
	return m.clampRow()
}

func (m Model) moveCursorRight() Model {
	if m.Cursor.Col < len(m.Board.Columns)-1 {
		m.Cursor.Col++
	}
	return m.clampRow()
}

func (m Model) clampRow() Model {
	maxRow := len(m.Board.Columns[m.Cursor.Col].Tasks) - 1
	if maxRow < 0 {
		m.Cursor.Row = 0
		return m
	}
	if m.Cursor.Row > maxRow {
		m.Cursor.Row = maxRow
	}
	return m
}

func (m Model) moveTask(delta int) Model {
	targetIdx := m.Cursor.Col + delta
	targetStatus, ok := statusAt(targetIdx)
	if !ok {
		return m
	}

	col := &m.Board.Columns[m.Cursor.Col]
	if len(col.Tasks) == 0 || m.Cursor.Row >= len(col.Tasks) {
		return m
	}

	task := col.Tasks[m.Cursor.Row]
	task.Status = targetStatus
	task.UpdatedAt = time.Now()

	arrow := "→"
	if delta < 0 {
		arrow = "←"
	}
	m.StatusMsg = fmt.Sprintf("%s Moved to %s", arrow, targetStatus)
	m.StatusMsgType = MsgInfo

	col.Tasks = append(col.Tasks[:m.Cursor.Row], col.Tasks[m.Cursor.Row+1:]...)
	m.Board.Columns[targetIdx].Tasks = append(
		m.Board.Columns[targetIdx].Tasks, task,
	)

	return m.clampRow()
}
