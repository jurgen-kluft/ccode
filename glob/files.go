package glob

import (
	"os"
	"path/filepath"
	"strings"
)

func GlobMatching(filepath string, globs []string) (bool, int) {
	for i, g := range globs {
		g = strings.Replace(g, "^", "", 1)
		if match, err := PathMatch(g, filepath); match == true && err == nil {
			return match, i
		} else if err != nil {
			return false, -1
		}
	}
	return false, 0
}

func GlobFiles(dirpath string, globs []string) (filepaths []string, err error) {
	err = filepath.Walk(dirpath, func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() == false {
			path = path[len(dirpath)+1:]
			match, index := GlobMatching(path, globs)
			if index >= 0 && match {
				//fmt.Println(path)
				filepaths = append(filepaths, path)
			}
		}
		return nil
	})
	return
}
