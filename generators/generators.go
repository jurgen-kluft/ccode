package ide_generators

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jurgen-kluft/ccode/denv"
)

// ----------------------------------------------------------------------------------------------
// IDE generator
// ----------------------------------------------------------------------------------------------
type Generator struct {
	Dev             DevEnum
	BuildTarget     denv.BuildTarget
	Verbose         bool
	WorkspacePath   string // $(GOPATH)/src/github.com/user, absolute path
	ExclusionFilter *denv.ExclusionFilter
}

func NewGenerator(dev string, buildTarget denv.BuildTarget, verbose bool) *Generator {
	g := &Generator{}
	g.Dev = NewDevEnum(dev)
	g.BuildTarget = buildTarget
	g.Verbose = verbose
	return g
}

func (g *Generator) Generate(pkg *denv.Package) {
	var ws *Workspace
	var err error

	if ws, err = g.GenerateWorkspace(pkg, g.Dev, g.BuildTarget); err != nil {
		fmt.Printf("Error generating workspace: %v\n", err)
		return
	}

	switch g.Dev {
	case DevTundra:
		gg := NewTundraGenerator(ws)
		gg.Generate()
	case DevMake:
		gg := NewMakeGenerator2(ws)
		err = gg.Generate()
	case DevXcode:
		gg := NewXcodeGenerator(ws)
		gg.Generate()
	case DevVs2015, DevVs2017, DevVs2019, DevVs2022:
		gg := NewMsDevGenerator(ws)
		gg.Generate(g.BuildTarget)
	case DevClay:
		gg := NewClayGenerator(ws, g.Verbose)
		err = gg.Generate()
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func (g *Generator) GenerateWorkspace(_pkg *denv.Package, _dev DevEnum, _buildTarget denv.BuildTarget) (*Workspace, error) {
	g.WorkspacePath = filepath.Join(os.Getenv("GOPATH"), "src", _pkg.RepoPath)

	mainApps := _pkg.GetMainApp()
	mainTests := _pkg.GetUnittest()
	testLibs := _pkg.GetTestLib()
	mainLibs := _pkg.GetMainLib()

	if (len(mainApps) == 0 && len(mainTests) == 0) && len(mainLibs) == 0 {
		return nil, fmt.Errorf("this package has no application(s), unittest(s) or main lib(s)")
	}

	wsc := NewWorkspaceConfig(_dev, _buildTarget.Os(), g.WorkspacePath, _pkg.RepoName)
	for _, app := range mainApps {
		if app.BuildType.IsApplication() {
			wsc.StartupProject = app.Name
			break
		}
	}
	if len(wsc.StartupProject) == 0 {
		for _, test := range mainTests {
			if test.BuildType.IsUnittest() {
				wsc.StartupProject = test.Name
				break
			}
		}
	}
	if len(wsc.StartupProject) == 0 {
		for _, lib := range mainLibs {
			if lib.BuildType.IsLibrary() {
				wsc.StartupProject = lib.Name
				break
			}
		}
	}
	wsc.MultiThreadedBuild = true

	ws := NewWorkspace(wsc)
	ws.WorkspaceName = _pkg.RepoName
	ws.WorkspaceAbsPath = filepath.Join(g.WorkspacePath, _pkg.RepoName)
	ws.GenerateAbsPath = filepath.Join(ws.WorkspaceAbsPath, "target", ws.Config.Dev.ToString())
	g.ExclusionFilter = denv.NewExclusionFilter(_buildTarget)

	projectStack := denv.NewDevProjectList()
	projectStack.AddMany(mainApps)
	projectStack.AddMany(mainTests)
	projectStack.AddMany(mainLibs)
	projectStack.AddMany(testLibs)

	projectDependencies := denv.NewDevProjectList()
	for projectStack.Len() > 0 {
		prj := projectStack.Pop()
		project := g.getOrCreateProject(prj, ws)
		projectDependencies.Reset()
		prj.CollectProjectDependencies(projectDependencies)
		for _, dp := range projectDependencies.Values {
			depProject := g.getOrCreateProject(dp, ws)
			if project.Dependencies.Add(depProject) {
				projectStack.Add(dp)
			}
		}
	}

	if err := ws.Resolve(ws.Config.Dev); err != nil {
		return nil, err
	}

	return ws, nil
}

func (g *Generator) getOrCreateProject(devProj *denv.DevProject, ws *Workspace) *Project {

	if p, ok := ws.ProjectList.Get(devProj.Name); ok {
		return p
	}

	projectSettings := NewProjectSettings()
	{
		if !devProj.BuildType.IsExecutable() {
			if devProj.BuildType.IsUnittest() {
				projectSettings.Group = "unittest/cpp-library"
			} else {
				projectSettings.Group = "main/cpp-library"
			}
			//projectSettings.Type = ProjectTypeCppLib
		} else {
			if devProj.BuildType.IsUnittest() {
				projectSettings.Group = "unittest/cpp-exe"
			} else if devProj.BuildType.IsCli() {
				projectSettings.Group = "cli/cpp-exe"
			} else {
				projectSettings.Group = "main/cpp-exe"
			}
			//projectSettings.Type = ProjectTypeCppExe
		}

		projectSettings.IsGuiApp = false
		projectSettings.PchHeader = ""
		projectSettings.MultiThreadedBuild = true
		projectSettings.CppAsObjCpp = false

		newProject := ws.NewProject2(devProj, projectSettings)
		newProject.ProjectFilename = devProj.Name

		exclusionFilter := func(_filepath string) bool { return g.ExclusionFilter.IsExcluded(_filepath) }

		// Glob all required files (*.cpp, *.c, *.h, *.hpp, *.csv)
		for _, srcPinnedPath := range devProj.SourceDirs {
			newProject.GlobFiles(filepath.Join(srcPinnedPath.Path.Root, srcPinnedPath.Path.Base), srcPinnedPath.Path.Sub, srcPinnedPath.Glob, exclusionFilter)
		}

		return newProject
	}
}
