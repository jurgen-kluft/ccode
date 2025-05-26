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
