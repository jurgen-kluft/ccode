package axe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/denv"
)

// ----------------------------------------------------------------------------------------------
// IDE generator
// ----------------------------------------------------------------------------------------------

type AxeGenerator struct {
	Dev       DevEnum
	Os        string
	Arch      string
	GoPathAbs string // $(GOPATH)/src, absolute path
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

	mainApp := pkg.GetMainApp()
	unittestApp := pkg.GetUnittest()

	if mainApp == nil && unittestApp == nil {
		return nil, fmt.Errorf("this package has no main app or unittest")
	}

	app := unittestApp
	if app == nil {
		app = mainApp
	}

	wsc := NewWorkspaceConfig(dev, g.GoPathAbs, app.Name)
	wsc.StartupProject = app.Name
	wsc.MultiThreadedBuild = true

	ws := NewWorkspace(wsc)
	ws.WorkspaceName = app.Name
	ws.WorkspaceAbsPath = g.GoPathAbs
	ws.GenerateAbsPath = filepath.Join(g.GoPathAbs, app.PackageURL, "target", ws.Config.Dev.String())

	if unittestApp != nil {
		for _, cfg := range unittestApp.Configs {
			g.addWorkspaceConfiguration(ws, cfg, true)
		}
	} else {
		for _, cfg := range mainApp.Configs {
			g.addWorkspaceConfiguration(ws, cfg, true)
		}
	}

	// Create the main and dependency projects and also setup the list of dependencies of each project
	if app == mainApp {
		mainAppDependencies := g.collectProjectDependencies(app)
		mainAppProject := g.createProject(app, ws, false)
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
		unittestDependencies := g.collectProjectDependencies(app)
		unittestProject := g.createProject(app, ws, true)
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

		if executable && unittest {
			// Unittest executable
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "test", "cpp", "^**", "*.cpp"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "test", "include", "^**", "*.h"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "test", "include", "^**", "*.hpp"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "test", "include", "^**", "*.inl"))
		} else if executable {
			// Application
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "cpp", "^**", "*.cpp"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "include", "^**", "*.h"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "include", "^**", "*.hpp"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "include", "^**", "*.inl"))
			if ws.MakeTarget.OSIsMac() {
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "cpp", "^**", "*.m"))
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "app", "cpp", "^**", "*.mm"))
			}
		} else {
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.h"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.hpp"))
			newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "include", "^**", "*.inl"))
			if ws.MakeTarget.OSIsMac() {
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.m"))
				newProject.GlobFiles(projAbsPath, filepath.Join("source", "main", "cpp", "^**", "*.mm"))
			}
		}

		for _, cfg := range devProj.Configs {
			g.createProjectConfiguration(newProject, cfg, executable, unittest)
		}
		return newProject
	}
}

func (g *AxeGenerator) createProjectConfiguration(p *Project, cfg *denv.Config, executable bool, unittest bool) *Config {
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
	for _, include := range cfg.IncludeDirs {
		config.AddIncludeDir(include)
	}

	config.AddIncludeDir("source/main/include")

	if !unittest && executable {
		config.AddIncludeDir("source/app/include")
	}

	if unittest && executable {
		config.AddIncludeDir("source/test/include")
	}

	if p.Name == "cunittest" {
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
		//		config.CppFlags.ValuesToAdd("-std=c++14", "-Wall")
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
