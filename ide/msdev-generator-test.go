package ide

import (
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/axe"
)

type MsDevTestGenerator struct {
}

func NewMsDevTestGenerator() *MsDevTestGenerator {
	return &MsDevTestGenerator{}
}

func (m *MsDevTestGenerator) TestRun(ccoreAbsPath string, projectName string) error {

	visualStudioVersion := axe.VisualStudio2022

	wsc := axe.NewWorkspaceConfig(ccoreAbsPath, projectName)
	wsc.StartupProject = "cbase_unittest"
	wsc.MultiThreadedBuild = true

	ws := axe.NewWorkspace(wsc)
	ws.Generator = axe.GeneratorMsDev
	ws.WorkspaceName = projectName
	ws.WorkspaceAbsPath = ccoreAbsPath
	ws.GenerateAbsPath = filepath.Join(ccoreAbsPath, projectName, "target", ws.Generator.String())
	m.addWorkspaceConfiguration(ws, "DebugTest")
	m.addWorkspaceConfiguration(ws, "ReleaseTest")

	var cbase_lib *axe.Project
	var ccore_lib *axe.Project
	var cunittest_lib *axe.Project
	var cbase_unittest *axe.Project

	// cbase library project
	cbaseProjectConfig := axe.NewVisualStudioProjectConfig(visualStudioVersion)
	{
		cbaseProjectConfig.Group = "cpp-library"
		cbaseProjectConfig.Type = axe.ProjectTypeCppLib
		cbaseProjectConfig.IsGuiApp = false
		cbaseProjectConfig.PchHeader = ""
		cbaseProjectConfig.Dependencies = []string{"ccore_lib"}
		cbaseProjectConfig.MultiThreadedBuild = true
		cbaseProjectConfig.CppAsObjCpp = false

		cbase_lib = ws.NewProject("cbase_lib", "cbase", axe.ProjectTypeCppLib, cbaseProjectConfig)
		cbase_lib.ProjectFilename = "cbase_lib"
		cbase_lib.GlobFiles(filepath.Join(ccoreAbsPath, "ccore"), filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
		cbase_lib.GlobFiles(filepath.Join(ccoreAbsPath, "ccore"), filepath.Join("source", "main", "cpp", "^**", "*.m"))
		cbase_lib.GlobFiles(filepath.Join(ccoreAbsPath, "ccore"), filepath.Join("source", "main", "cpp", "^**", "*.mm"))
		cbase_lib.GlobFiles(filepath.Join(ccoreAbsPath, "ccore"), filepath.Join("source", "main", "include", "^**", "*.h"))
		cbase_lib.GlobFiles(filepath.Join(ccoreAbsPath, "ccore"), filepath.Join("source", "main", "include", "^**", "*.inl"))

		m.createDefaultProjectConfiguration(cbase_lib, "DebugTest")
		m.createDefaultProjectConfiguration(cbase_lib, "ReleaseTest")
	}

	// ccore library project
	ccoreProjectConfig := axe.NewVisualStudioProjectConfig(visualStudioVersion)
	{
		ccoreProjectConfig.Group = "cpp-library"
		ccoreProjectConfig.Type = axe.ProjectTypeCppLib
		ccoreProjectConfig.IsGuiApp = false
		ccoreProjectConfig.PchHeader = ""
		ccoreProjectConfig.Dependencies = []string{}
		ccoreProjectConfig.MultiThreadedBuild = true
		ccoreProjectConfig.CppAsObjCpp = false

		ccore_lib = ws.NewProject("ccore_lib", "ccore", axe.ProjectTypeCppLib, ccoreProjectConfig)
		ccore_lib.ProjectFilename = "ccore_lib"
		ccore_lib.GlobFiles(filepath.Join(ccoreAbsPath, "ccore"), filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
		ccore_lib.GlobFiles(filepath.Join(ccoreAbsPath, "ccore"), filepath.Join("source", "main", "cpp", "^**", "*.m"))
		ccore_lib.GlobFiles(filepath.Join(ccoreAbsPath, "ccore"), filepath.Join("source", "main", "cpp", "^**", "*.mm"))
		ccore_lib.GlobFiles(filepath.Join(ccoreAbsPath, "ccore"), filepath.Join("source", "main", "include", "^**", "*.h"))
		ccore_lib.GlobFiles(filepath.Join(ccoreAbsPath, "ccore"), filepath.Join("source", "main", "include", "^**", "*.inl"))

		m.createDefaultProjectConfiguration(ccore_lib, "DebugTest")
		m.createDefaultProjectConfiguration(ccore_lib, "ReleaseTest")
	}

	// cunittest library project
	cunittestProjectConfig := axe.NewVisualStudioProjectConfig(visualStudioVersion)
	{
		cunittestProjectConfig.Group = "unittest/cpp-library"
		cunittestProjectConfig.Type = axe.ProjectTypeCppLib
		cunittestProjectConfig.IsGuiApp = false
		cunittestProjectConfig.PchHeader = ""
		cunittestProjectConfig.Dependencies = []string{}
		cunittestProjectConfig.MultiThreadedBuild = true
		cunittestProjectConfig.CppAsObjCpp = false

		cunittest_lib = ws.NewProject("cunittest_lib", "cunittest", axe.ProjectTypeCppLib, cunittestProjectConfig)
		cunittest_lib.ProjectFilename = "cunittest_lib"
		cunittest_lib.GlobFiles(filepath.Join(ccoreAbsPath, "cunittest"), filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
		cunittest_lib.GlobFiles(filepath.Join(ccoreAbsPath, "cunittest"), filepath.Join("source", "main", "include", "^**", "*.h"))

		m.createDefaultProjectConfiguration(cunittest_lib, "DebugTest")
		m.createDefaultProjectConfiguration(cunittest_lib, "ReleaseTest")
	}

	// cbase unittest project, this is an executable
	cbaseTestProjectConfig := axe.NewVisualStudioProjectConfig(visualStudioVersion)
	{
		cbaseTestProjectConfig.Group = "unittest/cpp-exe"
		cbaseTestProjectConfig.Type = axe.ProjectTypeCppExe
		cbaseTestProjectConfig.IsGuiApp = false
		cbaseTestProjectConfig.PchHeader = ""
		cbaseTestProjectConfig.Dependencies = []string{"cbase_lib", "ccore_lib", "cunittest_lib"}
		cbaseTestProjectConfig.MultiThreadedBuild = true
		cbaseTestProjectConfig.CppAsObjCpp = false

		cbase_unittest = ws.NewProject("cbase_unittest", "cbase", axe.ProjectTypeCppExe, cbaseTestProjectConfig)
		cbase_unittest.ProjectFilename = "cbase_unittest"
		cbase_unittest.GlobFiles(filepath.Join(ccoreAbsPath, "cbase"), filepath.Join("source", "test", "cpp", "^**", "*.cpp"))
		cbase_unittest.GlobFiles(filepath.Join(ccoreAbsPath, "cbase"), filepath.Join("source", "test", "include", "^**", "*.h"))

		m.createDefaultProjectConfiguration(cbase_unittest, "DebugTest")
		m.createDefaultProjectConfiguration(cbase_unittest, "ReleaseTest")
	}

	if err := ws.Resolve(); err != nil {
		return err
	}

	g := axe.NewMsDevGenerator(ws)
	g.Generate()

	return nil
}

func (m *MsDevTestGenerator) createDefaultProjectConfiguration(p *axe.Project, configName string) *axe.Config {
	config := p.GetOrCreateConfig(configName)

	config.AddIncludeDir("source/main/include")

	if strings.HasSuffix(configName, "Test") {
		config.AddIncludeDir("source/test/include")
		config.VisualStudioClCompile.AddOrSet("ExceptionHandling", "Sync")
	}

	p.Configs.Add(config)
	return config
}

func (m *MsDevTestGenerator) addWorkspaceConfiguration(ws *axe.Workspace, configName string) {
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
