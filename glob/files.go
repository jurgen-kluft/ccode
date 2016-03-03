package glob

import (
	"os"
	"path/filepath"
)

func GlobMatching(filepath string, g string) bool {
	if match, err := PathMatch(g, filepath); match == true && err == nil {
		return match
	}
	return false
}

func GlobFiles(dirpath string, glob string) (filepaths []string, err error) {
	err = filepath.Walk(dirpath, func(path string, fi os.FileInfo, err error) error {
		if err == nil && fi.IsDir() == false {
			path = path[len(dirpath):]
			match := GlobMatching(path, glob)
			if match {
				filepaths = append(filepaths, path)
			}
		}
		return nil
	})
	return
}
