package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// dataDir returns ~/.local/share/kanter, creating it if missing.
func dataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home dir: %w", err)
	}
	dir := filepath.Join(home, ".local", "share", "kanter")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create data dir %s: %w", dir, err)
	}
	return dir, nil
}

func boardPath(name string) (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".json"), nil
}

func currentBoardPath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".current"), nil
}

// currentBoard reads the saved board name. Defaults to "kanter".
func currentBoard() (string, error) {
	path, err := currentBoardPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "kanter", nil
		}
		return "", fmt.Errorf("read current board: %w", err)
	}
	name := strings.TrimSpace(string(data))
	if name == "" {
		return "kanter", nil
	}
	return name, nil
}

func setCurrentBoard(name string) error {
	path, err := currentBoardPath()
	if err != nil {
		return err
	}
	return writeFileAtomic(path, []byte(name), 0o644)
}

func loadBoard(name string) (Board, error) {
	path, err := boardPath(name)
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

func saveBoard(board Board, name string) error {
	path, err := boardPath(name)
	if err != nil {
		return err
	}
	return saveBoardPath(board, path)
}

func saveBoardPath(board Board, path string) error {
	data, err := json.MarshalIndent(board, "", "\t")
	if err != nil {
		return fmt.Errorf("marshal board: %w", err)
	}
	return writeFileAtomic(path, data, 0o644)
}

func listBoards() ([]string, error) {
	dir, err := dataDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read data dir: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), ".json"))
	}
	sort.Strings(names)
	return names, nil
}

// isBoardMissing reports whether the error indicates a missing board file.
func isBoardMissing(err error) bool {
	return errors.Is(err, fs.ErrNotExist)
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
