package main

import (
	"path/filepath"

	"github.com/jurgen-kluft/ccode/axe"
)

type XcodeTestGenerator struct {
}

func NewXcodeTestGenerator() *XcodeTestGenerator {
	return &XcodeTestGenerator{}
}

func (m *XcodeTestGenerator) TestRun(rootAbsPath string, projectName string) error {

	devVersion := axe.XCODE

	wsc := axe.NewWorkspaceConfig(devVersion, rootAbsPath, projectName)
	wsc.StartupProject = "cbase_unittest"
	wsc.MultiThreadedBuild = true

	ws := axe.NewWorkspace(wsc)
	ws.WorkspaceName = projectName
	ws.WorkspaceAbsPath = rootAbsPath
	ws.GenerateAbsPath = filepath.Join(rootAbsPath, projectName, "target", ws.Config.Dev.String())
	m.addWorkspaceConfiguration(ws, axe.ConfigTypeDebug|axe.ConfigTypeTest)
	m.addWorkspaceConfiguration(ws, axe.ConfigTypeRelease|axe.ConfigTypeTest)

	var cbase_lib *axe.Project
	var ccore_lib *axe.Project
	var cunittest_lib *axe.Project
	var cbase_unittest *axe.Project

	// cbase library project
	cbaseProjectConfig := axe.NewVisualStudioProjectConfig(devVersion)
	{
		cbaseProjectConfig.Group = "cpp-library"
		cbaseProjectConfig.Type = axe.ProjectTypeCppLib
		cbaseProjectConfig.IsGuiApp = false
		cbaseProjectConfig.PchHeader = ""
		cbaseProjectConfig.Dependencies = []string{"ccore_lib"}
		cbaseProjectConfig.MultiThreadedBuild = true
		cbaseProjectConfig.CppAsObjCpp = false

		cbaseAbsPath := filepath.Join(rootAbsPath, "cbase")
		cbase_lib = ws.NewProject("cbase_lib", cbaseAbsPath, axe.ProjectTypeCppLib, cbaseProjectConfig)
		cbase_lib.ProjectFilename = "cbase_lib"
		cbase_lib.GlobFiles(filepath.Join(rootAbsPath, "cbase"), filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
		cbase_lib.GlobFiles(filepath.Join(rootAbsPath, "cbase"), filepath.Join("source", "main", "cpp", "^**", "*.m"))
		cbase_lib.GlobFiles(filepath.Join(rootAbsPath, "cbase"), filepath.Join("source", "main", "cpp", "^**", "*.mm"))
		cbase_lib.GlobFiles(filepath.Join(rootAbsPath, "cbase"), filepath.Join("source", "main", "include", "^**", "*.h"))
		cbase_lib.GlobFiles(filepath.Join(rootAbsPath, "cbase"), filepath.Join("source", "main", "include", "^**", "*.inl"))

		m.createDefaultProjectConfiguration(cbase_lib, axe.ConfigTypeDebug|axe.ConfigTypeTest)
		m.createDefaultProjectConfiguration(cbase_lib, axe.ConfigTypeRelease|axe.ConfigTypeTest)
	}

	// ccore library project
	ccoreProjectConfig := axe.NewVisualStudioProjectConfig(devVersion)
	{
		ccoreProjectConfig.Group = "cpp-library"
		ccoreProjectConfig.Type = axe.ProjectTypeCppLib
		ccoreProjectConfig.IsGuiApp = false
		ccoreProjectConfig.PchHeader = ""
		ccoreProjectConfig.Dependencies = []string{}
		ccoreProjectConfig.MultiThreadedBuild = true
		ccoreProjectConfig.CppAsObjCpp = false

		ccoreAbsPath := filepath.Join(rootAbsPath, "ccore")
		ccore_lib = ws.NewProject("ccore_lib", ccoreAbsPath, axe.ProjectTypeCppLib, ccoreProjectConfig)
		ccore_lib.ProjectFilename = "ccore_lib"
		ccore_lib.GlobFiles(filepath.Join(rootAbsPath, "ccore"), filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
		ccore_lib.GlobFiles(filepath.Join(rootAbsPath, "ccore"), filepath.Join("source", "main", "cpp", "^**", "*.m"))
		ccore_lib.GlobFiles(filepath.Join(rootAbsPath, "ccore"), filepath.Join("source", "main", "cpp", "^**", "*.mm"))
		ccore_lib.GlobFiles(filepath.Join(rootAbsPath, "ccore"), filepath.Join("source", "main", "include", "^**", "*.h"))
		ccore_lib.GlobFiles(filepath.Join(rootAbsPath, "ccore"), filepath.Join("source", "main", "include", "^**", "*.inl"))

		m.createDefaultProjectConfiguration(ccore_lib, axe.ConfigTypeDebug|axe.ConfigTypeTest)
		m.createDefaultProjectConfiguration(ccore_lib, axe.ConfigTypeRelease|axe.ConfigTypeTest)
	}

	// cunittest library project
	cunittestProjectConfig := axe.NewVisualStudioProjectConfig(devVersion)
	{
		cunittestProjectConfig.Group = "unittest/cpp-library"
		cunittestProjectConfig.Type = axe.ProjectTypeCppLib
		cunittestProjectConfig.IsGuiApp = false
		cunittestProjectConfig.PchHeader = ""
		cunittestProjectConfig.Dependencies = []string{}
		cunittestProjectConfig.MultiThreadedBuild = true
		cunittestProjectConfig.CppAsObjCpp = false

		cunittestAbsPath := filepath.Join(rootAbsPath, "cunittest")
		cunittest_lib = ws.NewProject("cunittest_lib", cunittestAbsPath, axe.ProjectTypeCppLib, cunittestProjectConfig)
		cunittest_lib.ProjectFilename = "cunittest_lib"
		cunittest_lib.GlobFiles(filepath.Join(rootAbsPath, "cunittest"), filepath.Join("source", "main", "cpp", "^**", "*.cpp"))
		cunittest_lib.GlobFiles(filepath.Join(rootAbsPath, "cunittest"), filepath.Join("source", "main", "include", "^**", "*.h"))

		m.createDefaultProjectConfiguration(cunittest_lib, axe.ConfigTypeDebug|axe.ConfigTypeTest)
		m.createDefaultProjectConfiguration(cunittest_lib, axe.ConfigTypeRelease|axe.ConfigTypeTest)
	}

	// cbase unittest project, this is an executable
	cbaseTestProjectConfig := axe.NewVisualStudioProjectConfig(devVersion)
	{
		cbaseTestProjectConfig.Group = "unittest/cpp-exe"
		cbaseTestProjectConfig.Type = axe.ProjectTypeCppExe
		cbaseTestProjectConfig.IsGuiApp = false
		cbaseTestProjectConfig.PchHeader = ""
		cbaseTestProjectConfig.Dependencies = []string{"cbase_lib", "ccore_lib", "cunittest_lib"}
		cbaseTestProjectConfig.MultiThreadedBuild = true
		cbaseTestProjectConfig.CppAsObjCpp = false

		cbaseUnittestAbsPath := filepath.Join(rootAbsPath, "cbase")
		cbase_unittest = ws.NewProject("cbase_unittest", cbaseUnittestAbsPath, axe.ProjectTypeCppExe, cbaseTestProjectConfig)
		cbase_unittest.ProjectFilename = "cbase_unittest"
		cbase_unittest.GlobFiles(filepath.Join(rootAbsPath, "cbase"), filepath.Join("source", "test", "cpp", "^**", "*.cpp"))
		cbase_unittest.GlobFiles(filepath.Join(rootAbsPath, "cbase"), filepath.Join("source", "test", "include", "^**", "*.h"))

		m.createDefaultProjectConfiguration(cbase_unittest, axe.ConfigTypeDebug|axe.ConfigTypeTest)
		m.createDefaultProjectConfiguration(cbase_unittest, axe.ConfigTypeRelease|axe.ConfigTypeTest)
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
