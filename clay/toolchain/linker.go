package toolchain

// Linker is an interface that defines the methods required for linking
// object and archive files into an executable.
type Linker interface {
	GenerateMapFile()
	AddLibraryPath(path string)
	AddLibraryFile(lib string)
	SetupArgs(userVars Vars)
	Link(inputArchiveAbsFilepaths []string, outputAppRelFilepathNoExt string) error
}

// Note: Linker dependency management.
// An executable is a collection of archive files that together form an executable.
// The linker before linking the list of archive files, should query the depTrackr
// to check if the archive files are up to date.
// After linking, the linker should add the executable file + the archive files as
// an item with dependencies to the depTrackr.
