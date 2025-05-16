package axe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/denv"
)

// ----------------------------------------------------------------------------------------------
// Exclusion filter
// ----------------------------------------------------------------------------------------------
func IsExcludedOnMac(str string) bool {
	exclude := []string{"win", "pc", "win32", "win64", "windows", "d3d11", "d3d12", "linux", "unix", "nob", "ios"}
	for _, e := range exclude {
		if strings.HasPrefix(str, e+"_") || strings.HasSuffix(str, "_"+e) || strings.EqualFold(str, e) {
			return true
		}
	}
	return false
}

func IsExcludedOnWindows(str string) bool {
	exclude := []string{"mac", "macos", "darwin", "linux", "unix", "nob", "cocoa", "metal", "osx", "ios"}
	for _, e := range exclude {
		if strings.HasPrefix(str, e+"_") || strings.HasSuffix(str, "_"+e) || strings.EqualFold(str, e) {
			return true
		}
	}
	return false
}

func IsExcludedOnLinux(str string) bool {
	exclude := []string{"mac", "macos", "darwin", "win", "pc", "win32", "win64", "windows", "d3d11", "d3d12", "cocoa", "metal", "osx", "ios", "nob"}
	for _, e := range exclude {
		if strings.HasPrefix(str, e+"_") || strings.HasSuffix(str, "_"+e) || strings.EqualFold(str, e) {
			return true
		}
	}
	return false
}

func IsExcludedDefault(str string) bool {
	if strings.HasSuffix(str, "_nob") {
		return true
	}
	return false
}

func NewExclusionFilter(target MakeTarget) *ExclusionFilter {
	if target.OSIsMac() {
		return &ExclusionFilter{Filter: IsExcludedOnMac}
	} else if target.OSIsWindows() {
		return &ExclusionFilter{Filter: IsExcludedOnWindows}
	} else if target.OSIsLinux() {
		return &ExclusionFilter{Filter: IsExcludedOnLinux}
	}
	return &ExclusionFilter{Filter: IsExcludedDefault}
}

type ExclusionFilter struct {
	Filter func(filepath string) bool
}

func (f *ExclusionFilter) IsExcluded(filepath string) bool {
	parts := PathSplitRelativeFilePath(filepath, true)
	for i := 0; i < len(parts)-1; i++ {
		p := strings.ToLower(parts[i])
		if f.Filter(p) {
			return true
		}
	}
	return false
}

// ----------------------------------------------------------------------------------------------
// IDE generator
// ----------------------------------------------------------------------------------------------
type Generator struct {
	Dev             DevEnum
	Os              string
	Arch            string
	Verbose         bool
	GoPathAbs       string // $(GOPATH)/src, absolute path
	ExclusionFilter *ExclusionFilter
}

func NewGenerator(dev string, os string, arch string, verbose bool) *Generator {
	g := &Generator{}
	g.Dev = DevEnumFromString(dev)
	g.Os = strings.ToLower(os)
	g.Arch = strings.ToLower(arch)
	g.Verbose = verbose
	return g
}

func (g *Generator) Generate(pkg *denv.Package) error {
	var ws *Workspace
	var err error

	if ws, err = g.GenerateWorkspace(pkg, g.Dev, g.Os, g.Arch); err != nil {
		return err
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
		gg.Generate()
	case DevEspMake:
		gg := NewEspMakeGenerator(ws, g.Verbose)
		err = gg.Generate()
	}

	return err
}

func (g *Generator) GenerateWorkspace(pkg *denv.Package, _dev DevEnum, _os string, _arch string) (*Workspace, error) {
	g.GoPathAbs = filepath.Join(os.Getenv("GOPATH"), "src")

	mainLib := pkg.GetMainLib()
	mainApp := pkg.GetMainApp()
	mainTest := pkg.GetUnittest()

	if mainApp == nil && mainTest == nil && mainLib == nil {
		return nil, fmt.Errorf("this package has no main app, main lib or unittest")
	}

	app := mainTest
	if app == nil {
		app = mainApp
	}
	if app == nil {
		app = mainLib
	}

	wsc := NewWorkspaceConfig(_dev, _os, _arch, g.GoPathAbs, app.Name)
	wsc.StartupProject = app.Name
	wsc.MultiThreadedBuild = true

	ws := NewWorkspace(wsc)
	ws.WorkspaceName = app.Name
	ws.WorkspaceAbsPath = g.GoPathAbs
	ws.GenerateAbsPath = filepath.Join(g.GoPathAbs, app.PackageURL, "target", ws.Config.Dev.String())

	g.ExclusionFilter = NewExclusionFilter(ws.MakeTarget)

	// Create the main and dependency projects and also set up the list of dependencies of each project

	if mainTest != nil {
		mainTestProject := g.getOrCreateProject(mainTest, ws)
		mainTestProject.AddConfigurations(mainTest.Configs)

		mainTestDependencies := mainTest.CollectProjectDependencies()
		for _, dp := range mainTestDependencies {
			dpProject := g.getOrCreateProject(dp, ws)
			mainTestProject.Dependencies.Add(dpProject)

			dpProject.AddConfigurations(dp.Configs)

			dpDependencies := dp.CollectProjectDependencies()
			for _, dpd := range dpDependencies {
				dpdProject := g.getOrCreateProject(dpd, ws)
				dpProject.Dependencies.Add(dpdProject)
			}
		}
	}

	if mainApp != nil {
		mainAppProject := g.getOrCreateProject(mainApp, ws)
		mainAppProject.AddConfigurations(mainApp.Configs)

		mainAppDependencies := mainApp.CollectProjectDependencies()
		for _, dp := range mainAppDependencies {
			depProject := g.getOrCreateProject(dp, ws)
			mainAppProject.Dependencies.Add(depProject)

			depProject.AddConfigurations(dp.Configs)

			dpDependencies := dp.CollectProjectDependencies()
			for _, dpd := range dpDependencies {
				dpdProject := g.getOrCreateProject(dpd, ws)
				depProject.Dependencies.Add(dpdProject)
			}
		}
	}

	if mainLib != nil {
		mainLibProject := g.getOrCreateProject(mainLib, ws)
		mainLibProject.AddConfigurations(mainLib.Configs)

		mainLibDependencies := mainLib.CollectProjectDependencies()
		for _, dp := range mainLibDependencies {
			depProject := g.getOrCreateProject(dp, ws)
			mainLibProject.Dependencies.Add(depProject)

			depProject.AddConfigurations(dp.Configs)

			dpDependencies := dp.CollectProjectDependencies()
			for _, dpd := range dpDependencies {
				dpdProject := g.getOrCreateProject(dpd, ws)
				depProject.Dependencies.Add(dpdProject)
			}
		}
	}

	if err := ws.Resolve(ws.Config.Dev); err != nil {
		return nil, err
	}

	return ws, nil
}

func (g *Generator) getOrCreateProject(devProj *denv.Project, ws *Workspace) *Project {

	if p, ok := ws.ProjectList.Get(devProj.Name); ok {
		return p
	}

	projectConfig := NewProjectConfig()
	{
		if !devProj.Type.IsExecutable() {
			if devProj.Type.IsUnitTest() {
				projectConfig.Group = "unittest/cpp-library"
			} else {
				projectConfig.Group = "main/cpp-library"
			}
			//projectConfig.Type = ProjectTypeCppLib
		} else {
			if devProj.Type.IsUnitTest() {
				projectConfig.Group = "unittest/cpp-exe"
			} else {
				projectConfig.Group = "main/cpp-exe"
			}
			//projectConfig.Type = ProjectTypeCppExe
		}

		projectConfig.IsGuiApp = false
		projectConfig.PchHeader = ""
		projectConfig.MultiThreadedBuild = true
		projectConfig.CppAsObjCpp = false

		projAbsPath := filepath.Join(g.GoPathAbs, devProj.PackageURL)
		newProject := ws.NewProject(devProj.Name, projAbsPath, devProj.Type, projectConfig)
		newProject.ProjectFilename = devProj.Name

		exclusionFilter := func(_filepath string) bool { return g.ExclusionFilter.IsExcluded(_filepath) }

		if devProj.Type.IsUnitTest() {
			// Unittest executable
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "test", "cpp", "^**", "*.cpp"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "test", "include", "^**", "*.h"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "test", "include", "^**", "*.hpp"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "test", "include", "^**", "*.inl"), exclusionFilter)
		} else if devProj.Type.IsApplication() {
			// Application
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "cpp", "^**", "*.cpp"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "include", "^**", "*.h"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "include", "^**", "*.hpp"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "include", "^**", "*.inl"), exclusionFilter)
			if ws.MakeTarget.OSIsMac() {
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "cpp", "^**", "*.m"), exclusionFilter)
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "cpp", "^**", "*.mm"), exclusionFilter)
			}
		} else if devProj.Type.IsLibrary() {
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.c"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.cpp"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.h"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.hpp"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.inl"), exclusionFilter)
			if ws.MakeTarget.OSIsMac() {
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.m"), exclusionFilter)
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.mm"), exclusionFilter)
			}
		}

		return newProject
	}
}
