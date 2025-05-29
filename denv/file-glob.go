package denv

import (
	"os"
	"path/filepath"

	utils "github.com/jurgen-kluft/ccode/utils"
)

func GlobFiles(dirpath string, glob string) (filepaths []string, err error) {
	dirpath = filepath.Clean(dirpath)
	err = filepath.Walk(dirpath, func(path string, fi os.FileInfo, err error) error {
		if err == nil && fi.IsDir() == false {
			path = path[len(dirpath)+1:]
			match := utils.GlobMatching(path, glob)
			if match {
				filepaths = append(filepaths, path)
			}
		}
		return err
	})
	return
}
