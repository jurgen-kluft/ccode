package foundation

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

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

func FileExists(path string) bool {
	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func FileRead(path string) ([]byte, error) {
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

func FileEnumerate(rootPath string, dirFunc func(string, string) bool, fileFunc func(string, string)) {
	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
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
}
