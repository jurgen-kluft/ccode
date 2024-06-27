package fixr

import (
	"bufio"
	"os"
	"path/filepath"
)

// ----------------------------------------------------------------------------
// File I/O, read lines from file, write lines to file
// ----------------------------------------------------------------------------

func readLinesFromFile(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	lines := make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	return lines, scanner.Err()
}

func writeLinesToFile(filename string, lines []string) error {
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
