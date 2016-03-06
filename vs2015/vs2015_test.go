package vs2015_test

import (
	"github.com/jurgen-kluft/xcode/denv"
	"github.com/jurgen-kluft/xcode/vs2015"
	"os"
	"testing"
)

// (projectname string, sourcefiles []string, headerfiles []string, platforms []string, configs []string, depprojectnames []string, vars vars.Variables, replacer vars.Replacer, writer ProjectWriter) {
func TestSimpleProject(t *testing.T) {

	//xunittestproject := denv.SetupDefaultCppProject("xunittest", "github.com\\jurgen-kluft")
	xbaseproject := denv.SetupDefaultCppProject("xbase", "github.com/jurgen-kluft")
	//xbaseproject.Dependencies = append(xbaseproject.Dependencies, xunittestproject)

	xtestproject := denv.SetupDefaultCppProject("xtest", "/Users/Jurgen/golang/src/github.com/jurgen-kluft")
	//xtestproject.Dependencies = append(xtestproject.Dependencies, xunittestproject)
	xtestproject.Dependencies = append(xtestproject.Dependencies, xbaseproject)

	os.Chdir(xtestproject.Path)

	vs2015.GenerateVisualStudio2015Solution(xtestproject)
}
