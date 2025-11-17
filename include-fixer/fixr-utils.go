package fixr

import (
	"path/filepath"
	"strings"
)

// ChangeExtension changes the extension of a file
// Example: ChangeExtension("foo/bar.hpp", ".hpp", ".h") -> "foo/bar.h"
func ChangeExtension(filepath string, from string, to string) string {
	// Change the extension of the file
	if strings.HasSuffix(filepath, from) {
		return filepath[:len(filepath)-len(from)] + to
	}
	return filepath
}

// ParentPath returns the parent directory
func ParentPath(path string) string {
	if path == "." {
		return "."
	}
	path = filepath.Clean(path)
	return filepath.Dir(path)
}
