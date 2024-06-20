package axe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/denv"
)

type DevEnum uint

// All development environment
const (
	TUNDRA       DevEnum = 0x020000
	CMAKE        DevEnum = 0x040000
	XCODE        DevEnum = 0x080000
	VISUALSTUDIO DevEnum = 0x100000
	VS2015       DevEnum = VISUALSTUDIO | 2015
	VS2017       DevEnum = VISUALSTUDIO | 2017
	VS2019       DevEnum = VISUALSTUDIO | 2019
	VS2022       DevEnum = VISUALSTUDIO | 2022
	INVALID      DevEnum = 0xFFFFFFFF
)

func GetDevEnum(dev string) DevEnum {
	dev = strings.ToLower(dev)
	if dev == "tundra" {
		return TUNDRA
	} else if dev == "cmake" {
		return CMAKE
	} else if dev == "xcode" {
		return XCODE
	}
	return ParseVisualStudio(dev)
}

// ParseVisualStudio returns a value for type IDE deduced from the incoming string @dev
func ParseVisualStudio(dev string) DevEnum {
	dev = strings.ToLower(dev)
	if dev == "vs2022" {
		return VS2022
	} else if dev == "vs2019" {
		return VS2019
	} else if dev == "vs2017" {
		return VS2017
	} else if dev == "vs2015" {
		return VS2015
	}
	return INVALID
}

// ----------------------------------------------------------------------------------------------
// IDE generator
// ----------------------------------------------------------------------------------------------

type AxeGenerator struct {
	Dev         DevEnum
	Os          string
	Arch        string
	RootAbsPath string
}

func NewAxeGenerator(dev string, os string, arch string) *AxeGenerator {
	g := &AxeGenerator{}
	g.Dev = GetDevEnum(strings.ToLower(dev))
	g.Os = strings.ToLower(os)
	g.Arch = strings.ToLower(arch)
	return g
}

func (g *AxeGenerator) IsVisualStudio() bool {
	return g.Dev&VISUALSTUDIO != 0
}
func (g *AxeGenerator) IsTundra() bool {
	return g.Dev == TUNDRA
}
func (g *AxeGenerator) IsCMake() bool {
	return g.Dev == CMAKE
}
func (g *AxeGenerator) IsXCode() bool {
	return g.Dev == XCODE
}

func (g *AxeGenerator) GenerateMsDev(pkg *denv.Package) error {
	var ws *Workspace
	var err error

	if ws, err = g.GenerateWorkspace(pkg, GeneratorMsDev); err != nil {
		return err
	}

	gg := NewMsDevGenerator(ws)
	gg.Generate()

	return nil
}

func (g *AxeGenerator) GenerateTundra(pkg *denv.Package) error {
	var ws *Workspace
	var err error

	if ws, err = g.GenerateWorkspace(pkg, GeneratorTundra); err != nil {
		return err
	}

	gg := NewTundraGenerator(ws)
	gg.Generate()

	return nil
}

func (g *AxeGenerator) GenerateCMake(pkg *denv.Package) error {
	var ws *Workspace
	var err error

	if ws, err = g.GenerateWorkspace(pkg, GeneratorCMake); err != nil {
		return err
	}

	gg := NewCMakeGenerator(ws)
	gg.Generate()

	return nil
}

func (g *AxeGenerator) GenerateXcode(pkg *denv.Package) error {
	var ws *Workspace
	var err error

	if ws, err = g.GenerateWorkspace(pkg, GeneratorXcode); err != nil {
		return err
	}

	gg := NewXcodeGenerator(ws)
	gg.Generate()

	return nil
}

func (g *AxeGenerator) GenerateWorkspace(pkg *denv.Package, generatorType GeneratorType) (*Workspace, error) {
	g.RootAbsPath = filepath.Join(os.Getenv("GOPATH"), "src")

	mainApp := pkg.GetMainApp()
	unittestApp := pkg.GetUnittest()

	if mainApp == nil && unittestApp == nil {
		return nil, fmt.Errorf("this package has no main app or unittest")
	}

	app := unittestApp
	if app == nil {
		app = mainApp
	}

	wsc := NewWorkspaceConfig(g.RootAbsPath, app.Name)
	wsc.StartupProject = app.Name
	wsc.MultiThreadedBuild = true

	ws := NewWorkspace(wsc)
	ws.Generator = generatorType
	ws.WorkspaceName = app.Name
	ws.WorkspaceAbsPath = g.RootAbsPath
	ws.GenerateAbsPath = filepath.Join(g.RootAbsPath, app.PackageURL, "target", ws.Generator.String())
	if unittestApp != nil {
		g.addWorkspaceConfiguration(ws, ConfigTypeDebug|ConfigTypeTest)
		g.addWorkspaceConfiguration(ws, ConfigTypeRelease|ConfigTypeTest)
	} else {
		g.addWorkspaceConfiguration(ws, ConfigTypeDebug)
		g.addWorkspaceConfiguration(ws, ConfigTypeRelease)
	}

	// Create the main and dependency projects and also setup the list of dependencies of each project
	if app == mainApp {
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
	} else if app == unittestApp {
		unittestDependencies := g.collectProjectDependencies(unittestApp)
		unittestProject := g.createProject(unittestApp, ws, true)
		for _, dp := range unittestDependencies {
			unittestProject.Settings.Dependencies = append(unittestProject.Settings.Dependencies, dp.Name)
		}
		for _, dp := range unittestDependencies {
			depProjectDependencies := g.collectProjectDependencies(dp)
			depProject := g.createProject(dp, ws, true)
			for _, ddp := range depProjectDependencies {
				depProject.Settings.Dependencies = append(depProject.Settings.Dependencies, ddp.Name)
			}
		}
	}

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

func (g *AxeGenerator) createProject(proj *denv.Project, ws *Workspace, unittest bool) *Project {
	projectConfig := NewVisualStudioProjectConfig(g.Dev)
	{
		executable := proj.Type == denv.Executable
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

		for _, dp := range proj.Dependencies {
			projectConfig.Dependencies = append(projectConfig.Dependencies, dp.Name)
		}

		projAbsPath := filepath.Join(g.RootAbsPath, proj.PackageURL)
		newProject := ws.NewProject(proj.Name, projAbsPath, ProjectTypeCppLib, projectConfig)
		newProject.ProjectFilename = proj.Name

		if unittest && executable {
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.h"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.inl"))

			newProject.GlobFiles(projAbsPath, filepath.Join("source", "test", "cpp", "^**", "*.cpp"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "test", "include", "^**", "*.h"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "test", "include", "^**", "*.inl"))

			g.createDefaultProjectConfiguration(newProject, ConfigTypeDebug|ConfigTypeTest, true)
			g.createDefaultProjectConfiguration(newProject, ConfigTypeRelease|ConfigTypeTest, true)
		} else if unittest && !executable {
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.h"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.inl"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
			if ws.MakeTarget.OSIsMac() {
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.m"))
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.mm"))
			}

			g.createDefaultProjectConfiguration(newProject, ConfigTypeDebug|ConfigTypeTest, true)
			g.createDefaultProjectConfiguration(newProject, ConfigTypeRelease|ConfigTypeTest, true)
		} else if !unittest && executable {
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.h"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.inl"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
			if ws.MakeTarget.OSIsMac() {
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.m"))
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.mm"))
			}

			g.createDefaultProjectConfiguration(newProject, ConfigTypeDebug, false)
			g.createDefaultProjectConfiguration(newProject, ConfigTypeDebug, false)
		} else {
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.inl"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
			if ws.MakeTarget.OSIsMac() {
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.m"))
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.mm"))
			}
			g.createDefaultProjectConfiguration(newProject, ConfigTypeDebug, false)
			g.createDefaultProjectConfiguration(newProject, ConfigTypeDebug, false)
		}

		return newProject
	}
}

func (g *AxeGenerator) createDefaultProjectConfiguration(p *Project, configType ConfigType, unittest bool) *Config {
	config := p.GetOrCreateConfig(configType)

	config.AddIncludeDir("source/main/include")

	if unittest {
		config.AddIncludeDir("source/test/include")
		config.VisualStudioClCompile.AddOrSet("ExceptionHandling", "Sync")
	}

	p.Configs.Add(config)
	return config
}

func (g *AxeGenerator) addWorkspaceConfiguration(ws *Workspace, configType ConfigType) {
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

	config.CppDefines.ValuesToAdd("_UNICODE", "UNICODE")

	// clang
	if ws.MakeTarget.CompilerIsClang() {
		config.CppFlags.ValuesToAdd("-std=c++11", "-Wall", "-Wfatal-errors", "-Werror", "-Wno-switch")
		config.LinkFlags.ValuesToAdd("-lstdc++")
		if configType.IsDebug() {
			config.CppFlags.ValuesToAdd("-g")
		}
	}

	ws.AddConfig(config)
}
