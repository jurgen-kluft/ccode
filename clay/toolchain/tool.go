package toolchain

type Tool interface {
	SetupArgs(userVars Vars, config string)
	Execute() error
}
