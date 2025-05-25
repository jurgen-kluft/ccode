package toolchain

type ArchiverConfig struct {
	Switches              []string // Compile switches (flags)
	LibraryPaths          []string // Library paths for the linker (system)
	Libraries             []string // Libraries to link against
	OutputMapFile         bool     // Whether to generate a map file for the linker output
	ResponseFileLdFlags   string   // FilePath to the linker flags file (optional)
	ResponseFileLdScripts string   // FilePath to the linker scripts file (optional)
	ResponseFileLdLibs    string   // FilePath to the linker libraries file (optional)
}

type Archiver interface {
	Archive(inputObjFilepaths []string, outputArchiveFilepath string) error
}
