package axe

import (
	"os"
	"path/filepath"
)

func WriteTextFile(filename string, text string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(text)
	if err != nil {
		return err
	}
	return nil
}

func WriteLinesToFile(filename string, lines []string) error {
	// Make sure the full directory exists, if not create it
	path := filepath.Dir(filename)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, line := range lines {
		_, err = f.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func SliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}
