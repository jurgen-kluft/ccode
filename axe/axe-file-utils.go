package axe

import (
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

func DirExists(path string) bool {
	di, err := os.Stat(path)
	return err == nil && di.IsDir()
}

func FindFileMatching(path string, findMatch func(file string) bool) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if findMatch(entry.Name()) {
			return filepath.Join(path, entry.Name()), nil
		}
	}
	return "", nil
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

func SliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}
