package toolchain

// Compiler is an interface that defines the methods required for compiling source files
type Compiler interface {
	AddDefine(define string)
	AddIncludePath(path string)
	SetupArgs(userVars Vars)
	Compile(sourceAbsFilepath string, sourceRelFilepath string) (objRelFilepath string, err error)
}
