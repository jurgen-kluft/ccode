package vs2015_test

import (
	"path"
	"testing"

	"github.com/jurgen-kluft/xcode/denv"
	"github.com/jurgen-kluft/xcode/uid"
	"github.com/jurgen-kluft/xcode/vars"
	"github.com/jurgen-kluft/xcode/vs2015"
)

// (projectname string, sourcefiles []string, headerfiles []string, platforms []string, configs []string, depprojectnames []string, vars vars.Variables, replacer vars.Replacer, writer ProjectWriter) {
func TestSimpleProject(t *testing.T) {

	variables := vars.NewVars()
	replacer := vars.NewReplacer()

	variables.AddVar("xtest:OPTIMIZATION[Win32][DevDebug]", "Disabled") // Disabled or Full
	variables.AddVar("xtest:OPTIMIZATION[Win32][DevRelease]", "Full")   // Disabled or Full
	variables.AddVar("xtest:DEBUG_INFO[Win32][DevDebug]", "true")
	variables.AddVar("xtest:DEBUG_INFO[Win32][DevRelease]", "false")
	variables.AddVar("xtest:USE_DEBUG_LIBS[Win32][DevDebug]", "true")
	variables.AddVar("xtest:USE_DEBUG_LIBS[Win32][DevRelease]", "false")

	variables.AddVar("xtest:OPTIMIZATION[x64][DevDebug]", "Disabled") // Disabled or Full
	variables.AddVar("xtest:OPTIMIZATION[x64][DevRelease]", "Full")   // Disabled or Full
	variables.AddVar("xtest:DEBUG_INFO[x64][DevDebug]", "true")
	variables.AddVar("xtest:DEBUG_INFO[x64][DevRelease]", "false")
	variables.AddVar("xtest:USE_DEBUG_LIBS[x64][DevDebug]", "true")
	variables.AddVar("xtest:USE_DEBUG_LIBS[x64][DevRelease]", "false")

	variables.AddVar("xtest:GUID", uid.GetGUID("xtest"))

	variables.AddVar("TOOLSET[Win32]", "v140")
	variables.AddVar("TOOLSET[x64]", "v140")

	//xunittestproject := denv.SetupDefaultCppProject("xunittest", "github.com\\jurgen-kluft")

	xbaseproject := denv.SetupDefaultCppProject("xbase", "github.com\\jurgen-kluft")
	//xbaseproject.Dependencies = append(xbaseproject.Dependencies, xunittestproject)

	xtestproject := denv.SetupDefaultCppProject("xtest", "github.com\\jurgen-kluft")
	//xtestproject.Dependencies = append(xtestproject.Dependencies, xunittestproject)
	xtestproject.Dependencies = append(xtestproject.Dependencies, xbaseproject)

	xtestsln := denv.Solution{}
	xtestsln.Projects = make([]*denv.Project, 0)
	xtestsln.Projects = append(xtestsln.Projects, xtestproject)
	xtestsln.Projects = append(xtestsln.Projects, xbaseproject)
	//xtestsln.Projects = append(xtestsln.Projects, xunittestproject)

	for _, prj := range xtestsln.Projects {
		prjwriter := &denv.ProjectTextWriter{}
		prjwriter.Open(path.Join(prj.Path, prj.Name+".vcxproj"))
		vs2015.GenerateVisualStudio2015Project(prj, variables, replacer, prjwriter)
		prjwriter.Close()

		prjwriter.Open(path.Join(prj.Path, prj.Name+".vcxproj.filters"))
		vs2015.GenerateVisualStudio2015ProjectFilters(prj, prjwriter)
		prjwriter.Close()
	}

	writer := &denv.ProjectTextWriter{}
	writer.Open(path.Join(xtestproject.Path, "xtest.sln"))
	vs2015.GenerateVisualStudio2015Solution(xtestsln, writer)
	writer.Close()
}
