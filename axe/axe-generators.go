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
	if strings.HasPrefix(str, "win_") || strings.HasPrefix(str, "pc_") || strings.HasPrefix(str, "win32_") || strings.HasPrefix(str, "win64_") {
		return true
	}
	if strings.HasSuffix(str, "_win") || strings.HasSuffix(str, "_pc") || strings.HasSuffix(str, "_win32") || strings.HasSuffix(str, "_win64") {
		return true
	}
	if strings.HasPrefix(str, "linux_") || strings.HasPrefix(str, "unix_") {
		return true
	}
	if strings.HasSuffix(str, "_linux") || strings.HasSuffix(str, "_unix") {
		return true
	}
	if strings.EqualFold(str, "windows") || strings.EqualFold(str, "linux") {
		return true
	}
	if strings.EqualFold(str, "d3d11") || strings.EqualFold(str, "d3d12") {
		return true
	}
	if strings.HasSuffix(str, "_nob") {
		return true
	}
	return false
}

func IsExcludedOnWindows(str string) bool {
	if strings.HasPrefix(str, "mac_") || strings.HasPrefix(str, "macos_") || strings.HasPrefix(str, "darwin_") || strings.HasPrefix(str, "linux_") || strings.HasPrefix(str, "unix_") {
		return true
	}
	if strings.HasSuffix(str, "_mac") || strings.HasSuffix(str, "_macos") || strings.HasSuffix(str, "_darwin") || strings.HasSuffix(str, "_linux") || strings.HasSuffix(str, "_unix") {
		return true
	}
	if strings.EqualFold(str, "macos") || strings.EqualFold(str, "linux") {
		return true
	}
	if strings.EqualFold(str, "cocoa") || strings.EqualFold(str, "metal") {
		return true
	}
	if strings.HasSuffix(str, "_nob") {
		return true
	}
	return false
}

func IsExcludedOnLinux(str string) bool {
	if strings.HasPrefix(str, "mac_") || strings.HasPrefix(str, "macos_") || strings.HasPrefix(str, "darwin_") {
		return true
	}
	if strings.HasPrefix(str, "win_") || strings.HasPrefix(str, "pc_") || strings.HasPrefix(str, "win32_") || strings.HasPrefix(str, "win64_") {
		return true
	}
	if strings.EqualFold(str, "windows") || strings.EqualFold(str, "macos") {
		return true
	}
	if strings.EqualFold(str, "d3d11") || strings.EqualFold(str, "d3d12") || strings.EqualFold(str, "cocoa") || strings.EqualFold(str, "metal") {
		return true
	}
	if strings.HasSuffix(str, "_nob") {
		return true
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
type AxeGenerator struct {
	Dev             DevEnum
	Os              string
	Arch            string
	GoPathAbs       string // $(GOPATH)/src, absolute path
	ExclusionFilter *ExclusionFilter
}

func NewAxeGenerator(dev string, os string, arch string) *AxeGenerator {
	g := &AxeGenerator{}
	g.Dev = GetDevEnum(strings.ToLower(dev))
	g.Os = strings.ToLower(os)
	g.Arch = strings.ToLower(arch)
	return g
}

func (g *AxeGenerator) IsValid() bool {
	return g.Dev != DevInvalid
}
func (g *AxeGenerator) IsVisualStudio() bool {
	return g.Dev&DevVisualStudio == DevVisualStudio
}
func (g *AxeGenerator) IsTundra() bool {
	return g.Dev == DevTundra
}
func (g *AxeGenerator) IsMake() bool {
	return g.Dev == DevMake
}
func (g *AxeGenerator) IsCMake() bool {
	return g.Dev == DevCmake
}
func (g *AxeGenerator) IsXCode() bool {
	return g.Dev == DevXcode
}

func (g *AxeGenerator) Generate(pkg *denv.Package) error {
	var ws *Workspace
	var err error

	if ws, err = g.GenerateWorkspace(pkg, g.Dev); err != nil {
		return err
	}

	switch g.Dev {
	case DevTundra:
		gg := NewTundraGenerator(ws)
		gg.Generate()
	case DevCmake:
		gg := NewCMakeGenerator(ws)
		gg.Generate()
	case DevMake:
		gg := NewMakeGenerator2(ws)
		gg.Generate()
	case DevXcode:
		gg := NewXcodeGenerator(ws)
		gg.Generate()
	case DevVs2015, DevVs2017, DevVs2019, DevVs2022:
		gg := NewMsDevGenerator(ws)
		gg.Generate()
	}

	return nil
}

func (g *AxeGenerator) GenerateWorkspace(pkg *denv.Package, dev DevEnum) (*Workspace, error) {
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

	wsc := NewWorkspaceConfig(dev, g.GoPathAbs, app.Name)
	wsc.StartupProject = app.Name
	wsc.MultiThreadedBuild = true

	ws := NewWorkspace(wsc)
	ws.WorkspaceName = app.Name
	ws.WorkspaceAbsPath = g.GoPathAbs
	ws.GenerateAbsPath = filepath.Join(g.GoPathAbs, app.PackageURL, "target", ws.Config.Dev.String())

	g.ExclusionFilter = NewExclusionFilter(ws.MakeTarget)

	if mainLib != nil {
		for _, cfg := range mainLib.Configs {
			g.addWorkspaceConfiguration(ws, cfg, true)
		}
	}

	if mainApp != nil {
		for _, cfg := range mainApp.Configs {
			g.addWorkspaceConfiguration(ws, cfg, true)
		}
	}

	if mainTest != nil {
		for _, cfg := range mainTest.Configs {
			g.addWorkspaceConfiguration(ws, cfg, true)
		}
	}

	// Create the main and dependency projects and also setup the list of dependencies of each project

	// Note: Can we not create all 3: App, Lib and Unittest projects ?
	if mainApp != nil {
		mainAppDependencies := g.collectProjectDependencies(mainApp)
		mainAppProject := g.createProject(mainApp, ws, false)
		for _, dp := range mainAppDependencies {
			mainAppProject.Settings.Dependencies = append(mainAppProject.Settings.Dependencies, dp.Name)
		}
		for _, dp := range mainAppDependencies {
			depProjectDependencies := g.collectProjectDependencies(dp)
			depProject := g.createProject(dp, ws, false)
			for _, ddp := range depProjectDependencies {
				depProject.Settings.Dependencies = append(depProject.Settings.Dependencies, ddp.Name)
			}
		}
	}

	if mainLib != nil {
		mainLibDependencies := g.collectProjectDependencies(mainLib)
		mainLibProject := g.createProject(mainLib, ws, false)
		for _, dp := range mainLibDependencies {
			mainLibProject.Settings.Dependencies = append(mainLibProject.Settings.Dependencies, dp.Name)
		}
		for _, dp := range mainLibDependencies {
			depProjectDependencies := g.collectProjectDependencies(dp)
			depProject := g.createProject(dp, ws, false)
			for _, ddp := range depProjectDependencies {
				depProject.Settings.Dependencies = append(depProject.Settings.Dependencies, ddp.Name)
			}
		}
	}

	if mainTest != nil {
		mainTestDependencies := g.collectProjectDependencies(mainTest)
		mainTestProject := g.createProject(mainTest, ws, true)
		for _, dp := range mainTestDependencies {
			mainTestProject.Settings.Dependencies = append(mainTestProject.Settings.Dependencies, dp.Name)
		}
		for _, dp := range mainTestDependencies {
			depProjectDependencies := g.collectProjectDependencies(dp)
			depProject := g.createProject(dp, ws, true)
			for _, ddp := range depProjectDependencies {
				depProject.Settings.Dependencies = append(depProject.Settings.Dependencies, ddp.Name)
			}
		}
	}

	// if app == mainApp {
	// 	mainAppDependencies := g.collectProjectDependencies(app)
	// 	mainAppProject := g.createProject(app, ws, false)
	// 	for _, dp := range mainAppDependencies {
	// 		mainAppProject.Settings.Dependencies = append(mainAppProject.Settings.Dependencies, dp.Name)
	// 	}
	// 	for _, dp := range mainAppDependencies {
	// 		depProjectDependencies := g.collectProjectDependencies(dp)
	// 		depProject := g.createProject(dp, ws, false)
	// 		for _, ddp := range depProjectDependencies {
	// 			depProject.Settings.Dependencies = append(depProject.Settings.Dependencies, ddp.Name)
	// 		}
	// 	}
	// } else if app == mainTest {
	// 	mainTestDependencies := g.collectProjectDependencies(app)
	// 	mainTestProject := g.createProject(app, ws, true)
	// 	for _, dp := range mainTestDependencies {
	// 		mainTestProject.Settings.Dependencies = append(mainTestProject.Settings.Dependencies, dp.Name)
	// 	}
	// 	for _, dp := range mainTestDependencies {
	// 		depProjectDependencies := g.collectProjectDependencies(dp)
	// 		depProject := g.createProject(dp, ws, true)
	// 		for _, ddp := range depProjectDependencies {
	// 			depProject.Settings.Dependencies = append(depProject.Settings.Dependencies, ddp.Name)
	// 		}
	// 	}
	// }

	if err := ws.Resolve(); err != nil {
		return nil, err
	}

	return ws, nil
}

func (g *AxeGenerator) collectProjectDependencies(proj *denv.Project) []*denv.Project {

	// Traverse and collect all dependencies

	projectMap := map[string]int{}
	projectList := []*denv.Project{}
	for _, dp := range proj.Dependencies {
		if _, ok := projectMap[dp.Name]; !ok {
			projectMap[dp.Name] = len(projectList)
			projectList = append(projectList, dp)
		}
	}

	projectIdx := 0
	for projectIdx < len(projectList) {
		cp := projectList[projectIdx]
		projectIdx++

		for _, dp := range cp.Dependencies {
			if _, ok := projectMap[dp.Name]; !ok {
				projectMap[dp.Name] = len(projectList)
				projectList = append(projectList, dp)
			}
		}
	}
	return projectList
}

func (g *AxeGenerator) createProject(devProj *denv.Project, ws *Workspace, unittest bool) *Project {
	projectConfig := NewVisualStudioProjectConfig(g.Dev)
	{
		executable := devProj.Type == denv.Executable
		if !executable {
			if unittest {
				projectConfig.Group = "unittest/cpp-library"
			} else {
				projectConfig.Group = "mainapp/cpp-library"
			}
			projectConfig.Type = ProjectTypeCppLib
		} else {
			if unittest {
				projectConfig.Group = "unittest/cpp-exe"
			} else {
				projectConfig.Group = "mainapp/cpp-exe"
			}
			projectConfig.Type = ProjectTypeCppExe
		}
		projectConfig.IsGuiApp = false
		projectConfig.PchHeader = ""
		projectConfig.Dependencies = []string{}
		projectConfig.MultiThreadedBuild = true
		projectConfig.CppAsObjCpp = false

		for _, dp := range devProj.Dependencies {
			projectConfig.Dependencies = append(projectConfig.Dependencies, dp.Name)
		}

		projAbsPath := filepath.Join(g.GoPathAbs, devProj.PackageURL)
		newProject := ws.NewProject(devProj.Name, projAbsPath, ProjectTypeCppLib, projectConfig)
		newProject.ProjectFilename = devProj.Name

		exclusionFilter := func(_filepath string) bool { return g.ExclusionFilter.IsExcluded(_filepath) }

		if executable && unittest {
			// Unittest executable
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "test", "cpp", "^**", "*.cpp"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "test", "include", "^**", "*.h"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "test", "include", "^**", "*.hpp"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "test", "include", "^**", "*.inl"), exclusionFilter)
		} else if executable {
			// Application
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "cpp", "^**", "*.cpp"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "include", "^**", "*.h"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "include", "^**", "*.hpp"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "include", "^**", "*.inl"), exclusionFilter)
			if ws.MakeTarget.OSIsMac() {
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "cpp", "^**", "*.m"), exclusionFilter)
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "cpp", "^**", "*.mm"), exclusionFilter)
			}
		} else {
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.cpp"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.h"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.hpp"), exclusionFilter)
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.inl"), exclusionFilter)
			if ws.MakeTarget.OSIsMac() {
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.m"), exclusionFilter)
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.mm"), exclusionFilter)
			}
		}

		for _, cfg := range devProj.Configs {
			g.createProjectConfiguration(newProject, devProj, cfg, executable, unittest)
		}
		return newProject
	}
}

func (g *AxeGenerator) createProjectConfiguration(p *Project, prj *denv.Project, cfg *denv.Config, executable bool, unittest bool) *Config {
	configType := ConfigTypeNone
	if cfg.Config == "Debug" {
		configType = ConfigTypeDebug
	} else if cfg.Config == "Release" {
		configType = ConfigTypeRelease
	}
	if unittest {
		configType |= ConfigTypeTest
	}

	config := p.GetOrCreateConfig(configType)

	for _, define := range cfg.Defines {
		config.CppDefines.ValuesToAdd(define)
	}

	for _, include := range prj.IncludeDirs {
		config.AddIncludeDir(include)
	}

	// config.AddIncludeDir("source/main/include")

	// if !unittest && executable {
	// 	config.AddIncludeDir("source/app/include")
	// }

	// if unittest && executable {
	// 	config.AddIncludeDir("source/test/include")
	// }

	if configType.IsTest() {
		config.VisualStudioClCompile.AddOrSet("ExceptionHandling", "Sync")
	}

	p.Configs.Add(config)
	return config
}

func (g *AxeGenerator) addWorkspaceConfiguration(ws *Workspace, cfg *denv.Config, unittest bool) {
	configType := ConfigTypeNone
	if cfg.Config == "Debug" {
		configType = ConfigTypeDebug
	} else if cfg.Config == "Release" {
		configType = ConfigTypeRelease
	}
	if unittest {
		configType |= ConfigTypeTest
	}

	config := NewConfig(configType, ws, nil)

	if configType.IsDebug() {
		config.CppDefines.ValuesToAdd("TARGET_DEBUG", "TARGET_DEV", "_DEBUG")
	} else {
		config.CppDefines.ValuesToAdd("TARGET_RELEASE", "TARGET_DEV", "NDEBUG")
	}

	if ws.MakeTarget.OSIsWindows() {
		config.CppDefines.ValuesToAdd("TARGET_PC")
	} else if ws.MakeTarget.OSIsLinux() {
		config.CppDefines.ValuesToAdd("TARGET_LINUX")
	} else if ws.MakeTarget.OSIsMac() {
		config.CppDefines.ValuesToAdd("TARGET_MAC")
	}

	if ws.MakeTarget.OSIsMac() {
		config.LinkFlags.ValuesToAdd("-ObjC")
		config.LinkFlags.ValuesToAdd("-framework Foundation")
		config.LinkFlags.ValuesToAdd("-framework Cocoa")
		config.LinkFlags.ValuesToAdd("-framework Carbon")
		config.LinkFlags.ValuesToAdd("-framework Metal")
		config.LinkFlags.ValuesToAdd("-framework OpenGL")
		config.LinkFlags.ValuesToAdd("-framework IOKit")
		config.LinkFlags.ValuesToAdd("-framework AppKit")
		config.LinkFlags.ValuesToAdd("-framework CoreVideo")
		config.LinkFlags.ValuesToAdd("-framework QuartzCore")
		//		config.LinkFlags.ValuesToAdd("-framework AudioToolBox")
		//		config.LinkFlags.ValuesToAdd("-framework OpenAL")
	}

	config.CppDefines.ValuesToAdd("_UNICODE", "UNICODE")

	// clang
	if ws.MakeTarget.CompilerIsClang() {
		config.CppFlags.ValuesToAdd("-std=c++17")
		config.CppFlags.ValuesToAdd("-Wno-switch", "-Wno-unused-variable", "-Wno-unused-function", "-Wno-unused-private-field")
		config.CppFlags.ValuesToAdd("-Wno-unused-but-set-variable")
		//config.CppFlags.ValuesToAdd("-Wfatal-errors", "-Werror")
		config.LinkFlags.ValuesToAdd("-lstdc++")
		if configType.IsDebug() {
			config.CppFlags.ValuesToAdd("-g")
		}
	}

	ws.AddConfig(config)
}
