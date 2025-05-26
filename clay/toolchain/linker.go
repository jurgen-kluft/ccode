package toolchain

type Linker interface {
	GenerateMapFile()
	AddLibraryPath(path string)
	AddLibraryFile(lib string)
	SetupArgs(userVars Vars)
	Link(inputArchiveAbsFilepaths []string, outputAppRelFilepathNoExt string) error
}
