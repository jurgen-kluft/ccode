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
	RootAbsPath         string
	VisualStudioVersion EnumVisualStudio
}

func NewAxeGenerator(dev string, os string, arch string) *AxeGenerator {
	return &AxeGenerator{}
}

func (a *AxeGenerator) IsVisualStudio(dev string, os string, arch string) bool {
	return ParseVisualStudio(dev) != INVALID
}
func (a *AxeGenerator) IsTundra(dev string, os string, arch string) bool {
	dev = strings.ToLower(dev)
	return dev == "tundra"
}
func (a *AxeGenerator) IsCMake(dev string, os string, arch string) bool {
	dev = strings.ToLower(dev)
	return dev == "cmake"
}
func (a *AxeGenerator) IsXCode(dev string, os string, arch string) bool {
	dev = strings.ToLower(dev)
	return dev == "xcode"
}

func (a *AxeGenerator) GenerateMsDev(msdev DevEnum, pkg *denv.Package) error {
	var ws *Workspace
	var err error

	a.VisualStudioVersion = VisualStudio2022
	if ws, err = a.GenerateWorkspace(pkg, GeneratorMsDev); err != nil {
		return err
	}

	g := NewMsDevGenerator(ws)
	g.Generate()

	return nil
}

func (a *AxeGenerator) GenerateTundra(pkg *denv.Package) error {
	var ws *Workspace
	var err error

	if ws, err = a.GenerateWorkspace(pkg, GeneratorTundra); err != nil {
		return err
	}

	g := NewTundraGenerator(ws)
	g.Generate()

	return nil
}

func (a *AxeGenerator) GenerateCMake(pkg *denv.Package) error {
	var ws *Workspace
	var err error

	if ws, err = a.GenerateWorkspace(pkg, GeneratorCMake); err != nil {
		return err
	}

	g := NewCMakeGenerator(ws)
	g.Generate()

	return nil
}

func (a *AxeGenerator) GenerateWorkspace(pkg *denv.Package, generatorType GeneratorType) (*Workspace, error) {
	a.RootAbsPath = filepath.Join(os.Getenv("GOPATH"), "src")

	mainApp := pkg.GetMainApp()
	unittestApp := pkg.GetUnittest()

	if mainApp == nil && unittestApp == nil {
		return nil, fmt.Errorf("this package has no main app or unittest")
	}

	app := unittestApp
	if app == nil {
		app = mainApp
	}

	wsc := NewWorkspaceConfig(a.RootAbsPath, app.Name)
	wsc.StartupProject = app.Name
	wsc.MultiThreadedBuild = true

	ws := NewWorkspace(wsc)
	ws.Generator = generatorType
	ws.WorkspaceName = app.Name
	ws.WorkspaceAbsPath = a.RootAbsPath
	ws.GenerateAbsPath = filepath.Join(a.RootAbsPath, app.PackageURL, "target", ws.Generator.String())
	if unittestApp != nil {
		a.addWorkspaceConfiguration(ws, ConfigTypeDebug|ConfigTypeTest)
		a.addWorkspaceConfiguration(ws, ConfigTypeRelease|ConfigTypeTest)
	} else {
		a.addWorkspaceConfiguration(ws, ConfigTypeDebug)
		a.addWorkspaceConfiguration(ws, ConfigTypeRelease)
	}

	projectMap := map[string]int{}
	projectList := []*denv.Project{}
	projectIdx := 0
	for _, dp := range app.Dependencies {
		if _, ok := projectMap[dp.Name]; !ok {
			projectMap[dp.Name] = len(projectList)
			projectList = append(projectList, dp)
		}
	}

	// Traverse and collect all dependencies
	for {
		cp := projectList[projectIdx]
		projectIdx++

		for _, dp := range cp.Dependencies {
			if _, ok := projectMap[dp.Name]; !ok {
				projectMap[dp.Name] = len(projectList)
				projectList = append(projectList, dp)
			}
		}
		if projectIdx == len(projectList) {
			break
		}
	}

	// Create the main project
	if app == mainApp {
		a.createProject(mainApp, ws, false)
		for _, dp := range projectList {
			a.createProject(dp, ws, false)
		}
	} else {
		a.createProject(unittestApp, ws, true)
		for _, dp := range projectList {
			a.createProject(dp, ws, true)
		}
	}

	if err := ws.Resolve(); err != nil {
		return nil, err
	}

	return ws, nil
}

func (a *AxeGenerator) createProject(proj *denv.Project, ws *Workspace, unittest bool) {
	projectConfig := NewVisualStudioProjectConfig(a.VisualStudioVersion)
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

		projAbsPath := filepath.Join(a.RootAbsPath, proj.PackageURL)
		lib := ws.NewProject(proj.Name, projAbsPath, ProjectTypeCppLib, projectConfig)
		lib.ProjectFilename = proj.Name

		if unittest && executable {
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.h"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.inl"))

			lib.GlobFiles(projAbsPath, filepath.Join("source", "test", "cpp", "^**", "*.cpp"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "test", "include", "^**", "*.h"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "test", "include", "^**", "*.inl"))

			a.createDefaultProjectConfiguration(lib, ConfigTypeDebug|ConfigTypeTest, true)
			a.createDefaultProjectConfiguration(lib, ConfigTypeRelease|ConfigTypeTest, true)
		} else if unittest && !executable {
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.h"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.inl"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
			if ws.MakeTarget.OSIsMac() {
				lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.m"))
				lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.mm"))
			}

			a.createDefaultProjectConfiguration(lib, ConfigTypeDebug|ConfigTypeTest, true)
			a.createDefaultProjectConfiguration(lib, ConfigTypeRelease|ConfigTypeTest, true)
		} else if !unittest && executable {
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.h"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.inl"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
			if ws.MakeTarget.OSIsMac() {
				lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.m"))
				lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.mm"))
			}

			a.createDefaultProjectConfiguration(lib, ConfigTypeDebug, false)
			a.createDefaultProjectConfiguration(lib, ConfigTypeDebug, false)
		} else {
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.inl"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
			if ws.MakeTarget.OSIsMac() {
				lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.m"))
				lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.mm"))
			}
			a.createDefaultProjectConfiguration(lib, ConfigTypeDebug, false)
			a.createDefaultProjectConfiguration(lib, ConfigTypeDebug, false)
		}
	}
}

func (a *AxeGenerator) createDefaultProjectConfiguration(p *Project, configType ConfigType, unittest bool) *Config {
	config := p.GetOrCreateConfig(configType)

	config.AddIncludeDir("source/main/include")

	if unittest {
		config.AddIncludeDir("source/test/include")
		config.VisualStudioClCompile.AddOrSet("ExceptionHandling", "Sync")
	}

	p.Configs.Add(config)
	return config
}

func (a *AxeGenerator) addWorkspaceConfiguration(ws *Workspace, configType ConfigType) {
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
