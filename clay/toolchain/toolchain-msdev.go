package toolchain

type ToolchainMsdev struct {
	ToolchainInstance
}

func NewToolchainMsdev() *ToolchainMsdev {
	return &ToolchainMsdev{}
}

func (ms *ToolchainMsdev) NewCCompiler(config *Config) Compiler {
	return nil
}
func (ms *ToolchainMsdev) NewCppCompiler(config *Config) Compiler {
	return nil
}
func (ms *ToolchainMsdev) NewArchiver(config *Config) Archiver {
	return nil
}
func (ms *ToolchainMsdev) NewLinker(config *Config) Linker {
	return nil
}

func (t *ToolchainMsdev) NewBurner(config *Config) Burner {
	return &ToolchainEmptyBurner{}
}
