package vs_test

import (
	"os"
	"testing"

	"github.com/jurgen-kluft/ccode/denv"
	"github.com/jurgen-kluft/ccode/vs"
)

// (projectname string, sourcefiles []string, headerfiles []string, platforms []string, configs []string, depprojectnames []string, vars vars.Variables, replacer vars.Replacer, writer ProjectWriter) {
func TestSimpleProject(t *testing.T) {

	xunittestproject := denv.SetupDefaultCppLibProject("xunittest", "github.com\\jurgen-kluft\\xunittest")

	xbaseproject := denv.SetupDefaultCppLibProject("xbase", "github.com\\jurgen-kluft\\xbase")
	xbaseproject.Dependencies = append(xbaseproject.Dependencies, xunittestproject)

	xtestproject := denv.SetupDefaultCppLibProject("xtest", "github.com\\jurgen-kluft\\xtest")
	xtestproject.Type = denv.Executable
	xtestproject.Dependencies = append(xtestproject.Dependencies, xunittestproject)
	xtestproject.Dependencies = append(xtestproject.Dependencies, xbaseproject)

	// Since we are running the test at the wrong location we need to change
	// the current work directory to the actual package directory
	os.Chdir("/Users/Jurgen/golang/src/github.com/jurgen-kluft/xtest")
	//os.Chdir(denv.Path("D:/Dev.Go/src/github.com/jurgen-kluft/xtest"))

	vs.GenerateVisualStudio2015Solution(xtestproject)
}
