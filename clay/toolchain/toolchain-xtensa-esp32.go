package toolchain

import (
	_ "embed"
)

type ToolchainXtensaEsp32 struct {
	CompilerConfig *CompilerConfig // The compiler configuration
	ArchiverConfig *ArchiverConfig // The archiver configuration
	LinkerConfig   *LinkerConfig   // The linker configuration
}

//go:embed toolchain-xtensa.esp32.config
var toolchainXtensaEsp32Config string

func NewToolchainXtensaEsp32() *ToolchainXtensaEsp32 {
	return &ToolchainXtensaEsp32{}
}

func (ms *ToolchainXtensaEsp32) GetCompiler() *Compiler {
	return nil
}
func (ms *ToolchainXtensaEsp32) GetArchiver() *Archiver {
	return nil
}
func (ms *ToolchainXtensaEsp32) GetLinker() *Linker {
	return nil
}
