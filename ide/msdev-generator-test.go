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

// TestRun("~/dev.go/src/github.com/jurgen-kluft", "cbase"")
//
//	ccore = "~/dev.go/src/github.com/jurgen-kluft/ccore"
//	cbase = "~/dev.go/src/github.com/jurgen-kluft/cbase"
//	cunittest = "~/dev.go/src/github.com/jurgen-kluft/cunittest"

// C++ source files for the cbase library are in "source/main/cpp" and should be globbed
// C++ header files for the cbase library are in "source/main/include" and should be globbed
// C++ source files for the cbase unittest project are in "source/test/cpp" and should be globbed

// The cbase C++ library dependencies are:
// - ccore
// - cunittest

// The ccore C++ library has no dependencies
// The cunittest C++ library has no dependencies

// For all the libraries there are 4 build configurations:
// - debug
// - release
// - debug_test
// - release_test

// For the unittest project there are 2 build configurations:
// - debug_test
// - release_test

// The cbase library is a static library
// The ccore library is a static library
// The cunittest library is a static library
// The cbase unittest project is an executable

// The above are all the details needed to generate the Visual Studio solution and project files

func (m *MsDevTestGenerator) TestRun(ccoreAbsPath string, projectName string) error {

	ws := axe.NewWorkspace(ccoreAbsPath, projectName)

	ws.Config.ConfigList = []string{"DebugTest", "ReleaseTest"}
	ws.GenerateAbsPath = "target"
	ws.Config.StartupProject = "cbase_unittest"
	ws.Config.MultiThreadedBuild = true

	// Add the configurations
	for _, name := range ws.Config.ConfigList {
		m.addWorkspaceConfiguration(ws, name)
	}

	visualStudioVersion := axe.VisualStudio2022

	ws.Generator = axe.GeneratorMsDev
	ws.MakeTarget = axe.NewDefaultMakeTarget()
	ws.WorkspaceName = projectName
	ws.WorkspaceAbsPath = ccoreAbsPath
	ws.GenerateAbsPath = filepath.Join(ccoreAbsPath, projectName, "target", ws.Generator.String())

	debugTestConfig := axe.NewConfig("DebugTest", ws, nil)
	releaseTestConfig := axe.NewConfig("ReleaseTest", ws, nil)

	ws.AddConfig(debugTestConfig)
	ws.AddConfig(releaseTestConfig)

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
		cbase_lib.GlobFiles(filepath.Join(ccoreAbsPath, "cbase"), "source/main/cpp/^**/*.cpp")
		cbase_lib.GlobFiles(filepath.Join(ccoreAbsPath, "cbase"), "source/main/cpp/^**/*.m")
		cbase_lib.GlobFiles(filepath.Join(ccoreAbsPath, "cbase"), "source/main/cpp/^**/*.mm")
		cbase_lib.GlobFiles(filepath.Join(ccoreAbsPath, "cbase"), "source/main/include/^**/*.h")
		cbase_lib.GlobFiles(filepath.Join(ccoreAbsPath, "cbase"), "source/main/include/^**/*.inl")

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
		ccore_lib.GlobFiles(filepath.Join(ccoreAbsPath, "ccore"), "source/main/cpp/^**/*.cpp")
		ccore_lib.GlobFiles(filepath.Join(ccoreAbsPath, "ccore"), "source/main/cpp/^**/*.m")
		ccore_lib.GlobFiles(filepath.Join(ccoreAbsPath, "ccore"), "source/main/cpp/^**/*.mm")
		ccore_lib.GlobFiles(filepath.Join(ccoreAbsPath, "ccore"), "source/main/include/^**/*.h")
		ccore_lib.GlobFiles(filepath.Join(ccoreAbsPath, "ccore"), "source/main/include/^**/*.inl")

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
		cunittest_lib.GlobFiles(filepath.Join(ccoreAbsPath, "cunittest"), "source/main/cpp/^**/*.cpp")
		cunittest_lib.GlobFiles(filepath.Join(ccoreAbsPath, "cunittest"), "source/main/include/^**/*.h")

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
		cbase_unittest.GlobFiles(filepath.Join(ccoreAbsPath, "cbase"), "source/test/cpp/^**/*.cpp")
		cbase_unittest.GlobFiles(filepath.Join(ccoreAbsPath, "cbase"), "source/test/include/^**/*.h")

		m.createDefaultProjectConfiguration(cbase_unittest, "DebugTest")
		m.createDefaultProjectConfiguration(cbase_unittest, "ReleaseTest")
	}

	if err := ws.Finalize(); err != nil {
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
		config.VisualStudioClCompile.Add("ExceptionHandling", "Sync")
	}

	config.CppDefines.ValuesToAdd("TARGET_PC")

	p.Configs[configName] = config
	return config
}

func (m *MsDevTestGenerator) addWorkspaceConfiguration(ws *axe.Workspace, configName string) {
	config := axe.NewConfig(configName, ws, nil)

	if config.IsDebug {
		config.CppDefines.ValuesToAdd("DEBUG", "_DEBUG")
	} else {
		config.CppDefines.ValuesToAdd("NDEBUG")
	}
	config.CppDefines.ValuesToAdd("_UNICODE", "UNICODE", "TARGET_PC")
	config.LinkFlags.ValuesToAdd("-ObjC", "-framework Foundation")
	config.XcodeSettings.Add("MACOSX_DEPLOYMENT_TARGET", "10.11")

	// clang
	config.CppFlags.ValuesToAdd("-std=c++11", "-Wall", "-Wfatal-errors", "-Werror", "-Wno-switch")
	config.LinkFlags.ValuesToAdd("-lstdc++")
	if config.IsDebug {
		config.CppFlags.ValuesToAdd("-g")
	}

	ws.AddConfig(config)
}
