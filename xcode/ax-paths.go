package xcode

import (
	"os"
	"path/filepath"
	"strings"
)

type PathParts struct {
	Driver string
	Dir    string
	Name   string
	Ext    string
}

func PathNormalize(path string) string {

	// if the path is empty, return it
	if len(path) < 1 {
		return path
	}

	// remove any '..' and trailing slashes
	// adjust for the platform we are on
	path = filepath.Clean(path)

	// if the path is absolute, normalize it
	if !filepath.IsAbs(path) {
		return filepath.Join(PathGetCurrentDir(), path)
	}

	return path
}

func PathIsAbs(path string) bool {
	return filepath.IsAbs(path)
}

func PathDirname(path string) string {
	return filepath.Dir(path)
}

func PathBasename(path string, withExtension bool) string {

	pivot := strings.IndexAny(path, "/\\")
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

	return path[pivot:dot]
}

func PathUp(path string) (parent, sub string) {
	parent = filepath.Dir(path)
	sub = filepath.Base(path)
	return
}

func PathExtension(path string) string {
	return filepath.Ext(path)
}

func PathSplit(path string) PathParts {
	var parts PathParts

	parts.Driver = filepath.VolumeName(path)
	parts.Dir, parts.Name = filepath.Split(path)
	parts.Ext = filepath.Ext(parts.Name)

	if len(parts.Ext) == 0 {
		parts.Dir = filepath.Join(parts.Dir, parts.Name)
		parts.Name = ""
	}

	return parts
}

func PathMakeFullPath(dir, path string) string {
	if !PathIsAbs(path) {
		return PathGetAbs(path)
	}
	return PathGetAbs(filepath.Join(dir, path))
}

func PathGetAbs(path string) string {

	// if the path is already absolute, return it
	if PathIsAbs(path) {
		return path
	}

	// if the path is relative, get the current directory
	// and append the path to it
	dir := PathGetCurrentDir()
	return PathGetAbs(dir + "/" + path)
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
