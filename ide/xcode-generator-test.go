package ide

import (
	"path/filepath"

	"github.com/jurgen-kluft/ccode/axe"
)

type XcodeTestGenerator struct {
}

func NewXcodeTestGenerator() *XcodeTestGenerator {
	return &XcodeTestGenerator{}
}

func (x *XcodeTestGenerator) TestRun(ccoreAbsPath string, projectName string) error {

	visualStudioVersion := axe.VisualStudio2022

	wsc := axe.NewWorkspaceConfig(ccoreAbsPath, projectName)
	wsc.StartupProject = "cbase_unittest"
	wsc.MultiThreadedBuild = true

	ws := axe.NewWorkspace(wsc)
	ws.Generator = axe.GeneratorMsDev
	ws.WorkspaceName = projectName
	ws.WorkspaceAbsPath = ccoreAbsPath
	ws.GenerateAbsPath = filepath.Join(ccoreAbsPath, projectName, "target", ws.Generator.String())

	x.addWorkspaceConfiguration(ws, axe.ConfigTypeDebug|axe.ConfigTypeTest)
	x.addWorkspaceConfiguration(ws, axe.ConfigTypeRelease|axe.ConfigTypeTest)

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

		config := x.createDefaultProjectConfiguration(cbase_lib, axe.ConfigTypeDebug)
		config.CppDefines.ValuesToAdd("")
		x.createDefaultProjectConfiguration(cbase_lib, axe.ConfigTypeRelease)
		x.createDefaultProjectConfiguration(cbase_lib, axe.ConfigTypeDebug|axe.ConfigTypeTest)
		x.createDefaultProjectConfiguration(cbase_lib, axe.ConfigTypeRelease|axe.ConfigTypeTest)
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

		x.createDefaultProjectConfiguration(ccore_lib, axe.ConfigTypeDebug)
		x.createDefaultProjectConfiguration(ccore_lib, axe.ConfigTypeRelease)
		x.createDefaultProjectConfiguration(ccore_lib, axe.ConfigTypeDebug|axe.ConfigTypeTest)
		x.createDefaultProjectConfiguration(ccore_lib, axe.ConfigTypeRelease|axe.ConfigTypeTest)
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
		cunittest_lib.ProjectFilename = "cunittest"
		cunittest_lib.GlobFiles(filepath.Join(ccoreAbsPath, "cunittest"), "source/main/cpp/^**/*.cpp")
		cunittest_lib.GlobFiles(filepath.Join(ccoreAbsPath, "cunittest"), "source/main/include/^**/*.h")

		x.createDefaultProjectConfiguration(cunittest_lib, axe.ConfigTypeDebug|axe.ConfigTypeTest)
		x.createDefaultProjectConfiguration(cunittest_lib, axe.ConfigTypeRelease|axe.ConfigTypeTest)
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

		x.createDefaultProjectConfiguration(cbase_unittest, axe.ConfigTypeDebug|axe.ConfigTypeTest)
		x.createDefaultProjectConfiguration(cbase_unittest, axe.ConfigTypeRelease|axe.ConfigTypeTest)
	}

	if err := ws.Resolve(); err != nil {
		return err
	}

	g := axe.NewXcodeGenerator(ws)
	g.Generate()

	return nil
}

func (x *XcodeTestGenerator) createDefaultProjectConfiguration(p *axe.Project, configType axe.ConfigType) *axe.Config {
	config := p.GetOrCreateConfig(configType)

	config.AddIncludeDir("source/main/include")
	if configType.IsTest() {
		config.AddIncludeDir("source/test/include")
	}

	p.Configs.Add(config)
	return config
}

func (x *XcodeTestGenerator) addWorkspaceConfiguration(ws *axe.Workspace, configType axe.ConfigType) {
	config := axe.NewConfig(configType, ws, nil)

	if configType.IsDebug() {
		config.CppDefines.ValuesToAdd("DEBUG", "_DEBUG")
	} else {
		config.CppDefines.ValuesToAdd("NDEBUG")
	}
	config.CppDefines.ValuesToAdd("_UNICODE", "UNICODE")
	config.LinkFlags.ValuesToAdd("-ObjC", "-framework Foundation")
	config.XcodeSettings.AddOrSet("MACOSX_DEPLOYMENT_TARGET", "10.11")

	// clang
	config.CppFlags.ValuesToAdd("-std=c++11", "-Wall", "-Wfatal-errors", "-Werror", "-Wno-switch")
	config.LinkFlags.ValuesToAdd("-lstdc++")
	if configType.IsDebug() {
		config.CppFlags.ValuesToAdd("-g")
	}

	ws.AddConfig(config)
}
