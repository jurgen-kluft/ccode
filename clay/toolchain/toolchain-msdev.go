package toolchain

type ToolchainMsdev struct {
	ToolchainInstance
}

// Compiler options
//      https://github.com/MicrosoftDocs/cpp-docs/blob/main/docs/build/reference/compiler-options-listed-by-category.md
// Linker options
//      https://github.com/MicrosoftDocs/cpp-docs/blob/main/docs/build/reference/linker-options.md

func NewToolchainMsdev() *ToolchainMsdev {
	return &ToolchainMsdev{}
}

func (ms *ToolchainMsdev) NewCompiler(config *Config) Compiler {
	return nil
}
func (ms *ToolchainMsdev) NewArchiver(a ArchiverType, config *Config) Archiver {
	return nil
}
func (ms *ToolchainMsdev) NewLinker(config *Config) Linker {
	return nil
}

func (t *ToolchainMsdev) NewBurner(config *Config) Burner {
	return &ToolchainEmptyBurner{}
}
