package toolchain

type ToolchainMsdev struct {
	ToolchainInstance
}

func NewToolchainMsdev() *ToolchainMsdev {
	return &ToolchainMsdev{}
}

func (ms *ToolchainMsdev) NewCCompiler(config string) Compiler {
	return nil
}
func (ms *ToolchainMsdev) NewCppCompiler(config string) Compiler {
	return nil
}
func (ms *ToolchainMsdev) NewArchiver(config string) Archiver {
	return nil
}
func (ms *ToolchainMsdev) NewLinker(config string) Linker {
	return nil
}

func (t *ToolchainMsdev) NewBurner(config string) Burner {
	return &ToolchainEmptyBurner{}
}
