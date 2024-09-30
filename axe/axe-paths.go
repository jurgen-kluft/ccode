package axe

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// ----------------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------------

type PinnedPath struct {
	Root string
	Path string
}

func (fp *PinnedPath) String() string {
	return path.Join(fp.Root, fp.Path)
}

func (fp *PinnedPath) RelativeTo(root string) string {
	return PathGetRelativeTo(fp.String(), root)
}

// ----------------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------------

type PinnedPathSet struct {
	Entries map[string]int
	Values  []*PinnedPath
}

func NewPinnedPathSet() *PinnedPathSet {
	d := &PinnedPathSet{}
	d.Entries = make(map[string]int)
	d.Values = make([]*PinnedPath, 0)
	return d
}

func (d *PinnedPathSet) Merge(other *PinnedPathSet) {
	for _, value := range other.Values {
		d.AddOrSet(value.Root, value.Path)
	}
}

func (d *PinnedPathSet) Copy() *PinnedPathSet {
	c := NewPinnedPathSet()
	c.Merge(d)
	return c
}

func (d *PinnedPathSet) Extend(rhs *PinnedPathSet) {
	for _, fp := range rhs.Values {
		d.AddOrSet(fp.Root, fp.Path)
	}
}

func (d *PinnedPathSet) UniqueExtend(rhs *PinnedPathSet) {
	for _, fp := range rhs.Values {
		fullpath := fp.String()
		if _, ok := d.Entries[fullpath]; !ok {
			d.AddOrSet(fp.Root, fp.Path)
		}
	}
}

func (d *PinnedPathSet) AddOrSet(base string, dir string) {
	fp := &PinnedPath{Root: base, Path: dir}
	fullpath := fp.String()
	i, ok := d.Entries[fullpath]
	if !ok {
		d.Entries[fullpath] = len(d.Values)
		d.Values = append(d.Values, fp)
	} else if strings.Compare(d.Values[i].String(), fullpath) != 0 {
		d.Values[i] = fp
	}
}

// Enumerate will call the enumerator function for each key-value pair in the dictionary.
//
//	'last' will be 0 for all but the last key-value pair, and 1 for the last key-value pair.
func (d *PinnedPathSet) Enumerate(enumerator func(i int, base string, dir string, last int)) {
	for i, fp := range d.Values {
		base := fp.Root
		dir := fp.Path
		last := i / (len(d.Values) - 1)
		enumerator(i, base, dir, last)
	}
}

func (d *PinnedPathSet) Concatenated(prefix string, suffix string, modifier func(base string, dir string) string) string {
	concat := ""
	for _, fp := range d.Values {
		newFullPath := modifier(fp.Root, fp.Path)
		concat += prefix + newFullPath + suffix
	}
	return concat
}

// ----------------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------------

// PathSlash returns the slash of the current OS
func PathSlash() string {
	return string(os.PathSeparator)
}

// PathNormalize returns a normalized path, fixing slashes and removing '..' and trailing slashes
func PathNormalize(path string) string {

	// if the path is empty, return it
	if len(path) < 1 {
		return path
	}

	// adjust for the forward and backward slashes
	if os.PathSeparator == '\\' {
		path = strings.Replace(path, "/", "\\", -1)
	} else {
		path = strings.Replace(path, "\\", "/", -1)
	}

	// remove any '..' and trailing slashes
	// adjust for the platform we are on
	path = filepath.Clean(path)

	return path
}

// PathDirname returns the directory from the path
func PathDirname(path string) string {
	return filepath.Dir(path)
}

// PathFilename returns the filename from the path with or without the extension
func PathFilename(path string, withExtension bool) string {

	path = PathNormalize(path)

	pivot := strings.LastIndexAny(path, PathSlash())
	if pivot < 0 {
		pivot = 0
	} else {
		pivot++
	}

	if withExtension {
		return path[pivot:]
	}

	// Search backwards for the last '.' character but not beyond pivot
	dot := strings.LastIndex(path[pivot:], ".")
	if dot < 0 {
		return path[pivot:]
	}

	return path[pivot : pivot+dot]
}

// PathFileExtension returns the extension of the file in the path
func PathFileExtension(path string) string {
	path = PathNormalize(path)
	return filepath.Ext(path)
}

// PathUp returns the parent directory and the sub directory
func PathUp(path string) (parent, sub string) {
	path = PathNormalize(path)
	parent = filepath.Dir(path)
	sub = filepath.Base(path)
	return
}

// PathParent returns the parent directory
func PathParent(path string) string {
	path = PathNormalize(path)
	return filepath.Dir(path)
}

// PathSplitRelativeFilePath first makes sure the path is relative, then it splits
//
//	the path into each directory, filename and extension
func PathSplitRelativeFilePath(path string, splitFilenameAndExtension bool) []string {
	// e.g        Documents/Books/Sci-fi/Asimov/IRobot.epub
	// split into [Documents, Books, Sci-fi, Asimov, IRobot, epub]

	path = PathNormalize(path)

	// make sure the path is relative
	if filepath.IsAbs(path) {
		return nil
	}

	parts := []string{}
	parts = strings.Split(path, PathSlash()) // split the path into parts where the last part is the filename

	if splitFilenameAndExtension { // do we keep the filename as it is or split it into filename and extension
		filename := parts[len(parts)-1]                         // Get the filename
		ext := filepath.Ext(filename)                           // Get the extension of the filename
		parts[len(parts)-1] = strings.TrimSuffix(filename, ext) // Remove the extension from the filename
		parts = append(parts, ext)                              // Add the extension to the parts
	}
	return parts
}

func PathGetRelativeTo(path, root string) string {
	if rel, err := filepath.Rel(root, path); err == nil {
		return rel
	}
	return ""
}

func PathWindowsPath(path string) (outPath string) {
	outPath = strings.ReplaceAll(path, "/", "\\")
	return
}

func PathGetCurrentDir() string {
	pwd, err := os.Getwd()
	if err == nil {
		return pwd
	}
	return ""
}

func PathMakeDir(path string) bool {
	if PathDirExists(path) {
		return false
	}

	parent := PathDirname(path)
	if parent != "" {
		PathMakeDir(parent)
	}

	err := os.Mkdir(path, os.ModePerm)
	if err != nil {
		return false
	}
	return true
}

func PathFileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func PathDirExists(path string) bool {
	if fi, err := os.Stat(path); err == nil {
		return fi.IsDir()
	}
	return false
}

func PathIsDir(path string) bool {
	if fi, err := os.Stat(path); err == nil {
		return fi.IsDir()
	}
	return false
}

func PathIsFile(path string) bool {
	if fi, err := os.Stat(path); err == nil {
		return !fi.IsDir()
	}
	return false
}

func MatchCharCaseInsensitive(a rune, b rune) bool {
	if a >= 'A' && a <= 'Z' {
		a = (a - 'A') + 'a'
	}
	if b >= 'A' && b <= 'Z' {
		b = (b - 'A') + 'a'
	}
	return a == b
}
func PathMatchWildcard(path, wildcard string, ignoreCase bool) bool {
	pb := 0
	pe := len(path)

	wb := 0
	we := len(wildcard)

	for pb < pe && wb < we {

		pc, ps := utf8.DecodeRuneInString(path[pb:])
		wc, ws := utf8.DecodeRuneInString(wildcard[wb:])

		if wc == '*' {
			w1 := wb + ws
			if w1 >= we {
				return true
			}

			p1 := pb + ps
			if p1 >= pe {
				return false
			}

			pb = p1

			pc, ps = utf8.DecodeRuneInString(path[pb:])
			wc, ws = utf8.DecodeRuneInString(wildcard[w1:])

			if pc == wc {
				wb = w1
			}

			continue
		}

		if ignoreCase {
			if MatchCharCaseInsensitive(wc, pc) {
				pb += ps
				wb += ws
				continue
			}
		} else {
			if wc == pc {
				pb += ps
				wb += ws
				continue
			}
		}

		if wc == '?' {
			pb += ps
			wb += ws
			continue
		}

		return false
	}

	if pb == pe && wb == we {
		return true
	}

	return false
}

func PathMatchWildcardOptimized(path, wildcard string, ignoreCase bool) bool {
	pb := 0
	pe := len(path)

	wb := 0
	we := len(wildcard)

	for pb < pe && wb < we {

		pc, ps := utf8.DecodeRuneInString(path[pb:])
		wc, ws := utf8.DecodeRuneInString(wildcard[wb:])

		if wc == '*' {

			w1 := wb + ws
			if w1 >= we {
				return true
			}
			wc, ws = utf8.DecodeRuneInString(wildcard[w1:])

			pb = pb + ps
			for pb < pe {
				pc, ps = utf8.DecodeRuneInString(path[pb:])
				if pc == wc {
					goto next
				}
				pb += ps
			}
			return false

		next:
			wb = w1
			continue
		}

		if wc == '?' {
			pb += ps
			wb += ws
			continue
		}

		if ignoreCase {
			if MatchCharCaseInsensitive(wc, pc) {
				pb += ps
				wb += ws
				continue
			}
		} else {
			if wc == pc {
				pb += ps
				wb += ws
				continue
			}
		}

		return false
	}

	if pb == pe && wb == we {
		return true
	}

	return false
}
