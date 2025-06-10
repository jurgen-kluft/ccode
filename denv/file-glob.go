package denv

import (
	"os"
	"path/filepath"

	"github.com/jurgen-kluft/ccode/foundation"
)

func GlobFiles(dirpath string, glob string, isExcluded func(string) bool) (filepaths []string, err error) {
	dirpath = filepath.Clean(dirpath)
	err = filepath.Walk(dirpath, func(path string, fi os.FileInfo, err error) error {
		if err == nil && fi.IsDir() == false {
			path = path[len(dirpath)+1:]
			match := foundation.GlobMatching(path, glob)
			if match && !isExcluded(path) {
				filepaths = append(filepaths, path)
			}
		}
		return err
	})
	return
}
