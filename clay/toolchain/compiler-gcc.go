package toolchain

type CompilerGcc struct {
	Path string // The path to the clang compiler executable
}

func NewGccCompiler(path string) *CompilerGcc {

	return nil
}

func (cl *CompilerGcc) Compile(sourceFilepath string, objectFilepath string) error {

	return nil
}
