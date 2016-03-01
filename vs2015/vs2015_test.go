package vs2015_test

import (
	"github.com/jurgen-kluft/xcode/denv"
	"github.com/jurgen-kluft/xcode/glob"
	"github.com/jurgen-kluft/xcode/uid"
	"github.com/jurgen-kluft/xcode/vars"
	"github.com/jurgen-kluft/xcode/vs2015"
	"os"
	"path"
	"testing"
)

type TestProjectWriter struct {
	fhnd *os.File
}

func (writer *TestProjectWriter) Open(filepath string) (err error) {
	writer.fhnd, err = os.OpenFile(filepath, os.O_CREATE|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	return nil
}
func (writer *TestProjectWriter) Close() (err error) {
	err = writer.fhnd.Close()
	return err
}

func (writer *TestProjectWriter) Write(str string) (err error) {
	_, err = writer.fhnd.WriteString(str)
	return err
}

// (projectname string, sourcefiles []string, headerfiles []string, platforms []string, configs []string, depprojectnames []string, vars vars.Variables, replacer vars.Replacer, writer ProjectWriter) {
func TestSimpleProject(t *testing.T) {
	platforms := []string{"Win32", "x64"}
	configs := []string{"DevDebug", "DevRelease"}

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

	xunittestproject := denv.Project{Name: "xunittest"}
	xunittestproject.GUID = uid.GetGUID(xunittestproject.Name)
	xunittestproject.ProjectID = vs2015.CPPprojectID
	xunittestproject.Platforms = platforms
	xunittestproject.Configs = configs

	xbaseproject := denv.Project{Name: "xbase"}
	xbaseproject.Dependencies = []denv.Project{xunittestproject}
	xbaseproject.GUID = uid.GetGUID(xbaseproject.Name)
	xbaseproject.ProjectID = vs2015.CPPprojectID
	xbaseproject.Platforms = platforms
	xbaseproject.Configs = configs

	xtestproject := denv.Project{Name: "xtest"}
	xtestproject.Dependencies = []denv.Project{}
	xtestproject.GUID = uid.GetGUID(xtestproject.Name)
	xtestproject.ProjectID = vs2015.CPPprojectID
	xtestproject.Platforms = platforms
	xtestproject.Configs = configs
	xtestproject.Path = "d:\\Test.xcode\\xtest"
	xtestproject.SrcGlobPaths = []string{"source\\main\\^cpp\\**\\*.cpp"}
	xtestproject.HdrGlobPaths = []string{"source\\main\\^include\\**\\*.h"}
	xtestproject.SrcFiles, _ = glob.GlobFiles(xtestproject.Path, xtestproject.SrcGlobPaths)
	xtestproject.HdrFiles, _ = glob.GlobFiles(xtestproject.Path, xtestproject.HdrGlobPaths)

	xunittestproject.Path = "d:\\Test.xcode\\xtest\\vendor\\xunittest"
	xunittestproject.SrcGlobPaths = []string{"source\\main\\^cpp\\**\\*.cpp"}
	xunittestproject.HdrGlobPaths = []string{"source\\main\\^include\\**\\*.h"}
	xunittestproject.SrcFiles, _ = glob.GlobFiles(xunittestproject.Path, xunittestproject.SrcGlobPaths)
	xunittestproject.HdrFiles, _ = glob.GlobFiles(xunittestproject.Path, xunittestproject.HdrGlobPaths)

	xbaseproject.Path = "d:\\Test.xcode\\xtest\\vendor\\xbase"
	xbaseproject.SrcGlobPaths = []string{"source\\main\\^cpp\\**\\*.cpp"}
	xbaseproject.HdrGlobPaths = []string{"source\\main\\^include\\**\\*.h"}
	xbaseproject.SrcFiles, _ = glob.GlobFiles(xbaseproject.Path, xbaseproject.SrcGlobPaths)
	xbaseproject.HdrFiles, _ = glob.GlobFiles(xbaseproject.Path, xbaseproject.HdrGlobPaths)

	xtestsln := denv.Solution{}
	xtestsln.Projects = make([]denv.Project, 0)
	xtestsln.Projects = append(xtestsln.Projects, xtestproject)
	xtestsln.Projects = append(xtestsln.Projects, xbaseproject)
	xtestsln.Projects = append(xtestsln.Projects, xunittestproject)

	for _, prj := range xtestsln.Projects {
		prjwriter := &TestProjectWriter{}
		prjwriter.Open(path.Join(prj.Path, prj.Name+".vcxproj"))
		vs2015.GenerateVisualStudio2015Project(prj, variables, replacer, prjwriter)
		prjwriter.Close()

		prjwriter.Open(path.Join(prj.Path, prj.Name+".vcxproj.filters"))
		vs2015.GenerateVisualStudio2015ProjectFilters(prj, prjwriter)
		prjwriter.Close()
	}

	writer := &TestProjectWriter{}
	writer.Open(path.Join(xtestproject.Path, "xtest.sln"))
	vs2015.GenerateVisualStudio2015Solution(xtestsln, writer)
	writer.Close()
}
