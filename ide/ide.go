package ide

import "os"

func GenerateIde() {
	workspacePath := "$HOME/dev.go/src/github.com/jurgen-kluft"
	workspacePath = os.ExpandEnv(workspacePath)
	TestRun(workspacePath, "cbase")
}
