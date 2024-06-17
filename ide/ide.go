package ide

import "os"
import "runtime"

func GenerateMsDevIde() {
	workspacePath := "$HOME/dev.go/src/github.com/jurgen-kluft"
	if runtime.GOOS == "windows" {
		workspacePath = "d:\\Dev.Go\\src\\github.com\\jurgen-kluft"
	}
	workspacePath = os.ExpandEnv(workspacePath)
	generator := NewMsDevTestGenerator()
	generator.TestRun(workspacePath, "cbase")
}

func GenerateXcodeIde() {
	workspacePath := "$HOME/dev.go/src/github.com/jurgen-kluft"
	workspacePath = os.ExpandEnv(workspacePath)
	generator := NewXcodeTestGenerator()
	generator.TestRun(workspacePath, "cbase")
}
