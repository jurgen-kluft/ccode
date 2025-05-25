package toolchain

type Toolchain interface {
	GetCompiler() *Compiler
	GetArchiver() *Archiver
	GetLinker() *Linker
}
