package toolchain

type CompilerMsdev struct {
	Path string // The path to the clang compiler executable
}

func (cl *CompilerMsdev) Compile(sourceFilepath string, objectFilepath string) error {

	return nil
}
