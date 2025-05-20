package denv

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cutils "github.com/jurgen-kluft/ccode/cutils"
)

// ----------------------------------------------------------------------------------------------
// Exclusion filter
// ----------------------------------------------------------------------------------------------
var gValidSuffixDefault = 0
var gValidSuffixWindows = 1
var gValidSuffixMac = 2
var gValidSuffixIOs = 3
var gValidSuffixLinux = 4
var gValidSuffixArduino = 5

var gValidSuffixDB = [][]string{
	[]string{"_nob", "_null", "_nill"},
	[]string{"_win", "_pc", "_win32", "_win64", "_windows", "_d3d11", "_d3d12"},
	[]string{"_mac", "_macos", "_darwin", "_cocoa", "_metal", "_osx"},
	[]string{"_ios", "_iphone", "_ipad", "_ipod"},
	[]string{"_linux", "_unix"},
	[]string{"arduino", "esp32"},
}

func IsExcludedOn(str string, os int) bool {
	for i, l := range gValidSuffixDB {
		if i == os {
			continue
		}
		for _, e := range l {
			if strings.HasSuffix(str, e) {
				return true
			}
		}
	}
	return false
}

func IsExcludedOnMac(str string) bool {
	return IsExcludedOn(str, gValidSuffixMac)
}
func IsExcludedOnWindows(str string) bool {
	return IsExcludedOn(str, gValidSuffixWindows)
}
func IsExcludedOnLinux(str string) bool {
	return IsExcludedOn(str, gValidSuffixLinux)
}
func IsExcludedDefault(str string) bool {
	return IsExcludedOn(str, gValidSuffixDefault)
}

func NewExclusionFilter(target *MakeTarget) *ExclusionFilter {
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
	parts := cutils.PathSplitRelativeFilePath(filepath, true)
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

func (g *Generator) Generate(pkg *Package) error {
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
	case DevEsp32:
		gg := NewEsp32Generator(ws, g.Verbose)
		err = gg.Generate()
	}

	return err
}

func (g *Generator) GenerateWorkspace(pkg *Package, _dev DevEnum, _os string, _arch string) (*Workspace, error) {
	g.GoPathAbs = filepath.Join(os.Getenv("GOPATH"), "src")

	mainApps := pkg.GetMainApp()
	mainTests := pkg.GetUnittest()
	mainLibs := pkg.GetMainLib()

	if (len(mainApps) == 0 && len(mainTests) == 0) && len(mainLibs) == 0 {
		return nil, fmt.Errorf("this package has no application(s), unittest(s) or main lib(s)")
	}

	wsc := NewWorkspaceConfig(_dev, _os, _arch, g.GoPathAbs, pkg.Name)
	if len(mainApps) > 0 {
		wsc.StartupProject = mainApps[0].Name
	} else if len(mainTests) > 0 {
		wsc.StartupProject = mainTests[0].Name
	} else if len(mainLibs) > 0 {
		wsc.StartupProject = mainLibs[0].Name
	}
	wsc.MultiThreadedBuild = true

	ws := NewWorkspace(wsc)
	ws.WorkspaceName = pkg.Name
	ws.WorkspaceAbsPath = g.GoPathAbs
	if len(mainApps) > 0 {
		ws.GenerateAbsPath = filepath.Join(g.GoPathAbs, mainApps[0].PackageURL, "target", ws.Config.Dev.String())
	} else if len(mainTests) > 0 {
		ws.GenerateAbsPath = filepath.Join(g.GoPathAbs, mainTests[0].PackageURL, "target", ws.Config.Dev.String())
	} else if len(mainLibs) > 0 {
		ws.GenerateAbsPath = filepath.Join(g.GoPathAbs, mainLibs[0].PackageURL, "target", ws.Config.Dev.String())
	}
	g.ExclusionFilter = NewExclusionFilter(ws.BuildTarget)

	for _, mainTest := range mainTests {
		if mainTest.Type.IsTest() {
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
	}

	for _, mainApp := range mainApps {
		if mainApp.Type.IsApplication() {
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
	}

	for _, mainLib := range mainLibs {
		if mainLib.Type.IsLibrary() {
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
	}

	if err := ws.Resolve(ws.Config.Dev); err != nil {
		return nil, err
	}

	return ws, nil
}

func (g *Generator) getOrCreateProject(devProj *DevProject, ws *Workspace) *Project {

	if p, ok := ws.ProjectList.Get(devProj.Name); ok {
		return p
	}

	projectConfig := NewProjectConfig()
	{
		if !devProj.Type.IsExecutable() {
			if devProj.Type.IsTest() {
				projectConfig.Group = "unittest/cpp-library"
			} else {
				projectConfig.Group = "main/cpp-library"
			}
			//projectConfig.Type = ProjectTypeCppLib
		} else {
			if devProj.Type.IsTest() {
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
		newProject := ws.NewProject2(devProj, devProj.Name, projAbsPath, projectConfig)
		newProject.ProjectFilename = devProj.Name

		// Generate file entry dictionaries for groups of external source files to the new project
		for _, externalSource := range devProj.ExternalSources {
			if externalSource.BuildTargets.Contains(ws.BuildTarget.BuildTarget) {
				externalSrcFileDict := NewFileEntryDict(externalSource.Path)
				for _, srcFile := range externalSource.SrcFiles {
					externalSrcFileDict.Add(srcFile)
				}
				newProject.ExternalSrcFiles = append(newProject.ExternalSrcFiles, externalSrcFileDict)
			}
		}

		exclusionFilter := func(_filepath string) bool { return g.ExclusionFilter.IsExcluded(_filepath) }

		if devProj.Type.IsTest() {
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
			if ws.BuildTarget.OSIsMac() {
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "cpp", "^**", "*.m"), exclusionFilter)
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "cpp", "^**", "*.mm"), exclusionFilter)
			}
		} else if devProj.Type.IsLibrary() {
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.c"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.cpp"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.h"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.hpp"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.inl"), exclusionFilter)
			if ws.BuildTarget.OSIsMac() {
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.m"), exclusionFilter)
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.mm"), exclusionFilter)
			}
		}

		return newProject
	}
}
