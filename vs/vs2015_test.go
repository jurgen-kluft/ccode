package vs_test

import (
	"os"
	"testing"

	"github.com/jurgen-kluft/ccode/denv"
	"github.com/jurgen-kluft/ccode/vs"
)

// (projectname string, sourcefiles []string, headerfiles []string, platforms []string, configs []string, depprojectnames []string, vars vars.Variables, replacer vars.Replacer, writer ProjectWriter) {
func TestSimpleProject(t *testing.T) {

	cunittestproject := denv.SetupDefaultCppLibProject("cunittest", "github.com\\jurgen-kluft\\cunittest")

	cbaseproject := denv.SetupDefaultCppLibProject("cbase", "github.com\\jurgen-kluft\\cbase")
	cbaseproject.Dependencies = append(cbaseproject.Dependencies, cunittestproject)

	ctestproject := denv.SetupDefaultCppLibProject("ctest", "github.com\\jurgen-kluft\\ctest")
	ctestproject.Type = denv.Executable
	ctestproject.Dependencies = append(ctestproject.Dependencies, cunittestproject)
	ctestproject.Dependencies = append(ctestproject.Dependencies, cbaseproject)

	// Since we are running the test at the wrong location we need to change
	// the current work directory to the actual package directory
	os.Chdir("/Users/Jurgen/golang/src/github.com/jurgen-kluft/ctest")
	//os.Chdir(denv.Path("D:/Dev.Go/src/github.com/jurgen-kluft/ctest"))

	vs.GenerateVisualStudio2015Solution(ctestproject)
}
