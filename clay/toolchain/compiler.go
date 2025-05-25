package toolchain

type CompilerConfig struct {
	Defines              []string // Compile defines
	Switches             []string // Compile switches (flags)
	WarningSwitches      []string // Warning switches (flags)
	ResponseFileFlags    string   // FilePath to the C compiler flags file (optional)
	ResponseFileDefines  string   // FilePath to the C compiler defines file (optional)
	ResponseFileIncludes string   // FilePath to the C compiler includes file (optional)
	PrefixIncludePaths   []string // Include paths for the compiler (system level)
	IncludePaths         []string // Include paths for the compiler (system level)
}

type Compiler interface {
	Compile(sourceFilepath string, objectFilepath string) error
}
