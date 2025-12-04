package toolchain

import (
	"io"
	"os"
	"path/filepath"
)

// FileCommander is an interface that defines the methods for copying files
// to the build directory.
type FileCommander interface {
	Setup(buildPath string)

	// CopyDir copies the specified directory to the build directory.
	// The fileFilter function is used to determine which files to copy (true = copy, false = skip).
	// The dirFilter function is used to determine which directories to traverse (true = traverse, false = skip).
	// It returns slices of source and destination absolute file paths that were copied.
	CopyDir(dir string, fileFilter func(file string) bool, dirFilter func(file string) bool) ([]string, []string, error)

	// CopyFiles copies the specified files to the build directory.
	// The specified dir is the root directory from which the files are copied.
	// The files slice contains the relative paths of the files to be copied.
	CopyFiles(dir string, files []string) error
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Empty FileCommander

type EmptyFileCommander struct {
	// For some toolchains, this may be empty
}

func (cl *EmptyFileCommander) Setup(buildPath string) {
}

func (cl *EmptyFileCommander) CopyDir(dir string, fileFilter func(file string) bool, dirFilter func(file string) bool) ([]string, []string, error) {
	return []string{}, []string{}, nil
}

func (cl *EmptyFileCommander) CopyFiles(dir string, files []string) error {
	return nil
}

// --------------------------------------------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
// Basic FileCommander

type BasicFileCommander struct {
	buildPath string
}

func (cl *BasicFileCommander) Setup(buildPath string) {
	cl.buildPath = buildPath
}

func (cl *BasicFileCommander) CopyDir(dir string, fileFilter func(file string) bool, dirFilter func(file string) bool) (srcFiles []string, dstFiles []string, result error) {
	relFiles := []string{}
	result = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if it's a directory
		if info.IsDir() {
			// Apply the dirFilter
			if !dirFilter(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// Apply the fileFilter
		if !fileFilter(path) {
			return nil
		}

		// Get the relative path of the file
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		relFiles = append(relFiles, relPath)
		return nil
	})

	if result != nil {
		return nil, nil, result
	}

	// Now copy the collected files
	result = cl.CopyFiles(dir, relFiles)

	if result != nil {
		return nil, nil, result
	}

	srcFiles = []string{}
	dstFiles = []string{}
	for _, relFile := range relFiles {
		srcFiles = append(srcFiles, filepath.Join(dir, relFile))
		dstFiles = append(dstFiles, filepath.Join(cl.buildPath, relFile))
	}

	return srcFiles, dstFiles, nil
}

func (cl *BasicFileCommander) CopyFiles(dir string, files []string) error {
	for _, file := range files {
		srcFile := filepath.Join(dir, file)
		destFile := filepath.Join(cl.buildPath, srcFile)

		// Open the source file
		src, err := os.Open(srcFile)
		if err != nil {
			return err
		}
		src.Close()

		// Create the destination file
		dest, err := os.Create(destFile)
		if err != nil {
			return err
		}
		defer dest.Close()

		// Copy the contents from source to destination
		_, err = io.Copy(dest, src)
		if err != nil {
			return err
		}
	}
	return nil
}
