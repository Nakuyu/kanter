package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func boardPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".local", "share", "kanter")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "kanter.json"), nil
}

func loadBoard() (Board, error) {
	path, err := boardPath()
	if err != nil {
		return Board{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Board{}, err
	}

	var board Board
	if err := json.Unmarshal(data, &board); err != nil {
		return Board{}, err
	}
	return board, nil
}

func saveBoard(board Board) error {
	path, err := boardPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(board, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
