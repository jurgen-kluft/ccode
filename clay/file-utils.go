package clay

import (
	"fmt"
	"io"
	"os"
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

func MakeDir(path string) error {
	// Create the directory if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	return nil
}

func CopyFiles(src, dst string) error {

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
