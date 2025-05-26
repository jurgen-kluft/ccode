package toolchain

type Compiler interface {
	AddDefine(define string)
	AddIncludePath(path string)
	SetupArgs(userVars Vars)
	Compile(sourceAbsFilepath string, sourceRelFilepath string) (string, error)
}
