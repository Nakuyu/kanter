package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func boardPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home dir: %w", err)
	}
	dir := filepath.Join(home, ".local", "share", "kanter")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create data dir %s: %w", dir, err)
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
		return Board{}, fmt.Errorf("read %s: %w", path, err)
	}

	var board Board
	if err := json.Unmarshal(data, &board); err != nil {
		return Board{}, fmt.Errorf("parse %s: %w", path, err)
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
		return fmt.Errorf("marshal board: %w", err)
	}

	return writeFileAtomic(path, data, 0o644)
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) (err error) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	tmp, err := os.CreateTemp(dir, base+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		if err != nil {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err = tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("fsync temp file: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err = os.Chmod(tmpName, perm); err != nil {
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err = os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename into place: %w", err)
	}
	return nil
}

func isBoardMissing(err error) bool {
	return errors.Is(err, fs.ErrNotExist)
}
