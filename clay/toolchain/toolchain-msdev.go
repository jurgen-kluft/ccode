package toolchain

type ToolchainMsdev struct {
	CompilerConfig *CompilerConfig // The compiler configuration
	ArchiverConfig *ArchiverConfig // The archiver configuration
	LinkerConfig   *LinkerConfig   // The linker configuration
}

func NewToolchainMsdev() *ToolchainMsdev {
	return &ToolchainMsdev{}
}

func (ms *ToolchainMsdev) GetCompiler() *Compiler {
	return nil
}
func (ms *ToolchainMsdev) GetArchiver() *Archiver {
	return nil
}
func (ms *ToolchainMsdev) GetLinker() *Linker {
	return nil
}
