package fixr

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type FileRename struct {
	CurrentFilePath string
	NewFilePath     string
}

func (f *Fixr) globAndRename(dirpath string, renamer func(_filepath string) (renameNeeded bool, renamedFilepath string), isSourceFile func(_filepath string) bool, isHeaderFile func(_filepath string) bool) {

	filesToRename := make([]FileRename, 0, 16)

	err := filepath.WalkDir(dirpath, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			path = filepath.Dir(path)
			relpath, _ := filepath.Rel(dirpath, path)
			_filepath := filepath.Join(relpath, d.Name())
			if isSourceFile(_filepath) || isHeaderFile(_filepath) {
				renameNeeded, renamedFilepath := renamer(_filepath)
				if renameNeeded {
					filesToRename = append(filesToRename, FileRename{CurrentFilePath: _filepath, NewFilePath: renamedFilepath})
				}
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
			if !f.DryRun() {
				err = os.Rename(r.CurrentFilePath, r.NewFilePath)
			} else {
				err = nil
			}
			if err == nil {
				if f.Verbose() {
					fmt.Printf(" * renamed file   %s  ->  %s\n", r.CurrentFilePath, r.NewFilePath)
				}

				if isHeaderFile(r.CurrentFilePath) {
					f.renamedHeaderFiles[r.CurrentFilePath] = r
					continue
				}
			} else {
				fmt.Println(err)
			}
		}
	}
}
