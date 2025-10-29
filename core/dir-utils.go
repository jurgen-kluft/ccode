package corepkg

import (
	"fmt"
	"os"
	"path/filepath"
)

func DirExists(path string) bool {
	// Check if the directory exists
	if info, err := os.Stat(path); err == nil {
		return info.IsDir()
	}
	return false
}

func DirMake(path string) error {
	// Create the directory if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	return nil
}

func DirList(path string) ([]string, error) {
	// Open the directory
	dir, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open directory %s: %w", path, err)
	}
	defer dir.Close()

	// Read the directory entries
	entries, err := dir.Readdirnames(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	return entries, nil
}

func FindDirMatching(path string, findMatch func(dir string) bool) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if findMatch(entry.Name()) {
			return filepath.Join(path, entry.Name()), nil
		}
	}
	return "", nil
}
