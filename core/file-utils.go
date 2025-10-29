package corepkg

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func FileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
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

func FileChangeExtension(filename, newExt string) string {
	// Find the last dot in the filename
	lastDot := -1
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			lastDot = i
			break
		}
	}

	// If no dot is found, just append the new extension
	if lastDot == -1 {
		return filename + newExt
	}

	// Replace the old extension with the new one
	return filename[:lastDot] + newExt
}

func FileOpenWriteClose(path string, data []byte) error {
	// Create or truncate the file for writing
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer file.Close()

	return FileWrite(file, data)
}

func FileWrite(f *os.File, data []byte) error {
	// Write the data to the file
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	// Sync the file to ensure all data is written
	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}
	return nil
}

func FileOpenReadClose(path string) ([]byte, error) {
	// Open the file for reading
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	// Read the entire file into a byte slice
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return data, nil
}

func FileRead(f *os.File, size int) (data []byte, err error) {
	// Create a byte slice to hold the data
	if size > 0 {
		data = make([]byte, size)
		// Read the specified number of bytes from the file
		if n, err := f.Read(data); err != nil {
			if err == io.EOF {
				return data[:n], nil // Return the bytes read if EOF
			}
			return nil, fmt.Errorf("failed to read from file: %w", err)
		} else {
			if n < size {
				return data[:n], nil // Return the bytes read if less than requested
			}
		}
	} else {
		// If size is 0, read the entire file
		data, err = io.ReadAll(f)
		if err != nil {
			return nil, fmt.Errorf("failed to read from file: %w", err)
		}
	}
	return data, nil // Return the entire file content
}

func FileCopy(src, dst string) error {

	// Assume the files are binary files
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}

	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}

	defer dstFile.Close()

	// Copy the source file to the destination file
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file %s to %s: %w", src, dst, err)
	}

	// Sync the destination file to ensure all data is written
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file %s: %w", dst, err)
	}

	return nil
}

func FileEnumerate(rootPath string, dirFunc func(string, string) bool, fileFunc func(string, string)) error {
	rootPath = filepath.Clean(rootPath)
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			relPath := path[len(rootPath):]
			relPath = strings.TrimLeft(relPath, "/")
			if dirFunc(rootPath, relPath) {
				return nil // Continue walking the tree
			}
			return filepath.SkipDir
		}
		relPath := path[len(rootPath)+1:]
		fileFunc(rootPath, relPath)
		return nil
	})

	return err
}
