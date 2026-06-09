package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func handleHelp() {
	fmt.Println(`usage: kanter [--board <name>] [<command>]

commands:
  kanter                    launch TUI
  kanter board              show current board and task counts
  kanter list               list all tasks grouped by column
  kanter list-boards        list all boards
  kanter add <title> [body] add a task to TODO
  kanter remove <query>     remove a task by ID or title
  kanter change-status <query> <status>
                            move task to todo/doing/done

flags:
  --board <name>            work with a specific board`)
}

func handleBoard(name string) {
	board, err := loadBoard(name)
	if err != nil {
		boardMissingExit(name, err)
	}

	var counts [3]int
	for i := range board.Columns {
		counts[i] = len(board.Columns[i].Tasks)
	}
	fmt.Printf("Board: %s  (%d TODO \u00b7 %d DOING \u00b7 %d DONE)\n",
		name, counts[0], counts[1], counts[2])
}

func handleList(name string) {
	board, err := loadBoard(name)
	if err != nil {
		boardMissingExit(name, err)
	}

	for i := range board.Columns {
		col := &board.Columns[i]
		status := col.Status.String()
		fmt.Printf("\n%s\n", status)
		for _, task := range col.Tasks {
			id := task.ID[:min(len(task.ID), 8)]
			if task.Body != "" {
				fmt.Printf("  %s  %s  (%s)\n", id, task.Title, task.Body)
			} else {
				fmt.Printf("  %s  %s\n", id, task.Title)
			}
		}
	}
	fmt.Println()
}

func handleListBoards(currentBoardName string) {
	names, err := listBoards()
	if err != nil {
		fmt.Fprintf(os.Stderr, "kanter: %v\n", err)
		os.Exit(1)
	}

	for _, name := range names {
		if name == currentBoardName {
			fmt.Printf("  %s (current)\n", name)
		} else {
			fmt.Printf("  %s\n", name)
		}
	}
}

func handleAdd(name, title, body string) {
	board, err := loadBoard(name)
	if err != nil {
		boardMissingExit(name, err)
	}

	task := Task{
		ID:        newID(),
		Title:     title,
		Body:      body,
		Status:    StatusTodo,
		UpdatedAt: time.Now(),
	}
	board.Columns[columnIndex(task.Status)].Tasks = append(
		board.Columns[columnIndex(task.Status)].Tasks, task,
	)

	if err := saveBoard(board, name); err != nil {
		fmt.Fprintf(os.Stderr, "kanter: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("added: %s  %s\n", task.ID, task.Title)
}

func handleRemove(name, query string) {
	board, err := loadBoard(name)
	if err != nil {
		boardMissingExit(name, err)
	}

	refs := findTask(&board, query)
	if len(refs) == 0 {
		fmt.Fprintf(os.Stderr, "kanter: no task matching %q\n", query)
		os.Exit(1)
	}
	if len(refs) > 1 {
		fmt.Fprintf(os.Stderr, "kanter: %q matches %d tasks — be more specific\n", query, len(refs))
		os.Exit(1)
	}

	r := refs[0]
	title := board.Columns[r.Col].Tasks[r.Row].Title
	board.Columns[r.Col].Tasks = append(
		board.Columns[r.Col].Tasks[:r.Row],
		board.Columns[r.Col].Tasks[r.Row+1:]...,
	)

	if err := saveBoard(board, name); err != nil {
		fmt.Fprintf(os.Stderr, "kanter: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("removed: %s\n", title)
}

func handleChangeStatus(name, query, statusStr string) {
	board, err := loadBoard(name)
	if err != nil {
		boardMissingExit(name, err)
	}

	target, ok := parseStatus(statusStr)
	if !ok {
		fmt.Fprintf(os.Stderr, "kanter: status must be todo, doing, or done (got %q)\n", statusStr)
		os.Exit(1)
	}

	refs := findTask(&board, query)
	if len(refs) == 0 {
		fmt.Fprintf(os.Stderr, "kanter: no task matching %q\n", query)
		os.Exit(1)
	}
	if len(refs) > 1 {
		fmt.Fprintf(os.Stderr, "kanter: %q matches %d tasks — be more specific\n", query, len(refs))
		os.Exit(1)
	}

	r := refs[0]
	task := board.Columns[r.Col].Tasks[r.Row]
	if r.Col == columnIndex(target) {
		fmt.Fprintf(os.Stderr, "kanter: %q is already in %s\n", task.Title, target)
		os.Exit(1)
	}

	board.Columns[r.Col].Tasks = append(
		board.Columns[r.Col].Tasks[:r.Row],
		board.Columns[r.Col].Tasks[r.Row+1:]...,
	)
	task.Status = target
	task.UpdatedAt = time.Now()
	board.Columns[columnIndex(target)].Tasks = append(
		board.Columns[columnIndex(target)].Tasks, task,
	)

	if err := saveBoard(board, name); err != nil {
		fmt.Fprintf(os.Stderr, "kanter: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("moved: %s  %s \u2192 %s\n", task.Title, r.Status.String(), target)
}

type taskRef struct {
	Col    int
	Row    int
	Status Status
}

func findTask(board *Board, query string) []taskRef {
	for i := range board.Columns {
		for j := range board.Columns[i].Tasks {
			if board.Columns[i].Tasks[j].ID == query {
				return []taskRef{{
					Col:    i,
					Row:    j,
					Status: board.Columns[i].Tasks[j].Status,
				}}
			}
		}
	}

	lower := strings.ToLower(query)
	var matches []taskRef
	for i := range board.Columns {
		for j := range board.Columns[i].Tasks {
			if strings.Contains(strings.ToLower(board.Columns[i].Tasks[j].Title), lower) {
				matches = append(matches, taskRef{
					Col:    i,
					Row:    j,
					Status: board.Columns[i].Tasks[j].Status,
				})
			}
		}
	}
	return matches
}

func parseStatus(s string) (Status, bool) {
	switch strings.ToLower(s) {
	case "todo", "0":
		return StatusTodo, true
	case "doing", "1":
		return StatusDoing, true
	case "done", "2":
		return StatusDone, true
	default:
		return 0, false
	}
}

func boardMissingExit(name string, err error) {
	if isBoardMissing(err) {
		fmt.Fprintf(os.Stderr, "kanter: board %q does not exist yet — launch the TUI first\n", name)
	} else {
		fmt.Fprintf(os.Stderr, "kanter: %v\n", err)
	}
	os.Exit(1)
}
