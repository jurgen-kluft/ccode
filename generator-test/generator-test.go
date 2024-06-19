package generator_test

import (
	"os"
	"runtime"
)

// ----------------------------------------------------------------------------------------------
// IDE generator tests
// ----------------------------------------------------------------------------------------------

func TestGenerateMsDevIde() {
	workspacePath := "$HOME/dev.go/src/github.com/jurgen-kluft"
	if runtime.GOOS == "windows" {
		workspacePath = "d:\\Dev.Go\\src\\github.com\\jurgen-kluft"
	}
	workspacePath = os.ExpandEnv(workspacePath)
	generator := NewMsDevTestGenerator()
	generator.TestRun(workspacePath, "cbase")
}

func TestGenerateTundra() {
	workspacePath := "$HOME/dev.go/src/github.com/jurgen-kluft"
	if runtime.GOOS == "windows" {
		workspacePath = "d:\\Dev.Go\\src\\github.com\\jurgen-kluft"
	}
	workspacePath = os.ExpandEnv(workspacePath)
	generator := NewTundraTestGenerator()
	generator.TestRun(workspacePath, "cbase")
}

func TestGenerateXcode() {
	workspacePath := "$HOME/dev.go/src/github.com/jurgen-kluft"
	workspacePath = os.ExpandEnv(workspacePath)
	generator := NewXcodeTestGenerator()
	generator.TestRun(workspacePath, "cbase")
}
