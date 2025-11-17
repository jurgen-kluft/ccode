package fixr

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type rootTree struct {
	dir          *fdir
	dirs         map[string]*fdir
	files        map[string]*ffile
	filenames    map[string][]*ffile
	fuzzyMatcher *ClosestMatch
}

func newRoot() *rootTree {
	rootDir := &fdir{path: "."}
	r := &rootTree{
		dir:       rootDir,
		dirs:      map[string]*fdir{".": rootDir},
		files:     map[string]*ffile{},
		filenames: map[string][]*ffile{},
	}
	return r
}

func (r *rootTree) finalize() {
	possible := make([]string, 0, len(r.files))
	for _, file := range r.files {
		possible = append(possible, file.path)
	}
	r.fuzzyMatcher = NewClosestMatch(possible, []int{2, 3, 4})
	fmt.Printf("Include header file structure, fuzzy matching accuracy: %f\n", r.fuzzyMatcher.AccuracyMutatingWords())
}

func (r *rootTree) addFile(file *ffile) {
	r.files[strings.ToLower(file.path)] = file
	r.filenames[strings.ToLower(file.name)] = append(r.filenames[file.name], file)
}

func (r *rootTree) scanDir(dirpath string, filter func(filename string) bool) error {

	err := filepath.WalkDir(dirpath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		path = path[len(dirpath):]
		path = strings.TrimPrefix(path, "/")
		path = strings.TrimPrefix(path, "\\")
		if len(path) == 0 {
			path = "."
		}

		if d.IsDir() {
			if _, ok := r.dirs[path]; !ok {
				parentPath := ParentPath(path)
				parent := r.dirs[parentPath]
				dir := newFDir(d.Name(), path, parent)
				r.dirs[path] = dir
			}
		} else {
			path = filepath.Dir(path)

			_filepath := filepath.Join(path, d.Name())
			if filter(_filepath) {
				if dir, ok := r.dirs[path]; !ok {
					parentPath := ParentPath(path)
					parent := r.dirs[parentPath]
					dir = newFDir(d.Name(), path, parent)
					r.dirs[path] = dir
					file := newFFile(d.Name(), _filepath, dir)
					dir.files = append(dir.files, file)
					r.addFile(file)
				} else {
					file := newFFile(d.Name(), _filepath, dir)
					dir.files = append(dir.files, file)
					r.addFile(file)
				}
			}
		}
		return nil
	})

	return err
}
