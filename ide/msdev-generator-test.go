package ide

import (
	"path/filepath"
	"strings"

	"github.com/jurgen-kluft/ccode/axe"
)

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

func TestRun(workspacePath string, libraryPath string) error {

	ws := axe.NewWorkspace()

	ws.Config.ConfigList = []string{"Debug", "Release", "Debug Test", "Release Test"}
	ws.GenerateAbsPath = "target"
	ws.Config.StartupProject = "cbase_unittest"
	ws.Config.MultiThreadedBuild = true

	// Add the configurations
	for _, name := range ws.Config.ConfigList {
		addWorkspaceConfiguration(ws, name)
	}

	visualStudioVersion := axe.VisualStudio2022

	ws.Generator = "msdev"
	ws.MakeTarget = axe.NewDefaultMakeTarget()
	ws.WorkspaceName = "cbase"
	ws.WorkspaceAbsPath = workspacePath
	ws.GenerateAbsPath = filepath.Join(workspacePath, "cbase", "target")

	debugConfig := axe.NewConfig("Debug", ws, nil)
	releaseConfig := axe.NewConfig("Release", ws, nil)
	debugTestConfig := axe.NewConfig("Debug Test", ws, nil)
	releaseTestConfig := axe.NewConfig("Release Test", ws, nil)

	ws.Configs["Debug"] = debugConfig
	ws.Configs["Release"] = releaseConfig
	ws.Configs["Debug Test"] = debugTestConfig
	ws.Configs["Release Test"] = releaseTestConfig

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

		cbase_lib = axe.NewProject(ws, "cbase_lib", "cbase", axe.ProjectTypeCppLib, cbaseProjectConfig)
		cbase_lib.GlobFiles(filepath.Join(workspacePath, libraryPath), "source/main/cpp/^**/*.cpp")
		cbase_lib.GlobFiles(filepath.Join(workspacePath, libraryPath), "source/main/cpp/^**/*.m")
		cbase_lib.GlobFiles(filepath.Join(workspacePath, libraryPath), "source/main/cpp/^**/*.mm")
		cbase_lib.GlobFiles(filepath.Join(workspacePath, libraryPath), "source/main/include/^**/*.h")
		cbase_lib.GlobFiles(filepath.Join(workspacePath, libraryPath), "source/main/include/^**/*.inl")

		createDefaultProjectConfiguration(cbase_lib, "Debug")
		createDefaultProjectConfiguration(cbase_lib, "Release")
		createDefaultProjectConfiguration(cbase_lib, "Debug Test")
		createDefaultProjectConfiguration(cbase_lib, "Release Test")
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

		ccore_lib = axe.NewProject(ws, "ccore_lib", "ccore", axe.ProjectTypeCppLib, ccoreProjectConfig)
		ccore_lib.GlobFiles(filepath.Join(workspacePath, libraryPath), "source/main/cpp/^**/*.cpp")
		ccore_lib.GlobFiles(filepath.Join(workspacePath, libraryPath), "source/main/cpp/^**/*.m")
		ccore_lib.GlobFiles(filepath.Join(workspacePath, libraryPath), "source/main/cpp/^**/*.mm")
		ccore_lib.GlobFiles(filepath.Join(workspacePath, libraryPath), "source/main/include/^**/*.h")
		ccore_lib.GlobFiles(filepath.Join(workspacePath, libraryPath), "source/main/include/^**/*.inl")

		createDefaultProjectConfiguration(ccore_lib, "Debug")
		createDefaultProjectConfiguration(ccore_lib, "Release")
		createDefaultProjectConfiguration(ccore_lib, "Debug Test")
		createDefaultProjectConfiguration(ccore_lib, "Release Test")
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

		cunittest_lib = axe.NewProject(ws, "cunittest_lib", "cunittest", axe.ProjectTypeCppLib, cunittestProjectConfig)
		cunittest_lib.GlobFiles(filepath.Join(workspacePath, libraryPath), "source/main/cpp/^**/*.cpp")
		cunittest_lib.GlobFiles(filepath.Join(workspacePath, libraryPath), "source/main/include/^**/*.h")

		createDefaultProjectConfiguration(cunittest_lib, "Debug Test")
		createDefaultProjectConfiguration(cunittest_lib, "Release Test")
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

		cbase_unittest = axe.NewProject(ws, "cbase_unittest", "cbase", axe.ProjectTypeCppExe, cbaseTestProjectConfig)
		cbase_unittest.GlobFiles(filepath.Join(workspacePath, libraryPath), "source/test/cpp/^**/*.cpp")
		cbase_unittest.GlobFiles(filepath.Join(workspacePath, libraryPath), "source/test/include/^**/*.h")

		createDefaultProjectConfiguration(cbase_unittest, "Debug Test")
		createDefaultProjectConfiguration(cbase_unittest, "Release Test")
	}

	ws.Projects["cbase_lib"] = cbase_lib
	ws.Projects["ccore_lib"] = ccore_lib
	ws.Projects["cunittest_lib"] = cunittest_lib
	ws.Projects["cbase_unittest"] = cbase_unittest

	if err := ws.Finalize(); err != nil {
		return err
	}

	g := NewMsDevGenerator(ws)
	g.Generate()

	return nil
}

func createDefaultProjectConfiguration(p *axe.Project, configName string) {
	config := p.GetOrCreateConfig(configName)
	config.WarningAsError = true

	config.AddIncludeDir("source/main/include")
	if strings.HasSuffix(configName, "Test") {
		config.AddIncludeDir("source/test/include")
	}

	p.Configs[configName] = config
}

func addWorkspaceConfiguration(ws *axe.Workspace, configName string) {
	config := axe.NewConfig(configName, ws, nil)

	config.WarningAsError = true
	if config.IsDebug {
		config.CppDefines.ValuesToAdd("DEBUG", "_DEBUG")
	} else {
		config.CppDefines.ValuesToAdd("NDEBUG")
	}
	config.CppDefines.ValuesToAdd("_UNICODE", "UNICODE")
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
