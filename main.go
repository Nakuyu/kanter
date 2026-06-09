package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	args := os.Args[1:]
	boardName := ""

	for len(args) > 0 && strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "--board":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "kanter: --board requires a name")
				os.Exit(1)
			}
			boardName = args[1]
			args = args[2:]
		case "--help", "-h":
			handleHelp()
			return
		default:
			fmt.Fprintf(os.Stderr, "kanter: unknown flag %s\n", args[0])
			os.Exit(1)
		}
	}

	if boardName == "" {
		var err error
		boardName, err = currentBoard()
		if err != nil {
			fmt.Fprintf(os.Stderr, "kanter: %v\n", err)
			os.Exit(1)
		}
	}

	if len(args) > 0 {
		switch args[0] {
		case "help":
			handleHelp()
		case "board":
			handleBoard(boardName)
		case "list":
			handleList(boardName)
		case "list-boards":
			handleListBoards(boardName)
		case "add":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "usage: kanter add <title> [body]")
				os.Exit(1)
			}
			title := args[1]
			body := ""
			if len(args) > 2 {
				body = strings.Join(args[2:], " ")
			}
			handleAdd(boardName, title, body)
		case "remove":
			if len(args) < 2 {
				fmt.Fprintln(os.Stderr, "usage: kanter remove <id-or-title>")
				os.Exit(1)
			}
			handleRemove(boardName, strings.Join(args[1:], " "))
		case "change-status":
			if len(args) < 3 {
				fmt.Fprintln(os.Stderr, "usage: kanter change-status <id-or-title> <todo|doing|done>")
				os.Exit(1)
			}
			handleChangeStatus(boardName, args[1], args[2])
		default:
			fmt.Fprintf(os.Stderr, "kanter: unknown command %q\n", args[0])
			os.Exit(1)
		}
		return
	}

	p := tea.NewProgram(
		NewModel(boardName),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "kanter: %v\n", err)
		os.Exit(1)
	}
}
