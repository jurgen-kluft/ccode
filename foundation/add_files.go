package foundation

import (
	"os"
	"path/filepath"
	"strings"
)

func AddFilesFrom(rootPath string, dirFunc func(string, string) bool, fileFunc func(string, string)) {
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
