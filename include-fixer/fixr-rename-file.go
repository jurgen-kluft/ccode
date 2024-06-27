package fixr

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type FileRename struct {
	CurrentFilePath string
	NewFilePath     string
}

var CCoreFileNamingPolicy = func(_filepath string) (bool, string) {
	if strings.HasSuffix(_filepath, ".hpp") {
		return true, strings.TrimSuffix(_filepath, ".hpp") + ".h"
	}
	return false, _filepath
}

func (f *Fixr) globAndRename(dirpath string, renamer func(_filepath string) (renameNeeded bool, renamedFilepath string), isHeaderFileFunc func(_filepath string) bool) {

	filesToRename := make([]FileRename, 0, 16)

	err := filepath.WalkDir(dirpath, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			path = filepath.Dir(path)
			relpath, _ := filepath.Rel(dirpath, path)
			_filepath := filepath.Join(relpath, d.Name())
			renameNeeded, renamedFilepath := renamer(_filepath)
			if renameNeeded {
				filesToRename = append(filesToRename, FileRename{CurrentFilePath: _filepath, NewFilePath: renamedFilepath})
			}
		}
		return err
	})

	if err != nil {
		fmt.Println(err)
	}

	// Rename all necessary files and fix all include directives that use any of the renamed header files
	if len(filesToRename) > 0 {
		for _, r := range filesToRename {
			if !f.DryRun {
				err = os.Rename(r.CurrentFilePath, r.NewFilePath)
			} else {
				err = nil
			}
			if err == nil {
				if f.Verbose {
					fmt.Printf(" * renamed file   %s  ->  %s\n", r.CurrentFilePath, r.NewFilePath)
				}

				if isHeaderFileFunc(r.CurrentFilePath) {
					f.renamedHeaderFiles[r.CurrentFilePath] = r
					continue
				}
			} else {
				fmt.Println(err)
			}
		}
	}
}
