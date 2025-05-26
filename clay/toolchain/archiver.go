package toolchain

// Archiver (also known as Lib) is an interface that defines the methods required for
// creating an archive (.a, .lib) file.
type Archiver interface {
	// SetupArgs prepares the arguments for the archiver based on user-defined variables.
	// It should be called before using the Archive method.
	SetupArgs(userVars Vars)

	// Archive takes a list of input object file paths and an output archive file path.
	// Both paths are relative to the build path.
	Archive(inputObjAbsFilepaths []string, outputArchiveRelFilepath string) error
}
