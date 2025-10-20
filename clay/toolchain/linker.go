package toolchain

// Linker is an interface that defines the methods required for linking
// object and archive files into an executable.
type Linker interface {

	// Returns the final linked filename for the given filepath. Please
	// provide the path and name of the file without an extension.
	// e.g. "path/to/exe/name" will return "path/to/exe/name.exe" on Windows
	// and "path/to/exe/name" on Unix-like systems.
	LinkedFilepath(filepath string) string

	// SetupArgs prepares the linker arguments based on the provided options.
	SetupArgs(libraryPaths []string, libraryFiles []string)

	// Link takes a list of input object file paths and an output file path
	Link(inputObjectsAbsFilepaths, inputArchivesAbsFilepaths []string, outputAppRelFilepathNoExt string) error
}
