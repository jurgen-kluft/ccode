package axe

import (
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type PathParts struct {
	Driver string
	Dir    string
	Name   string
	Ext    string
}

func PathSlash() string {
	return string(os.PathSeparator)
}

func PathOtherSlash() string {
	slash := os.PathSeparator
	if slash == '/' {
		return "\\"
	}
	return "/"
}

func PathNormalize(path string) string {

	// if the path is empty, return it
	if len(path) < 1 {
		return path
	}

	// For each OS, adjust for the forward and backward slashes
	path = strings.ReplaceAll(path, PathSlash(), PathOtherSlash())

	// remove any '..' and trailing slashes
	// adjust for the platform we are on
	path = filepath.Clean(path)

	return path
}

func PathIsAbs(path string) bool {
	return filepath.IsAbs(path)
}

func PathDirname(path string) string {
	return filepath.Dir(path)
}

func PathBasename(path string, withExtension bool) string {

	path = PathNormalize(path)

	pivot := strings.LastIndexAny(path, "/\\")
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

func PathUp(path string) (parent, sub string) {
	path = PathNormalize(path)
	parent = filepath.Dir(path)
	sub = filepath.Base(path)
	return
}

func PathExtension(path string) string {
	path = PathNormalize(path)
	return filepath.Ext(path)
}

func PathSplit(path string) PathParts {
	var parts PathParts

	path = PathNormalize(path)

	parts.Driver = filepath.VolumeName(path)
	parts.Dir, parts.Name = filepath.Split(path)
	parts.Ext = filepath.Ext(parts.Name)

	if len(parts.Ext) == 0 {
		parts.Dir = filepath.Join(parts.Dir, parts.Name)
		parts.Name = ""
	}

	return parts
}

// PathSplitRelativeFilePath first makes sure the path is relative, then it splits
//
//	the path into each directory, filename and extension
func PathSplitRelativeFilePath(path string, splitFilenameAndExtension bool) []string {
	// e.g        /Documents/Books/Sci-fi/Asimov/IRobot.epub
	// split into [Documents, Books, Sci-fi, Asimov, IRobot, epub]

	path = PathNormalize(path)

	// make sure the path is relative
	if PathIsAbs(path) {
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

func PathMakeFullPath(dir, path string) string {
	dir = PathNormalize(dir)
	path = PathNormalize(path)
	if !PathIsAbs(path) {
		return PathGetAbs(path)
	}
	return PathGetAbs(filepath.Join(dir, path))
}

func PathGetAbs(path string) string {

	path = PathNormalize(path)

	// if the path is already absolute, return it
	if PathIsAbs(path) {
		return path
	}

	// if the path is relative, get the current directory
	// and append the path to it
	dir := PathGetCurrentDir()
	return filepath.Join(dir, path)
}

func PathGetRel(path, relativeTo string) string {
	if rel, err := filepath.Rel(relativeTo, path); err == nil {
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
		a += (a - 'A') + 'a'
	}
	if b >= 'A' && b <= 'Z' {
		b += (b - 'A') + 'a'
	}
	return a == b
}

// Make this UTF-8 safe
func PathMatchWildcard(path, wildcard string, ignoreCase bool) bool {
	pb := 0
	pe := len(path)

	wb := 0
	we := len(wildcard)

	for pb < pe && wb < we {

		pc, ps := utf8.DecodeRuneInString(path[pb:])
		wc, ws := utf8.DecodeRuneInString(wildcard[wb:])

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
			if wildcard[wb] == path[pb] {
				pb += ps
				wb += ws
				continue
			}
		}

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

		return false
	}

	if pb == pe && wb == we {
		return true
	}

	return false
}
