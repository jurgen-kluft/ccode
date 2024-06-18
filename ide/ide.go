package ide

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jurgen-kluft/ccode/axe"
	"github.com/jurgen-kluft/ccode/denv"
)

func GenerateMsDevIde() {
	workspacePath := "$HOME/dev.go/src/github.com/jurgen-kluft"
	if runtime.GOOS == "windows" {
		workspacePath = "d:\\Dev.Go\\src\\github.com\\jurgen-kluft"
	}
	workspacePath = os.ExpandEnv(workspacePath)
	generator := NewMsDevTestGenerator()
	generator.TestRun(workspacePath, "cbase")
}

func GenerateXcodeIde() {
	workspacePath := "$HOME/dev.go/src/github.com/jurgen-kluft"
	workspacePath = os.ExpandEnv(workspacePath)
	generator := NewXcodeTestGenerator()
	generator.TestRun(workspacePath, "cbase")
}

// ----------------------------------------------------------------------------------------------
// IDE generator
// ----------------------------------------------------------------------------------------------

type AxeGenerator struct {
	RootAbsPath         string
	VisualStudioVersion axe.EnumVisualStudio
}

func NewAxeGenerator() *AxeGenerator {
	return &AxeGenerator{}
}

func (a *AxeGenerator) GenerateMsDev(msdev denv.DevEnum, pkg *denv.Package) error {
	a.VisualStudioVersion = axe.VisualStudio2022
	a.RootAbsPath = filepath.Join(os.Getenv("GOPATH"), "src")

	mainApp := pkg.GetMainApp()
	unittestApp := pkg.GetUnittest()

	if mainApp == nil && unittestApp == nil {
		return fmt.Errorf("this package has no main app or unittest")
	}

	app := unittestApp
	if app == nil {
		app = mainApp
	}

	wsc := axe.NewWorkspaceConfig(a.RootAbsPath, app.Name)
	wsc.StartupProject = app.Name
	wsc.MultiThreadedBuild = true

	ws := axe.NewWorkspace(wsc)
	ws.Generator = axe.GeneratorMsDev
	ws.WorkspaceName = app.Name
	ws.WorkspaceAbsPath = a.RootAbsPath
	ws.GenerateAbsPath = filepath.Join(a.RootAbsPath, app.PackageURL, "target", ws.Generator.String())
	if unittestApp != nil {
		a.addWorkspaceConfiguration(ws, "DebugTest")
		a.addWorkspaceConfiguration(ws, "ReleaseTest")
	} else {
		a.addWorkspaceConfiguration(ws, "Debug")
		a.addWorkspaceConfiguration(ws, "Release")
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
		return err
	}

	g := axe.NewMsDevGenerator(ws)
	g.Generate()

	return nil
}

func (a *AxeGenerator) createProject(proj *denv.Project, ws *axe.Workspace, unittest bool) {
	projectConfig := axe.NewVisualStudioProjectConfig(a.VisualStudioVersion)
	{
		executable := proj.Type == denv.Executable
		if !executable {
			if unittest {
				projectConfig.Group = "unittest/cpp-library"
			} else {
				projectConfig.Group = "mainapp/cpp-library"
			}
			projectConfig.Type = axe.ProjectTypeCppLib
		} else {
			if unittest {
				projectConfig.Group = "unittest/cpp-exe"
			} else {
				projectConfig.Group = "mainapp/cpp-exe"
			}
			projectConfig.Type = axe.ProjectTypeCppExe
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
		lib := ws.NewProject(proj.Name, projAbsPath, axe.ProjectTypeCppLib, projectConfig)
		lib.ProjectFilename = proj.Name

		if unittest && executable {
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.h"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.inl"))

			lib.GlobFiles(projAbsPath, filepath.Join("source", "test", "cpp", "^**", "*.cpp"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "test", "include", "^**", "*.h"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "test", "include", "^**", "*.inl"))

			a.createDefaultProjectConfiguration(lib, "DebugTest", true)
			a.createDefaultProjectConfiguration(lib, "ReleaseTest", true)
		} else if unittest && !executable {
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.h"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.inl"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
			if ws.MakeTarget.OSIsMac() {
				lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.m"))
				lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.mm"))
			}

			a.createDefaultProjectConfiguration(lib, "DebugTest", true)
			a.createDefaultProjectConfiguration(lib, "ReleaseTest", true)
		} else if !unittest && executable {
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.h"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.inl"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
			if ws.MakeTarget.OSIsMac() {
				lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.m"))
				lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.mm"))
			}

			a.createDefaultProjectConfiguration(lib, "Debug", false)
			a.createDefaultProjectConfiguration(lib, "Release", false)
		} else {
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.inl"))
			lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
			if ws.MakeTarget.OSIsMac() {
				lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.m"))
				lib.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.mm"))
			}
			a.createDefaultProjectConfiguration(lib, "Debug", false)
			a.createDefaultProjectConfiguration(lib, "Release", false)
		}
	}
}

func (a *AxeGenerator) createDefaultProjectConfiguration(p *axe.Project, configName string, unittest bool) *axe.Config {
	config := p.GetOrCreateConfig(configName)

	config.AddIncludeDir("source/main/include")

	if unittest {
		config.AddIncludeDir("source/test/include")
		config.VisualStudioClCompile.AddOrSet("ExceptionHandling", "Sync")
	}

	p.Configs.Add(config)
	return config
}

func (a *AxeGenerator) addWorkspaceConfiguration(ws *axe.Workspace, configName string) {
	config := axe.NewConfig(configName, ws, nil)

	if config.IsDebug {
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
		if config.IsDebug {
			config.CppFlags.ValuesToAdd("-g")
		}
	}

	ws.AddConfig(config)
}
