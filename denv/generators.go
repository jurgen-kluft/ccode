package denv

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	corepkg "github.com/jurgen-kluft/ccode/core"
	"github.com/jurgen-kluft/ccode/dev"
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

func NewExclusionFilter(target dev.BuildTarget) *ExclusionFilter {
	if target.Mac() {
		return &ExclusionFilter{Filter: IsExcludedOnMac}
	} else if target.Windows() {
		return &ExclusionFilter{Filter: IsExcludedOnWindows}
	} else if target.Linux() {
		return &ExclusionFilter{Filter: IsExcludedOnLinux}
	}
	return &ExclusionFilter{Filter: IsExcludedDefault}
}

type ExclusionFilter struct {
	Filter func(filepath string) bool
}

func (f *ExclusionFilter) IsExcluded(filepath string) bool {
	parts := corepkg.PathSplitRelativeFilePath(filepath, true)
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
	BuildTarget     dev.BuildTarget
	Verbose         bool
	WorkspacePath   string // $(GOPATH)/src/github.com/user, absolute path
	ExclusionFilter *ExclusionFilter
}

func NewGenerator(dev string, buildTarget dev.BuildTarget, verbose bool) *Generator {
	g := &Generator{}
	g.Dev = NewDevEnum(dev)
	g.BuildTarget = buildTarget
	g.Verbose = verbose
	return g
}

func (g *Generator) Generate(pkg *Package) {
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
		gg.Generate()
	case DevClay:
		gg := NewClayGenerator(ws, g.Verbose)
		err = gg.Generate()
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func (g *Generator) GenerateWorkspace(pkg *Package, _dev DevEnum, _buildTarget dev.BuildTarget) (*Workspace, error) {
	g.WorkspacePath = filepath.Join(os.Getenv("GOPATH"), "src", pkg.RepoPath)

	mainApps := pkg.GetMainApp()
	mainTests := pkg.GetUnittest()
	testLibs := pkg.GetTestLib()
	mainLibs := pkg.GetMainLib()

	if (len(mainApps) == 0 && len(mainTests) == 0) && len(mainLibs) == 0 {
		return nil, fmt.Errorf("this package has no application(s), unittest(s) or main lib(s)")
	}

	wsc := NewWorkspaceConfig(_dev, _buildTarget, g.WorkspacePath, pkg.RepoName)
	if len(mainApps) > 0 {
		wsc.StartupProject = mainApps[0].Name
	} else if len(mainTests) > 0 {
		wsc.StartupProject = mainTests[0].Name
	} else if len(mainLibs) > 0 {
		wsc.StartupProject = mainLibs[0].Name
	}
	wsc.MultiThreadedBuild = true

	ws := NewWorkspace(wsc)
	ws.WorkspaceName = pkg.RepoName
	ws.WorkspaceAbsPath = filepath.Join(g.WorkspacePath, pkg.RepoName)
	ws.GenerateAbsPath = filepath.Join(ws.WorkspaceAbsPath, "target", ws.Config.Dev.ToString())
	g.ExclusionFilter = NewExclusionFilter(ws.BuildTarget)

	projectStack := make([]*DevProject, 0)
	projectStack = append(projectStack, mainApps...)
	projectStack = append(projectStack, mainTests...)
	projectStack = append(projectStack, mainLibs...)
	projectStack = append(projectStack, testLibs...)
	projectHistory := make(map[*DevProject]string)

	for len(projectStack) > 0 {
		prj := projectStack[0]
		projectStack = projectStack[1:]
		project := g.getOrCreateProject(prj, ws)
		projectDependencies := prj.CollectProjectDependencies()
		for _, dp := range projectDependencies.Values {
			depProject := g.getOrCreateProject(dp, ws)
			project.Dependencies.Add(depProject)
			if _, ok := projectHistory[dp]; !ok {
				projectHistory[dp] = dp.Name
				projectStack = append(projectStack, dp)
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
