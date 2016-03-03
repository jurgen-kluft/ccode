package vs2015_test

import (
	"testing"

	"github.com/jurgen-kluft/xcode/denv"
	"github.com/jurgen-kluft/xcode/vs2015"
)

// (projectname string, sourcefiles []string, headerfiles []string, platforms []string, configs []string, depprojectnames []string, vars vars.Variables, replacer vars.Replacer, writer ProjectWriter) {
func TestSimpleProject(t *testing.T) {

	//xunittestproject := denv.SetupDefaultCppProject("xunittest", "github.com\\jurgen-kluft")

	xbaseproject := denv.SetupDefaultCppProject("xbase", "github.com\\jurgen-kluft")
	//xbaseproject.Dependencies = append(xbaseproject.Dependencies, xunittestproject)

	xtestproject := denv.SetupDefaultCppProject("xtest", "D:\\dev.go\\src\\github.com\\jurgen-kluft")
	//xtestproject.Dependencies = append(xtestproject.Dependencies, xunittestproject)
	xtestproject.Dependencies = append(xtestproject.Dependencies, xbaseproject)

	vs2015.GenerateVisualStudio2015Solution(xtestproject)
}
