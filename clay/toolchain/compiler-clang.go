package toolchain

type CompilerClang struct {
	Path string // The path to the clang compiler executable
}

func NewClangCompiler(path string) *CompilerClang {
	return nil
}

func (cl *CompilerClang) Compile(sourceFilepath string, objectFilepath string) error {

	return nil
}
